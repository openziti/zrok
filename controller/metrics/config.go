package metrics

import (
	"github.com/michaelquigley/cf"
	"github.com/openziti/zrok/controller/store"
	"github.com/pkg/errors"
)

type Config struct {
	Source interface{}
	Influx *InfluxConfig
	Store  *store.Config
}

type InfluxConfig struct {
	Url    string
	Bucket string
	Org    string
	Token  string `cf:"+secret"`
}

func LoadConfig(path string) (*Config, error) {
	cfg := &Config{}
	if err := cf.BindYaml(cfg, path, GetCfOptions()); err != nil {
		return nil, errors.Wrapf(err, "error loading config from '%v'", path)
	}
	return cfg, nil
}
