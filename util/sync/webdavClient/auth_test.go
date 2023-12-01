package webdavClient

import (
	"bytes"
	"net/http"
	"strings"
	"testing"
)

func TestEmptyAuth(t *testing.T) {
	auth := NewEmptyAuth()
	srv, _, _ := newAuthSrv(t, basicAuth)
	defer srv.Close()
	cli := NewAuthClient(srv.URL, auth)
	if err := cli.Connect(); err == nil {
		t.Fatalf("got nil want error")
	}
}

func TestRedirectAuthWIP(t *testing.T) {
	hasPassedAuthServer := false
	authHandler := func(h http.Handler) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			if user, passwd, ok := r.BasicAuth(); ok {
				if user == "user" && passwd == "password" {
					hasPassedAuthServer = true
					w.WriteHeader(200)
					return
				}
			}
			w.Header().Set("Www-Authenticate", `Basic realm="x"`)
			w.WriteHeader(401)
		}
	}

	psrv, _, _ := newAuthSrv(t, authHandler)
	defer psrv.Close()

	dataHandler := func(h http.Handler) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			hasAuth := strings.Contains(r.Header.Get("Authorization"), "Basic dXNlcjpwYXNzd29yZA==")

			if hasPassedAuthServer && hasAuth {
				h.ServeHTTP(w, r)
				return
			}
			w.Header().Set("Www-Authenticate", `Basic realm="x"`)
			http.Redirect(w, r, psrv.URL+"/", 302)
		}
	}

	srv, _, _ := newAuthSrv(t, dataHandler)
	defer srv.Close()
	cli := NewClient(srv.URL, "user", "password")
	data, err := cli.Read("/hello.txt")
	if err != nil {
		t.Logf("WIP got error=%v; want nil", err)
	}
	if bytes.Compare(data, []byte("hello gowebdav\n")) != 0 {
		t.Logf("WIP got data=%v; want=hello gowebdav", data)
	}
}
