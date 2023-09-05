package publicProxy

import (
	"github.com/michaelquigley/cf"
	"github.com/pkg/errors"
)

type Config struct {
	Identity  string
	Address   string
	HostMatch string
	Oauth     *OauthConfig
}

type OauthConfig struct {
	Port        int
	RedirectUrl string
	HashKeyRaw  string
	Providers   []*OauthProviderSecrets
}

func (oc *OauthConfig) GetProvider(name string) *OauthProviderSecrets {
	for _, provider := range oc.Providers {
		if provider.Name == name {
			return provider
		}
	}
	return nil
}

type OauthProviderSecrets struct {
	Name         string
	ClientId     string
	ClientSecret string
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
