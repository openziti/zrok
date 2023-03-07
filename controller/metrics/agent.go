package metrics

import (
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type MetricsAgent struct {
	src  Source
	join chan struct{}
}

func Run(cfg *Config) (*MetricsAgent, error) {
	logrus.Info("starting")

	if cfg.Source == nil {
		return nil, errors.New("no 'source' configured; exiting")
	}

	src, ok := cfg.Source.(Source)
	if !ok {
		return nil, errors.New("invalid 'source'; exiting")
	}

	events := make(chan map[string]interface{})
	join, err := src.Start(events)
	if err != nil {
		return nil, errors.Wrap(err, "error starting source")
	}

	go func() {
		for {
			select {
			case event := <-events:
				logrus.Info(Ingest(event))
			}
		}
	}()

	return &MetricsAgent{src, join}, nil
}

func (ma *MetricsAgent) Stop() {
	logrus.Info("stopping")
	ma.src.Stop()
}

func (ma *MetricsAgent) Join() {
	<-ma.join
}
