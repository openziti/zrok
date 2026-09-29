package controller

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/openziti/zrok/v2/controller/automation"
	"github.com/openziti/zrok/v2/controller/automation/zitifake"
	"github.com/openziti/zrok/v2/controller/dynamicProxyController"
	"github.com/openziti/zrok/v2/controller/store"
	"github.com/openziti/zrok/v2/rest_model_zrok"
	"github.com/openziti/zrok/v2/rest_server_zrok/operations/admin"
	shareops "github.com/openziti/zrok/v2/rest_server_zrok/operations/share"
	"github.com/stretchr/testify/require"
)

// recordedUpdate is a published mapping update, with what the store showed when it arrived.
type recordedUpdate struct {
	frontendToken string
	mapping       dynamicProxyController.Mapping
	// readable is false when the store could not be read, which on the single-connection sqlite store
	// means the publishing transaction was still open.
	readable bool
	// rowPresent reports whether a committed frontend_mappings row existed for the update's frontend and
	// name.
	rowPresent bool
}

type recordingPublisher struct {
	mu      sync.Mutex
	updates []recordedUpdate
	fail    error
}

func (p *recordingPublisher) Publish(ctx context.Context, frontendToken string, m dynamicProxyController.Mapping) error {
	rec := recordedUpdate{frontendToken: frontendToken, mapping: m}
	readCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	if trx, err := str.BeginContext(readCtx); err == nil {
		rec.readable = true
		_, err := str.FindFrontendMappingByFrontendTokenAndNameWithShareState(frontendToken, m.Name, trx)
		rec.rowPresent = err == nil
		_ = trx.Rollback()
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	p.updates = append(p.updates, rec)
	return p.fail
}

func (p *recordingPublisher) recorded() []recordedUpdate {
	p.mu.Lock()
	defer p.mu.Unlock()
	return append([]recordedUpdate(nil), p.updates...)
}

func (p *recordingPublisher) reset() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.updates = nil
}

func installRecordingPublisher(t *testing.T) *recordingPublisher {
	t.Helper()
	prev := mappingPub
	t.Cleanup(func() { mappingPub = prev })
	pub := &recordingPublisher{}
	mappingPub = pub
	return pub
}

// useZitiFake points cfg.Ziti at a fresh fake, for fixtures that set up the store alone.
func useZitiFake(t *testing.T) (*zitifake.Server, *automation.ZitiAutomation) {
	t.Helper()
	fake := zitifake.New()
	t.Cleanup(fake.Close)
	cfg.Ziti = &automation.Config{ApiEndpoint: fake.URL, Username: "admin", Password: "admin"}
	ziti, err := automation.NewZitiAutomation(cfg.Ziti)
	require.NoError(t, err)
	return fake, ziti
}

func insertFrontendMapping(t *testing.T, frontendToken, name, shareToken string) {
	t.Helper()
	trx, err := str.Begin()
	require.NoError(t, err)
	_, err = str.CreateFrontendMapping(&store.FrontendMapping{FrontendToken: frontendToken, Name: name, ShareToken: shareToken}, trx)
	require.NoError(t, err)
	require.NoError(t, trx.Commit())
}

func requireFrontendMappings(t *testing.T, shareToken string, want int) {
	t.Helper()
	trx, err := str.Begin()
	require.NoError(t, err)
	defer func() { _ = trx.Rollback() }()
	fms, err := str.FindFrontendMappingsByShareToken(shareToken, trx)
	require.NoError(t, err)
	require.Len(t, fms, want)
}

func unbindOf(frontendToken, name string) recordedUpdate {
	return recordedUpdate{
		frontendToken: frontendToken,
		mapping:       dynamicProxyController.Mapping{Operation: dynamicProxyController.OperationUnbind, Name: name},
		readable:      true,
	}
}

