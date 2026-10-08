package store

import (
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
)

type repairStoreFixture struct {
	str   *Store
	trx   *sqlx.Tx
	acct  int
	env   int
	nsId  int
	names map[string]int
}

func setupRepairStoreFixture(t *testing.T) *repairStoreFixture {
	t.Helper()
	str, err := Open(&Config{Path: ":memory:", Type: "sqlite3"})
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, str.Close()) })
	trx, err := str.Begin()
	require.NoError(t, err)
	t.Cleanup(func() { _ = trx.Rollback() })
	f := &repairStoreFixture{str: str, trx: trx, names: make(map[string]int)}
	f.acct, err = str.CreateAccount(&Account{Email: "owner@example.com", Salt: "salt", Password: "password", Token: "acct-token"}, trx)
	require.NoError(t, err)
	f.env, err = str.CreateEnvironment(f.acct, &Environment{Description: "env", Host: "host", Address: "address", ZId: "env-zid"}, trx)
	require.NoError(t, err)
	f.nsId, err = str.CreateNamespace(&Namespace{Token: "public", Name: "example.com", Open: true}, trx)
	require.NoError(t, err)
	return f
}

// mapping creates a share with token holding the name through a mapping, then marks the share or the
// name deleted as asked, and returns the mapping id.
func (f *repairStoreFixture) mapping(t *testing.T, token, name string, reserved, shareDeleted, nameDeleted bool) int {
	t.Helper()
	nameId, err := f.str.CreateName(&Name{NamespaceId: f.nsId, Name: name, AccountId: f.acct, Reserved: reserved}, f.trx)
	require.NoError(t, err)
	f.names[name] = nameId
	shareId, err := f.str.CreateShare(f.env, &Share{ZId: name + "-zid", Token: token, ShareMode: "public", BackendMode: "proxy", PermissionMode: OpenPermissionMode}, f.trx)
	require.NoError(t, err)
	id, err := f.str.CreateShareNameMapping(&ShareNameMapping{ShareId: shareId, NameId: nameId}, f.trx)
	require.NoError(t, err)
	if shareDeleted {
		require.NoError(t, f.str.DeleteShare(shareId, f.trx))
	}
	if nameDeleted {
		require.NoError(t, f.str.DeleteName(nameId, f.trx))
	}
	return id
}

func TestShareNameMappingRepairQueries(t *testing.T) {
	f := setupRepairStoreFixture(t)
	res := f.mapping(t, "dead-res", "res", true, true, false)
	auto := f.mapping(t, "dead-auto", "auto", false, true, false)
	autoGone := f.mapping(t, "dead-auto-gone", "auto-gone", false, true, true)
	deadName := f.mapping(t, "live-dn", "gone", true, false, true)
	f.mapping(t, "live", "live", true, false, false)
	torn := f.mapping(t, "torn", "torn", false, true, false)
	require.NoError(t, f.str.DeleteShareNameMapping(torn, f.trx))

	count, err := f.str.CountShareNameMappingsToDeletedShares(true, f.trx)
	require.NoError(t, err)
	require.Equal(t, 1, count)
	count, err = f.str.CountShareNameMappingsToDeletedShares(false, f.trx)
	require.NoError(t, err)
	require.Equal(t, 2, count)
	count, err = f.str.CountAllocatedNamesHeldByDeletedShares(f.trx)
	require.NoError(t, err)
	require.Equal(t, 1, count)
	count, err = f.str.CountShareNameMappingsToDeletedNames(f.trx)
	require.NoError(t, err)
	require.Equal(t, 1, count)

	details, err := f.str.FindShareNameMappingsToDeletedShares(true, 10, f.trx)
	require.NoError(t, err)
	require.Len(t, details, 1)
	require.Equal(t, ShareNameMappingRepairDetail{MappingId: res, NameId: f.names["res"], Name: "res", Reserved: true, NamespaceToken: "public", ShareToken: "dead-res"}, *details[0])

	details, err = f.str.FindShareNameMappingsToDeletedShares(false, 1, f.trx)
	require.NoError(t, err)
	require.Len(t, details, 1)
	require.Equal(t, auto, details[0].MappingId)
	details, err = f.str.FindShareNameMappingsToDeletedShares(false, 10, f.trx)
	require.NoError(t, err)
	require.Len(t, details, 2)
	require.Equal(t, autoGone, details[1].MappingId)
	require.True(t, details[1].NameDeleted)

	details, err = f.str.FindShareNameMappingsToDeletedNames(10, f.trx)
	require.NoError(t, err)
	require.Len(t, details, 1)
	require.Equal(t, deadName, details[0].MappingId)
	require.Equal(t, "live-dn", details[0].ShareToken)

	// a reserved name is never released, a deleted name is not counted, and a deleted mapping is not
	// deleted again
	n, err := f.str.DeleteAllocatedNames([]int{f.names["res"], f.names["auto"], f.names["auto-gone"]}, f.trx)
	require.NoError(t, err)
	require.Equal(t, int64(1), n)
	_, err = f.str.GetName(f.names["res"], f.trx)
	require.NoError(t, err)
	_, err = f.str.GetName(f.names["auto"], f.trx)
	require.Error(t, err)

	n, err = f.str.DeleteShareNameMappings([]int{res, auto, torn}, f.trx)
	require.NoError(t, err)
	require.Equal(t, int64(2), n)
	count, err = f.str.CountShareNameMappingsToDeletedShares(true, f.trx)
	require.NoError(t, err)
	require.Equal(t, 0, count)

	n, err = f.str.DeleteShareNameMappings(nil, f.trx)
	require.NoError(t, err)
	require.Zero(t, n)
}
