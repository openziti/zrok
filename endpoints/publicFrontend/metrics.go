package publicFrontend

import (
	"github.com/openziti/sdk-golang/ziti"
	"github.com/openziti/sdk-golang/ziti/config"
	"github.com/openziti/zrok/model"
	"github.com/openziti/zrok/zrokdir"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"gopkg.in/mgo.v2/bson"
	"time"
)

type metricsAgent struct {
	cfg      *Config
	accum    map[string]model.SessionMetrics
	updates  chan metricsUpdate
	lastSend time.Time
	zCtx     ziti.Context
}

type metricsUpdate struct {
	id           string
	bytesRead    int64
	bytesWritten int64
}

func newMetricsAgent(cfg *Config) (*metricsAgent, error) {
	zif, err := zrokdir.ZitiIdentityFile(cfg.Identity)
	if err != nil {
		return nil, errors.Wrapf(err, "error getting '%v' identity file", cfg.Identity)
	}
	zCfg, err := config.NewFromFile(zif)
	if err != nil {
		return nil, errors.Wrapf(err, "error loading '%v' identity", cfg.Identity)
	}
	logrus.Infof("loaded '%v' identity", cfg.Identity)
	return &metricsAgent{
		cfg:      cfg,
		accum:    make(map[string]model.SessionMetrics),
		updates:  make(chan metricsUpdate, 10240),
		lastSend: time.Now(),
		zCtx:     ziti.NewContextWithConfig(zCfg),
	}, nil
}

func (ma *metricsAgent) run() {
	for {
		select {
		case update := <-ma.updates:
			ma.pushUpdate(update)
			if time.Since(ma.lastSend) >= ma.cfg.Metrics.SendTimeout {
				if err := ma.sendMetrics(); err != nil {
					logrus.Errorf("error sending metrics: %v", err)
				}
			}

		case <-time.After(5 * time.Second):
			if err := ma.sendMetrics(); err != nil {
				logrus.Errorf("error sending metrics: %v", err)
			}
		}
	}
}

func (ma *metricsAgent) pushUpdate(mu metricsUpdate) {
	if sm, found := ma.accum[mu.id]; found {
		ma.accum[mu.id] = model.SessionMetrics{
			BytesRead:    sm.BytesRead + mu.bytesRead,
			BytesWritten: sm.BytesWritten + mu.bytesWritten,
			LastUpdate:   time.Now().UnixMilli(),
		}
	} else {
		ma.accum[mu.id] = model.SessionMetrics{
			BytesRead:    mu.bytesRead,
			BytesWritten: mu.bytesWritten,
			LastUpdate:   time.Now().UnixMilli(),
		}
	}
}

func (ma *metricsAgent) sendMetrics() error {
	if len(ma.accum) > 0 {
		m := &model.Metrics{
			Namespace: ma.cfg.Identity,
			Sessions:  ma.accum,
		}
		metricsJson, err := bson.Marshal(m)
		if err != nil {
			return errors.Wrap(err, "error marshaling metrics")
		}
		conn, err := ma.zCtx.Dial(ma.cfg.Metrics.Service)
		if err != nil {
			return errors.Wrap(err, "error connecting to metrics service")
		}
		n, err := conn.Write(metricsJson)
		if err != nil {
			return errors.Wrap(err, "error sending metrics")
		}
		defer func() { _ = conn.Close() }()
		if n != len(metricsJson) {
			return errors.Wrap(err, "short metrics write")
		}
		logrus.Infof("sent %d bytes of metrics data", n)
		ma.accum = make(map[string]model.SessionMetrics)
		ma.lastSend = time.Now()
	}
	return nil
}
