package metrics

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

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
