package limits

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/michaelquigley/df/dd"
	"github.com/michaelquigley/df/dl"
	"github.com/openziti/zrok/v2/controller/automation"
	"github.com/openziti/zrok/v2/controller/emailUi"
	"github.com/openziti/zrok/v2/controller/metrics"
	"github.com/openziti/zrok/v2/controller/store"
	"github.com/openziti/zrok/v2/sdk/golang/sdk"
	"github.com/stretchr/testify/require"
)

// these fixtures prove the fast path makes no store read by pointing the agent at a closed store before an
// event the fast path should finish: every store call on it fails, so an event that finishes without error,
// is counted as finished by the fast path and makes no reader call made no store read. full-path runs are
// counted by the agent's own report counters. time is injected through the caches' clock.

const testPeriod = 60 * time.Minute

type countingReader struct {
	total int64
	err   error
	calls int
}

func (r *countingReader) totalRxTxForAccount(context.Context, int64, time.Duration) (int64, int64, error) {
	r.calls++
	if r.err != nil {
		return 0, 0, r.err
	}
	return r.total, 0, nil
}
func (r *countingReader) totalRxTxForEnvironment(context.Context, int64, time.Duration) (int64, int64, error) {
	panic("not used by enforcement")
}
func (r *countingReader) totalRxTxForShare(context.Context, string, time.Duration) (int64, int64, error) {
	panic("not used by enforcement")
}

type recordingAction struct{ classes []store.BandwidthClass }

func (r *recordingAction) HandleAccount(_ *store.Account, _, _ int64, bwc store.BandwidthClass, _ *userLimits, _ *sqlx.Tx) error {
	r.classes = append(r.classes, bwc)
	return nil
}

type fastPathFixture struct {
	t        *testing.T
	str      *store.Store
	closed   *store.Store
	agent    *Agent
	reader   *countingReader
	warnings *recordingAction
	limits   *recordingAction
	now      time.Time
	acct     int
}

const (
	proxyShare = "proxy-share"
	tcpShare   = "tcp-share"
)

// newFastPathFixture is an account with a proxy share and a tcpTunnel share, under global bandwidth classes
// with the given totals over testPeriod.
func newFastPathFixture(t *testing.T, warningTotal, limitTotal int64, tweak func(*Config)) *fastPathFixture {
	t.Helper()
	str, err := store.Open(&store.Config{Path: ":memory:", Type: "sqlite3"})
	require.NoError(t, err)
	t.Cleanup(func() { _ = str.Close() })
	closed, err := store.Open(&store.Config{Path: ":memory:", Type: "sqlite3"})
	require.NoError(t, err)
	require.NoError(t, closed.Close())

	cfg := DefaultConfig()
	cfg.Bandwidth = &BandwidthPerPeriod{
		Period:  testPeriod,
		Warning: &Bandwidth{Rx: store.Unlimited, Tx: store.Unlimited, Total: warningTotal},
		Limit:   &Bandwidth{Rx: store.Unlimited, Tx: store.Unlimited, Total: limitTotal},
	}
	if tweak != nil {
		tweak(cfg)
	}
	a, err := NewAgent(cfg, &metrics.InfluxConfig{Url: "http://localhost:8086"}, &automation.Config{}, &emailUi.Config{}, str)
	require.NoError(t, err)
	f := &fastPathFixture{t: t, str: str, closed: closed, agent: a, reader: &countingReader{}, warnings: &recordingAction{}, limits: &recordingAction{}, now: time.Unix(1_000_000, 0)}
	a.ifx = f.reader
	a.warningActions = []AccountAction{f.warnings}
	a.limitActions = []AccountAction{f.limits}
	a.caches.now = func() time.Time { return f.now }

	trx, err := str.Begin()
	require.NoError(t, err)
	defer func() { _ = trx.Rollback() }()
	f.acct, err = str.CreateAccount(&store.Account{Email: "user@example.com", Salt: "salt", Password: "password", Token: "user-token"}, trx)
	require.NoError(t, err)
	env, err := str.CreateEnvironment(f.acct, &store.Environment{Description: "env", Host: "host", Address: "address", ZId: "env-zid"}, trx)
	require.NoError(t, err)
	_, err = str.CreateShare(env, &store.Share{ZId: "proxy-zid", Token: proxyShare, ShareMode: string(sdk.PublicShareMode), BackendMode: string(sdk.ProxyBackendMode), PermissionMode: store.OpenPermissionMode}, trx)
	require.NoError(t, err)
	_, err = str.CreateShare(env, &store.Share{ZId: "tcp-zid", Token: tcpShare, ShareMode: string(sdk.PrivateShareMode), BackendMode: string(sdk.TcpTunnelBackendMode), PermissionMode: store.OpenPermissionMode}, trx)
	require.NoError(t, err)
	require.NoError(t, trx.Commit())
	return f
}

