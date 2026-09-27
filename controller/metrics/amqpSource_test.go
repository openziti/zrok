package metrics

import (
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestAmqpSourcePrefetchDefaults(t *testing.T) {
	for _, value := range []int{0, 7} {
		s, err := newAmqpSource(&AmqpSourceConfig{Prefetch: value})
		require.NoError(t, err)
		if value == 0 {
			require.Equal(t, 64, s.cfg.Prefetch)
		} else {
			require.Equal(t, value, s.cfg.Prefetch)
		}
		s.cancel()
	}
	_, err := newAmqpSource(&AmqpSourceConfig{Prefetch: -1})
	require.Error(t, err)
}

func TestAmqpSourceStopDuringHandshake(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer listener.Close()
	accepted := make(chan net.Conn, 1)
	go func() {
		conn, err := listener.Accept()
		if err == nil {
			accepted <- conn
		}
	}()
	s, err := newAmqpSource(&AmqpSourceConfig{Url: "amqp://guest:guest@" + listener.Addr().String(), QueueName: "events"})
	require.NoError(t, err)
	_, err = s.Start(make(chan ZitiEventMsg))
	require.NoError(t, err)
	t.Cleanup(s.Stop)
	select {
	case conn := <-accepted:
		defer conn.Close()
	case <-time.After(time.Second):
		t.Fatal("source did not connect")
	}
	done := make(chan struct{})
	go func() { s.Stop(); close(done) }()
	select {
	case <-done:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("source did not cancel handshake")
	}
}
