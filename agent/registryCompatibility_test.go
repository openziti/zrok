package agent

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/michaelquigley/df/dd"
	"github.com/stretchr/testify/require"
)

func TestRegistryReadsV208WriterFixtures(t *testing.T) {
	retry, err := time.Parse(time.RFC3339Nano, "2026-10-28T12:34:56.123456789-04:00")
	require.NoError(t, err)
	// these files were generated with Registry.Save and the request structs from the v2.0.8 tag.
	// retain its null slices and omitted optional fields; hand-written populated fixtures missed this regression.
	cases := map[string]*Registry{
		"empty": {V: RegistryV},
		"public": {V: RegistryV, PublicShares: []*PublicShareRegistryEntry{{Request: &SharePublicRequest{
			Target: "http://localhost:8080", BackendMode: "proxy",
			NameSelections: []NameSelection{{NamespaceToken: "public", Name: "fixture-app"}},
		}}}},
		"private": {V: RegistryV,
			PrivateShares: []*PrivateShareRegistryEntry{{Request: &SharePrivateRequest{
				Target: "8081", BackendMode: "proxy", PrivateShareToken: "reserved-token",
			}}},
			PrivateAccesses: []*AccessRegistryEntry{{Request: &AccessPrivateRequest{
				ShareToken: "access-token", BindAddress: "127.0.0.1:8082",
			}}},
		},
		"all": {V: RegistryV,
			PublicShares: []*PublicShareRegistryEntry{{Request: &SharePublicRequest{
				Target: "8080", BackendMode: "proxy", NameSelections: []NameSelection{{NamespaceToken: "team", Name: "app"}},
				BasicAuth: []string{"fixture:password"}, Insecure: true, Closed: true,
				AccessGrants: []string{"fixture-account"}, OauthRefreshInterval: "3h0m0s",
			}, Failure: &FailureEntry{Count: 2, LastError: "fixture failure", NextRetry: retry}}},
			PrivateShares: []*PrivateShareRegistryEntry{{Request: &SharePrivateRequest{
				Target: "8081", BackendMode: "tcpTunnel", PrivateShareToken: "reserved-token",
				Insecure: true, Closed: true, AccessGrants: []string{"fixture-account"},
			}}},
			PrivateAccesses: []*AccessRegistryEntry{{Request: &AccessPrivateRequest{
				ShareToken: "access-token", BindAddress: "127.0.0.1:8082", AutoMode: true,
				AutoAddress: "127.0.0.1", AutoStartPort: 1024, AutoEndPort: 2048,
				ResponseHeaders: []string{"X-Fixture: true"},
			}, Failure: &FailureEntry{}}},
		},
	}
	for name, expected := range cases {
		t.Run(name, func(t *testing.T) {
			original, err := os.ReadFile(filepath.Join("testdata", "registry_v208", name+".json"))
			require.NoError(t, err)
			path := filepath.Join(t.TempDir(), "registry.json")
			require.NoError(t, os.WriteFile(path, original, 0600))
			loaded, err := LoadRegistry(path)
			require.NoError(t, err)
			require.Equal(t, expected, loaded)
			unchanged, err := os.ReadFile(path)
			require.NoError(t, err)
			require.Equal(t, original, unchanged, "loading must not rewrite the old file")

			require.NoError(t, loaded.Save(path))
			roundTrip, err := LoadRegistry(path)
			require.NoError(t, err)
			before, err := dd.Unbind(expected, registryBindingOptions())
			require.NoError(t, err)
			after, err := dd.Unbind(roundTrip, registryBindingOptions())
			require.NoError(t, err)
			require.Equal(t, before, after, "saving may normalize empty slices but must preserve requests and failures")
		})
	}
}

func TestRegistryCompatibilityKeepsInvalidInput(t *testing.T) {
	for name, raw := range map[string]string{
		"invalid json":     `{"v":"2",`,
		"unknown version":  `{"v":"3","public_shares":null}`,
		"wrong slice type": `{"v":"2","public_shares":[{"request":{"basic_auth":false}}]}`,
		"null entry":       `{"v":"2","public_shares":[null]}`,
		"null request":     `{"v":"2","public_shares":[{"request":null}]}`,
		"missing request":  `{"v":"2","public_shares":[{}]}`,
	} {
		t.Run(name, func(t *testing.T) {
			a := newLifecycleAgent(t)
			path, err := a.root.AgentRegistry()
			require.NoError(t, err)
			require.NoError(t, os.WriteFile(path, []byte(raw), 0600))
			require.Error(t, a.imported())
			require.ErrorContains(t, a.SaveRegistry(), "preserving unread file")
			data, err := os.ReadFile(path)
			require.NoError(t, err)
			require.Equal(t, raw, string(data))
		})
	}
}
