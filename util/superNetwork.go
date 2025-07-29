package util

import (
	"github.com/openziti/sdk-golang/ziti"
	"github.com/sirupsen/logrus"
)

func EnableSuperNetwork(zCfg *ziti.Config) {
	zCfg.MaxControlConnections = 2
	zCfg.MaxDefaultConnections = 1
	logrus.Warnf("super networking enabled")
}
