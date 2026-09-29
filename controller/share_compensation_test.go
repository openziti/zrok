package controller

import (
	"bytes"
	"slices"
	"strings"
	"sync"
	"testing"

	"github.com/michaelquigley/df/dl"
	"github.com/openziti/zrok/v2/controller/automation"
	"github.com/openziti/zrok/v2/controller/automation/zitifake"
	controllerConfig "github.com/openziti/zrok/v2/controller/config"
	"github.com/openziti/zrok/v2/controller/store"
	"github.com/openziti/zrok/v2/rest_model_zrok"
	shareops "github.com/openziti/zrok/v2/rest_server_zrok/operations/share"
	"github.com/stretchr/testify/require"
)

type lockedBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *lockedBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Write(p)
}

func (b *lockedBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.String()
}

type shareCreateFixture struct {
	fake      *zitifake.Server
	logs      *lockedBuffer
	principal *rest_model_zrok.Principal
}

// setupShareCreateFixture reaches the fake through the production NewZitiAutomation; the plain-http
// endpoint skips the ca fetch. the namespace 'public' has one frontend, so public shares selecting it
// get a dial policy.
func setupShareCreateFixture(t *testing.T) *shareCreateFixture {
	t.Helper()

	prevStore, prevCfg, prevProxyConfigId := str, cfg, zrokProxyConfigId
	t.Cleanup(func() {
		str, cfg, zrokProxyConfigId = prevStore, prevCfg, prevProxyConfigId
	})

	fake := zitifake.New()
	t.Cleanup(fake.Close)
	cfg = controllerConfig.DefaultConfig()
	cfg.Ziti = &automation.Config{ApiEndpoint: fake.URL, Username: "admin", Password: "admin"}
	zrokProxyConfigId = "zrok-proxy-config-type"

	logs := &lockedBuffer{}
	dl.Init(dl.DefaultOptions().JSON().SetOutput(logs))
	t.Cleanup(func() { dl.Init() })

	var err error
	str, err = store.Open(&store.Config{Path: ":memory:", Type: "sqlite3"})
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, str.Close()) })

	trx, err := str.Begin()
	require.NoError(t, err)
	accountID, err := str.CreateAccount(&store.Account{Email: "owner@example.com", Salt: "salt", Password: "password", Token: "acct-token"}, trx)
	require.NoError(t, err)
	_, err = str.CreateEnvironment(accountID, &store.Environment{Description: "test environment", Host: "host", Address: "address", ZId: "env-zid"}, trx)
	require.NoError(t, err)
	namespaceID, err := str.CreateNamespace(&store.Namespace{Token: "public", Name: "example.com", Description: "public namespace", Open: true}, trx)
	require.NoError(t, err)
	frontendID, err := str.CreateGlobalFrontend(&store.Frontend{Token: "frontend", ZId: "frontend-zid", PermissionMode: store.OpenPermissionMode}, trx)
	require.NoError(t, err)
	_, err = str.CreateNamespaceFrontendMapping(namespaceID, frontendID, true, trx)
	require.NoError(t, err)
	require.NoError(t, trx.Commit())

	return &shareCreateFixture{
		fake:      fake,
		logs:      logs,
		principal: &rest_model_zrok.Principal{ID: int64(accountID), Email: "owner@example.com"},
	}
}

func publicShareRequest(permissionMode string, accessGrants ...string) *rest_model_zrok.ShareRequest {
	return &rest_model_zrok.ShareRequest{
		EnvZID:         "env-zid",
		ShareMode:      "public",
		BackendMode:    "proxy",
		Target:         "http://127.0.0.1:8080",
		AuthScheme:     "none",
		PermissionMode: permissionMode,
		AccessGrants:   accessGrants,
		NameSelections: []*rest_model_zrok.NameSelection{{NamespaceToken: "public"}},
	}
}

func privateShareRequest(token string) *rest_model_zrok.ShareRequest {
	return &rest_model_zrok.ShareRequest{
		EnvZID:            "env-zid",
		ShareMode:         "private",
		BackendMode:       "proxy",
		Target:            "http://127.0.0.1:8080",
		AuthScheme:        "none",
		PermissionMode:    string(store.OpenPermissionMode),
		PrivateShareToken: token,
	}
}

func (f *shareCreateFixture) share(body *rest_model_zrok.ShareRequest) interface{} {
	return newShareHandler().Handle(shareops.ShareParams{Body: body}, f.principal)
}

// recordToken captures the token the handler allocates for, from the name of its service create.
func (f *shareCreateFixture) recordToken() *string {
	token := new(string)
	f.fake.OnBeforeCreate(func(kind, name string) {
		if kind == zitifake.Services {
			*token = name
		}
	})
	return token
}

func (f *shareCreateFixture) requireNoShares(t *testing.T) {
	t.Helper()
	trx, err := str.Begin()
	require.NoError(t, err)
	defer func() { _ = trx.Rollback() }()
	shares, err := str.FindAllShares(trx)
	require.NoError(t, err)
	require.Empty(t, shares)
}

// requireCompensated asserts that every object created through the api was deleted, in reverse order.
func (f *shareCreateFixture) requireCompensated(t *testing.T, token string, wantCreates int) {
	t.Helper()
	created, deleted := f.fake.Log()
	require.Len(t, created, wantCreates)
	reversed := slices.Clone(created)
	slices.Reverse(reversed)
	require.Equal(t, reversed, deleted)
	require.Equal(t, f.fake.ConfigCreates, f.fake.ConfigDeletes)
	require.Equal(t, f.fake.ServiceCreates, f.fake.ServiceDeletes)
	require.Equal(t, f.fake.PolicyCreates, f.fake.PolicyDeletes)
	require.Equal(t, f.fake.SerpCreates, f.fake.SerpDeletes)
	require.Empty(t, f.fake.Tagged("zrokShareToken", token))
}

