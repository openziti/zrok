package webBackend

import (
	"github.com/openziti/sdk-golang/ziti"
	"github.com/openziti/sdk-golang/ziti/config"
	"github.com/openziti/sdk-golang/ziti/edge"
	"github.com/pkg/errors"
	"net/http"
	"time"
)

type Config struct {
	IdentityPath string
	WebRoot      string
	ShrToken     string
}

type backend struct {
	cfg      *Config
	listener edge.Listener
	handler  http.Handler
}

func NewBackend(cfg *Config) (*backend, error) {
	options := ziti.ListenOptions{
		ConnectTimeout: 5 * time.Minute,
		MaxConnections: 64,
	}
	zcfg, err := config.NewFromFile(cfg.IdentityPath)
	if err != nil {
		return nil, errors.Wrap(err, "error loading config")
	}
	listener, err := ziti.NewContextWithConfig(zcfg).ListenWithOptions(cfg.ShrToken, &options)
	if err != nil {
		return nil, errors.Wrap(err, "error listening")
	}

	return &backend{
		cfg:      cfg,
		listener: listener,
		handler:  http.FileServer(http.Dir(cfg.WebRoot)),
	}, nil
}

func (self *backend) Run() error {
	if err := http.Serve(self.listener, self.handler); err != nil {
		return err
	}
	return nil
}

func (self *backend) Requests() func() int32 {
	return func() int32 { return 0 }
}
