package zrokdir

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/adrg/xdg"
	"github.com/pkg/errors"
)

type ZrokXDG struct {
	Env        *Environment
	Cfg        *Config
	identities map[string]struct{}
}

func Load() (*ZrokXDG, error) {
	if err := checkMetadata(); err != nil {
		return nil, err
	}

	zrd := &ZrokXDG{}

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

func (zrd *ZrokXDG) Save() error {
	if err := writeMetadata(); err != nil {
		return errors.Wrap(err, "error saving metadata")
	}
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
	cfgDir, err := configFile()
	if err != nil {
		return err
	}
	if err := os.RemoveAll(filepath.Dir(cfgDir)); err != nil {
		return err
	}

	dataDir, err := environmentFile()
	if err != nil {
		return err
	}
	if err := os.RemoveAll(filepath.Dir(dataDir)); err != nil {
		return err
	}
	return nil
}

func listIdentities() (map[string]struct{}, error) {
	ids := make(map[string]struct{})

	idd, err := identitiesDir()
	if err != nil {
		return nil, errors.Wrap(err, "error getting zrokdir identities path")
	}
	_, err = os.Stat(idd)
	if os.IsNotExist(err) {
		return ids, nil
	}
	if err != nil {
		return nil, errors.Wrapf(err, "error stat-ing zrokdir identities root '%v'", idd)
	}
	des, err := os.ReadDir(idd)
	if err != nil {
		return nil, errors.Wrapf(err, "error listing zrokdir identities from '%v'", idd)
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
	return xdg.ConfigFile("zrok/config.json")
}

func environmentFile() (string, error) {
	return xdg.DataFile("zrok/environment.json")
}

func identityFile(name string) (string, error) {
	return xdg.DataFile(fmt.Sprintf("zrok/identities/%v.json", name))
}

func identitiesDir() (string, error) {
	dataDir, err := environmentFile()
	if err != nil {
		return "", err
	}
	return filepath.Join(filepath.Dir(dataDir), "identities"), nil
}

func metadataFile() (string, error) {
	return xdg.DataFile("zrok/metadata.json")
}
