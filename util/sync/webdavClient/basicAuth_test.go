package webdavClient

import (
	"net/http"
	"testing"
)

func TestNewBasicAuth(t *testing.T) {
	a := &BasicAuth{user: "user", pw: "password"}

	ex := "BasicAuth login: user"
	if a.String() != ex {
		t.Error("expected: " + ex + " got: " + a.String())
	}

	if a.Clone() != a {
		t.Error("expected the same instance")
	}

	if a.Close() != nil {
		t.Error("expected close without errors")
	}
}

func TestBasicAuthAuthorize(t *testing.T) {
	a := &BasicAuth{user: "user", pw: "password"}
	rq, _ := http.NewRequest("GET", "http://localhost/", nil)
	a.Authorize(nil, rq, "/")
	if rq.Header.Get("Authorization") != "Basic dXNlcjpwYXNzd29yZA==" {
		t.Error("got wrong Authorization header: " + rq.Header.Get("Authorization"))
	}
}

func TestPreemtiveBasicAuth(t *testing.T) {
	a := &BasicAuth{user: "user", pw: "password"}
	auth := NewPreemptiveAuth(a)
	n, b := auth.NewAuthenticator(nil)
	if b != nil {
		t.Error("expected body to be nil")
	}
	if n != a {
		t.Error("expected the same instance")
	}

	srv, _, _ := newAuthSrv(t, basicAuth)
	defer srv.Close()
	cli := NewAuthClient(srv.URL, auth)
	if err := cli.Connect(); err != nil {
		t.Fatalf("got error: %v, want nil", err)
	}
}
