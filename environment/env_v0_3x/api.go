package env_v0_3x

import (
	"github.com/openziti/zrok/environment/env_core"
	"github.com/openziti/zrok/rest_client_zrok"
)

func (r *Root) Metadata() *env_core.Metadata {
	return r.meta
}

func (r *Root) HasConfig() (bool, error) {
	return r.cfg != nil, nil
}

func (r *Root) Config() *env_core.Config {
	return r.cfg
}

func (r *Root) SetConfig(cfg *env_core.Config) error {
	if err := saveConfig(cfg); err != nil {
		return err
	}
	r.cfg = cfg
	return nil
}

func (r *Root) Client() (*rest_client_zrok.Zrok, error) {
	return nil, nil
}

func (r *Root) ApiEndpoint() (string, string) {
	if r.env != nil {
		return r.env.ApiEndpoint, "env"
	}
	return "", ""
}

func (r *Root) Environment() *env_core.Environment {
	return r.env
}

func (r *Root) DeleteEnvironment() error {
	return nil
}

func (r *Root) IsEnabled() (bool, error) {
	return r.env != nil, nil
}

func (r *Root) ZitiIdentityFile(name string) (string, error) {
	return "", nil
}

func (r *Root) SaveZitiIdentity(name, data string) error {
	return nil
}

func (r *Root) DeleteZitiIdentity(name string) error {
	return nil
}

func (r *Root) Obliterate() error {
	return nil
}
