package limits

import (
	"bytes"
	"testing"

	"github.com/openziti/zrok/v2/sdk/golang/sdk"
	"github.com/stretchr/testify/require"
)

func TestRepairDialPolicies(t *testing.T) {
	f := newEmptyRelaxFixture(t)
	unlimited := f.addMappedPublicShare(t, "unlimited@example.com", false, "public")
	limited := f.addMappedPublicShare(t, "limited@example.com", true, "public-limited")
	healthy := f.addMappedPublicShare(t, "healthy@example.com", false, "public-healthy")

	trx, err := f.str.Begin()
	require.NoError(t, err)
	shr, err := f.str.GetShare(healthy.share, trx)
	require.NoError(t, err)
	desired, err := desiredDialPolicies(f.str, shr, trx)
	require.NoError(t, err)
	require.NoError(t, trx.Rollback())
	require.Len(t, desired, 1)
	_, err = ensureDialPolicy(f.ziti(), desired[0])
	require.NoError(t, err)
	baseline := f.policyCreates()

	var out bytes.Buffer
	rpt, err := RepairDialPolicies(f.str, f.ziti(), false, &out)
	require.NoError(t, err)
	require.Len(t, rpt.Missing, 1)
	require.Equal(t, unlimited.publicDialPolicy(), rpt.Missing[0].Name)
	require.Equal(t, unlimited.token, rpt.Missing[0].ShareToken)
	require.Equal(t, []string{"limited@example.com"}, rpt.SkippedAccounts)
	require.Equal(t, 2, rpt.Checked)
	require.Equal(t, 0, rpt.Created)
	require.Equal(t, baseline, f.policyCreates())
	require.Contains(t, out.String(), unlimited.publicDialPolicy())
	require.Contains(t, out.String(), "skipped limited account 'limited@example.com'")

	rpt, err = RepairDialPolicies(f.str, f.ziti(), true, &out)
	require.NoError(t, err)
	require.Len(t, rpt.Missing, 1)
	require.Equal(t, 1, rpt.Created)
	require.Equal(t, 0, rpt.Failed)
	require.Equal(t, baseline+1, f.policyCreates())
	require.NotNil(t, f.policy(t, unlimited.publicDialPolicy()))
	require.Nil(t, f.policy(t, limited.publicDialPolicy()))
	require.False(t, journalEmpty(t, f.str, limited.account))

	rpt, err = RepairDialPolicies(f.str, f.ziti(), true, &out)
	require.NoError(t, err)
	require.Empty(t, rpt.Missing)
	require.Equal(t, 0, rpt.Created)
	require.Equal(t, baseline+1, f.policyCreates())
}

func TestRepairDialPoliciesReportsFailures(t *testing.T) {
	f := newEmptyRelaxFixture(t)
	failing := f.addMappedPublicShare(t, "failing@example.com", false, "public")
	working := f.addMappedPublicShare(t, "working@example.com", false, "public-working")
	f.fake.RejectCreates(failing.publicDialPolicy(), true)

	var out bytes.Buffer
	rpt, err := RepairDialPolicies(f.str, f.ziti(), true, &out)
	require.Error(t, err)
	require.Len(t, rpt.Missing, 2)
	require.Equal(t, 1, rpt.Created)
	require.Equal(t, 1, rpt.Failed)
	require.NotNil(t, f.policy(t, working.publicDialPolicy()))
}

func TestRepairDialPoliciesReportsMissingFrontend(t *testing.T) {
	f := newEmptyRelaxFixture(t)
	unlimited := f.addMappedPublicShare(t, "unlimited@example.com", false, "public")
	trx, err := f.str.Begin()
	require.NoError(t, err)
	stale := addShare(t, f.str, trx, "stale@example.com", sdk.PublicShareMode)
	shr, err := f.str.GetShare(stale.share, trx)
	require.NoError(t, err)
	gone := "gone"
	shr.FrontendSelection = &gone
	require.NoError(t, f.str.UpdateShare(shr, trx))
	require.NoError(t, trx.Commit())

	var out bytes.Buffer
	rpt, err := RepairDialPolicies(f.str, f.ziti(), true, &out)
	require.Error(t, err)
	require.Equal(t, 1, rpt.Failed)
	require.Len(t, rpt.FailedShares, 1)
	require.Equal(t, "stale@example.com", rpt.FailedShares[0].Account)
	require.Equal(t, stale.token, rpt.FailedShares[0].ShareToken)
	require.Equal(t, "gone", rpt.FailedShares[0].SelectedFrontend)
	require.Contains(t, out.String(), "account 'stale@example.com', share '"+stale.token+"', selects missing frontend 'gone'")
	require.Len(t, rpt.Missing, 1)
	require.Equal(t, unlimited.publicDialPolicy(), rpt.Missing[0].Name)
	require.Equal(t, 1, rpt.Created)
	require.NotNil(t, f.policy(t, unlimited.publicDialPolicy()))
}
