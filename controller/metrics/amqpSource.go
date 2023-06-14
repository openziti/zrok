package metrics

import (
	"github.com/michaelquigley/cf"
	"github.com/openziti/zrok/controller/env"
	"github.com/pkg/errors"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/sirupsen/logrus"
	"time"
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
	cfg   *AmqpSourceConfig
	conn  *amqp.Connection
	ch    *amqp.Channel
	queue amqp.Queue
	msgs  <-chan amqp.Delivery
	errs  chan *amqp.Error
	join  chan struct{}
}

func newAmqpSource(cfg *AmqpSourceConfig) (*amqpSource, error) {
	as := &amqpSource{cfg: cfg, join: make(chan struct{})}
	if err := as.connect(); err != nil {
		return nil, err
	}
	return as, nil
}

func (s *amqpSource) Start(events chan ZitiEventMsg) (join chan struct{}, err error) {
	go func() {
		logrus.Info("started")
		defer logrus.Info("stopped")

		reconnect := false
		for {
			if reconnect || s.errs == nil {
				if err := s.reconnect(); err != nil {
					logrus.Errorf("error reconnecting: %v", err)
					continue
				}
				reconnect = false
			}
			select {
			case event := <-s.msgs:
				if event.Body != nil {
					events <- &ZitiEventAMQP{
						data: ZitiEventJson(event.Body),
						msg:  event,
					}
				}
			case err, ok := <-s.errs:
				if err != nil || !ok {
					logrus.Error(err)
					reconnect = true
				}
			}
		}
	}()
	return s.join, nil
}

func (s *amqpSource) Stop() {
	if s.ch != nil {
		if err := s.ch.Close(); err != nil {
			logrus.Error(err)
		}
	}
	close(s.join)
}

func (s *amqpSource) connect() error {
	conn, err := amqp.Dial(s.cfg.Url)
	if err != nil {
		return errors.Wrap(err, "error dialing amqp broker")
	}

	ch, err := conn.Channel()
	if err != nil {
		return errors.Wrap(err, "error getting amqp channel")
	}

	queue, err := ch.QueueDeclare(s.cfg.QueueName, true, false, false, false, nil)
	if err != nil {
		return errors.Wrap(err, "error declaring queue")
	}

	msgs, err := ch.Consume(s.cfg.QueueName, "zrok", false, false, false, false, nil)
	if err != nil {
		return errors.Wrap(err, "error consuming")
	}

	s.errs = make(chan *amqp.Error)
	conn.NotifyClose(s.errs)
	s.conn = conn
	s.ch = ch
	s.queue = queue
	s.msgs = msgs

	logrus.Infof("connected to '%v'", s.cfg.Url)

	return nil
}

func (s *amqpSource) reconnect() error {
	s.conn = nil
	s.ch = nil
	s.msgs = nil
	s.errs = nil

	logrus.Infof("reconnecting; delay for reconnect")
	time.Sleep(10 * time.Second)
	return s.connect()
}
