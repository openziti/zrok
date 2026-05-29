package proxyUi

import (
	"html/template"
	"net/http"
	"time"

	"github.com/michaelquigley/df/dl"
	"github.com/pkg/errors"
)

var tokenTmpl *template.Template

func init() {
	if data, err := FS.ReadFile("token.html"); err == nil {
		tokenTmpl, err = template.New("token").Parse(string(data))
		if err != nil {
			panic(errors.Wrap(err, "unable to parse embedded template 'token.html'"))
		}
	} else {
		panic(errors.Wrap(err, "unable to load embedded template 'token.html'"))
	}
}

type tokenDisplayData struct {
	Token      string
	Expiry     string
	TargetHost string
}

// WriteTokenDisplay renders the session token display page. The token value is
// HTML-escaped by html/template, so arbitrary JWT content is safe to render.
func WriteTokenDisplay(w http.ResponseWriter, token string, expiry time.Time, targetHost string) {
	data := tokenDisplayData{
		Token:      token,
		Expiry:     expiry.UTC().Format(time.RFC1123),
		TargetHost: targetHost,
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	if err := tokenTmpl.Execute(w, data); err != nil {
		dl.Errorf("failed to execute token template: %v", err)
	}
}
