package metrics

import (
	"github.com/michaelquigley/df/dl"
	"github.com/openziti/zrok/v2/controller/store"
	"github.com/pkg/errors"
)

type Agent struct {
	events  chan ZitiEventMsg
	src     ZitiEventJsonSource
	srcJoin chan struct{}
	cache   *cache
	snks    []UsageSink
}

func NewAgent(cfg *AgentConfig, str *store.Store, ifxCfg *InfluxConfig) (*Agent, error) {
	a := &Agent{}
	if v, ok := cfg.Source.(ZitiEventJsonSource); ok {
		a.src = v
	} else {
		return nil, errors.New("invalid event json source")
	}
	a.cache = newShareCache(str)
	a.snks = append(a.snks, newInfluxWriter(ifxCfg))
	return a, nil
}

func (a *Agent) AddUsageSink(snk UsageSink) {
	a.snks = append(a.snks, snk)
}

func (a *Agent) Start() error {
	a.events = make(chan ZitiEventMsg)
	srcJoin, err := a.src.Start(a.events)
	if err != nil {
		return err
	}
	a.srcJoin = srcJoin

	go func() {
		dl.Info("started")
		defer dl.Info("stopped")
		for {
			select {
			case event := <-a.events:
				if usage, err := Ingest(event.Data()); err == nil {
					if usage.ZitiServiceId != "" {
						if err := a.cache.addZrokDetail(usage); err != nil {
							dl.Debugf("unable to add zrok detail for: %v: %v", usage.String(), err)
						}
					}
					shouldAck := true
					for _, snk := range a.snks {
						if err := snk.Handle(usage); err != nil {
							dl.Errorf("error handling usage: %v", err)
							if shouldAck {
								shouldAck = false
							}
						}
					}
					if shouldAck {
						if err := event.Ack(); err != nil {
							dl.Errorf("unable to ack handled message: %v", err)
						}
					}
				} else {
					dl.Errorf("unable to ingest '%v': %v", event.Data(), err)
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
