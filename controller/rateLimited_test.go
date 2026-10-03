package controller

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/runtime/middleware"
	"github.com/jmoiron/sqlx"
	"github.com/openziti/sdk-golang/ziti"
	"github.com/openziti/zrok/v2/controller/automation"
	"github.com/openziti/zrok/v2/controller/automation/zitifake"
	"github.com/openziti/zrok/v2/controller/store"
	"github.com/openziti/zrok/v2/rest_server_zrok/operations/environment"
	shareops "github.com/openziti/zrok/v2/rest_server_zrok/operations/share"
	"github.com/stretchr/testify/require"
)

// requireRateLimitedResponse writes resp and asserts the 503 a client sees, with its Retry-After.
func requireRateLimitedResponse(t *testing.T, resp interface{}) {
	t.Helper()
	responder, ok := resp.(middleware.Responder)
	require.True(t, ok, "%T", resp)
	rec := httptest.NewRecorder()
	responder.WriteResponse(rec, runtime.JSONProducer())
	require.Equal(t, http.StatusServiceUnavailable, rec.Code, "%T", resp)
	require.Equal(t, "5", rec.Header().Get("Retry-After"))
}

func (f *shareCreateFixture) access(h *accessHandler, token string) interface{} {
	return h.Handle(shareops.AccessParams{Body: shareops.AccessBody{EnvZID: "env-zid", ShareToken: token}}, f.principal)
}

func (f *shareCreateFixture) createPrivateShare(t *testing.T, token string) {
	t.Helper()
	resp := f.share(privateShareRequest(token))
	require.IsType(t, &shareops.ShareCreated{}, resp)
}

// accessDialPolicies counts the access dial policies the fake holds for a private share: its service
// policies less the share's own bind policy.
func (f *shareCreateFixture) accessDialPolicies(token string) int {
	policies := 0
	for _, entry := range f.fake.Tagged("zrokShareToken", token) {
		if strings.HasPrefix(entry, zitifake.ServicePolicies+"/") {
			policies++
		}
	}
	return policies - 1
}

// environmentFrontends returns the frontends recorded for the fixture's environment.
func (f *shareCreateFixture) environmentFrontends(t *testing.T) []*store.Frontend {
	t.Helper()
	trx, err := str.Begin()
	require.NoError(t, err)
	defer func() { _ = trx.Rollback() }()
	env, err := str.FindEnvironmentForAccount("env-zid", int(f.principal.ID), trx)
	require.NoError(t, err)
	fes, err := str.FindFrontendsForEnvironment(env.Id, trx)
	require.NoError(t, err)
	return fes
}

func TestShareRateLimitedAnswersServiceUnavailable(t *testing.T) {
	f := setupShareCreateFixture(t)
	token := f.recordToken()
	f.fake.RateLimitCreates(zitifake.ServicePolicies, true)

	resp := f.share(publicShareRequest(string(store.OpenPermissionMode)))

	requireRateLimitedResponse(t, resp)
	require.Contains(t, f.logs.String(), "ziti rate limited by 'SERVER_TOO_MANY_REQUESTS' on 'POST /edge/management/v1/service-policies'")
	f.requireNoShares(t)
	// the config and service created before the 429 are compensated.
	f.requireCompensated(t, *token, 2)
}

