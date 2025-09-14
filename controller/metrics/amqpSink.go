package metrics

import (
	"context"
	"time"

	"github.com/michaelquigley/df/dd"
	"github.com/pkg/errors"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/sirupsen/logrus"
)

const AmqpSinkType = "amqpSink"

type AmqpSinkConfig struct {
	Url       string `dd:"+secret"`
	QueueName string
}

func LoadAmqpSink(v map[string]any) (dd.Dynamic, error) {
	cfg, err := dd.New[AmqpSinkConfig](v)
	if err != nil {
		return nil, err
	}
	return &amqpSink{cfg: cfg}, nil
}

type amqpSink struct {
	cfg       *AmqpSinkConfig
	conn      *amqp.Connection
	ch        *amqp.Channel
	queue     amqp.Queue
	connected bool
}

func (s *amqpSink) Type() string                   { return AmqpSinkType }
func (s *amqpSink) ToMap() (map[string]any, error) { return nil, nil }

func (s *amqpSink) Handle(event ZitiEventJson) error {
	if !s.connected {
		if err := s.connect(); err != nil {
			return err
		}
		logrus.Infof("connected to '%v'", s.cfg.Url)
		s.connected = true
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	logrus.Infof("pushing '%v'", event)
	err := s.ch.PublishWithContext(ctx, "", s.queue.Name, false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        []byte(event),
	})
	if err != nil {
		s.connected = false
	}
	return err
}

func (s *amqpSink) connect() (err error) {
	s.conn, err = amqp.Dial(s.cfg.Url)
	if err != nil {
		return errors.Wrapf(err, "error dialing '%v'", s.cfg.Url)
	}
	s.ch, err = s.conn.Channel()
	if err != nil {
		return errors.Wrapf(err, "error getting amqp channel from '%v'", s.cfg.Url)
	}
	s.queue, err = s.ch.QueueDeclare(s.cfg.QueueName, true, false, false, false, nil)
	if err != nil {
		return errors.Wrapf(err, "error declaring queue '%v' with '%v'", s.cfg.QueueName, s.cfg.Url)
	}
	return nil
}
