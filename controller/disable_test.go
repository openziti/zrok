package controller

import (
	"testing"

	"github.com/openziti/zrok/v2/controller/automation/zitifake"
	"github.com/openziti/zrok/v2/controller/store"
	"github.com/openziti/zrok/v2/rest_model_zrok"
	"github.com/openziti/zrok/v2/rest_server_zrok/operations/admin"
	"github.com/openziti/zrok/v2/rest_server_zrok/operations/environment"
	"github.com/stretchr/testify/require"
)

type disableFixture struct {
	*shareCreateFixture
	envID int
}

// setupDisableFixture adds a share and a frontend to the share-create fixture's environment 'env-zid'.
// the fake holds no ziti objects for either, so their teardown deletes find nothing.
func setupDisableFixture(t *testing.T) *disableFixture {
	t.Helper()
	f := setupShareCreateFixture(t)
	trx, err := str.Begin()
	require.NoError(t, err)
	envs, err := str.FindEnvironmentsForAccount(int(f.principal.ID), trx)
	require.NoError(t, err)
	require.Len(t, envs, 1)
	envID := envs[0].Id
	_, err = str.CreateShare(envID, &store.Share{ZId: "share-zid", Token: "share-token", ShareMode: "private", BackendMode: "proxy", PermissionMode: store.OpenPermissionMode}, trx)
	require.NoError(t, err)
	_, err = str.CreateFrontend(envID, &store.Frontend{Token: "frontend-token", ZId: "env-zid", PermissionMode: store.OpenPermissionMode}, trx)
	require.NoError(t, err)
	require.NoError(t, trx.Commit())
	return &disableFixture{shareCreateFixture: f, envID: envID}
}

func (f *disableFixture) disable() interface{} {
	return newDisableHandler().Handle(environment.DisableParams{Body: environment.DisableBody{Identity: "env-zid"}}, f.principal)
}

func (f *disableFixture) requireEnvironmentRemoved(t *testing.T, removed bool) {
	t.Helper()
	trx, err := str.Begin()
	require.NoError(t, err)
	defer func() { _ = trx.Rollback() }()
	env, err := str.GetEnvironment(f.envID, trx)
	require.NoError(t, err)
	require.Equal(t, removed, env.Deleted)
	shrs, err := str.FindSharesForEnvironment(f.envID, trx)
	require.NoError(t, err)
	fes, err := str.FindFrontendsForEnvironment(f.envID, trx)
	require.NoError(t, err)
	if removed {
		require.Empty(t, shrs)
		require.Empty(t, fes)
	} else {
		require.Len(t, shrs, 1)
		require.Len(t, fes, 1)
	}
}

func TestDisableWithAbsentIdentitySucceeds(t *testing.T) {
	f := setupDisableFixture(t)

	resp := f.disable()

	require.IsType(t, &environment.DisableOK{}, resp)
	f.requireEnvironmentRemoved(t, true)
	require.Contains(t, f.logs.String(), "identity 'env-zid' for environment already deleted")
}

func TestDisableDeletesPresentIdentity(t *testing.T) {
	f := setupDisableFixture(t)
	f.fake.SeedWithID(zitifake.Identities, "env-zid", "env-zid", nil)
	f.fake.SeedWithID(zitifake.EdgeRouterPolicies, "erp-env", "env-zid", nil)

	resp := f.disable()

	require.IsType(t, &environment.DisableOK{}, resp)
	require.False(t, f.fake.Has(zitifake.Identities, "env-zid"))
	require.False(t, f.fake.Has(zitifake.EdgeRouterPolicies, "erp-env"))
	f.requireEnvironmentRemoved(t, true)
	require.NotContains(t, f.logs.String(), "already deleted")
}

func TestDisableOtherFailuresStillFail(t *testing.T) {
	f := setupDisableFixture(t)
	f.fake.SeedWithID(zitifake.Identities, "env-zid", "env-zid", nil)
	f.fake.RejectOperations(true)

	resp := f.disable()

	require.IsType(t, &environment.DisableInternalServerError{}, resp)
	f.fake.RejectOperations(false)
	require.True(t, f.fake.Has(zitifake.Identities, "env-zid"))
	f.requireEnvironmentRemoved(t, false)
}

func TestDeleteAccountWithAbsentIdentitySucceeds(t *testing.T) {
	f := setupDisableFixture(t)

	resp := newDeleteAccountHandler().Handle(admin.DeleteAccountParams{Body: admin.DeleteAccountBody{Email: "owner@example.com"}}, &rest_model_zrok.Principal{Admin: true})

	require.IsType(t, &admin.DeleteAccountOK{}, resp)
	f.requireEnvironmentRemoved(t, true)
	trx, err := str.Begin()
	require.NoError(t, err)
	defer func() { _ = trx.Rollback() }()
	_, err = str.FindAccountWithEmail("owner@example.com", trx)
	require.Error(t, err)
	envs, err := str.FindEnvironmentsForAccount(int(f.principal.ID), trx)
	require.NoError(t, err)
	require.Empty(t, envs)
}

func TestDeleteIdentityWithAbsentIdentitySucceeds(t *testing.T) {
	f := setupShareCreateFixture(t)

	resp := newDeleteIdentityHandler().Handle(admin.DeleteIdentityParams{Body: admin.DeleteIdentityBody{ZID: "env-zid"}}, &rest_model_zrok.Principal{Admin: true})

	require.IsType(t, &admin.DeleteIdentityOK{}, resp)
	require.Contains(t, f.logs.String(), "identity 'env-zid' already deleted")
}
