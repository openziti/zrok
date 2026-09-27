package metrics

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/michaelquigley/df/dl"
	"github.com/openziti/zrok/v2/controller/store"
	"github.com/stretchr/testify/require"
)

const validUsage ZitiEventJson = `{"namespace":"fabric.usage","interval_start_utc":1,"tags":{"serviceId":"unrelated"},"usage":{}}`

type testEvent struct {
	data  ZitiEventJson
	acks  atomic.Int32
	nacks atomic.Int32
	done  chan bool
}

func (e *testEvent) Data() ZitiEventJson { return e.data }
func (e *testEvent) Ack() error          { e.acks.Add(1); e.done <- true; return nil }
func (e *testEvent) Nack(requeue bool) error {
	e.nacks.Add(1)
	if requeue {
		panic("consumer must never requeue")
	}
	e.done <- false
	return nil
}

type testSource struct {
	drainOnStop bool
	messages    []ZitiEventMsg
	stop        chan struct{}
	join        chan struct{}
	once        sync.Once
}

func (*testSource) Type() string                   { return "test" }
func (*testSource) ToMap() (map[string]any, error) { return nil, nil }
func (s *testSource) Start(events chan ZitiEventMsg) (chan struct{}, error) {
	s.stop, s.join = make(chan struct{}), make(chan struct{})
	go func() {
		defer close(s.join)
		if s.drainOnStop {
			<-s.stop
			for _, event := range s.messages {
				events <- event
			}
			return
		}
		for _, event := range s.messages {
			select {
			case events <- event:
			case <-s.stop:
				return
			}
		}
		<-s.stop
	}()
	return s.join, nil
}
func (s *testSource) Stop() { s.once.Do(func() { close(s.stop) }); <-s.join }

type sinkFunc func(*Usage) error

func (f sinkFunc) Handle(u *Usage) error { return f(u) }

func startTestAgent(t *testing.T, cfg AgentConfig, events []ZitiEventMsg, sinks ...UsageSink) *Agent {
	t.Helper()
	cfg.Source = &testSource{messages: events}
	str, err := store.Open(&store.Config{Type: "sqlite3", Path: ":memory:"})
	require.NoError(t, err)
	t.Cleanup(func() { _ = str.Close() })
	a, err := NewAgent(&cfg, str, &InfluxConfig{Url: "http://127.0.0.1:1"})
	require.NoError(t, err)
	a.durable.(*influxWriter).idb.Close()
	a.durable = sinks[0]
	for _, sink := range sinks[1:] {
		a.AddUsageSink(sink)
	}
	require.NoError(t, a.Start())
	t.Cleanup(a.Stop)
	return a
}

func awaitEvent(t *testing.T, e *testEvent) bool {
	t.Helper()
	select {
	case ack := <-e.done:
		return ack
	case <-time.After(2 * time.Second):
		t.Fatal("delivery was not settled")
		return false
	}
}

func TestIngestFailureNacksImmediately(t *testing.T) {
	e := &testEvent{data: "not-json", done: make(chan bool, 2)}
	var calls atomic.Int32
	startTestAgent(t, AgentConfig{}, []ZitiEventMsg{e}, sinkFunc(func(*Usage) error { calls.Add(1); return nil }))
	require.False(t, awaitEvent(t, e))
	require.EqualValues(t, 1, e.nacks.Load())
	require.Zero(t, e.acks.Load())
	require.Zero(t, calls.Load())
}

func TestSinkRetriesThenNacks(t *testing.T) {
	e := &testEvent{data: validUsage, done: make(chan bool, 2)}
	var calls, downstream atomic.Int32
	started := time.Now()
	startTestAgent(t, AgentConfig{RetryAttempts: 3, RetryInitialBackoff: 5 * time.Millisecond, RetryMaxBackoff: 10 * time.Millisecond, RetryBudget: time.Second}, []ZitiEventMsg{e},
		sinkFunc(func(*Usage) error { calls.Add(1); return errors.New("offline") }),
		sinkFunc(func(*Usage) error { downstream.Add(1); return nil }))
	require.False(t, awaitEvent(t, e))
	require.EqualValues(t, 3, calls.Load())
	require.Zero(t, downstream.Load())
	require.EqualValues(t, 1, e.nacks.Load())
	require.Zero(t, e.acks.Load())
	require.GreaterOrEqual(t, time.Since(started), 15*time.Millisecond)
	require.Less(t, time.Since(started), time.Second)
}

