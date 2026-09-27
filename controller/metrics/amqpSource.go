package metrics

import (
	"context"
	"net"
	"time"

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
	Prefetch  int
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
	cfg           *AmqpSourceConfig
	conn          *amqp.Connection
	ch            *amqp.Channel
	queue         amqp.Queue
	msgs          <-chan amqp.Delivery
	errs          chan *amqp.Error
	events        chan ZitiEventMsg
	ctx           context.Context
	cancel        context.CancelFunc
	transport     net.Conn
	stopTransport func() bool
	join          chan struct{}
}

func newAmqpSource(cfg *AmqpSourceConfig) (*amqpSource, error) {
	if cfg.Prefetch < 0 {
		return nil, errors.New("prefetch must be positive")
	}
	config := *cfg
	if config.Prefetch == 0 {
		config.Prefetch = 64
	}
	ctx, cancel := context.WithCancel(context.Background())
	as := &amqpSource{
		cfg:    &config,
		ctx:    ctx,
		cancel: cancel,
		join:   make(chan struct{}),
	}
	return as, nil
}

func (s *amqpSource) Start(events chan ZitiEventMsg) (join chan struct{}, err error) {
	s.events = events
	go s.run()
	return s.join, nil
}

func (s *amqpSource) Stop() {
	s.cancel()
	<-s.join
}

func (s *amqpSource) run() {
	logrus.Info("started")
	defer close(s.join)
	defer logrus.Info("stopped")
	defer s.disconnect()

mainLoop:
	for {
		if s.ctx.Err() != nil {
			break mainLoop
		}
		s.disconnect()
		logrus.Infof("connecting to '%v'", s.cfg.Url)
		if err := s.connect(); err != nil {
			logrus.Errorf("error connecting to '%v': %v", s.cfg.Url, err)
			select {
			case <-time.After(10 * time.Second):
				continue mainLoop
			case <-s.ctx.Done():
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

			case <-s.ctx.Done():
				break mainLoop

			case event, ok := <-s.msgs:
				if !ok {
					logrus.Debug("selecting on msg !ok")
					break msgLoop
				}
				select {
				case s.events <- &ZitiEventAMQP{
					data: ZitiEventJson(event.Body),
					msg:  event,
				}:
				case <-s.ctx.Done():
					// leave the delivery unsettled for the broker to requeue on connection close.
					break mainLoop
				}
			}
		}
	}
}

func (s *amqpSource) connect() (err error) {
	ready := false
	defer func() {
		if !ready {
			s.disconnect()
		}
	}()
	conn, err := amqp.DialConfig(s.cfg.Url, amqp.Config{Dial: func(network, addr string) (net.Conn, error) {
		dialer := net.Dialer{Timeout: 10 * time.Second}
		raw, err := dialer.DialContext(s.ctx, network, addr)
		if err != nil {
			return nil, err
		}
		s.transport = raw
		s.stopTransport = context.AfterFunc(s.ctx, func() { _ = raw.Close() })
		if err := raw.SetDeadline(time.Now().Add(10 * time.Second)); err != nil {
			return nil, err
		}
		return raw, nil
	}})
	if err != nil {
		return errors.Wrap(err, "error dialing amqp broker")
	}
	// the AMQP handshake clears its deadline; bound queue and consumer setup too.
	if err := s.transport.SetDeadline(time.Now().Add(10 * time.Second)); err != nil {
		return err
	}

	ch, err := conn.Channel()
	if err != nil {
		return errors.Wrap(err, "error getting amqp channel")
	}

	queue, err := ch.QueueDeclare(s.cfg.QueueName, true, false, false, false, nil)
	if err != nil {
		return errors.Wrap(err, "error declaring queue")
	}
	if err := ch.Qos(s.cfg.Prefetch, 0, false); err != nil {
		return errors.Wrap(err, "error setting prefetch")
	}

	msgs, err := ch.Consume(s.cfg.QueueName, "zrok", false, false, false, false, nil)
	if err != nil {
		return errors.Wrap(err, "error consuming")
	}

	if err := s.transport.SetDeadline(time.Time{}); err != nil {
		return err
	}
	s.errs = make(chan *amqp.Error, 1)
	conn.NotifyClose(s.errs)
	s.conn = conn
	s.ch = ch
	s.queue = queue
	s.msgs = msgs
	ready = true

	return nil
}

func (s *amqpSource) disconnect() {
	if s.stopTransport != nil {
		s.stopTransport()
		s.stopTransport = nil
	}
	if s.transport != nil {
		_ = s.transport.Close()
		s.transport = nil
	}
}
