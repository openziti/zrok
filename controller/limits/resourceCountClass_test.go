package limits

import (
	"fmt"
	"testing"

	"github.com/openziti/zrok/v2/controller/store"
	"github.com/stretchr/testify/assert"
)

func TestNewConfigResourceCountClass(t *testing.T) {
	cfg := &Config{
		Environments:   10,
		Shares:         50,
		ReservedShares: 5,
		UniqueNames:    2,
		ShareFrontends: 100,
	}

	rcc := newConfigResourceCountClass(cfg)

	assert.NotNil(t, rcc)
	assert.True(t, rcc.IsGlobal())
	assert.Equal(t, -1, rcc.GetLimitClassId())
	assert.Equal(t, 10, rcc.GetEnvironments())
	assert.Equal(t, 50, rcc.GetShares())
	assert.Equal(t, 5, rcc.GetReservedShares())
	assert.Equal(t, 2, rcc.GetUniqueNames())
	assert.Equal(t, 100, rcc.GetShareFrontends())
}

func TestConfigResourceCountClass_String(t *testing.T) {
	tests := []struct {
		name     string
		cfg      *Config
		expected string
	}{
		{
			name: "all values set",
			cfg: &Config{
				Environments:   10,
				Shares:         50,
				ReservedShares: 5,
				UniqueNames:    2,
				ShareFrontends: 100,
			},
			expected: "Config<environments: 10, shares: 50, reservedShares: 5, uniqueNames: 2, share_frontends: 100>",
		},
		{
			name: "all unlimited",
			cfg: &Config{
				Environments:   store.Unlimited,
				Shares:         store.Unlimited,
				ReservedShares: store.Unlimited,
				UniqueNames:    store.Unlimited,
				ShareFrontends: store.Unlimited,
			},
			expected: fmt.Sprintf("Config<environments: %d, shares: %d, reservedShares: %d, uniqueNames: %d, share_frontends: %d>",
				store.Unlimited, store.Unlimited, store.Unlimited, store.Unlimited, store.Unlimited),
		},
		{
			name: "mixed values",
			cfg: &Config{
				Environments:   1,
				Shares:         store.Unlimited,
				ReservedShares: 0,
				UniqueNames:    5,
				ShareFrontends: store.Unlimited,
			},
			expected: fmt.Sprintf("Config<environments: 1, shares: %d, reservedShares: 0, uniqueNames: 5, share_frontends: %d>",
				store.Unlimited, store.Unlimited),
		},
		{
			name: "zero values",
			cfg: &Config{
				Environments:   0,
				Shares:         0,
				ReservedShares: 0,
				UniqueNames:    0,
				ShareFrontends: 0,
			},
			expected: "Config<environments: 0, shares: 0, reservedShares: 0, uniqueNames: 0, share_frontends: 0>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rcc := newConfigResourceCountClass(tt.cfg)
			result := rcc.String()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestConfigResourceCountClass_IsGlobal(t *testing.T) {
	cfg := DefaultConfig()
	rcc := newConfigResourceCountClass(cfg)

	// configResourceCountClass should always be global
	assert.True(t, rcc.IsGlobal())
}

func TestConfigResourceCountClass_GetLimitClassId(t *testing.T) {
	cfg := DefaultConfig()
	rcc := newConfigResourceCountClass(cfg)

	// configResourceCountClass should always return -1 for limit class id
	assert.Equal(t, -1, rcc.GetLimitClassId())
}

func TestConfigResourceCountClass_GetterMethods(t *testing.T) {
	testCases := []struct {
		name string
		cfg  *Config
	}{
		{
			name: "positive values",
			cfg: &Config{
				Environments:   100,
				Shares:         200,
				ReservedShares: 50,
				UniqueNames:    10,
				ShareFrontends: 500,
			},
		},
		{
			name: "zero values",
			cfg: &Config{
				Environments:   0,
				Shares:         0,
				ReservedShares: 0,
				UniqueNames:    0,
				ShareFrontends: 0,
			},
		},
		{
			name: "unlimited values",
			cfg: &Config{
				Environments:   store.Unlimited,
				Shares:         store.Unlimited,
				ReservedShares: store.Unlimited,
				UniqueNames:    store.Unlimited,
				ShareFrontends: store.Unlimited,
			},
		},
		{
			name: "large values",
			cfg: &Config{
				Environments:   999999,
				Shares:         888888,
				ReservedShares: 777777,
				UniqueNames:    666666,
				ShareFrontends: 555555,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rcc := newConfigResourceCountClass(tc.cfg)

			assert.Equal(t, tc.cfg.Environments, rcc.GetEnvironments())
			assert.Equal(t, tc.cfg.Shares, rcc.GetShares())
			assert.Equal(t, tc.cfg.ReservedShares, rcc.GetReservedShares())
			assert.Equal(t, tc.cfg.UniqueNames, rcc.GetUniqueNames())
			assert.Equal(t, tc.cfg.ShareFrontends, rcc.GetShareFrontends())
		})
	}
}

func BenchmarkConfigResourceCountClass_String(b *testing.B) {
	cfg := &Config{
		Environments:   10,
		Shares:         50,
		ReservedShares: 5,
		UniqueNames:    2,
		ShareFrontends: 100,
	}
	rcc := newConfigResourceCountClass(cfg)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = rcc.String()
	}
}

func TestConfigResourceCountClass_InterfaceImplementation(t *testing.T) {
	cfg := &Config{
		Environments:   5,
		Shares:         10,
		ReservedShares: 2,
		UniqueNames:    1,
		ShareFrontends: 20,
	}

	rcc := newConfigResourceCountClass(cfg)

	// verify it implements the ResourceCountClass interface
	var _ store.ResourceCountClass = rcc

	// verify it also implements the BaseLimitClass interface
	var _ store.BaseLimitClass = rcc

	// test through interface
	var iface store.ResourceCountClass = rcc
	assert.True(t, iface.IsGlobal())
	assert.Equal(t, -1, iface.GetLimitClassId())
	assert.Equal(t, 5, iface.GetEnvironments())
	assert.Equal(t, 10, iface.GetShares())
	assert.Equal(t, 2, iface.GetReservedShares())
	assert.Equal(t, 1, iface.GetUniqueNames())
	assert.Equal(t, 20, iface.GetShareFrontends())
	assert.NotEmpty(t, iface.String())
}