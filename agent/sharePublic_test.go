package agent

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/openziti/zrok/v2/environment/env_core"
	"github.com/openziti/zrok/v2/sdk/golang/sdk"
	"github.com/stretchr/testify/require"
)

type preparationRoot struct {
	env_core.Root
	meta *env_core.Metadata
}

func (r *preparationRoot) Metadata() *env_core.Metadata { return r.meta }

func TestPublicPreparationFreshDefault(t *testing.T) {
	t.Setenv("ZROK2_DEFAULT_NAMESPACE", "")
	root := &preparationRoot{meta: &env_core.Metadata{RootPath: t.TempDir()}}
	a := &Agent{root: root}
	path := filepath.Join(root.meta.RootPath, "config.json")
	req := &SharePublicRequest{BackendMode: "proxy", Target: "8080", OauthProvider: "provider", OauthEmailDomains: []string{"*@example.com"}, Closed: true}
	require.NoError(t, os.WriteFile(path, []byte(`{"default_namespace":"team-a"}`), 0600))
	first, err := a.preparePublicRequest(req)
	require.NoError(t, err)
	require.Equal(t, 3*time.Hour, first.OauthRefreshInterval)
	require.NoError(t, os.WriteFile(path, []byte(`{"default_namespace":"team-b"}`), 0600))
	second, err := a.preparePublicRequest(req)
	require.NoError(t, err)
	require.Equal(t, "team-a", first.NameSelections[0].NamespaceToken)
	require.Equal(t, "team-b", second.NameSelections[0].NamespaceToken)
	require.Nil(t, req.NameSelections)
	require.Empty(t, req.OauthRefreshInterval)
	require.Equal(t, "8080", req.Target)
	require.Same(t, root, a.root)
	req.NameSelections = []NameSelection{{NamespaceToken: "cli-default"}, {NamespaceToken: "team", Name: "app"}}
	// explicit selections do not need a readable config.
	require.NoError(t, os.WriteFile(path, []byte("{"), 0600))
	explicit, err := a.preparePublicRequest(req)
	require.NoError(t, err)
	require.Equal(t, []sdk.NameSelection{{NamespaceToken: "cli-default"}, {NamespaceToken: "team", Name: "app"}}, explicit.NameSelections)
	req.NameSelections = nil
	_, _, err = a.SharePublic(req)
	require.Error(t, err) // fails before launching a child
	t.Setenv("ZROK2_DEFAULT_NAMESPACE", "process-default")
	override, err := a.preparePublicRequest(req)
	require.NoError(t, err)
	require.Equal(t, "process-default", override.NameSelections[0].NamespaceToken)
}

func TestPublicPreparationValidation(t *testing.T) {
	a := &Agent{}
	explicit := []NameSelection{{NamespaceToken: "team", Name: "app"}}
	for _, req := range []*SharePublicRequest{
		nil,
		{BackendMode: "tcpTunnel", Target: "8080", NameSelections: explicit},
		{BackendMode: "proxy", Target: "8080", NameSelections: explicit, OauthRefreshInterval: "invalid"},
		{BackendMode: "proxy", Target: "8080", NameSelections: explicit, BasicAuth: []string{"u:p"}, OauthProvider: "provider"},
		{BackendMode: "proxy", Target: "8080", NameSelections: explicit, OauthProvider: "provider", OauthEmailDomains: []string{"["}},
	} {
		_, _, err := a.SharePublic(req)
		require.Error(t, err)
	}
	for _, mode := range []string{"web", "drive", "caddy"} {
		req := &SharePublicRequest{BackendMode: mode, Target: "relative", NameSelections: explicit}
		out, err := a.preparePublicRequest(req)
		require.NoError(t, err)
		want, err := filepath.Abs("relative")
		require.NoError(t, err)
		require.Equal(t, want, out.Target)
		require.Equal(t, "relative", req.Target)
	}
}

func TestAgentPublicRequestMatchesCLIContract(t *testing.T) {
	a := &Agent{}
	req := &SharePublicRequest{BackendMode: "proxy", Target: "443", NameSelections: []NameSelection{{NamespaceToken: "team", Name: "app"}, {NamespaceToken: "other"}}, BasicAuth: []string{"u:p"}, AccessGrants: []string{"account"}, Closed: true}
	out, err := a.preparePublicRequest(req)
	require.NoError(t, err)
	require.Equal(t, &sdk.ShareRequest{BackendMode: sdk.ProxyBackendMode, ShareMode: sdk.PublicShareMode, Target: "https://127.0.0.1:443", NameSelections: []sdk.NameSelection{{NamespaceToken: "team", Name: "app"}, {NamespaceToken: "other"}}, BasicAuth: []string{"u:p"}, AccessGrants: []string{"account"}, PermissionMode: sdk.ClosedPermissionMode}, out)
	out.BasicAuth[0] = "changed"
	out.AccessGrants[0] = "changed"
	out.NameSelections[0].Name = "changed"
	require.Equal(t, "u:p", req.BasicAuth[0])
	require.Equal(t, "account", req.AccessGrants[0])
	require.Equal(t, "app", req.NameSelections[0].Name)
}

func TestAgentPublicNameSelectionPreparation(t *testing.T) {
	t.Setenv("ZROK2_DEFAULT_NAMESPACE", "team:app")
	a := &Agent{}
	for _, selections := range [][]NameSelection{nil, {}} {
		req := &SharePublicRequest{BackendMode: "proxy", Target: "8080", NameSelections: selections}
		out, err := a.preparePublicRequest(req)
		require.NoError(t, err)
		require.Equal(t, []sdk.NameSelection{{NamespaceToken: "team", Name: "app"}}, out.NameSelections)
		require.Equal(t, selections, req.NameSelections)
	}
	for _, selections := range [][]NameSelection{{{}}, {{Name: "app"}}, {{NamespaceToken: "team"}, {}}} {
		req := &SharePublicRequest{BackendMode: "proxy", Target: "8080", NameSelections: selections}
		out, err := a.preparePublicRequest(req)
		require.ErrorContains(t, err, "namespace")
		require.Nil(t, out)
		require.Equal(t, selections, req.NameSelections)
	}
}
