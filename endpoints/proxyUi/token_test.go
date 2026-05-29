package proxyUi

import (
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestWriteTokenDisplayStatusOK(t *testing.T) {
	w := httptest.NewRecorder()
	WriteTokenDisplay(w, "my.jwt.token", time.Now().Add(time.Hour), "share.example.com")
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestWriteTokenDisplayContainsToken(t *testing.T) {
	w := httptest.NewRecorder()
	token := "header.payload.signature"
	WriteTokenDisplay(w, token, time.Now().Add(time.Hour), "share.example.com")
	if !strings.Contains(w.Body.String(), token) {
		t.Fatalf("expected body to contain token %q", token)
	}
}

func TestWriteTokenDisplayEscapesHTML(t *testing.T) {
	w := httptest.NewRecorder()
	rawToken := "<script>alert(1)</script>"
	WriteTokenDisplay(w, rawToken, time.Now().Add(time.Hour), "share.example.com")
	body := w.Body.String()
	if strings.Contains(body, rawToken) {
		t.Fatal("expected HTML-escaped token, got raw unescaped token value")
	}
}

func TestWriteTokenDisplayContainsHeaderName(t *testing.T) {
	w := httptest.NewRecorder()
	WriteTokenDisplay(w, "tok", time.Now().Add(time.Hour), "share.example.com")
	if !strings.Contains(w.Body.String(), "X-Zrok-Session") {
		t.Fatal("expected body to reference X-Zrok-Session header")
	}
}

func TestWriteTokenDisplayPostMessageScriptPresentWithTargetHost(t *testing.T) {
	w := httptest.NewRecorder()
	WriteTokenDisplay(w, "tok", time.Now().Add(time.Hour), "share.example.com")
	if !strings.Contains(w.Body.String(), "postMessage") {
		t.Fatal("expected postMessage script when targetHost is set")
	}
}

func TestWriteTokenDisplayNoPostMessageScriptWithoutTargetHost(t *testing.T) {
	w := httptest.NewRecorder()
	WriteTokenDisplay(w, "tok", time.Now().Add(time.Hour), "")
	if strings.Contains(w.Body.String(), "postMessage") {
		t.Fatal("expected no postMessage script when targetHost is empty")
	}
}
