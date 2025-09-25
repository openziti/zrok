package canary

import (
	"github.com/michaelquigley/df/dd"
	"github.com/michaelquigley/df/dl"
	"github.com/openziti/zrok/controller/metrics"
	"github.com/pkg/errors"
)

const ConfigVersion = 1

type Config struct {
	V      int
	Influx *metrics.InfluxConfig
}

func LoadConfig(path string) (*Config, error) {
	cfg, err := dd.NewFromYAML[Config](path)
	if err != nil {
		return nil, err
	}
	if cfg.V != ConfigVersion {
		return nil, errors.Errorf("expecting canary configuration version '%v', got '%v'", ConfigVersion, cfg.V)
	}
	dl.Info(dd.MustInspect(cfg))
	return cfg, nil
}
