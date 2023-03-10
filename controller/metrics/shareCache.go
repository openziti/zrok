package metrics

import (
	"github.com/openziti/zrok/controller/store"
	"github.com/pkg/errors"
)

type shareCache struct {
	str *store.Store
}

func newShareCache(cfg *store.Config) (*shareCache, error) {
	str, err := store.Open(cfg)
	if err != nil {
		return nil, errors.Wrap(err, "error opening store")
	}
	return &shareCache{str}, nil
}

func (sc *shareCache) getToken(svcZId string) (string, error) {
	tx, err := sc.str.Begin()
	if err != nil {
		return "", err
	}
	defer func() { _ = tx.Rollback() }()
	shr, err := sc.str.FindShareWithZIdAndDeleted(svcZId, tx)
	if err != nil {
		return "", err
	}
	return shr.Token, nil
}
