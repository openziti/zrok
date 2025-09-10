package dynamicProxy

import (
	"github.com/michaelquigley/df"
)

type Service struct {
	app        *df.Application[Config]
	subscriber *AmqpSubscriber
}

func NewService(cfgPath string) (*Service, error) {
	defaults := Config{}
	pp := &Service{app: df.NewApplication[Config](defaults)}
	df.WithFactoryFunc(pp.app, pp.buildSubscriber)
	if err := pp.app.Initialize(cfgPath); err != nil {
		return nil, err
	}
	return pp, nil
}

func (p *Service) Start() error {
	return p.app.Start()
}

func (p *Service) Stop() error {
	return p.app.Stop()
}

func (p *Service) buildSubscriber(app *df.Application[Config]) error {
	if app.Cfg.AmqpSubscriber == nil {
		return nil // amqp subscriber is optional
	}
	subscriber, err := NewAmqpSubscriber(app.Cfg.AmqpSubscriber)
	if err != nil {
		return err
	}
	df.Set(app.C, subscriber)
	return nil
}
