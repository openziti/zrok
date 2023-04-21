package tcpTunnel

import (
	"github.com/openziti/sdk-golang/ziti"
	"github.com/openziti/sdk-golang/ziti/config"
	"github.com/openziti/transport/v2"
	"github.com/openziti/zrok/endpoints"
	"github.com/openziti/zrok/model"
	"github.com/openziti/zrok/zrokdir"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"io"
)

type FrontendConfig struct {
	BindAddress  string
	IdentityName string
	ShrToken     string
}

type Frontend struct {
	cfg      *FrontendConfig
	zCtx     ziti.Context
	listener transport.Address
	closer   io.Closer
}

func NewFrontend(cfg *FrontendConfig) (*Frontend, error) {
	addr, err := transport.ParseAddress(cfg.BindAddress)
	if err != nil {
		return nil, errors.Wrapf(err, "error parsing '%v'", cfg.BindAddress)
	}
	zCfgPath, err := zrokdir.ZitiIdentityFile(cfg.IdentityName)
	if err != nil {
		return nil, errors.Wrapf(err, "error getting ziti identity '%v' from zrokdir", cfg.IdentityName)
	}
	zCfg, err := config.NewFromFile(zCfgPath)
	if err != nil {
		return nil, errors.Wrap(err, "error loading config")
	}
	zCfg.ConfigTypes = []string{model.ZrokProxyConfig}
	zCtx := ziti.NewContextWithConfig(zCfg)
	return &Frontend{
		cfg:      cfg,
		zCtx:     zCtx,
		listener: addr,
	}, nil
}

func (f *Frontend) Run() error {
	closer, err := f.listener.Listen(f.cfg.ShrToken, nil, f.accept, nil)
	if err != nil {
		return err
	}
	f.closer = closer
	return nil
}

func (f *Frontend) Stop() {
	if f.closer != nil {
		_ = f.closer.Close()
	}
}

func (f *Frontend) accept(conn transport.Conn) {
	if zConn, err := f.zCtx.Dial(f.cfg.ShrToken); err == nil {
		go endpoints.TXer(conn, zConn)
		go endpoints.TXer(zConn, conn)
		logrus.Infof("accepted '%v' <=> '%v'", conn.RemoteAddr(), zConn.RemoteAddr())
	} else {
		logrus.Errorf("error dialing '%v': %v", f.cfg.ShrToken, err)
		_ = conn.Close()
	}
}
