package dynamicProxy

import (
	"github.com/michaelquigley/df"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type mappings struct {
	amqp *AmqpSubscriber
	ctrl *ControllerClient
}

func buildMappings(app *df.Application[*Config]) error {
	mappings := newMappings()
	df.Set(app.C, mappings)
	return nil
}

func newMappings() *mappings {
	return &mappings{}
}

func (m *mappings) Link(c *df.Container) error {
	var found bool
	m.amqp, found = df.Get[*AmqpSubscriber](c)
	if !found {
		return errors.New("no amqp subscriber found")
	}
	logrus.Infof("linked '%T'", m.amqp)

	m.ctrl, found = df.Get[*ControllerClient](c)
	if !found {
		return errors.New("no controller client found")
	}
	logrus.Infof("linked '%T'", m.ctrl)
	return nil
}

func (m *mappings) Start() error {
	logrus.Infof("started")
	return nil
}

func (m *mappings) Stop() error {
	logrus.Infof("stopped")
	return nil
}
