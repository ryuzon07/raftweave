package jwt

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"testing"
	"time"

	jwtlib "github.com/golang-jwt/jwt/v5"
	"github.com/raftweave/raftweave/internal/auth/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testKey = func() []byte {
	key, _ := rsa.GenerateKey(rand.Reader, 2048)
	return pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	})
}()

func testUser() *domain.User {
	return &domain.User{
		ID:       "user-001",
		Email:    "test@raftweave.io",
		Name:     "Test User",
		Provider: domain.ProviderGitHub,
	}
}

func testRoles() map[string]string {
	return map[string]string{"ws-1": "ADMIN", "ws-2": "MEMBER"}
}

func TestIssueAndValidate_RoundTrip_Success(t *testing.T) {
	issuer, err := New(testKey, "https://auth.raftweave.io", []string{"https://api.raftweave.io"})
	require.NoError(t, err)

	user := testUser()
	roles := testRoles()

	token, err := issuer.IssueAccessToken(context.Background(), user, "session-1", roles)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	claims, err := issuer.Validate(context.Background(), token)
	require.NoError(t, err)

	assert.Equal(t, user.ID, claims.UserID)
	assert.Equal(t, user.Email, claims.Email)
	assert.Equal(t, user.Name, claims.Name)
	assert.Equal(t, user.Provider, claims.Provider)
	assert.Equal(t, "session-1", claims.SessionID)
	assert.Equal(t, roles, claims.Roles)
	assert.Equal(t, "access", claims.TokenType)
	assert.Equal(t, "https://auth.raftweave.io", claims.Issuer)
	assert.Contains(t, claims.Audience, "https://api.raftweave.io")

	// Verify TTL is ~15 minutes
	expectedExp := time.Now().UTC().Add(15 * time.Minute)
	assert.WithinDuration(t, expectedExp, claims.ExpiresAt.Time, 5*time.Second)
}

func TestValidate_ExpiredToken_ReturnsErrTokenExpired(t *testing.T) {
	// Parse the key manually to create a token with a past expiry.
	block, _ := pem.Decode(testKey)
	privKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	require.NoError(t, err)

	now := time.Now().UTC().Add(-1 * time.Hour)
	claims := Claims{
		RegisteredClaims: jwtlib.RegisteredClaims{
			Issuer:    "https://auth.raftweave.io",
			Audience:  jwtlib.ClaimStrings{"https://api.raftweave.io"},
			ExpiresAt: jwtlib.NewNumericDate(now.Add(15 * time.Minute)), // expired ~45 min ago
			NotBefore: jwtlib.NewNumericDate(now),
			IssuedAt:  jwtlib.NewNumericDate(now),
			ID:        "jti-expired",
			Subject:   "user-001",
		},
		UserID:    "user-001",
		TokenType: "access",
	}
	token := jwtlib.NewWithClaims(jwtlib.SigningMethodRS256, claims)
	signed, err := token.SignedString(privKey)
	require.NoError(t, err)

	issuer, err := New(testKey, "https://auth.raftweave.io", []string{"https://api.raftweave.io"})
	require.NoError(t, err)

	_, err = issuer.Validate(context.Background(), signed)
	assert.ErrorIs(t, err, domain.ErrTokenExpired)
}

func TestValidate_WrongAlgorithm_HS256_ReturnsErrTokenInvalid(t *testing.T) {
	// Security-critical: HS256 algorithm confusion attack.
	// An attacker creates a token signed with HS256 using the public key as secret.
	issuer, err := New(testKey, "https://auth.raftweave.io", []string{"https://api.raftweave.io"})
	require.NoError(t, err)

	claims := Claims{
		RegisteredClaims: jwtlib.RegisteredClaims{
			Issuer:    "https://auth.raftweave.io",
			Audience:  jwtlib.ClaimStrings{"https://api.raftweave.io"},
			ExpiresAt: jwtlib.NewNumericDate(time.Now().UTC().Add(15 * time.Minute)),
			NotBefore: jwtlib.NewNumericDate(time.Now().UTC()),
			IssuedAt:  jwtlib.NewNumericDate(time.Now().UTC()),
			ID:        "jti-hs256",
			Subject:   "user-001",
		},
		UserID:    "user-001",
		TokenType: "access",
	}

	// Sign with HS256 using the public key bytes as the HMAC secret.
	pubKeyBytes := x509.MarshalPKCS1PublicKey(issuer.PublicKey())
	token := jwtlib.NewWithClaims(jwtlib.SigningMethodHS256, claims)
	signed, err := token.SignedString(pubKeyBytes)
	require.NoError(t, err)

	_, err = issuer.Validate(context.Background(), signed)
	assert.ErrorIs(t, err, domain.ErrTokenInvalid)
}

func TestValidate_AlgorithmNone_ReturnsErrTokenInvalid(t *testing.T) {
	// Security-critical: "none" algorithm bypass attack.
	issuer, err := New(testKey, "https://auth.raftweave.io", []string{"https://api.raftweave.io"})
	require.NoError(t, err)

	claims := Claims{
		RegisteredClaims: jwtlib.RegisteredClaims{
			Issuer:    "https://auth.raftweave.io",
			Audience:  jwtlib.ClaimStrings{"https://api.raftweave.io"},
			ExpiresAt: jwtlib.NewNumericDate(time.Now().UTC().Add(15 * time.Minute)),
			NotBefore: jwtlib.NewNumericDate(time.Now().UTC()),
			IssuedAt:  jwtlib.NewNumericDate(time.Now().UTC()),
			ID:        "jti-none",
			Subject:   "user-001",
		},
		UserID:    "user-001",
		TokenType: "access",
	}

	// jwtlib.UnsafeAllowNoneSignatureType is needed to sign with "none" for testing.
	token := jwtlib.NewWithClaims(jwtlib.SigningMethodNone, claims)
	signed, err := token.SignedString(jwtlib.UnsafeAllowNoneSignatureType)
	require.NoError(t, err)

	_, err = issuer.Validate(context.Background(), signed)
	assert.ErrorIs(t, err, domain.ErrTokenInvalid)
}

