package metrics

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
	"github.com/openziti/zrok/controller/store"
	"github.com/pkg/errors"
)

var (
	errNotAShare = errors.New("service is not a zrok share")
	errNoAccount = errors.New("environment has no account")
)

type enrichmentStore interface {
	BeginContext(context.Context) (*sqlx.Tx, error)
	FindShareWithZIdAndDeletedContext(context.Context, string, *sqlx.Tx) (*store.Share, error)
	GetEnvironmentContext(context.Context, int, *sqlx.Tx) (*store.Environment, error)
}

type cache struct {
	str enrichmentStore
}

func newShareCache(str *store.Store) *cache {
	return &cache{str}
}

func (c *cache) addZrokDetail(ctx context.Context, u *Usage) error {
	trx, err := c.str.BeginContext(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = trx.Rollback() }()

	shr, err := c.str.FindShareWithZIdAndDeletedContext(ctx, u.ZitiServiceId, trx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errNotAShare
		}
		return err
	}
	env, err := c.str.GetEnvironmentContext(ctx, shr.EnvironmentId, trx)
	if err != nil {
		return err
	}
	u.ShareToken = shr.Token
	u.EnvironmentId = int64(env.Id)
	if env.AccountId == nil {
		return errors.Wrapf(errNoAccount, "environment '%d'", env.Id)
	}
	u.AccountId = int64(*env.AccountId)

	return nil
}
