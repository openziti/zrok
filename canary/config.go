package canary

import (
	"github.com/michaelquigley/df"
	"github.com/openziti/zrok/controller/metrics"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const ConfigVersion = 1

type Config struct {
	V      int
	Influx *metrics.InfluxConfig
}

func LoadConfig(path string) (*Config, error) {
	cfg, err := df.NewFromYAML[Config](path)
	if err != nil {
		return nil, err
	}
	if cfg.V != ConfigVersion {
		return nil, errors.Errorf("expecting canary configuration version '%v', got '%v'", ConfigVersion, cfg.V)
	}
	logrus.Info(df.MustInspect(cfg))
	return cfg, nil
}
