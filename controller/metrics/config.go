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

	// ShareCacheLifetime is how long a confirmed share attribution is trusted; negative disables it.
	ShareCacheLifetime time.Duration
	// NotShareCacheLifetime is how long a confirmed "not a share" answer is trusted; negative disables it.
	NotShareCacheLifetime time.Duration
	// ShareCacheSize is the most attribution entries held, of both kinds together.
	ShareCacheSize int
	// CacheReportInterval is how often the attribution lookup counts are logged; negative disables the report.
	CacheReportInterval time.Duration
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
	// zero is the default; a negative lifetime or interval is kept, and turns that cache or report off.
	if c.ShareCacheLifetime == 0 {
		c.ShareCacheLifetime = 10 * time.Minute
	}
	if c.NotShareCacheLifetime == 0 {
		c.NotShareCacheLifetime = time.Minute
	}
	if c.ShareCacheSize <= 0 {
		c.ShareCacheSize = 100000
	}
	if c.CacheReportInterval == 0 {
		c.CacheReportInterval = 5 * time.Minute
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
