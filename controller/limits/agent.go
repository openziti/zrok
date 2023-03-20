package limits

import (
	"github.com/openziti/zrok/controller/metrics"
	"github.com/openziti/zrok/controller/store"
	"github.com/openziti/zrok/controller/zrokEdgeSdk"
	"github.com/openziti/zrok/util"
	"github.com/sirupsen/logrus"
	"time"
)

type Agent struct {
	cfg   *Config
	ifx   *influxReader
	zCfg  *zrokEdgeSdk.Config
	str   *store.Store
	close chan struct{}
	join  chan struct{}
}

func NewAgent(cfg *Config, ifxCfg *metrics.InfluxConfig, zCfg *zrokEdgeSdk.Config, str *store.Store) (*Agent, error) {
	return &Agent{
		cfg:   cfg,
		ifx:   newInfluxReader(ifxCfg),
		zCfg:  zCfg,
		str:   str,
		close: make(chan struct{}),
		join:  make(chan struct{}),
	}, nil
}

func (a *Agent) Start() {
	go a.run()
}

func (a *Agent) Stop() {
	close(a.close)
	<-a.join
}

func (a *Agent) Handle(u *metrics.Usage) error {
	logrus.Infof("handling: %v", u)
	acctRx, acctTx, err := a.ifx.totalForAccount(u.AccountId, 24*time.Hour)
	if err != nil {
		logrus.Error(err)
	}
	envRx, envTx, err := a.ifx.totalForEnvironment(u.EnvironmentId, 24*time.Hour)
	if err != nil {
		logrus.Error(err)
	}
	shareRx, shareTx, err := a.ifx.totalForShare(u.ShareToken, 24*time.Hour)
	if err != nil {
		logrus.Error(err)
	}
	logrus.Infof("'%v': acct:{rx: %v, tx: %v}, env:{rx: %v, tx: %v}, share:{rx: %v, tx: %v}",
		u.ShareToken,
		util.BytesToSize(acctRx), util.BytesToSize(acctTx),
		util.BytesToSize(envRx), util.BytesToSize(envTx),
		util.BytesToSize(shareRx), util.BytesToSize(shareTx),
	)
	return nil
}

func (a *Agent) run() {
	logrus.Info("started")
	defer logrus.Info("stopped")

mainLoop:
	for {
		select {
		case <-time.After(a.cfg.Cycle):
			logrus.Info("inspection cycle")

		case <-a.close:
			close(a.join)
			break mainLoop
		}
	}
}
