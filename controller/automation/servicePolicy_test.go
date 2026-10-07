package automation

import (
	"testing"

	"github.com/openziti/edge-api/rest_model"
	"github.com/openziti/zrok/v2/controller/automation/zitifake"
	"github.com/stretchr/testify/require"
)

// typedPolicyFixture holds a dial and a bind policy for share 'one', and a dial policy for share 'two'.
func typedPolicyFixture(t *testing.T, integerTypeMatchesNothing bool) (*zitifake.Server, *ZitiAutomation) {
	t.Helper()
	fake := zitifake.New()
	t.Cleanup(fake.Close)
	fake.IntegerTypeMatchesNothing(integerTypeMatchesNothing)
	fake.SeedPolicyWithID("one-dial", "one-dial", ZrokShareTags("one").ToRestModel(), rest_model.DialBindDial)
	fake.SeedPolicyWithID("one-bind", "one-bind", ZrokShareTags("one").ToRestModel(), rest_model.DialBindBind)
	fake.SeedPolicyWithID("two-dial", "two-dial", ZrokShareTags("two").ToRestModel(), rest_model.DialBindDial)
	return fake, NewZitiAutomationWithEdge(fake.Edge())
}

func TestDeleteByTagAndTypeSelectsByResponseType(t *testing.T) {
	for _, gen := range zitifake.Generations {
		t.Run(gen.Name, func(t *testing.T) {
			fake, ziti := typedPolicyFixture(t, gen.IntegerTypeMatchesNothing)

			require.NoError(t, ziti.ServicePolicies.DeleteByTagAndType(BuildTagFilter("zrokShareToken", "one"), rest_model.DialBindDial))

			_, deleted := fake.Log()
			require.Equal(t, []string{zitifake.ServicePolicies + "/one-dial"}, deleted)

			require.NoError(t, ziti.ServicePolicies.DeleteByTagAndType(BuildTagFilter("zrokShareToken", "one"), rest_model.DialBindBind))

			_, deleted = fake.Log()
			require.Equal(t, []string{zitifake.ServicePolicies + "/one-dial", zitifake.ServicePolicies + "/one-bind"}, deleted)
			require.True(t, fake.Has(zitifake.ServicePolicies, "two-dial"))
		})
	}
}

func TestDeleteByTagAndTypeContinuesPastVanished(t *testing.T) {
	fake, ziti := typedPolicyFixture(t, false)
	fake.SeedPolicyWithID("one-dial-second", "one-dial-second", ZrokShareTags("one").ToRestModel(), rest_model.DialBindDial)
	// the first dial policy disappears between the listing and its delete.
	fake.VanishBeforeDelete(zitifake.ServicePolicies, "one-dial")

	require.NoError(t, ziti.ServicePolicies.DeleteByTagAndType(BuildTagFilter("zrokShareToken", "one"), rest_model.DialBindDial))

	_, deleted := fake.Log()
	require.Equal(t, []string{zitifake.ServicePolicies + "/one-dial-second"}, deleted)
	require.False(t, fake.Has(zitifake.ServicePolicies, "one-dial"))
	require.True(t, fake.Has(zitifake.ServicePolicies, "one-bind"))
}

func TestDeleteByTagAndTypePropagatesOtherErrors(t *testing.T) {
	fake, ziti := typedPolicyFixture(t, false)
	fake.RejectDeletes(zitifake.ServicePolicies, true)

	err := ziti.ServicePolicies.DeleteByTagAndType(BuildTagFilter("zrokShareToken", "one"), rest_model.DialBindDial)

	require.ErrorContains(t, err, "status 500")
	require.False(t, IsNotFound(err))
	require.True(t, fake.Has(zitifake.ServicePolicies, "one-dial"))
}
