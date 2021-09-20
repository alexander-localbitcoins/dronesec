package main

import (
	"errors"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/alexander-localbitcoins/dronesec"
	"github.com/alexander-localbitcoins/dronesec/client"
	"github.com/alexander-localbitcoins/logger"
)

const usage string = `USAGE: Secret must be a file named "secret_name.secret" where the filename without extension will be used as the secret name and content as data. You place multiple secrets if you want. Configuration is done using the following environmental variables:
DRONE_SYSTEM_PROTO
DRONE_SYSTEM_HOST
DRONE_REPO_OWNER
DRONE_REPO_NAME
DRONE_TOKEN
INSECURE // Ignore certificates
DEBUG // Debug log
QUIET // Ignore most messages
`

var SecretNotFoundError = errors.New("Error: did not find secret file of form: \"secret_name.secret\"")

type Secret struct {
	Name string
	Data string
}

func findSecretFile() (res []*Secret, err error) {
	matchingFiles, err := filepath.Glob("*.secret")
	for _, f := range matchingFiles {
		data, err := ioutil.ReadFile(f)
		if err != nil {
			return nil, err
		}
		res = append(res, &Secret{strings.TrimSuffix(f, filepath.Ext(f)), string(data)})
	}
	return res, nil
}

// USAGE, run container with a file named secret_name.secret at root, where the file name will be used as secret name and content as data
func main() {
	var log logger.Logger
	if os.Getenv("DEBUG") != "" {
		log = logger.NewLogger(logger.Debug)
	} else if os.Getenv("QUIET") != "" {
		log = logger.NewLogger(logger.Quiet)
	} else {
		log = logger.NewLogger(0)
	}
	secs, err := findSecretFile()
	if err != nil {
		log.Error(SecretNotFoundError)
		os.Exit(1)
	}
	var flags client.BuilderFlags
	if os.Getenv("INSECURE") != "" {
		flags = client.Insecure
	} else {
		flags = 0
	}
	ds, err := dronesec.NewDroneSec((&url.URL{Scheme: os.Getenv("DRONE_SYSTEM_PROTO"), Host: os.Getenv("DRONE_SYSTEM_HOST")}).String(), os.Getenv("DRONE_REPO_OWNER"), os.Getenv("DRONE_REPO_NAME"), os.Getenv("DRONE_TOKEN"), "", flags, log)
	if err != nil {
		log.Debug(err)
		log.Info(usage)
		os.Exit(2)
	}
	for _, s := range secs {
		ds.Create(s.Name, s.Data)
		if err != nil {
			log.Error(err)
			os.Exit(3)
		}
	}
}
