package controller

import (
	"testing"

	"github.com/openziti/zrok/v2/controller/store"
	"github.com/openziti/zrok/v2/rest_model_zrok"
	"github.com/openziti/zrok/v2/rest_server_zrok/operations/share"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnaccessRejectsGlobalFrontend(t *testing.T) {
	testStore := newTestControllerStore(t)
	principal, env := newTestPrincipalWithEnvironment(t, testStore, "owner@test.com", "owner-env")
	frontend := newTestGlobalFrontend(t, testStore, "global-frontend")

	resp := newUnaccessHandler().Handle(share.UnaccessParams{
		Body: share.UnaccessBody{
			EnvZID:        env.ZId,
			FrontendToken: frontend.Token,
			ShareToken:    "irrelevant",
		},
	}, principal)

	assert.IsType(t, &share.UnaccessNotFound{}, resp)
	assertFrontendDeletedState(t, testStore, frontend.Id, false)
}

func TestUnaccessRejectsForeignFrontend(t *testing.T) {
	testStore := newTestControllerStore(t)
	principal, env := newTestPrincipalWithEnvironment(t, testStore, "owner@test.com", "owner-env")
	foreignPrincipal, foreignEnv := newTestPrincipalWithEnvironment(t, testStore, "other@test.com", "other-env")
	_ = foreignPrincipal
	frontend := newTestEnvironmentFrontend(t, testStore, foreignEnv.Id, "foreign-frontend")

	resp := newUnaccessHandler().Handle(share.UnaccessParams{
		Body: share.UnaccessBody{
			EnvZID:        env.ZId,
			FrontendToken: frontend.Token,
			ShareToken:    "irrelevant",
		},
	}, principal)

	assert.IsType(t, &share.UnaccessNotFound{}, resp)
	assertFrontendDeletedState(t, testStore, frontend.Id, false)
}

func TestUnaccessRejectsUnownedEnvironment(t *testing.T) {
	testStore := newTestControllerStore(t)
	principal, _ := newTestPrincipalWithEnvironment(t, testStore, "owner@test.com", "owner-env")
	_, foreignEnv := newTestPrincipalWithEnvironment(t, testStore, "other@test.com", "other-env")

	resp := newUnaccessHandler().Handle(share.UnaccessParams{
		Body: share.UnaccessBody{
			EnvZID:        foreignEnv.ZId,
			FrontendToken: "irrelevant",
			ShareToken:    "irrelevant",
		},
	}, principal)

	assert.IsType(t, &share.UnaccessUnauthorized{}, resp)
}

func newTestControllerStore(t *testing.T) *store.Store {
	testStore, err := store.Open(&store.Config{Path: ":memory:", Type: "sqlite3"})
	require.NoError(t, err)

	previousStore := str
	str = testStore

	t.Cleanup(func() {
		str = previousStore
		require.NoError(t, testStore.Close())
	})

	return testStore
}

func newTestPrincipalWithEnvironment(t *testing.T, testStore *store.Store, email string, envZId string) (*rest_model_zrok.Principal, *store.Environment) {
	trx, err := testStore.Begin()
	require.NoError(t, err)

	accountId, err := testStore.CreateAccount(&store.Account{
		Email:    email,
		Salt:     "salt",
		Password: "password",
		Token:    email + "-token",
	}, trx)
	require.NoError(t, err)

	envId, err := testStore.CreateEnvironment(accountId, &store.Environment{
		Description: email + "-env",
		Host:        email + "-host",
		Address:     email + "-address",
		ZId:         envZId,
	}, trx)
	require.NoError(t, err)

	env, err := testStore.GetEnvironment(envId, trx)
	require.NoError(t, err)

	require.NoError(t, trx.Commit())

	return &rest_model_zrok.Principal{
		ID:    int64(accountId),
		Email: email,
		Token: email + "-token",
	}, env
}

func newTestGlobalFrontend(t *testing.T, testStore *store.Store, token string) *store.Frontend {
	trx, err := testStore.Begin()
	require.NoError(t, err)

	publicName := token + "-public"
	frontend := &store.Frontend{
		Token:          token,
		ZId:            token + "-zid",
		PublicName:     &publicName,
		Reserved:       true,
		PermissionMode: store.OpenPermissionMode,
	}
	frontendId, err := testStore.CreateGlobalFrontend(frontend, trx)
	require.NoError(t, err)

	storedFrontend, err := testStore.GetFrontend(frontendId, trx)
	require.NoError(t, err)

	require.NoError(t, trx.Commit())

	return storedFrontend
}

func newTestEnvironmentFrontend(t *testing.T, testStore *store.Store, envId int, token string) *store.Frontend {
	trx, err := testStore.Begin()
	require.NoError(t, err)

	frontend := &store.Frontend{
		Token:          token,
		ZId:            token + "-zid",
		PermissionMode: store.ClosedPermissionMode,
	}
	frontendId, err := testStore.CreateFrontend(envId, frontend, trx)
	require.NoError(t, err)

	storedFrontend, err := testStore.GetFrontend(frontendId, trx)
	require.NoError(t, err)

	require.NoError(t, trx.Commit())

	return storedFrontend
}

func assertFrontendDeletedState(t *testing.T, testStore *store.Store, frontendId int, expectedDeleted bool) {
	trx, err := testStore.Begin()
	require.NoError(t, err)
	defer func() { _ = trx.Rollback() }()

	frontend, err := testStore.GetFrontend(frontendId, trx)
	require.NoError(t, err)
	assert.Equal(t, expectedDeleted, frontend.Deleted)
}
