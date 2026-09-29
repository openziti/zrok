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
