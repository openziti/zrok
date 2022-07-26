package zrokdir

import (
	"os"
	"path/filepath"
)

func ReadToken() (string, error) {
	path, err := tokenFile()
	if err != nil {
		return "", err
	}
	token, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(token), nil
}

func WriteToken(token string) error {
	path, err := tokenFile()
	if err != nil {
		return err
	}
	if err := os.WriteFile(path, []byte(token), os.FileMode(400)); err != nil {
		return err
	}
	return nil
}

func ReadIdentityId() (string, error) {
	path, err := IdentityIdFile()
	if err != nil {
		return "", err
	}
	id, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(id), nil
}

func WriteIdentityId(id string) error {
	path, err := IdentityIdFile()
	if err != nil {
		return err
	}
	if err := os.WriteFile(path, []byte(id), os.FileMode(400)); err != nil {
		return err
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

func IdentityIdFile() (string, error) {
	zrok, err := zrokDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(zrok, "identity.id"), nil
}

func IdentityConfigFile() (string, error) {
	zrok, err := zrokDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(zrok, "identity.json"), nil
}

func tokenFile() (string, error) {
	zrok, err := zrokDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(zrok, "token"), nil
}

func zrokDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".zrok"), nil
}
