package environment

import (
	"github.com/openziti/zrok/environment/env_core"
	"github.com/openziti/zrok/rest_client_zrok"
)

type Root interface {
	Metadata() *env_core.Metadata
	IsLatest() bool
	HasConfig() (bool, error)
	Config() *env_core.Config
	SetConfig(cfg *env_core.Config) error
	Client() (*rest_client_zrok.Zrok, error)
	ApiEndpoint() (string, string)
	Environment() *env_core.Environment
	DeleteEnvironment() error
	IsEnabled() (bool, error)
	ZitiIdentityFile(name string) (string, error)
	SaveZitiIdentity(name, data string) error
	DeleteZitiIdentity(name string) error
	Obliterate() error
}

func ListRoots() ([]*env_core.Metadata, error) {
	return nil, nil
}

func LoadRoot() (Root, error) {
	return nil, nil
}

func LoadRootVersion(m *env_core.Metadata) (Root, error) {
	return nil, nil
}

func UpdateRoot(r Root) (Root, error) {
	return nil, nil
}