func (f *fastPathFixture) advance(d time.Duration) { f.now = f.now.Add(d) }

func (f *fastPathFixture) usage(token string, mode sdk.BackendMode) *metrics.Usage {
	return &metrics.Usage{AccountId: int64(f.acct), ShareToken: token, BackendMode: string(mode)}
}

func (f *fastPathFixture) proxy() *metrics.Usage { return f.usage(proxyShare, sdk.ProxyBackendMode) }
func (f *fastPathFixture) tcp() *metrics.Usage   { return f.usage(tcpShare, sdk.TcpTunnelBackendMode) }

// fast requires the fast path to finish the event with no store read and no reader call.
func (f *fastPathFixture) fast(u *metrics.Usage) {
	f.t.Helper()
	live := f.agent.str
	f.agent.str = f.closed
	defer func() { f.agent.str = live }()
	finished, full, calls := f.agent.caches.finished, f.agent.caches.full, f.reader.calls
	require.NoError(f.t, f.agent.handleUsage(u))
	require.Equal(f.t, finished+1, f.agent.caches.finished, "expected the fast path to finish the event")
	require.Equal(f.t, full, f.agent.caches.full)
	require.Equal(f.t, calls, f.reader.calls, "the fast path must not call the reader")
}

// full requires the event to take the full path against the live store.
func (f *fastPathFixture) full(u *metrics.Usage) {
	f.t.Helper()
	full := f.agent.caches.full
	require.NoError(f.t, f.agent.handleUsage(u))
	require.Equal(f.t, full+1, f.agent.caches.full, "expected the event to take the full path")
}

func (f *fastPathFixture) exec(query string, args ...any) {
	f.t.Helper()
	trx, err := f.str.Begin()
	require.NoError(f.t, err)
	defer func() { _ = trx.Rollback() }()
	_, err = trx.Exec(query, args...)
	require.NoError(f.t, err)
	require.NoError(f.t, trx.Commit())
}

// applyClass creates a bandwidth limit class over testPeriod and applies it to the account.
func (f *fastPathFixture) applyClass(mode *sdk.BackendMode, action store.LimitAction, total int64) int {
	f.t.Helper()
	trx, err := f.str.Begin()
	require.NoError(f.t, err)
	defer func() { _ = trx.Rollback() }()
	lcId, err := f.str.CreateLimitClass(&store.LimitClass{
		BackendMode: mode, Environments: store.Unlimited, Shares: store.Unlimited, ReservedShares: store.Unlimited,
		UniqueNames: store.Unlimited, ShareFrontends: store.Unlimited, PeriodMinutes: int(testPeriod.Minutes()),
		RxBytes: store.Unlimited, TxBytes: store.Unlimited, TotalBytes: total, LimitAction: action,
	}, trx)
	require.NoError(f.t, err)
	_, err = f.str.ApplyLimitClass(&store.AppliedLimitClass{AccountId: f.acct, LimitClassId: lcId}, trx)
	require.NoError(f.t, err)
	require.NoError(f.t, trx.Commit())
	return lcId
}

func (f *fastPathFixture) journal() []*store.BandwidthLimitJournalEntry {
	f.t.Helper()
	trx, err := f.str.Begin()
	require.NoError(f.t, err)
	defer func() { _ = trx.Rollback() }()
	entries, err := f.str.FindAllBandwidthLimitJournal(trx)
	require.NoError(f.t, err)
	return entries
}

func modePtr(mode sdk.BackendMode) *sdk.BackendMode { return &mode }

