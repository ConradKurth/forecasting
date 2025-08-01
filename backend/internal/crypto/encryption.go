package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"database/sql/driver"
	"encoding/base64"
	"io"
	"strings"

	"github.com/ConradKurth/forecasting/backend/internal/config"
	"github.com/pkg/errors"
)

const EncryptionVersion string = "v1"

var encryptionKey []byte

func init() {
	hasher := sha256.New()
	hasher.Write([]byte(config.Values.Encryption.SecretKey))
	encryptionKey = hasher.Sum(nil)
}

type EncryptedSecret string

func (s EncryptedSecret) Value() (driver.Value, error) {
	if s == "" {
		return "", nil
	}

	return s.EncryptedString()
}

func (s EncryptedSecret) EncryptedString() (string, error) {
	return encryptWithVersion(string(s))
}

func (s *EncryptedSecret) Scan(value any) error {
	var encrypted string

	switch v := value.(type) {
	case []byte:
		encrypted = string(v)
	case string:
		encrypted = v
	case nil:
		*s = ""
		return nil
	default:
		return errors.New("unsupported type for EncryptedSecret.Scan")
	}

	if encrypted == "" {
		*s = ""
		return nil
	}

	plaintext, err := decryptWithVersion(encrypted)
	if err != nil {
		return err
	}
	*s = EncryptedSecret(plaintext)

	return nil
}

func (s EncryptedSecret) String() string {
	return string(s)
}

func encrypt(plaintext string) (string, error) {
	if plaintext == "" {
		return "", nil
	}

	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, aesGCM.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := aesGCM.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func encryptWithVersion(plaintext string) (string, error) {
	if plaintext == "" {
		return "", nil
	}

	encrypted, err := encrypt(plaintext)
	if err != nil {
		return "", err
	}

	return EncryptionVersion + ":" + encrypted, nil
}

func decryptWithVersion(encrypted string) (string, error) {
	if encrypted == "" {
		return "", nil
	}

	version, ciphertext, found := strings.Cut(encrypted, ":")
	if !found {
		return "", errors.New("invalid secret format")
	}

	switch version {
	case "", EncryptionVersion:
		return decrypt(ciphertext)
	default:
		return "", errors.Errorf("unsupported encryption version: %s", version)
	}
}

func decrypt(ciphertext string) (string, error) {
	if ciphertext == "" {
		return "", nil
	}

	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := aesGCM.NonceSize()
	if len(data) < nonceSize {
		return "", errors.New("ciphertext too short")
	}

	nonce, ciphertextData := data[:nonceSize], data[nonceSize:]
	plaintext, err := aesGCM.Open(nil, nonce, ciphertextData, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}
