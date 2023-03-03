package metrics

import (
	"github.com/michaelquigley/cf"
	"github.com/pkg/errors"
)

type Config struct {
	Source interface{}
}

func LoadConfig(path string) (*Config, error) {
	cfg := &Config{}
	if err := cf.BindYaml(cfg, path, GetCfOptions()); err != nil {
		return nil, errors.Wrapf(err, "error loading config from '%v'", path)
	}
	return cfg, nil
}
