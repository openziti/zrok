package zrokdir

import (
	"fmt"
	"github.com/pkg/errors"
	"os"
	"path/filepath"
	"strings"
)

type ZrokDir struct {
	Env        *Environment
	Cfg        *Config
	identities map[string]struct{}
}

func Initialize() (*ZrokDir, error) {
	zrd, err := zrokDir()
	if err != nil {
		return nil, errors.Wrap(err, "error getting zrokdir path")
	}
	if err := os.MkdirAll(zrd, os.FileMode(0700)); err != nil {
		return nil, errors.Wrapf(err, "error creating zrokdir root path '%v'", zrd)
	}
	if err := DeleteEnvironment(); err != nil {
		return nil, errors.Wrap(err, "error deleting environment")
	}
	idd, err := identitiesDir()
	if err != nil {
		return nil, errors.Wrap(err, "error getting zrokdir identities path")
	}
	if err := os.MkdirAll(idd, os.FileMode(0700)); err != nil {
		return nil, errors.Wrapf(err, "error creating zrokdir identities root path '%v'", idd)
	}
	return Load()
}

func Load() (*ZrokDir, error) {
	zrd := &ZrokDir{}

	ids, err := listIdentities()
	if err != nil {
		return nil, err
	}
	zrd.identities = ids

	hasCfg, err := hasConfig()
	if err != nil {
		return nil, err
	}
	if hasCfg {
		cfg, err := loadConfig()
		if err != nil {
			return nil, err
		}
		zrd.Cfg = cfg
	}

	hasEnv, err := hasEnvironment()
	if err != nil {
		return nil, err
	}
	if hasEnv {
		env, err := loadEnvironment()
		if err != nil {
			return nil, err
		}
		zrd.Env = env
	}

	return zrd, nil
}

func (zrd *ZrokDir) Save() error {
	if zrd.Env != nil {
		if err := saveEnvironment(zrd.Env); err != nil {
			return errors.Wrap(err, "error saving environment")
		}
	}
	if zrd.Cfg != nil {
		if err := saveConfig(zrd.Cfg); err != nil {
			return errors.Wrap(err, "error saving config")
		}
	}
	return nil
}

func Obliterate() error {
	zrd, err := zrokDir()
	if err != nil {
		return err
	}
	if err := os.RemoveAll(zrd); err != nil {
		return err
	}
	return nil
}

func listIdentities() (map[string]struct{}, error) {
	idd, err := identitiesDir()
	if err != nil {
		return nil, errors.Wrap(err, "error getting zrokdir identities path")
	}
	des, err := os.ReadDir(idd)
	if err != nil {
		return nil, errors.Wrapf(err, "error listing zrokdir identities from '%v'", idd)
	}
	ids := make(map[string]struct{})
	for _, de := range des {
		if strings.HasSuffix(de.Name(), ".json") && !de.IsDir() {
			name := strings.TrimSuffix(de.Name(), ".json")
			ids[name] = struct{}{}
		}
	}
	return ids, nil
}

func configFile() (string, error) {
	zrd, err := zrokDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(zrd, "config.json"), nil
}

func environmentFile() (string, error) {
	zrd, err := zrokDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(zrd, "environment.json"), nil
}

func identityFile(name string) (string, error) {
	idd, err := identitiesDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(idd, fmt.Sprintf("%v.json", name)), nil
}

func identitiesDir() (string, error) {
	zrd, err := zrokDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(zrd, "identities"), nil
}

func zrokDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".zrok"), nil
}
