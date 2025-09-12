package dynamicProxy

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/michaelquigley/df"
	"github.com/openziti/zrok/dynamicProxyModel"
	"github.com/pkg/errors"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/sirupsen/logrus"
)

type amqpSubscriberConfig struct {
	Url          string `df:"+required"`
	ExchangeName string `df:"+required"`
	QueueDepth   int
}

type amqpSubscriber struct {
	cfg        *config
	conn       *amqp.Connection
	ch         *amqp.Channel
	queue      amqp.Queue
	ctx        context.Context
	cancel     context.CancelFunc
	done       chan struct{}
	instanceId string
	updates    chan *dynamicProxyModel.Mapping
}

func buildAmqpSubscriber(app *df.Application[*config]) error {
	subscriber, err := newAmqpSubscriber(app.Cfg)
	if err != nil {
		return err
	}
	df.Set(app.C, subscriber)
	return nil
}

func newAmqpSubscriber(cfg *config) (*amqpSubscriber, error) {
	ctx, cancel := context.WithCancel(context.Background())

	s := &amqpSubscriber{
		cfg:        cfg,
		ctx:        ctx,
		cancel:     cancel,
		done:       make(chan struct{}),
		instanceId: uuid.New().String(),
		updates:    make(chan *dynamicProxyModel.Mapping, cfg.AmqpSubscriber.QueueDepth),
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
	logrus.Infof("amqp subscriber started for frontend token '%s'", s.cfg.FrontendToken)
	defer logrus.Infof("amqp subscriber stopped for frontend token '%s'", s.cfg.FrontendToken)
	defer close(s.done)

mainLoop:
	for {
		select {
		case <-s.ctx.Done():
			break mainLoop
		default:
			logrus.Infof("connecting to amqp broker at '%s'", s.cfg.AmqpSubscriber.Url)
			if err := s.connect(); err != nil {
				logrus.Errorf("failed to connect to amqp broker: %v", err)
				select {
				case <-time.After(10 * time.Second):
					continue mainLoop
				case <-s.ctx.Done():
					break mainLoop
				}
			}
			logrus.Infof("connected to amqp broker, consuming messages for frontend '%s'", s.cfg.FrontendToken)

			if err := s.consume(); err != nil {
				logrus.Errorf("consume error: %v", err)
				s.disconnect()
			}
		}
	}

	s.disconnect()
}

func (s *amqpSubscriber) connect() error {
	conn, err := amqp.Dial(s.cfg.AmqpSubscriber.Url)
	if err != nil {
		return errors.Wrapf(err, "failed to dial amqp broker at '%s'", s.cfg.AmqpSubscriber.Url)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return errors.Wrap(err, "failed to create amqp channel")
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
		ch.Close()
		conn.Close()
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
		ch.Close()
		conn.Close()
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
		ch.Close()
		conn.Close()
		return errors.Wrapf(err, "failed to bind queue '%s' to exchange '%s' with routing key '%s'",
			queue.Name, s.cfg.AmqpSubscriber.ExchangeName, s.cfg.FrontendToken)
	}

	s.conn = conn
	s.ch = ch
	s.queue = queue

	logrus.Debugf("created ephemeral queue '%s' bound to frontend token '%s'", queue.Name, s.cfg.FrontendToken)
	return nil
}

func (s *amqpSubscriber) consume() error {
	msgs, err := s.ch.Consume(
		s.queue.Name, // queue
		"",           // consumer tag (auto-generated)
		false,        // auto-ack: false (manual ack for reliability)
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

			if err := s.handleMessage(msg); err != nil {
				logrus.Errorf("failed to handle message: %v", err)
				// negative acknowledgment - message will be requeued
				msg.Nack(false, true)
			} else {
				// positive acknowledgment
				msg.Ack(false)
			}
		}
	}
}

func (s *amqpSubscriber) handleMessage(delivery amqp.Delivery) error {
	var data map[string]any
	if err := json.Unmarshal(delivery.Body, &data); err != nil {
		return errors.Wrap(err, "failed to unmarshal mapping data")
	}
	update, err := df.New[dynamicProxyModel.Mapping](data)
	if err != nil {
		return err
	}
	logrus.Infof("mapping update -> %v", update)

	select {
	case s.updates <- update:
		logrus.Debugf("published mapping update to channel")
	case <-s.ctx.Done():
		return errors.New("context cancelled while publishing update")
	default:
		logrus.Warnf("updates channel full, dropping mapping update")
	}

	return nil
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
}

func (s *amqpSubscriber) Updates() <-chan *dynamicProxyModel.Mapping {
	return s.updates
}

func (s *amqpSubscriber) generateQueueName() string {
	return "frontend-" + s.cfg.FrontendToken + "-" + s.instanceId
}
