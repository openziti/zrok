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
	cfg   *AmqpSinkConfig
	conn  *amqp.Connection
	ch    *amqp.Channel
	queue amqp.Queue
	join  chan struct{}
}

func newAmqpSink(cfg *AmqpSinkConfig) (*amqpSink, error) {
	return &amqpSink{
		cfg:  cfg,
		join: make(chan struct{}),
	}, nil
}

func (s *amqpSink) Start() (join chan struct{}, err error) {
	logrus.Info("started")
	return s.join, nil
}

func (s *amqpSink) Handle(event ZitiEventJson) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	logrus.Infof("pushing '%v'", event)
	return s.ch.PublishWithContext(ctx, "", s.queue.Name, false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        []byte(event),
	})
}

func (s *amqpSink) Stop() {
	close(s.join)
	logrus.Info("stopped")
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

	logrus.Infof("connected to amqp broker at '%v'", s.cfg.Url)

	return nil
}
