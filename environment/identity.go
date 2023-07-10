package environment

import (
	"github.com/pkg/errors"
	"os"
	"path/filepath"
)

func ZitiIdentityFile(name string) (string, error) {
	return identityFile(name)
}

func SaveZitiIdentity(name, data string) error {
	zif, err := ZitiIdentityFile(name)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(zif), os.FileMode(0700)); err != nil {
		return errors.Wrapf(err, "error creating environment path '%v'", filepath.Dir(zif))
	}
	if err := os.WriteFile(zif, []byte(data), os.FileMode(0600)); err != nil {
		return errors.Wrapf(err, "error writing ziti identity file '%v'", zif)
	}
	return nil
}

func DeleteZitiIdentity(name string) error {
	zif, err := ZitiIdentityFile(name)
	if err != nil {
		return errors.Wrapf(err, "error getting ziti identity file path for '%v'", name)
	}
	if err := os.Remove(zif); err != nil {
		return errors.Wrapf(err, "error removing ziti identity file '%v'", zif)
	}
	return nil
}
