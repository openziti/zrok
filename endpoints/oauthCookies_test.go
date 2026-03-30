package endpoints

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"
)

type testOAuthCookieConfig struct {
	cookieName      string
	cookieDomain    string
	maxCookieSize   int
	maxCookieChunks int
	sessionLifetime time.Duration
}

func (c testOAuthCookieConfig) GetCookieName() string             { return c.cookieName }
func (c testOAuthCookieConfig) GetCookieDomain() string           { return c.cookieDomain }
func (c testOAuthCookieConfig) GetMaxCookieSize() int             { return c.maxCookieSize }
func (c testOAuthCookieConfig) GetMaxCookieChunks() int           { return c.maxCookieChunks }
func (c testOAuthCookieConfig) GetSessionLifetime() time.Duration { return c.sessionLifetime }

func TestGetSessionCookieSingleRoundTrip(t *testing.T) {
	cfg := testOAuthCookieConfig{
		cookieName:      "zrok-auth-session",
		maxCookieSize:   3072,
		maxCookieChunks: 10,
		sessionLifetime: time.Hour,
	}

	token := "single-session-token"
	req := requestWithSessionCookie(t, cfg, token)

	cookie, err := GetSessionCookie(req, cfg)
	if err != nil {
		t.Fatalf("expected cookie to round trip: %v", err)
	}

	if cookie.Value != token {
		t.Fatalf("expected token %q, got %q", token, cookie.Value)
	}
}

func TestGetSessionCookieStripedRoundTrip(t *testing.T) {
	cfg := testOAuthCookieConfig{
		cookieName:      "zrok-auth-session",
		maxCookieSize:   16,
		maxCookieChunks: 32,
		sessionLifetime: time.Hour,
	}

	token := repeatedToken(2048)
	req := requestWithStripedSessionCookie(t, cfg, token)

	cookie, err := GetSessionCookie(req, cfg)
	if err != nil {
		t.Fatalf("expected striped cookie to round trip: %v", err)
	}

	if cookie.Value != token {
		t.Fatalf("expected token %q, got %q", token, cookie.Value)
	}
}

func TestGetSessionCookieRejectsTooManyChunks(t *testing.T) {
	cfg := testOAuthCookieConfig{
		cookieName:      "zrok-auth-session",
		maxCookieSize:   3072,
		maxCookieChunks: 10,
		sessionLifetime: time.Hour,
	}

	req := httptest.NewRequest(http.MethodGet, "https://example.com", nil)
	req.AddCookie(&http.Cookie{Name: cfg.cookieName, Value: "11|x"})

	_, err := GetSessionCookie(req, cfg)
	if err == nil || !strings.Contains(err.Error(), "exceeds maximum") {
		t.Fatalf("expected oversized chunk count error, got %v", err)
	}
}

func TestGetSessionCookieRejectsInvalidChunkCount(t *testing.T) {
	cfg := testOAuthCookieConfig{
		cookieName:      "zrok-auth-session",
		maxCookieSize:   3072,
		maxCookieChunks: 10,
		sessionLifetime: time.Hour,
	}

	testCases := []string{"0|x", "-1|x", "abc|x", strings.Repeat("9", 40) + "|x"}
	for _, value := range testCases {
		req := httptest.NewRequest(http.MethodGet, "https://example.com", nil)
		req.AddCookie(&http.Cookie{Name: cfg.cookieName, Value: value})

		_, err := GetSessionCookie(req, cfg)
		if err == nil || !strings.Contains(err.Error(), "invalid cookie chunk count") {
			t.Fatalf("expected invalid chunk count error for %q, got %v", value, err)
		}
	}
}

func TestGetSessionCookieMissingChunk(t *testing.T) {
	cfg := testOAuthCookieConfig{
		cookieName:      "zrok-auth-session",
		maxCookieSize:   3072,
		maxCookieChunks: 10,
		sessionLifetime: time.Hour,
	}

	req := httptest.NewRequest(http.MethodGet, "https://example.com", nil)
	req.AddCookie(&http.Cookie{Name: cfg.cookieName, Value: "2|x"})

	_, err := GetSessionCookie(req, cfg)
	if err == nil || !strings.Contains(err.Error(), "missing cookie chunk") {
		t.Fatalf("expected missing chunk error, got %v", err)
	}
}

func TestSetSessionCookieRejectsTooManyChunks(t *testing.T) {
	cfg := testOAuthCookieConfig{
		cookieName:      "zrok-auth-session",
		maxCookieSize:   16,
		maxCookieChunks: 1,
		sessionLifetime: time.Hour,
	}

	w := httptest.NewRecorder()
	err := SetSessionCookie(w, cfg.cookieName, repeatedToken(2048), cfg)
	if err == nil || !strings.Contains(err.Error(), "requires") {
		t.Fatalf("expected chunk limit error, got %v", err)
	}
}

func TestGetEffectiveMaxCookieChunksClamp(t *testing.T) {
	cfg := testOAuthCookieConfig{maxCookieChunks: hardMaxCookieChunks + 5}
	if got, want := getEffectiveMaxCookieChunks(cfg), hardMaxCookieChunks; got != want {
		t.Fatalf("expected hard max %d, got %d", want, got)
	}

	cfg.maxCookieChunks = 0
	if got, want := getEffectiveMaxCookieChunks(cfg), defaultMaxCookieChunks; got != want {
		t.Fatalf("expected default max %d, got %d", want, got)
	}
}

func requestWithSessionCookie(t *testing.T, cfg testOAuthCookieConfig, token string) *http.Request {
	t.Helper()

	w := httptest.NewRecorder()
	if err := SetSessionCookie(w, cfg.cookieName, token, cfg); err != nil {
		t.Fatalf("expected cookie to be set: %v", err)
	}

	res := w.Result()
	t.Cleanup(func() {
		_ = res.Body.Close()
	})

	req := httptest.NewRequest(http.MethodGet, "https://example.com", nil)
	for _, cookie := range res.Cookies() {
		req.AddCookie(cookie)
	}
	return req
}

func requestWithStripedSessionCookie(t *testing.T, cfg testOAuthCookieConfig, token string) *http.Request {
	t.Helper()

	compressedToken, err := CompressToken(token)
	if err != nil {
		t.Fatalf("expected token to compress: %v", err)
	}

	chunkCount, firstChunkSize, err := getChunkCount(compressedToken, cfg.maxCookieSize)
	if err != nil {
		t.Fatalf("expected chunk count to be calculated: %v", err)
	}
	if chunkCount <= 1 {
		t.Fatalf("expected striped cookie fixture to require more than one chunk, got %d", chunkCount)
	}

	req := httptest.NewRequest(http.MethodGet, "https://example.com", nil)
	req.AddCookie(&http.Cookie{
		Name:  cfg.cookieName,
		Value: strconv.Itoa(chunkCount) + "|" + compressedToken[:firstChunkSize],
	})

	offset := firstChunkSize
	for i := 1; i < chunkCount; i++ {
		end := min(offset+cfg.maxCookieSize, len(compressedToken))
		req.AddCookie(&http.Cookie{
			Name:  cfg.cookieName + "_" + strconv.Itoa(i),
			Value: compressedToken[offset:end],
		})
		offset = end
	}

	return req
}

func repeatedToken(size int) string {
	var b strings.Builder
	b.Grow(size)
	for i := 0; i < size; i++ {
		b.WriteString(strconv.Itoa(i % 10))
		b.WriteByte(byte('a' + (i % 26)))
	}
	return b.String()
}
