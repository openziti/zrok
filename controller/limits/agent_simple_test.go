package limits

import (
	"testing"

	"github.com/openziti/zrok/controller/automation"
	"github.com/openziti/zrok/controller/emailUi"
	"github.com/openziti/zrok/controller/metrics"
	"github.com/openziti/zrok/controller/store"
	"github.com/stretchr/testify/assert"
)

func TestNewAgentSimple(t *testing.T) {
	cfg := DefaultConfig()
	ifxCfg := &metrics.InfluxConfig{
		Url:    "http://localhost:8086",
		Token:  "test-token",
		Org:    "test-org",
		Bucket: "test-bucket",
	}
	zCfg := &automation.Config{}
	emailCfg := &emailUi.Config{}

	// use a nil store for basic constructor test
	agent, err := NewAgent(cfg, ifxCfg, zCfg, emailCfg, &store.Store{})

	assert.NoError(t, err)
	assert.NotNil(t, agent)
	assert.Equal(t, cfg, agent.cfg)
	assert.NotNil(t, agent.ifx)
	assert.Equal(t, zCfg, agent.zCfg)
	assert.NotNil(t, agent.queue)
	assert.NotNil(t, agent.warningActions)
	assert.NotNil(t, agent.limitActions)
	assert.NotNil(t, agent.relaxActions)
	assert.NotNil(t, agent.close)
	assert.NotNil(t, agent.join)
}

func TestAgent_TransferBytesExceeded(t *testing.T) {
	agent := &Agent{}

	tests := []struct {
		name     string
		rx       int64
		tx       int64
		bwc      store.BandwidthClass
		expected bool
	}{
		{
			name: "under all limits",
			rx:   50,
			tx:   50,
			bwc: &configBandwidthClass{
				bw: &Bandwidth{
					Rx:    100,
					Tx:    100,
					Total: 200,
				},
			},
			expected: false,
		},
		{
			name: "rx exceeded",
			rx:   150,
			tx:   50,
			bwc: &configBandwidthClass{
				bw: &Bandwidth{
					Rx:    100,
					Tx:    200,
					Total: 300,
				},
			},
			expected: true,
		},
		{
			name: "tx exceeded",
			rx:   50,
			tx:   150,
			bwc: &configBandwidthClass{
				bw: &Bandwidth{
					Rx:    200,
					Tx:    100,
					Total: 300,
				},
			},
			expected: true,
		},
		{
			name: "total exceeded",
			rx:   100,
			tx:   100,
			bwc: &configBandwidthClass{
				bw: &Bandwidth{
					Rx:    200,
					Tx:    200,
					Total: 150,
				},
			},
			expected: true,
		},
		{
			name: "unlimited values",
			rx:   1000000,
			tx:   1000000,
			bwc: &configBandwidthClass{
				bw: &Bandwidth{
					Rx:    store.Unlimited,
					Tx:    store.Unlimited,
					Total: store.Unlimited,
				},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := agent.transferBytesExceeded(tt.rx, tt.tx, tt.bwc)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAgent_Handle(t *testing.T) {
	agent := &Agent{
		queue: make(chan *metrics.Usage, 10),
	}

	usage := &metrics.Usage{
		AccountId:  1,
		ShareToken: "test-token",
	}

	err := agent.Handle(usage)
	assert.NoError(t, err)

	// verify the usage was queued
	select {
	case queued := <-agent.queue:
		assert.Equal(t, usage, queued)
	default:
		t.Fatal("usage was not queued")
	}
}