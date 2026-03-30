package proxyUi

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestWriteUnauthorizedEscapesErrorHtml(t *testing.T) {
	w := httptest.NewRecorder()

	WriteUnauthorized(w, UnauthorizedData().WithError(errors.New(`x</code><script>alert(1)</script><code>`)))

	if got, want := w.Code, http.StatusUnauthorized; got != want {
		t.Fatalf("expected status %d, got %d", want, got)
	}

	body := w.Body.String()
	if strings.Contains(body, "<script>alert(1)</script>") {
		t.Fatalf("expected script tag to be escaped, body was %q", body)
	}
	if !strings.Contains(body, "&lt;script&gt;alert(1)&lt;/script&gt;") {
		t.Fatalf("expected escaped script tag in body, got %q", body)
	}
}

func TestWriteNotFoundPreservesIntentionalHtml(t *testing.T) {
	w := httptest.NewRecorder()

	WriteNotFound(w, NotFoundData("share-token"))

	body := w.Body.String()
	if !strings.Contains(body, "<code>share-token</code>") {
		t.Fatalf("expected banner html to be preserved, got %q", body)
	}
	if !strings.Contains(body, "<code>zrok2 share</code>") {
		t.Fatalf("expected message html to be preserved, got %q", body)
	}
}
