package dynamicProxy

import (
	"github.com/michaelquigley/df"
	"github.com/sirupsen/logrus"
)

type Config struct {
	V              int                   `df:"+match=1"`
	AmqpSubscriber *AmqpSubscriberConfig `df:"+required"`
}

type Service struct {
	app        *df.Application[*Config]
	subscriber *AmqpSubscriber
}

func NewService(cfgPath string) (*Service, error) {
	defaults := &Config{}
	svc := &Service{app: df.NewApplication[*Config](defaults)}
	df.WithFactoryFunc(svc.app, svc.buildSubscriber)
	if err := svc.app.Initialize(cfgPath); err != nil {
		return nil, err
	}
	logrus.Info(df.Inspect(svc.app.Cfg))
	return svc, nil
}

func (p *Service) Start() error {
	return p.app.Start()
}

func (p *Service) Stop() error {
	return p.app.Stop()
}

func (p *Service) buildSubscriber(app *df.Application[*Config]) error {
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