// requireShareReleased asserts the end state of a share torn down through teardownShare: the share and
// its name mapping are deleted, no frontend mapping carries its token, an auto-allocated name is
// released and a reserved name survives and can be selected again.
func requireShareReleased(t *testing.T, shareID, nameID int, shareToken, name string, reserved bool, principal *rest_model_zrok.Principal) {
	t.Helper()
	trx, err := str.Begin()
	require.NoError(t, err)
	defer func() { _ = trx.Rollback() }()

	var shareDeleted bool
	require.NoError(t, trx.QueryRow("select deleted from shares where id = $1", shareID).Scan(&shareDeleted))
	require.True(t, shareDeleted)

	var mappingDeleted bool
	require.NoError(t, trx.QueryRow("select deleted from share_name_mappings where share_id = $1", shareID).Scan(&mappingDeleted))
	require.True(t, mappingDeleted)

	// the availability query share creation uses
	mappings, err := str.FindShareNameMappingsByNameId(nameID, trx)
	require.NoError(t, err)
	require.Empty(t, mappings)

	fms, err := str.FindFrontendMappingsByShareToken(shareToken, trx)
	require.NoError(t, err)
	require.Empty(t, fms)

	var nameDeleted bool
	require.NoError(t, trx.QueryRow("select deleted from names where id = $1", nameID).Scan(&nameDeleted))
	require.Equal(t, !reserved, nameDeleted)

	if reserved {
		endpoints, nameIds, err := (&shareHandler{}).processNameSelections(
			[]*rest_model_zrok.NameSelection{{NamespaceToken: "public", Name: name}},
			"new-share-token", principal, trx)
		require.NoError(t, err)
		require.Equal(t, []string{name + ".example.com"}, endpoints)
		require.Equal(t, []int{nameID}, nameIds)
	}
}

func TestDisableReleasesShareNames(t *testing.T) {
	for _, reserved := range []bool{true, false} {
		t.Run(map[bool]string{true: "reserved", false: "allocated"}[reserved], func(t *testing.T) {
			fixture := setupShareNameFixture(t, reserved)
			attachDynamicFrontend(t, fixture, "dynamic-fe")
			insertFrontendMapping(t, "dynamic-fe", "demo.example.com", "share-token")
			_, ziti := useZitiFake(t)

			trx, err := str.Begin()
			require.NoError(t, err)
			env, err := str.GetEnvironment(fixture.environmentID, trx)
			require.NoError(t, err)
			updates, err := disableEnvironment(env, trx, ziti)
			require.NoError(t, err)
			require.NoError(t, trx.Commit())

			require.Equal(t, []pendingMappingUpdate{{
				frontendToken: "dynamic-fe",
				mapping:       dynamicProxyController.Mapping{Operation: dynamicProxyController.OperationUnbind, Name: "demo.example.com"},
			}}, updates)
			requireShareReleased(t, fixture.shareID, fixture.nameID, "share-token", "demo", reserved, fixture.principal)
		})
	}
}

func TestUnshareReleasesShareNames(t *testing.T) {
	for _, reserved := range []bool{true, false} {
		t.Run(map[bool]string{true: "reserved", false: "allocated"}[reserved], func(t *testing.T) {
			fixture := setupShareNameFixture(t, reserved)
			attachDynamicFrontend(t, fixture, "dynamic-fe")
			insertFrontendMapping(t, "dynamic-fe", "demo.example.com", "share-token")
			useZitiFake(t)
			pub := installRecordingPublisher(t)

			resp := newUnshareHandler().Handle(shareops.UnshareParams{Body: shareops.UnshareBody{EnvZID: "env-zid", ShareToken: "share-token"}}, fixture.principal)

			require.IsType(t, &shareops.UnshareOK{}, resp)
			requireShareReleased(t, fixture.shareID, fixture.nameID, "share-token", "demo", reserved, fixture.principal)
			require.Equal(t, []recordedUpdate{unbindOf("dynamic-fe", "demo.example.com")}, pub.recorded())
		})
	}
}

