package controller

import (
	"github.com/sirupsen/logrus"
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
	cfg      *MetricsConfig
	shutdown chan struct{}
	joined   chan struct{}
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

	<-mtr.shutdown
}

func (mtr *metricsAgent) stop() {
	close(mtr.shutdown)
}

func (mtr *metricsAgent) join() {
	<-mtr.joined
}
