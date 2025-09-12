package dynamicProxy

import (
	"github.com/michaelquigley/df"
	"github.com/sirupsen/logrus"
)

type Service struct {
	app *df.Application[*config]
}

func NewService(cfgPath string) (*Service, error) {
	svc := &Service{app: df.NewApplication[*config](defaults())}
	df.WithFactoryFunc(svc.app, buildAmqpSubscriber)
	df.WithFactoryFunc(svc.app, buildControllerClient)
	df.WithFactoryFunc(svc.app, buildMappings)
	df.WithFactoryFunc(svc.app, buildHttpListener)
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
