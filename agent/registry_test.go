package agent

import (
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
	require.Equal(t, RegistryV, loaded.V)
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
