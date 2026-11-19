package metrics

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/michaelquigley/df/dd"
	"github.com/michaelquigley/df/dl"
	"github.com/openziti/zrok/v2/controller/store"
	"github.com/stretchr/testify/require"
)

// testClock is the injected time source for the attribution cache; no fixture sleeps for a lifetime.
type testClock struct{ at atomic.Int64 }

func newTestClock() *testClock {
	c := &testClock{}
	c.at.Store(time.Unix(1_000_000, 0).UnixNano())
	return c
}

func (c *testClock) now() time.Time            { return time.Unix(0, c.at.Load()) }
func (c *testClock) advance(d time.Duration)   { c.at.Add(int64(d)) }
func (s *flakyShareStore) transactions() int32 { return s.beginCalls.Load() }
func (s *flakyShareStore) selects() int32      { return s.calls.Load() + s.envCalls.Load() }
func (s *flakyShareStore) totalCalls() int32   { return s.transactions() + s.selects() }
func (s *flakyShareStore) shareSelects() int32 { return s.calls.Load() }

type enrichmentFixture struct {
	str       *store.Store
	accountId int
	envId     int
	ephEnvId  int
}

// newEnrichmentFixture creates a share with an account on service 'svc-share' (proxy), a share in an ephemeral
// environment on service 'svc-ephemeral' (tcpTunnel), and nothing on any other service id.
func newEnrichmentFixture(t *testing.T) *enrichmentFixture {
	t.Helper()
	str, err := store.Open(&store.Config{Type: "sqlite3", Path: ":memory:"})
	require.NoError(t, err)
	t.Cleanup(func() { _ = str.Close() })
	trx, err := str.Begin()
	require.NoError(t, err)
	defer func() { _ = trx.Rollback() }()
	f := &enrichmentFixture{str: str}
	f.accountId, err = str.CreateAccount(&store.Account{Email: "test@example.com", Token: "test", Password: "password"}, trx)
	require.NoError(t, err)
	f.envId, err = str.CreateEnvironment(f.accountId, &store.Environment{ZId: "environment"}, trx)
	require.NoError(t, err)
	_, err = str.CreateShare(f.envId, &store.Share{ZId: "svc-share", Token: "share", ShareMode: "public", BackendMode: "proxy", PermissionMode: store.OpenPermissionMode}, trx)
	require.NoError(t, err)
	f.ephEnvId, err = str.CreateEphemeralEnvironment(&store.Environment{ZId: "ephemeral"}, trx)
	require.NoError(t, err)
	_, err = str.CreateShare(f.ephEnvId, &store.Share{ZId: "svc-ephemeral", Token: "ephemeral-share", ShareMode: "private", BackendMode: "tcpTunnel", PermissionMode: store.OpenPermissionMode}, trx)
	require.NoError(t, err)
	require.NoError(t, trx.Commit())
	return f
}

func (f *enrichmentFixture) cache(cfg AgentConfig) (*cache, *flakyShareStore, *testClock) {
	flaky := &flakyShareStore{Store: f.str}
	clk := newTestClock()
	c := newShareCache(f.str, cfg.withDefaults())
	c.str = flaky
	c.now = clk.now
	c.lastReport = clk.now()
	return c, flaky, clk
}

func TestCachedAttributionMatchesStore(t *testing.T) {
	f := newEnrichmentFixture(t)
	c, flaky, _ := f.cache(AgentConfig{})

	first := &Usage{ZitiServiceId: "svc-share"}
	require.NoError(t, c.addZrokDetail(context.Background(), first))
	require.EqualValues(t, 1, flaky.transactions())
	require.EqualValues(t, 2, flaky.selects())
	require.Equal(t, "share", first.ShareToken)
	require.EqualValues(t, f.envId, first.EnvironmentId)
	require.EqualValues(t, f.accountId, first.AccountId)
	require.Equal(t, "proxy", first.BackendMode)

	second := &Usage{ZitiServiceId: "svc-share"}
	require.NoError(t, c.addZrokDetail(context.Background(), second))
	require.EqualValues(t, 3, flaky.totalCalls(), "a cached attribution must make no store call")
	require.Equal(t, *first, *second)
}

