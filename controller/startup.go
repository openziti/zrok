package controller

import (
	"fmt"

	"github.com/michaelquigley/df/dl"
	"github.com/openziti/zrok/v2/controller/automation"
	"github.com/openziti/zrok/v2/sdk/golang/sdk"
	"github.com/pkg/errors"
)

var zrokProxyConfigId string

func controllerStartup() error {
	if err := inspectZiti(); err != nil {
		return err
	}
	return nil
}

func inspectZiti() error {
	dl.Infof("inspecting ziti controller configuration")

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
	dl.Infof("found '%v' config type with id '%v'", sdk.ZrokProxyConfig, *configTypes[0].ID)
	zrokProxyConfigId = *configTypes[0].ID

	return nil
}
