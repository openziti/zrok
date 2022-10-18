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
	cfg       *MetricsConfig
	influx    influxdb2.Client
	writeApi  api.WriteAPIBlocking
	zCtx      ziti.Context
	zListener edge.Listener
	shutdown  chan struct{}
	joined    chan struct{}
}

func newMetricsAgent(cfg *MetricsConfig) *metricsAgent {
	ma := &metricsAgent{
		cfg:      cfg,
		shutdown: make(chan struct{}),
		joined:   make(chan struct{}),
	}
	if cfg.Influx != nil {
		ma.influx = influxdb2.NewClient(cfg.Influx.Url, cfg.Influx.Token)
		ma.writeApi = ma.influx.WriteAPIBlocking(cfg.Influx.Org, cfg.Influx.Bucket)
	}
	return ma
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
	go mtr.listen()
	return nil
}

func (mtr *metricsAgent) listen() {
	logrus.Info("started")
	defer logrus.Info("exited")
	for {
		conn, err := mtr.zListener.Accept()
		if err != nil {
			logrus.Errorf("error accepting: %v", err)
			return
		}
		logrus.Debugf("accepted metrics connetion from '%v'", conn.RemoteAddr())
		go newMetricsHandler(conn, mtr.writeApi).run()
	}
}

func (mtr *metricsAgent) stop() {
	close(mtr.shutdown)
}

func (mtr *metricsAgent) join() {
	<-mtr.joined
}

type metricsHandler struct {
	conn     net.Conn
	writeApi api.WriteAPIBlocking
}

func newMetricsHandler(conn net.Conn, writeApi api.WriteAPIBlocking) *metricsHandler {
	return &metricsHandler{conn, writeApi}
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
	mtr := &model.Metrics{}
	if err := bson.Unmarshal(mtrBuf.Bytes(), &mtr); err == nil {
		out := "metrics = {\n"
		var pts []*write.Point
		for k, v := range mtr.Sessions {
			if mh.writeApi != nil {
				pt := influxdb2.NewPoint("xfer",
					map[string]string{"namespace": mtr.Namespace, "session": k},
					map[string]interface{}{"bytesRead": v.BytesRead, "bytesWritten": v.BytesWritten},
					time.UnixMilli(v.LastUpdate))
				pts = append(pts, pt)
			}
			out += fmt.Sprintf("\t[%v.%v]: %v/%v\n", mtr.Namespace, k, util.BytesToSize(v.BytesRead), util.BytesToSize(v.BytesWritten))
		}
		out += "}"
		logrus.Info(out)

		if len(pts) > 0 {
			if err := mh.writeApi.WritePoint(context.Background(), pts...); err == nil {
				logrus.Debugf("wrote metrics to influx")
			} else {
				logrus.Errorf("error writing points to influxdb: %v", err)
			}
		}

	} else {
		logrus.Errorf("error unmarshaling metrics: %v", err)
	}
}