func TestSinkRecoversWithinBudget(t *testing.T) {
	e := &testEvent{data: validUsage, done: make(chan bool, 2)}
	var calls, downstream atomic.Int32
	startTestAgent(t, AgentConfig{RetryInitialBackoff: time.Millisecond}, []ZitiEventMsg{e},
		sinkFunc(func(*Usage) error {
			if calls.Add(1) <= 2 {
				return errors.New("temporarily offline")
			}
			return nil
		}),
		sinkFunc(func(*Usage) error { downstream.Add(1); return nil }))
	require.True(t, awaitEvent(t, e))
	require.EqualValues(t, 3, calls.Load())
	require.EqualValues(t, 1, downstream.Load())
	require.EqualValues(t, 1, e.acks.Load())
	require.Zero(t, e.nacks.Load())
}

func TestRetryBudgetBoundsBackoff(t *testing.T) {
	e := &testEvent{data: validUsage, done: make(chan bool, 2)}
	var calls atomic.Int32
	started := time.Now()
	startTestAgent(t, AgentConfig{RetryAttempts: 100, RetryBudget: 30 * time.Millisecond, RetryInitialBackoff: time.Second}, []ZitiEventMsg{e},
		sinkFunc(func(*Usage) error { calls.Add(1); return errors.New("offline") }))
	require.False(t, awaitEvent(t, e))
	require.EqualValues(t, 1, calls.Load())
	require.Less(t, time.Since(started), 500*time.Millisecond)
}

func TestStopInterruptsHangingSink(t *testing.T) {
	e := &testEvent{data: validUsage, done: make(chan bool, 2)}
	entered, release, exited := make(chan struct{}), make(chan struct{}), make(chan struct{})
	a := startTestAgent(t, AgentConfig{}, []ZitiEventMsg{e}, sinkFunc(func(*Usage) error {
		close(entered)
		<-release
		defer close(exited)
		return errors.New("offline")
	}))
	defer func() { close(release); <-exited }()
	<-entered
	stopped := make(chan struct{})
	go func() { a.Stop(); close(stopped) }()
	select {
	case <-stopped:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("stop waited for the retry budget")
	}
	require.Zero(t, e.acks.Load())
	require.Zero(t, e.nacks.Load())
}

func TestHungSinkDoesNotAccumulateWorkers(t *testing.T) {
	first := &testEvent{data: validUsage, done: make(chan bool, 2)}
	second := &testEvent{data: validUsage, done: make(chan bool, 2)}
	release, exited := make(chan struct{}), make(chan struct{})
	var calls atomic.Int32
	startTestAgent(t, AgentConfig{RetryBudget: 20 * time.Millisecond, RetryInitialBackoff: time.Millisecond}, []ZitiEventMsg{first, second}, sinkFunc(func(*Usage) error {
		calls.Add(1)
		defer close(exited)
		<-release
		return errors.New("offline")
	}))
	defer func() { close(release); <-exited }()
	require.False(t, awaitEvent(t, first))
	require.False(t, awaitEvent(t, second))
	require.EqualValues(t, 1, calls.Load())
}

func TestStopInterruptsRetryBackoff(t *testing.T) {
	e := &testEvent{data: validUsage, done: make(chan bool, 2)}
	entered := make(chan struct{}, 1)
	a := startTestAgent(t, AgentConfig{RetryInitialBackoff: time.Minute, RetryMaxBackoff: time.Minute}, []ZitiEventMsg{e}, sinkFunc(func(*Usage) error {
		entered <- struct{}{}
		return errors.New("offline")
	}))
	<-entered
	done := make(chan struct{})
	go func() { a.Stop(); close(done) }()
	select {
	case <-done:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("stop waited for backoff")
	}
	require.Zero(t, e.acks.Load())
	require.Zero(t, e.nacks.Load())
}

type observedEvent struct {
	*testEvent
	read chan struct{}
}

func (e *observedEvent) Data() ZitiEventJson {
	close(e.read)
	return e.testEvent.Data()
}

