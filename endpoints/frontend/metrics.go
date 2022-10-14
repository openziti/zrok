package frontend

import (
	"encoding/json"
	"github.com/openziti-test-kitchen/zrok/zrokdir"
	"github.com/openziti/sdk-golang/ziti"
	"github.com/openziti/sdk-golang/ziti/config"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"time"
)

type metricsAgent struct {
	metrics map[string]sessionMetrics
	updates chan metricsUpdate
	zCtx    ziti.Context
}

type sessionMetrics struct {
	BytesRead    int64
	BytesWritten int64
	LastUpdate   time.Time
}

type metricsUpdate struct {
	id           string
	bytesRead    int64
	bytesWritten int64
}

func newMetricsAgent(identityName string) (*metricsAgent, error) {
	zif, err := zrokdir.ZitiIdentityFile(identityName)
	if err != nil {
		return nil, errors.Wrapf(err, "error getting '%v' identity file", identityName)
	}
	zCfg, err := config.NewFromFile(zif)
	if err != nil {
		return nil, errors.Wrapf(err, "error loading '%v' identity", identityName)
	}
	logrus.Infof("loaded '%v' identity", identityName)
	return &metricsAgent{
		metrics: make(map[string]sessionMetrics),
		updates: make(chan metricsUpdate, 10240),
		zCtx:    ziti.NewContextWithConfig(zCfg),
	}, nil
}

func (ma *metricsAgent) run() {
	for {
		select {
		case update := <-ma.updates:
			if sm, found := ma.metrics[update.id]; found {
				sm.BytesRead += update.bytesRead
				sm.BytesWritten += update.bytesWritten
				sm.LastUpdate = time.Now()
				ma.metrics[update.id] = sm
			} else {
				sm := sessionMetrics{
					BytesRead:    update.bytesRead,
					BytesWritten: update.bytesWritten,
					LastUpdate:   time.Now(),
				}
				ma.metrics[update.id] = sm
			}

		case <-time.After(5 * time.Second):
			if metricsJson, err := json.MarshalIndent(ma.metrics, "", "  "); err == nil {
				logrus.Info(string(metricsJson))
			} else {
				logrus.Errorf("error marshaling metrics: %v", err)
			}
		}
	}
}
