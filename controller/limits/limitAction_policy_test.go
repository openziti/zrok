package limits

import (
	"testing"

	"github.com/openziti/edge-api/rest_model"
	"github.com/openziti/zrok/v2/controller/automation"
	"github.com/openziti/zrok/v2/controller/automation/zitifake"
	"github.com/openziti/zrok/v2/sdk/golang/sdk"
	"github.com/stretchr/testify/require"
)

// a limit removes the share's dial policy and keeps its bind policy, on either ziti generation.
func TestLimitRemovesOnlyDialPolicy(t *testing.T) {
	for _, gen := range zitifake.Generations {
		t.Run(gen.Name, func(t *testing.T) {
			f := newEmptyRelaxFixture(t)
			f.fake.IntegerTypeMatchesNothing(gen.IntegerTypeMatchesNothing)
			trx, err := f.str.Begin()
			require.NoError(t, err)
			defer func() { _ = trx.Rollback() }()
			ls := addShare(t, f.str, trx, "limited@example.com", sdk.PublicShareMode)
			other := addShare(t, f.str, trx, "other@example.com", sdk.PublicShareMode)
			f.fake.SeedPolicyWithID("dial", "dial", automation.ZrokShareTags(ls.token).ToRestModel(), rest_model.DialBindDial)
			f.fake.SeedPolicyWithID("bind", "bind", automation.ZrokShareTags(ls.token).ToRestModel(), rest_model.DialBindBind)
			f.fake.SeedPolicyWithID("other-dial", "other-dial", automation.ZrokShareTags(other.token).ToRestModel(), rest_model.DialBindDial)
			acct, err := f.str.GetAccount(ls.account, trx)
			require.NoError(t, err)
			bwc := newConfigBandwidthClasses(f.agent.cfg.Bandwidth)[1]

			require.NoError(t, newLimitAction(f.str, f.agent.newZiti).HandleAccount(acct, 0, 0, bwc, &userLimits{}, trx))

			_, deleted := f.fake.Log()
			require.Equal(t, []string{zitifake.ServicePolicies + "/dial"}, deleted)
			require.True(t, f.fake.Has(zitifake.ServicePolicies, "bind"))
			require.True(t, f.fake.Has(zitifake.ServicePolicies, "other-dial"))
		})
	}
}
