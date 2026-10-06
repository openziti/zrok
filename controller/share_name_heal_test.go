package controller

import (
	"testing"

	"github.com/openziti/zrok/v2/controller/dynamicProxyController"
	"github.com/openziti/zrok/v2/controller/store"
	"github.com/openziti/zrok/v2/rest_model_zrok"
	shareops "github.com/openziti/zrok/v2/rest_server_zrok/operations/share"
	"github.com/stretchr/testify/require"
)

// seedNameHolder reserves the name 'demo' in the share-create fixture's namespace and creates a share
// with the given token holding it through a share name mapping; deleted marks that share row deleted and
// leaves the mapping live, which is what a severed name looks like.
func seedNameHolder(t *testing.T, f *shareCreateFixture, token string, deleted bool) (nameID, shareID int) {
	t.Helper()
	trx, err := str.Begin()
	require.NoError(t, err)
	ns, err := str.FindNamespaceWithToken("public", trx)
	require.NoError(t, err)
	env, err := str.FindEnvironmentForAccount("env-zid", int(f.principal.ID), trx)
	require.NoError(t, err)
	nameID, err = str.CreateName(&store.Name{NamespaceId: ns.Id, Name: "demo", AccountId: int(f.principal.ID), Reserved: true}, trx)
	require.NoError(t, err)
	shareID, err = str.CreateShare(env.Id, &store.Share{
		ZId:            token + "-zid",
		Token:          token,
		ShareMode:      "public",
		BackendMode:    "proxy",
		PermissionMode: store.OpenPermissionMode,
	}, trx)
	require.NoError(t, err)
	_, err = str.CreateShareNameMapping(&store.ShareNameMapping{ShareId: shareID, NameId: nameID}, trx)
	require.NoError(t, err)
	if deleted {
		require.NoError(t, str.DeleteShare(shareID, trx))
	}
	require.NoError(t, trx.Commit())
	return nameID, shareID
}

func demoShareRequest(permissionMode string, accessGrants ...string) *rest_model_zrok.ShareRequest {
	req := publicShareRequest(permissionMode, accessGrants...)
	req.NameSelections = []*rest_model_zrok.NameSelection{{NamespaceToken: "public", Name: "demo"}}
	return req
}

// requireNameMappings asserts the live share name mappings for nameID, as share id and share deleted flag.
func requireNameMappings(t *testing.T, nameID int, want map[int]bool) {
	t.Helper()
	trx, err := str.Begin()
	require.NoError(t, err)
	defer func() { _ = trx.Rollback() }()
	mappings, err := str.FindShareNameMappingsByNameIdWithShare(nameID, trx)
	require.NoError(t, err)
	got := make(map[int]bool)
	for _, m := range mappings {
		got[m.ShareId] = m.ShareDeleted
	}
	require.Equal(t, want, got)
}

func TestShareCreateHealsSeveredName(t *testing.T) {
	f := setupShareCreateFixture(t)
	addDynamicFrontend(t, "dynamic-fe")
	pub := installRecordingPublisher(t)
	nameID, deadID := seedNameHolder(t, f, "dead-token", true)
	insertFrontendMapping(t, "dynamic-fe", "demo.example.com", "dead-token")

	resp := f.share(demoShareRequest(string(store.OpenPermissionMode)))

	created, ok := resp.(*shareops.ShareCreated)
	require.True(t, ok, "%T", resp)
	token := created.Payload.ShareToken
	require.Equal(t, []string{"demo.example.com"}, created.Payload.FrontendProxyEndpoints)

	trx, err := str.Begin()
	require.NoError(t, err)
	var deadMappingDeleted bool
	require.NoError(t, trx.QueryRow("select deleted from share_name_mappings where share_id = $1", deadID).Scan(&deadMappingDeleted))
	require.True(t, deadMappingDeleted)
	shr, err := str.FindShareWithToken(token, trx)
	require.NoError(t, err)
	fm, err := str.FindFrontendMappingByFrontendTokenAndNameWithShareState("dynamic-fe", "demo.example.com", trx)
	require.NoError(t, err)
	require.NoError(t, trx.Rollback())

	// the healed mapping's unbind goes out ahead of the new bind for the same name, both after commit
	unbind := unbindOf("dynamic-fe", "demo.example.com")
	unbind.rowPresent = true
	require.Equal(t, []recordedUpdate{unbind, {
		frontendToken: "dynamic-fe",
		mapping: dynamicProxyController.Mapping{
			Id:         fm.Id,
			Operation:  dynamicProxyController.OperationBind,
			Name:       "demo.example.com",
			ShareToken: token,
		},
		readable:   true,
		rowPresent: true,
	}}, pub.recorded())

	requireNameMappings(t, nameID, map[int]bool{shr.Id: false})
	requireFrontendMappings(t, "dead-token", 0)
	requireFrontendMappings(t, token, 1)
	require.Contains(t, f.logs.String(), "healed severed name 'demo'")
	require.Contains(t, f.logs.String(), "deleted share 'dead-token'")
}

