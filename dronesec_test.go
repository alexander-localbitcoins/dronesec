package dronesec

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/alexander-localbitcoins/dronesec/client"
	"github.com/alexander-localbitcoins/logger/mock"
)

type MethodAndEndpoint struct {
	Method   string
	Endpoint string
}

var enpointsAndPayload = map[MethodAndEndpoint]string{
	MethodAndEndpoint{"GET", "/api/user"}:                             `{"id":10,"login":"username","email":"","machine":false,"admin":false,"active":true,"avatar":"","syncing":false,"synced":12389,"created":1631169300,"updated":1631169400,"last_login":1631169500}`,
	MethodAndEndpoint{"GET", "/api/repos/owner/repo/secrets/test"}:    `{"id":1,"repo_id":2,"name":"test"}`,
	MethodAndEndpoint{"POST", "/api/repos/owner/repo/secrets"}:        `{"message":"UNIQUE constraint failed: secrets.secret_repo_id, secrets.secret_name"}`,
	MethodAndEndpoint{"PATCH", "/api/repos/owner/repo/secrets/test"}:  `{"id":2,"repo_id":2,"name":"test"}`,
	MethodAndEndpoint{"DELETE", "/api/repos/owner/repo/secrets/test"}: `{"message":"sql: no rows in result set"}`,
	// MethodAndEndpoint{"DELETE", "/api/repos/owner/repo/secrets/test"}: "", // There is not response if secret is found and deleted
}

const rawToken = "SUPERSECRETTOKEN"

func TestDroneSec(t *testing.T) {
	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		if strings.Split(token, " ")[1] != rawToken {
			http.Error(w, "Invalid token", 401)
			return
		}
		fmt.Fprintln(w, enpointsAndPayload[MethodAndEndpoint{r.Method, r.URL.String()}])
	}))
	l := new(mock.MockLogger)
	ds, err := NewDroneSec(ts.URL, "owner", "repo", rawToken, "./testdata/test_certs/", client.Insecure, l)
	if err != nil {
		t.Log(err)
		t.Fatal("Failed to build dronesec")
	}
	if err := ds.Create("test", "test"); err != nil {
		for _, w := range l.Warnings {
			t.Log(w)
		}
		for _, d := range l.Debugs {
			t.Log(d)
		}
		t.Fatal(err)
	}
	if err := ds.Delete("test"); err != nil {
		for _, w := range l.Warnings {
			t.Log(w)
		}
		for _, d := range l.Debugs {
			t.Log(d)
		}
		t.Fatal(err)
	}
}

func TestDroneSecWithRealEndpoint(t *testing.T) {
	l := new(mock.MockLogger)
	ds, err := NewDroneSec((&url.URL{Scheme: os.Getenv("DRONE_SYSTEM_PROTO"), Host: os.Getenv("DRONE_SYSTEM_HOST")}).String(), os.Getenv("DRONE_REPO_OWNER"), os.Getenv("DRONE_REPO_NAME"), os.Getenv("DRONE_TOKEN"), "", client.Insecure, l)
	if errors.Is(err, EmptyInput) {
		t.Log(err)
		t.Skip("Some inputs required for real life test not provided (passed as env vars), skipping test")
	}
	if err != nil {
		for _, w := range l.Warnings {
			t.Log(w)
		}
		for _, d := range l.Debugs {
			t.Log(d)
		}
		t.Fatal(err)
	}
	if err := ds.Create("test", "test"); err != nil {
		for _, w := range l.Warnings {
			t.Log(w)
		}
		for _, d := range l.Debugs {
			t.Log(d)
		}
		t.Fatal(err)
	}
	if err := ds.Delete("test"); err != nil {
		for _, w := range l.Warnings {
			t.Log(w)
		}
		for _, d := range l.Debugs {
			t.Log(d)
		}
		t.Fatal(err)
	}
}