func limitClassId(bwc store.BandwidthClass) int {
	if bwc.IsGlobal() {
		return 0
	}
	return bwc.GetLimitClassId()
}

func TestFastPathNothingToDoCostsNothing(t *testing.T) {
	f := newFastPathFixture(t, 100, 200, nil)
	f.reader.total = 50

	f.full(f.proxy())
	require.Equal(t, 1, f.reader.calls)
	for i := 0; i < 5; i++ {
		f.advance(10 * time.Second)
		f.fast(f.proxy())
	}
	f.advance(10 * time.Second)
	f.full(f.proxy())
	require.Equal(t, 2, f.reader.calls)
	require.Empty(t, f.journal())

	t.Run("limitless", func(t *testing.T) {
		f.exec("update accounts set limitless = true where id = $1", f.acct)
		f.advance(time.Minute)
		f.full(f.proxy())
		f.reader.total = 1 << 40
		f.fast(f.proxy())
		f.fast(f.tcp())
	})
}

func TestActingPathReadsFresh(t *testing.T) {
	longUsage := func(cfg *Config) { cfg.UsageCacheLifetime = 10 * time.Minute }

	t.Run("lowered class", func(t *testing.T) {
		f := newFastPathFixture(t, 100, 200, longUsage)
		f.applyClass(nil, store.WarningLimitAction, 900)
		lcId := f.applyClass(nil, store.LimitLimitAction, 1000)
		f.reader.total = 500
		f.full(f.proxy())

		f.exec("update limit_classes set total_bytes = 400 where id = $1", lcId)
		// until the account entry expires the cached class decides only that there is nothing to do.
		f.advance(30 * time.Second)
		f.fast(f.proxy())
		require.Empty(t, f.journal())

		f.advance(30 * time.Second)
		f.full(f.proxy())
		require.Len(t, f.limits.classes, 1)
		require.Equal(t, lcId, limitClassId(f.limits.classes[0]))
		require.EqualValues(t, 400, f.limits.classes[0].GetTotalBytes(), "the action uses the class read from the store")
		require.Len(t, f.journal(), 1)
	})

	// a wrong implementation would hand the cached class to enforcement, and journal and act on it.
	t.Run("raised class", func(t *testing.T) {
		f := newFastPathFixture(t, 900, 1000, nil)
		lcId := f.applyClass(modePtr(sdk.ProxyBackendMode), store.LimitLimitAction, 400)
		f.reader.total = 500
		// the tcpTunnel event caches the account with the proxy class, and totals over it, without crossing it.
		f.full(f.tcp())
		require.Empty(t, f.journal())

		f.exec("update limit_classes set total_bytes = 1000 where id = $1", lcId)
		f.full(f.proxy())
		require.Empty(t, f.journal())
		require.Empty(t, f.limits.classes)
		require.Empty(t, f.warnings.classes)
	})

	// a wrong implementation would judge from the cached totals, and journal and act on them.
	t.Run("lowered totals", func(t *testing.T) {
		f := newFastPathFixture(t, 900, 1000, nil)
		f.applyClass(modePtr(sdk.ProxyBackendMode), store.LimitLimitAction, 400)
		f.reader.total = 500
		f.full(f.tcp())
		calls := f.reader.calls

		f.reader.total = 300
		f.full(f.proxy())
		require.Equal(t, calls+1, f.reader.calls, "the full path reads the period's totals afresh")
		require.Empty(t, f.journal())
		require.Empty(t, f.limits.classes)
		require.Empty(t, f.warnings.classes)
	})
}

func TestCrossingNoticedWithinUsageLifetime(t *testing.T) {
	f := newFastPathFixture(t, 100, 200, func(cfg *Config) { cfg.UsageCacheLifetime = 30 * time.Second })
	f.reader.total = 50
	f.full(f.proxy())

	f.reader.total = 250
	f.advance(29 * time.Second)
	f.fast(f.proxy())
	require.Empty(t, f.journal())

	f.advance(time.Second)
	f.full(f.proxy())
	entries := f.journal()
	require.Len(t, entries, 1)
	require.Equal(t, store.LimitLimitAction, entries[0].Action)
	require.Len(t, f.limits.classes, 1)
}

