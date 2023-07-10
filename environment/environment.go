package environment

import (
	"encoding/json"
	"github.com/pkg/errors"
	"os"
	"path/filepath"
)

type Environment struct {
	Token       string `json:"zrok_token"`
	ZId         string `json:"ziti_identity"`
	ApiEndpoint string `json:"api_endpoint"`
}

func hasEnvironment() (bool, error) {
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

func loadEnvironment() (*Environment, error) {
	ef, err := environmentFile()
	if err != nil {
		return nil, errors.Wrap(err, "error getting environment file")
	}
	data, err := os.ReadFile(ef)
	if err != nil {
		return nil, errors.Wrapf(err, "error reading environment file '%v'", ef)
	}
	env := &Environment{}
	if err := json.Unmarshal(data, env); err != nil {
		return nil, errors.Wrapf(err, "error unmarshaling environment file '%v'", ef)
	}
	return env, nil
}

func saveEnvironment(env *Environment) error {
	data, err := json.MarshalIndent(env, "", "  ")
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

func DeleteEnvironment() error {
	ef, err := environmentFile()
	if err != nil {
		return errors.Wrap(err, "error getting environment file")
	}
	if err := os.Remove(ef); err != nil {
		return errors.Wrap(err, "error removing environment file")
	}

	return nil
}
