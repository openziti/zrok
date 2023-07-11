package env_v0_3

import (
	"fmt"
	"github.com/pkg/errors"
	"os"
	"path/filepath"
	"strings"
)

type Root struct {
	Env        *Environment
	Cfg        *Config
	identities map[string]struct{}
}

func Initialize() (*Root, error) {
	zrd, err := rootDir()
	if err != nil {
		return nil, errors.Wrap(err, "error getting environment path")
	}
	if err := os.MkdirAll(zrd, os.FileMode(0700)); err != nil {
		return nil, errors.Wrapf(err, "error creating environment root path '%v'", zrd)
	}
	if err := DeleteEnvironment(); err != nil {
		return nil, errors.Wrap(err, "error deleting environment")
	}
	idd, err := identitiesDir()
	if err != nil {
		return nil, errors.Wrap(err, "error getting environment identities path")
	}
	if err := os.MkdirAll(idd, os.FileMode(0700)); err != nil {
		return nil, errors.Wrapf(err, "error creating environment identities root path '%v'", idd)
	}
	return Load()
}

func Load() (*Root, error) {
	if err := checkMetadata(); err != nil {
		return nil, err
	}

	zrd := &Root{}

	ids, err := listIdentities()
	if err != nil {
		return nil, err
	}
	zrd.identities = ids

	hasCfg, err := HasConfig()
	if err != nil {
		return nil, err
	}
	if hasCfg {
		cfg, err := LoadConfig()
		if err != nil {
			return nil, err
		}
		zrd.Cfg = cfg
	}

	hasEnv, err := IsEnabled()
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

func (r *Root) Save() error {
	if err := writeMetadata(); err != nil {
		return errors.Wrap(err, "error saving metadata")
	}
	if r.Env != nil {
		if err := saveEnvironment(r.Env); err != nil {
			return errors.Wrap(err, "error saving environment")
		}
	}
	if r.Cfg != nil {
		if err := SaveConfig(r.Cfg); err != nil {
			return errors.Wrap(err, "error saving config")
		}
	}
	return nil
}

func Obliterate() error {
	zrd, err := rootDir()
	if err != nil {
		return err
	}
	if err := os.RemoveAll(zrd); err != nil {
		return err
	}
	return nil
}

func listIdentities() (map[string]struct{}, error) {
	ids := make(map[string]struct{})

	idd, err := identitiesDir()
	if err != nil {
		return nil, errors.Wrap(err, "error getting environment identities path")
	}
	_, err = os.Stat(idd)
	if os.IsNotExist(err) {
		return ids, nil
	}
	if err != nil {
		return nil, errors.Wrapf(err, "error stat-ing environment identities root '%v'", idd)
	}
	des, err := os.ReadDir(idd)
	if err != nil {
		return nil, errors.Wrapf(err, "error listing environment identities from '%v'", idd)
	}
	for _, de := range des {
		if strings.HasSuffix(de.Name(), ".json") && !de.IsDir() {
			name := strings.TrimSuffix(de.Name(), ".json")
			ids[name] = struct{}{}
		}
	}
	return ids, nil
}

func configFile() (string, error) {
	zrd, err := rootDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(zrd, "config.json"), nil
}

func environmentFile() (string, error) {
	zrd, err := rootDir()
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
	zrd, err := rootDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(zrd, "identities"), nil
}

func metadataFile() (string, error) {
	zrd, err := rootDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(zrd, "metadata.json"), nil
}

func rootDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".zrok"), nil
}
