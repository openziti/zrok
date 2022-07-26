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

func WriteIdentity(data string) error {
	path, err := IdentityFile()
	if err != nil {
		return err
	}
	if err := os.WriteFile(path, []byte(data), os.FileMode(400)); err != nil {
		return err
	}
	return nil
}

func IdentityFile() (string, error) {
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
