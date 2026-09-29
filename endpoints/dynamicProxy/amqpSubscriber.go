package dynamicProxy

import (
	"context"
	"encoding/json"
	"net"
	"time"

	"github.com/google/uuid"
	"github.com/michaelquigley/df/da"
	"github.com/michaelquigley/df/dd"
	"github.com/michaelquigley/df/dl"
	"github.com/openziti/zrok/v2/controller/dynamicProxyController"
	"github.com/pkg/errors"
	amqp "github.com/rabbitmq/amqp091-go"
)

type amqpSubscriberConfig struct {
	Url          string `dd:"+required"`
	ExchangeName string `dd:"+required"`
	QueueDepth   int
	Prefetch     int
}

// amqpDialTimeout bounds the tcp dial, the amqp handshake and the queue setup, each separately.
const amqpDialTimeout = 10 * time.Second

type amqpSubscriber struct {
	cfg           *config
	transport     net.Conn
	stopTransport func() bool
	conn          *amqp.Connection
	ch            *amqp.Channel
	queue         amqp.Queue
	ctx           context.Context
	cancel        context.CancelFunc
	done          chan struct{}
	instanceId    string
	updates       chan *dynamicProxyController.Mapping
}

func buildAmqpSubscriber(app *da.Application[*config]) error {
	subscriber, err := newAmqpSubscriber(app.Cfg)
	if err != nil {
		return err
	}
	da.Set(app.C, subscriber)
	return nil
}

func newAmqpSubscriber(cfg *config) (*amqpSubscriber, error) {
	if cfg.AmqpSubscriber.Prefetch <= 0 {
		return nil, errors.Errorf("amqp_subscriber.prefetch must be positive, got '%d'", cfg.AmqpSubscriber.Prefetch)
	}
	ctx, cancel := context.WithCancel(context.Background())

	s := &amqpSubscriber{
		cfg:        cfg,
		ctx:        ctx,
		cancel:     cancel,
		done:       make(chan struct{}),
		instanceId: uuid.New().String(),
		updates:    make(chan *dynamicProxyController.Mapping, cfg.AmqpSubscriber.QueueDepth),
	}

	return s, nil
}

func (s *amqpSubscriber) Start() error {
	go s.run()
	return nil
}

func (s *amqpSubscriber) Stop() error {
	s.cancel()
	<-s.done
	return nil
}

func (s *amqpSubscriber) run() {
	dl.Infof("amqp subscriber started for frontend token '%s'", s.cfg.FrontendToken)
	defer dl.Infof("amqp subscriber stopped for frontend token '%s'", s.cfg.FrontendToken)
	defer close(s.done)

mainLoop:
	for {
		select {
		case <-s.ctx.Done():
			break mainLoop
		default:
			dl.Infof("connecting to amqp broker at '%s'", s.cfg.AmqpSubscriber.Url)
			if err := s.connect(); err != nil {
				dl.Errorf("failed to connect to amqp broker: %v", err)
				select {
				case <-time.After(10 * time.Second):
					continue mainLoop
				case <-s.ctx.Done():
					break mainLoop
				}
			}
			dl.Infof("connected to amqp broker, consuming messages for frontend '%s'", s.cfg.FrontendToken)

			if err := s.consume(); err != nil {
				dl.Errorf("consume error: %v", err)
				s.disconnect()
			}
		}
	}

	s.disconnect()
}

func (s *amqpSubscriber) connect() (err error) {
	ready := false
	defer func() {
		if !ready {
			s.disconnect()
		}
	}()

	// the dial observes the shutdown context, and shutdown closes the socket, so a stop during an unreachable or
	// unresponsive broker returns at once; the deadline bounds a handshake that never completes
	conn, err := amqp.DialConfig(s.cfg.AmqpSubscriber.Url, amqp.Config{Dial: func(network, addr string) (net.Conn, error) {
		dialer := net.Dialer{Timeout: amqpDialTimeout}
		raw, err := dialer.DialContext(s.ctx, network, addr)
		if err != nil {
			return nil, err
		}
		s.transport = raw
		s.stopTransport = context.AfterFunc(s.ctx, func() { _ = raw.Close() })
		if err := raw.SetDeadline(time.Now().Add(amqpDialTimeout)); err != nil {
			return nil, err
		}
		return raw, nil
	}})
	if err != nil {
		return errors.Wrapf(err, "failed to dial amqp broker at '%s'", s.cfg.AmqpSubscriber.Url)
	}
	s.conn = conn
	// the amqp handshake clears its deadline; bound the channel and queue setup too
	if err := s.transport.SetDeadline(time.Now().Add(amqpDialTimeout)); err != nil {
		return errors.Wrap(err, "failed to set amqp setup deadline")
	}

	ch, err := conn.Channel()
	if err != nil {
		return errors.Wrap(err, "failed to create amqp channel")
	}
	s.ch = ch

	// bound the deliveries the broker hands over unacknowledged
	if err := ch.Qos(s.cfg.AmqpSubscriber.Prefetch, 0, false); err != nil {
		return errors.Wrapf(err, "failed to set prefetch '%d'", s.cfg.AmqpSubscriber.Prefetch)
	}

	// declare exchange (should already exist from publisher side)
	err = ch.ExchangeDeclare(
		s.cfg.AmqpSubscriber.ExchangeName, // name
		"topic",                           // type
		true,                              // durable
		false,                             // auto-deleted
		false,                             // internal
		false,                             // no-wait
		nil,                               // arguments
	)
	if err != nil {
		return errors.Wrapf(err, "failed to declare exchange '%s'", s.cfg.AmqpSubscriber.ExchangeName)
	}

	// create ephemeral queue for this process instance
	queueName := s.generateQueueName()
	queue, err := ch.QueueDeclare(
		queueName, // name with instance ID for uniqueness
		false,     // durable: false (ephemeral)
		true,      // delete when unused: true (auto-cleanup)
		true,      // exclusive: true (only this connection)
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		return errors.Wrapf(err, "failed to declare queue '%s'", queueName)
	}

	// bind queue to exchange with frontend token as routing key
	err = ch.QueueBind(
		queue.Name,                        // queue name
		s.cfg.FrontendToken,               // routing key (frontend token)
		s.cfg.AmqpSubscriber.ExchangeName, // exchange
		false,                             // no-wait
		nil,                               // arguments
	)
	if err != nil {
		return errors.Wrapf(err, "failed to bind queue '%s' to exchange '%s' with routing key '%s'",
			queue.Name, s.cfg.AmqpSubscriber.ExchangeName, s.cfg.FrontendToken)
	}

	if err := s.transport.SetDeadline(time.Time{}); err != nil {
		return errors.Wrap(err, "failed to clear amqp setup deadline")
	}
	s.queue = queue
	ready = true

	dl.Debugf("created ephemeral queue '%s' bound to frontend token '%s'", queue.Name, s.cfg.FrontendToken)
	return nil
}

