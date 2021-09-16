package client

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/alexander-localbitcoins/logger/mock"
)

const (
	numberOfCertsInDir = 1
)

func TestNewDroneBuilder(t *testing.T) {
	// Test cert directory
	if b, _ := NewDroneBuilder(new(mock.MockLogger), "../testdata/test_certs/", 0); b == nil {
		t.Error("Failed to load certificates in directory")
	} else {
		if got := len(b.certs.Subjects()); got != numberOfCertsInDir {
			t.Errorf("Found incorrect amount of certs. Expected: %d, got: %d", numberOfCertsInDir, got)
		}
	}

	// Test direct path to cert
	if b, _ := NewDroneBuilder(new(mock.MockLogger), "../testdata/test_certs/cert.pem", 0); b == nil {
		t.Error("Failed to load single certificate")
	} else {
		if got := len(b.certs.Subjects()); got != 1 {
			t.Errorf("Found incorrect amount of certs. Expected: %d, got: %d", 1, got)
		}
	}

	// test insecure flag
	l := new(mock.MockLogger)
	b, _ := NewDroneBuilder(l, "", Insecure)
	if !b.insecure {
		t.Error("insecure flag not set")
	}
	if l.NotInWarnings("Ignoring certificates, this is dangerous!") {
		t.Error("Warning about insecure flag not found")
	}

	// Test empty cert dir
	l = new(mock.MockLogger)
	if b, _ := NewDroneBuilder(l, "../testdata/", 0); b.certs != nil {
		t.Error("Found certificates when it shouldn't have")
	} else if l.NotInWarnings("No certificates found or failed to load, attempting without") {
		t.Error("Didn't find error about no certificate")
	}

	// test invalid certificate
	l = new(mock.MockLogger)
	if b, _ := NewDroneBuilder(l, "../testdata/invalid_cert.pem", 0); b != nil {
		t.Error("Builder should've failed but didn't")
	} else if l.NotInErrors(InvalidCertificateError) {
		t.Error("Invalid certificate error not found")
	}

	// test invalid path
	l = new(mock.MockLogger)
	if b, _ := NewDroneBuilder(l, "../testdata/asoidnjasdnbasodi", 0); b != nil {
		t.Error("Builder should've failed but didn't")
	} else if l.NotInDebugs(os.ErrNotExist) {
		t.Error("Path does not exist error not found")
	}
}

const userEndpointTestPayload string = `{"id":11,"login":"drone-lbc","email":"","machine":false,"admin":false,"active":true,"avatar":"https://avatars.githubusercontent.com/u/89182792?v=4","syncing":false,"synced":1631169513,"created":1631169512,"updated":1631169512,"last_login":1631169512}`
const rawToken string = "SUPERSECRETTOKEN"

func TestBuildDroneClient(t *testing.T) {
	l := new(mock.MockLogger)
	b, _ := NewDroneBuilder(l, "", Insecure)
	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		if strings.Split(token, " ")[1] != rawToken {
			http.Error(w, "Invalid token", 401)
			return
		}
		fmt.Fprintln(w, userEndpointTestPayload)
	}))
	if _, err := b.BuildDroneClient(ts.URL, rawToken); err != nil {
		t.Error("Failed to verify connection, err :", err)
	}
	if _, err := b.BuildDroneClient(ts.URL, ""); !errors.Is(err, EmptyToken) {
		t.Error("Empty token did not return correct error")
	}
	if _, err := b.BuildDroneClient(ts.URL, "asoidjasdoi"); !strings.Contains(err.Error(), "Invalid token") && !strings.Contains(err.Error(), "client error 401") {
		t.Error("Did not receive server error, err:", err)
	}
	ts.Close()
}
