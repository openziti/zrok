package udpTunnel

import (
	"net"
	"time"

	"github.com/openziti/sdk-golang/ziti"
	"github.com/openziti/sdk-golang/ziti/edge"
	"github.com/openziti/zrok/endpoints"
	"github.com/openziti/zrok/util"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type BackendConfig struct {
	IdentityPath    string
	EndpointAddress string
	ShrToken        string
	RequestsChan    chan *endpoints.Request
	SuperNetwork    bool
}

type Backend struct {
	cfg      *BackendConfig
	listener edge.Listener
}

func NewBackend(cfg *BackendConfig) (*Backend, error) {
	options := ziti.ListenOptions{
		ConnectTimeout:               5 * time.Minute,
		WaitForNEstablishedListeners: 1,
	}
	zcfg, err := ziti.NewConfigFromFile(cfg.IdentityPath)
	if err != nil {
		return nil, errors.Wrap(err, "error loading config")
	}
	if cfg.SuperNetwork {
		util.EnableSuperNetwork(zcfg)
	}
	zctx, err := ziti.NewContext(zcfg)
	if err != nil {
		return nil, errors.Wrap(err, "error loading ziti context")
	}
	listener, err := zctx.ListenWithOptions(cfg.ShrToken, &options)
	if err != nil {
		return nil, errors.Wrap(err, "error listening")
	}
	b := &Backend{
		cfg:      cfg,
		listener: listener,
	}
	return b, nil
}

func (b *Backend) Run() error {
	logrus.Info("started")
	defer logrus.Info("exited")

	for {
		if conn, err := b.listener.Accept(); err == nil {
			go b.handle(conn)
		} else {
			return err
		}
	}
}

func (b *Backend) handle(conn net.Conn) {
	logrus.Debugf("handling '%v'", conn.RemoteAddr())
	if rAddr, err := net.ResolveUDPAddr("udp", b.cfg.EndpointAddress); err == nil {
		if rConn, err := net.DialUDP("udp", nil, rAddr); err == nil {
			go endpoints.TXer(conn, rConn)
			go endpoints.TXer(rConn, conn)
			if b.cfg.RequestsChan != nil {
				b.cfg.RequestsChan <- &endpoints.Request{
					Stamp:      time.Now(),
					RemoteAddr: conn.RemoteAddr().String(),
					Method:     "ACCEPT",
					Path:       rAddr.String(),
				}
			}
		} else {
			logrus.Errorf("error dialing '%v': %v", b.cfg.EndpointAddress, err)
			_ = conn.Close()
			return
		}
	} else {
		logrus.Errorf("error resolving '%v': %v", b.cfg.EndpointAddress, err)
	}
}
