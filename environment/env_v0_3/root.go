package env_v0_3

import (
	"encoding/json"
	"github.com/openziti/zrok/environment/env_core"
	"github.com/pkg/errors"
	"os"
	"path/filepath"
)

const V = "v0.3"

type Root struct {
	meta *env_core.Metadata
	cfg  *env_core.Config
	env  *env_core.Environment
}

func Assert() (bool, error) {
	exists, err := rootExists()
	if err != nil {
		return true, err
	}
	if exists {
		meta, err := loadMetadata()
		if err != nil {
			return true, err
		}
		return meta.V == V, nil
	}
	return false, nil
}

func Load() (*Root, error) {
	r := &Root{}
	exists, err := rootExists()
	if err != nil {
		return nil, err
	}
	if exists {
		if meta, err := loadMetadata(); err == nil {
			r.meta = meta
		} else {
			return nil, err
		}

		if cfg, err := loadConfig(); err == nil {
			r.cfg = cfg
		}

		if env, err := loadEnvironment(); err == nil {
			r.env = env
		}

	} else {
		root, err := rootDir()
		if err != nil {
			return nil, err
		}
		r.meta = &env_core.Metadata{
			V:        V,
			RootPath: root,
		}
	}
	return r, nil
}

func rootExists() (bool, error) {
	mf, err := metadataFile()
	if err != nil {
		return false, err
	}
	_, err = os.Stat(mf)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func assertMetadata() error {
	exists, err := rootExists()
	if err != nil {
		return err
	}
	if !exists {
		if err := writeMetadata(); err != nil {
			return err
		}
	}
	return nil
}

func loadMetadata() (*env_core.Metadata, error) {
	mf, err := metadataFile()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(mf)
	if err != nil {
		return nil, err
	}
	m := &metadata{}
	if err := json.Unmarshal(data, m); err != nil {
		return nil, errors.Wrapf(err, "error unmarshaling metadata file '%v'", mf)
	}
	if m.V != V {
		return nil, errors.Errorf("got metadata version '%v', expected '%v'", m.V, V)
	}
	rf, err := rootDir()
	if err != nil {
		return nil, err
	}
	out := &env_core.Metadata{
		V:        m.V,
		RootPath: rf,
	}
	return out, nil
}

func writeMetadata() error {
	mf, err := metadataFile()
	if err != nil {
		return err
	}
	data, err := json.Marshal(&metadata{V: V})
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(mf), os.FileMode(0700)); err != nil {
		return err
	}
	if err := os.WriteFile(mf, data, os.FileMode(0600)); err != nil {
		return err
	}
	return nil
}

func loadConfig() (*env_core.Config, error) {
	cf, err := configFile()
	if err != nil {
		return nil, errors.Wrap(err, "error getting config file path")
	}
	data, err := os.ReadFile(cf)
	if err != nil {
		return nil, errors.Wrapf(err, "error reading config file '%v'", cf)
	}
	cfg := &config{}
	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, errors.Wrapf(err, "error unmarshaling config file '%v'", cf)
	}
	out := &env_core.Config{
		ApiEndpoint: cfg.ApiEndpoint,
	}
	return out, nil
}

func saveConfig(cfg *env_core.Config) error {
	in := &config{ApiEndpoint: cfg.ApiEndpoint}
	data, err := json.MarshalIndent(in, "", "  ")
	if err != nil {
		return errors.Wrap(err, "error marshaling config")
	}
	cf, err := configFile()
	if err != nil {
		return errors.Wrap(err, "error getting config file path")
	}
	if err := os.MkdirAll(filepath.Dir(cf), os.FileMode(0700)); err != nil {
		return errors.Wrapf(err, "error creating environment path '%v'", filepath.Dir(cf))
	}
	if err := os.WriteFile(cf, data, os.FileMode(0600)); err != nil {
		return errors.Wrap(err, "error saving config file")
	}
	return nil
}

func isEnabled() (bool, error) {
	ef, err := environmentFile()
	if err != nil {
		return false, errors.Wrap(err, "error getting environment file path")
	}
	_, err = os.Stat(ef)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, errors.Wrapf(err, "error stat-ing environment file '%v'", ef)
	}
	return true, nil
}

func loadEnvironment() (*env_core.Environment, error) {
	ef, err := environmentFile()
	if err != nil {
		return nil, errors.Wrap(err, "error getting environment file")
	}
	data, err := os.ReadFile(ef)
	if err != nil {
		return nil, errors.Wrapf(err, "error reading environment file '%v'", ef)
	}
	env := &environment{}
	if err := json.Unmarshal(data, env); err != nil {
		return nil, errors.Wrapf(err, "error unmarshaling environment file '%v'", ef)
	}
	out := &env_core.Environment{
		AccountToken: env.AccountToken,
		ZitiIdentity: env.ZId,
		ApiEndpoint:  env.ApiEndpoint,
	}
	return out, nil
}

func saveEnvironment(env *env_core.Environment) error {
	in := &environment{
		AccountToken: env.AccountToken,
		ZId:          env.ZitiIdentity,
		ApiEndpoint:  env.ApiEndpoint,
	}
	data, err := json.MarshalIndent(in, "", "  ")
	if err != nil {
		return errors.Wrap(err, "error marshaling environment")
	}
	ef, err := environmentFile()
	if err != nil {
		return errors.Wrap(err, "error getting environment file")
	}
	if err := os.MkdirAll(filepath.Dir(ef), os.FileMode(0700)); err != nil {
		return errors.Wrapf(err, "error creating environment path '%v'", filepath.Dir(ef))
	}
	if err := os.WriteFile(ef, data, os.FileMode(0600)); err != nil {
		return errors.Wrap(err, "error saving environment file")
	}
	return nil
}

func deleteEnvironment() error {
	ef, err := environmentFile()
	if err != nil {
		return errors.Wrap(err, "error getting environment file")
	}
	if err := os.Remove(ef); err != nil {
		return errors.Wrap(err, "error removing environment file")
	}

	return nil
}

type metadata struct {
	V string `json:"v"`
}

type config struct {
	ApiEndpoint string `json:"api_endpoint"`
}

type environment struct {
	AccountToken string `json:"zrok_token"`
	ZId          string `json:"ziti_identity"`
	ApiEndpoint  string `json:"api_endpoint"`
}
