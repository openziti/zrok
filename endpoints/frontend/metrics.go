package frontend

import (
	"github.com/sirupsen/logrus"
	"time"
)

type metricsAgent struct {
	metrics map[string]sessionMetrics
	updates chan metricsUpdate
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

func newMetricsAgent() *metricsAgent {
	return &metricsAgent{
		metrics: make(map[string]sessionMetrics),
		updates: make(chan metricsUpdate, 10240),
	}
}

func (ma *metricsAgent) run() {
	for {
		select {
		case update := <-ma.updates:
			logrus.Infof("update: [%v] read: %d, written: %d", update.id, update.bytesRead, update.bytesWritten)
		}
	}
}