func (s *amqpSubscriber) consume() error {
	msgs, err := s.ch.Consume(
		s.queue.Name, // queue
		"",           // consumer tag (auto-generated)
		false,        // auto-ack: false (settled per delivery, see settle)
		false,        // exclusive
		false,        // no-local
		false,        // no-wait
		nil,          // args
	)
	if err != nil {
		return errors.Wrap(err, "failed to start consuming messages")
	}

	for {
		select {
		case <-s.ctx.Done():
			return nil
		case msg, ok := <-msgs:
			if !ok {
				return errors.New("message channel closed")
			}

			s.settle(msg, s.handleMessage(msg))
		}
	}
}

// deliveryOutcome is what handleMessage did with a delivery; settle turns it into the acknowledgement.
type deliveryOutcome int

const (
	// deliveryForwarded: parsed and handed to the mapping loop; acknowledged
	deliveryForwarded deliveryOutcome = iota
	// deliveryDropped: parsed but the updates channel was full; acknowledged, reconciliation covers it
	deliveryDropped
	// deliveryRejected: unparseable or an unknown operation; nacked without requeue, never forwarded
	deliveryRejected
	// deliveryCancelled: shutdown while in flight; nacked without requeue
	deliveryCancelled
)

// settle acknowledges or rejects a delivery. nothing requeues: the queue is exclusive to this process and deleted with
// it, so a requeued message can only come back here, and a rejected message would come back forever. a message lost
// with the queue is what reconciliation recovers.
func (s *amqpSubscriber) settle(delivery amqp.Delivery, outcome deliveryOutcome) {
	var err error
	switch outcome {
	case deliveryForwarded, deliveryDropped:
		err = delivery.Ack(false)
	default:
		err = delivery.Nack(false, false)
	}
	if err != nil {
		dl.Errorf("failed to settle delivery: %v", err)
	}
}

func (s *amqpSubscriber) handleMessage(delivery amqp.Delivery) deliveryOutcome {
	update, err := parseMapping(delivery.Body)
	if err != nil {
		dl.Errorf("rejecting mapping update '%s': %v", bodyPrefix(delivery.Body), err)
		return deliveryRejected
	}

	switch update.Operation {
	case dynamicProxyController.OperationBind:
		dl.Debugf("adding mapping for '%v' -> '%v'", update.Name, update.ShareToken)

	case dynamicProxyController.OperationUnbind:
		dl.Debugf("removing mapping for '%v'", update.Name)

	default:
		dl.Errorf("rejecting mapping update '%s': unknown operation '%v'", bodyPrefix(delivery.Body), update.Operation)
		return deliveryRejected
	}

	select {
	case <-s.ctx.Done():
		return deliveryCancelled
	default:
	}

	select {
	case s.updates <- update:
		dl.Debugf("published mapping update to channel")
		return deliveryForwarded
	default:
		dl.Warnf("updates channel full, dropping mapping update for '%v'", update.Name)
		return deliveryDropped
	}
}

func parseMapping(body []byte) (*dynamicProxyController.Mapping, error) {
	var data map[string]any
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal mapping data")
	}
	return dd.New[dynamicProxyController.Mapping](data)
}

// bodyPrefix returns at most the first hundred bytes of a message body for logging.
func bodyPrefix(body []byte) []byte {
	const limit = 100
	if len(body) > limit {
		return body[:limit]
	}
	return body
}

func (s *amqpSubscriber) disconnect() {
	if s.ch != nil {
		s.ch.Close()
		s.ch = nil
	}
	if s.conn != nil {
		s.conn.Close()
		s.conn = nil
	}
	if s.stopTransport != nil {
		s.stopTransport()
		s.stopTransport = nil
	}
	if s.transport != nil {
		_ = s.transport.Close()
		s.transport = nil
	}
}

func (s *amqpSubscriber) Updates() <-chan *dynamicProxyController.Mapping {
	return s.updates
}

func (s *amqpSubscriber) generateQueueName() string {
	return "frontend-" + s.cfg.FrontendToken + "-" + s.instanceId
}
