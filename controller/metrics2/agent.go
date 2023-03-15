package metrics2

import (
	"github.com/openziti/zrok/controller/store"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type Agent struct {
	events  chan ZitiEventJson
	src     ZitiEventJsonSource
	srcJoin chan struct{}
	cache   *cache
	snk     UsageSink
}

func NewAgent(cfg *AgentConfig, str *store.Store, ifxCfg *InfluxConfig) (*Agent, error) {
	a := &Agent{}
	if v, ok := cfg.Source.(ZitiEventJsonSource); ok {
		a.src = v
	} else {
		return nil, errors.New("invalid event json source")
	}
	a.cache = newShareCache(str)
	a.snk = newInfluxWriter(ifxCfg)
	return a, nil
}

func (a *Agent) Start() error {
	a.events = make(chan ZitiEventJson)
	srcJoin, err := a.src.Start(a.events)
	if err != nil {
		return err
	}
	a.srcJoin = srcJoin

	go func() {
		logrus.Info("started")
		defer logrus.Info("stopped")
		for {
			select {
			case event := <-a.events:
				if usage, err := Ingest(event); err == nil {
					if err := a.cache.addZrokDetail(usage); err != nil {
						logrus.Error(err)
					}
					if err := a.snk.Handle(usage); err != nil {
						logrus.Error(err)
					}
				} else {
					logrus.Error(err)
				}
			}
		}
	}()

	return nil
}

func (a *Agent) Stop() {
	a.src.Stop()
	close(a.events)
}
