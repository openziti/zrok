package controller

import (
	"path/filepath"
	"testing"

	"github.com/openziti/edge-api/rest_management_api_client"
	"github.com/openziti/zrok/controller/store"
	"github.com/openziti/zrok/controller/store/storetest"
	"github.com/openziti/zrok/controller/zrokEdgeSdk"
	"github.com/openziti/zrok/controller/zrokEdgeSdk/zitifake"
	"github.com/openziti/zrok/rest_model_zrok"
	"github.com/openziti/zrok/rest_server_zrok/operations/share"
	"github.com/openziti/zrok/sdk/golang/sdk"
	"github.com/stretchr/testify/require"
)

type unshareFixture struct {
	fake      *zitifake.Server
	handler   *unshareHandler
	edgeCalls int
	principal *rest_model_zrok.Principal
	envZId    string
}

func newUnshareFixture(t *testing.T) *unshareFixture {
	t.Helper()
	testStr, err := store.Open(&store.Config{Path: ":memory:", Type: "sqlite3"})
	require.NoError(t, err)
	return newUnshareFixtureOn(t, testStr)
}

func newUnshareFixtureOn(t *testing.T, testStr *store.Store) *unshareFixture {
	t.Helper()
	oldStr := str
	str = testStr
	t.Cleanup(func() {
		str = oldStr
		require.NoError(t, testStr.Close())
	})
	fake := zitifake.New()
	t.Cleanup(fake.Close)
	f := &unshareFixture{fake: fake, envZId: "env-zid"}
	f.handler = &unshareHandler{edge: func() (*rest_management_api_client.ZitiEdgeManagement, error) {
		f.edgeCalls++
		return fake.Edge(), nil
	}}
	trx, err := str.Begin()
	require.NoError(t, err)
	acctID, err := str.CreateAccount(&store.Account{Email: "unshare@example.com", Salt: "salt", Password: "password", Token: "account-token"}, trx)
	require.NoError(t, err)
	_, err = str.CreateEnvironment(acctID, &store.Environment{Description: "unshare", Host: "host", Address: "address", ZId: f.envZId}, trx)
	require.NoError(t, err)
	require.NoError(t, trx.Commit())
	f.principal = &rest_model_zrok.Principal{ID: int64(acctID), Email: "unshare@example.com"}
	return f
}

func (f *unshareFixture) unshare(token string) interface{} {
	return f.handler.Handle(share.UnshareParams{Body: share.UnshareBody{EnvZID: f.envZId, ShareToken: token}}, f.principal)
}

func TestUnshareMissingShareSkipsZiti(t *testing.T) {
	f := newUnshareFixture(t)
	before := f.fake.Requests()
	require.IsType(t, &share.UnshareNotFound{}, f.unshare("no-such-share"))
	require.Equal(t, before, f.fake.Requests())
	require.Zero(t, f.edgeCalls)
}

func TestUnshareExistingShareDeallocates(t *testing.T) {
	f := newUnshareFixture(t)
	edge := f.fake.Edge()
	shrToken := "existing-share"
	shrZId, err := zrokEdgeSdk.CreateShareService(f.envZId, shrToken, "config-zid", edge)
	require.NoError(t, err)
	tags := zrokEdgeSdk.ZrokShareTags(shrToken).SubTags
	require.NoError(t, zrokEdgeSdk.CreateServicePolicyBind(shrToken+"-bind", shrZId, f.envZId, tags, edge))
	require.NoError(t, zrokEdgeSdk.CreateServicePolicyDial(shrToken+"-dial", shrZId, []string{f.envZId}, tags, edge))

	trx, err := str.Begin()
	require.NoError(t, err)
	envs, err := str.FindEnvironmentsForAccount(int(f.principal.ID), trx)
	require.NoError(t, err)
	shrID, err := str.CreateShare(envs[0].Id, &store.Share{ZId: shrZId, Token: shrToken, ShareMode: string(sdk.PrivateShareMode), BackendMode: string(sdk.ProxyBackendMode), PermissionMode: store.OpenPermissionMode}, trx)
	require.NoError(t, err)
	require.NoError(t, trx.Commit())

	require.IsType(t, &share.UnshareOK{}, f.unshare(shrToken))
	require.Equal(t, 1, f.edgeCalls)
	_, policyDeletes, _, serviceDeletes := f.fake.Counts()
	require.Equal(t, 2, policyDeletes)
	require.Equal(t, 1, serviceDeletes)
	_, err = zrokEdgeSdk.FindShareService(shrZId, edge)
	require.Error(t, err)

	trx, err = str.Begin()
	require.NoError(t, err)
	defer func() { _ = trx.Rollback() }()
	_, err = str.FindShareWithToken(shrToken, trx)
	require.Error(t, err)
	shr, err := str.GetShare(shrID, trx)
	require.NoError(t, err)
	require.True(t, shr.Deleted)
}

