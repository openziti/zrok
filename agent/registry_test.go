package agent

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/michaelquigley/df/dd"
	"github.com/stretchr/testify/require"
)

func TestRegistryV2CompatibilityAndSnapshots(t *testing.T) {
	path := filepath.Join(t.TempDir(), "registry.json")
	raw := `{"v":"2","private_accesses":[{"request":{"share_token":"share","bind_address":"127.0.0.1:8000","auto_start_port":1024,"auto_end_port":2048,"response_headers":["X-Test: true"]},"failure":{"count":2,"last_error":"offline","next_retry":"2026-10-28T12:34:56.123456789-04:00"}}],"public_shares":[{"request":{"target":"8080","backend_mode":"proxy","name_selections":[{"namespace_token":"team","name":"app"}],"basic_auth":["user:password"],"oauth_refresh_interval":"3h0m0s","access_grants":["account"]}}],"private_shares":[{"request":{"target":"8081","backend_mode":"proxy","private_share_token":"reserved","access_grants":["account"]}}]}`
	require.NoError(t, os.WriteFile(path, []byte(raw), 0600))
	loaded, err := LoadRegistry(path)
	require.NoError(t, err)
	require.Equal(t, "2", loaded.V, "a version 2 file is read as it is")
	require.Equal(t, "app", loaded.PublicShares[0].Request.NameSelections[0].Name)
	require.Equal(t, "user:password", loaded.PublicShares[0].Request.BasicAuth[0])
	require.EqualValues(t, 1024, loaded.PrivateAccesses[0].Request.AutoStartPort)
	require.Equal(t, "2026-10-28T12:34:56.123456789-04:00", loaded.PrivateAccesses[0].Failure.NextRetry.Format(time.RFC3339Nano))
	require.NoError(t, loaded.Save(path))
	roundTrip, err := LoadRegistry(path)
	require.NoError(t, err)
	before, err := dd.Unbind(loaded, registryBindingOptions())
	require.NoError(t, err)
	after, err := dd.Unbind(roundTrip, registryBindingOptions())
	require.NoError(t, err)
	require.Equal(t, before, after)
}

func TestRegistryVersions(t *testing.T) {
	t.Run("version 2 is saved as version 3 with its requests intact", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "registry.json")
		raw := `{"v":"2","public_shares":[{"request":{"target":"8080","backend_mode":"proxy","name_selections":[{"namespace_token":"public","name":"app"}],"basic_auth":["admin:secret"]}}]}`
		require.NoError(t, os.WriteFile(path, []byte(raw), 0600))
		loaded, err := LoadRegistry(path)
		require.NoError(t, err)
		require.NoError(t, loaded.Save(path))
		var saved map[string]any
		data, err := os.ReadFile(path)
		require.NoError(t, err)
		require.NoError(t, json.Unmarshal(data, &saved))
		require.Equal(t, "3", saved["v"])
		reloaded, err := LoadRegistry(path)
		require.NoError(t, err)
		require.Equal(t, "3", reloaded.V)
		require.Len(t, reloaded.PublicShares, 1)
		// saving normalizes absent slices to empty ones; the request is the same request.
		require.True(t, sameRequest(loaded.PublicShares[0].Request, reloaded.PublicShares[0].Request))
		require.Equal(t, "8080", reloaded.PublicShares[0].Request.Target)
		require.Equal(t, []NameSelection{{NamespaceToken: "public", Name: "app"}}, reloaded.PublicShares[0].Request.NameSelections)
		require.Equal(t, []string{"admin:secret"}, reloaded.PublicShares[0].Request.BasicAuth)
	})
	t.Run("version 3 loads", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "registry.json")
		require.NoError(t, os.WriteFile(path, []byte(`{"v":"3","public_shares":[{"request":{"target":"8080","backend_mode":"proxy","basic_auth":["admin:s3:cret"]}}]}`), 0600))
		loaded, err := LoadRegistry(path)
		require.NoError(t, err)
		require.Equal(t, "3", loaded.V)
		require.Equal(t, []string{"admin:s3:cret"}, loaded.PublicShares[0].Request.BasicAuth)
	})
	t.Run("an unknown version is refused", func(t *testing.T) {
		for _, raw := range []string{`{"v":"4"}`, `{"v":"1"}`, `{}`} {
			path := filepath.Join(t.TempDir(), "registry.json")
			require.NoError(t, os.WriteFile(path, []byte(raw), 0600))
			_, err := LoadRegistry(path)
			require.ErrorContains(t, err, "invalid registry version", raw)
			data, err := os.ReadFile(path)
			require.NoError(t, err)
			require.Equal(t, raw, string(data))
		}
	})
	t.Run("a registry with no entries is saved as version 3", func(t *testing.T) {
		for _, registry := range []*Registry{{}, {V: "2"}} {
			path := filepath.Join(t.TempDir(), "registry.json")
			require.NoError(t, registry.Save(path))
			var saved map[string]any
			data, err := os.ReadFile(path)
			require.NoError(t, err)
			require.NoError(t, json.Unmarshal(data, &saved))
			require.Equal(t, "3", saved["v"])
		}
	})
}
