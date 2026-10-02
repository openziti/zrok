package controller

import (
	"os"

	"github.com/michaelquigley/df/dl"
	"github.com/openziti/zrok/v2/controller/automation"
	zrok_config "github.com/openziti/zrok/v2/controller/config"
	"github.com/openziti/zrok/v2/controller/limits"
	"github.com/openziti/zrok/v2/controller/store"
	"github.com/pkg/errors"
)

// RepairDialPolicies reports, and with apply creates, the dial service policies missing from live shares
// whose accounts the bandwidth limit journal does not hold limited.
func RepairDialPolicies(cfg *zrok_config.Config, apply bool) error {
	str, err := store.Open(cfg.Store)
	if err != nil {
		return errors.Wrap(err, "error opening store")
	}
	defer func() {
		if err := str.Close(); err != nil {
			dl.Errorf("error closing store: %v", err)
		}
	}()
	ziti, err := automation.NewZitiAutomation(cfg.Ziti)
	if err != nil {
		return err
	}
	_, err = limits.RepairDialPolicies(str, ziti, apply, os.Stdout)
	return err
}
