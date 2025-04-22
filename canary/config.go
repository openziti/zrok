package canary

import (
	"github.com/michaelquigley/cf"
	"github.com/openziti/zrok/controller/metrics"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const ConfigVersion = 1

type Config struct {
	V      int
	Influx *metrics.InfluxConfig
}

func DefaultConfig() *Config {
	return &Config{}
}

func LoadConfig(path string) (*Config, error) {
	cfg := DefaultConfig()
	if err := cf.BindYaml(cfg, path, cf.DefaultOptions()); err != nil {
		return nil, errors.Wrapf(err, "error loading canary configuration '%v'", path)
	}
	if cfg.V != ConfigVersion {
		return nil, errors.Errorf("expecting canary configuration version '%v', got '%v'", ConfigVersion, cfg.V)
	}
	logrus.Info(cf.Dump(cfg, cf.DefaultOptions()))
	return cfg, nil
}
