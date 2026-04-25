package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"

	"github.com/raftweave/raftweave/internal/ingestion/domain"
)

type AESEncryptor struct {
	keys           map[string][]byte
	currentVersion string
}

// NewAESEncryptor creates a new domain.Encryptor using AES-256-GCM.
// Ensure all provided keys are exactly 32 bytes long.
func NewAESEncryptor(keys map[string][]byte, currentVersion string) (*AESEncryptor, error) {
	if len(keys) == 0 {
		return nil, errors.New("no keys provided")
	}
	if _, ok := keys[currentVersion]; !ok {
		return nil, fmt.Errorf("current key version %q not found in keys map", currentVersion)
	}

	for version, key := range keys {
		if len(key) != 32 {
			return nil, fmt.Errorf("invalid key length for version %q: must be exactly 32 bytes for AES-256", version)
		}
	}

	return &AESEncryptor{
		keys:           keys,
		currentVersion: currentVersion,
	}, nil
}

// Encrypt returns a base64 encoded byte slice containing the nonce, ciphertext, and tag.
func (e *AESEncryptor) Encrypt(plaintext []byte) (ciphertext []byte, keyVersion string, err error) {
	key := e.keys[e.currentVersion]
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, "", domain.ErrEncryptionFailed
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, "", domain.ErrEncryptionFailed
	}

	nonce := make([]byte, aesgcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, "", domain.ErrEncryptionFailed
	}

	sealed := aesgcm.Seal(nil, nonce, plaintext, nil)

	// Combine nonce + sealed payload
	combined := append(nonce, sealed...)

	// Base64 encode the final combination per requirements
	encoded := make([]byte, base64.StdEncoding.EncodedLen(len(combined)))
	base64.StdEncoding.Encode(encoded, combined)

	return encoded, e.currentVersion, nil
}

// Decrypt decodes the base64 ciphertext, extracts the nonce, and decrypts the payload.
func (e *AESEncryptor) Decrypt(ciphertext []byte, keyVersion string) (plaintext []byte, err error) {
	key, ok := e.keys[keyVersion]
	if !ok {
		return nil, domain.ErrDecryptionFailed
	}

	// Base64 Decode
	decoded := make([]byte, base64.StdEncoding.DecodedLen(len(ciphertext)))
	n, err := base64.StdEncoding.Decode(decoded, ciphertext)
	if err != nil {
		return nil, domain.ErrDecryptionFailed
	}
	decoded = decoded[:n]

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, domain.ErrDecryptionFailed
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, domain.ErrDecryptionFailed
	}

	nonceSize := aesgcm.NonceSize()
	if len(decoded) < nonceSize {
		return nil, domain.ErrDecryptionFailed
	}

	nonce, cipherBytes := decoded[:nonceSize], decoded[nonceSize:]
	plaintext, err = aesgcm.Open(nil, nonce, cipherBytes, nil)
	if err != nil {
		// e.g. authentication tag validation failed
		return nil, domain.ErrDecryptionFailed
	}

	return plaintext, nil
}