func TestShareCreateLiveHolderConflictNamesToken(t *testing.T) {
	f := setupShareCreateFixture(t)
	addDynamicFrontend(t, "dynamic-fe")
	nameID, liveID := seedNameHolder(t, f, "live-token", false)
	insertFrontendMapping(t, "dynamic-fe", "demo.example.com", "live-token")

	resp := f.share(demoShareRequest(string(store.OpenPermissionMode)))

	conflict, ok := resp.(*shareops.ShareConflict)
	require.True(t, ok, "%T", resp)
	require.Equal(t, "name 'demo' in namespace 'public' is in use by share 'live-token'; run 'zrok2 delete share live-token' to release it", string(conflict.Payload))
	requireNameMappings(t, nameID, map[int]bool{liveID: false})
	requireFrontendMappings(t, "live-token", 1)
	created, _ := f.fake.Log()
	require.Empty(t, created)
	require.NotContains(t, f.logs.String(), "healed severed name")
}

func TestShareCreateLiveFrontendMappingIsConflict(t *testing.T) {
	f := setupShareCreateFixture(t)
	addDynamicFrontend(t, "dynamic-fe")
	// the name's share mapping is severed, but a live share still holds its frontend mapping
	nameID, deadID := seedNameHolder(t, f, "dead-token", true)
	trx, err := str.Begin()
	require.NoError(t, err)
	env, err := str.FindEnvironmentForAccount("env-zid", int(f.principal.ID), trx)
	require.NoError(t, err)
	_, err = str.CreateShare(env.Id, &store.Share{ZId: "live-zid", Token: "live-token", ShareMode: "public", BackendMode: "proxy", PermissionMode: store.OpenPermissionMode}, trx)
	require.NoError(t, err)
	require.NoError(t, trx.Commit())
	insertFrontendMapping(t, "dynamic-fe", "demo.example.com", "live-token")

	resp := f.share(demoShareRequest(string(store.OpenPermissionMode)))

	conflict, ok := resp.(*shareops.ShareConflict)
	require.True(t, ok, "%T", resp)
	require.Contains(t, string(conflict.Payload), "in use by share 'live-token'")
	requireNameMappings(t, nameID, map[int]bool{deadID: true})
	requireFrontendMappings(t, "live-token", 1)
}

func TestNamePathsAgreeOnLiveHolder(t *testing.T) {
	fixture := setupShareNameFixture(t, true)
	attachDynamicFrontend(t, fixture, "dynamic-fe")
	pub := installRecordingPublisher(t)

	trx, err := str.Begin()
	require.NoError(t, err)
	_, _, _, err = (&shareHandler{}).processNameSelections(
		[]*rest_model_zrok.NameSelection{{NamespaceToken: "public", Name: "demo"}},
		"new-share-token", fixture.principal, trx)
	require.Error(t, err)
	require.NoError(t, trx.Rollback())
	want := "name 'demo' in namespace 'public' is in use by share 'share-token'; run 'zrok2 delete share share-token' to release it"
	require.Equal(t, want, err.Error())

	resp := newDeleteShareNameHandler().Handle(shareops.DeleteShareNameParams{
		Body: shareops.DeleteShareNameBody{NamespaceToken: "public", Name: "demo"},
	}, fixture.principal)
	deleteConflict, ok := resp.(*shareops.DeleteShareNameConflict)
	require.True(t, ok, "%T", resp)
	require.Equal(t, want, string(deleteConflict.Payload))

	// create-name meets the same holder through its frontend mapping once the name itself is gone
	trx, err = str.Begin()
	require.NoError(t, err)
	require.NoError(t, str.DeleteName(fixture.nameID, trx))
	_, err = str.CreateFrontendMapping(&store.FrontendMapping{FrontendToken: "dynamic-fe", Name: "demo.example.com", ShareToken: "share-token"}, trx)
	require.NoError(t, err)
	require.NoError(t, trx.Commit())

	resp = newCreateShareNameHandler().Handle(shareops.CreateShareNameParams{
		Body: shareops.CreateShareNameBody{NamespaceToken: "public", Name: "demo"},
	}, fixture.principal)
	createConflict, ok := resp.(*shareops.CreateShareNameConflict)
	require.True(t, ok, "%T", resp)
	require.Equal(t, want, string(createConflict.Payload))

	// the conflicting create-name left no name behind
	trx, err = str.Begin()
	require.NoError(t, err)
	defer func() { _ = trx.Rollback() }()
	available, err := str.CheckNameAvailability(fixture.namespaceID, "demo", trx)
	require.NoError(t, err)
	require.True(t, available)
	require.Empty(t, pub.recorded())
}

