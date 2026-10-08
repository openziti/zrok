package controller

import (
	"encoding/json"
	"testing"

	"github.com/openziti/zrok/v2/controller/store"
	"github.com/openziti/zrok/v2/rest_model_zrok"
	"github.com/openziti/zrok/v2/rest_server_zrok/operations/admin"
	shareops "github.com/openziti/zrok/v2/rest_server_zrok/operations/share"
	"github.com/stretchr/testify/require"
)

func getNameAdminPrincipal(t *testing.T) *rest_model_zrok.Principal {
	t.Helper()
	trx, err := str.Begin()
	require.NoError(t, err)
	id, err := str.CreateAccount(&store.Account{Email: "admin@example.com", Salt: "salt", Password: "password", Token: "admin-token"}, trx)
	require.NoError(t, err)
	require.NoError(t, trx.Commit())
	return &rest_model_zrok.Principal{ID: int64(id), Email: "admin@example.com", Admin: true}
}

func TestGetNameRejectsNonAdmin(t *testing.T) {
	f := setupShareNameFixture(t, true)
	resp := newGetNameHandler().Handle(admin.GetNameParams{NamespaceToken: "public", Name: "demo"}, f.principal)
	require.IsType(t, &admin.GetNameUnauthorized{}, resp)
}

func TestGetNameFindsAnotherAccountsLiveShare(t *testing.T) {
	f := setupShareNameFixture(t, true)
	principal := getNameAdminPrincipal(t)
	require.NotEqual(t, f.accountID, int(principal.ID))
	resp := newGetNameHandler().Handle(admin.GetNameParams{NamespaceToken: "public", Name: "demo"}, principal)
	ok, isOK := resp.(*admin.GetNameOK)
	require.True(t, isOK)
	require.Equal(t, "public", *ok.Payload.NamespaceToken)
	require.Equal(t, "example.com", *ok.Payload.NamespaceName)
	require.Equal(t, "demo", *ok.Payload.Name)
	require.Equal(t, "test@example.com", *ok.Payload.AccountEmail)
	require.Equal(t, "share-token", *ok.Payload.ShareToken)
	require.True(t, *ok.Payload.Reserved)
	require.Positive(t, *ok.Payload.CreatedAt)
}

func TestGetNameAfterShareTeardownHasEmptyShareToken(t *testing.T) {
	f := setupShareNameFixture(t, true)
	principal := getNameAdminPrincipal(t)
	useZitiFake(t)
	unshared := newUnshareHandler().Handle(shareops.UnshareParams{Body: shareops.UnshareBody{EnvZID: "env-zid", ShareToken: "share-token"}}, f.principal)
	require.IsType(t, &shareops.UnshareOK{}, unshared)

	resp := newGetNameHandler().Handle(admin.GetNameParams{NamespaceToken: "public", Name: "demo"}, principal)
	ok, isOK := resp.(*admin.GetNameOK)
	require.True(t, isOK)
	require.Empty(t, *ok.Payload.ShareToken)
	encoded, err := json.Marshal(ok.Payload)
	require.NoError(t, err)
	require.Contains(t, string(encoded), `"shareToken":""`)
}

func TestGetNameUnknownNameIsNotFound(t *testing.T) {
	setupShareNameFixture(t, true)
	resp := newGetNameHandler().Handle(admin.GetNameParams{NamespaceToken: "public", Name: "unknown"}, getNameAdminPrincipal(t))
	require.IsType(t, &admin.GetNameNotFound{}, resp)
}

func TestGetNameUnknownNamespaceIsNotFound(t *testing.T) {
	setupShareNameFixture(t, true)
	resp := newGetNameHandler().Handle(admin.GetNameParams{NamespaceToken: "unknown", Name: "demo"}, getNameAdminPrincipal(t))
	require.IsType(t, &admin.GetNameNotFound{}, resp)
}

func TestGetNameStoreFailureIsInternalError(t *testing.T) {
	setupShareNameFixture(t, true)
	principal := getNameAdminPrincipal(t)
	trx, err := str.Begin()
	require.NoError(t, err)
	_, err = trx.Exec("alter table names rename to names_gone")
	require.NoError(t, err)
	require.NoError(t, trx.Commit())

	resp := newGetNameHandler().Handle(admin.GetNameParams{NamespaceToken: "public", Name: "demo"}, principal)
	require.IsType(t, &admin.GetNameInternalServerError{}, resp)
}
