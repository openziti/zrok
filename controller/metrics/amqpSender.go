package metrics

import (
	"context"
	"github.com/pkg/errors"
	amqp "github.com/rabbitmq/amqp091-go"
	"time"
)

type AmqpSenderConfig struct {
	Url   string `cf:"+secret"`
	Queue string
}

type AmqpSender struct {
	conn  *amqp.Connection
	ch    *amqp.Channel
	queue amqp.Queue
}

func NewAmqpSender(cfg *AmqpSenderConfig) (*AmqpSender, error) {
	conn, err := amqp.Dial(cfg.Url)
	if err != nil {
		return nil, errors.Wrap(err, "error dialing amqp broker")
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, errors.Wrap(err, "error getting channel from amqp connection")
	}

	queue, err := ch.QueueDeclare(cfg.Queue, true, false, false, false, nil)
	if err != nil {
		return nil, errors.Wrap(err, "error creating amqp queue")
	}

	return &AmqpSender{conn, ch, queue}, nil
}

func (s *AmqpSender) Send(json string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err := s.ch.PublishWithContext(ctx, "", s.queue.Name, false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        []byte(json),
	})
	if err != nil {
		return errors.Wrap(err, "error sending")
	}
	return nil
}
