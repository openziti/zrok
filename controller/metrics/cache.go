package metrics

import (
	"github.com/openziti/zrok/v2/controller/store"
)

type cache struct {
	str *store.Store
}

func newShareCache(str *store.Store) *cache {
	return &cache{str}
}

func (c *cache) addZrokDetail(u *Usage) error {
	trx, err := c.str.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = trx.Rollback() }()

	shr, err := c.str.FindShareWithZIdAndDeleted(u.ZitiServiceId, trx)
	if err != nil {
		return err
	}
	u.ShareToken = shr.Token
	env, err := c.str.GetEnvironment(shr.EnvironmentId, trx)
	if err != nil {
		return err
	}
	u.EnvironmentId = int64(env.Id)
	u.AccountId = int64(*env.AccountId)

	return nil
}
