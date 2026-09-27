package metrics

import (
	"time"

	"github.com/michaelquigley/df/dd"
)

type Config struct {
	Influx *InfluxConfig
	Agent  *AgentConfig
}

type AgentConfig struct {
	Source              dd.Dynamic
	RetryAttempts       int
	RetryInitialBackoff time.Duration
	RetryMaxBackoff     time.Duration
	RetryBudget         time.Duration
}

func (c AgentConfig) withDefaults() AgentConfig {
	if c.RetryAttempts <= 0 {
		c.RetryAttempts = 5
	}
	if c.RetryInitialBackoff <= 0 {
		c.RetryInitialBackoff = time.Second
	}
	if c.RetryMaxBackoff <= 0 {
		c.RetryMaxBackoff = 30 * time.Second
	}
	if c.RetryBudget <= 0 {
		c.RetryBudget = 2 * time.Minute
	}
	return c
}

type InfluxConfig struct {
	Url          string
	Bucket       string
	Org          string
	Token        string `dd:"+secret"`
	WriteTimeout time.Duration
}
