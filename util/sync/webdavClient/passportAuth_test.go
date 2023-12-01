package webdavClient

import (
	"bytes"
	"net/http"
	"net/url"
	"regexp"
	"testing"
)

// testing the creation is enough as it handles the authorization during init
func TestNewPassportAuth(t *testing.T) {
	user := "user"
	pass := "password"
	p1 := "some,comma,separated,values"
	token := "from-PP='token'"

	authHandler := func(h http.Handler) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			reg, err := regexp.Compile("Passport1\\.4 sign-in=" + url.QueryEscape(user) + ",pwd=" + url.QueryEscape(pass) + ",OrgVerb=GET,OrgUrl=.*," + p1)
			if err != nil {
				t.Error(err)
			}
			if reg.MatchString(r.Header.Get("Authorization")) {
				w.Header().Set("Authentication-Info", token)
				w.WriteHeader(200)
				return
			}
		}
	}
	authsrv, _, _ := newAuthSrv(t, authHandler)
	defer authsrv.Close()

	dataHandler := func(h http.Handler) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			reg, err := regexp.Compile("Passport1\\.4 " + token)
			if err != nil {
				t.Error(err)
			}
			if reg.MatchString(r.Header.Get("Authorization")) {
				w.Header().Set("Set-Cookie", "Pass=port")
				h.ServeHTTP(w, r)
				return
			}
			for _, c := range r.Cookies() {
				if c.Name == "Pass" && c.Value == "port" {
					h.ServeHTTP(w, r)
					return
				}
			}
			w.Header().Set("Www-Authenticate", "Passport1.4 "+p1)
			http.Redirect(w, r, authsrv.URL+"/", 302)
		}
	}
	srv, _, _ := newAuthSrv(t, dataHandler)
	defer srv.Close()

	cli := NewClient(srv.URL, user, pass)
	data, err := cli.Read("/hello.txt")
	if err != nil {
		t.Errorf("got error=%v; want nil", err)
	}
	if !bytes.Equal(data, []byte("hello gowebdav\n")) {
		t.Logf("got data=%v; want=hello gowebdav", data)
	}
}
