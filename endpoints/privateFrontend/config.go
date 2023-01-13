package privateFrontend

import "github.com/openziti/zrok/endpoints"

type Config struct {
	IdentityName string
	ShrToken     string
	Address      string
	RequestsChan chan *endpoints.Request
}

func DefaultConfig(identityName string) *Config {
	return &Config{
		IdentityName: identityName,
		Address:      "0.0.0.0:8080",
	}
}
