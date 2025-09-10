package metrics

import (
	"time"

	"github.com/michaelquigley/df"
	"github.com/pkg/errors"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/sirupsen/logrus"
)

const AmqpSourceType = "amqpSource"

type AmqpSourceConfig struct {
	Url       string `df:",secret"`
	QueueName string
}

func LoadAmqpSource(v map[string]any) (df.Dynamic, error) {
	cfg, err := df.New[AmqpSourceConfig](v)
	if err != nil {
		return nil, err
	}
	return newAmqpSource(cfg)
}

type amqpSource struct {
	cfg    *AmqpSourceConfig
	conn   *amqp.Connection
	ch     *amqp.Channel
	queue  amqp.Queue
	msgs   <-chan amqp.Delivery
	errs   chan *amqp.Error
	events chan ZitiEventMsg
	close  chan struct{}
	join   chan struct{}
}

func newAmqpSource(cfg *AmqpSourceConfig) (*amqpSource, error) {
	as := &amqpSource{
		cfg:   cfg,
		close: make(chan struct{}),
		join:  make(chan struct{}),
	}
	return as, nil
}

func (s *amqpSource) Type() string                   { return AmqpSourceType }
func (s *amqpSource) ToMap() (map[string]any, error) { return nil, nil }

func (s *amqpSource) Start(events chan ZitiEventMsg) (join chan struct{}, err error) {
	s.events = events
	go s.run()
	return s.join, nil
}

func (s *amqpSource) Stop() {
	close(s.close)
	<-s.join
}

func (s *amqpSource) run() {
	logrus.Info("started")
	defer logrus.Info("stopped")
	defer close(s.join)

mainLoop:
	for {
		logrus.Infof("connecting to '%v'", s.cfg.Url)
		if err := s.connect(); err != nil {
			logrus.Errorf("error connecting to '%v': %v", s.cfg.Url, err)
			select {
			case <-time.After(10 * time.Second):
				continue mainLoop
			case <-s.close:
				break mainLoop
			}
		}
		logrus.Infof("connected to '%v'", s.cfg.Url)

	msgLoop:
		for {
			select {
			case err, ok := <-s.errs:
				if err != nil || !ok {
					logrus.Error(err)
					break msgLoop
				}

			case <-s.close:
				break mainLoop

			case event, ok := <-s.msgs:
				if !ok {
					logrus.Debug("selecting on msg !ok")
					break msgLoop
				}
				if event.Body != nil {
					s.events <- &ZitiEventAMQP{
						data: ZitiEventJson(event.Body),
						msg:  event,
					}
				} else {
					logrus.Debug("event body was nil!")
					break msgLoop
				}
			}
		}
	}
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

	return nil
}
