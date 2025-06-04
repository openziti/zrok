package env_v0_4

import (
	"fmt"
	"os"
	"path/filepath"
)

var rootDirName = ".zrok"

func SetRootDirName(name string) {
	rootDirName = name
}

func rootDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, rootDirName), nil
}

func metadataFile() (string, error) {
	zrd, err := rootDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(zrd, "metadata.json"), nil
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

func identitiesDir() (string, error) {
	zrd, err := rootDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(zrd, "identities"), nil
}

func identityFile(name string) (string, error) {
	idd, err := identitiesDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(idd, fmt.Sprintf("%v.json", name)), nil
}

func agentSocket() (string, error) {
	zrd, err := rootDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(zrd, "agent.socket"), nil
}

func agentRegistry() (string, error) {
	zrd, err := rootDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(zrd, "agent-registry.json"), nil
}

func agentEnrollment() (string, error) {
	zrd, err := rootDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(zrd, "agent-enrollment.json"), nil
}
