package store

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFindNameWithShareTokenByNamespaceAndName(t *testing.T) {
	f := setupRepairStoreFixture(t)
	f.mapping(t, "live-token", "live", true, false, false)
	f.mapping(t, "deleted-token", "deleted-share", true, true, false)
	_, err := f.str.CreateName(&Name{NamespaceId: f.nsId, Name: "unmapped", AccountId: f.acct, Reserved: true}, f.trx)
	require.NoError(t, err)

	for _, tc := range []struct {
		name      string
		wantToken string
	}{
		{name: "live", wantToken: "live-token"},
		{name: "deleted-share"},
		{name: "unmapped"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			an, err := f.str.FindNameWithShareTokenByNamespaceAndName(f.nsId, tc.name, f.trx)
			require.NoError(t, err)
			require.Equal(t, tc.name, an.Name.Name)
			if tc.wantToken == "" {
				require.Nil(t, an.ShareToken)
			} else {
				require.Equal(t, tc.wantToken, *an.ShareToken)
			}
		})
	}

	t.Run("absent", func(t *testing.T) {
		an, err := f.str.FindNameWithShareTokenByNamespaceAndName(f.nsId, "absent", f.trx)
		require.Nil(t, an)
		require.True(t, errors.Is(err, sql.ErrNoRows))
	})
}
