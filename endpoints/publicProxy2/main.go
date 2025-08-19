package publicProxy2

import (
	"github.com/michaelquigley/df"
	"github.com/openziti/zrok/controller/store"
)

type Config struct {
	Store *store.Config
}

type PublicProxy struct {
	app *df.Application[Config]
}

func NewPublicProxy(cfgPath string) (*PublicProxy, error) {
	defaults := Config{}
	app := df.NewApplication[Config](defaults)
	df.WithFactory(app, &storeFactory{})
	if err := app.Initialize(cfgPath); err != nil {
		return nil, err
	}
	return &PublicProxy{app: app}, nil
}

type storeFactory struct{}

func (f *storeFactory) Build(app *df.Application[Config]) error {
	return nil
}
