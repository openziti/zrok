package limits

import (
	"context"
	"testing"
	"time"

	"github.com/openziti/zrok/v2/controller/metrics"
	"github.com/stretchr/testify/require"
)

func TestFullHandoffDropsWithinTimeout(t *testing.T) {
	a := &Agent{cfg: &Config{HandoffTimeout: 20 * time.Millisecond}, queue: make(chan *metrics.Usage, 1), close: make(chan struct{})}
	a.queue <- &metrics.Usage{}
	started := time.Now()
	require.NoError(t, a.Handle(&metrics.Usage{ShareToken: "share"}))
	require.GreaterOrEqual(t, time.Since(started), 20*time.Millisecond)
	require.Less(t, time.Since(started), 500*time.Millisecond)
	require.EqualValues(t, 1, a.DroppedEvents())
	require.Len(t, a.queue, 1)
}

func TestHandoffObservesConsumerShutdown(t *testing.T) {
	a := &Agent{cfg: &Config{HandoffTimeout: time.Minute}, queue: make(chan *metrics.Usage), close: make(chan struct{})}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	started := time.Now()
	require.NoError(t, a.HandleContext(ctx, &metrics.Usage{ShareToken: "share"}))
	require.Less(t, time.Since(started), 500*time.Millisecond)
	require.EqualValues(t, 1, a.DroppedEvents())
}
