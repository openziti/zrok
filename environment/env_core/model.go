package env_core

import "github.com/openziti/zrok/rest_client_zrok"

// Root is the primary interface encapsulating the on-disk environment data.
type Root interface {
	Metadata() *Metadata
	Obliterate() error

	HasConfig() (bool, error)
	Config() *Config
	SetConfig(cfg *Config) error

	Client() (*rest_client_zrok.Zrok, error)
	ApiEndpoint() (string, string)
	DefaultFrontend() (string, string)
	Headless() (bool, string)

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
}

type Environment struct {
	AccountToken string
	ZitiIdentity string
	ApiEndpoint  string
}

type Config struct {
	ApiEndpoint     string
	DefaultFrontend string
	Headless        bool
}

type Metadata struct {
	V        string
	RootPath string
}
