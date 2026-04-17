package dynamicProxy

import (
	"testing"

	"github.com/openziti/zrok/v2/controller/dynamicProxyController"
)

// withTestMappings swaps in a populated oauthRouter+mappings for the duration of the
// test, so resolveCookieDomain has something to look up. Restores the previous router
// on cleanup.
func withTestMappings(t *testing.T, hosts ...string) {
	t.Helper()
	m := newMappings()
	for _, h := range hosts {
		m.nameMap[h] = &dynamicProxyController.FrontendMapping{Name: h, ShareToken: "sh-" + h}
	}
	prev := globalOAuthRouter
	globalOAuthRouter = &oauthRouter{mappings: m}
	t.Cleanup(func() { globalOAuthRouter = prev })
}

func TestGetNamespaceForHost(t *testing.T) {
	m := newMappings()
	m.nameMap["alice.customer.io"] = &dynamicProxyController.FrontendMapping{Name: "alice.customer.io", ShareToken: "sh1"}
	m.nameMap["bob.example.co.uk"] = &dynamicProxyController.FrontendMapping{Name: "bob.example.co.uk", ShareToken: "sh2"}
	m.nameMap["weird."] = &dynamicProxyController.FrontendMapping{Name: "weird.", ShareToken: "sh3"}
	m.nameMap[".leading"] = &dynamicProxyController.FrontendMapping{Name: ".leading", ShareToken: "sh4"}
	m.nameMap["nodots"] = &dynamicProxyController.FrontendMapping{Name: "nodots", ShareToken: "sh5"}

	tests := []struct {
		name     string
		host     string
		wantNs   string
		wantOk   bool
	}{
		{"simple namespace", "alice.customer.io", "customer.io", true},
		{"multi-label namespace", "bob.example.co.uk", "example.co.uk", true},
		{"unmapped host rejected", "stranger.unknown.io", "", false},
		{"empty host rejected", "", "", false},
		{"name ending in dot rejected", "weird.", "", false},
		{"leading dot rejected", ".leading", "", false},
		{"name with no dot rejected", "nodots", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ns, ok := m.getNamespaceForHost(tt.host)
			if ok != tt.wantOk {
				t.Fatalf("ok = %v, want %v", ok, tt.wantOk)
			}
			if ns != tt.wantNs {
				t.Fatalf("ns = %q, want %q", ns, tt.wantNs)
			}
		})
	}
}

func TestResolveCookieDomain(t *testing.T) {
	withTestMappings(t, "alice.customer.io", "bob.example.co.uk")

	tests := []struct {
		name   string
		input  string
		wantNs string
		wantOk bool
	}{
		{"bare host", "alice.customer.io", "customer.io", true},
		{"host with port", "alice.customer.io:8080", "customer.io", true},
		{"host with path", "alice.customer.io/some/path", "customer.io", true},
		{"host with port and path", "alice.customer.io:8080/some/path", "customer.io", true},
		{"multi-label tld with port", "bob.example.co.uk:443", "example.co.uk", true},
		{"unmapped host rejected", "stranger.unknown.io:8080", "", false},
		{"empty rejected", "", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ns, ok := resolveCookieDomain(tt.input)
			if ok != tt.wantOk {
				t.Fatalf("ok = %v, want %v", ok, tt.wantOk)
			}
			if ns != tt.wantNs {
				t.Fatalf("ns = %q, want %q", ns, tt.wantNs)
			}
		})
	}
}

func TestResolveCookieDomainNilRouter(t *testing.T) {
	prev := globalOAuthRouter
	globalOAuthRouter = nil
	t.Cleanup(func() { globalOAuthRouter = prev })

	if _, ok := resolveCookieDomain("alice.customer.io"); ok {
		t.Fatal("expected rejection when globalOAuthRouter is nil")
	}
}