// openV2Store opens a store that also carries the v2 mapping tables; the store is reopened after creating them so
// that its probe sees them.
func openV2Store(t *testing.T) *store.Store {
	t.Helper()
	cfg := &store.Config{Path: filepath.Join(t.TempDir(), "zrok.db"), Type: "sqlite3"}
	v1Str, err := store.Open(cfg)
	require.NoError(t, err)
	trx, err := v1Str.Begin()
	require.NoError(t, err)
	require.NoError(t, storetest.CreateV2Tables(trx))
	require.NoError(t, trx.Commit())
	require.NoError(t, v1Str.Close())
	v2Str, err := store.Open(cfg)
	require.NoError(t, err)
	return v2Str
}

func (f *unshareFixture) createMappedShare(t *testing.T, shrZId, shrToken string) int {
	t.Helper()
	trx, err := str.Begin()
	require.NoError(t, err)
	envs, err := str.FindEnvironmentsForAccount(int(f.principal.ID), trx)
	require.NoError(t, err)
	shrID, err := str.CreateShare(envs[0].Id, &store.Share{ZId: shrZId, Token: shrToken, ShareMode: string(sdk.PublicShareMode), BackendMode: string(sdk.ProxyBackendMode), PermissionMode: store.OpenPermissionMode}, trx)
	require.NoError(t, err)
	_, err = storetest.MapShare(trx, int(f.principal.ID), shrID, shrToken, shrToken)
	require.NoError(t, err)
	require.NoError(t, trx.Commit())
	return shrID
}

func requireNoLiveMappings(t *testing.T, shrID int, shrToken string) {
	t.Helper()
	trx, err := str.Begin()
	require.NoError(t, err)
	defer func() { _ = trx.Rollback() }()
	names, frontends, err := storetest.LiveMappings(trx, shrID, shrToken)
	require.NoError(t, err)
	require.Zero(t, names)
	require.Zero(t, frontends)
}

func TestUnshareReleasesV2Mappings(t *testing.T) {
	f := newUnshareFixtureOn(t, openV2Store(t))
	edge := f.fake.Edge()
	shrToken := "mapped-share"
	shrZId, err := zrokEdgeSdk.CreateShareService(f.envZId, shrToken, "config-zid", edge)
	require.NoError(t, err)
	shrID := f.createMappedShare(t, shrZId, shrToken)

	require.IsType(t, &share.UnshareOK{}, f.unshare(shrToken))
	requireNoLiveMappings(t, shrID, shrToken)
}

func TestDisableReleasesV2Mappings(t *testing.T) {
	f := newUnshareFixtureOn(t, openV2Store(t))
	shrToken := "disabled-share"
	shrID := f.createMappedShare(t, "disabled-zid", shrToken)

	trx, err := str.Begin()
	require.NoError(t, err)
	env, err := str.FindEnvironmentForAccount(f.envZId, int(f.principal.ID), trx)
	require.NoError(t, err)
	require.NoError(t, removeEnvironmentFromStore(env, trx))
	require.NoError(t, trx.Commit())
	requireNoLiveMappings(t, shrID, shrToken)
}