func TestWarnedAccountKeepsSending(t *testing.T) {
	f := newFastPathFixture(t, 100, 200, nil)
	f.reader.total = 150
	f.full(f.proxy())
	require.Len(t, f.warnings.classes, 1)
	for i := 0; i < 5; i++ {
		f.advance(5 * time.Second)
		f.fast(f.proxy())
	}
	// expiry takes the full path once, which finds the entry already journaled and records it again.
	f.advance(time.Minute)
	f.full(f.proxy())
	f.fast(f.proxy())
	require.Len(t, f.journal(), 1)
	require.Len(t, f.warnings.classes, 1)
}

// a wrong implementation that stops at the first exceeded class it finds recorded answers "done" here.
func TestRecordedWarningDoesNotHideLimit(t *testing.T) {
	t.Run("global", func(t *testing.T) {
		f := newFastPathFixture(t, 100, 200, func(cfg *Config) { cfg.UsageCacheLifetime = 30 * time.Second })
		f.applyClass(modePtr(sdk.ProxyBackendMode), store.WarningLimitAction, 100)
		f.reader.total = 150
		f.full(f.tcp())
		require.Len(t, f.warnings.classes, 1)
		require.True(t, f.warnings.classes[0].IsGlobal())

		// a proxy event refreshes the totals over the global limit, and selects its own scoped class, which
		// comes last in its array.
		f.reader.total = 250
		f.advance(30 * time.Second)
		f.full(f.proxy())
		require.Len(t, f.warnings.classes, 2)
		require.False(t, f.warnings.classes[1].IsGlobal())
		require.Empty(t, f.limits.classes)

		// for tcpTunnel the global limit is now the last exceeded class, and nothing has recorded it.
		f.full(f.tcp())
		require.Len(t, f.limits.classes, 1)
		require.True(t, f.limits.classes[0].IsGlobal())
		require.Len(t, f.journal(), 3)
		f.fast(f.tcp())
	})

	t.Run("scoped", func(t *testing.T) {
		f := newFastPathFixture(t, 100, 200, nil)
		lcId := f.applyClass(modePtr(sdk.ProxyBackendMode), store.LimitLimitAction, 240)
		f.reader.total = 250
		f.full(f.tcp())
		require.Len(t, f.limits.classes, 1)
		require.True(t, f.limits.classes[0].IsGlobal())

		// for proxy the scoped class is the last exceeded, and only the global limit is recorded.
		f.full(f.proxy())
		require.Len(t, f.limits.classes, 2)
		require.Equal(t, lcId, limitClassId(f.limits.classes[1]))
		require.Len(t, f.journal(), 2)
		f.fast(f.proxy())
		f.fast(f.tcp())
	})
}

func TestRelaxedAccountCrossesAgain(t *testing.T) {
	t.Run("same class", func(t *testing.T) {
		f := newFastPathFixture(t, 100, 200, nil)
		f.reader.total = 150
		f.full(f.proxy())
		f.exec("delete from bandwidth_limit_journal where account_id = $1", f.acct)

		f.advance(59 * time.Second)
		f.fast(f.proxy())
		require.Empty(t, f.journal())
		f.advance(time.Second)
		f.full(f.proxy())
		require.Len(t, f.journal(), 1)
		require.Len(t, f.warnings.classes, 2)
	})

	// a wrong implementation that extends every crossing record when the account is refreshed keeps the
	// proxy record alive through the tcpTunnel traffic, and never journals the second crossing.
	t.Run("refreshed by another backend mode", func(t *testing.T) {
		f := newFastPathFixture(t, 900, 1000, func(cfg *Config) { cfg.UsageCacheLifetime = 10 * time.Second })
		lcId := f.applyClass(modePtr(sdk.ProxyBackendMode), store.WarningLimitAction, 100)
		f.reader.total = 150
		f.full(f.proxy())
		require.Len(t, f.warnings.classes, 1)
		f.exec("delete from bandwidth_limit_journal where account_id = $1 and limit_class_id = $2", f.acct, lcId)

		for i := 0; i < 5; i++ {
			f.advance(10 * time.Second)
			f.full(f.tcp())
		}
		f.fast(f.proxy())
		require.Empty(t, f.journal())

		f.advance(10 * time.Second)
		f.full(f.tcp())
		f.full(f.proxy())
		require.Len(t, f.journal(), 1)
		require.Len(t, f.warnings.classes, 2)
	})
}

