package frontend

import (
	"fmt"
	"github.com/openziti-test-kitchen/zrok/util"
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
	bytesRead    int64
	bytesWritten int64
	lastUpdate   time.Time
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
				sm.bytesRead += update.bytesRead
				sm.bytesWritten += update.bytesWritten
				sm.lastUpdate = time.Now()
				ma.metrics[update.id] = sm
			} else {
				sm := sessionMetrics{
					bytesRead:    update.bytesRead,
					bytesWritten: update.bytesWritten,
					lastUpdate:   time.Now(),
				}
				ma.metrics[update.id] = sm
			}

		case <-time.After(5 * time.Second):
			now := time.Now()
			out := "metrics = {\n"
			for k, v := range ma.metrics {
				age := now.Sub(v.lastUpdate)
				out += fmt.Sprintf("\t[%v]: %s/%s (%s)\n", k, util.BytesToSize(v.bytesRead), util.BytesToSize(v.bytesWritten), age.String())
			}
			out += "}\n"

		}
	}
}
