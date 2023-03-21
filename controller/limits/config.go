package limits

import "time"

const Unlimited = -1

type Config struct {
	Environments int
	Shares       int
	Bandwidth    *BandwidthConfig
	Cycle        time.Duration
	Enforcing    bool
}

type BandwidthConfig struct {
	PerAccount     *BandwidthPerPeriod
	PerEnvironment *BandwidthPerPeriod
	PerShare       *BandwidthPerPeriod
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
		Period: 15 * (24 * time.Hour),
		Warning: &Bandwidth{
			Rx:    Unlimited,
			Tx:    Unlimited,
			Total: Unlimited,
		},
		Limit: &Bandwidth{
			Rx:    Unlimited,
			Tx:    Unlimited,
			Total: Unlimited,
		},
	}
}

func DefaultConfig() *Config {
	return &Config{
		Environments: Unlimited,
		Shares:       Unlimited,
		Bandwidth: &BandwidthConfig{
			PerAccount:     DefaultBandwidthPerPeriod(),
			PerEnvironment: DefaultBandwidthPerPeriod(),
			PerShare:       DefaultBandwidthPerPeriod(),
		},
		Enforcing: false,
		Cycle:     15 * time.Minute,
	}
}
