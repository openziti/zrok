package limits

import (
	"context"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/openziti/edge-api/rest_model"
	"github.com/openziti/zrok/v2/controller/automation"
	"github.com/openziti/zrok/v2/controller/automation/zitifake"
	"github.com/openziti/zrok/v2/controller/emailUi"
	"github.com/openziti/zrok/v2/controller/metrics"
	"github.com/openziti/zrok/v2/controller/store"
	"github.com/openziti/zrok/v2/sdk/golang/sdk"
	"github.com/stretchr/testify/require"
)

type fixedBandwidthReader struct{}

func (fixedBandwidthReader) totalRxTxForAccount(context.Context, int64, time.Duration) (int64, int64, error) {
	return 0, 0, nil
}
func (fixedBandwidthReader) totalRxTxForEnvironment(context.Context, int64, time.Duration) (int64, int64, error) {
	return 0, 0, nil
}
func (fixedBandwidthReader) totalRxTxForShare(context.Context, string, time.Duration) (int64, int64, error) {
	return 0, 0, nil
}

type relaxFixture struct {
	str                     *store.Store
	agent                   *Agent
	fake                    *zitifake.Server
	badAccount, goodAccount int
	bad, good               *limitedShare
}

// newRelaxFixture holds two limited accounts. the bad account's public share is mapped to a name in a
// namespace with one frontend, and the fake refuses to create its dial policy; the good account's private
// share has one access.
func newRelaxFixture(t *testing.T) *relaxFixture {
	t.Helper()
	f := newEmptyRelaxFixture(t)
	trx, err := f.str.Begin()
	require.NoError(t, err)
	f.bad = addLimitedShare(t, f.str, trx, "bad@example.com", sdk.PublicShareMode)
	nsID := addNamespace(t, f.str, trx, "bad-ns", "bad-public")
	mapShareToName(t, f.str, trx, f.bad, nsID, "bad")
	f.good = addLimitedShare(t, f.str, trx, "good@example.com", sdk.PrivateShareMode)
	addAccessFrontend(t, f.str, trx, f.good, "good@example.com-frontend")
	require.NoError(t, trx.Commit())
	f.fake.RejectCreates(f.bad.publicDialPolicy(), true)
	f.badAccount, f.goodAccount = f.bad.account, f.good.account
	return f
}

// newEmptyRelaxFixture is an agent over an empty store and an empty ziti fake.
func newEmptyRelaxFixture(t *testing.T) *relaxFixture {
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
	return &relaxFixture{str: str, agent: agent, fake: fake}
}

func (f *relaxFixture) ziti() *automation.ZitiAutomation {
	return automation.NewZitiAutomationWithEdge(f.fake.Edge())
}

// policy returns the service policy named name, or nil.
func (f *relaxFixture) policy(t *testing.T, name string) *rest_model.ServicePolicyDetail {
	t.Helper()
	policies, err := f.ziti().ServicePolicies.Find(&automation.FilterOptions{Filter: automation.BuildFilter("name", name)})
	require.NoError(t, err)
	require.LessOrEqual(t, len(policies), 1)
	if len(policies) == 0 {
		return nil
	}
	return policies[0]
}

func (f *relaxFixture) policyCreates() int {
	created, _, _, _ := f.fake.Counts()
	return created
}

type limitedShare struct {
	email                 string
	account, env, share   int
	envZId, shrZId, token string
}

func (ls *limitedShare) publicDialPolicy() string { return ls.envZId + "-" + ls.shrZId + "-dial" }

func (ls *limitedShare) accessDialPolicy(feToken string) string {
	return feToken + "-" + ls.envZId + "-" + ls.shrZId + "-dial"
}

// addLimitedShare creates an account with one environment and one share, and journals a global limit for
// the account.
func addLimitedShare(t *testing.T, str *store.Store, trx *sqlx.Tx, email string, mode sdk.ShareMode) *limitedShare {
	t.Helper()
	ls := addShare(t, str, trx, email, mode)
	_, err := str.CreateBandwidthLimitJournalEntry(&store.BandwidthLimitJournalEntry{AccountId: ls.account, Action: store.LimitLimitAction}, trx)
	require.NoError(t, err)
	return ls
}

func addShare(t *testing.T, str *store.Store, trx *sqlx.Tx, email string, mode sdk.ShareMode) *limitedShare {
	t.Helper()
	ls := &limitedShare{email: email, envZId: email + "-env", shrZId: email + "-share", token: email + "-share-token"}
	var err error
	ls.account, err = str.CreateAccount(&store.Account{Email: email, Salt: "salt", Password: "password", Token: email + "-token"}, trx)
	require.NoError(t, err)
	ls.env, err = str.CreateEnvironment(ls.account, &store.Environment{Description: email, Host: "host", Address: "address", ZId: ls.envZId}, trx)
	require.NoError(t, err)
	ls.share, err = str.CreateShare(ls.env, &store.Share{ZId: ls.shrZId, Token: ls.token, ShareMode: string(mode), BackendMode: string(sdk.ProxyBackendMode), PermissionMode: store.OpenPermissionMode}, trx)
	require.NoError(t, err)
	return ls
}

// addNamespace creates a namespace served by a global frontend for each of frontendZIds.
func addNamespace(t *testing.T, str *store.Store, trx *sqlx.Tx, token string, frontendZIds ...string) int {
	t.Helper()
	nsID, err := str.CreateNamespace(&store.Namespace{Token: token, Name: token + ".example.com"}, trx)
	require.NoError(t, err)
	for _, zId := range frontendZIds {
		feID, err := str.CreateGlobalFrontend(&store.Frontend{Token: zId + "-token", ZId: zId, PermissionMode: store.OpenPermissionMode}, trx)
		require.NoError(t, err)
		_, err = str.CreateNamespaceFrontendMapping(nsID, feID, false, trx)
		require.NoError(t, err)
	}
	return nsID
}

// mapShareToName creates a name in the namespace and maps the share to it, as the v2 share path does.
func mapShareToName(t *testing.T, str *store.Store, trx *sqlx.Tx, ls *limitedShare, nsID int, name string) {
	t.Helper()
	nameID, err := str.CreateName(&store.Name{NamespaceId: nsID, Name: name, AccountId: ls.account}, trx)
	require.NoError(t, err)
	_, err = str.CreateShareNameMapping(&store.ShareNameMapping{ShareId: ls.share, NameId: nameID}, trx)
	require.NoError(t, err)
}

// addAccessFrontend records a private access to the share from the share's own environment.
func addAccessFrontend(t *testing.T, str *store.Store, trx *sqlx.Tx, ls *limitedShare, feToken string) {
	t.Helper()
	_, err := str.CreateFrontend(ls.env, &store.Frontend{PrivateShareId: &ls.share, Token: feToken, ZId: ls.envZId, PermissionMode: store.ClosedPermissionMode}, trx)
	require.NoError(t, err)
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
