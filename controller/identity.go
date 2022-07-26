package controller

import (
	"crypto/rand"
	"encoding/hex"
	"github.com/pkg/errors"
)

func generateApiToken() (string, error) {
	bytes := make([]byte, 64)
	if _, err := rand.Read(bytes); err != nil {
		return "", errors.Wrap(err, "error generating random api token")
	}
	return hex.EncodeToString(bytes), nil
}

func generateIdentityId() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", errors.Wrap(err, "error generating random identity id")
	}
	return hex.EncodeToString(bytes), nil
}
