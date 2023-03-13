package metrics

import (
	"github.com/openziti/zrok/controller/store"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type MetricsAgent struct {
	src   Source
	cache *cache
	join  chan struct{}
}

func Run(cfg *Config, strCfg *store.Config) (*MetricsAgent, error) {
	logrus.Info("starting")

	cache, err := newShareCache(strCfg)
	if err != nil {
		return nil, errors.Wrap(err, "error creating share cache")
	}

	if cfg.Strategies == nil || cfg.Strategies.Source == nil {
		return nil, errors.New("no 'strategies/source' configured; exiting")
	}

	src, ok := cfg.Strategies.Source.(Source)
	if !ok {
		return nil, errors.New("invalid 'strategies/source'; exiting")
	}

	if cfg.Influx == nil {
		return nil, errors.New("no 'influx' configured; exiting")
	}

	idb := openInfluxDb(cfg.Influx)

	events := make(chan map[string]interface{})
	join, err := src.Start(events)
	if err != nil {
		return nil, errors.Wrap(err, "error starting source")
	}

	go func() {
		for {
			select {
			case event := <-events:
				usage := Ingest(event)
				if err := cache.addZrokDetail(usage); err == nil {
					if err := idb.Write(usage); err != nil {
						logrus.Error(err)
					}
				} else {
					logrus.Error(err)
				}
			}
		}
	}()

	return &MetricsAgent{src: src, join: join}, nil
}

func (ma *MetricsAgent) Stop() {
	logrus.Info("stopping")
	ma.src.Stop()
}

func (ma *MetricsAgent) Join() {
	<-ma.join
}
