//go:build agent_smoke

package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestPublicResponseEndpointSchemes(t *testing.T) {
	lb := &loopback{path: "/" + randomHex(16), body: randomHex(32)}
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != lb.path || r.Header.Get("skip_zrok_interstitial") != "true" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(lb.body))
	})
	tlsServer := httptest.NewTLSServer(handler)
	t.Cleanup(tlsServer.Close)
	httpServer := httptest.NewServer(handler)
	t.Cleanup(httpServer.Close)
	// trust only the test server's certificate through the client's normal transport path.
	previous := http.DefaultTransport
	http.DefaultTransport = tlsServer.Client().Transport
	t.Cleanup(func() { http.DefaultTransport = previous })

	for _, tc := range []struct {
		name     string
		endpoint string
		want     string
	}{
		{"bare endpoint uses https", strings.TrimPrefix(tlsServer.URL, "https://"), tlsServer.URL + lb.path},
		{"explicit https", tlsServer.URL, tlsServer.URL + lb.path},
		{"explicit http", httpServer.URL, httpServer.URL + lb.path},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cfg := defaultConfig()
			cfg.pollInterval = 10 * time.Millisecond
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			got, err := awaitBody(ctx, cfg, tc.endpoint, lb)
			require.NoError(t, err)
			require.Equal(t, tc.want, got)
		})
	}
}
