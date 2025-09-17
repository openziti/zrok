package controller

import (
	"fmt"

	"github.com/openziti/zrok/controller/automation"
	"github.com/openziti/zrok/controller/config"
	"github.com/openziti/zrok/sdk/golang/sdk"
	"github.com/sirupsen/logrus"
)

func Unbootstrap(cfg *config.Config) error {
	ziti, err := automation.NewZitiAutomation(cfg.Ziti)
	if err != nil {
		return err
	}

	// cleanup all resources tagged with zrok (this handles most cleanup)
	if err := ziti.CleanupByTag("zrok", "*"); err != nil {
		logrus.Errorf("error cleaning up zrok-tagged resources: %v", err)
	}

	// cleanup the specific config type that isn't tagged with zrok
	configTypeFilter := fmt.Sprintf("name=\"%v\"", sdk.ZrokProxyConfig)
	if err := ziti.ConfigTypes.DeleteWithFilter(configTypeFilter); err != nil {
		logrus.Errorf("error unbootstrapping config type: %v", err)
	}

	return nil
}
