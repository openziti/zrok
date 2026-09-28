package store

import (
	"testing"

	"github.com/openziti/zrok/controller/store/storetest"
	"github.com/stretchr/testify/require"
)

func openTestStore(t *testing.T) *Store {
	t.Helper()
	str, err := Open(&Config{Path: ":memory:", Type: "sqlite3"})
	require.NoError(t, err)
	t.Cleanup(func() { _ = str.Close() })
	return str
}

func createTestShare(t *testing.T, str *Store, token string) (acctId, shrId int) {
	t.Helper()
	trx, err := str.Begin()
	require.NoError(t, err)
	acctId, err = str.CreateAccount(&Account{Email: token + "@example.com", Salt: "salt", Password: "password", Token: token + "-acct"}, trx)
	require.NoError(t, err)
	envId, err := str.CreateEnvironment(acctId, &Environment{Description: "d", Host: "h", Address: "a", ZId: token + "-env"}, trx)
	require.NoError(t, err)
	shrId, err = str.CreateShare(envId, &Share{ZId: token + "-zid", Token: token, ShareMode: "public", BackendMode: "proxy", PermissionMode: OpenPermissionMode}, trx)
	require.NoError(t, err)
	require.NoError(t, trx.Commit())
	return acctId, shrId
}

func TestDeleteShareReleasesV2Mappings(t *testing.T) {
	str := openTestStore(t)
	require.False(t, str.v2Mappings)
	trx, err := str.Begin()
	require.NoError(t, err)
	require.NoError(t, storetest.CreateV2Tables(trx))
	require.NoError(t, trx.Commit())
	str.probeV2Mappings()
	require.True(t, str.v2Mappings)

	acctId, shrId := createTestShare(t, str, "torn")
	otherAcctId, otherShrId := createTestShare(t, str, "other")
	trx, err = str.Begin()
	require.NoError(t, err)
	v2, err := storetest.MapShare(trx, acctId, shrId, "torn", "torn")
	require.NoError(t, err)
	otherV2, err := storetest.MapShare(trx, otherAcctId, otherShrId, "other", "other")
	require.NoError(t, err)
	require.NoError(t, trx.Commit())

	trx, err = str.Begin()
	require.NoError(t, err)
	require.NoError(t, str.DeleteShare(shrId, trx))
	require.NoError(t, trx.Commit())

	trx, err = str.Begin()
	require.NoError(t, err)
	defer func() { _ = trx.Rollback() }()
	shr, err := str.GetShare(shrId, trx)
	require.NoError(t, err)
	require.True(t, shr.Deleted)
	names, frontends, err := storetest.LiveMappings(trx, shrId, "torn")
	require.NoError(t, err)
	require.Zero(t, names)
	require.Zero(t, frontends)
	deleted, err := storetest.NameDeleted(trx, v2.AutoNameId)
	require.NoError(t, err)
	require.True(t, deleted)
	deleted, err = storetest.NameDeleted(trx, v2.ReservedNameId)
	require.NoError(t, err)
	require.False(t, deleted)

	names, frontends, err = storetest.LiveMappings(trx, otherShrId, "other")
	require.NoError(t, err)
	require.Equal(t, 2, names)
	require.Equal(t, 2, frontends)
	for _, nameId := range []int{otherV2.AutoNameId, otherV2.ReservedNameId} {
		deleted, err = storetest.NameDeleted(trx, nameId)
		require.NoError(t, err)
		require.False(t, deleted)
	}
}

func TestDeleteShareWithoutV2Tables(t *testing.T) {
	str := openTestStore(t)
	require.False(t, str.v2Mappings)
	_, shrId := createTestShare(t, str, "plain")

	trx, err := str.Begin()
	require.NoError(t, err)
	require.NoError(t, str.DeleteShare(shrId, trx))
	require.NoError(t, trx.Commit())

	trx, err = str.Begin()
	require.NoError(t, err)
	defer func() { _ = trx.Rollback() }()
	shr, err := str.GetShare(shrId, trx)
	require.NoError(t, err)
	require.True(t, shr.Deleted)
}
