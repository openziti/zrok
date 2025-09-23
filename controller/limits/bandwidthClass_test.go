package limits

import (
	"fmt"
	"testing"
	"time"

	"github.com/openziti/zrok/controller/store"
	"github.com/openziti/zrok/sdk/golang/sdk"
	"github.com/stretchr/testify/assert"
)

func TestNewConfigBandwidthClasses(t *testing.T) {
	cfg := &BandwidthPerPeriod{
		Period: 24 * time.Hour,
		Warning: &Bandwidth{
			Rx:    1024 * 1024 * 100,
			Tx:    1024 * 1024 * 100,
			Total: 1024 * 1024 * 200,
		},
		Limit: &Bandwidth{
			Rx:    1024 * 1024 * 200,
			Tx:    1024 * 1024 * 200,
			Total: 1024 * 1024 * 400,
		},
	}

	classes := newConfigBandwidthClasses(cfg)

	assert.Len(t, classes, 2)

	// test warning class
	warning := classes[0]
	assert.True(t, warning.IsGlobal())
	assert.False(t, warning.IsScoped())
	assert.Equal(t, -1, warning.GetLimitClassId())
	assert.Equal(t, sdk.BackendMode(""), warning.GetBackendMode())
	assert.Equal(t, int(cfg.Period.Minutes()), warning.GetPeriodMinutes())
	assert.Equal(t, cfg.Warning.Rx, warning.GetRxBytes())
	assert.Equal(t, cfg.Warning.Tx, warning.GetTxBytes())
	assert.Equal(t, cfg.Warning.Total, warning.GetTotalBytes())
	assert.Equal(t, store.WarningLimitAction, warning.GetLimitAction())

	// test limit class
	limit := classes[1]
	assert.True(t, limit.IsGlobal())
	assert.False(t, limit.IsScoped())
	assert.Equal(t, -1, limit.GetLimitClassId())
	assert.Equal(t, sdk.BackendMode(""), limit.GetBackendMode())
	assert.Equal(t, int(cfg.Period.Minutes()), limit.GetPeriodMinutes())
	assert.Equal(t, cfg.Limit.Rx, limit.GetRxBytes())
	assert.Equal(t, cfg.Limit.Tx, limit.GetTxBytes())
	assert.Equal(t, cfg.Limit.Total, limit.GetTotalBytes())
	assert.Equal(t, store.LimitLimitAction, limit.GetLimitAction())
}

func TestConfigBandwidthClass_String(t *testing.T) {
	tests := []struct {
		name     string
		bc       *configBandwidthClass
		expected string
	}{
		{
			name: "unlimited bandwidth",
			bc: &configBandwidthClass{
				periodInMinutes: 1440,
				bw: &Bandwidth{
					Rx:    store.Unlimited,
					Tx:    store.Unlimited,
					Total: store.Unlimited,
				},
				limitAction: store.WarningLimitAction,
			},
			expected: "ConfigClass<periodMinutes: 1440, limitAction: warning>",
		},
		{
			name: "rx only limit",
			bc: &configBandwidthClass{
				periodInMinutes: 60,
				bw: &Bandwidth{
					Rx:    1024 * 1024 * 100,
					Tx:    store.Unlimited,
					Total: store.Unlimited,
				},
				limitAction: store.LimitLimitAction,
			},
			expected: "ConfigClass<periodMinutes: 60, rxBytes: 104.9 MB, limitAction: limit>",
		},
		{
			name: "tx only limit",
			bc: &configBandwidthClass{
				periodInMinutes: 60,
				bw: &Bandwidth{
					Rx:    store.Unlimited,
					Tx:    1024 * 1024 * 50,
					Total: store.Unlimited,
				},
				limitAction: store.WarningLimitAction,
			},
			expected: "ConfigClass<periodMinutes: 60, txBytes: 52.4 MB, limitAction: warning>",
		},
		{
			name: "total only limit",
			bc: &configBandwidthClass{
				periodInMinutes: 120,
				bw: &Bandwidth{
					Rx:    store.Unlimited,
					Tx:    store.Unlimited,
					Total: 1024 * 1024 * 200,
				},
				limitAction: store.LimitLimitAction,
			},
			expected: "ConfigClass<periodMinutes: 120, totalBytes: 209.7 MB, limitAction: limit>",
		},
		{
			name: "all limits set",
			bc: &configBandwidthClass{
				periodInMinutes: 30,
				bw: &Bandwidth{
					Rx:    1024 * 1024 * 10,
					Tx:    1024 * 1024 * 20,
					Total: 1024 * 1024 * 30,
				},
				limitAction: store.WarningLimitAction,
			},
			expected: "ConfigClass<periodMinutes: 30, rxBytes: 10.5 MB, txBytes: 21.0 MB, totalBytes: 31.5 MB, limitAction: warning>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.bc.String()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestConfigBandwidthClass_Methods(t *testing.T) {
	bc := &configBandwidthClass{
		periodInMinutes: 60,
		bw: &Bandwidth{
			Rx:    100,
			Tx:    200,
			Total: 300,
		},
		limitAction: store.LimitLimitAction,
	}

	// test all getter methods
	assert.True(t, bc.IsGlobal())
	assert.False(t, bc.IsScoped())
	assert.Equal(t, -1, bc.GetLimitClassId())
	assert.Equal(t, sdk.BackendMode(""), bc.GetBackendMode())
	assert.Equal(t, 60, bc.GetPeriodMinutes())
	assert.Equal(t, int64(100), bc.GetRxBytes())
	assert.Equal(t, int64(200), bc.GetTxBytes())
	assert.Equal(t, int64(300), bc.GetTotalBytes())
	assert.Equal(t, store.LimitLimitAction, bc.GetLimitAction())
}

func BenchmarkConfigBandwidthClass_String(b *testing.B) {
	bc := &configBandwidthClass{
		periodInMinutes: 1440,
		bw: &Bandwidth{
			Rx:    1024 * 1024 * 100,
			Tx:    1024 * 1024 * 200,
			Total: 1024 * 1024 * 300,
		},
		limitAction: store.WarningLimitAction,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = bc.String()
	}
}

func TestNewConfigBandwidthClasses_DifferentPeriods(t *testing.T) {
	testCases := []struct {
		name           string
		period         time.Duration
		expectedMinutes int
	}{
		{"1 hour", 1 * time.Hour, 60},
		{"6 hours", 6 * time.Hour, 360},
		{"12 hours", 12 * time.Hour, 720},
		{"24 hours", 24 * time.Hour, 1440},
		{"1 week", 7 * 24 * time.Hour, 10080},
		{"30 minutes", 30 * time.Minute, 30},
		{"15 minutes", 15 * time.Minute, 15},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := &BandwidthPerPeriod{
				Period:  tc.period,
				Warning: &Bandwidth{Rx: 100, Tx: 100, Total: 200},
				Limit:   &Bandwidth{Rx: 200, Tx: 200, Total: 400},
			}

			classes := newConfigBandwidthClasses(cfg)

			assert.Equal(t, tc.expectedMinutes, classes[0].GetPeriodMinutes(),
				fmt.Sprintf("Warning class period minutes mismatch for %s", tc.name))
			assert.Equal(t, tc.expectedMinutes, classes[1].GetPeriodMinutes(),
				fmt.Sprintf("Limit class period minutes mismatch for %s", tc.name))
		})
	}
}