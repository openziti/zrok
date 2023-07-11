package environment

import (
	"github.com/openziti/zrok/environment/env_v0_3"
)

type Root interface {
}

func Load() (Root, error) {
	return nil, nil
}

func IsEnabled() (bool, error) {
	return env_v0_3.IsEnabled()
}

func DeleteEnvironment() error {
	return env_v0_3.DeleteEnvironment()
}

func HasConfig() (bool, error) {
	return env_v0_3.HasConfig()
}

func ZitiIdentityFile(name string) (string, error) {
	return env_v0_3.ZitiIdentityFile(name)
}

func SaveZitiIdentity(name, data string) error {
	return env_v0_3.SaveZitiIdentity(name, data)
}

func DeleteZitiIdentity(name string) error {
	return env_v0_3.DeleteZitiIdentity(name)
}
