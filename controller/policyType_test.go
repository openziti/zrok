package controller

import (
	"sort"
	"testing"

	"github.com/openziti/edge-api/rest_model"
	"github.com/openziti/zrok/v2/controller/automation"
	"github.com/openziti/zrok/v2/controller/automation/zitifake"
	"github.com/openziti/zrok/v2/controller/store"
	"github.com/openziti/zrok/v2/rest_server_zrok/operations/agent"
	"github.com/openziti/zrok/v2/rest_server_zrok/operations/environment"
	shareops "github.com/openziti/zrok/v2/rest_server_zrok/operations/share"
	"github.com/stretchr/testify/require"
)

// fixtureEnvironmentId returns the store id of the fixture's environment.
func (f *shareCreateFixture) fixtureEnvironmentId(t *testing.T) int {
	t.Helper()
	trx, err := str.Begin()
	require.NoError(t, err)
	defer func() { _ = trx.Rollback() }()
	env, err := str.FindEnvironmentForAccount("env-zid", int(f.principal.ID), trx)
	require.NoError(t, err)
	return env.Id
}

// a disable withdraws the access of the environment's frontend by its dial policy alone, on either
// ziti generation.
func TestDisableRemovesOnlyFrontendDialPolicy(t *testing.T) {
	for _, gen := range zitifake.Generations {
		t.Run(gen.Name, func(t *testing.T) {
			f := setupShareCreateFixture(t)
			f.fake.IntegerTypeMatchesNothing(gen.IntegerTypeMatchesNothing)
			envId := f.fixtureEnvironmentId(t)
			trx, err := str.Begin()
			require.NoError(t, err)
			_, err = str.CreateFrontend(envId, &store.Frontend{Token: "env-fe", ZId: "env-fe-zid", PermissionMode: store.OpenPermissionMode}, trx)
			require.NoError(t, err)
			require.NoError(t, trx.Commit())
			f.fake.SeedPolicyWithID("fe-dial", "fe-dial", automation.ZrokTags().WithTag("zrokFrontendToken", "env-fe").ToRestModel(), rest_model.DialBindDial)
			f.fake.SeedPolicyWithID("fe-bind", "fe-bind", automation.ZrokTags().WithTag("zrokFrontendToken", "env-fe").ToRestModel(), rest_model.DialBindBind)
			f.fake.SeedPolicyWithID("other-fe-dial", "other-fe-dial", automation.ZrokTags().WithTag("zrokFrontendToken", "other-fe").ToRestModel(), rest_model.DialBindDial)

			resp := newDisableHandler().Handle(environment.DisableParams{Body: environment.DisableBody{Identity: "env-zid"}}, f.principal)

			require.IsType(t, &environment.DisableOK{}, resp)
			_, deleted := f.fake.Log()
			require.Equal(t, []string{zitifake.ServicePolicies + "/fe-dial"}, deleted)
			require.True(t, f.fake.Has(zitifake.ServicePolicies, "fe-bind"))
			require.True(t, f.fake.Has(zitifake.ServicePolicies, "other-fe-dial"))
		})
	}
}

// an unenroll removes the remote's dial and bind policies, its service edge router policy and its
// service, and nothing of any other remote, on either ziti generation.
func TestAgentUnenrollRemovesOnlyRemotePolicies(t *testing.T) {
	for _, gen := range zitifake.Generations {
		t.Run(gen.Name, func(t *testing.T) {
			f := setupShareCreateFixture(t)
			f.fake.IntegerTypeMatchesNothing(gen.IntegerTypeMatchesNothing)
			envId := f.fixtureEnvironmentId(t)
			trx, err := str.Begin()
			require.NoError(t, err)
			_, err = str.CreateAgentEnrollment(envId, "remote", trx)
			require.NoError(t, err)
			require.NoError(t, trx.Commit())
			remote := automation.ZrokAgentRemoteTags("remote", "env-zid").ToRestModel()
			other := automation.ZrokAgentRemoteTags("other-remote", "other-env").ToRestModel()
			f.fake.SeedPolicyWithID("remote-dial", "remote-dial", remote, rest_model.DialBindDial)
			f.fake.SeedPolicyWithID("remote-bind", "remote-bind", remote, rest_model.DialBindBind)
			f.fake.SeedWithID(zitifake.ServiceEdgeRouterPolicies, "remote-serp", "remote", remote)
			f.fake.SeedWithID(zitifake.Services, "remote-service", "remote", remote)
			f.fake.SeedPolicyWithID("other-dial", "other-dial", other, rest_model.DialBindDial)
			f.fake.SeedPolicyWithID("other-bind", "other-bind", other, rest_model.DialBindBind)
			f.fake.SeedPolicyWithID("share-dial", "share-dial", automation.ZrokShareTags("share").ToRestModel(), rest_model.DialBindDial)

			resp := newAgentUnenrollHandler().Handle(agent.UnenrollParams{Body: agent.UnenrollBody{EnvZID: "env-zid"}}, f.principal)

			require.IsType(t, &agent.UnenrollOK{}, resp)
			_, deleted := f.fake.Log()
			sort.Strings(deleted)
			require.Equal(t, []string{
				zitifake.ServiceEdgeRouterPolicies + "/remote-serp",
				zitifake.ServicePolicies + "/remote-bind",
				zitifake.ServicePolicies + "/remote-dial",
				zitifake.Services + "/remote-service",
			}, deleted)
			require.Equal(t, 3, f.fake.Len(zitifake.ServicePolicies))
		})
	}
}

// an unaccess removes the frontend's dial policy and keeps the share's bind policy, on either ziti
// generation.
func TestUnaccessRemovesOnlyFrontendDialPolicy(t *testing.T) {
	for _, gen := range zitifake.Generations {
		t.Run(gen.Name, func(t *testing.T) {
			f := setupShareCreateFixture(t)
			f.fake.IntegerTypeMatchesNothing(gen.IntegerTypeMatchesNothing)
			f.createPrivateShare(t, "private-access")
			created, ok := f.access(newAccessHandler(), "private-access").(*shareops.AccessCreated)
			require.True(t, ok)
			require.Equal(t, 1, f.accessDialPolicies("private-access"))

			resp := newUnaccessHandler().Handle(shareops.UnaccessParams{Body: shareops.UnaccessBody{EnvZID: "env-zid", ShareToken: "private-access", FrontendToken: created.Payload.FrontendToken}}, f.principal)

			require.IsType(t, &shareops.UnaccessOK{}, resp)
			require.Zero(t, f.accessDialPolicies("private-access"))
			_, deleted := f.fake.Log()
			require.Len(t, deleted, 1)
		})
	}
}
