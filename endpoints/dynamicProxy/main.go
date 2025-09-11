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
	df.WithFactoryFunc(svc.app, svc.buildSubscriber)
	df.WithFactoryFunc(svc.app, svc.buildGrpcClient)
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
	subscriber, err := NewAmqpSubscriber(app.Cfg.AmqpSubscriber)
	if err != nil {
		return err
	}
	df.Set(app.C, subscriber)
	return nil
}

func (p *Service) buildGrpcClient(app *df.Application[*Config]) error {
	client, err := NewControllerClient(app.Cfg.Controller)
	if err != nil {
		return err
	}
	df.Set(app.C, client)
	return nil
}

func (p *Service) getDynamicProxyClient() *ControllerClient {
	if client, found := df.Get[*ControllerClient](p.app.C); found {
		return client
	}
	return nil
}
