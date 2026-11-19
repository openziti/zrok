package limits

import (
	"time"

	"github.com/openziti/zrok/v2/controller/store"
)

// DefaultQueryTimeout bounds each of the limits agent's influx queries.
const DefaultQueryTimeout = 30 * time.Second

const (
	// DefaultAccountCacheLifetime is how long an account's limitless flag and limit classes are trusted.
	DefaultAccountCacheLifetime = 60 * time.Second
	// DefaultUsageCacheLifetime is how long an account's transfer totals are trusted.
	DefaultUsageCacheLifetime = 60 * time.Second
	// DefaultCacheReportInterval is how often the fast-path counts are logged.
	DefaultCacheReportInterval = 5 * time.Minute
)

type Config struct {
	Environments   int
	Shares         int
	ReservedShares int
	UniqueNames    int
	ShareFrontends int
	Bandwidth      *BandwidthPerPeriod
	Cycle          time.Duration
	Enforcing      bool
	HandoffTimeout time.Duration
	QueryTimeout   time.Duration

	// AccountCacheLifetime is how long the fast path trusts an account's limitless flag, limit classes and
	// recorded crossings. zero is the default; negative turns the cache off.
	AccountCacheLifetime time.Duration
	// UsageCacheLifetime is how long the fast path trusts an account's transfer totals. zero is the default;
	// negative turns the cache off.
	UsageCacheLifetime time.Duration
	// CacheReportInterval is how often the fast-path counts are logged. zero is the default; negative turns
	// the report off.
	CacheReportInterval time.Duration
}

// these resolve the cache settings where they are used, so a zero from the file or from a Config built by
// hand gets the default, and a negative value is kept.

func (cfg *Config) accountCacheLifetime() time.Duration {
	if cfg == nil || cfg.AccountCacheLifetime == 0 {
		return DefaultAccountCacheLifetime
	}
	return cfg.AccountCacheLifetime
}

func (cfg *Config) usageCacheLifetime() time.Duration {
	if cfg == nil || cfg.UsageCacheLifetime == 0 {
		return DefaultUsageCacheLifetime
	}
	return cfg.UsageCacheLifetime
}

func (cfg *Config) cacheReportInterval() time.Duration {
	if cfg == nil || cfg.CacheReportInterval == 0 {
		return DefaultCacheReportInterval
	}
	return cfg.CacheReportInterval
}

type BandwidthPerPeriod struct {
	Period  time.Duration
	Warning *Bandwidth
	Limit   *Bandwidth
}

type Bandwidth struct {
	Rx    int64
	Tx    int64
	Total int64
}

func DefaultBandwidthPerPeriod() *BandwidthPerPeriod {
	return &BandwidthPerPeriod{
		Period: 24 * time.Hour,
		Warning: &Bandwidth{
			Rx:    store.Unlimited,
			Tx:    store.Unlimited,
			Total: store.Unlimited,
		},
		Limit: &Bandwidth{
			Rx:    store.Unlimited,
			Tx:    store.Unlimited,
			Total: store.Unlimited,
		},
	}
}

func DefaultConfig() *Config {
	return &Config{
		Environments:   store.Unlimited,
		Shares:         store.Unlimited,
		ReservedShares: store.Unlimited,
		UniqueNames:    store.Unlimited,
		ShareFrontends: store.Unlimited,
		Bandwidth:      DefaultBandwidthPerPeriod(),
		Enforcing:      false,
		Cycle:          15 * time.Minute,
		QueryTimeout:   DefaultQueryTimeout,

		AccountCacheLifetime: DefaultAccountCacheLifetime,
		UsageCacheLifetime:   DefaultUsageCacheLifetime,
		CacheReportInterval:  DefaultCacheReportInterval,
	}
}