func TestBlockedEnrichmentIsBounded(t *testing.T) {
	for _, shutdown := range []bool{false, true} {
		name := "budget"
		if shutdown {
			name = "shutdown"
		}
		t.Run(name, func(t *testing.T) {
			str, err := store.Open(&store.Config{Type: "sqlite3", Path: ":memory:", DisableAutoMigration: true})
			require.NoError(t, err)
			t.Cleanup(func() { _ = str.Close() })
			// sqlite has one connection; holding it parks enrichment at transaction begin.
			trx, err := str.Begin()
			require.NoError(t, err)
			t.Cleanup(func() { _ = trx.Rollback() })
			e := &observedEvent{testEvent: &testEvent{
				data: validUsage,
				done: make(chan bool, 2),
			}, read: make(chan struct{})}
			budget := 30 * time.Millisecond
			if shutdown {
				budget = time.Minute
			}
			a, err := NewAgent(&AgentConfig{Source: &testSource{messages: []ZitiEventMsg{e}}, RetryBudget: budget}, str, &InfluxConfig{})
			require.NoError(t, err)
			a.durable.(*influxWriter).idb.Close()
			var calls atomic.Int32
			a.durable = sinkFunc(func(*Usage) error { calls.Add(1); return nil })
			started := time.Now()
			require.NoError(t, a.Start())
			t.Cleanup(a.Stop)
			<-e.read
			if shutdown {
				done := make(chan struct{})
				go func() { a.Stop(); close(done) }()
				select {
				case <-done:
				case <-time.After(500 * time.Millisecond):
					t.Fatal("stop waited for the store connection")
				}
			}
			if shutdown {
				require.Zero(t, e.acks.Load())
				require.Zero(t, e.nacks.Load())
				require.Zero(t, calls.Load())
			} else {
				require.False(t, awaitEvent(t, e.testEvent))
				require.Zero(t, calls.Load())
				require.EqualValues(t, 1, e.nacks.Load())
				require.Zero(t, e.acks.Load())
			}
			require.Less(t, time.Since(started), 500*time.Millisecond)
		})
	}
}

func TestUnknownShareIsBestEffort(t *testing.T) {
	var logs bytes.Buffer
	dl.Init(dl.DefaultOptions().JSON().SetLevel(slog.LevelDebug).SetOutput(&logs))
	t.Cleanup(func() { dl.Init() })
	e := &testEvent{data: validUsage, done: make(chan bool, 2)}
	var calls atomic.Int32
	var shareToken string
	a := startTestAgent(t, AgentConfig{RetryInitialBackoff: time.Minute, RetryMaxBackoff: time.Minute}, []ZitiEventMsg{e},
		sinkFunc(func(u *Usage) error {
			shareToken = u.ShareToken
			calls.Add(1)
			return nil
		}))
	require.True(t, awaitEvent(t, e))
	a.Stop()
	require.Empty(t, shareToken)
	require.EqualValues(t, 1, calls.Load())
	require.EqualValues(t, 1, e.acks.Load())
	require.Zero(t, e.nacks.Load())
	require.Equal(t, 1, strings.Count(logs.String(), "unable to add zrok detail"))
	require.NotContains(t, logs.String(), "usage processing attempt")
}

func TestLaterSinkFailureLogsAndAcks(t *testing.T) {
	var logs bytes.Buffer
	dl.Init(dl.DefaultOptions().JSON().SetOutput(&logs))
	t.Cleanup(func() { dl.Init() })
	e := &testEvent{data: validUsage, done: make(chan bool, 2)}
	var durableCalls, laterCalls atomic.Int32
	a := startTestAgent(t, AgentConfig{}, []ZitiEventMsg{e},
		sinkFunc(func(*Usage) error { durableCalls.Add(1); return nil }),
		sinkFunc(func(*Usage) error { laterCalls.Add(1); return errors.New("later sink unavailable") }))
	require.True(t, awaitEvent(t, e))
	a.Stop()
	require.EqualValues(t, 1, durableCalls.Load())
	require.EqualValues(t, 1, laterCalls.Load())
	require.EqualValues(t, 1, e.acks.Load())
	require.Zero(t, e.nacks.Load())
	require.Contains(t, logs.String(), "later sink unavailable")
	require.Contains(t, logs.String(), "delivery will be acknowledged")
}

