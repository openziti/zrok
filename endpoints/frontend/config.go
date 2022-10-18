package frontend

import (
	"github.com/michaelquigley/cf"
	"github.com/pkg/errors"
	"time"
)

type Config struct {
	Identity  string
	Metrics   *MetricsConfig
	Address   string
	HostMatch string
}

type MetricsConfig struct {
	Service     string
	SendTimeout time.Duration
}

func DefaultConfig() *Config {
	return &Config{
		Identity: "frontend",
		Metrics: &MetricsConfig{
			Service:     "metrics",
			SendTimeout: 5 * time.Second,
		},
		Address: "0.0.0.0:8080",
	}
}

func (c *Config) Load(path string) error {
	if err := cf.BindYaml(c, path, cf.DefaultOptions()); err != nil {
		return errors.Wrapf(err, "error loading frontend config '%v'", path)
	}
	return nil
}
