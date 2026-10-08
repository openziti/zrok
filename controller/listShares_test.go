package controller

import (
	"testing"

	"github.com/openziti/zrok/v2/controller/store"
	"github.com/openziti/zrok/v2/rest_server_zrok/operations/metadata"
	"github.com/stretchr/testify/require"
)

func TestListSharesLimitJournalStoreFailureIsInternalError(t *testing.T) {
	f := setupShareCreateFixture(t)
	// the account limit check's journal query fails on a missing table
	trx, err := str.Begin()
	require.NoError(t, err)
	_, err = trx.Exec("alter table bandwidth_limit_journal rename to bandwidth_limit_journal_gone")
	require.NoError(t, err)
	require.NoError(t, trx.Commit())

	resp := newListSharesHandler().Handle(metadata.ListSharesParams{}, f.principal)

	require.IsType(t, &metadata.ListSharesInternalServerError{}, resp)
	require.Contains(t, f.logs.String(), "error checking account limit journal for 'owner@example.com'")
}

func TestListSharesFrontendEndpointsStoreFailureIsInternalError(t *testing.T) {
	f := setupShareCreateFixture(t)
	trx, err := str.Begin()
	require.NoError(t, err)
	envs, err := str.FindEnvironmentsForAccount(int(f.principal.ID), trx)
	require.NoError(t, err)
	_, err = str.CreateShare(envs[0].Id, &store.Share{ZId: "share-zid", Token: "share-token", ShareMode: "public", BackendMode: "proxy", PermissionMode: store.OpenPermissionMode}, trx)
	require.NoError(t, err)
	// the frontend endpoints' name query fails on a missing table
	_, err = trx.Exec("alter table share_name_mappings rename to share_name_mappings_gone")
	require.NoError(t, err)
	require.NoError(t, trx.Commit())

	resp := newListSharesHandler().Handle(metadata.ListSharesParams{}, f.principal)

	require.IsType(t, &metadata.ListSharesInternalServerError{}, resp)
	require.Contains(t, f.logs.String(), "error building frontend endpoints for user 'owner@example.com'")
	require.Contains(t, f.logs.String(), "error finding names for share 'share-token'")
}
