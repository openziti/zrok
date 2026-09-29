package dynamicProxy

import (
	"encoding/json"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/michaelquigley/df/dd"
	"github.com/openziti/zrok/v2/controller/dynamicProxyController"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/require"
)

// fakeAcknowledger records how a delivery was settled.
type fakeAcknowledger struct {
	acks    int
	nacks   int
	requeue bool
}

func (a *fakeAcknowledger) Ack(uint64, bool) error { a.acks++; return nil }

func (a *fakeAcknowledger) Nack(_ uint64, _ bool, requeue bool) error {
	a.nacks++
	a.requeue = a.requeue || requeue
	return nil
}

func (a *fakeAcknowledger) Reject(_ uint64, requeue bool) error {
	a.nacks++
	a.requeue = a.requeue || requeue
	return nil
}

func newTestSubscriber(t *testing.T, depth int) *amqpSubscriber {
	cfg := &config{FrontendToken: "fe", AmqpSubscriber: &amqpSubscriberConfig{QueueDepth: depth, Prefetch: 64}}
	s, err := newAmqpSubscriber(cfg)
	require.NoError(t, err)
	t.Cleanup(s.cancel)
	return s
}

// mappingBody encodes a mapping the way the controller's publisher does.
func mappingBody(t *testing.T, m dynamicProxyController.Mapping) []byte {
	data, err := dd.Unbind(m)
	require.NoError(t, err)
	body, err := json.Marshal(data)
	require.NoError(t, err)
	return body
}

// deliver runs one delivery through handling and settlement, as consume does.
func deliver(s *amqpSubscriber, body []byte) (deliveryOutcome, *fakeAcknowledger) {
	ack := &fakeAcknowledger{}
	d := amqp.Delivery{Acknowledger: ack, Body: body}
	outcome := s.handleMessage(d)
	s.settle(d, outcome)
	return outcome, ack
}

func TestSubscriberRejectsMalformedBody(t *testing.T) {
	logs := captureLogs(t)
	s := newTestSubscriber(t, 1)
	body := []byte("{not json" + strings.Repeat("x", 200))

	outcome, ack := deliver(s, body)

	require.Equal(t, deliveryRejected, outcome)
	require.Equal(t, 0, ack.acks)
	require.Equal(t, 1, ack.nacks)
	require.False(t, ack.requeue)
	require.Empty(t, s.updates)
	require.Contains(t, logs.String(), string(body[:100]))
	require.NotContains(t, logs.String(), string(body[:101]))
}

func TestSubscriberRejectsUnknownOperation(t *testing.T) {
	s := newTestSubscriber(t, 1)

	outcome, ack := deliver(s, mappingBody(t, dynamicProxyController.Mapping{Id: 1, Operation: "zzz", Name: "a", ShareToken: "sa"}))

	require.Equal(t, deliveryRejected, outcome)
	require.Equal(t, 0, ack.acks)
	require.Equal(t, 1, ack.nacks)
	require.False(t, ack.requeue)
	require.Empty(t, s.updates)
}

func TestSubscriberForwardsValidBind(t *testing.T) {
	s := newTestSubscriber(t, 1)
	want := dynamicProxyController.Mapping{Id: 7, Operation: dynamicProxyController.OperationBind, Name: "a", ShareToken: "sa"}

	outcome, ack := deliver(s, mappingBody(t, want))

	require.Equal(t, deliveryForwarded, outcome)
	require.Equal(t, 1, ack.acks)
	require.Equal(t, 0, ack.nacks)
	require.Equal(t, want, *<-s.updates)
}

func TestSubscriberFullChannelDropsAndAcks(t *testing.T) {
	logs := captureLogs(t)
	s := newTestSubscriber(t, 1)
	body := mappingBody(t, dynamicProxyController.Mapping{Id: 1, Operation: dynamicProxyController.OperationUnbind, Name: "a"})
	_, _ = deliver(s, body)

	outcome, ack := deliver(s, body)

	require.Equal(t, deliveryDropped, outcome)
	require.Equal(t, 1, ack.acks)
	require.Equal(t, 0, ack.nacks)
	require.Len(t, s.updates, 1)
	require.Contains(t, logs.String(), "updates channel full")
	require.Contains(t, logs.String(), `"level":"WARN"`)
}

func TestSubscriberShutdownNacksInFlight(t *testing.T) {
	s := newTestSubscriber(t, 1)
	s.cancel()

	outcome, ack := deliver(s, mappingBody(t, dynamicProxyController.Mapping{Id: 1, Operation: dynamicProxyController.OperationBind, Name: "a", ShareToken: "sa"}))

	require.Equal(t, deliveryCancelled, outcome)
	require.Equal(t, 1, ack.nacks)
	require.False(t, ack.requeue)
	require.Empty(t, s.updates)
}

func TestSubscriberPrefetchMustBePositive(t *testing.T) {
	for _, prefetch := range []int{0, -1} {
		_, err := newAmqpSubscriber(&config{AmqpSubscriber: &amqpSubscriberConfig{Prefetch: prefetch}})
		require.Error(t, err)
	}
}

func TestSubscriberStopDuringUnresponsiveBrokerReturnsPromptly(t *testing.T) {
	// a listener that accepts the connection and never speaks amqp, so the dial hangs in the handshake
	l, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	t.Cleanup(func() { _ = l.Close() })
	accepted := make(chan net.Conn, 1)
	go func() {
		if c, err := l.Accept(); err == nil {
			accepted <- c
		}
	}()

	cfg := &config{
		FrontendToken: "fe",
		AmqpSubscriber: &amqpSubscriberConfig{
			Url:          "amqp://guest:guest@" + l.Addr().String() + "/",
			ExchangeName: "dynamicProxy",
			QueueDepth:   1,
			Prefetch:     64,
		},
	}
	s, err := newAmqpSubscriber(cfg)
	require.NoError(t, err)
	require.NoError(t, s.Start())

	select {
	case c := <-accepted:
		t.Cleanup(func() { _ = c.Close() })
	case <-time.After(5 * time.Second):
		t.Fatal("subscriber never dialed the listener")
	}

	stopped := make(chan struct{})
	go func() {
		_ = s.Stop()
		close(stopped)
	}()
	select {
	case <-stopped:
	case <-time.After(2 * time.Second):
		t.Fatal("stop did not return while the dial was hung in the handshake")
	}
}
