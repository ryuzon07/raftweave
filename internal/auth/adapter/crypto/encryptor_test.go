package crypto

import (
	"crypto/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func generateTestKey(t *testing.T) []byte {
	t.Helper()
	key := make([]byte, 32)
	_, err := rand.Read(key)
	require.NoError(t, err)
	return key
}

func TestEncryptDecrypt_RoundTrip(t *testing.T) {
	enc, err := NewEncryptor(generateTestKey(t))
	require.NoError(t, err)

	plaintext := []byte("ghp_XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX")
	ciphertext, err := enc.Encrypt(plaintext)
	require.NoError(t, err)

	decrypted, err := enc.Decrypt(ciphertext)
	require.NoError(t, err)
	assert.Equal(t, plaintext, decrypted)
}

func TestEncrypt_ProducesUniqueCiphertexts(t *testing.T) {
	enc, err := NewEncryptor(generateTestKey(t))
	require.NoError(t, err)

	plaintext := []byte("same-input")
	ct1, err := enc.Encrypt(plaintext)
	require.NoError(t, err)
	ct2, err := enc.Encrypt(plaintext)
	require.NoError(t, err)

	// Due to random nonce, ciphertexts must differ even for same plaintext.
	assert.NotEqual(t, ct1, ct2)
}

func TestDecrypt_TamperedCiphertext_ReturnsError(t *testing.T) {
	enc, err := NewEncryptor(generateTestKey(t))
	require.NoError(t, err)

	ciphertext, err := enc.Encrypt([]byte("secret"))
	require.NoError(t, err)

	// Tamper with the last byte.
	ciphertext[len(ciphertext)-1] ^= 0xFF

	_, err = enc.Decrypt(ciphertext)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "decryption failed")
}

func TestDecrypt_WrongKey_ReturnsError(t *testing.T) {
	enc1, err := NewEncryptor(generateTestKey(t))
	require.NoError(t, err)
	enc2, err := NewEncryptor(generateTestKey(t))
	require.NoError(t, err)

	ciphertext, err := enc1.Encrypt([]byte("secret"))
	require.NoError(t, err)

	_, err = enc2.Decrypt(ciphertext)
	assert.Error(t, err)
}

func TestNewEncryptor_InvalidKeyLength_ReturnsError(t *testing.T) {
	_, err := NewEncryptor([]byte("short"))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "key must be exactly 32 bytes")
}

func TestDecrypt_TooShort_ReturnsError(t *testing.T) {
	enc, err := NewEncryptor(generateTestKey(t))
	require.NoError(t, err)

	_, err = enc.Decrypt([]byte{0x01, 0x02, 0x03})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ciphertext too short")
}

func TestDecrypt_UnsupportedVersion_ReturnsError(t *testing.T) {
	enc, err := NewEncryptor(generateTestKey(t))
	require.NoError(t, err)

	ciphertext, err := enc.Encrypt([]byte("data"))
	require.NoError(t, err)

	// Change the version byte to an unsupported value.
	ciphertext[0] = 0xFF

	_, err = enc.Decrypt(ciphertext)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported version")
}

func TestEncrypt_EmptyPlaintext_Success(t *testing.T) {
	enc, err := NewEncryptor(generateTestKey(t))
	require.NoError(t, err)

	ciphertext, err := enc.Encrypt([]byte{})
	require.NoError(t, err)

	decrypted, err := enc.Decrypt(ciphertext)
	require.NoError(t, err)
	assert.Equal(t, []byte{}, decrypted)
}

func TestEncrypt_NilPlaintext_ReturnsError(t *testing.T) {
	enc, err := NewEncryptor(generateTestKey(t))
	require.NoError(t, err)

	_, err = enc.Encrypt(nil)
	assert.Error(t, err)
}

func TestDecrypt_NilCiphertext_ReturnsError(t *testing.T) {
	enc, err := NewEncryptor(generateTestKey(t))
	require.NoError(t, err)

	_, err = enc.Decrypt(nil)
	assert.Error(t, err)
}

func TestCiphertext_ContainsVersionByte(t *testing.T) {
	enc, err := NewEncryptor(generateTestKey(t))
	require.NoError(t, err)

	ciphertext, err := enc.Encrypt([]byte("data"))
	require.NoError(t, err)

	assert.Equal(t, byte(0x01), ciphertext[0], "first byte should be version 0x01")
}
