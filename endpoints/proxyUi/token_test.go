package proxyUi

import (
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestWriteTokenDisplayStatusOK(t *testing.T) {
	w := httptest.NewRecorder()
	WriteTokenDisplay(w, "my.jwt.token", time.Now().Add(time.Hour))
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestWriteTokenDisplayContainsToken(t *testing.T) {
	w := httptest.NewRecorder()
	token := "header.payload.signature"
	WriteTokenDisplay(w, token, time.Now().Add(time.Hour))
	if !strings.Contains(w.Body.String(), token) {
		t.Fatalf("expected body to contain token %q", token)
	}
}

func TestWriteTokenDisplayEscapesHTML(t *testing.T) {
	w := httptest.NewRecorder()
	WriteTokenDisplay(w, "<script>alert(1)</script>", time.Now().Add(time.Hour))
	body := w.Body.String()
	if strings.Contains(body, "<script>") {
		t.Fatal("expected HTML-escaped token, got raw <script> tag")
	}
}

func TestWriteTokenDisplayContainsHeaderName(t *testing.T) {
	w := httptest.NewRecorder()
	WriteTokenDisplay(w, "tok", time.Now().Add(time.Hour))
	if !strings.Contains(w.Body.String(), "X-Zrok-Session") {
		t.Fatal("expected body to reference X-Zrok-Session header")
	}
}
