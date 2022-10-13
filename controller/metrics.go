package controller

import (
	"github.com/openziti-test-kitchen/zrok/zrokdir"
	"github.com/openziti/sdk-golang/ziti"
	"github.com/openziti/sdk-golang/ziti/config"
	"github.com/openziti/sdk-golang/ziti/edge"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"net"
	"time"
)

type MetricsConfig struct {
	ServiceName string
	Influx      *InfluxConfig
}

type InfluxConfig struct {
	Url    string
	Bucket string
	Org    string
	Token  string
}

type metricsAgent struct {
	cfg       *MetricsConfig
	shutdown  chan struct{}
	joined    chan struct{}
	zCtx      ziti.Context
	zListener edge.Listener
}

func newMetricsAgent(cfg *MetricsConfig) *metricsAgent {
	return &metricsAgent{
		cfg:      cfg,
		shutdown: make(chan struct{}),
		joined:   make(chan struct{}),
	}
}

func (mtr *metricsAgent) run() {
	logrus.Info("starting")
	defer logrus.Info("exiting")
	defer close(mtr.joined)

	if err := mtr.bindService(); err != nil {
		logrus.Errorf("error binding metrics service: %v", err)
		return
	}

	<-mtr.shutdown

	if err := mtr.zListener.Close(); err != nil {
		logrus.Errorf("error closing metrics service listener: %v", err)
	}
}

func (mtr *metricsAgent) bindService() error {
	zif, err := zrokdir.ZitiIdentityFile("ctrl")
	if err != nil {
		return errors.Wrap(err, "error getting 'ctrl' identity")
	}
	zCfg, err := config.NewFromFile(zif)
	if err != nil {
		return errors.Wrap(err, "error loading 'ctrl' identity")
	}
	mtr.zCtx = ziti.NewContextWithConfig(zCfg)
	opts := &ziti.ListenOptions{
		ConnectTimeout: 5 * time.Minute,
		MaxConnections: 1024,
	}
	mtr.zListener, err = mtr.zCtx.ListenWithOptions(mtr.cfg.ServiceName, opts)
	if err != nil {
		return errors.Wrapf(err, "error listening for metrics on '%v'", mtr.cfg.ServiceName)
	}
	return nil
}

func (mtr *metricsAgent) listen() {
	for {
		conn, err := mtr.zListener.Accept()
		if err != nil {
			logrus.Errorf("error accepting: %v", err)
			return
		}
		go newMetricsHandler(conn).run()
	}
}

func (mtr *metricsAgent) stop() {
	close(mtr.shutdown)
}

func (mtr *metricsAgent) join() {
	<-mtr.joined
}

type metricsHandler struct {
	conn net.Conn
}

func newMetricsHandler(conn net.Conn) *metricsHandler {
	return &metricsHandler{conn}
}

func (mh *metricsHandler) run() {
	logrus.Infof("handling metrics connection: %v", mh.conn.RemoteAddr())
	if err := mh.conn.Close(); err != nil {
		logrus.Errorf("error closing metrics connection")
	}
}