func TestValidate_TamperedSignature_ReturnsErrTokenInvalid(t *testing.T) {
	issuer, err := New(testKey, "https://auth.raftweave.io", []string{"https://api.raftweave.io"})
	require.NoError(t, err)

	token, err := issuer.IssueAccessToken(context.Background(), testUser(), "session-1", testRoles())
	require.NoError(t, err)

	// Tamper with the last character of the signature.
	tampered := token[:len(token)-1] + "X"

	_, err = issuer.Validate(context.Background(), tampered)
	assert.ErrorIs(t, err, domain.ErrTokenInvalid)
}

func TestValidate_WrongIssuer_ReturnsErrTokenInvalid(t *testing.T) {
	// Create a token with the wrong issuer.
	block, _ := pem.Decode(testKey)
	privKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	require.NoError(t, err)

	claims := Claims{
		RegisteredClaims: jwtlib.RegisteredClaims{
			Issuer:    "https://evil.attacker.com",
			Audience:  jwtlib.ClaimStrings{"https://api.raftweave.io"},
			ExpiresAt: jwtlib.NewNumericDate(time.Now().UTC().Add(15 * time.Minute)),
			NotBefore: jwtlib.NewNumericDate(time.Now().UTC()),
			IssuedAt:  jwtlib.NewNumericDate(time.Now().UTC()),
			ID:        "jti-issuer",
			Subject:   "user-001",
		},
		UserID:    "user-001",
		TokenType: "access",
	}
	token := jwtlib.NewWithClaims(jwtlib.SigningMethodRS256, claims)
	signed, err := token.SignedString(privKey)
	require.NoError(t, err)

	issuer, err := New(testKey, "https://auth.raftweave.io", []string{"https://api.raftweave.io"})
	require.NoError(t, err)

	_, err = issuer.Validate(context.Background(), signed)
	assert.ErrorIs(t, err, domain.ErrTokenInvalid)
}

func TestValidate_WrongAudience_ReturnsErrTokenInvalid(t *testing.T) {
	block, _ := pem.Decode(testKey)
	privKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	require.NoError(t, err)

	claims := Claims{
		RegisteredClaims: jwtlib.RegisteredClaims{
			Issuer:    "https://auth.raftweave.io",
			Audience:  jwtlib.ClaimStrings{"https://wrong-audience.com"},
			ExpiresAt: jwtlib.NewNumericDate(time.Now().UTC().Add(15 * time.Minute)),
			NotBefore: jwtlib.NewNumericDate(time.Now().UTC()),
			IssuedAt:  jwtlib.NewNumericDate(time.Now().UTC()),
			ID:        "jti-aud",
			Subject:   "user-001",
		},
		UserID:    "user-001",
		TokenType: "access",
	}
	token := jwtlib.NewWithClaims(jwtlib.SigningMethodRS256, claims)
	signed, err := token.SignedString(privKey)
	require.NoError(t, err)

	issuer, err := New(testKey, "https://auth.raftweave.io", []string{"https://api.raftweave.io"})
	require.NoError(t, err)

	_, err = issuer.Validate(context.Background(), signed)
	assert.ErrorIs(t, err, domain.ErrTokenInvalid)
}

func TestValidate_RefreshMetaTokenAsAccess_ReturnsErrTokenInvalid(t *testing.T) {
	block, _ := pem.Decode(testKey)
	privKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	require.NoError(t, err)

	claims := Claims{
		RegisteredClaims: jwtlib.RegisteredClaims{
			Issuer:    "https://auth.raftweave.io",
			Audience:  jwtlib.ClaimStrings{"https://api.raftweave.io"},
			ExpiresAt: jwtlib.NewNumericDate(time.Now().UTC().Add(15 * time.Minute)),
			NotBefore: jwtlib.NewNumericDate(time.Now().UTC()),
			IssuedAt:  jwtlib.NewNumericDate(time.Now().UTC()),
			ID:        "jti-refresh",
			Subject:   "user-001",
		},
		UserID:    "user-001",
		TokenType: "refresh_meta", // NOT an access token
	}
	token := jwtlib.NewWithClaims(jwtlib.SigningMethodRS256, claims)
	signed, err := token.SignedString(privKey)
	require.NoError(t, err)

	issuer, err := New(testKey, "https://auth.raftweave.io", []string{"https://api.raftweave.io"})
	require.NoError(t, err)

	_, err = issuer.Validate(context.Background(), signed)
	assert.ErrorIs(t, err, domain.ErrTokenInvalid)
}

func TestIssue_JTIIsUniquePerToken(t *testing.T) {
	issuer, err := New(testKey, "https://auth.raftweave.io", []string{"https://api.raftweave.io"})
	require.NoError(t, err)

	user := testUser()
	roles := testRoles()

	const count = 100
	jtiSet := make(map[string]struct{}, count)

	for i := 0; i < count; i++ {
		token, err := issuer.IssueAccessToken(context.Background(), user, "session-1", roles)
		require.NoError(t, err)

		claims, err := issuer.Validate(context.Background(), token)
		require.NoError(t, err)

		_, exists := jtiSet[claims.ID]
		assert.False(t, exists, "JTI collision detected: %s", claims.ID)
		jtiSet[claims.ID] = struct{}{}
	}

	assert.Len(t, jtiSet, count)
}
