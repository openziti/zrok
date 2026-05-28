package dynamicProxy

import (
	"testing"
	"time"
)

// testEncryptionKey is a 32-byte key suitable for AES-256-GCM.
var testEncryptionKey = []byte("01234567890123456789012345678901")

func TestBuildSessionJWTRejectsEmptyTargetHost(t *testing.T) {
	cfg := &oauthConfig{SessionLifetime: time.Hour}
	_, err := buildSessionJWT(sessionCookieRequest{
		oauthCfg:      cfg,
		targetHost:    "",
		signingKey:    []byte("test-signing-key"),
		encryptionKey: testEncryptionKey,
	})
	if err == nil {
		t.Fatal("expected error for empty targetHost, got nil")
	}
}

func TestBuildSessionJWTRoundTrip(t *testing.T) {
	cfg := &oauthConfig{SessionLifetime: time.Hour}

	got, err := buildSessionJWT(sessionCookieRequest{
		oauthCfg:        cfg,
		email:           "user@example.com",
		provider:        "github",
		targetHost:      "share.example.com",
		refreshInterval: 3 * time.Hour,
		signingKey:      []byte("test-signing-key"),
		encryptionKey:   testEncryptionKey,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == "" {
		t.Fatal("expected non-empty JWT string")
	}
}

func TestBuildSessionJWTStripsPathFromTargetHost(t *testing.T) {
	cfg := &oauthConfig{SessionLifetime: time.Hour}

	got, err := buildSessionJWT(sessionCookieRequest{
		oauthCfg:      cfg,
		targetHost:    "share.example.com/some/path",
		signingKey:    []byte("test-signing-key"),
		encryptionKey: testEncryptionKey,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == "" {
		t.Fatal("expected non-empty JWT string")
	}
}
