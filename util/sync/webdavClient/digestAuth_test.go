package webdavClient

import (
	"net/http"
	"strings"
	"testing"
)

func TestNewDigestAuth(t *testing.T) {
	a := &DigestAuth{user: "user", pw: "password", digestParts: make(map[string]string, 0)}

	ex := "DigestAuth login: user"
	if a.String() != ex {
		t.Error("expected: " + ex + " got: " + a.String())
	}

	if a.Clone() == a {
		t.Error("expected a different instance")
	}

	if a.Close() != nil {
		t.Error("expected close without errors")
	}
}

func TestDigestAuthAuthorize(t *testing.T) {
	a := &DigestAuth{user: "user", pw: "password", digestParts: make(map[string]string, 0)}
	rq, _ := http.NewRequest("GET", "http://localhost/", nil)
	a.Authorize(nil, rq, "/")
	// TODO this is a very lazy test it cuts of cnonce
	ex := `Digest username="user", realm="", nonce="", uri="/", nc=1, cnonce="`
	if strings.Index(rq.Header.Get("Authorization"), ex) != 0 {
		t.Error("got wrong Authorization header: " + rq.Header.Get("Authorization"))
	}
}
