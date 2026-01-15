package controller

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"image/png"
	"io"
	"strings"

	"github.com/pkg/errors"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
	"golang.org/x/crypto/argon2"
)

const (
	totpIssuer          = "zrok"
	totpPeriod          = 30
	totpDigits          = otp.DigitsSix
	totpAlgorithm       = otp.AlgorithmSHA1
	recoveryCodeLength  = 8
	recoveryCodeCount   = 10
)

// TotpSetupResult contains the data needed for a user to set up TOTP
type TotpSetupResult struct {
	Secret          string
	QrCode          string // base64 encoded PNG
	ProvisioningUri string
}

// GenerateTotpSecret creates a new TOTP secret for the given account email
func GenerateTotpSecret(accountEmail string) (*TotpSetupResult, error) {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      totpIssuer,
		AccountName: accountEmail,
		Period:      totpPeriod,
		Digits:      totpDigits,
		Algorithm:   totpAlgorithm,
	})
	if err != nil {
		return nil, errors.Wrap(err, "error generating TOTP key")
	}

	// Generate QR code as PNG
	qrImg, err := key.Image(200, 200)
	if err != nil {
		return nil, errors.Wrap(err, "error generating QR code image")
	}

	// Encode image to base64 PNG
	var imgBuf bytes.Buffer
	if err := png.Encode(&imgBuf, qrImg); err != nil {
		return nil, errors.Wrap(err, "error encoding QR code to PNG")
	}

	qrCode := "data:image/png;base64," + base64.StdEncoding.EncodeToString(imgBuf.Bytes())

	return &TotpSetupResult{
		Secret:          key.Secret(),
		QrCode:          qrCode,
		ProvisioningUri: key.URL(),
	}, nil
}

// ValidateTotpCode validates a TOTP code against the secret
// Accepts codes within ±1 time period for clock drift tolerance
func ValidateTotpCode(secret, code string) bool {
	return totp.Validate(code, secret)
}

// EncryptTotpSecret encrypts a TOTP secret using AES-256-GCM
// The key should be a 32-byte base64-encoded string
// Returns base64-encoded ciphertext with nonce prepended
func EncryptTotpSecret(plaintext, keyBase64 string) (string, error) {
	key, err := base64.StdEncoding.DecodeString(keyBase64)
	if err != nil {
		return "", errors.Wrap(err, "error decoding encryption key")
	}

	if len(key) != 32 {
		return "", errors.New("encryption key must be 32 bytes")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", errors.Wrap(err, "error creating cipher")
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", errors.Wrap(err, "error creating GCM")
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", errors.Wrap(err, "error generating nonce")
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// DecryptTotpSecret decrypts a TOTP secret that was encrypted with EncryptTotpSecret
func DecryptTotpSecret(ciphertextBase64, keyBase64 string) (string, error) {
	key, err := base64.StdEncoding.DecodeString(keyBase64)
	if err != nil {
		return "", errors.Wrap(err, "error decoding encryption key")
	}

	if len(key) != 32 {
		return "", errors.New("encryption key must be 32 bytes")
	}

	ciphertext, err := base64.StdEncoding.DecodeString(ciphertextBase64)
	if err != nil {
		return "", errors.Wrap(err, "error decoding ciphertext")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", errors.Wrap(err, "error creating cipher")
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", errors.Wrap(err, "error creating GCM")
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", errors.New("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", errors.Wrap(err, "error decrypting")
	}

	return string(plaintext), nil
}

// GenerateRecoveryCodes generates a set of recovery codes
// Returns both the plaintext codes (for showing to user) and their hashes (for storage)
func GenerateRecoveryCodes() ([]string, []string, error) {
	codes := make([]string, recoveryCodeCount)
	hashes := make([]string, recoveryCodeCount)

	for i := 0; i < recoveryCodeCount; i++ {
		code, err := generateRecoveryCode()
		if err != nil {
			return nil, nil, errors.Wrap(err, "error generating recovery code")
		}
		codes[i] = code
		hashes[i] = HashRecoveryCode(code)
	}

	return codes, hashes, nil
}

// generateRecoveryCode generates a single recovery code
// Format: XXXX-XXXX (8 alphanumeric characters with dash)
func generateRecoveryCode() (string, error) {
	const charset = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789" // Excludes easily confused chars: 0, O, 1, I
	bytes := make([]byte, recoveryCodeLength)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	for i := range bytes {
		bytes[i] = charset[int(bytes[i])%len(charset)]
	}

	// Format as XXXX-XXXX
	return fmt.Sprintf("%s-%s", string(bytes[:4]), string(bytes[4:])), nil
}

// HashRecoveryCode hashes a recovery code for secure storage
// Uses argon2id for resistance against brute force attacks
func HashRecoveryCode(code string) string {
	// Normalize the code: remove dashes and convert to uppercase
	normalized := strings.ToUpper(strings.ReplaceAll(code, "-", ""))

	// Use a fixed salt derived from the code itself for deterministic hashing
	// This allows us to verify codes without storing the salt separately
	saltInput := sha256.Sum256([]byte("zrok-recovery-" + normalized))
	salt := saltInput[:16]

	hash := argon2.IDKey([]byte(normalized), salt, 1, 3*1024, 4, 32)
	return hex.EncodeToString(hash)
}

// GenerateSecureToken generates a cryptographically secure random token
func GenerateSecureToken(length int) (string, error) {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return "", errors.Wrap(err, "error generating random bytes")
	}
	return base64.URLEncoding.EncodeToString(b), nil
}