// severName marks the share-name fixture's share deleted and gives it a frontend mapping on 'dynamic-fe',
// so that its name is severed on both halves.
func severName(t *testing.T, fixture *shareNameFixture) {
	t.Helper()
	attachDynamicFrontend(t, fixture, "dynamic-fe")
	trx, err := str.Begin()
	require.NoError(t, err)
	require.NoError(t, str.DeleteShare(fixture.shareID, trx))
	_, err = str.CreateFrontendMapping(&store.FrontendMapping{FrontendToken: "dynamic-fe", Name: "demo.example.com", ShareToken: "share-token"}, trx)
	require.NoError(t, err)
	require.NoError(t, trx.Commit())
}

func TestCreateShareNamePublishesHealedUnbindAfterCommit(t *testing.T) {
	fixture := setupShareNameFixture(t, true)
	severName(t, fixture)
	trx, err := str.Begin()
	require.NoError(t, err)
	require.NoError(t, str.DeleteName(fixture.nameID, trx))
	require.NoError(t, trx.Commit())
	pub := installRecordingPublisher(t)

	resp := newCreateShareNameHandler().Handle(shareops.CreateShareNameParams{
		Body: shareops.CreateShareNameBody{NamespaceToken: "public", Name: "demo"},
	}, fixture.principal)

	require.IsType(t, &shareops.CreateShareNameCreated{}, resp)
	require.Equal(t, []recordedUpdate{unbindOf("dynamic-fe", "demo.example.com")}, pub.recorded())
	requireFrontendMappings(t, "share-token", 0)
}

func TestDeleteShareNamePublishesHealedUnbindAfterCommit(t *testing.T) {
	fixture := setupShareNameFixture(t, true)
	severName(t, fixture)
	pub := installRecordingPublisher(t)

	resp := newDeleteShareNameHandler().Handle(shareops.DeleteShareNameParams{
		Body: shareops.DeleteShareNameBody{NamespaceToken: "public", Name: "demo"},
	}, fixture.principal)

	require.IsType(t, &shareops.DeleteShareNameOK{}, resp)
	require.Equal(t, []recordedUpdate{unbindOf("dynamic-fe", "demo.example.com")}, pub.recorded())
	requireFrontendMappings(t, "share-token", 0)
	requireNameMappings(t, fixture.nameID, map[int]bool{})
}

func TestShareCreateNameCheckStoreFailureIsInternalError(t *testing.T) {
	f := setupShareCreateFixture(t)
	seedNameHolder(t, f, "dead-token", true)
	// the name check's mapping query fails on a missing table
	trx, err := str.Begin()
	require.NoError(t, err)
	_, err = trx.Exec("alter table share_name_mappings rename to share_name_mappings_gone")
	require.NoError(t, err)
	require.NoError(t, trx.Commit())

	resp := f.share(demoShareRequest(string(store.OpenPermissionMode)))

	require.IsType(t, &shareops.ShareInternalServerError{}, resp)
	require.Contains(t, f.logs.String(), "error finding share name mappings for name 'demo'")
	created, _ := f.fake.Log()
	require.Empty(t, created)
}

func TestShareCreateHealingRollsBackWithRequest(t *testing.T) {
	f := setupShareCreateFixture(t)
	addDynamicFrontend(t, "dynamic-fe")
	pub := installRecordingPublisher(t)
	nameID, deadID := seedNameHolder(t, f, "dead-token", true)
	insertFrontendMapping(t, "dynamic-fe", "demo.example.com", "dead-token")

	// an unknown access grant fails the request after the name was healed
	resp := f.share(demoShareRequest(string(store.ClosedPermissionMode), "nobody@example.com"))

	require.IsType(t, &shareops.ShareInternalServerError{}, resp)
	f.requireNoShares(t)
	requireNameMappings(t, nameID, map[int]bool{deadID: true})
	requireFrontendMappings(t, "dead-token", 1)
	require.Empty(t, pub.recorded())

	// the next attempt finds the severed name intact and heals it again
	resp = f.share(demoShareRequest(string(store.OpenPermissionMode)))

	created, ok := resp.(*shareops.ShareCreated)
	require.True(t, ok, "%T", resp)
	requireFrontendMappings(t, "dead-token", 0)
	requireFrontendMappings(t, created.Payload.ShareToken, 1)
	updates := pub.recorded()
	require.Len(t, updates, 2)
	require.Equal(t, dynamicProxyController.OperationUnbind, updates[0].mapping.Operation)
	require.Equal(t, dynamicProxyController.OperationBind, updates[1].mapping.Operation)
}