func TestShutdownDrainsWithoutSettling(t *testing.T) {
	events := []*testEvent{
		{data: validUsage, done: make(chan bool, 2)},
		{data: validUsage, done: make(chan bool, 2)},
		{data: validUsage, done: make(chan bool, 2)},
	}
	src := &testSource{drainOnStop: true}
	for _, event := range events {
		src.messages = append(src.messages, event)
	}
	a, err := NewAgent(&AgentConfig{Source: src}, nil, &InfluxConfig{})
	require.NoError(t, err)
	a.durable.(*influxWriter).idb.Close()
	var calls atomic.Int32
	a.durable = sinkFunc(func(*Usage) error { calls.Add(1); return nil })
	require.NoError(t, a.Start())
	t.Cleanup(a.Stop)
	done := make(chan struct{})
	go func() { a.Stop(); close(done) }()
	select {
	case <-done:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("stop failed to drain the source")
	}
	for _, event := range events {
		require.Zero(t, event.acks.Load())
		require.Zero(t, event.nacks.Load())
	}
	require.Zero(t, calls.Load())
}

func TestEphemeralShareKeepsAttribution(t *testing.T) {
	str, err := store.Open(&store.Config{Type: "sqlite3", Path: ":memory:"})
	require.NoError(t, err)
	t.Cleanup(func() { _ = str.Close() })
	trx, err := str.Begin()
	require.NoError(t, err)
	defer func() { _ = trx.Rollback() }()
	envID, err := str.CreateEphemeralEnvironment(&store.Environment{ZId: "ephemeral"}, trx)
	require.NoError(t, err)
	_, err = str.CreateShare(envID, &store.Share{ZId: "unrelated", Token: "ephemeral-share", ShareMode: "private", BackendMode: "tcpTunnel", PermissionMode: "open"}, trx)
	require.NoError(t, err)
	require.NoError(t, trx.Commit())

	e := &testEvent{data: validUsage, done: make(chan bool, 2)}
	a, err := NewAgent(&AgentConfig{Source: &testSource{messages: []ZitiEventMsg{e}}}, str, &InfluxConfig{})
	require.NoError(t, err)
	a.durable.(*influxWriter).idb.Close()
	var got Usage
	var calls atomic.Int32
	a.durable = sinkFunc(func(u *Usage) error { got = *u; calls.Add(1); return nil })
	require.NoError(t, a.Start())
	t.Cleanup(a.Stop)
	require.True(t, awaitEvent(t, e))
	a.Stop()
	require.EqualValues(t, 1, calls.Load())
	require.Equal(t, "ephemeral-share", got.ShareToken)
	require.EqualValues(t, envID, got.EnvironmentId)
	require.Zero(t, got.AccountId)
	require.EqualValues(t, 1, e.acks.Load())
	require.Zero(t, e.nacks.Load())
}

type contextualTestSink struct{ sinkFunc }

func (s contextualTestSink) HandleContext(_ context.Context, u *Usage) error { return s.Handle(u) }

func TestDurablePanicNacksAndContinues(t *testing.T) {
	for _, contextual := range []bool{false, true} {
		name := "legacy"
		if contextual {
			name = "contextual"
		}
		t.Run(name, func(t *testing.T) {
			var logs bytes.Buffer
			dl.Init(dl.DefaultOptions().JSON().SetOutput(&logs))
			t.Cleanup(func() { dl.Init() })
			poison := &testEvent{data: validUsage, done: make(chan bool, 2)}
			next := &testEvent{data: validUsage, done: make(chan bool, 2)}
			var calls, laterCalls atomic.Int32
			fn := sinkFunc(func(*Usage) error {
				if calls.Add(1) == 1 {
					panic("poison marker")
				}
				return nil
			})
			var durable UsageSink = fn
			if contextual {
				durable = contextualTestSink{fn}
			}
			a := startTestAgent(t, AgentConfig{}, []ZitiEventMsg{poison, next}, durable,
				sinkFunc(func(*Usage) error { laterCalls.Add(1); return nil }))
			require.False(t, awaitEvent(t, poison))
			require.True(t, awaitEvent(t, next))
			a.Stop()
			require.EqualValues(t, 1, poison.nacks.Load())
			require.Zero(t, poison.acks.Load())
			require.EqualValues(t, 1, next.acks.Load())
			require.Zero(t, next.nacks.Load())
			require.EqualValues(t, 2, calls.Load())
			require.EqualValues(t, 1, laterCalls.Load())
			require.Contains(t, logs.String(), "panic processing usage body")
			require.Contains(t, logs.String(), "poison marker")
		})
	}
}
