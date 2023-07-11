package env_v0_3

import (
	"encoding/json"
	"github.com/pkg/errors"
	"os"
	"path/filepath"
)

type Config struct {
	ApiEndpoint string `json:"api_endpoint"`
}

func HasConfig() (bool, error) {
	cf, err := configFile()
	if err != nil {
		return false, errors.Wrap(err, "error getting config file path")
	}
	_, err = os.Stat(cf)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, errors.Wrapf(err, "error stat-ing config file '%v'", cf)
	}
	return true, nil
}

func LoadConfig() (*Config, error) {
	cf, err := configFile()
	if err != nil {
		return nil, errors.Wrap(err, "error getting config file path")
	}
	data, err := os.ReadFile(cf)
	if err != nil {
		return nil, errors.Wrapf(err, "error reading config file '%v'", cf)
	}
	cfg := &Config{}
	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, errors.Wrapf(err, "error unmarshaling config file '%v'", cf)
	}
	return cfg, nil
}

func SaveConfig(cfg *Config) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
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
