package zrokdir

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
)

func migrate() error {
	zrd, err := zrokDirOld()
	if err != nil {
		return err
	}
	_, err = os.Stat(zrd)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	if err := moveFileHelper(configFileOld, configFile, "config"); err != nil {
		return err
	}
	if err := moveFileHelper(environmentFileOld, environmentFile, "environment"); err != nil {
		return err
	}

	if err := moveIdentities(); err != nil {
		return err
	}

	if err := moveFileHelper(metadataFileOld, metadataFile, "metadata"); err != nil {
		return err
	}

	return nil
}

func moveIdentities() error {
	oi, err := listIdentitiesOld()
	if err != nil {
		return fmt.Errorf("unable to list old identities: %v", err)
	}
	ifo := func(name string) func() (string, error) {
		return func() (string, error) {
			return identityFileOld(name)
		}
	}
	ifn := func(name string) func() (string, error) {
		return func() (string, error) {
			return identityFile(name)
		}
	}

	fmt.Println(oi)

	for id := range oi {
		if err := moveFileHelper(ifo(id), ifn(id), fmt.Sprintf("identity/%s", id)); err != nil {
			return fmt.Errorf("unable to move identity directory: %v", err)
		}
	}

	return nil
}

func moveFileHelper(old, new func() (string, error), name string) error {
	of, err := old()
	if err != nil {
		return fmt.Errorf("unable to load old %f file: %v", name, err)
	}
	nf, err := new()
	if err != nil {
		return fmt.Errorf("unable to load new %f file: %v", name, err)
	}
	if err := moveFile(of, nf); err != nil {
		return fmt.Errorf("unable to move %s file: %v", name, err)
	}
	return nil
}

func moveFile(source, dest string) error {
	sourceFile, err := os.Open(source)
	if err != nil {
		return fmt.Errorf("unable to open source file: %v", err)
	}
	defer sourceFile.Close()
	destFile, err := os.Open(dest)
	if err != nil {
		return fmt.Errorf("unable to open destination file: %v", err)
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return fmt.Errorf("error writing to destination file: %v", err)
	}
	return nil
}

func listIdentitiesOld() (map[string]struct{}, error) {
	ids := make(map[string]struct{})

	idd, err := identitiesDirOld()
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

func configFileOld() (string, error) {
	zrd, err := zrokDirOld()
	if err != nil {
		return "", err
	}
	return filepath.Join(zrd, "config.json"), nil
}

func environmentFileOld() (string, error) {
	zrd, err := zrokDirOld()
	if err != nil {
		return "", err
	}
	return filepath.Join(zrd, "environment.json"), nil
}

func identityFileOld(name string) (string, error) {
	idd, err := identitiesDirOld()
	if err != nil {
		return "", err
	}
	return filepath.Join(idd, fmt.Sprintf("%v.json", name)), nil
}

func identitiesDirOld() (string, error) {
	zrd, err := zrokDirOld()
	if err != nil {
		return "", err
	}
	return filepath.Join(zrd, "identities"), nil
}

func metadataFileOld() (string, error) {
	zrd, err := zrokDirOld()
	if err != nil {
		return "", err
	}
	return filepath.Join(zrd, "metadata.json"), nil
}

func zrokDirOld() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".zrok"), nil
}
