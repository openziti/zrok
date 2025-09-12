package dynamicProxy

import (
	"time"

	"github.com/michaelquigley/df"
	"github.com/sirupsen/logrus"
)

type config struct {
	V              int                     `df:"+match=1"`
	AmqpSubscriber *amqpSubscriberConfig   `df:"+required"`
	Controller     *controllerClientConfig `df:"+required"`
}

type Service struct {
	app *df.Application[*config]
}

func NewService(cfgPath string) (*Service, error) {
	defaults := &config{
		Controller: &controllerClientConfig{
			Timeout: 30 * time.Second,
		},
	}
	svc := &Service{app: df.NewApplication[*config](defaults)}
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

func (p *Service) getAmqpSubscriber() *amqpSubscriber {
	if subscriber, found := df.Get[*amqpSubscriber](p.app.C); found {
		return subscriber
	}
	return nil
}

func (p *Service) getControllerClient() *controllerClient {
	if client, found := df.Get[*controllerClient](p.app.C); found {
		return client
	}
	return nil
}
