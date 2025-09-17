package controller

import (
	"fmt"

	"github.com/openziti/zrok/controller/automation"
	"github.com/openziti/zrok/sdk/golang/sdk"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var zrokProxyConfigId string

func controllerStartup() error {
	if err := inspectZiti(); err != nil {
		return err
	}
	return nil
}

func inspectZiti() error {
	logrus.Infof("inspecting ziti controller configuration")

	ziti, err := automation.NewZitiAutomation(cfg.Ziti)
	if err != nil {
		return errors.Wrap(err, "error getting automation client")
	}
	if err := findZrokProxyConfigType(ziti); err != nil {
		return err
	}

	return nil
}

func findZrokProxyConfigType(ziti *automation.ZitiAutomation) error {
	filter := fmt.Sprintf("name=\"%v\"", sdk.ZrokProxyConfig)
	filterOpts := &automation.FilterOptions{
		Filter: filter,
		Limit:  100,
		Offset: 0,
	}

	configTypes, err := ziti.ConfigTypes.Find(filterOpts)
	if err != nil {
		return err
	}
	if len(configTypes) != 1 {
		return errors.Errorf("expected 1 zrok proxy config type, found '%d'", len(configTypes))
	}
	logrus.Infof("found '%v' config type with id '%v'", sdk.ZrokProxyConfig, *configTypes[0].ID)
	zrokProxyConfigId = *configTypes[0].ID

	return nil
}