func TestUnshareRateLimitedRollsBack(t *testing.T) {
	f := setupShareCreateFixture(t)
	addDynamicFrontend(t, "dynamic-fe")
	installRecordingPublisher(t)
	token := f.createPublicShare(t)
	for _, kind := range []string{zitifake.Configs, zitifake.Services, zitifake.ServicePolicies, zitifake.ServiceEdgeRouterPolicies} {
		f.fake.RateLimitDeletes(kind, true)
	}

	resp := f.unshare(token)

	requireRateLimitedResponse(t, resp)
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

func TestAccessCommitFailureCompensates(t *testing.T) {
	f := setupShareCreateFixture(t)
	f.createPrivateShare(t, "private-access")
	h := newAccessHandler()
	h.commit = func(*sqlx.Tx) error { return errors.New("commit failed") }

	resp := f.access(h, "private-access")

	require.IsType(t, &shareops.AccessInternalServerError{}, resp)
	require.Zero(t, f.accessDialPolicies("private-access"))
	created, deleted := f.fake.Log()
	require.Equal(t, created[len(created)-1:], deleted)
	require.Empty(t, f.environmentFrontends(t))
}

func TestAccessSuccessKeepsDialPolicy(t *testing.T) {
	f := setupShareCreateFixture(t)
	f.createPrivateShare(t, "private-access")

	resp := f.access(newAccessHandler(), "private-access")

	created, ok := resp.(*shareops.AccessCreated)
	require.True(t, ok, "%T", resp)
	require.Equal(t, 1, f.accessDialPolicies("private-access"))
	require.Len(t, f.fake.Tagged("zrokFrontendToken", created.Payload.FrontendToken), 1)
	_, deleted := f.fake.Log()
	require.Empty(t, deleted)
}

func TestAccessRateLimitedAnswersServiceUnavailable(t *testing.T) {
	f := setupShareCreateFixture(t)
	f.createPrivateShare(t, "private-access")
	f.fake.RateLimitCreates(zitifake.ServicePolicies, true)

	resp := f.access(newAccessHandler(), "private-access")

	requireRateLimitedResponse(t, resp)
	require.Zero(t, f.accessDialPolicies("private-access"))
	require.Empty(t, f.environmentFrontends(t))
}

func TestUnaccessRateLimitedAnswersServiceUnavailable(t *testing.T) {
	f := setupShareCreateFixture(t)
	f.createPrivateShare(t, "private-access")
	created, ok := f.access(newAccessHandler(), "private-access").(*shareops.AccessCreated)
	require.True(t, ok)
	f.fake.RateLimitDeletes(zitifake.ServicePolicies, true)

	resp := newUnaccessHandler().Handle(shareops.UnaccessParams{Body: shareops.UnaccessBody{EnvZID: "env-zid", ShareToken: "private-access", FrontendToken: created.Payload.FrontendToken}}, f.principal)

	requireRateLimitedResponse(t, resp)
	require.Equal(t, 1, f.accessDialPolicies("private-access"))
	require.Len(t, f.environmentFrontends(t), 1)
}

func TestEnableRateLimitedAnswersServiceUnavailable(t *testing.T) {
	f := setupShareCreateFixture(t)
	f.fake.RateLimitCreates(zitifake.Identities, true)

	resp := newEnableHandler().Handle(environment.EnableParams{HTTPRequest: httptest.NewRequest(http.MethodPost, "/enable", nil), Body: environment.EnableBody{Description: "rate limited", Host: "host"}}, f.principal)

	requireRateLimitedResponse(t, resp)
}

func TestDisableRateLimitedRollsBack(t *testing.T) {
	f := setupDisableFixture(t)
	f.fake.SeedWithID(zitifake.Identities, "env-zid", "env-zid", nil)
	f.fake.RateLimitDeletes(zitifake.Identities, true)

	resp := f.disable()

	requireRateLimitedResponse(t, resp)
	require.True(t, f.fake.Has(zitifake.Identities, "env-zid"))
	f.requireEnvironmentRemoved(t, false)
}

// testEnableHandler skips enrollment, which needs a real enrollment endpoint the fake does not serve.
func testEnableHandler() *enableHandler {
	h := newEnableHandler()
	h.enroll = func(*automation.ZitiAutomation, string) (*ziti.Config, error) { return &ziti.Config{}, nil }
	return h
}

func (f *shareCreateFixture) enable(h *enableHandler) interface{} {
	return h.Handle(environment.EnableParams{HTTPRequest: httptest.NewRequest(http.MethodPost, "/enable", nil), Body: environment.EnableBody{Description: "compensated", Host: "host"}}, f.principal)
}

// requireEnvironmentCompensated asserts the enable created an identity and deleted everything it created.
func (f *shareCreateFixture) requireEnvironmentCompensated(t *testing.T, wantCreated []string) {
	t.Helper()
	created, deleted := f.fake.Log()
	var kinds []string
	for _, entry := range created {
		kind, _, _ := strings.Cut(entry, "/")
		kinds = append(kinds, kind)
	}
	require.Equal(t, wantCreated, kinds)
	reversed := slices.Clone(created)
	slices.Reverse(reversed)
	require.Equal(t, reversed, deleted)
	require.Zero(t, f.fake.Len(zitifake.Identities))
	require.Zero(t, f.fake.Len(zitifake.EdgeRouterPolicies))
	trx, err := str.Begin()
	require.NoError(t, err)
	defer func() { _ = trx.Rollback() }()
	envs, err := str.FindEnvironmentsForAccount(int(f.principal.ID), trx)
	require.NoError(t, err)
	require.Len(t, envs, 1)
}

func TestEnableCommitFailureCompensates(t *testing.T) {
	f := setupShareCreateFixture(t)
	h := testEnableHandler()
	h.commit = func(*sqlx.Tx) error { return errors.New("commit failed") }

	resp := f.enable(h)

	require.IsType(t, &environment.EnableInternalServerError{}, resp)
	f.requireEnvironmentCompensated(t, []string{zitifake.Identities, zitifake.EdgeRouterPolicies})
	require.Contains(t, f.logs.String(), "compensated failed environment '")
}

func TestEnableRateLimitedPolicyCompensates(t *testing.T) {
	f := setupShareCreateFixture(t)
	f.fake.RateLimitCreates(zitifake.EdgeRouterPolicies, true)

	resp := f.enable(testEnableHandler())

	requireRateLimitedResponse(t, resp)
	f.requireEnvironmentCompensated(t, []string{zitifake.Identities})
}

func TestEnableSuccessKeepsIdentity(t *testing.T) {
	f := setupShareCreateFixture(t)

	resp := f.enable(testEnableHandler())

	created, ok := resp.(*environment.EnableCreated)
	require.True(t, ok, "%T", resp)
	require.True(t, f.fake.Has(zitifake.Identities, created.Payload.Identity))
	require.Equal(t, 1, f.fake.Len(zitifake.EdgeRouterPolicies))
	_, deleted := f.fake.Log()
	require.Empty(t, deleted)
}
