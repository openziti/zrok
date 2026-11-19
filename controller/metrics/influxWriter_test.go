package metrics

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	protocol "github.com/influxdata/line-protocol"
	"github.com/stretchr/testify/require"
)

type influxPointKey struct {
	measurement string
	tags        string
	timestamp   int64
}

// fakeInflux retains points by InfluxDB identity, replacing an earlier write to
// the same measurement, tag set and timestamp. fields are not part of the key.
type fakeInflux struct {
	mu         sync.Mutex
	points     map[influxPointKey]protocol.Metric
	precisions []string
}

func newTestInfluxWriter(t *testing.T) (*influxWriter, *fakeInflux) {
	t.Helper()
	f := &fakeInflux{points: make(map[influxPointKey]protocol.Metric)}
	srv := httptest.NewServer(http.HandlerFunc(f.write))
	t.Cleanup(srv.Close)
	w := newInfluxWriter(&InfluxConfig{Url: srv.URL, Org: "org", Bucket: "bucket"})
	t.Cleanup(w.idb.Close)
	return w, f
}

func (f *fakeInflux) write(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost || r.URL.Path != "/api/v2/write" {
		http.Error(w, "expected a write request", http.StatusBadRequest)
		return
	}
	data, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	handler := protocol.NewMetricHandler()
	handler.SetTimePrecision(time.Nanosecond)
	points, err := protocol.NewParser(handler).Parse(data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	f.precisions = append(f.precisions, r.URL.Query().Get("precision"))
	for _, pt := range points {
		var tags []string
		for _, tag := range pt.TagList() {
			tags = append(tags, fmt.Sprintf("%q=%q", tag.Key, tag.Value))
		}
		sort.Strings(tags)
		key := influxPointKey{pt.Name(), strings.Join(tags, ","), pt.Time().UnixNano()}
		f.points[key] = pt
	}
	w.WriteHeader(http.StatusNoContent)
}

func (f *fakeInflux) snapshot(t *testing.T) map[influxPointKey]protocol.Metric {
	t.Helper()
	f.mu.Lock()
	defer f.mu.Unlock()
	require.NotEmpty(t, f.precisions)
	for _, precision := range f.precisions {
		require.Equal(t, "ns", precision)
	}
	out := make(map[influxPointKey]protocol.Metric, len(f.points))
	for key, pt := range f.points {
		out[key] = pt
	}
	return out
}

func pointIdentityUsage() Usage {
	return Usage{
		IntervalStart: time.Unix(1700000000, 0),
		ShareToken:    "share", EnvironmentId: 7, AccountId: 11,
		ZitiCircuitId: "circuit-1", ZitiRouterId: "router-1", ZitiReportTime: "2026-11-19T12:00:00.123456789Z",
		BackendTx: 1000, BackendRx: 2000, FrontendTx: 3000, FrontendRx: 4000,
	}
}

func assertStoredUsage(t *testing.T, f *fakeInflux, count int, backendTx, backendRx, frontendTx, frontendRx int64) {
	t.Helper()
	points := f.snapshot(t)
	require.Len(t, points, 3*count)
	counts := make(map[string]int)
	totals := make(map[string]int64)
	for _, pt := range points {
		tags := make(map[string]string)
		for _, tag := range pt.TagList() {
			tags[tag.Key] = tag.Value
		}
		expectedTags := map[string]string{"share": "share", "envId": "7", "acctId": "11"}
		kind := pt.Name()
		if kind == "xfer" {
			kind = tags["namespace"]
			require.Contains(t, []string{"backend", "frontend"}, kind)
			expectedTags["namespace"] = kind
		} else {
			require.Equal(t, "circuits", kind)
		}
		require.Equal(t, expectedTags, tags)
		counts[kind]++
		if kind != "circuits" {
			require.Len(t, pt.FieldList(), 2)
			for _, field := range pt.FieldList() {
				value, ok := field.Value.(int64)
				require.True(t, ok)
				totals[kind+"/"+field.Key] += value
			}
		}
	}
	require.Equal(t, map[string]int{"circuits": count, "backend": count, "frontend": count}, counts)
	require.Equal(t, map[string]int64{
		"backend/tx": backendTx, "backend/rx": backendRx,
		"frontend/tx": frontendTx, "frontend/rx": frontendRx,
	}, totals)
}

func TestInfluxSeveralCircuitsAreKept(t *testing.T) {
	w, f := newTestInfluxWriter(t)
	for i, bytes := range []int64{1000, 2000, 4000} {
		u := pointIdentityUsage()
		u.ZitiCircuitId = fmt.Sprintf("circuit-%d", i)
		u.BackendTx, u.BackendRx = bytes, 2*bytes
		u.FrontendTx, u.FrontendRx = 3*bytes, 4*bytes
		require.NoError(t, w.Handle(&u))
	}
	assertStoredUsage(t, f, 3, 7000, 14000, 21000, 28000)
}

func TestInfluxRedeliveryReplacesItself(t *testing.T) {
	w, f := newTestInfluxWriter(t)
	u := pointIdentityUsage()
	require.NoError(t, w.Handle(&u))
	before := f.snapshot(t)
	// a new consumer has no shared counter or other process-local state.
	w2 := newInfluxWriter(&InfluxConfig{Url: w.idb.ServerURL(), Org: "org", Bucket: "bucket"})
	t.Cleanup(w2.idb.Close)
	u.ProcessedStamp = time.Now()
	require.NoError(t, w2.Handle(&u))
	require.Equal(t, before, f.snapshot(t))
	assertStoredUsage(t, f, 1, 1000, 2000, 3000, 4000)
}

func TestInfluxReportIdentity(t *testing.T) {
	for _, part := range []string{"report time", "router"} {
		t.Run(part, func(t *testing.T) {
			w, f := newTestInfluxWriter(t)
			u := pointIdentityUsage()
			require.NoError(t, w.Handle(&u))
			if part == "report time" {
				u.ZitiReportTime = "2026-11-19T12:01:00.123456789Z"
			} else {
				u.ZitiRouterId = "router-2"
			}
			require.NoError(t, w.Handle(&u))
			assertStoredUsage(t, f, 2, 2000, 4000, 6000, 8000)
		})
	}
}

func TestInfluxIdentityFromEvent(t *testing.T) {
	const event = `{"namespace":"fabric.usage","interval_start_utc":1700000000,"tags":{"serviceId":"service"},"circuit_id":"circuit-1","source_id":"router-1","timestamp":"2026-11-19T12:00:00.123456789Z","usage":{"egress.rx":1000,"egress.tx":2000,"ingress.tx":3000,"ingress.rx":4000}}`
	for _, tc := range []struct{ name, old, replacement string }{
		{"report time", "2026-11-19T12:00:00.123456789Z", "2026-11-19T12:01:00.123456789Z"},
		{"router", "router-1", "router-2"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			w, f := newTestInfluxWriter(t)
			for _, data := range []string{event, strings.Replace(event, tc.old, tc.replacement, 1)} {
				u, err := Ingest(ZitiEventJson(data))
				require.NoError(t, err)
				u.ShareToken, u.EnvironmentId, u.AccountId = "share", 7, 11
				require.NoError(t, w.Handle(u))
			}
			assertStoredUsage(t, f, 2, 2000, 4000, 6000, 8000)
		})
	}
}

