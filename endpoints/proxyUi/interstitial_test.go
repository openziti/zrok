package proxyUi

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestEmbeddedInterstitialIsServed(t *testing.T) {
	embedded, err := FS.ReadFile("interstitial.html")
	if err != nil {
		t.Fatalf("expected embedded interstitial page, got %v", err)
	}
	if len(embedded) == 0 {
		t.Fatal("expected non-empty embedded interstitial page")
	}

	w := httptest.NewRecorder()
	WriteInterstitialAnnounce(w, "")

	if got, want := w.Code, http.StatusOK; got != want {
		t.Fatalf("expected status %d, got %d", want, got)
	}
	if !bytes.Equal(w.Body.Bytes(), embedded) {
		t.Fatalf("expected the embedded interstitial page, got %q", w.Body.String())
	}
}
