package limits

import (
	"github.com/openziti/zrok/controller/metrics"
	"github.com/openziti/zrok/controller/store"
	"github.com/openziti/zrok/controller/zrokEdgeSdk"
	"github.com/sirupsen/logrus"
)

type Agent struct {
}

func NewAgent(cfg *Config, ifxCfg *metrics.InfluxConfig, zCfg *zrokEdgeSdk.Config, str *store.Store) (*Agent, error) {
	return &Agent{}, nil
}

func (a *Agent) Start() error {
	return nil
}

func (a *Agent) Stop() {
}

func (a *Agent) Handle(u *metrics.Usage) error {
	logrus.Infof("handling: %v", u)
	return nil
}
