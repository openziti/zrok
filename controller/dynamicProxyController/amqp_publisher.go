package dynamicProxyController

import (
	"context"
	"encoding/json"
	"time"

	"github.com/michaelquigley/df"
	"github.com/openziti/zrok/dynamicProxyModel"
	"github.com/pkg/errors"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/sirupsen/logrus"
)

type AmqpPublisherConfig struct {
	Url          string `df:"+required"`
	ExchangeName string `df:"+required"`
}

type AmqpPublisher struct {
	cfg       *AmqpPublisherConfig
	conn      *amqp.Connection
	ch        *amqp.Channel
	connected bool
}

func NewAmqpPublisher(cfg *AmqpPublisherConfig) (*AmqpPublisher, error) {
	p := &AmqpPublisher{cfg: cfg}
	if err := p.connect(); err != nil {
		return nil, err
	}
	return p, nil
}

func (p *AmqpPublisher) connect() error {
	conn, err := amqp.Dial(p.cfg.Url)
	if err != nil {
		return errors.Wrapf(err, "failed to dial amqp broker at '%s'", p.cfg.Url)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return errors.Wrap(err, "failed to create amqp channel")
	}

	// declare topic exchange for routing messages by frontend token
	err = ch.ExchangeDeclare(
		p.cfg.ExchangeName, // name
		"topic",            // type
		true,               // durable
		false,              // auto-deleted
		false,              // internal
		false,              // no-wait
		nil,                // arguments
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return errors.Wrapf(err, "failed to declare exchange '%s'", p.cfg.ExchangeName)
	}

	p.conn = conn
	p.ch = ch
	p.connected = true

	logrus.Infof("amqp publisher connected to '%s', exchange: '%s'", p.cfg.Url, p.cfg.ExchangeName)
	return nil
}

func (p *AmqpPublisher) Publish(ctx context.Context, frontendToken string, m dynamicProxyModel.Mapping) error {
	if !p.connected {
		if err := p.connect(); err != nil {
			return err
		}
	}

	data, err := df.Unbind(m)
	if err != nil {
		return errors.Wrap(err, "failed to serialize mapping")
	}

	body, err := json.Marshal(data)
	if err != nil {
		return errors.Wrap(err, "failed to marshal mapping data")
	}

	publishCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	err = p.ch.PublishWithContext(
		publishCtx,
		p.cfg.ExchangeName, // exchange
		frontendToken,      // routing key (frontend token)
		false,              // mandatory
		false,              // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp.Persistent, // persist messages for reliability
			Timestamp:    time.Now(),
		},
	)

	if err != nil {
		p.connected = false
		return errors.Wrapf(err, "failed to publish mapping update for frontend '%s'", frontendToken)
	}

	logrus.Debugf("published mapping update for frontend '%s': %+v", frontendToken, m)
	return nil
}

func (p *AmqpPublisher) Close() error {
	var errs []error

	if p.ch != nil {
		if err := p.ch.Close(); err != nil {
			errs = append(errs, errors.Wrap(err, "failed to close amqp channel"))
		}
	}

	if p.conn != nil {
		if err := p.conn.Close(); err != nil {
			errs = append(errs, errors.Wrap(err, "failed to close amqp connection"))
		}
	}

	p.connected = false

	if len(errs) > 0 {
		return errors.Errorf("errors closing amqp publisher: %v", errs)
	}

	return nil
}
