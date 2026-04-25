// Package crypto provides AES-256-GCM encryption for secrets stored at rest.
//
// This adapter encrypts sensitive credentials like GitHub OAuth tokens before
// persistence. The ciphertext format is versioned to support future key rotation:
//
//	[ 1-byte version | 12-byte nonce | GCM ciphertext+tag ]
package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
)

const (
	// currentVersion is the encryption format version.
	// Bump this when the key or algorithm changes to allow transparent migration.
	currentVersion byte = 0x01

	// nonceSize is the standard GCM nonce size in bytes.
	nonceSize = 12

	// keySize is the required AES-256 key length in bytes.
	keySize = 32
)

// Encryptor provides authenticated encryption using AES-256-GCM.
type Encryptor interface {
	// Encrypt encrypts plaintext and returns versioned ciphertext.
	// The nonce is randomly generated and prepended to the output.
	Encrypt(plaintext []byte) (ciphertext []byte, err error)

	// Decrypt decrypts versioned ciphertext produced by Encrypt.
	// Returns the original plaintext on success.
	Decrypt(ciphertext []byte) (plaintext []byte, err error)
}

type aesGCMEncryptor struct {
	aead cipher.AEAD
}

// NewEncryptor creates a new AES-256-GCM encryptor from a 32-byte key.
// The key MUST be loaded from a secure source (e.g., Kubernetes Secret, env var)
// and MUST NOT be logged or included in error messages.
func NewEncryptor(key []byte) (Encryptor, error) {
	if len(key) != keySize {
		return nil, fmt.Errorf("crypto.NewEncryptor: key must be exactly %d bytes, got %d", keySize, len(key))
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("crypto.NewEncryptor: %w", err)
	}

	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("crypto.NewEncryptor: %w", err)
	}

	return &aesGCMEncryptor{aead: aead}, nil
}

// Encrypt encrypts plaintext using AES-256-GCM with a random nonce.
// Output format: [ version(1) | nonce(12) | ciphertext+tag ]
func (e *aesGCMEncryptor) Encrypt(plaintext []byte) ([]byte, error) {
	if plaintext == nil {
		return nil, errors.New("crypto.Encrypt: plaintext must not be nil")
	}

	nonce := make([]byte, nonceSize)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("crypto.Encrypt: failed to generate nonce: %w", err)
	}

	// Seal appends the ciphertext+tag to the dst slice.
	ciphertext := e.aead.Seal(nil, nonce, plaintext, nil)

	// Prepend version byte and nonce.
	out := make([]byte, 0, 1+nonceSize+len(ciphertext))
	out = append(out, currentVersion)
	out = append(out, nonce...)
	out = append(out, ciphertext...)

	return out, nil
}

// Decrypt decrypts versioned ciphertext produced by Encrypt.
func (e *aesGCMEncryptor) Decrypt(data []byte) ([]byte, error) {
	if data == nil {
		return nil, errors.New("crypto.Decrypt: ciphertext must not be nil")
	}

	// Minimum length: version(1) + nonce(12) + tag(16) = 29 bytes
	minLen := 1 + nonceSize + e.aead.Overhead()
	if len(data) < minLen {
		return nil, errors.New("crypto.Decrypt: ciphertext too short")
	}

	version := data[0]
	if version != currentVersion {
		return nil, fmt.Errorf("crypto.Decrypt: unsupported version %d", version)
	}

	nonce := data[1 : 1+nonceSize]
	ciphertext := data[1+nonceSize:]

	plaintext, err := e.aead.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("crypto.Decrypt: decryption failed (tampered or wrong key): %w", err)
	}

	if len(plaintext) == 0 {
		return []byte{}, nil
	}

	return plaintext, nil
}
