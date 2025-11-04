package tcpTunnel

import (
	"net"
	"time"

	"github.com/michaelquigley/df/dl"
	"github.com/openziti/sdk-golang/ziti"
	"github.com/openziti/zrok/endpoints"
	"github.com/openziti/zrok/environment"
	"github.com/openziti/zrok/sdk/golang/sdk"
	"github.com/openziti/zrok/util"
	"github.com/pkg/errors"
)

type FrontendConfig struct {
	BindAddress  string
	IdentityName string
	ShrToken     string
	RequestsChan chan *endpoints.Request
	SuperNetwork bool
}

type Frontend struct {
	cfg   *FrontendConfig
	zCtx  ziti.Context
	lAddr *net.TCPAddr
}

func NewFrontend(cfg *FrontendConfig) (*Frontend, error) {
	lAddr, err := net.ResolveTCPAddr("tcp", cfg.BindAddress)
	if err != nil {
		return nil, errors.Wrapf(err, "error resolving tcp address '%v'", cfg.BindAddress)
	}
	env, err := environment.LoadRoot()
	if err != nil {
		return nil, errors.Wrap(err, "error loading environment root")
	}
	zCfgPath, err := env.ZitiIdentityNamed(cfg.IdentityName)
	if err != nil {
		return nil, errors.Wrapf(err, "error getting ziti identity '%v' from environment", cfg.IdentityName)
	}
	zCfg, err := ziti.NewConfigFromFile(zCfgPath)
	if err != nil {
		return nil, errors.Wrap(err, "error loading config")
	}
	zCfg.ConfigTypes = []string{sdk.ZrokProxyConfig}
	superNetwork, _ := env.SuperNetwork()
	if superNetwork {
		util.EnableSuperNetwork(zCfg)
	}
	zCtx, err := ziti.NewContext(zCfg)
	if err != nil {
		return nil, errors.Wrap(err, "error loading ziti context")
	}
	return &Frontend{
		cfg:   cfg,
		zCtx:  zCtx,
		lAddr: lAddr,
	}, nil
}

func (f *Frontend) Run() error {
	l, err := net.ListenTCP("tcp", f.lAddr)
	if err != nil {
		return errors.Wrapf(err, "error listening at '%v'", f.lAddr)
	}
	for {
		if conn, err := l.Accept(); err == nil {
			go f.accept(conn)
			dl.Debugf("accepted tcp connection from '%v'", conn.RemoteAddr())
		} else {
			return err
		}
	}
}

func (f *Frontend) accept(conn net.Conn) {
	if zConn, err := f.zCtx.DialWithOptions(f.cfg.ShrToken, &ziti.DialOptions{ConnectTimeout: 30 * time.Second}); err == nil {
		go endpoints.TXer(conn, zConn)
		go endpoints.TXer(zConn, conn)
		if f.cfg.RequestsChan != nil {
			f.cfg.RequestsChan <- &endpoints.Request{
				Stamp:      time.Now(),
				RemoteAddr: conn.RemoteAddr().String(),
				Method:     "ACCEPT",
				Path:       f.cfg.ShrToken,
			}
		}
	} else {
		dl.Errorf("error dialing '%v': %v", f.cfg.ShrToken, err)
		_ = conn.Close()
	}
}
