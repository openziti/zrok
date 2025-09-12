package dynamicProxy

import (
	"context"
	"time"

	"github.com/michaelquigley/df"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type mappings struct {
	cfg    *config
	amqp   *amqpSubscriber
	ctrl   *controllerClient
	ctx    context.Context
	cancel context.CancelFunc
}

func buildMappings(app *df.Application[*config]) error {
	mappings := newMappings()
	df.Set(app.C, mappings)
	return nil
}

func newMappings() *mappings {
	return &mappings{}
}

func (m *mappings) Link(c *df.Container) error {
	var found bool
	m.cfg, found = df.Get[*config](c)
	if !found {
		return errors.New("no config found")
	}

	m.amqp, found = df.Get[*amqpSubscriber](c)
	if !found {
		return errors.New("no amqp subscriber found")
	}

	m.ctrl, found = df.Get[*controllerClient](c)
	if !found {
		return errors.New("no controller client found")
	}
	return nil
}

func (m *mappings) Start() error {
	m.ctx, m.cancel = context.WithCancel(context.Background())
	go m.run()
	return nil
}

func (m *mappings) Stop() error {
	if m.cancel != nil {
		m.cancel()
	}
	return nil
}

func (m *mappings) run() {
	logrus.Infof("started")
	defer logrus.Infof("stopped")

	start := time.Now()
	mappings, err := m.ctrl.getAllFrontendMappings(m.cfg.AmqpSubscriber.FrontendToken, 0)
	if err != nil {
		logrus.Fatal(err)
	}
	logrus.Infof("retrieved '%d' mappings in '%v'", len(mappings), time.Since(start))

	for {
		select {
		case <-m.ctx.Done():
			return
		default:
			// do work
		}
	}
}