func TestNotAShareIsCachedForItsLifetime(t *testing.T) {
	f := newEnrichmentFixture(t)
	c, flaky, clk := f.cache(AgentConfig{NotShareCacheLifetime: time.Minute})

	for i := 0; i < 10; i++ {
		u := &Usage{ZitiServiceId: "svc-missing"}
		require.ErrorIs(t, c.addZrokDetail(context.Background(), u), errNotAShare)
		require.Empty(t, u.ShareToken)
	}
	require.EqualValues(t, 1, flaky.shareSelects())

	clk.advance(time.Minute - time.Second)
	require.ErrorIs(t, c.addZrokDetail(context.Background(), &Usage{ZitiServiceId: "svc-missing"}), errNotAShare)
	require.EqualValues(t, 1, flaky.shareSelects())

	clk.advance(time.Second)
	require.ErrorIs(t, c.addZrokDetail(context.Background(), &Usage{ZitiServiceId: "svc-missing"}), errNotAShare)
	require.EqualValues(t, 2, flaky.shareSelects(), "an expired negative entry must go back to the store")
}

func TestShareWithoutAccountIsCached(t *testing.T) {
	f := newEnrichmentFixture(t)
	c, flaky, _ := f.cache(AgentConfig{})

	for i := 0; i < 2; i++ {
		u := &Usage{ZitiServiceId: "svc-ephemeral"}
		err := c.addZrokDetail(context.Background(), u)
		require.ErrorIs(t, err, errNoAccount)
		require.Equal(t, "ephemeral-share", u.ShareToken)
		require.EqualValues(t, f.ephEnvId, u.EnvironmentId)
		require.Zero(t, u.AccountId)
		require.Equal(t, "tcpTunnel", u.BackendMode)
	}
	require.EqualValues(t, 1, flaky.shareSelects())
}

// a wrong implementation would remember a failure, most plausibly as "not a share", and answer the next
// lookup from memory: no second store call and no attribution.
func TestStoreFailureIsNotCached(t *testing.T) {
	for _, stage := range []string{"begin", "share select", "environment select", "cancelled context"} {
		t.Run(stage, func(t *testing.T) {
			f := newEnrichmentFixture(t)
			c, flaky, _ := f.cache(AgentConfig{})
			ctx := context.Background()
			switch stage {
			case "begin":
				flaky.beginFailures = 1
			case "share select":
				flaky.failures = 1
			case "environment select":
				flaky.envFailures = 1
			case "cancelled context":
				cancelled, cancel := context.WithCancel(ctx)
				cancel()
				ctx = cancelled
			}

			failed := &Usage{ZitiServiceId: "svc-share"}
			err := c.addZrokDetail(ctx, failed)
			require.Error(t, err)
			require.NotErrorIs(t, err, errNotAShare)
			require.NotErrorIs(t, err, errNoAccount)
			require.Empty(t, failed.ShareToken)
			require.Empty(t, c.entries, "a failure must leave no entry")

			before := flaky.transactions()
			u := &Usage{ZitiServiceId: "svc-share"}
			require.NoError(t, c.addZrokDetail(context.Background(), u))
			require.Equal(t, before+1, flaky.transactions(), "the next lookup must go to the store")
			require.Equal(t, "share", u.ShareToken)
			require.EqualValues(t, f.accountId, u.AccountId)
		})
	}
}

func TestNegativeLifetimesDisableTheCache(t *testing.T) {
	f := newEnrichmentFixture(t)
	c, flaky, _ := f.cache(AgentConfig{ShareCacheLifetime: -1, NotShareCacheLifetime: -1})
	require.Negative(t, c.shareLifetime)
	require.Negative(t, c.notShareLifetime)

	for i := 0; i < 3; i++ {
		u := &Usage{ZitiServiceId: "svc-share"}
		require.NoError(t, c.addZrokDetail(context.Background(), u))
		require.Equal(t, "share", u.ShareToken)
		require.ErrorIs(t, c.addZrokDetail(context.Background(), &Usage{ZitiServiceId: "svc-missing"}), errNotAShare)
	}
	require.EqualValues(t, 6, flaky.shareSelects())
	require.EqualValues(t, 6, flaky.transactions())
	require.Empty(t, c.entries)
}

