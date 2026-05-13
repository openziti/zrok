package controller

import (
	"testing"

	controllerConfig "github.com/openziti/zrok/v2/controller/config"
	"github.com/openziti/zrok/v2/controller/store"
	"github.com/openziti/zrok/v2/rest_model_zrok"
	shareops "github.com/openziti/zrok/v2/rest_server_zrok/operations/share"
	"github.com/stretchr/testify/require"
)

type shareNameFixture struct {
	accountID     int
	environmentID int
	namespaceID   int
	shareID       int
	nameID        int
	principal     *rest_model_zrok.Principal
}

func setupShareNameFixture(t *testing.T, reserved bool) *shareNameFixture {
	t.Helper()

	prevStore := str
	prevCfg := cfg
	t.Cleanup(func() {
		str = prevStore
		cfg = prevCfg
	})
	cfg = controllerConfig.DefaultConfig()

	var err error
	str, err = store.Open(&store.Config{Path: ":memory:", Type: "sqlite3"})
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, str.Close())
	})

	trx, err := str.Begin()
	require.NoError(t, err)

	accountID, err := str.CreateAccount(&store.Account{
		Email:    "test@example.com",
		Salt:     "salt",
		Password: "password",
		Token:    "acct-token",
	}, trx)
	require.NoError(t, err)

	environmentID, err := str.CreateEnvironment(accountID, &store.Environment{
		Description: "test environment",
		Host:        "host",
		Address:     "address",
		ZId:         "env-zid",
	}, trx)
	require.NoError(t, err)

	namespaceID, err := str.CreateNamespace(&store.Namespace{
		Token:       "public",
		Name:        "example.com",
		Description: "public namespace",
		Open:        true,
	}, trx)
	require.NoError(t, err)

	shareID, err := str.CreateShare(environmentID, &store.Share{
		ZId:            "share-zid",
		Token:          "share-token",
		ShareMode:      "public",
		BackendMode:    "proxy",
		PermissionMode: store.OpenPermissionMode,
	}, trx)
	require.NoError(t, err)

	nameID, err := str.CreateName(&store.Name{
		NamespaceId: namespaceID,
		Name:        "demo",
		AccountId:   accountID,
		Reserved:    reserved,
	}, trx)
	require.NoError(t, err)

	_, err = str.CreateShareNameMapping(&store.ShareNameMapping{
		ShareId: shareID,
		NameId:  nameID,
	}, trx)
	require.NoError(t, err)

	require.NoError(t, trx.Commit())

	return &shareNameFixture{
		accountID:     accountID,
		environmentID: environmentID,
		namespaceID:   namespaceID,
		shareID:       shareID,
		nameID:        nameID,
		principal: &rest_model_zrok.Principal{
			ID:    int64(accountID),
			Email: "test@example.com",
		},
	}
}

func attachDynamicFrontend(t *testing.T, fixture *shareNameFixture, frontendToken string) {
	t.Helper()

	trx, err := str.Begin()
	require.NoError(t, err)

	frontendID, err := str.CreateGlobalFrontend(&store.Frontend{
		Token:          frontendToken,
		ZId:            frontendToken + "-zid",
		Dynamic:        true,
		PermissionMode: store.OpenPermissionMode,
	}, trx)
	require.NoError(t, err)

	_, err = str.CreateNamespaceFrontendMapping(fixture.namespaceID, frontendID, false, trx)
	require.NoError(t, err)
	require.NoError(t, trx.Commit())
}

func TestDeleteShareNameConflictsWhenAttachedToActiveShare(t *testing.T) {
	fixture := setupShareNameFixture(t, true)

	handler := newDeleteShareNameHandler()
	resp := handler.Handle(shareops.DeleteShareNameParams{
		Body: shareops.DeleteShareNameBody{
			NamespaceToken: "public",
			Name:           "demo",
		},
	}, fixture.principal)

	conflict, ok := resp.(*shareops.DeleteShareNameConflict)
	require.True(t, ok)
	require.Contains(t, string(conflict.Payload), "share-token")

	trx, err := str.Begin()
	require.NoError(t, err)
	defer func() { _ = trx.Rollback() }()

	name, err := str.GetName(fixture.nameID, trx)
	require.NoError(t, err)
	require.False(t, name.Deleted)

	mappings, err := str.FindShareNameMappingsByNameId(fixture.nameID, trx)
	require.NoError(t, err)
	require.Len(t, mappings, 1)
}

