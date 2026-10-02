package limits

import (
	"context"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/openziti/zrok/v2/controller/metrics"
	"github.com/openziti/zrok/v2/sdk/golang/sdk"
	"github.com/stretchr/testify/require"
)

// silentInflux accepts connections and never answers them. accepted receives once per connection.
func silentInflux(t *testing.T) (string, <-chan struct{}) {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	accepted := make(chan struct{}, 16)
	var mu sync.Mutex
	var conns []net.Conn
	go func() {
		for {
			conn, err := l.Accept()
			if err != nil {
				return
			}
			mu.Lock()
			conns = append(conns, conn)
			mu.Unlock()
			select {
			case accepted <- struct{}{}:
			default:
			}
		}
	}()
	t.Cleanup(func() {
		_ = l.Close()
		mu.Lock()
		defer mu.Unlock()
		for _, conn := range conns {
			_ = conn.Close()
		}
	})
	return "http://" + l.Addr().String(), accepted
}

func TestInfluxQueryHonoursTimeout(t *testing.T) {
	url, _ := silentInflux(t)
	reader := newInfluxReader(&metrics.InfluxConfig{Url: url, Bucket: "zrok", Org: "zrok"}, 200*time.Millisecond)

	start := time.Now()
	_, _, err := reader.totalRxTxForAccount(context.Background(), 1, time.Hour)
	require.Error(t, err)
	require.Less(t, time.Since(start), 5*time.Second)
}

func TestInfluxQueryEndsWhenAgentStops(t *testing.T) {
	url, accepted := silentInflux(t)
	f := newEmptyRelaxFixture(t)
	trx, err := f.str.Begin()
	require.NoError(t, err)
	addLimitedShare(t, f.str, trx, "hung@example.com", sdk.PrivateShareMode)
	require.NoError(t, trx.Commit())
	f.agent.cfg.Cycle = 10 * time.Millisecond
	f.agent.ifx = newInfluxReader(&metrics.InfluxConfig{Url: url, Bucket: "zrok", Org: "zrok"}, time.Hour)

	f.agent.Start()
	select {
	case <-accepted:
	case <-time.After(5 * time.Second):
		t.Fatal("relax cycle never queried influx")
	}

	stopped := make(chan struct{})
	go func() {
		f.agent.Stop()
		close(stopped)
	}()
	select {
	case <-stopped:
	case <-time.After(5 * time.Second):
		t.Fatal("agent did not stop while its influx query was in flight")
	}
}
