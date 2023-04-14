package tunnelBackend

import (
	"github.com/openziti/sdk-golang/ziti"
	"github.com/openziti/sdk-golang/ziti/config"
	"github.com/openziti/sdk-golang/ziti/edge"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"net"
	"time"
)

type Config struct {
	IdentityPath    string
	EndpointAddress string
	ShrToken        string
}

type Backend struct {
	cfg      *Config
	listener edge.Listener
}

func New(cfg *Config) (*Backend, error) {
	options := ziti.ListenOptions{
		ConnectTimeout: 5 * time.Minute,
		MaxConnections: 64,
	}
	zcfg, err := config.NewFromFile(cfg.IdentityPath)
	if err != nil {
		return nil, errors.Wrap(err, "error loading config")
	}
	listener, err := ziti.NewContextWithConfig(zcfg).ListenWithOptions(cfg.ShrToken, &options)
	if err == nil {
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
	if err := conn.Close(); err != nil {
		logrus.Errorf("error closing: %v", err)
	}
}
