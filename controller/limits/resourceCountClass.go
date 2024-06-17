package limits

import (
	"fmt"
	"github.com/openziti/zrok/controller/store"
)

type configResourceCountClass struct {
	cfg *Config
}

func newConfigResourceCountClass(cfg *Config) store.ResourceCountClass {
	return &configResourceCountClass{cfg}
}

func (rcc *configResourceCountClass) IsGlobal() bool {
	return true
}

func (rcc *configResourceCountClass) GetLimitClassId() int {
	return -1
}

func (rcc *configResourceCountClass) GetEnvironments() int {
	return rcc.cfg.Environments
}

func (rcc *configResourceCountClass) GetShares() int {
	return rcc.cfg.Shares
}

func (rcc *configResourceCountClass) GetReservedShares() int {
	return rcc.cfg.ReservedShares
}

func (rcc *configResourceCountClass) GetUniqueNames() int {
	return rcc.cfg.UniqueNames
}

func (rcc *configResourceCountClass) GetShareFrontends() int {
	return rcc.cfg.ShareFrontends
}

func (rcc *configResourceCountClass) String() string {
	return fmt.Sprintf("Config<environments: %d, shares: %d, reservedShares: %d, uniqueNames: %d>", rcc.cfg.Environments, rcc.cfg.Shares, rcc.cfg.ReservedShares, rcc.cfg.UniqueNames)
}
