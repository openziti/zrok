package tunnelBackend

import "github.com/openziti/sdk-golang/ziti/edge"

type Config struct {
	IdentityPath    string
	EndpointAddress string
	ShrToken        string
}

type Backend struct {
	cfg      *Config
	listener edge.Listener
}

func New