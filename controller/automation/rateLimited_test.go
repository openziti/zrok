package automation

import (
	"bytes"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/michaelquigley/df/dl"
	"github.com/openziti/zrok/v2/controller/automation/zitifake"
)

type logBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *logBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Write(p)
}

func (b *logBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.String()
}

func captureLogs(t *testing.T) *logBuffer {
	t.Helper()
	logs := &logBuffer{}
	dl.Init(dl.DefaultOptions().JSON().SetOutput(logs))
	t.Cleanup(func() { dl.Init() })
	return logs
}

func TestZitiRateLimitedCreate(t *testing.T) {
	fake, cfg, _ := sessionFixture(t)
	logs := captureLogs(t)
	ziti := mustSession(t, cfg)
	fake.RateLimitCreates(zitifake.Services, true)
	before := zitiRateLimited.Value()

	_, err := ziti.Services.Create(&ServiceOptions{BaseOptions: BaseOptions{Name: "limited"}})

	if !IsRateLimited(err) {
		t.Fatalf("want rate limited, got %T: %v", err, err)
	}
	var limited *RateLimitedError
	if !errors.As(err, &limited) || limited.Limiter != "SERVER_TOO_MANY_REQUESTS" || limited.Method != "POST" || limited.Path != "/edge/management/v1/services" {
		t.Fatalf("rate limited error = %+v", limited)
	}
	if got := zitiRateLimited.Value() - before; got != 1 {
		t.Fatalf("rate limited counter delta = %d, want 1", got)
	}
	if line := logs.String(); !bytes.Contains([]byte(line), []byte("ziti rate limited by 'SERVER_TOO_MANY_REQUESTS' on 'POST /edge/management/v1/services'")) {
		t.Fatalf("missing rate limited log line: %s", line)
	}
	if IsNotFound(err) {
		t.Fatal("rate limited error matched not found")
	}
	if _, _, creates, _ := fake.Counts(); creates != 0 {
		t.Fatalf("created %d services, want 0", creates)
	}
}

func TestZitiRateLimitedAuthentication(t *testing.T) {
	fake, cfg, _ := sessionFixture(t)
	logs := captureLogs(t)
	oldClock := refreshClock
	t.Cleanup(func() { refreshClock = oldClock })
	ziti := mustSession(t, cfg)
	fake.ExpireAll()
	fake.RateLimitAuthentications(1)
	before := zitiRateLimited.Value()

	_, err := ziti.Services.Find(&FilterOptions{})

	var limited *RateLimitedError
	if !IsRateLimited(err) || !errors.As(err, &limited) || limited.Limiter != "authenticate" {
		t.Fatalf("want authenticate rate limited, got %T: %v", err, err)
	}
	if got := zitiRateLimited.Value() - before; got != 1 {
		t.Fatalf("rate limited counter delta = %d, want 1", got)
	}
	if line := logs.String(); !bytes.Contains([]byte(line), []byte("ziti rate limited by 'authenticate' on 'POST /edge/management/v1/authenticate'")) {
		t.Fatalf("missing rate limited log line: %s", line)
	}

	// the failed refresh is remembered like any other: no second login inside the window.
	if _, err := ziti.Services.Find(&FilterOptions{}); !IsRateLimited(err) {
		t.Fatalf("remembered failure = %v", err)
	}
	if _, attempts, _ := fake.AuthCounts(); attempts != 2 {
		t.Fatalf("authentication attempts = %d, want 2", attempts)
	}
	if got := zitiRateLimited.Value() - before; got != 1 {
		t.Fatalf("remembered failure counted again: delta = %d", got)
	}

	refreshClock = func() time.Time { return time.Now().Add(failedRefreshWindow) }
	if _, err := ziti.Services.Find(&FilterOptions{}); err != nil {
		t.Fatalf("refresh after window: %v", err)
	}
	if successful, attempts, _ := fake.AuthCounts(); successful != 2 || attempts != 3 {
		t.Fatalf("successful=%d attempts=%d", successful, attempts)
	}
}

func TestZitiRateLimitedInitialAuthentication(t *testing.T) {
	fake, cfg, _ := sessionFixture(t)
	fake.RateLimitAuthentications(1)

	if _, err := NewZitiAutomation(cfg); !IsRateLimited(err) {
		t.Fatalf("want rate limited build, got %v", err)
	}
	ziti := mustSession(t, cfg)
	if _, err := ziti.Services.Find(&FilterOptions{}); err != nil {
		t.Fatal(err)
	}
}

func TestIsRateLimitedOnlyMatchesItsType(t *testing.T) {
	if IsRateLimited(nil) || IsRateLimited(errors.New("ziti SERVER_TOO_MANY_REQUESTS")) || IsRateLimited(NewNotFoundError("service", "delete", nil)) {
		t.Fatal("IsRateLimited matched a foreign error")
	}
	wrapped := wrapEdgeError(&RateLimitedError{Limiter: "SERVER_TOO_MANY_REQUESTS", Method: "DELETE", Path: "/services/x"}, "error deleting service '%s'", "x")
	if !IsRateLimited(wrapped) {
		t.Fatal("wrapEdgeError lost the rate limited type")
	}
}
