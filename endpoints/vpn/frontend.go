package vpn

import (
	"encoding/json"
	"net"
	"time"

	"github.com/michaelquigley/df/dl"
	"github.com/net-byte/vtun/common/config"
	"github.com/net-byte/vtun/tun"
	"github.com/openziti/sdk-golang/ziti"
	"github.com/openziti/zrok/endpoints"
	"github.com/openziti/zrok/environment"
	"github.com/openziti/zrok/sdk/golang/sdk"
	"github.com/openziti/zrok/util"
	"github.com/pkg/errors"
)

type FrontendConfig struct {
	IdentityName string
	ShrToken     string
	RequestsChan chan *endpoints.Request
	SuperNetwork bool
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
	superNetwork, _ := env.SuperNetwork()
	if superNetwork {
		util.EnableSuperNetwork(zCfg)
	}
	zCtx, err := ziti.NewContext(zCfg)
	if err != nil {
		return nil, errors.Wrap(err, "error loading ziti context")
	}

	zConn, err := zCtx.DialWithOptions(cfg.ShrToken, &ziti.DialOptions{ConnectTimeout: 30 * time.Second})
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
	dl.Infof("connected: %v", cltCfg.Greeting)

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

	dl.Infof("created tun device: %s", iface.Name())

	go func() {
		defer func() {
			_ = f.conn.Close()
			_ = iface.Close()
		}()

		b := make([]byte, cltCfg.MTU)
		for {
			n, err := f.conn.Read(b)
			if err != nil {
				dl.Errorf("error reading from ziti: %v", err)
				return
			}
			p := packet(b[:n])
			dl.Log().With("packet", p).Debug("received packet from peer")
			_, err = iface.Write(p)
			if err != nil {
				dl.Errorf("error writing to device: %v", err)
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
		dl.Log().With("packet", pkt).Debug("read packet from tun device")
		_, err = f.conn.Write(pkt)

		if err != nil {
			return errors.Wrap(err, "error sending packet to ziti")
		}
	}
}
