package proxy

import (
	"io"
	"testing"

	"github.com/caddyserver/caddy/v2"
)

func TestCaddyWebBackendProvision(t *testing.T) {
	previousWriter := loggingRequests
	previousRequests := middlewareRequests
	t.Cleanup(func() {
		SetCaddyLoggingWriter(previousWriter)
		middlewareRequests = previousRequests
	})

	for _, mode := range []string{"tui", "headless"} {
		t.Run(mode, func(t *testing.T) {
			SetCaddyLoggingWriter(nil)
			if mode == "tui" {
				SetCaddyLoggingWriter(testCaddyLogWriter{Writer: io.Discard})
			}
			backend, err := NewCaddyWebBackend(&CaddyWebBackendConfig{
				WebRoot:  t.TempDir(),
				ShrToken: "test-share",
			})
			if err != nil {
				t.Fatal(err)
			}
			// provision the real config without opening a Ziti listener or contacting a controller
			if err := caddy.Validate(backend.caddyCfg); err != nil {
				t.Fatalf("provision web backend: %v", err)
			}
		})
	}
}

type testCaddyLogWriter struct {
	io.Writer
}

func (testCaddyLogWriter) Close() error { return nil }