func TestCacheSizeIsBounded(t *testing.T) {
	f := newEnrichmentFixture(t)
	c, flaky, clk := f.cache(AgentConfig{ShareCacheSize: 3, ShareCacheLifetime: 10 * time.Minute, NotShareCacheLifetime: time.Minute})

	require.NoError(t, c.addZrokDetail(context.Background(), &Usage{ZitiServiceId: "svc-share"}))
	require.ErrorIs(t, c.addZrokDetail(context.Background(), &Usage{ZitiServiceId: "missing-1"}), errNotAShare)
	require.ErrorIs(t, c.addZrokDetail(context.Background(), &Usage{ZitiServiceId: "missing-2"}), errNotAShare)
	require.Len(t, c.entries, 3)

	// a full map first drops its expired entries.
	clk.advance(2 * time.Minute)
	require.ErrorIs(t, c.addZrokDetail(context.Background(), &Usage{ZitiServiceId: "missing-3"}), errNotAShare)
	require.Len(t, c.entries, 2)
	selects := flaky.shareSelects()
	require.NoError(t, c.addZrokDetail(context.Background(), &Usage{ZitiServiceId: "svc-share"}))
	require.Equal(t, selects, flaky.shareSelects(), "an unexpired entry survives the sweep")

	// a full map with nothing expired is cleared.
	for i := 0; i < 20; i++ {
		_ = c.addZrokDetail(context.Background(), &Usage{ZitiServiceId: "more-" + string(rune('a'+i))})
		require.LessOrEqual(t, len(c.entries), 3)
	}
}

func TestCacheReport(t *testing.T) {
	var logs bytes.Buffer
	dl.Init(dl.DefaultOptions().JSON().SetLevel(slog.LevelInfo).SetOutput(&logs))
	t.Cleanup(func() { dl.Init() })
	f := newEnrichmentFixture(t)
	c, _, clk := f.cache(AgentConfig{CacheReportInterval: 5 * time.Minute})

	require.NoError(t, c.addZrokDetail(context.Background(), &Usage{ZitiServiceId: "svc-share"}))
	require.NoError(t, c.addZrokDetail(context.Background(), &Usage{ZitiServiceId: "svc-share"}))
	require.ErrorIs(t, c.addZrokDetail(context.Background(), &Usage{ZitiServiceId: "svc-missing"}), errNotAShare)
	require.ErrorIs(t, c.addZrokDetail(context.Background(), &Usage{ZitiServiceId: "svc-missing"}), errNotAShare)
	require.ErrorIs(t, c.addZrokDetail(context.Background(), &Usage{ZitiServiceId: "svc-missing"}), errNotAShare)
	c.report()
	require.NotContains(t, logs.String(), "share cache")

	clk.advance(5 * time.Minute)
	c.report()
	require.Equal(t, 1, strings.Count(logs.String(), "share cache"))
	require.Contains(t, logs.String(), "'1' share and '2' not-a-share lookups from memory, sent '2' lookups to the store, holds '2' entries")

	// no lookups since the last report: no line.
	clk.advance(5 * time.Minute)
	c.report()
	require.Equal(t, 1, strings.Count(logs.String(), "share cache"))

	off, _, offClk := f.cache(AgentConfig{CacheReportInterval: -1})
	require.NoError(t, off.addZrokDetail(context.Background(), &Usage{ZitiServiceId: "svc-share"}))
	offClk.advance(time.Hour)
	off.report()
	require.Equal(t, 1, strings.Count(logs.String(), "share cache"))
}

func TestAgentConfigCacheBinding(t *testing.T) {
	cfg := &AgentConfig{}
	require.NoError(t, dd.Bind(cfg, map[string]any{
		"share_cache_lifetime":     "30s",
		"not_share_cache_lifetime": "-1s",
		"share_cache_size":         10,
		"cache_report_interval":    "-1s",
	}))
	require.Equal(t, 30*time.Second, cfg.ShareCacheLifetime)
	require.Equal(t, -time.Second, cfg.NotShareCacheLifetime)
	require.Equal(t, 10, cfg.ShareCacheSize)
	require.Equal(t, -time.Second, cfg.CacheReportInterval)

	applied := cfg.withDefaults()
	require.Equal(t, 30*time.Second, applied.ShareCacheLifetime)
	require.Equal(t, -time.Second, applied.NotShareCacheLifetime)
	require.Equal(t, -time.Second, applied.CacheReportInterval)

	defaults := AgentConfig{}.withDefaults()
	require.Equal(t, 10*time.Minute, defaults.ShareCacheLifetime)
	require.Equal(t, time.Minute, defaults.NotShareCacheLifetime)
	require.Equal(t, 100000, defaults.ShareCacheSize)
	require.Equal(t, 5*time.Minute, defaults.CacheReportInterval)
}
