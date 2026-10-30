package env_core

import (
	"context"
	"fmt"

	"github.com/openziti/zrok/v2/rest_client_zrok"
)

// ContextClient is the optional bounded client-construction capability.
type ContextClient interface {
	ClientContext(context.Context) (*rest_client_zrok.Zrok, error)
}

// ClientContext requires bounded initialization rather than silently using a legacy client.
func ClientContext(ctx context.Context, root Root) (*rest_client_zrok.Zrok, error) {
	client, ok := root.(ContextClient)
	if !ok {
		return nil, fmt.Errorf("environment root does not support context-aware client initialization")
	}
	return client.ClientContext(ctx)
}

// Root is the primary interface encapsulating the on-disk environment data.
type Root interface {
	Metadata() *Metadata
	Obliterate() error

	HasConfig() (bool, error)
	Config() *Config
	SetConfig(cfg *Config) error

	Client() (*rest_client_zrok.Zrok, error)
	ApiEndpoint() (string, string)
	DefaultNamespace() (string, string)
	Headless() (bool, string)
	SuperNetwork() (bool, string)

	IsEnabled() bool
	Environment() *Environment
	SetEnvironment(env *Environment) error
	DeleteEnvironment() error

	PublicIdentityName() string
	EnvironmentIdentityName() string

	ZitiIdentityNamed(name string) (string, error)
	SaveZitiIdentityNamed(name, data string) error
	DeleteZitiIdentityNamed(name string) error

	AgentSocket() (string, error)
	AgentRegistry() (string, error)
	AgentEnrollment() (string, error)
}

type Environment struct {
	AccountToken string
	ZitiIdentity string
	ApiEndpoint  string
}

type Config struct {
	ApiEndpoint      string
	DefaultNamespace string
	Headless         bool
	SuperNetwork     bool
}

type Metadata struct {
	V        string
	RootPath string
}
