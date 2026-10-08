package main

import (
	"testing"

	"github.com/openziti/zrok/v2/rest_model_zrok"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var deleteShareFixtureShares = []*rest_model_zrok.ShareSummary{
	{ShareToken: "abc123", EnvZID: "env-1"},
	{ShareToken: "def456", EnvZID: "env-1"},
	{ShareToken: "ghi789", EnvZID: "env-2"},
}

var deleteShareFixtureNames = []*rest_model_zrok.Name{
	{Name: "myapp", NamespaceToken: "public", NamespaceName: "share.example.com", ShareToken: "abc123"},
	{Name: "idle", NamespaceToken: "public", NamespaceName: "share.example.com", Reserved: true},
	{Name: "twice", NamespaceToken: "public", NamespaceName: "share.example.com", ShareToken: "abc123"},
	{Name: "twice", NamespaceToken: "internal", NamespaceName: "internal.example.com", ShareToken: "def456"},
	{Name: "both", NamespaceToken: "public", NamespaceName: "share.example.com", ShareToken: "def456"},
	{Name: "both", NamespaceToken: "internal", NamespaceName: "internal.example.com", ShareToken: "def456"},
	{Name: "elsewhere", NamespaceToken: "public", NamespaceName: "share.example.com", ShareToken: "ghi789"},
}

func TestResolveShareTokenOrNameAcceptsShareToken(t *testing.T) {
	shrToken, err := resolveShareTokenOrName("def456", deleteShareFixtureShares, nil)
	require.NoError(t, err)
	assert.Equal(t, "def456", shrToken)
}

func TestResolveShareTokenOrNameResolvesName(t *testing.T) {
	for _, arg := range []string{"myapp", "public:myapp", "myapp.share.example.com"} {
		shrToken, err := resolveShareTokenOrName(arg, deleteShareFixtureShares, deleteShareFixtureNames)
		require.NoError(t, err, arg)
		assert.Equal(t, "abc123", shrToken, arg)
	}
}

func TestResolveShareTokenOrNameNameHeldInTwoNamespacesBySameShare(t *testing.T) {
	shrToken, err := resolveShareTokenOrName("both", deleteShareFixtureShares, deleteShareFixtureNames)
	require.NoError(t, err)
	assert.Equal(t, "def456", shrToken)
}

func TestResolveShareTokenOrNameAmbiguousNameIsPermanent(t *testing.T) {
	_, err := resolveShareTokenOrName("twice", deleteShareFixtureShares, deleteShareFixtureNames)
	require.Error(t, err)
	assert.Equal(t, exitPermanent, classifyFailure(err).exitCode)
	assert.Contains(t, err.Error(), "'public:twice' (share 'abc123')")
	assert.Contains(t, err.Error(), "'internal:twice' (share 'def456')")

	shrToken, err := resolveShareTokenOrName("internal:twice", deleteShareFixtureShares, deleteShareFixtureNames)
	require.NoError(t, err)
	assert.Equal(t, "def456", shrToken)
}

func TestResolveShareTokenOrNameNameWithoutLiveShareIsPermanent(t *testing.T) {
	_, err := resolveShareTokenOrName("idle", deleteShareFixtureShares, deleteShareFixtureNames)
	require.Error(t, err)
	assert.Equal(t, "name 'idle' is not held by a live share", err.Error())
	assert.Equal(t, exitPermanent, classifyFailure(err).exitCode)
}

func TestResolveShareTokenOrNameUnknownIsPermanent(t *testing.T) {
	_, err := resolveShareTokenOrName("nothing", deleteShareFixtureShares, deleteShareFixtureNames)
	require.Error(t, err)
	assert.Equal(t, "no share token or name 'nothing' found for this account", err.Error())
	assert.Equal(t, exitPermanent, classifyFailure(err).exitCode)
}

func TestResolveUnshareEnvZIdUsesResolvedSharesEnvironmentForName(t *testing.T) {
	shrToken, err := resolveShareTokenOrName("elsewhere", deleteShareFixtureShares, deleteShareFixtureNames)
	require.NoError(t, err)
	require.Equal(t, "ghi789", shrToken)
	assert.Equal(t, "env-2", resolveUnshareEnvZId("elsewhere", shrToken, "env-1", "", deleteShareFixtureShares))
}

func TestResolveUnshareEnvZIdOverrideWinsForName(t *testing.T) {
	assert.Equal(t, "env-override", resolveUnshareEnvZId("elsewhere", "ghi789", "env-1", "env-override", deleteShareFixtureShares))
}

func TestResolveUnshareEnvZIdUsesCurrentEnvironmentForToken(t *testing.T) {
	assert.Equal(t, "env-1", resolveUnshareEnvZId("ghi789", "ghi789", "env-1", "", deleteShareFixtureShares))
}
