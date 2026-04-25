package crypto_test

import (
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/raftweave/raftweave/internal/ingestion/adapter/crypto"
	"github.com/raftweave/raftweave/internal/ingestion/domain"
)

var (
	validKey1 = []byte("12345678901234567890123456789012") // 32 bytes
	validKey2 = []byte("abcdefghijklmnopqrstuvwxyz123456") // 32 bytes
)

func TestNewAESEncryptor_InvalidKeyLength(t *testing.T) {
	keys := map[string][]byte{"v1": []byte("too-short")}
	_, err := crypto.NewAESEncryptor(keys, "v1")
	assert.ErrorContains(t, err, "invalid key length")

	_, err = crypto.NewAESEncryptor(nil, "v1")
	assert.ErrorContains(t, err, "no keys provided")

	_, err = crypto.NewAESEncryptor(map[string][]byte{"v2": validKey1}, "v1")
	assert.ErrorContains(t, err, "not found in keys map")
}

func TestEncrypt_Decrypt_Roundtrip(t *testing.T) {
	enc, err := crypto.NewAESEncryptor(map[string][]byte{"v1": validKey1}, "v1")
	require.NoError(t, err)

	plaintext := []byte("highly-sensitive-aws-iam-credentials")

	ciphertext, keyVersion, err := enc.Encrypt(plaintext)
	require.NoError(t, err)
	assert.Equal(t, "v1", keyVersion)
	assert.NotEqual(t, plaintext, ciphertext)

	decrypted, err := enc.Decrypt(ciphertext, keyVersion)
	require.NoError(t, err)
	assert.Equal(t, plaintext, decrypted)
}

func TestEncrypt_ProducesUniqueNonces(t *testing.T) {
	enc, err := crypto.NewAESEncryptor(map[string][]byte{"v1": validKey1}, "v1")
	require.NoError(t, err)

	plaintext := []byte("some-secret-data")

	cipher1, _, _ := enc.Encrypt(plaintext)
	cipher2, _, _ := enc.Encrypt(plaintext)

	assert.NotEqual(t, cipher1, cipher2, "same plaintext must produce different ciphertexts due to unique nonces")
}

func TestDecrypt_WrongKey(t *testing.T) {
	enc, _ := crypto.NewAESEncryptor(map[string][]byte{"v1": validKey1}, "v1")
	
	plaintext := []byte("secret")
	ciphertext, _, _ := enc.Encrypt(plaintext)

	_, err := enc.Decrypt(ciphertext, "v2-missing")
	assert.ErrorIs(t, err, domain.ErrDecryptionFailed)
}

func TestDecrypt_TamperedCiphertext(t *testing.T) {
	enc, _ := crypto.NewAESEncryptor(map[string][]byte{"v1": validKey1}, "v1")
	
	plaintext := []byte("secret")
	ciphertext, keyVersion, _ := enc.Encrypt(plaintext)

	// tampering - decoding base64, modifying, encoding again
	decodedLen := base64.StdEncoding.DecodedLen(len(ciphertext))
	decoded := make([]byte, decodedLen)
	n, _ := base64.StdEncoding.Decode(decoded, ciphertext)
	decoded = decoded[:n]
	
	decoded[len(decoded)-1] ^= 0xFF // flip last byte of the GCM tag

	tamperedCipher := make([]byte, base64.StdEncoding.EncodedLen(len(decoded)))
	base64.StdEncoding.Encode(tamperedCipher, decoded)

	_, err := enc.Decrypt(tamperedCipher, keyVersion)
	assert.ErrorIs(t, err, domain.ErrDecryptionFailed)
}

func TestEncrypt_KeyVersioning(t *testing.T) {
	keys := map[string][]byte{
		"v1": validKey1,
		"v2": validKey2,
	}

	enc, err := crypto.NewAESEncryptor(keys, "v2") // currently rotating to v2
	require.NoError(t, err)

	plaintext := []byte("secret")

	// Verify encryption uses the current active version
	_, keyVersion, err := enc.Encrypt(plaintext)
	require.NoError(t, err)
	assert.Equal(t, "v2", keyVersion)

	// But decrypting an old credential still works if we pass 'v1'
	
	// Create a v1 ciphertext manually using AESEncryptor configured at v1
	encV1, _ := crypto.NewAESEncryptor(keys, "v1")
	oldCiphertext, oldVer, _ := encV1.Encrypt(plaintext)

	decryptedOld, err := enc.Decrypt(oldCiphertext, oldVer) // works through the keys map
	require.NoError(t, err)
	assert.Equal(t, plaintext, decryptedOld)
}
