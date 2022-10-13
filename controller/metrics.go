package controller

import (
	"github.com/sirupsen/logrus"
	"time"
)

type MetricsConfig struct {
	Influx *InfluxConfig
}

type InfluxConfig struct {
	Url    string
	Bucket string
	Org    string
	Token  string
}

type metricsAgent struct {
	cfg *MetricsConfig
}

func newMetricsAgent(cfg *MetricsConfig) *metricsAgent {
	return &metricsAgent{cfg: cfg}
}

func (mtr *metricsAgent) run() {
	logrus.Info("starting")
	defer logrus.Info("exiting")

	for {
		time.Sleep(24 * time.Hour)
	}
}
