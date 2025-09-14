package dynamicProxy

import (
	"github.com/michaelquigley/df/da"
	"github.com/michaelquigley/df/dd"
	"github.com/sirupsen/logrus"
)

type Service struct {
	app *da.Application[*config]
}

func NewService(cfgPath string) (*Service, error) {
	svc := &Service{app: da.NewApplication[*config](defaults())}
	da.WithFactoryFunc(svc.app, buildAmqpSubscriber)
	da.WithFactoryFunc(svc.app, buildControllerClient)
	da.WithFactoryFunc(svc.app, buildMappings)
	da.WithFactoryFunc(svc.app, buildOAuthRouter)
	da.WithFactoryFunc(svc.app, buildHttpListener)
	opts := &dd.Options{DynamicBinders: oauthBinders}
	if err := svc.app.InitializeWithOptions(opts, cfgPath); err != nil {
		return nil, err
	}
	logrus.Info(dd.MustInspect(svc.app.Cfg))
	return svc, nil
}

func (p *Service) Start() error {
	return p.app.Start()
}

func (p *Service) Stop() error {
	return p.app.Stop()
}
