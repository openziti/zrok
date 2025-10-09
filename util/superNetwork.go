package util

import (
	"github.com/michaelquigley/df/dl"
	"github.com/openziti/sdk-golang/ziti"
)

func EnableSuperNetwork(zCfg *ziti.Config) {
	zCfg.MaxControlConnections = 2
	zCfg.MaxDefaultConnections = 1
	dl.Warnf("super networking enabled")
}
