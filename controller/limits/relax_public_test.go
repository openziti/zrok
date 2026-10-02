package limits

import (
	"testing"

	"github.com/pkg/errors"

	"github.com/openziti/zrok/v2/controller/store"
	"github.com/openziti/zrok/v2/sdk/golang/sdk"
	"github.com/stretchr/testify/require"
)

// addMappedPublicShare creates a public share, journalled limited or not, mapped to a name in a fresh
// namespace served by frontendZIds.
func (f *relaxFixture) addMappedPublicShare(t *testing.T, email string, limited bool, frontendZIds ...string) *limitedShare {
	t.Helper()
	trx, err := f.str.Begin()
	require.NoError(t, err)
	var ls *limitedShare
	if limited {
		ls = addLimitedShare(t, f.str, trx, email, sdk.PublicShareMode)
	} else {
		ls = addShare(t, f.str, trx, email, sdk.PublicShareMode)
	}
	nsID := addNamespace(t, f.str, trx, email+"-ns", frontendZIds...)
	mapShareToName(t, f.str, trx, ls, nsID, "name")
	require.NoError(t, trx.Commit())
	return ls
}

func (f *relaxFixture) journalLimit(t *testing.T, ls *limitedShare) {
	t.Helper()
	trx, err := f.str.Begin()
	require.NoError(t, err)
	_, err = f.str.CreateBandwidthLimitJournalEntry(&store.BandwidthLimitJournalEntry{AccountId: ls.account, Action: store.LimitLimitAction}, trx)
	require.NoError(t, err)
	require.NoError(t, trx.Commit())
}

func TestRelaxRebuildsPublicDialPolicyFromMappings(t *testing.T) {
	f := newEmptyRelaxFixture(t)
	ls := f.addMappedPublicShare(t, "public@example.com", true, "public", "dynfe")

	require.NoError(t, f.agent.relax())

	policy := f.policy(t, ls.publicDialPolicy())
	require.NotNil(t, policy)
	require.ElementsMatch(t, []string{"@public", "@dynfe"}, policy.IdentityRoles)
	require.Equal(t, []string{"@" + ls.shrZId}, []string(policy.ServiceRoles))
	require.Equal(t, ls.token, policy.Tags.SubTags["zrokShareToken"])
	require.True(t, journalEmpty(t, f.str, ls.account))
}

func TestRelaxV1PublicShareUsesFrontendSelection(t *testing.T) {
	f := newEmptyRelaxFixture(t)
	trx, err := f.str.Begin()
	require.NoError(t, err)
	ls := addLimitedShare(t, f.str, trx, "v1@example.com", sdk.PublicShareMode)
	publicName := "public"
	_, err = f.str.CreateGlobalFrontend(&store.Frontend{Token: "v1-frontend", ZId: "v1-frontend-zid", PublicName: &publicName, PermissionMode: store.OpenPermissionMode}, trx)
	require.NoError(t, err)
	shr, err := f.str.GetShare(ls.share, trx)
	require.NoError(t, err)
	shr.FrontendSelection = &publicName
	require.NoError(t, f.str.UpdateShare(shr, trx))
	require.NoError(t, trx.Commit())

	require.NoError(t, f.agent.relax())

	policy := f.policy(t, ls.publicDialPolicy())
	require.NotNil(t, policy)
	require.Equal(t, []string{"@v1-frontend-zid"}, []string(policy.IdentityRoles))
	require.True(t, journalEmpty(t, f.str, ls.account))
}

func TestRelaxPublicShareWithoutFrontendsCompletes(t *testing.T) {
	f := newEmptyRelaxFixture(t)
	trx, err := f.str.Begin()
	require.NoError(t, err)
	ls := addLimitedShare(t, f.str, trx, "bare@example.com", sdk.PublicShareMode)
	require.NoError(t, trx.Commit())

	require.NoError(t, f.agent.relax())

	require.Equal(t, 0, f.policyCreates())
	require.True(t, journalEmpty(t, f.str, ls.account))
}

func TestRelaxPublicShareRestoresAccessPolicies(t *testing.T) {
	f := newEmptyRelaxFixture(t)
	ls := f.addMappedPublicShare(t, "accessed@example.com", true, "public")
	trx, err := f.str.Begin()
	require.NoError(t, err)
	addAccessFrontend(t, f.str, trx, ls, "accessed-frontend")
	require.NoError(t, trx.Commit())

	require.NoError(t, f.agent.relax())

	require.NotNil(t, f.policy(t, ls.publicDialPolicy()))
	access := f.policy(t, ls.accessDialPolicy("accessed-frontend"))
	require.NotNil(t, access)
	require.Equal(t, []string{"@" + ls.envZId}, []string(access.IdentityRoles))
	require.Equal(t, "accessed-frontend", access.Tags.SubTags["zrokFrontendToken"])
	require.True(t, journalEmpty(t, f.str, ls.account))
}

func TestRelaxIsIdempotent(t *testing.T) {
	f := newEmptyRelaxFixture(t)
	ls := f.addMappedPublicShare(t, "again@example.com", true, "public", "dynfe")
	trx, err := f.str.Begin()
	require.NoError(t, err)
	addAccessFrontend(t, f.str, trx, ls, "again-frontend")
	require.NoError(t, trx.Commit())

	require.NoError(t, f.agent.relax())
	require.Equal(t, 2, f.policyCreates())
	require.True(t, journalEmpty(t, f.str, ls.account))

	// limited again with its policies left in place, as after a cycle that crashed before its commit.
	f.journalLimit(t, ls)
	require.NoError(t, f.agent.relax())
	require.Equal(t, 2, f.policyCreates())
	require.True(t, journalEmpty(t, f.str, ls.account))
}

func TestRelaxV1ShareWithMissingFrontendIsShareFailure(t *testing.T) {
	f := newEmptyRelaxFixture(t)
	trx, err := f.str.Begin()
	require.NoError(t, err)
	v1 := addLimitedShare(t, f.str, trx, "stale@example.com", sdk.PublicShareMode)
	shr, err := f.str.GetShare(v1.share, trx)
	require.NoError(t, err)
	gone := "gone"
	shr.FrontendSelection = &gone
	require.NoError(t, f.str.UpdateShare(shr, trx))
	other := addLimitedShare(t, f.str, trx, "other@example.com", sdk.PrivateShareMode)
	addAccessFrontend(t, f.str, trx, other, "other-frontend")
	require.NoError(t, trx.Commit())

	// the share's failure is reported, and is not a store failure that would end the cycle.
	trx, err = f.str.Begin()
	require.NoError(t, err)
	acct, err := f.str.GetAccount(v1.account, trx)
	require.NoError(t, err)
	err = newRelaxAction(f.str, f.agent.newZiti).HandleAccount(acct, 0, 0, newConfigBandwidthClasses(f.agent.cfg.Bandwidth)[1], nil, trx)
	require.NoError(t, trx.Rollback())
	require.ErrorContains(t, err, "selects frontend 'gone', which does not exist")
	var storeErr storeRelaxError
	require.False(t, errors.As(err, &storeErr))

	require.NoError(t, f.agent.relax())
	require.False(t, journalEmpty(t, f.str, v1.account))
	require.True(t, journalEmpty(t, f.str, other.account))
	require.NotNil(t, f.policy(t, other.accessDialPolicy("other-frontend")))
}
