package limits

import (
	"testing"

	"github.com/openziti/zrok/controller/store"
	"github.com/openziti/zrok/sdk/golang/sdk"
	"github.com/stretchr/testify/assert"
)

func TestUserLimits_ToBandwidthArraySimple(t *testing.T) {
	bwWarning := &configBandwidthClass{
		periodInMinutes: 60,
		bw:              &Bandwidth{Rx: 100, Tx: 100, Total: 200},
		limitAction:     store.WarningLimitAction,
	}
	bwLimit := &configBandwidthClass{
		periodInMinutes: 60,
		bw:              &Bandwidth{Rx: 200, Tx: 200, Total: 400},
		limitAction:     store.LimitLimitAction,
	}

	tests := []struct {
		name        string
		hasScoped   bool
		backendMode sdk.BackendMode
		expectedLen int
	}{
		{
			name:        "no scoped limits",
			hasScoped:   false,
			backendMode: sdk.ProxyBackendMode,
			expectedLen: 2,
		},
		{
			name:        "with scoped limit for matching backend",
			hasScoped:   true,
			backendMode: sdk.ProxyBackendMode,
			expectedLen: 3,
		},
		{
			name:        "with scoped limit for different backend",
			hasScoped:   true,
			backendMode: sdk.WebBackendMode,
			expectedLen: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ul := &userLimits{
				bandwidth: []store.BandwidthClass{bwWarning, bwLimit},
				scopes:    make(map[sdk.BackendMode]*store.LimitClass),
			}

			if tt.hasScoped {
				proxyMode := sdk.ProxyBackendMode
				ul.scopes[sdk.ProxyBackendMode] = &store.LimitClass{
					Model:         store.Model{Id: 1},
					PeriodMinutes: 30,
					RxBytes:       50,
					TxBytes:       50,
					TotalBytes:    100,
					LimitAction:   store.LimitLimitAction,
					BackendMode:   &proxyMode,
				}
			}

			result := ul.toBandwidthArray(tt.backendMode)
			assert.Equal(t, tt.expectedLen, len(result))
		})
	}
}

func TestAgent_ClassificationMethods(t *testing.T) {
	a := &Agent{
		cfg: DefaultConfig(),
	}

	t.Run("isResourceCountClass", func(t *testing.T) {
		validRC := &store.LimitClass{
			Model:        store.Model{Id: 1},
			Environments: 10,
		}
		assert.True(t, a.isResourceCountClass(validRC))

		proxyMode := sdk.ProxyBackendMode
		invalidRC := &store.LimitClass{
			Model:       store.Model{Id: 2},
			BackendMode: &proxyMode,
		}
		assert.False(t, a.isResourceCountClass(invalidRC))
	})

	t.Run("isUnscopedBandwidthClass", func(t *testing.T) {
		validBwc := &store.LimitClass{
			Model:          store.Model{Id: 1},
			PeriodMinutes:  60,
			RxBytes:        1024,
			Environments:   store.Unlimited,
			Shares:         store.Unlimited,
			ReservedShares: store.Unlimited,
			UniqueNames:    store.Unlimited,
			ShareFrontends: store.Unlimited,
		}
		assert.True(t, a.isUnscopedBandwidthClass(validBwc))

		proxyMode := sdk.ProxyBackendMode
		invalidBwc := &store.LimitClass{
			Model:         store.Model{Id: 2},
			BackendMode:   &proxyMode,
			PeriodMinutes: 60,
			RxBytes:       1024,
		}
		assert.False(t, a.isUnscopedBandwidthClass(invalidBwc))
	})

	t.Run("isScopedLimitClass", func(t *testing.T) {
		proxyMode := sdk.ProxyBackendMode
		validSLC := &store.LimitClass{
			Model:          store.Model{Id: 1},
			BackendMode:    &proxyMode,
			PeriodMinutes:  60,
			RxBytes:        1024,
			Environments:   store.Unlimited,
			Shares:         store.Unlimited,
			ReservedShares: store.Unlimited,
			UniqueNames:    store.Unlimited,
			ShareFrontends: store.Unlimited,
		}
		assert.True(t, a.isScopedLimitClass(validSLC))

		invalidSLC := &store.LimitClass{
			Model:         store.Model{Id: 2},
			PeriodMinutes: 60,
			RxBytes:       1024,
		}
		assert.False(t, a.isScopedLimitClass(invalidSLC))
	})
}

// helper function to get string pointer
func strPtr(s string) *string {
	return &s
}