// a wrong implementation that rebuilds the records on each run drops one class's record whenever it confirms
// the other's, and takes the full path on every event.
func TestTwoBackendModesEachRecordTheirOwnCrossing(t *testing.T) {
	f := newFastPathFixture(t, 900, 1000, nil)
	f.applyClass(modePtr(sdk.ProxyBackendMode), store.WarningLimitAction, 100)
	f.applyClass(modePtr(sdk.TcpTunnelBackendMode), store.WarningLimitAction, 100)
	f.reader.total = 150
	f.full(f.proxy())
	f.full(f.tcp())
	for i := 0; i < 3; i++ {
		f.advance(5 * time.Second)
		f.fast(f.proxy())
		f.fast(f.tcp())
	}
	require.Len(t, f.journal(), 2)
	require.Len(t, f.warnings.classes, 2)
}

func TestScopedClassJudgedPerBackendMode(t *testing.T) {
	f := newFastPathFixture(t, 900, 1000, nil)
	lcId := f.applyClass(modePtr(sdk.ProxyBackendMode), store.WarningLimitAction, 100)
	f.reader.total = 150
	f.full(f.tcp())
	f.fast(f.tcp())
	require.Empty(t, f.journal())

	f.full(f.proxy())
	require.Len(t, f.warnings.classes, 1)
	require.Equal(t, lcId, limitClassId(f.warnings.classes[0]))
	f.fast(f.proxy())
	f.fast(f.tcp())
}

func TestFailedEnforcementCachesNothing(t *testing.T) {
	f := newFastPathFixture(t, 100, 200, nil)
	f.reader.err = errors.New("influx unavailable")
	require.Error(t, f.agent.handleUsage(f.proxy()))
	require.Empty(t, f.agent.caches.accounts)
	require.Empty(t, f.agent.caches.totals)

	f.reader.err = nil
	f.reader.total = 50
	f.full(f.proxy())
	f.fast(f.proxy())

	t.Run("panic", func(t *testing.T) {
		g := newFastPathFixture(t, 100, 200, nil)
		g.reader.total = 150
		g.agent.warningActions = []AccountAction{panicAction{}}
		require.Error(t, g.agent.handleUsage(g.proxy()))
		require.Empty(t, g.agent.caches.accounts)
		require.Empty(t, g.agent.caches.totals)
	})
}

type panicAction struct{}

func (panicAction) HandleAccount(*store.Account, int64, int64, store.BandwidthClass, *userLimits, *sqlx.Tx) error {
	panic("action failed")
}

// a wrong implementation would hand the event's backend mode to enforcement, which would then check the
// tcpTunnel classes and find nothing exceeded.
func TestEventBackendModeDoesNotReachEnforcement(t *testing.T) {
	f := newFastPathFixture(t, 900, 1000, nil)
	lcId := f.applyClass(modePtr(sdk.ProxyBackendMode), store.LimitLimitAction, 100)
	f.reader.total = 150
	f.full(f.usage(proxyShare, sdk.TcpTunnelBackendMode))
	require.Len(t, f.limits.classes, 1)
	require.Equal(t, lcId, limitClassId(f.limits.classes[0]))
	entries := f.journal()
	require.Len(t, entries, 1)
	require.NotNil(t, entries[0].LimitClassId)
	require.Equal(t, lcId, *entries[0].LimitClassId)
}

func TestEventsTheFastPathSkips(t *testing.T) {
	t.Run("account id zero", func(t *testing.T) {
		f := newFastPathFixture(t, 100, 200, nil)
		f.reader.total = 50
		for i := 0; i < 3; i++ {
			u := f.proxy()
			u.AccountId = 0
			_, want := f.agent.enforce(u)
			require.Error(t, want)
			require.Equal(t, want.Error(), f.agent.handleUsage(u).Error())
		}
		require.EqualValues(t, 3, f.agent.caches.full)
		require.Empty(t, f.agent.caches.accounts)
		require.Empty(t, f.agent.caches.totals)
	})

	t.Run("no backend mode", func(t *testing.T) {
		f := newFastPathFixture(t, 100, 200, nil)
		f.reader.total = 50
		for i := 0; i < 3; i++ {
			f.full(f.usage(proxyShare, ""))
		}
		require.Equal(t, 3, f.reader.calls)
	})
}

