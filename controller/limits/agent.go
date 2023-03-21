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
	queue chan *metrics.Usage
	close chan struct{}
	join  chan struct{}
}

func NewAgent(cfg *Config, ifxCfg *metrics.InfluxConfig, zCfg *zrokEdgeSdk.Config, str *store.Store) (*Agent, error) {
	return &Agent{
		cfg:   cfg,
		ifx:   newInfluxReader(ifxCfg),
		zCfg:  zCfg,
		str:   str,
		queue: make(chan *metrics.Usage, 1024),
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
	logrus.Debugf("handling: %v", u)
	a.queue <- u
	return nil
}

func (a *Agent) run() {
	logrus.Info("started")
	defer logrus.Info("stopped")

mainLoop:
	for {
		select {
		case usage := <-a.queue:
			a.enforce(usage)

		case <-time.After(a.cfg.Cycle):
			logrus.Info("inspection cycle")

		case <-a.close:
			close(a.join)
			break mainLoop
		}
	}
}

func (a *Agent) enforce(u *metrics.Usage) {
	acctPeriod := 24 * time.Hour
	acctLimit := DefaultBandwidthPerPeriod()
	if a.cfg.Bandwidth != nil && a.cfg.Bandwidth.PerAccount != nil {
		acctLimit = a.cfg.Bandwidth.PerAccount
	}
	if acctLimit.Period > 0 {
		acctPeriod = acctLimit.Period
	}
	acctRx, acctTx, err := a.ifx.totalRxTxForAccount(u.AccountId, acctPeriod)
	if err != nil {
		logrus.Error(err)
	}
	if acctLimit.Warning.Rx != Unlimited && acctRx > acctLimit.Warning.Rx {
		logrus.Warnf("'%v': account over rx warning limit '%v' at '%v'", u.ShareToken, util.BytesToSize(acctLimit.Warning.Rx), util.BytesToSize(acctRx))
	}
	if acctLimit.Warning.Tx != Unlimited && acctTx > acctLimit.Warning.Tx {
		logrus.Warnf("'%v': account over tx warning limit '%v' at '%v'", u.ShareToken, util.BytesToSize(acctLimit.Warning.Tx), util.BytesToSize(acctTx))
	}
	if acctLimit.Warning.Total != Unlimited && acctTx+acctRx > acctLimit.Warning.Total {
		logrus.Warnf("'%v': account over total warning limit '%v' at '%v'", u.ShareToken, util.BytesToSize(acctLimit.Warning.Total), util.BytesToSize(acctRx+acctTx))
	}

	envPeriod := 24 * time.Hour
	envLimit := DefaultBandwidthPerPeriod()
	if a.cfg.Bandwidth != nil && a.cfg.Bandwidth.PerEnvironment != nil {
		envLimit = a.cfg.Bandwidth.PerEnvironment
	}
	if envLimit.Period > 0 {
		envPeriod = envLimit.Period
	}
	envRx, envTx, err := a.ifx.totalRxTxForEnvironment(u.EnvironmentId, envPeriod)
	if err != nil {
		logrus.Error(err)
	}

	sharePeriod := 24 * time.Hour
	shareLimit := DefaultBandwidthPerPeriod()
	if a.cfg.Bandwidth != nil && a.cfg.Bandwidth.PerShare != nil {
		shareLimit = a.cfg.Bandwidth.PerShare
	}
	if shareLimit.Period > 0 {
		sharePeriod = shareLimit.Period
	}
	shareRx, shareTx, err := a.ifx.totalRxTxForShare(u.ShareToken, sharePeriod)
	if err != nil {
		logrus.Error(err)
	}
	logrus.Infof("'%v': acct:{rx: %v, tx: %v}/%v, env:{rx: %v, tx: %v}/%v, share:{rx: %v, tx: %v}/%v",
		u.ShareToken,
		util.BytesToSize(acctRx), util.BytesToSize(acctTx), acctPeriod,
		util.BytesToSize(envRx), util.BytesToSize(envTx), envPeriod,
		util.BytesToSize(shareRx), util.BytesToSize(shareTx), sharePeriod,
	)
}
