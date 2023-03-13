package metrics

import (
	"github.com/openziti/zrok/controller/store"
	"github.com/pkg/errors"
)

type cache struct {
	str *store.Store
}

func newShareCache(cfg *store.Config) (*cache, error) {
	str, err := store.Open(cfg)
	if err != nil {
		return nil, errors.Wrap(err, "error opening store")
	}
	return &cache{str}, nil
}

func (sc *cache) addZrokDetail(u *Usage) error {
	tx, err := sc.str.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	shr, err := sc.str.FindShareWithZIdAndDeleted(u.ZitiServiceId, tx)
	if err != nil {
		return err
	}
	u.ShareToken = shr.Token
	env, err := sc.str.GetEnvironment(shr.EnvironmentId, tx)
	if err != nil {
		return err
	}
	u.EnvironmentId = int64(env.Id)
	u.AccountId = int64(*env.AccountId)

	return nil
}