func TestNegativeLifetimesDisableTheFastPath(t *testing.T) {
	for _, tc := range []struct {
		name             string
		account, usage   time.Duration
		limitlessAnswers bool
	}{
		{"both", -1, -1, false},
		{"account", -1, 0, false},
		{"usage", 0, -1, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := newFastPathFixture(t, 100, 200, func(cfg *Config) {
				cfg.AccountCacheLifetime = tc.account
				cfg.UsageCacheLifetime = tc.usage
			})
			f.reader.total = 150
			for i := 0; i < 3; i++ {
				f.full(f.proxy())
			}
			require.Equal(t, 3, f.reader.calls)
			require.Len(t, f.journal(), 1)

			f.exec("update accounts set limitless = true where id = $1", f.acct)
			f.full(f.proxy())
			if tc.limitlessAnswers {
				f.fast(f.proxy())
			} else {
				f.full(f.proxy())
			}
		})
	}
}

func TestCacheSettingsResolveWhereUsed(t *testing.T) {
	for _, cfg := range []*Config{nil, {}} {
		require.Equal(t, DefaultAccountCacheLifetime, cfg.accountCacheLifetime())
		require.Equal(t, DefaultUsageCacheLifetime, cfg.usageCacheLifetime())
		require.Equal(t, DefaultCacheReportInterval, cfg.cacheReportInterval())
	}
	negative := &Config{AccountCacheLifetime: -1, UsageCacheLifetime: -1, CacheReportInterval: -1}
	require.Negative(t, negative.accountCacheLifetime())
	require.Negative(t, negative.usageCacheLifetime())
	require.Negative(t, negative.cacheReportInterval())

	bound := DefaultConfig()
	require.NoError(t, dd.Merge(bound, map[string]any{
		"account_cache_lifetime": "2m",
		"usage_cache_lifetime":   "-1s",
		"cache_report_interval":  "0s",
	}))
	require.Equal(t, 2*time.Minute, bound.accountCacheLifetime())
	require.Equal(t, -time.Second, bound.usageCacheLifetime())
	require.Equal(t, DefaultCacheReportInterval, bound.cacheReportInterval())
}

func TestFastPathReport(t *testing.T) {
	var logs bytes.Buffer
	dl.Init(dl.DefaultOptions().JSON().SetLevel(slog.LevelInfo).SetOutput(&logs))
	t.Cleanup(func() { dl.Init() })
	f := newFastPathFixture(t, 100, 200, nil)
	f.agent.caches.lastReport = f.now
	f.reader.total = 50
	f.full(f.proxy())
	f.fast(f.proxy())
	f.fast(f.proxy())
	f.agent.reportCaches()
	require.NotContains(t, logs.String(), "limits fast path")

	f.advance(5 * time.Minute)
	f.agent.reportCaches()
	require.Equal(t, 1, strings.Count(logs.String(), "limits fast path"))
	require.Contains(t, logs.String(), "finished '2' events, '1' events took the full path")

	f.advance(5 * time.Minute)
	f.agent.reportCaches()
	require.Equal(t, 1, strings.Count(logs.String(), "limits fast path"))

	f.agent.cfg.CacheReportInterval = -1
	f.full(f.proxy())
	f.advance(time.Hour)
	f.agent.reportCaches()
	require.Equal(t, 1, strings.Count(logs.String(), "limits fast path"))
}

func TestMakeRoomBoundsTheMap(t *testing.T) {
	m := map[int]bool{1: true, 2: false, 3: true}
	makeRoom(m, 3, func(expired bool) bool { return expired })
	require.Equal(t, map[int]bool{2: false}, m)
	m[4], m[5] = false, false
	makeRoom(m, 3, func(expired bool) bool { return expired })
	require.Empty(t, m)
	m[6] = false
	makeRoom(m, 3, func(expired bool) bool { return expired })
	require.Len(t, m, 1)
}
