package controller

import (
	"github.com/michaelquigley/cf"
	"github.com/openziti-test-kitchen/zrok/controller/store"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type Config struct {
	Endpoint *EndpointConfig
	Proxy    *ProxyConfig
	Store    *store.Config
	Ziti     *ZitiConfig
}

type EndpointConfig struct {
	Host string
	Port int
}

type ProxyConfig struct {
	UrlTemplate string
	Identities  []string
}

type ZitiConfig struct {
	ApiEndpoint string
	Username    string
	Password    string
}

func LoadConfig(path string) (*Config, error) {
	cfg := &Config{}
	if err := cf.BindYaml(cfg, path, cf.DefaultOptions()); err != nil {
		return nil, errors.Wrapf(err, "error loading controller config '%v'", path)
	}
	logrus.Info(cf.Dump(cfg, cf.DefaultOptions()))
	return cfg, nil
}
