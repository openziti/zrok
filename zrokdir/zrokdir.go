package zrokdir

import (
	"encoding/json"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"os"
	"path/filepath"
)

type Environment struct {
	ZrokToken      string `json:"zrok_token"`
	ZitiIdentityId string `json:"ziti_identity"`
}

func LoadEnvironment() (*Environment, error) {
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

func SaveEnvironment(env *Environment) error {
	logrus.Infof("saving environment")
	data, err := json.MarshalIndent(env, "", "  ")
	if err != nil {
		return errors.Wrap(err, "error marshaling environment")
	}
	ef, err := environmentFile()
	if err != nil {
		return errors.Wrap(err, "error getting environment file")
	}
	if err := os.WriteFile(ef, data, os.FileMode(0600)); err != nil {
		return errors.Wrap(err, "error saving environment file")
	}
	return nil
}

func WriteIdentityConfig(data string) error {
	path, err := IdentityConfigFile()
	if err != nil {
		return err
	}
	if err := os.WriteFile(path, []byte(data), os.FileMode(400)); err != nil {
		return err
	}
	return nil
}

func IdentityConfigFile() (string, error) {
	zrok, err := zrokDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(zrok, "identity.json"), nil
}

func environmentFile() (string, error) {
	zrd, err := zrokDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(zrd, "environment.json"), nil
}

func zrokDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".zrok"), nil
}
