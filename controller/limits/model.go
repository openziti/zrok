package limits

import "time"

const Unlimited = -1

type Config struct {
	Environments int
	Shares       int
	Bandwidth    *BandwidthConfig
}

type BandwidthConfig struct {
	PerAccount     *BandwidthPerPeriod
	PerEnvironment *BandwidthPerPeriod
	PerShare       *BandwidthPerPeriod
}

type BandwidthPerPeriod struct {
	Period time.Duration
	Rx     int64
	Tx     int64
	Total  int64
}

func DefaultConfig() *Config {
	return &Config{
		Environments: Unlimited,
		Shares:       Unlimited,
		Bandwidth: &BandwidthConfig{
			PerAccount: &BandwidthPerPeriod{
				Period: 365 * (24 * time.Hour),
				Rx:     Unlimited,
				Tx:     Unlimited,
				Total:  Unlimited,
			},
			PerEnvironment: &BandwidthPerPeriod{
				Period: 365 * (24 * time.Hour),
				Rx:     Unlimited,
				Tx:     Unlimited,
				Total:  Unlimited,
			},
			PerShare: &BandwidthPerPeriod{
				Period: 365 * (24 * time.Hour),
				Rx:     Unlimited,
				Tx:     Unlimited,
				Total:  Unlimited,
			},
		},
	}
}
