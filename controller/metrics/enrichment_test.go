package metrics

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/openziti/zrok/controller/store"
	"github.com/stretchr/testify/require"
)

type flakyShareStore struct {
	*store.Store
	failures      int32
	calls         atomic.Int32
	firstDeadline time.Time
}

func (s *flakyShareStore) FindShareWithZIdAndDeletedContext(ctx context.Context, id string, trx *sqlx.Tx) (*store.Share, error) {
	attempt := s.calls.Add(1)
	if attempt == 1 {
		s.firstDeadline, _ = ctx.Deadline()
	}
	if attempt <= s.failures {
		return nil, errors.New("store temporarily unavailable")
	}
	return s.Store.FindShareWithZIdAndDeletedContext(ctx, id, trx)
}

type contextualSinkFunc func(context.Context, *Usage) error

func (f contextualSinkFunc) Handle(u *Usage) error { return f(context.Background(), u) }
func (f contextualSinkFunc) HandleContext(ctx context.Context, u *Usage) error {
	return f(ctx, u)
}

func TestTransientEnrichment(t *testing.T) {
	for _, tc := range []struct {
		name     string
		failures int32
		attempts int
		budget   time.Duration
		wantAck  bool
	}{
		{"recovers", 2, 5, time.Second, true},
		{"budget exhausted", 1000, 1000, 30 * time.Millisecond, false},
		{"attempts exhausted", 3, 3, time.Second, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			str, err := store.Open(&store.Config{Type: "sqlite3", Path: ":memory:"})
			require.NoError(t, err)
			t.Cleanup(func() { _ = str.Close() })
			trx, err := str.Begin()
			require.NoError(t, err)
			defer func() { _ = trx.Rollback() }()
			accountID, err := str.CreateAccount(&store.Account{Email: "test@example.com", Token: "test", Password: "password"}, trx)
			require.NoError(t, err)
			envID, err := str.CreateEnvironment(accountID, &store.Environment{ZId: "environment"}, trx)
			require.NoError(t, err)
			_, err = str.CreateShare(envID, &store.Share{ZId: "unrelated", Token: "share", ShareMode: "private", BackendMode: "tcpTunnel", PermissionMode: store.OpenPermissionMode}, trx)
			require.NoError(t, err)
			require.NoError(t, trx.Commit())

			e := &testEvent{data: validUsage, done: make(chan bool, 2)}
			a, err := NewAgent(&AgentConfig{
				Source: &testSource{messages: []ZitiEventMsg{e}}, RetryAttempts: tc.attempts,
				RetryInitialBackoff: time.Millisecond, RetryMaxBackoff: 5 * time.Millisecond, RetryBudget: tc.budget,
			}, str, &InfluxConfig{})
			require.NoError(t, err)
			flaky := &flakyShareStore{Store: str, failures: tc.failures}
			a.cache.str = flaky
			a.durable.(*influxWriter).idb.Close()
			var got Usage
			var sinkDeadline time.Time
			var calls atomic.Int32
			a.durable = contextualSinkFunc(func(ctx context.Context, u *Usage) error {
				got = *u
				sinkDeadline, _ = ctx.Deadline()
				calls.Add(1)
				return nil
			})
			var laterCalls atomic.Int32
			a.AddUsageSink(sinkFunc(func(*Usage) error { laterCalls.Add(1); return nil }))
			started := time.Now()
			require.NoError(t, a.Start())
			t.Cleanup(a.Stop)
			require.Equal(t, tc.wantAck, awaitEvent(t, e))
			a.Stop()
			require.Less(t, time.Since(started), 500*time.Millisecond)
			if tc.wantAck {
				require.EqualValues(t, 3, flaky.calls.Load())
				require.EqualValues(t, 1, calls.Load())
				require.EqualValues(t, 1, laterCalls.Load())
				require.Equal(t, "share", got.ShareToken)
				require.EqualValues(t, envID, got.EnvironmentId)
				require.EqualValues(t, accountID, got.AccountId)
				require.Equal(t, flaky.firstDeadline, sinkDeadline, "enrichment and durable write must share one deadline")
				require.EqualValues(t, 1, e.acks.Load())
				require.Zero(t, e.nacks.Load())
			} else {
				require.Zero(t, calls.Load())
				require.Zero(t, laterCalls.Load())
				require.EqualValues(t, 1, e.nacks.Load())
				require.Zero(t, e.acks.Load())
				if tc.name == "budget exhausted" {
					require.GreaterOrEqual(t, time.Since(started), tc.budget)
					require.Greater(t, flaky.calls.Load(), int32(0))
					require.Less(t, int(flaky.calls.Load()), tc.attempts)
				} else {
					require.EqualValues(t, tc.attempts, flaky.calls.Load())
				}
			}
		})
	}
}
