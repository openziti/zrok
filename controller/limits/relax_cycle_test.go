package limits

import (
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/openziti/zrok/v2/controller/automation"
	"github.com/openziti/zrok/v2/controller/automation/zitifake"
	"github.com/openziti/zrok/v2/controller/emailUi"
	"github.com/openziti/zrok/v2/controller/metrics"
	"github.com/openziti/zrok/v2/controller/store"
	"github.com/openziti/zrok/v2/sdk/golang/sdk"
	"github.com/stretchr/testify/require"
)

type fixedBandwidthReader struct{}

func (fixedBandwidthReader) totalRxTxForAccount(int64, time.Duration) (int64, int64, error) {
	return 0, 0, nil
}
func (fixedBandwidthReader) totalRxTxForEnvironment(int64, time.Duration) (int64, int64, error) {
	return 0, 0, nil
}
func (fixedBandwidthReader) totalRxTxForShare(string, time.Duration) (int64, int64, error) {
	return 0, 0, nil
}

type relaxFixture struct {
	str                     *store.Store
	agent                   *Agent
	fake                    *zitifake.Server
	badAccount, goodAccount int
}

func newRelaxFixture(t *testing.T) *relaxFixture {
	t.Helper()
	str, err := store.Open(&store.Config{Path: ":memory:", Type: "sqlite3"})
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, str.Close()) })
	fake := zitifake.New()
	t.Cleanup(fake.Close)
	agent, err := NewAgent(DefaultConfig(), &metrics.InfluxConfig{Url: "http://localhost:8086"}, &automation.Config{}, &emailUi.Config{}, str)
	require.NoError(t, err)
	agent.ifx = fixedBandwidthReader{}
	agent.setZitiFactory(func() (*automation.ZitiAutomation, error) {
		return automation.NewZitiAutomationWithEdge(fake.Edge()), nil
	})
	trx, err := str.Begin()
	require.NoError(t, err)
	bad := addLimitedShare(t, str, trx, "bad@example.com", sdk.PublicShareMode)
	good := addLimitedShare(t, str, trx, "good@example.com", sdk.PrivateShareMode)
	require.NoError(t, trx.Commit())
	return &relaxFixture{str: str, agent: agent, fake: fake, badAccount: bad, goodAccount: good}
}

func addLimitedShare(t *testing.T, str *store.Store, trx *sqlx.Tx, email string, mode sdk.ShareMode) int {
	t.Helper()
	acctID, err := str.CreateAccount(&store.Account{Email: email, Salt: "salt", Password: "password", Token: email + "-token"}, trx)
	require.NoError(t, err)
	envID, err := str.CreateEnvironment(acctID, &store.Environment{Description: email, Host: "host", Address: "address", ZId: email + "-env"}, trx)
	require.NoError(t, err)
	shareID, err := str.CreateShare(envID, &store.Share{ZId: email + "-share", Token: email + "-share-token", ShareMode: string(mode), BackendMode: string(sdk.ProxyBackendMode), PermissionMode: store.OpenPermissionMode}, trx)
	require.NoError(t, err)
	if mode == sdk.PrivateShareMode {
		_, err = str.CreateFrontend(envID, &store.Frontend{PrivateShareId: &shareID, Token: email + "-frontend", ZId: email + "-frontend-zid", PermissionMode: store.ClosedPermissionMode}, trx)
		require.NoError(t, err)
	}
	_, err = str.CreateBandwidthLimitJournalEntry(&store.BandwidthLimitJournalEntry{AccountId: acctID, Action: store.LimitLimitAction}, trx)
	require.NoError(t, err)
	return acctID
}

func journalEmpty(t *testing.T, str *store.Store, acctID int) bool {
	t.Helper()
	trx, err := str.Begin()
	require.NoError(t, err)
	defer func() { _ = trx.Rollback() }()
	empty, err := str.IsBandwidthLimitJournalEmptyForGlobal(acctID, trx)
	require.NoError(t, err)
	return empty
}

func TestRelaxRetainsBadShareAndCompletesOtherAccount(t *testing.T) {
	f := newRelaxFixture(t)
	// a successful share on the retained account must not be recreated on retry.
	trx, err := f.str.Begin()
	require.NoError(t, err)
	envs, err := f.str.FindEnvironmentsForAccount(f.badAccount, trx)
	require.NoError(t, err)
	require.Len(t, envs, 1)
	shareID, err := f.str.CreateShare(envs[0].Id, &store.Share{ZId: "second-share", Token: "second-share-token", ShareMode: string(sdk.PrivateShareMode), BackendMode: "proxy", PermissionMode: store.ClosedPermissionMode}, trx)
	require.NoError(t, err)
	_, err = f.str.CreateFrontend(envs[0].Id, &store.Frontend{PrivateShareId: &shareID, Token: "second-frontend", ZId: "second-frontend-zid", PermissionMode: store.ClosedPermissionMode}, trx)
	require.NoError(t, err)
	require.NoError(t, trx.Commit())
	require.NoError(t, f.agent.relax())
	require.False(t, journalEmpty(t, f.str, f.badAccount))
	require.True(t, journalEmpty(t, f.str, f.goodAccount))
	created, _, _, _ := f.fake.Counts()
	require.Equal(t, 2, created)
	policies, err := automation.NewZitiAutomationWithEdge(f.fake.Edge()).ServicePolicies.Find(&automation.FilterOptions{Filter: automation.BuildFilter("name", "good@example.com-frontend-good@example.com-env-good@example.com-share-dial")})
	require.NoError(t, err)
	require.Len(t, policies, 1)
	require.NoError(t, f.agent.relax())
	created, _, _, _ = f.fake.Counts()
	require.Equal(t, 2, created)
}

type panicOnAccount struct {
	delegate  AccountAction
	accountID int
}

func (p panicOnAccount) HandleAccount(acct *store.Account, rx, tx int64, bwc store.BandwidthClass, ul *userLimits, trx *sqlx.Tx) error {
	if acct.Id == p.accountID {
		panic("test relax panic")
	}
	return p.delegate.HandleAccount(acct, rx, tx, bwc, ul, trx)
}

func TestRelaxIsolatesPanickingAccount(t *testing.T) {
	f := newRelaxFixture(t)
	f.agent.relaxActions = []AccountAction{panicOnAccount{delegate: f.agent.relaxActions[0], accountID: f.badAccount}}
	require.NoError(t, f.agent.relax())
	require.False(t, journalEmpty(t, f.str, f.badAccount))
	require.True(t, journalEmpty(t, f.str, f.goodAccount))
}
