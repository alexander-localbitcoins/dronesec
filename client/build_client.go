package client

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"

	"github.com/alexander-localbitcoins/logger"
	"github.com/drone/drone-go/drone"
	"golang.org/x/oauth2"
)

type BuilderFlags uint32

const (
	Insecure BuilderFlags = 1 << iota
)

const (
	certLoadErrorHeader = "Error loading certificates: %w"
)

var (
	FailedToLoadCertError   = errors.New("Failed to load certificates")
	InvalidCertificateError = errors.New("Invalid certificate")
	NullLogger              = errors.New("Logger cannot be null")
	EmptyToken              = errors.New("Empty token is invalid")
)

func NewDroneBuilder(log logger.Logger, certsLocation string, flag BuilderFlags) (*droneBuilder, error) {
	b := new(droneBuilder)
	if log == nil {
		return nil, NullLogger
	} else {
		b.log = log
	}
	if flag&Insecure != 0 {
		b.insecure = true
		b.log.Warning("Ignoring certificates, this is dangerous!")
	}
	if certsLocation != "" {
		certs, err := b.loadCerts(certsLocation)
		if err != nil {
			return nil, err
		}
		b.certs = certs
	} else {
		b.log.Info("No certificates provided")
	}
	return b, nil
}

type droneBuilder struct {
	log      logger.Logger
	certs    *x509.CertPool
	insecure bool
}

func (b *droneBuilder) BuildDroneClient(host string, token string) (drone.Client, error) {
	// create an http client with oauth authentication.
	if token == "" {
		return nil, EmptyToken
	}
	auther := new(oauth2.Config).Client(
		oauth2.NoContext,
		&oauth2.Token{
			AccessToken: token,
		},
	)
	trans := auther.Transport.(*oauth2.Transport)
	if b.certs != nil {
		trans.Base = &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs:            b.certs,
				InsecureSkipVerify: b.insecure,
			},
		}
	} else {
		trans.Base = &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: b.insecure,
			},
		}
	}
	// create the drone client with authenticator
	client := drone.NewClient(host, auther)

	// check connection and token
	_, err := client.Self()
	if err != nil {
		return nil, fmt.Errorf("Failed to connect to server: %w", err)
	}
	return client, nil
}

func (b *droneBuilder) loadCerts(location string) (*x509.CertPool, error) {
	caCertPool := x509.NewCertPool()
	locationInfo, err := os.Stat(location)
	if err != nil {
		b.log.Debug(certFormatErr(err))
		return nil, FailedToLoadCertError
	}
	if !locationInfo.IsDir() {
		caCert, err := ioutil.ReadFile(location)
		if err != nil {
			b.log.Debug(certFormatErr(err))
			return nil, FailedToLoadCertError
		}
		caCertPool.AppendCertsFromPEM(caCert)
	} else {
		fInfos, err := ioutil.ReadDir(location)
		if err != nil {
			b.log.Debug(certFormatErr(err))
			return nil, FailedToLoadCertError
		}
		for _, fInfo := range fInfos {
			if fInfo.IsDir() {
				continue
			}
			caCert, err := ioutil.ReadFile(filepath.Join(location, fInfo.Name()))
			if err != nil {
				b.log.Debug(certFormatErr(err))
				continue
			}
			caCertPool.AppendCertsFromPEM(caCert)
		}
	}
	if len(caCertPool.Subjects()) == 0 {
		if locationInfo.IsDir() {
			b.log.Warning("No certificates found or failed to load, attempting without")
			return nil, nil
		} else {
			b.log.Error(certFormatErr(InvalidCertificateError))
			return nil, FailedToLoadCertError
		}
	}
	return caCertPool, nil
}

func certFormatErr(err error) error {
	return fmt.Errorf(certLoadErrorHeader, err)
}
