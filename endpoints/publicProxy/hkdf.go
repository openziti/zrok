package publicProxy

import (
	"crypto/sha256"

	"golang.org/x/crypto/hkdf"
)

// DeriveKey uses HKDF to derive a fixed-size []byte key from a key string
func DeriveKey(keyString string, sz int) ([]byte, error) {
	hkdf := hkdf.New(sha256.New, []byte(keyString), nil, []byte("derived-key"))
	key := make([]byte, sz)
	_, err := hkdf.Read(key)
	if err != nil {
		return nil, err
	}
	return key, nil
}
