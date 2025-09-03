package dynamicProxy

import (
	"github.com/michaelquigley/df"
	"github.com/openziti/zrok/endpoints/dynamicProxy/config"
	"github.com/openziti/zrok/endpoints/dynamicProxy/store"
)

type Service struct {
	app *df.Application[config.Config]
}

func NewService(cfgPath string) (*Service, error) {
	defaults := config.Config{}
	pp := &Service{app: df.NewApplication[config.Config](defaults)}
	df.WithFactoryFunc(pp.app, pp.buildStore)
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

func (p *Service) buildStore(app *df.Application[config.Config]) error {
	str, err := store.Open(app.Cfg.Store)
	if err != nil {
		return err
	}
	df.Set(app.C, str)
	return nil
}