func TestShareAllocationFailureIsReported(t *testing.T) {
	f := setupShareCreateFixture(t)
	var token string
	f.fake.OnBeforeCreate(func(kind, name string) {
		if kind == zitifake.Services && token == "" {
			// an untagged service with the share's name makes the service create conflict.
			token = name
			f.fake.Seed(zitifake.Services, name, nil)
		}
	})

	resp := f.share(publicShareRequest(string(store.OpenPermissionMode)))

	require.IsType(t, &shareops.ShareInternalServerError{}, resp)
	logs := f.logs.String()
	require.Contains(t, logs, "error allocating share resources")
	require.Contains(t, logs, "error creating service '"+token+"': ziti Bad Request: service name conflict: "+token)
	require.NotContains(t, logs, "chk_z_id")
	f.requireNoShares(t)
	require.Equal(t, 1, f.fake.ConfigCreates)
	require.Equal(t, 1, f.fake.ConfigDeletes)
	require.Equal(t, 0, f.fake.ServiceCreates)
	require.Empty(t, f.fake.Tagged("zrokShareToken", token))
}

func TestShareFailureAfterAllocationCompensatesPrivate(t *testing.T) {
	f := setupShareCreateFixture(t)
	trx, err := str.Begin()
	require.NoError(t, err)
	envs, err := str.FindEnvironmentsForAccount(int(f.principal.ID), trx)
	require.NoError(t, err)
	existingID, err := str.CreateShare(envs[0].Id, &store.Share{ZId: "existing-zid", Token: "duplicate", ShareMode: "private", BackendMode: "proxy", PermissionMode: store.OpenPermissionMode}, trx)
	require.NoError(t, err)
	require.NoError(t, trx.Commit())

	// ziti has no service named 'duplicate', so the availability check passes and the insert fails.
	resp := f.share(privateShareRequest("duplicate"))

	require.IsType(t, &shareops.ShareInternalServerError{}, resp)
	require.Contains(t, f.logs.String(), "error creating share record")
	f.requireCompensated(t, "duplicate", 4)
	trx, err = str.Begin()
	require.NoError(t, err)
	defer func() { _ = trx.Rollback() }()
	existing, err := str.FindShareWithToken("duplicate", trx)
	require.NoError(t, err)
	require.Equal(t, existingID, existing.Id)
	require.Equal(t, "existing-zid", existing.ZId)
}

func TestShareFailureAfterAllocationCompensatesPublic(t *testing.T) {
	f := setupShareCreateFixture(t)
	token := f.recordToken()

	resp := f.share(publicShareRequest(string(store.ClosedPermissionMode), "nobody@example.com"))

	require.IsType(t, &shareops.ShareInternalServerError{}, resp)
	require.Contains(t, f.logs.String(), "unable to find account 'nobody@example.com'")
	f.requireCompensated(t, *token, 5)
	created, _ := f.fake.Log()
	var kinds []string
	for _, entry := range created {
		kind, _, _ := strings.Cut(entry, "/")
		kinds = append(kinds, kind)
	}
	require.Equal(t, []string{zitifake.Configs, zitifake.Services, zitifake.ServicePolicies, zitifake.ServicePolicies, zitifake.ServiceEdgeRouterPolicies}, kinds)
	f.requireNoShares(t)
}

func TestShareSuccessDoesNotCompensate(t *testing.T) {
	f := setupShareCreateFixture(t)

	resp := f.share(publicShareRequest(string(store.OpenPermissionMode)))

	created, ok := resp.(*shareops.ShareCreated)
	require.True(t, ok, "%T", resp)
	token := created.Payload.ShareToken
	createdObjects, deleted := f.fake.Log()
	require.Len(t, createdObjects, 5)
	require.Empty(t, deleted)
	require.Len(t, f.fake.Tagged("zrokShareToken", token), 5)
	trx, err := str.Begin()
	require.NoError(t, err)
	defer func() { _ = trx.Rollback() }()
	shr, err := str.FindShareWithToken(token, trx)
	require.NoError(t, err)
	require.NotEmpty(t, shr.ZId)
}

func TestSharePrivateTokenRaceLeavesWinnerIntact(t *testing.T) {
	f := setupShareCreateFixture(t)
	const token = "contested"
	var winner []string
	f.fake.OnBeforeCreate(func(kind, name string) {
		if kind != zitifake.Configs || winner != nil {
			return
		}
		// the winner allocates between the loser's availability check and its service create. config
		// names are unique in ziti, so the winner's config is left out: this interleaving has the loser's
		// config created and its service create conflicting.
		tags := automation.ZrokShareTags(token).ToRestModel()
		winner = []string{
			zitifake.Services + "/" + f.fake.Seed(zitifake.Services, token, tags),
			zitifake.ServicePolicies + "/" + f.fake.Seed(zitifake.ServicePolicies, "env-winner-bind", tags),
			zitifake.ServiceEdgeRouterPolicies + "/" + f.fake.Seed(zitifake.ServiceEdgeRouterPolicies, "env-winner-serp", tags),
		}
	})

	resp := f.share(privateShareRequest(token))

	require.IsType(t, &shareops.ShareInternalServerError{}, resp)
	require.Contains(t, f.logs.String(), "error creating service '"+token+"': ziti Bad Request: service name conflict: "+token)
	created, deleted := f.fake.Log()
	require.Len(t, created, 1)
	require.Equal(t, created, deleted)
	remaining := f.fake.Tagged("zrokShareToken", token)
	slices.Sort(winner)
	require.Equal(t, winner, remaining)
	f.requireNoShares(t)
}
