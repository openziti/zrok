package limits

import (
	"testing"
	"time"

	"github.com/openziti/zrok/v2/controller/store"
	"github.com/stretchr/testify/assert"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	assert.NotNil(t, cfg)
	assert.Equal(t, store.Unlimited, cfg.Environments)
	assert.Equal(t, store.Unlimited, cfg.Shares)
	assert.Equal(t, store.Unlimited, cfg.ReservedShares)
	assert.Equal(t, store.Unlimited, cfg.UniqueNames)
	assert.Equal(t, store.Unlimited, cfg.ShareFrontends)
	assert.False(t, cfg.Enforcing)
	assert.Equal(t, 15*time.Minute, cfg.Cycle)
	assert.NotNil(t, cfg.Bandwidth)
}

func TestDefaultBandwidthPerPeriod(t *testing.T) {
	bw := DefaultBandwidthPerPeriod()

	assert.NotNil(t, bw)
	assert.Equal(t, 24*time.Hour, bw.Period)

	assert.NotNil(t, bw.Warning)
	assert.Equal(t, int64(store.Unlimited), bw.Warning.Rx)
	assert.Equal(t, int64(store.Unlimited), bw.Warning.Tx)
	assert.Equal(t, int64(store.Unlimited), bw.Warning.Total)

	assert.NotNil(t, bw.Limit)
	assert.Equal(t, int64(store.Unlimited), bw.Limit.Rx)
	assert.Equal(t, int64(store.Unlimited), bw.Limit.Tx)
	assert.Equal(t, int64(store.Unlimited), bw.Limit.Total)
}

func TestConfigWithCustomValues(t *testing.T) {
	cfg := &Config{
		Environments:   10,
		Shares:         50,
		ReservedShares: 5,
		UniqueNames:    2,
		ShareFrontends: 100,
		Bandwidth: &BandwidthPerPeriod{
			Period: 1 * time.Hour,
			Warning: &Bandwidth{
				Rx:    1024 * 1024 * 100, // 100MB
				Tx:    1024 * 1024 * 100, // 100MB
				Total: 1024 * 1024 * 200, // 200MB
			},
			Limit: &Bandwidth{
				Rx:    1024 * 1024 * 200, // 200MB
				Tx:    1024 * 1024 * 200, // 200MB
				Total: 1024 * 1024 * 400, // 400MB
			},
		},
		Cycle:     5 * time.Minute,
		Enforcing: true,
	}

	assert.Equal(t, 10, cfg.Environments)
	assert.Equal(t, 50, cfg.Shares)
	assert.Equal(t, 5, cfg.ReservedShares)
	assert.Equal(t, 2, cfg.UniqueNames)
	assert.Equal(t, 100, cfg.ShareFrontends)
	assert.True(t, cfg.Enforcing)
	assert.Equal(t, 5*time.Minute, cfg.Cycle)
	assert.Equal(t, 1*time.Hour, cfg.Bandwidth.Period)
}