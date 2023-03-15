package metrics

import (
	"context"
	"github.com/michaelquigley/cf"
	"github.com/openziti/zrok/controller/env"
	"github.com/pkg/errors"
	amqp "github.com/rabbitmq/amqp091-go"
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
	conn  *amqp.Connection
	ch    *amqp.Channel
	queue amqp.Queue
}

func newAmqpSink(cfg *AmqpSinkConfig) (*amqpSink, error) {
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

	return &amqpSink{conn, ch, queue}, nil
}

func (s *amqpSink) Handle(event ZitiEventJson) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return s.ch.PublishWithContext(ctx, "", s.queue.Name, false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        []byte(event),
	})
}
