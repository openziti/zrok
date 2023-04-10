package publicProxyFrontend

import (
	"github.com/michaelquigley/cf"
	"github.com/pkg/errors"
)

type Config struct {
	Identity  string
	Address   string
	HostMatch string
}

func DefaultConfig() *Config {
	return &Config{
		Identity: "frontend",
		Address:  "0.0.0.0:8080",
	}
}

func (c *Config) Load(path string) error {
	if err := cf.BindYaml(c, path, cf.DefaultOptions()); err != nil {
		return errors.Wrapf(err, "error loading frontend config '%v'", path)
	}
	return nil
}
