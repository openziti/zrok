package dynamicProxy

import (
	"crypto/sha256"
	"fmt"

	"github.com/go-jose/go-jose/v4"
	"golang.org/x/crypto/hkdf"
)

// deriveKey uses HKDF to expand a "password" into a []byte to be used as a key; better than just a raw hash
func deriveKey(keyString string, sz int) ([]byte, error) {
	out := hkdf.New(sha256.New, []byte(keyString), nil, []byte("derived-key"))
	key := make([]byte, sz)
	_, err := out.Read(key)
	if err != nil {
		return nil, err
	}
	return key, nil
}

// encryptToken uses AES-GCM (256) to encrypt tokens for inclusion in session tokens so that they're opaque outside of
// the auth subsystem
func encryptToken(token string, key []byte) (string, error) {
	enc, err := jose.NewEncrypter(
		jose.A256GCM,
		jose.Recipient{
			Algorithm: jose.DIRECT,
			Key:       key,
		},
		nil,
	)
	if err != nil {
		return "", fmt.Errorf("failed to create encrypter: %v", err)
	}

	obj, err := enc.Encrypt([]byte(token))
	if err != nil {
		return "", fmt.Errorf("failed to encrypt token: %v", err)
	}

	return obj.CompactSerialize()
}

func decryptToken(encrypted string, key []byte) (string, error) {
	obj, err := jose.ParseEncrypted(encrypted, []jose.KeyAlgorithm{jose.DIRECT}, []jose.ContentEncryption{jose.A256GCM})
	if err != nil {
		return "", fmt.Errorf("failed to parse encrypted token: %v", err)
	}

	decrypted, err := obj.Decrypt(key)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt token: %v", err)
	}

	return string(decrypted), nil
}