func TestDeleteShareNameCleansStaleMappingForDeletedShare(t *testing.T) {
	fixture := setupShareNameFixture(t, true)

	trx, err := str.Begin()
	require.NoError(t, err)
	require.NoError(t, str.DeleteShare(fixture.shareID, trx))
	require.NoError(t, trx.Commit())

	handler := newDeleteShareNameHandler()
	resp := handler.Handle(shareops.DeleteShareNameParams{
		Body: shareops.DeleteShareNameBody{
			NamespaceToken: "public",
			Name:           "demo",
		},
	}, fixture.principal)

	_, ok := resp.(*shareops.DeleteShareNameOK)
	require.True(t, ok)

	trx, err = str.Begin()
	require.NoError(t, err)
	defer func() { _ = trx.Rollback() }()

	var deleted bool
	err = trx.QueryRow("select deleted from names where id = $1", fixture.nameID).Scan(&deleted)
	require.NoError(t, err)
	require.True(t, deleted)

	mappings, err := str.FindShareNameMappingsByNameId(fixture.nameID, trx)
	require.NoError(t, err)
	require.Empty(t, mappings)
}

func TestCleanupShareNameMappingsHandlesDeletedReservedName(t *testing.T) {
	fixture := setupShareNameFixture(t, true)

	trx, err := str.Begin()
	require.NoError(t, err)
	require.NoError(t, str.DeleteName(fixture.nameID, trx))

	handler := newUnshareHandler()
	require.NoError(t, handler.cleanupShareNameMappings(fixture.shareID, trx))
	require.NoError(t, trx.Commit())

	trx, err = str.Begin()
	require.NoError(t, err)
	defer func() { _ = trx.Rollback() }()

	var deleted bool
	err = trx.QueryRow("select deleted from names where id = $1", fixture.nameID).Scan(&deleted)
	require.NoError(t, err)
	require.True(t, deleted)

	mappings, err := str.FindShareNameMappingsByNameId(fixture.nameID, trx)
	require.NoError(t, err)
	require.Empty(t, mappings)
}

func TestCleanupShareNameMappingsDeletesDynamicName(t *testing.T) {
	fixture := setupShareNameFixture(t, false)

	trx, err := str.Begin()
	require.NoError(t, err)

	handler := newUnshareHandler()
	require.NoError(t, handler.cleanupShareNameMappings(fixture.shareID, trx))
	require.NoError(t, trx.Commit())

	trx, err = str.Begin()
	require.NoError(t, err)
	defer func() { _ = trx.Rollback() }()

	var deleted bool
	err = trx.QueryRow("select deleted from names where id = $1", fixture.nameID).Scan(&deleted)
	require.NoError(t, err)
	require.True(t, deleted)

	mappings, err := str.FindShareNameMappingsByNameId(fixture.nameID, trx)
	require.NoError(t, err)
	require.Empty(t, mappings)
}

func TestCreateShareNameHealsStaleFrontendMappingForDeletedShare(t *testing.T) {
	fixture := setupShareNameFixture(t, true)
	attachDynamicFrontend(t, fixture, "dynamic-fe")

	trx, err := str.Begin()
	require.NoError(t, err)
	require.NoError(t, str.DeleteName(fixture.nameID, trx))
	require.NoError(t, str.DeleteShare(fixture.shareID, trx))
	_, err = trx.Exec("insert into frontend_mappings (frontend_token, name, share_token) values ($1, $2, $3)", "dynamic-fe", "demo.example.com", "share-token")
	require.NoError(t, err)
	require.NoError(t, trx.Commit())

	handler := newCreateShareNameHandler()
	resp := handler.Handle(shareops.CreateShareNameParams{
		Body: shareops.CreateShareNameBody{
			NamespaceToken: "public",
			Name:           "demo",
		},
	}, fixture.principal)

	_, ok := resp.(*shareops.CreateShareNameCreated)
	require.True(t, ok)

	trx, err = str.Begin()
	require.NoError(t, err)
	defer func() { _ = trx.Rollback() }()

	mapping, err := str.FindFrontendMappingByFrontendTokenAndNameWithShareState("dynamic-fe", "demo.example.com", trx)
	require.Error(t, err)
	require.Nil(t, mapping)

	name, err := str.FindNameByNamespaceAndName(fixture.namespaceID, "demo", trx)
	require.NoError(t, err)
	require.False(t, name.Deleted)
	require.True(t, name.Reserved)
}

func TestCreateShareNameConflictsWhenActiveFrontendMappingExists(t *testing.T) {
	fixture := setupShareNameFixture(t, true)
	attachDynamicFrontend(t, fixture, "dynamic-fe")

	trx, err := str.Begin()
	require.NoError(t, err)
	require.NoError(t, str.DeleteName(fixture.nameID, trx))
	_, err = trx.Exec("insert into frontend_mappings (frontend_token, name, share_token) values ($1, $2, $3)", "dynamic-fe", "demo.example.com", "share-token")
	require.NoError(t, err)
	require.NoError(t, trx.Commit())

	handler := newCreateShareNameHandler()
	resp := handler.Handle(shareops.CreateShareNameParams{
		Body: shareops.CreateShareNameBody{
			NamespaceToken: "public",
			Name:           "demo",
		},
	}, fixture.principal)

	conflict, ok := resp.(*shareops.CreateShareNameConflict)
	require.True(t, ok)
	require.Contains(t, string(conflict.Payload), "share-token")
}
