package controller

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/binary"
	"github.com/michaelquigley/pfxlog"
	"golang.org/x/crypto/argon2"
)

type hashedPassword struct {
	Password string
	Salt     string
}

func salt() string {
	buf := make([]byte, binary.MaxVarintLen64)
	_, err := rand.Read(buf)

	if err != nil {
		pfxlog.Logger().Panic(err)
	}

	return base64.StdEncoding.EncodeToString(buf)
}

func hashPassword(password string) (*hashedPassword, error) {
	return rehashPassword(password, salt())
}

func rehashPassword(password string, salt string) (*hashedPassword, error) {
	s, err := base64.StdEncoding.DecodeString(salt)
	if err != nil {
		return nil, err
	}

	hash := argon2.IDKey([]byte(password), s, 1, 3*1024, 4, 32)

	return &hashedPassword{
		Password: base64.StdEncoding.EncodeToString(hash),
		Salt:     salt,
	}, nil
}
