package udpTunnel

import (
	"github.com/openziti/sdk-golang/ziti"
	"github.com/openziti/sdk-golang/ziti/config"
	"github.com/openziti/zrok/endpoints"
	"github.com/openziti/zrok/model"
	"github.com/openziti/zrok/zrokdir"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"net"
)

type FrontendConfig struct {
	BindAddress  string
	IdentityName string
	ShrToken     string
}

type Frontend struct {
	cfg   *FrontendConfig
	zCtx  ziti.Context
	lAddr *net.UDPAddr
}

func NewFrontend(cfg *FrontendConfig) (*Frontend, error) {
	lAddr, err := net.ResolveUDPAddr("udp", cfg.BindAddress)
	if err != nil {
		return nil, errors.Wrapf(err, "error resolving udp address '%v'", cfg.BindAddress)
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
		cfg:   cfg,
		zCtx:  zCtx,
		lAddr: lAddr,
	}, nil
}

func (f *Frontend) Run() error {
	for {
		if conn, err := net.ListenUDP("udp", f.lAddr); err == nil {
			go f.accept(conn)
			logrus.Infof("accepted udp connection from '%v'", conn.RemoteAddr())
		} else {
			return err
		}
	}
}

func (f *Frontend) accept(conn *net.UDPConn) {
	if zConn, err := f.zCtx.Dial(f.cfg.ShrToken); err == nil {
		go endpoints.TXer(conn, zConn)
		go endpoints.TXer(zConn, conn)
		logrus.Infof("accepted '%v' <=> '%v'", conn.RemoteAddr(), zConn.RemoteAddr())
	} else {
		logrus.Errorf("error dialing '%v': %v", f.cfg.ShrToken, err)
		_ = conn.Close()
	}
}
