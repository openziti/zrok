package limits

import (
	"github.com/openziti/zrok/v2/controller/store"
	"time"
)

// DefaultQueryTimeout bounds each of the limits agent's influx queries.
const DefaultQueryTimeout = 30 * time.Second

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
	}
}
