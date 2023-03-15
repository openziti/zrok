package metrics2

import (
	"github.com/michaelquigley/cf"
	"github.com/openziti/zrok/controller/env"
	"github.com/pkg/errors"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/sirupsen/logrus"
)

func init() {
	env.GetCfOptions().AddFlexibleSetter("amqpSource", loadAmqpSourceConfig)
}

type AmqpSourceConfig struct {
	Url       string `cf:"+secret"`
	QueueName string
}

func loadAmqpSourceConfig(v interface{}, _ *cf.Options) (interface{}, error) {
	if submap, ok := v.(map[string]interface{}); ok {
		cfg := &AmqpSourceConfig{}
		if err := cf.Bind(cfg, submap, cf.DefaultOptions()); err != nil {
			return nil, err
		}
		return newAmqpSource(cfg)
	}
	return nil, errors.New("invalid config structure for 'amqpSource'")
}

type amqpSource struct {
	conn  *amqp.Connection
	ch    *amqp.Channel
	queue amqp.Queue
	msgs  <-chan amqp.Delivery
	join  chan struct{}
}

func newAmqpSource(cfg *AmqpSourceConfig) (*amqpSource, error) {
	conn, err := amqp.Dial(cfg.Url)
	if err != nil {
		return nil, errors.Wrap(err, "error dialing amqp broker")
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, errors.Wrap(err, "error getting amqp channel")
	}

	queue, err := ch.QueueDeclare(cfg.QueueName, true, false, false, false, nil)
	if err != nil {
		return nil, errors.Wrap(err, "error declaring queue")
	}

	msgs, err := ch.Consume(cfg.QueueName, "zrok", true, false, false, false, nil)
	if err != nil {
		return nil, errors.Wrap(err, "error consuming")
	}

	return &amqpSource{
		conn,
		ch,
		queue,
		msgs,
		make(chan struct{}),
	}, nil
}

func (s *amqpSource) Start(events chan ZitiEventJson) (join chan struct{}, err error) {
	go func() {
		logrus.Info("started")
		defer logrus.Info("stopped")
		for event := range s.msgs {
			events <- ZitiEventJson(event.Body)
		}
		close(s.join)
	}()
	return s.join, nil
}

func (s *amqpSource) Stop() {
	if err := s.ch.Close(); err != nil {
		logrus.Error(err)
	}
	<-s.join
}
