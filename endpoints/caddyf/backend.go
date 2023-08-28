package caddyf

import (
	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/openziti/zrok/endpoints"
	"github.com/openziti/zrok/sdk"
	"github.com/sirupsen/logrus"
)

type BackendConfig struct {
	CaddyfilePath string
	Shr           *sdk.Share
	Requests      chan *endpoints.Request
}

type Backend struct {
	cdycfg []byte
}

func NewBackend(cfg *BackendConfig) (*Backend, error) {
	cdyf, err := preprocessCaddyfile(cfg.CaddyfilePath, cfg.Shr)
	if err != nil {
		return nil, err
	}
	var adapter caddyfile.Adapter
	adapter.ServerType = httpcaddyfile.ServerType{}
	cdycfg, warnings, err := adapter.Adapt([]byte(cdyf), map[string]interface{}{"filename": cfg.CaddyfilePath})
	if err != nil {
		return nil, err
	}
	for _, warning := range warnings {
		logrus.Warnf("%v [%d] (%v): %v", cfg.CaddyfilePath, warning.Line, warning.Directive, warning.Message)
	}
	return &Backend{cdycfg: cdycfg}, nil
}

func (b *Backend) Run() error {
	if err := caddy.Run(&caddy.Config{}); err != nil {
		return err
	}
	if err := caddy.Load(b.cdycfg, true); err != nil {
		return err
	}
	return nil
}

func (b *Backend) Requests() func() int32 {
	return nil
}
