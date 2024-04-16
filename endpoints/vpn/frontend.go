package vpn

import (
	"encoding/json"
	"github.com/net-byte/vtun/common/config"
	"github.com/net-byte/vtun/tun"
	"github.com/openziti/sdk-golang/ziti"
	"github.com/openziti/zrok/endpoints"
	"github.com/openziti/zrok/environment"
	"github.com/openziti/zrok/sdk/golang/sdk"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"net"
	"time"
)

type FrontendConfig struct {
	IdentityName string
	ShrToken     string
	RequestsChan chan *endpoints.Request
}

type Frontend struct {
	cfg  *FrontendConfig
	ztx  ziti.Context
	conn net.Conn
}

func NewFrontend(cfg *FrontendConfig) (*Frontend, error) {
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
	zCtx, err := ziti.NewContext(zCfg)
	if err != nil {
		return nil, errors.Wrap(err, "error loading ziti context")
	}

	zConn, err := zCtx.Dial(cfg.ShrToken)
	if err != nil {
		zCtx.Close()
		return nil, errors.Wrap(err, "error connecting to ziti")
	}
	return &Frontend{
		cfg:  cfg,
		ztx:  zCtx,
		conn: zConn,
	}, nil
}

func (f *Frontend) Run() error {
	var cltCfg ClientConfig
	d := json.NewDecoder(f.conn)
	if err := d.Decode(&cltCfg); err != nil {
		return errors.Wrap(err, "error decoding vpn config")
	}
	f.cfg.RequestsChan <- &endpoints.Request{
		Stamp:      time.Now(),
		RemoteAddr: cltCfg.ServerIP,
		Method:     "CONNECTED",
		Path:       cltCfg.Greeting,
	}
	logrus.Info("connected:", cltCfg.Greeting)

	defer func() {
		f.cfg.RequestsChan <- &endpoints.Request{
			Stamp:      time.Now(),
			RemoteAddr: cltCfg.ServerIP,
			Method:     "Disconnected",
		}
	}()

	cfg := config.Config{
		ServerIP:   cltCfg.ServerIP,
		CIDR:       cltCfg.CIDR,
		ServerIPv6: cltCfg.ServerIPv6,
		CIDRv6:     cltCfg.CIDR6,
		MTU:        cltCfg.MTU,
		Verbose:    false,
	}
	iface := tun.CreateTun(cfg)

	logrus.Infof("created tun device: %s", iface.Name())

	go func() {
		defer func() {
			_ = f.conn.Close()
			_ = iface.Close()
		}()

		b := make([]byte, cltCfg.MTU)
		for {
			n, err := f.conn.Read(b)
			if err != nil {
				logrus.WithError(err).Error("error reading from ziti")
				return
			}
			p := packet(b[:n])
			logrus.WithField("packet", p).Trace("received packet from peer")
			_, err = iface.Write(p)
			if err != nil {
				logrus.WithError(err).Error("error writing to device")
				return
			}
		}
	}()

	buf := make([]byte, cltCfg.MTU)
	for {
		n, err := iface.Read(buf)
		if err != nil {
			return errors.Wrap(err, "error reading packet")
		}
		pkt := packet(buf[:n])
		logrus.WithField("packet", pkt).Trace("read packet from tun device")
		_, err = f.conn.Write(pkt)

		if err != nil {
			return errors.Wrap(err, "error sending packet to ziti")
		}
	}
}
