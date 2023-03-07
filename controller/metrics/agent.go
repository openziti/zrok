package metrics

import (
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type MetricsAgent struct {
	src   Source
	cache *shareCache
	join  chan struct{}
}

func Run(cfg *Config) (*MetricsAgent, error) {
	logrus.Info("starting")

	if cfg.Store == nil {
		return nil, errors.New("no 'store' configured; exiting")
	}
	cache, err := newShareCache(cfg.Store)
	if err != nil {
		return nil, errors.Wrap(err, "error creating share cache")
	}

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
				usage := Ingest(event)
				if shrToken, err := cache.getToken(usage.ZitiServiceId); err == nil {
					logrus.Infof("share: %v, circuit: %v, rx: %d, tx: %d", shrToken, usage.ZitiCircuitId, usage.BackendRx, usage.BackendTx)
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
