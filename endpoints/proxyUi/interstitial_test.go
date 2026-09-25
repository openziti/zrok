package proxyUi

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// Regression for issue #1262: the embedded interstitial filename must match
// what WriteInterstitialAnnounce reads from FS.
func TestEmbeddedInterstitialReadable(t *testing.T) {
	data, err := FS.ReadFile("interstitial.html")
	if err != nil {
		t.Fatalf("FS.ReadFile(\"interstitial.html\") failed: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("embedded interstitial.html is empty")
	}
	if !strings.Contains(string(data), "<html") && !strings.Contains(string(data), "<!DOCTYPE") && !strings.Contains(strings.ToLower(string(data)), "html") {
		t.Fatalf("embedded interstitial.html does not look like HTML (len=%d)", len(data))
	}
}

func TestWriteInterstitialAnnounceUsesEmbedded(t *testing.T) {
	// Ensure we exercise the embedded path (no external override).
	externalFile = nil

	w := httptest.NewRecorder()
	WriteInterstitialAnnounce(w, "")

	if got, want := w.Code, http.StatusOK; got != want {
		t.Fatalf("expected status %d, got %d", want, got)
	}
	if w.Body.Len() == 0 {
		t.Fatal("expected non-empty interstitial body from embedded HTML")
	}
}
