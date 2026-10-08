package store

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFrontendMappingsByShareToken(t *testing.T) {
	str, err := Open(&Config{Path: ":memory:", Type: "sqlite3"})
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, str.Close()) })

	trx, err := str.Begin()
	require.NoError(t, err)
	defer func() { _ = trx.Rollback() }()

	for _, fm := range []*FrontendMapping{
		{FrontendToken: "fe-a", Name: "one.example.com", ShareToken: "share-1"},
		{FrontendToken: "fe-b", Name: "one.example.com", ShareToken: "share-1"},
		{FrontendToken: "fe-a", Name: "two.example.com", ShareToken: "share-2"},
	} {
		_, err := str.CreateFrontendMapping(fm, trx)
		require.NoError(t, err)
	}

	fms, err := str.FindFrontendMappingsByShareToken("share-1", trx)
	require.NoError(t, err)
	require.Len(t, fms, 2)
	require.Equal(t, "fe-a", fms[0].FrontendToken)
	require.Equal(t, "fe-b", fms[1].FrontendToken)

	require.NoError(t, str.DeleteFrontendMappingsByShareToken("share-1", trx))
	fms, err = str.FindFrontendMappingsByShareToken("share-1", trx)
	require.NoError(t, err)
	require.Empty(t, fms)

	// other shares' rows are untouched, and deleting a token with no rows is not an error
	fms, err = str.FindFrontendMappingsByShareToken("share-2", trx)
	require.NoError(t, err)
	require.Len(t, fms, 1)
	require.NoError(t, str.DeleteFrontendMappingsByShareToken("absent", trx))
}

func TestFrontendMappingsWithoutLiveShare(t *testing.T) {
	f := setupRepairStoreFixture(t)
	f.mapping(t, "dead", "dead", false, true, false)
	f.mapping(t, "live", "live", false, false, false)
	// a reused token: a deleted row and a live row share it, so its mapping has a live share
	f.mapping(t, "reused", "reused-old", false, true, false)
	f.mapping(t, "reused", "reused-new", false, false, false)

	ids := make(map[string]int64)
	for _, fm := range []*FrontendMapping{
		{FrontendToken: "fe", Name: "dead.example.com", ShareToken: "dead"},
		{FrontendToken: "fe", Name: "ghost.example.com", ShareToken: "ghost"},
		{FrontendToken: "fe", Name: "live.example.com", ShareToken: "live"},
		{FrontendToken: "fe", Name: "reused.example.com", ShareToken: "reused"},
	} {
		id, err := f.str.CreateFrontendMapping(fm, f.trx)
		require.NoError(t, err)
		ids[fm.ShareToken] = int64(id)
	}

	count, err := f.str.CountFrontendMappingsWithoutLiveShare(f.trx)
	require.NoError(t, err)
	require.Equal(t, 2, count)
	fms, err := f.str.FindFrontendMappingsWithoutLiveShare(10, f.trx)
	require.NoError(t, err)
	require.Len(t, fms, 2)
	require.Equal(t, ids["dead"], fms[0].Id)
	require.True(t, fms[0].ShareRow)
	require.Equal(t, ids["ghost"], fms[1].Id)
	require.False(t, fms[1].ShareRow)
	fms, err = f.str.FindFrontendMappingsWithoutLiveShare(1, f.trx)
	require.NoError(t, err)
	require.Len(t, fms, 1)

	n, err := f.str.DeleteFrontendMappings([]int64{ids["dead"], ids["ghost"]}, f.trx)
	require.NoError(t, err)
	require.Equal(t, int64(2), n)
	count, err = f.str.CountFrontendMappingsWithoutLiveShare(f.trx)
	require.NoError(t, err)
	require.Zero(t, count)
	live, err := f.str.FindFrontendMappingsByShareToken("live", f.trx)
	require.NoError(t, err)
	require.Len(t, live, 1)
}

func TestFrontendMappingShareStateOverReusedToken(t *testing.T) {
	f := setupRepairStoreFixture(t)
	f.mapping(t, "dead", "dead", false, true, false)
	f.mapping(t, "reused", "reused-old", false, true, false)
	f.mapping(t, "reused", "reused-new", false, false, false)
	for _, token := range []string{"dead", "reused", "ghost"} {
		_, err := f.str.CreateFrontendMapping(&FrontendMapping{FrontendToken: "fe", Name: token + ".example.com", ShareToken: token}, f.trx)
		require.NoError(t, err)
	}

	for token, want := range map[string]*bool{"dead": boolPtr(true), "reused": boolPtr(false), "ghost": nil} {
		fm, err := f.str.FindFrontendMappingByFrontendTokenAndNameWithShareState("fe", token+".example.com", f.trx)
		require.NoError(t, err)
		require.Equal(t, want, fm.ShareDeleted, token)
	}
}

func boolPtr(v bool) *bool { return &v }