func TestDeleteAccountTearsDownEveryShare(t *testing.T) {
	fixture := setupShareNameFixture(t, true)
	attachDynamicFrontend(t, fixture, "dynamic-fe")
	insertFrontendMapping(t, "dynamic-fe", "demo.example.com", "share-token")
	useZitiFake(t)
	pub := installRecordingPublisher(t)

	// a second environment with a share on an auto-allocated name
	trx, err := str.Begin()
	require.NoError(t, err)
	env2ID, err := str.CreateEnvironment(fixture.accountID, &store.Environment{Description: "second", Host: "host", Address: "address", ZId: "env-zid-2"}, trx)
	require.NoError(t, err)
	share2ID, err := str.CreateShare(env2ID, &store.Share{ZId: "share-zid-2", Token: "share-token-2", ShareMode: "public", BackendMode: "proxy", PermissionMode: store.OpenPermissionMode}, trx)
	require.NoError(t, err)
	name2ID, err := str.CreateName(&store.Name{NamespaceId: fixture.namespaceID, Name: "share-token-2", AccountId: fixture.accountID}, trx)
	require.NoError(t, err)
	_, err = str.CreateShareNameMapping(&store.ShareNameMapping{ShareId: share2ID, NameId: name2ID}, trx)
	require.NoError(t, err)
	require.NoError(t, trx.Commit())
	insertFrontendMapping(t, "dynamic-fe", "share-token-2.example.com", "share-token-2")

	resp := newDeleteAccountHandler().Handle(admin.DeleteAccountParams{Body: admin.DeleteAccountBody{Email: "test@example.com"}}, &rest_model_zrok.Principal{Admin: true})

	require.IsType(t, &admin.DeleteAccountOK{}, resp)
	requireShareReleased(t, fixture.shareID, fixture.nameID, "share-token", "demo", true, fixture.principal)
	requireShareReleased(t, share2ID, name2ID, "share-token-2", "share-token-2", false, fixture.principal)
	require.ElementsMatch(t, []recordedUpdate{
		unbindOf("dynamic-fe", "demo.example.com"),
		unbindOf("dynamic-fe", "share-token-2.example.com"),
	}, pub.recorded())
}

// addDynamicFrontend maps a dynamic frontend onto the share-create fixture's namespace 'public'.
func addDynamicFrontend(t *testing.T, token string) {
	t.Helper()
	trx, err := str.Begin()
	require.NoError(t, err)
	ns, err := str.FindNamespaceWithToken("public", trx)
	require.NoError(t, err)
	feID, err := str.CreateGlobalFrontend(&store.Frontend{Token: token, ZId: token + "-zid", Dynamic: true, PermissionMode: store.OpenPermissionMode}, trx)
	require.NoError(t, err)
	_, err = str.CreateNamespaceFrontendMapping(ns.Id, feID, false, trx)
	require.NoError(t, err)
	require.NoError(t, trx.Commit())
}

func (f *shareCreateFixture) unshare(token string) interface{} {
	return newUnshareHandler().Handle(shareops.UnshareParams{Body: shareops.UnshareBody{EnvZID: "env-zid", ShareToken: token}}, f.principal)
}

func (f *shareCreateFixture) createPublicShare(t *testing.T) string {
	t.Helper()
	resp := f.share(publicShareRequest(string(store.OpenPermissionMode)))
	created, ok := resp.(*shareops.ShareCreated)
	require.True(t, ok, "%T", resp)
	return created.Payload.ShareToken
}

func TestShareCreatePublishesBindAfterCommit(t *testing.T) {
	f := setupShareCreateFixture(t)
	addDynamicFrontend(t, "dynamic-fe")
	pub := installRecordingPublisher(t)

	token := f.createPublicShare(t)

	trx, err := str.Begin()
	require.NoError(t, err)
	fm, err := str.FindFrontendMappingByFrontendTokenAndNameWithShareState("dynamic-fe", token+".example.com", trx)
	require.NoError(t, err)
	require.NoError(t, trx.Rollback())
	require.Equal(t, []recordedUpdate{{
		frontendToken: "dynamic-fe",
		mapping: dynamicProxyController.Mapping{
			Id:         fm.Id,
			Operation:  dynamicProxyController.OperationBind,
			Name:       token + ".example.com",
			ShareToken: token,
		},
		readable:   true,
		rowPresent: true,
	}}, pub.recorded())
}

func TestShareCreateFailurePublishesNothing(t *testing.T) {
	f := setupShareCreateFixture(t)
	addDynamicFrontend(t, "dynamic-fe")
	pub := installRecordingPublisher(t)
	token := f.recordToken()

	resp := f.share(publicShareRequest(string(store.ClosedPermissionMode), "nobody@example.com"))

	require.IsType(t, &shareops.ShareInternalServerError{}, resp)
	require.Empty(t, pub.recorded())
	requireFrontendMappings(t, *token, 0)
	f.requireNoShares(t)
}

