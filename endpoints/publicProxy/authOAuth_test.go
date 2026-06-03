package publicProxy

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

var testOauthCfg = &OauthConfig{
	EndpointUrl: "https://oauth.example.com",
	CookieName:  "zrok_session",
}

func TestOauthUnauthorizedWithOriginSetsCORSHeaders(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/resource", nil)
	req.Header.Set("Origin", "https://app.example.com")

	oauthUnauthorized(rec, req, testOauthCfg, "github", "share.example.com/api", 3*time.Hour)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "https://app.example.com" {
		t.Fatalf("expected Access-Control-Allow-Origin %q, got %q", "https://app.example.com", got)
	}
	if got := rec.Header().Get("Access-Control-Allow-Credentials"); got != "true" {
		t.Fatalf("expected Access-Control-Allow-Credentials %q, got %q", "true", got)
	}
	if got := rec.Header().Get("Vary"); !strings.Contains(got, "Origin") {
		t.Fatalf("expected Vary to contain %q, got %q", "Origin", got)
	}
}

func TestOauthUnauthorizedWithoutOriginOmitsCORSHeaders(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/resource", nil)

	oauthUnauthorized(rec, req, testOauthCfg, "github", "share.example.com/api", 3*time.Hour)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("expected no Access-Control-Allow-Origin, got %q", got)
	}
	if got := rec.Header().Get("Access-Control-Allow-Credentials"); got != "" {
		t.Fatalf("expected no Access-Control-Allow-Credentials, got %q", got)
	}
}

func TestOauthUnauthorizedResponseBody(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/resource", nil)

	oauthUnauthorized(rec, req, testOauthCfg, "github", "share.example.com/api", 3*time.Hour)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
	if got := rec.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("expected Content-Type application/json, got %q", got)
	}
	var body struct {
		LoginURL string `json:"login_url"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}
	if !strings.Contains(body.LoginURL, "/github/login") {
		t.Fatalf("expected login_url to contain /github/login, got %q", body.LoginURL)
	}
	if !strings.Contains(body.LoginURL, "return_token=true") {
		t.Fatalf("expected login_url to contain return_token=true, got %q", body.LoginURL)
	}
	if got := rec.Header().Get("Location"); got != body.LoginURL {
		t.Fatalf("expected Location header %q to match login_url %q", got, body.LoginURL)
	}
}
