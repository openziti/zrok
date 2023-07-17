package udpTunnel

import (
	"github.com/openziti/sdk-golang/ziti"
	"github.com/openziti/zrok/endpoints"
	"github.com/openziti/zrok/environment"
	"github.com/openziti/zrok/sdk"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"net"
	"sync"
	"time"
)

type FrontendConfig struct {
	BindAddress  string
	IdentityName string
	ShrToken     string
	RequestsChan chan *endpoints.Request
	IdleTime     time.Duration
}

type Frontend struct {
	cfg     *FrontendConfig
	zCtx    ziti.Context
	lAddr   *net.UDPAddr
	clients *sync.Map // map[net.Addr]*clientConn
}

type clientConn struct {
	zitiConn net.Conn
	conn     *net.UDPConn
	addr     *net.UDPAddr
	closer   func(addr *net.UDPAddr)
	active   chan bool
}

func (c *clientConn) Read(b []byte) (n int, err error) {
	panic("write only connection!")
}

func (c *clientConn) Write(b []byte) (n int, err error) {
	return c.conn.WriteTo(b, c.addr)
}

func (c *clientConn) Close() error {
	select {
	case <-c.active:
		// if it's here client as already closed
	default:
		close(c.active)
		c.closer(c.addr)
	}
	return nil
}

func (c *clientConn) LocalAddr() net.Addr {
	return c.conn.LocalAddr()
}

func (c *clientConn) RemoteAddr() net.Addr {
	return c.addr
}

func (c *clientConn) SetDeadline(t time.Time) error {
	return nil
}

func (c *clientConn) SetReadDeadline(t time.Time) error {
	return nil
}

func (c *clientConn) SetWriteDeadline(t time.Time) error {
	return nil
}

func (c *clientConn) timeout(idle time.Duration) {
	t := time.NewTimer(idle)
	for {
		select {

		case active := <-c.active:
			if active {
				t.Stop()
				t.Reset(idle)
			} else {
				break
			}

		case <-t.C:
			_ = c.Close()
			return
		}
	}
}

func NewFrontend(cfg *FrontendConfig) (*Frontend, error) {
	lAddr, err := net.ResolveUDPAddr("udp", cfg.BindAddress)
	if err != nil {
		return nil, errors.Wrapf(err, "error resolving udp address '%v'", cfg.BindAddress)
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
	zCtx, err := ziti.NewContext(zCfg)
	if err != nil {
		return nil, errors.Wrap(err, "error loading ziti context")
	}
	logrus.Errorf("creating new frontend")
	return &Frontend{
		cfg:     cfg,
		zCtx:    zCtx,
		lAddr:   lAddr,
		clients: new(sync.Map),
	}, nil
}

func (f *Frontend) Run() error {
	l, err := net.ListenUDP("udp", f.lAddr)
	if err != nil {
		return errors.Wrapf(err, "error listening at '%v'", f.lAddr)
	}
	for {
		buf := make([]byte, 16*1024)
		count, srcAddr, err := l.ReadFromUDP(buf)
		if err != nil {
			return err
		}

		c, found := f.clients.Load(srcAddr.String())
		if found {
			clt := c.(*clientConn)
			clt.active <- true
			_, err := clt.zitiConn.Write(buf[:count])
			if err != nil {
				logrus.Errorf("error writing '%v': %v", f.cfg.ShrToken, err)
				f.clients.Delete(srcAddr)
				_ = clt.zitiConn.Close()
			}
		} else {
			zitiConn, err := f.zCtx.Dial(f.cfg.ShrToken)
			if err != nil {
				logrus.Errorf("error dialing '%v': %v", f.cfg.ShrToken, err)
				continue
			}

			_, err = zitiConn.Write(buf[:count])
			if err != nil {
				logrus.Errorf("error writing '%v': %v", f.cfg.ShrToken, err)
				_ = zitiConn.Close()
				continue
			}

			clt := f.makeClient(zitiConn, l, srcAddr)
			f.clients.Store(srcAddr.String(), clt)
		}
	}
}

func (f *Frontend) notify(msg string, addr *net.UDPAddr) {
	if f.cfg.RequestsChan != nil {
		f.cfg.RequestsChan <- &endpoints.Request{
			Stamp:      time.Now(),
			RemoteAddr: addr.String(),
			Method:     msg,
			Path:       f.cfg.ShrToken,
		}
	}
}

func (f *Frontend) makeClient(zitiConn net.Conn, l *net.UDPConn, addr *net.UDPAddr) *clientConn {
	clt := &clientConn{
		zitiConn: zitiConn,
		conn:     l,
		addr:     addr,
		closer:   f.closeClient,
		active:   make(chan bool),
	}
	go clt.timeout(f.cfg.IdleTime)
	go endpoints.TXer(zitiConn, clt)

	f.notify("ACCEPT", addr)
	return clt
}

func (f *Frontend) closeClient(addr *net.UDPAddr) {
	f.notify("CLOSED", addr)
	c, found := f.clients.LoadAndDelete(addr.String())
	if found {
		clt := c.(*clientConn)
		_ = clt.zitiConn.Close()
	}
}
