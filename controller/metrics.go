package controller

import (
	"bytes"
	"context"
	"fmt"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/influxdata/influxdb-client-go/v2/api/write"
	"github.com/openziti-test-kitchen/zrok/model"
	"github.com/openziti-test-kitchen/zrok/util"
	"github.com/openziti-test-kitchen/zrok/zrokdir"
	"github.com/openziti/sdk-golang/ziti"
	"github.com/openziti/sdk-golang/ziti/config"
	"github.com/openziti/sdk-golang/ziti/edge"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"gopkg.in/mgo.v2/bson"
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
	cfg          *MetricsConfig
	influx       influxdb2.Client
	writeApi     api.WriteAPIBlocking
	metricsQueue chan *model.Metrics
	envCache     map[string]*envCacheEntry
	zCtx         ziti.Context
	zListener    edge.Listener
	shutdown     chan struct{}
	joined       chan struct{}
}

type envCacheEntry struct {
	env        string
	lastAccess time.Time
}

func newMetricsAgent(cfg *MetricsConfig) *metricsAgent {
	ma := &metricsAgent{
		cfg:          cfg,
		metricsQueue: make(chan *model.Metrics, 1024),
		envCache:     make(map[string]*envCacheEntry),
		shutdown:     make(chan struct{}),
		joined:       make(chan struct{}),
	}
	if cfg.Influx != nil {
		ma.influx = influxdb2.NewClient(cfg.Influx.Url, cfg.Influx.Token)
		ma.writeApi = ma.influx.WriteAPIBlocking(cfg.Influx.Org, cfg.Influx.Bucket)
	}
	return ma
}

func (ma *metricsAgent) run() {
	logrus.Info("starting")
	defer logrus.Info("exiting")
	defer close(ma.joined)

	if err := ma.bindService(); err != nil {
		logrus.Errorf("error binding metrics service: %v", err)
		return
	}

work:
	for {
		select {
		case <-ma.shutdown:
			break work

		case m := <-ma.metricsQueue:
			if err := ma.processMetrics(m); err != nil {
				logrus.Errorf("error processing metrics: %v", err)
			}
		}
	}

	if err := ma.zListener.Close(); err != nil {
		logrus.Errorf("error closing metrics service listener: %v", err)
	}
}

func (ma *metricsAgent) bindService() error {
	zif, err := zrokdir.ZitiIdentityFile("ctrl")
	if err != nil {
		return errors.Wrap(err, "error getting 'ctrl' identity")
	}
	zCfg, err := config.NewFromFile(zif)
	if err != nil {
		return errors.Wrap(err, "error loading 'ctrl' identity")
	}
	ma.zCtx = ziti.NewContextWithConfig(zCfg)
	opts := &ziti.ListenOptions{
		ConnectTimeout: 5 * time.Minute,
		MaxConnections: 1024,
	}
	ma.zListener, err = ma.zCtx.ListenWithOptions(ma.cfg.ServiceName, opts)
	if err != nil {
		return errors.Wrapf(err, "error listening for metrics on '%v'", ma.cfg.ServiceName)
	}
	go ma.listen()
	return nil
}

func (ma *metricsAgent) listen() {
	logrus.Info("started")
	defer logrus.Info("exited")
	for {
		conn, err := ma.zListener.Accept()
		if err != nil {
			logrus.Errorf("error accepting: %v", err)
			return
		}
		logrus.Debugf("accepted metrics connetion from '%v'", conn.RemoteAddr())
		go newMetricsHandler(conn, ma.metricsQueue).run()
	}
}

func (ma *metricsAgent) processMetrics(m *model.Metrics) error {
	var pts []*write.Point
	if len(m.Sessions) > 0 {
		out := "metrics = {\n"
		for k, v := range m.Sessions {
			if ma.writeApi != nil {
				pt := influxdb2.NewPoint("xfer",
					map[string]string{"namespace": m.Namespace, "session": k},
					map[string]interface{}{"bytesRead": v.BytesRead, "bytesWritten": v.BytesWritten},
					time.UnixMilli(v.LastUpdate))
				pts = append(pts, pt)
			}
			out += fmt.Sprintf("\t[%v.%v]: %v/%v (%v)\n", m.Namespace, k, util.BytesToSize(v.BytesRead), util.BytesToSize(v.BytesWritten), time.Since(time.UnixMilli(v.LastUpdate)))
		}
		out += "}"
		logrus.Info(out)
	}

	if len(pts) > 0 {
		if err := ma.writeApi.WritePoint(context.Background(), pts...); err == nil {
			logrus.Debugf("wrote metrics to influx")
		} else {
			return err
		}
	}

	return nil
}

func (ma *metricsAgent) stop() {
	close(ma.shutdown)
}

func (ma *metricsAgent) join() {
	<-ma.joined
}

type metricsHandler struct {
	conn         net.Conn
	metricsQueue chan *model.Metrics
}

func newMetricsHandler(conn net.Conn, metricsQueue chan *model.Metrics) *metricsHandler {
	return &metricsHandler{conn, metricsQueue}
}

func (mh *metricsHandler) run() {
	logrus.Debugf("handling metrics connection: %v", mh.conn.RemoteAddr())
	var mtrBuf bytes.Buffer
	buf := make([]byte, 4096)
	for {
		n, err := mh.conn.Read(buf)
		if err != nil {
			break
		}
		mtrBuf.Write(buf[:n])
	}
	if err := mh.conn.Close(); err != nil {
		logrus.Errorf("error closing metrics connection")
	}
	m := &model.Metrics{}
	if err := bson.Unmarshal(mtrBuf.Bytes(), &m); err == nil {
		mh.metricsQueue <- m
	} else {
		logrus.Errorf("error unmarshaling metrics: %v", err)
	}
}
