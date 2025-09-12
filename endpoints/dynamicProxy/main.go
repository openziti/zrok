package dynamicProxy

import (
	"time"

	"github.com/michaelquigley/df"
	"github.com/sirupsen/logrus"
)

type Config struct {
	V              int                     `df:"+match=1"`
	AmqpSubscriber *AmqpSubscriberConfig   `df:"+required"`
	Controller     *ControllerClientConfig `df:"+required"`
}

type Service struct {
	app *df.Application[*Config]
}

func NewService(cfgPath string) (*Service, error) {
	defaults := &Config{
		Controller: &ControllerClientConfig{
			Timeout: 30 * time.Second,
		},
	}
	svc := &Service{app: df.NewApplication[*Config](defaults)}
	df.WithFactoryFunc(svc.app, buildAmqpSubscriber)
	df.WithFactoryFunc(svc.app, buildControllerClient)
	df.WithFactoryFunc(svc.app, buildMappings)
	if err := svc.app.Initialize(cfgPath); err != nil {
		return nil, err
	}
	logrus.Info(df.MustInspect(svc.app.Cfg))
	return svc, nil
}

func (p *Service) Start() error {
	return p.app.Start()
}

func (p *Service) Stop() error {
	return p.app.Stop()
}

func (p *Service) getAmqpSubscriber() *AmqpSubscriber {
	if subscriber, found := df.Get[*AmqpSubscriber](p.app.C); found {
		return subscriber
	}
	return nil
}

func (p *Service) getControllerClient() *ControllerClient {
	if client, found := df.Get[*ControllerClient](p.app.C); found {
		return client
	}
	return nil
}
