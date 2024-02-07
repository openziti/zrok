package socks

import (
	"github.com/openziti/sdk-golang/ziti"
	"github.com/openziti/sdk-golang/ziti/edge"
	"github.com/openziti/zrok/endpoints"
	"github.com/pkg/errors"
	"time"
)

type BackendConfig struct {
	IdentityPath string
	ShrToken     string
	Requests     chan *endpoints.Request
}

type Backend struct {
	cfg      *BackendConfig
	listener edge.Listener
	server   *Server
}

func NewBackend(cfg *BackendConfig) (*Backend, error) {
	options := ziti.ListenOptions{
		ConnectTimeout:               5 * time.Minute,
		WaitForNEstablishedListeners: 1,
	}
	zcfg, err := ziti.NewConfigFromFile(cfg.IdentityPath)
	if err != nil {
		return nil, errors.Wrap(err, "error loading ziti identity")
	}
	zctx, err := ziti.NewContext(zcfg)
	if err != nil {
		return nil, errors.Wrap(err, "error loading ziti context")
	}
	listener, err := zctx.ListenWithOptions(cfg.ShrToken, &options)
	if err != nil {
		return nil, err
	}

	return &Backend{
		cfg:      cfg,
		listener: listener,
		server:   &Server{},
	}, nil
}

func (b *Backend) Run() error {
	if err := b.server.Serve(b.listener); err != nil {
		return err
	}
	return nil
}
