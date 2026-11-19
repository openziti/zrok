package metrics

import (
	"context"
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/michaelquigley/df/dl"
	"github.com/openziti/zrok/v2/controller/store"
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

// attribution is a confirmed answer from the store for one ziti service id: either the share it belongs to,
// or the fact that it is not a share.
type attribution struct {
	notAShare     bool
	shareToken    string
	environmentId int64
	accountId     *int64
	backendMode   string
	expires       time.Time
}

// cache remembers confirmed attributions by ziti service id. it is used only from the metrics agent's run
// goroutine, so it takes no lock. only a confirmed store answer is ever stored; a failure is never an answer.
type cache struct {
	str              enrichmentStore
	shareLifetime    time.Duration
	notShareLifetime time.Duration
	size             int
	reportInterval   time.Duration
	now              func() time.Time
	entries          map[string]attribution
	lastReport       time.Time
	shareHits        int64
	notShareHits     int64
	storeLookups     int64
}

func newShareCache(str *store.Store, cfg AgentConfig) *cache {
	c := &cache{
		str:              str,
		shareLifetime:    cfg.ShareCacheLifetime,
		notShareLifetime: cfg.NotShareCacheLifetime,
		size:             cfg.ShareCacheSize,
		reportInterval:   cfg.CacheReportInterval,
		now:              time.Now,
		entries:          make(map[string]attribution),
	}
	c.lastReport = c.now()
	return c
}

func (c *cache) addZrokDetail(ctx context.Context, u *Usage) error {
	if a, found := c.entries[u.ZitiServiceId]; found && c.now().Before(a.expires) {
		if a.notAShare {
			c.notShareHits++
			return errNotAShare
		}
		c.shareHits++
		return a.apply(u)
	}

	c.storeLookups++
	a, err := c.lookup(ctx, u.ZitiServiceId)
	if err != nil {
		return err
	}
	c.remember(u.ZitiServiceId, a)
	if a.notAShare {
		return errNotAShare
	}
	return a.apply(u)
}

// lookup reads the attribution for a service from the store. it returns an attribution only for a confirmed
// answer; any other outcome is an error and must not be remembered.
func (c *cache) lookup(ctx context.Context, zId string) (attribution, error) {
	trx, err := c.str.BeginContext(ctx)
	if err != nil {
		return attribution{}, err
	}
	defer func() { _ = trx.Rollback() }()

	shr, err := c.str.FindShareWithZIdAndDeletedContext(ctx, zId, trx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return attribution{notAShare: true}, nil
		}
		return attribution{}, err
	}
	env, err := c.str.GetEnvironmentContext(ctx, shr.EnvironmentId, trx)
	if err != nil {
		return attribution{}, err
	}
	a := attribution{
		shareToken:    shr.Token,
		environmentId: int64(env.Id),
		backendMode:   shr.BackendMode,
	}
	if env.AccountId != nil {
		accountId := int64(*env.AccountId)
		a.accountId = &accountId
	}
	return a, nil
}

func (a attribution) apply(u *Usage) error {
	u.ShareToken = a.shareToken
	u.EnvironmentId = a.environmentId
	u.BackendMode = a.backendMode
	if a.accountId == nil {
		return errors.Wrapf(errNoAccount, "environment '%d'", a.environmentId)
	}
	u.AccountId = *a.accountId
	return nil
}

func (c *cache) remember(zId string, a attribution) {
	lifetime := c.shareLifetime
	if a.notAShare {
		lifetime = c.notShareLifetime
	}
	if lifetime < 0 {
		return
	}
	now := c.now()
	if _, found := c.entries[zId]; !found && len(c.entries) >= c.size {
		for k, v := range c.entries {
			if !now.Before(v.expires) {
				delete(c.entries, k)
			}
		}
		if len(c.entries) >= c.size {
			clear(c.entries)
		}
	}
	a.expires = now.Add(lifetime)
	c.entries[zId] = a
}

// report logs the lookup counts once the report interval has passed, and only when lookups happened since
// the last report. it is called from the run loop after each event.
func (c *cache) report() {
	if c.reportInterval < 0 {
		return
	}
	now := c.now()
	if now.Sub(c.lastReport) < c.reportInterval {
		return
	}
	c.lastReport = now
	if c.shareHits+c.notShareHits+c.storeLookups == 0 {
		return
	}
	dl.Infof("share cache answered '%d' share and '%d' not-a-share lookups from memory, sent '%d' lookups to the store, holds '%d' entries",
		c.shareHits, c.notShareHits, c.storeLookups, len(c.entries))
	c.shareHits, c.notShareHits, c.storeLookups = 0, 0, 0
}
