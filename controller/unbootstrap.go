package controller

import (
	"github.com/openziti/zrok/controller/config"
	"github.com/openziti/zrok/controller/zrokEdgeSdk"
)

func Unbootstrap(cfg *config.Config) error {
	_, err := zrokEdgeSdk.Client(cfg.Ziti)
	if err != nil {
		return err
	}
	return nil
}
