package tcpTunnel

import (
	"github.com/openziti/sdk-golang/ziti"
	"github.com/openziti/sdk-golang/ziti/config"
	"github.com/openziti/sdk-golang/ziti/edge"
	"github.com/openziti/transport/v2/tcp"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"net"
	"time"
)

type BackendConfig struct {
	IdentityPath    string
	EndpointAddress string
	ShrToken        string
}

type Backend struct {
	cfg      *BackendConfig
	listener edge.Listener
}

func NewBackend(cfg *BackendConfig) (*Backend, error) {
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
	logrus.Infof("handling '%v'", conn.RemoteAddr())
	rConn, err := tcp.Dial(b.cfg.EndpointAddress, "tcp", 30*time.Second)
	if err != nil {
		logrus.Errorf("error dialing '%v': %v", b.cfg.EndpointAddress, err)
		_ = conn.Close()
		return
	}
	go b.rxer(conn, rConn)
	go b.txer(conn, rConn)
}

func (b *Backend) rxer(conn, rConn net.Conn) {
	logrus.Infof("started '%v' <=> '%v'", conn.RemoteAddr(), rConn.RemoteAddr())
	defer logrus.Warnf("exited '%v' <=> '%v'", conn.RemoteAddr(), rConn.RemoteAddr())

	buf := make([]byte, 10240)
	for {
		if rxsz, err := conn.Read(buf); err == nil {
			if txsz, err := rConn.Write(buf[:rxsz]); err == nil {
				if txsz != rxsz {
					logrus.Errorf("short write '%v' (%d != %d)", rConn.RemoteAddr(), txsz, rxsz)
				}
			} else {
				logrus.Errorf("error writing '%v': %v", rConn.RemoteAddr(), err)
				_ = rConn.Close()
				_ = conn.Close()
				return
			}
		} else {
			logrus.Errorf("read error '%v': %v", rConn.RemoteAddr(), err)
			_ = rConn.Close()
			_ = conn.Close()
			return
		}
	}
}

func (b *Backend) txer(conn, rConn net.Conn) {
	logrus.Infof("started '%v' <=> '%v'", conn.RemoteAddr(), rConn.RemoteAddr())
	defer logrus.Warnf("exited '%v' <=> '%v'", conn.RemoteAddr(), rConn.RemoteAddr())

	buf := make([]byte, 10240)
	for {
		if rxsz, err := rConn.Read(buf); err == nil {
			if txsz, err := conn.Write(buf[:rxsz]); err == nil {
				if txsz != rxsz {
					logrus.Errorf("short write '%v' (%d != %d)", conn.RemoteAddr(), txsz, rxsz)
				}
			} else {
				logrus.Errorf("error writing '%v': %v", conn.RemoteAddr(), err)
				_ = rConn.Close()
				_ = conn.Close()
				return
			}
		} else {
			logrus.Errorf("read error '%v': %v", conn.RemoteAddr(), err)
			_ = rConn.Close()
			_ = conn.Close()
			return
		}
	}
}
