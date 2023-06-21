package metrics

import (
	"context"
	"github.com/michaelquigley/cf"
	"github.com/openziti/zrok/controller/env"
	"github.com/pkg/errors"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/sirupsen/logrus"
	"time"
)

func init() {
	env.GetCfOptions().AddFlexibleSetter("amqpSink", loadAmqpSinkConfig)
}

type AmqpSinkConfig struct {
	Url       string `cf:"+secret"`
	QueueName string
}

func loadAmqpSinkConfig(v interface{}, _ *cf.Options) (interface{}, error) {
	if submap, ok := v.(map[string]interface{}); ok {
		cfg := &AmqpSinkConfig{}
		if err := cf.Bind(cfg, submap, cf.DefaultOptions()); err != nil {
			return nil, err
		}
		return newAmqpSink(cfg)
	}
	return nil, errors.New("invalid config structure for 'amqpSink'")
}

type amqpSink struct {
	cfg       *AmqpSinkConfig
	conn      *amqp.Connection
	ch        *amqp.Channel
	queue     amqp.Queue
	connected bool
}

func newAmqpSink(cfg *AmqpSinkConfig) (*amqpSink, error) {
	as := &amqpSink{cfg: cfg}
	return as, nil
}

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
		return errors.Wrap(err, "error dialing amqp broker")
	}
	s.ch, err = s.conn.Channel()
	if err != nil {
		return errors.Wrap(err, "error getting amqp channel")
	}
	s.queue, err = s.ch.QueueDeclare(s.cfg.QueueName, true, false, false, false, nil)
	if err != nil {
		return errors.Wrap(err, "error declaring queue")
	}
	return nil
}
