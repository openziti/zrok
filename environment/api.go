package environment

import (
	"github.com/openziti/zrok/environment/env_core"
	"github.com/openziti/zrok/rest_client_zrok"
	"github.com/pkg/errors"
)

type Root interface {
	Metadata() *env_core.Metadata
	HasConfig() (bool, error)
	Config() *env_core.Config
	SetConfig(cfg *env_core.Config) error
	Client() (*rest_client_zrok.Zrok, error)
	ApiEndpoint() (string, string)
	Environment() *env_core.Environment
	SetEnvironment(env *env_core.Environment) error
	DeleteEnvironment() error
	IsEnabled() bool
	ZitiIdentityFile(name string) (string, error)
	SaveZitiIdentity(name, data string) error
	DeleteZitiIdentity(name string) error
	Obliterate() error
}

func LoadRoot() (Root, error) {
	return env_v0_3.Load()
}

func ListRoots() ([]*env_core.Metadata, error) {
	return nil, nil
}

func LoadRootVersion(m *env_core.Metadata) (Root, error) {
	if m == nil {
		return nil, errors.Errorf("specify metadata version")
	}
	switch m.V {
	case env_v0_3.V:
		return env_v0_3.Load()

	default:
		return nil, errors.Errorf("unknown metadata version '%v'", m.V)
	}
}

func NeedsUpdate(r Root) bool {
	return r.Metadata().V != env_v0_3.V
}

func UpdateRoot(r Root) (Root, error) {
	return nil, nil
}