func TestUnsharePublishesUnbindAfterCommit(t *testing.T) {
	f := setupShareCreateFixture(t)
	addDynamicFrontend(t, "dynamic-fe")
	pub := installRecordingPublisher(t)
	token := f.createPublicShare(t)
	pub.reset()

	resp := f.unshare(token)

	require.IsType(t, &shareops.UnshareOK{}, resp)
	require.Equal(t, []recordedUpdate{unbindOf("dynamic-fe", token+".example.com")}, pub.recorded())
	requireFrontendMappings(t, token, 0)
	require.Empty(t, f.fake.Tagged("zrokShareToken", token))
}

func TestPublishFailureDoesNotChangeResponse(t *testing.T) {
	f := setupShareCreateFixture(t)
	addDynamicFrontend(t, "dynamic-fe")
	pub := installRecordingPublisher(t)
	pub.fail = errors.New("broker unavailable")

	token := f.createPublicShare(t)
	requireFrontendMappings(t, token, 1)

	resp := f.unshare(token)

	require.IsType(t, &shareops.UnshareOK{}, resp)
	requireFrontendMappings(t, token, 0)
	require.Len(t, pub.recorded(), 2)
	require.Contains(t, f.logs.String(), "broker unavailable")
}

func TestShareFailedFrontendMappingInsertFailsRequest(t *testing.T) {
	f := setupShareCreateFixture(t)
	addDynamicFrontend(t, "dynamic-fe")
	pub := installRecordingPublisher(t)
	trx, err := str.Begin()
	require.NoError(t, err)
	ns, err := str.FindNamespaceWithToken("public", trx)
	require.NoError(t, err)
	_, err = str.CreateName(&store.Name{NamespaceId: ns.Id, Name: "demo", AccountId: int(f.principal.ID), Reserved: true}, trx)
	require.NoError(t, err)
	require.NoError(t, trx.Commit())
	// (frontend_token, name) is unique, so this row makes the create's insert fail
	insertFrontendMapping(t, "dynamic-fe", "demo.example.com", "other-token")
	token := f.recordToken()
	req := publicShareRequest(string(store.OpenPermissionMode))
	req.NameSelections = []*rest_model_zrok.NameSelection{{NamespaceToken: "public", Name: "demo"}}

	resp := f.share(req)

	require.IsType(t, &shareops.ShareInternalServerError{}, resp)
	require.Contains(t, f.logs.String(), "error recording frontend mapping 'demo.example.com'")
	f.requireNoShares(t)
	f.requireCompensated(t, *token, 5)
	requireFrontendMappings(t, "other-token", 1)
	require.Empty(t, pub.recorded())
}

func TestUnshareZitiFailureRollsBack(t *testing.T) {
	f := setupShareCreateFixture(t)
	addDynamicFrontend(t, "dynamic-fe")
	pub := installRecordingPublisher(t)
	token := f.createPublicShare(t)
	pub.reset()
	f.fake.RejectOperations(true)

	resp := f.unshare(token)

	require.IsType(t, &shareops.UnshareInternalServerError{}, resp)
	f.fake.RejectOperations(false)
	require.Empty(t, pub.recorded())
	require.Len(t, f.fake.Tagged("zrokShareToken", token), 5)
	trx, err := str.Begin()
	require.NoError(t, err)
	defer func() { _ = trx.Rollback() }()
	shr, err := str.FindShareWithToken(token, trx)
	require.NoError(t, err)
	require.False(t, shr.Deleted)
	ns, err := str.FindNamespaceWithToken("public", trx)
	require.NoError(t, err)
	name, err := str.FindNameByNamespaceAndName(ns.Id, token, trx)
	require.NoError(t, err)
	require.False(t, name.Deleted)
	mappings, err := str.FindShareNameMappingsByNameId(name.Id, trx)
	require.NoError(t, err)
	require.Len(t, mappings, 1)
	fms, err := str.FindFrontendMappingsByShareToken(token, trx)
	require.NoError(t, err)
	require.Len(t, fms, 1)
}