func TestInfluxMissingIdentity(t *testing.T) {
	w, f := newTestInfluxWriter(t)
	u := pointIdentityUsage()
	u.ZitiCircuitId, u.ZitiRouterId, u.ZitiReportTime = "circuit-1", "", ""
	require.NoError(t, w.Handle(&u))
	u.ZitiCircuitId, u.ZitiReportTime = "", "2026-11-19T12:00:00.123456789Z"
	require.NoError(t, w.Handle(&u))
	assertStoredUsage(t, f, 2, 2000, 4000, 6000, 8000)
	u.ZitiReportTime = ""
	require.NoError(t, w.Handle(&u))
	before := f.snapshot(t)
	require.NoError(t, w.Handle(&u))
	require.Equal(t, before, f.snapshot(t))
	assertStoredUsage(t, f, 3, 3000, 6000, 9000, 12000)
}

func TestInfluxPointTimes(t *testing.T) {
	for i := range 100 {
		t.Run(fmt.Sprintf("identity-%d", i), func(t *testing.T) {
			w, f := newTestInfluxWriter(t)
			u := pointIdentityUsage()
			u.ZitiCircuitId = fmt.Sprintf("circuit-%d", i)
			require.NoError(t, w.Handle(&u))
			points := f.snapshot(t)
			require.Len(t, points, 3)
			var stamp time.Time
			for _, pt := range points {
				require.True(t, pt.Time().After(u.IntervalStart))
				require.True(t, pt.Time().Before(u.IntervalStart.Add(time.Second)))
				if stamp.IsZero() {
					stamp = pt.Time()
				}
				require.Equal(t, stamp, pt.Time())
			}
		})
	}
}

func TestInfluxPointTimeIsPinned(t *testing.T) {
	w, f := newTestInfluxWriter(t)
	u := pointIdentityUsage()
	require.NoError(t, w.Handle(&u))
	points := f.snapshot(t)
	require.Len(t, points, 3)
	for _, pt := range points {
		require.Equal(t, u.IntervalStart.Add(640103973*time.Nanosecond), pt.Time())
	}
}

func TestHangingInfluxWriteIsBounded(t *testing.T) {
	entered, release := make(chan struct{}, 1), make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		entered <- struct{}{}
		select {
		case <-r.Context().Done():
		case <-release:
		}
	}))
	defer srv.Close()
	defer close(release)
	w := newInfluxWriter(&InfluxConfig{Url: srv.URL, Org: "org", Bucket: "bucket", WriteTimeout: 40 * time.Millisecond})
	defer w.idb.Close()
	started := time.Now()
	err := w.Handle(&Usage{ShareToken: "share", IntervalStart: time.Now(), BackendRx: 1})
	require.ErrorIs(t, err, context.DeadlineExceeded)
	require.Less(t, time.Since(started), 500*time.Millisecond)
	select {
	case <-entered:
	default:
		t.Fatal("request did not reach Influx")
	}
}

func TestInfluxWriteObservesShutdown(t *testing.T) {
	entered, release := make(chan struct{}), make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(entered)
		select {
		case <-r.Context().Done():
		case <-release:
		}
	}))
	defer srv.Close()
	defer close(release)
	w := newInfluxWriter(&InfluxConfig{Url: srv.URL, WriteTimeout: time.Minute})
	defer w.idb.Close()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() { done <- w.HandleContext(ctx, &Usage{ShareToken: "share", IntervalStart: time.Now()}) }()
	<-entered
	cancel()
	select {
	case err := <-done:
		require.ErrorIs(t, err, context.Canceled)
	case <-time.After(500 * time.Millisecond):
		t.Fatal("write ignored shutdown")
	}
}
