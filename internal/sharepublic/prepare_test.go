package sharepublic

import (
	"slices"
	"testing"
	"time"

	"github.com/openziti/zrok/v2/sdk/golang/sdk"
	"github.com/stretchr/testify/require"
)

func TestPrepareNameSelections(t *testing.T) {
	for _, tc := range []struct {
		name       string
		selections []sdk.NameSelection
		defaultNS  string
		want       []sdk.NameSelection
		invalid    bool
	}{
		{name: "nil named default", defaultNS: "team:app", want: []sdk.NameSelection{{NamespaceToken: "team", Name: "app"}}},
		{name: "empty named default", selections: []sdk.NameSelection{}, defaultNS: "team:app", want: []sdk.NameSelection{{NamespaceToken: "team", Name: "app"}}},
		{name: "nil namespace default", defaultNS: "team", want: []sdk.NameSelection{{NamespaceToken: "team"}}},
		{name: "empty namespace default", selections: []sdk.NameSelection{}, defaultNS: "team", want: []sdk.NameSelection{{NamespaceToken: "team"}}},
		{name: "empty explicit selection", selections: []sdk.NameSelection{{}}, defaultNS: "team", invalid: true},
		{name: "explicit name without namespace", selections: []sdk.NameSelection{{Name: "app"}}, defaultNS: "team", invalid: true},
		{name: "invalid selection after valid selection", selections: []sdk.NameSelection{{NamespaceToken: "team"}, {}}, defaultNS: "other", invalid: true},
		{name: "blank namespace", selections: []sdk.NameSelection{{NamespaceToken: " \t", Name: "app"}}, defaultNS: "team", invalid: true},
		{name: "empty default", invalid: true},
		{name: "default name without namespace", defaultNS: ":app", invalid: true},
		{name: "explicit namespace", selections: []sdk.NameSelection{{NamespaceToken: "explicit"}}, want: []sdk.NameSelection{{NamespaceToken: "explicit"}}},
		{name: "explicit names and namespaces", selections: []sdk.NameSelection{{NamespaceToken: "team", Name: "app"}, {NamespaceToken: "other"}}, defaultNS: "fallback:ignored", want: []sdk.NameSelection{{NamespaceToken: "team", Name: "app"}, {NamespaceToken: "other"}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			in := sdk.ShareRequest{BackendMode: sdk.ProxyBackendMode, Target: "8080", NameSelections: tc.selections}
			original := in
			original.NameSelections = slices.Clone(in.NameSelections)
			out, err := Prepare(in, tc.defaultNS)
			require.Equal(t, original, in)
			if tc.invalid {
				require.ErrorContains(t, err, "namespace")
				require.Nil(t, out)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tc.want, out.NameSelections)
			out.NameSelections[0] = sdk.NameSelection{NamespaceToken: "changed", Name: "changed"}
			require.Equal(t, original, in)
		})
	}
}

func TestTargets(t *testing.T) {
	for _, tc := range []struct{ mode, target, want string }{
		{"proxy", "443", "https://127.0.0.1:443"}, {"proxy", "8080", "http://127.0.0.1:8080"},
		{"proxy", "example.com/path", "http://example.com/path"}, {"proxy", "https://example.com", "https://example.com"},
		{"web", "./web", "./web"}, {"caddy", "Caddyfile", "Caddyfile"}, {"drive", "./files", "./files"},
	} {
		t.Run(tc.mode+tc.target, func(t *testing.T) {
			target, err := Target(tc.mode, []string{tc.target})
			require.NoError(t, err)
			require.Equal(t, tc.want, target)
		})
	}
	for _, mode := range []string{"tcpTunnel", "udpTunnel", "socks", "unknown"} {
		_, err := Target(mode, []string{"8080"})
		require.Error(t, err)
	}
	for _, args := range [][]string{nil, {"a", "b"}} {
		_, err := Target("proxy", args)
		require.Error(t, err)
	}
	for _, target := range []string{"0", "-1", "65536", "http://["} {
		_, err := Target("proxy", []string{target})
		require.Error(t, err)
	}
}

func TestPrepareValidationAndOwnership(t *testing.T) {
	in := sdk.ShareRequest{BackendMode: sdk.ProxyBackendMode, Target: "8080", BasicAuth: []string{" user : password "}, AccessGrants: []string{"account"}, NameSelections: []sdk.NameSelection{{NamespaceToken: "team", Name: "app"}}, PermissionMode: sdk.ClosedPermissionMode}
	out, err := Prepare(in, "other")
	require.NoError(t, err)
	require.Equal(t, sdk.PublicShareMode, out.ShareMode)
	require.Equal(t, "http://127.0.0.1:8080", out.Target)
	require.Equal(t, in.NameSelections, out.NameSelections)
	out.BasicAuth[0] = "changed"
	out.AccessGrants[0] = "changed"
	out.NameSelections[0].Name = "changed"
	require.Equal(t, " user : password ", in.BasicAuth[0])
	require.Equal(t, "account", in.AccessGrants[0])
	require.Equal(t, "app", in.NameSelections[0].Name)
	for _, selections := range [][]sdk.NameSelection{nil, {}} {
		out, err := Prepare(sdk.ShareRequest{BackendMode: sdk.ProxyBackendMode, Target: "443", NameSelections: selections}, "team")
		require.NoError(t, err)
		require.Equal(t, []sdk.NameSelection{{NamespaceToken: "team"}}, out.NameSelections)
	}
	in.OauthProvider = "provider"
	_, err = Prepare(in, "public")
	require.ErrorContains(t, err, "mutually exclusive")
	in.BasicAuth = nil
	in.OauthEmailAddressPatterns = []string{"["}
	_, err = Prepare(in, "public")
	require.ErrorContains(t, err, "glob")
	in.OauthEmailAddressPatterns = []string{"*@example.com"}
	in.OauthRefreshInterval = -time.Second
	out, err = Prepare(in, "public")
	require.NoError(t, err)
	require.Equal(t, -time.Second, out.OauthRefreshInterval)
	out.OauthEmailAddressPatterns[0] = "changed"
	require.Equal(t, "*@example.com", in.OauthEmailAddressPatterns[0])
	in.OauthProvider = ""
	in.BasicAuth = []string{"invalid"}
	_, err = Prepare(in, "public")
	require.ErrorContains(t, err, "username:password")
}
