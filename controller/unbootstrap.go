package controller

import (
	"github.com/openziti/zrok/controller/automation"
	"github.com/openziti/zrok/controller/config"
	"github.com/openziti/zrok/sdk/golang/sdk"
	"github.com/sirupsen/logrus"
)

func Unbootstrap(cfg *config.Config) error {
	automationCfg := &automation.Config{
		ApiEndpoint: cfg.Ziti.ApiEndpoint,
		Username:    cfg.Ziti.Username,
		Password:    cfg.Ziti.Password,
	}

	za, err := automation.NewZitiAutomation(automationCfg)
	if err != nil {
		return err
	}

	// cleanup all zrok-tagged resources using the automation framework
	logrus.Info("cleaning up zrok resources")
	if err := za.CleanupByTagFilter("zrok", "*"); err != nil {
		logrus.Errorf("error cleaning up zrok resources: %v", err)
	}

	// cleanup the zrok proxy config type specifically
	if err := unbootstrapConfigType(za); err != nil {
		logrus.Errorf("error unbootstrapping config type: %v", err)
	}

	return nil
}

func unbootstrapConfigType(za *automation.ZitiAutomation) error {
	// find and delete the zrok proxy config type specifically by name
	configType, err := za.ConfigTypes.GetByName(sdk.ZrokProxyConfig)
	if err != nil {
		return err
	}

	if configType != nil {
		if err := za.ConfigTypes.Delete(*configType.ID); err != nil {
			return err
		}
		logrus.Infof("deleted zrok proxy config type '%s'", *configType.ID)
	} else {
		logrus.Info("zrok proxy config type not found")
	}

	return nil
}
