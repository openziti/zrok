package frontend

import (
	"encoding/json"
	"github.com/openziti-test-kitchen/zrok/model"
	"github.com/openziti-test-kitchen/zrok/zrokdir"
	"github.com/openziti/sdk-golang/ziti"
	"github.com/openziti/sdk-golang/ziti/config"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"time"
)

type metricsAgent struct {
	metrics *model.Metrics
	updates chan metricsUpdate
	zCtx    ziti.Context
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
		metrics: &model.Metrics{},
		updates: make(chan metricsUpdate, 10240),
		zCtx:    ziti.NewContextWithConfig(zCfg),
	}, nil
}

func (ma *metricsAgent) run() {
	for {
		select {
		case update := <-ma.updates:
			ma.metrics.PushSession(update.id, model.SessionMetrics{
				BytesRead:    update.bytesRead,
				BytesWritten: update.bytesWritten,
				LastUpdate:   time.Now().UnixMilli(),
			})

		case <-time.After(5 * time.Second):
			if metricsJson, err := json.MarshalIndent(ma.metrics, "", "  "); err == nil {
				logrus.Info(string(metricsJson))
			} else {
				logrus.Errorf("error marshaling metrics: %v", err)
			}
		}
	}
}
