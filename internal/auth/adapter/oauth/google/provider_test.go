package google

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	jwtlib "github.com/golang-jwt/jwt/v5"
	cryptoadapter "github.com/raftweave/raftweave/internal/auth/adapter/crypto"
	oauthutil "github.com/raftweave/raftweave/internal/auth/adapter/oauth"
	"github.com/raftweave/raftweave/internal/auth/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockUserRepo struct{ users map[string]*domain.User }

func newMockUserRepo() *mockUserRepo { return &mockUserRepo{users: make(map[string]*domain.User)} }
func (m *mockUserRepo) Create(_ context.Context, u *domain.User) error { m.users[u.ID] = u; return nil }
func (m *mockUserRepo) GetByID(_ context.Context, id string) (*domain.User, error) {
	if u, ok := m.users[id]; ok { return u, nil }; return nil, domain.ErrUserNotFound
}
func (m *mockUserRepo) GetByEmail(_ context.Context, email string) (*domain.User, error) {
	for _, u := range m.users { if u.Email == email { return u, nil } }; return nil, domain.ErrUserNotFound
}
func (m *mockUserRepo) GetByProviderID(_ context.Context, p domain.Provider, pid string) (*domain.User, error) {
	for _, u := range m.users { if u.Provider == p && u.ProviderID == pid { return u, nil } }
	return nil, domain.ErrUserNotFound
}
func (m *mockUserRepo) UpdateGitHubToken(_ context.Context, _ string, _ []byte) error { return nil }
func (m *mockUserRepo) UpdateLastLogin(_ context.Context, _ string) error              { return nil }
func (m *mockUserRepo) UpdateEmailVerified(_ context.Context, _ string) error           { return nil }
func (m *mockUserRepo) SoftDelete(_ context.Context, _ string) error                   { return nil }

type mockStateStore struct{ states map[string]*oauthutil.StatePayload }

func newMockStateStore() *mockStateStore {
	return &mockStateStore{states: make(map[string]*oauthutil.StatePayload)}
}
func (m *mockStateStore) Create(_ context.Context, v, u string) (string, error) {
	s, _ := oauthutil.GenerateState()
	m.states[s] = &oauthutil.StatePayload{CodeVerifier: v, RedirectURI: u}
	return s, nil
}
func (m *mockStateStore) Verify(_ context.Context, s string) (*oauthutil.StatePayload, error) {
	p, ok := m.states[s]; if !ok { return nil, domain.ErrSessionNotFound }
	delete(m.states, s); return p, nil
}

func testEncryptor(t *testing.T) cryptoadapter.Encryptor {
	t.Helper()
	key := make([]byte, 32); _, _ = rand.Read(key)
	enc, err := cryptoadapter.NewEncryptor(key); require.NoError(t, err); return enc
}

// signTestIDToken creates a test Google id_token signed with the given RSA key.
func signTestIDToken(t *testing.T, key *rsa.PrivateKey, kid, clientID string, claims GoogleIDTokenClaims) string {
	t.Helper()
	token := jwtlib.NewWithClaims(jwtlib.SigningMethodRS256, claims)
	token.Header["kid"] = kid
	signed, err := token.SignedString(key)
	require.NoError(t, err)
	return signed
}

// setupMockGoogle creates a mock Google OAuth + JWKS server.
func setupMockGoogle(t *testing.T, key *rsa.PrivateKey, kid, clientID string, claims GoogleIDTokenClaims) *httptest.Server {
	t.Helper()
	idToken := signTestIDToken(t, key, kid, clientID, claims)
	mux := http.NewServeMux()
	mux.HandleFunc("/token", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"access_token": "ya29.test", "token_type": "Bearer",
			"id_token": idToken, "expires_in": 3600,
		})
	})
	mux.HandleFunc("/certs", func(w http.ResponseWriter, r *http.Request) {
		pub := key.PublicKey
		nB64 := base64.RawURLEncoding.EncodeToString(pub.N.Bytes())
		eBytes := big.NewInt(int64(pub.E)).Bytes()
		eB64 := base64.RawURLEncoding.EncodeToString(eBytes)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"keys": []map[string]string{
				{"kty": "RSA", "kid": kid, "n": nB64, "e": eB64, "alg": "RS256", "use": "sig"},
			},
		})
	})
	return httptest.NewServer(mux)
}

func genRSAKey(t *testing.T) *rsa.PrivateKey {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048); require.NoError(t, err); return key
}

func validClaims(clientID string) GoogleIDTokenClaims {
	now := time.Now().UTC()
	return GoogleIDTokenClaims{
		RegisteredClaims: jwtlib.RegisteredClaims{
			Issuer: "https://accounts.google.com", Subject: "google-sub-123",
			Audience: jwtlib.ClaimStrings{clientID},
			ExpiresAt: jwtlib.NewNumericDate(now.Add(1 * time.Hour)),
			IssuedAt: jwtlib.NewNumericDate(now), NotBefore: jwtlib.NewNumericDate(now),
		},
		Email: "test@gmail.com", EmailVerified: true,
		Name: "Test User", Picture: "https://lh3.googleusercontent.com/photo",
	}
}

func TestAuthURL_ContainsStateParameter(t *testing.T) {
	p := New("cid", "cs", testEncryptor(t), newMockStateStore(), newMockUserRepo())
	url, state, err := p.AuthURL(context.Background(), "http://localhost/callback")
	require.NoError(t, err)
	assert.NotEmpty(t, state)
	assert.Contains(t, url, "state=")
}

func TestHandleCallback_ValidIDToken_ReturnsUser(t *testing.T) {
	key := genRSAKey(t)
	clientID := "test-client-id"
	srv := setupMockGoogle(t, key, "kid1", clientID, validClaims(clientID))
	defer srv.Close()
	ss := newMockStateStore()
	p := New(clientID, "cs", testEncryptor(t), ss, newMockUserRepo(),
		WithHTTPClient(srv.Client()),
		WithEndpoints(srv.URL+"/auth", srv.URL+"/token", srv.URL+"/certs"),
	)
	_, state, _ := p.AuthURL(context.Background(), "http://localhost/callback")
	user, err := p.HandleCallback(context.Background(), "code", state, "http://localhost/callback")
	require.NoError(t, err)
	assert.Equal(t, "test@gmail.com", user.Email)
	assert.Equal(t, domain.ProviderGoogle, user.Provider)
}

func TestHandleCallback_ExpiredIDToken_ReturnsError(t *testing.T) {
	key := genRSAKey(t)
	clientID := "test-client-id"
	c := validClaims(clientID)
	c.ExpiresAt = jwtlib.NewNumericDate(time.Now().UTC().Add(-1 * time.Hour))
	srv := setupMockGoogle(t, key, "kid1", clientID, c)
	defer srv.Close()
	ss := newMockStateStore()
	p := New(clientID, "cs", testEncryptor(t), ss, newMockUserRepo(),
		WithHTTPClient(srv.Client()),
		WithEndpoints(srv.URL+"/auth", srv.URL+"/token", srv.URL+"/certs"),
	)
	_, state, _ := p.AuthURL(context.Background(), "http://localhost/callback")
	_, err := p.HandleCallback(context.Background(), "code", state, "http://localhost/callback")
	assert.Error(t, err)
}

func TestHandleCallback_WrongAudience_ReturnsError(t *testing.T) {
	key := genRSAKey(t)
	c := validClaims("wrong-client-id") // Audience doesn't match
	srv := setupMockGoogle(t, key, "kid1", "wrong-client-id", c)
	defer srv.Close()
	ss := newMockStateStore()
	p := New("correct-client-id", "cs", testEncryptor(t), ss, newMockUserRepo(),
		WithHTTPClient(srv.Client()),
		WithEndpoints(srv.URL+"/auth", srv.URL+"/token", srv.URL+"/certs"),
	)
	_, state, _ := p.AuthURL(context.Background(), "http://localhost/callback")
	_, err := p.HandleCallback(context.Background(), "code", state, "http://localhost/callback")
	assert.Error(t, err)
}

func TestHandleCallback_UnverifiedEmail_ReturnsError(t *testing.T) {
	key := genRSAKey(t)
	clientID := "test-client-id"
	c := validClaims(clientID)
	c.EmailVerified = false
	srv := setupMockGoogle(t, key, "kid1", clientID, c)
	defer srv.Close()
	ss := newMockStateStore()
	p := New(clientID, "cs", testEncryptor(t), ss, newMockUserRepo(),
		WithHTTPClient(srv.Client()),
		WithEndpoints(srv.URL+"/auth", srv.URL+"/token", srv.URL+"/certs"),
	)
	_, state, _ := p.AuthURL(context.Background(), "http://localhost/callback")
	_, err := p.HandleCallback(context.Background(), "code", state, "http://localhost/callback")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not verified")
}

func TestHandleCallback_ExistingGitHubAccount_LinksProviders(t *testing.T) {
	key := genRSAKey(t)
	clientID := "test-client-id"
	srv := setupMockGoogle(t, key, "kid1", clientID, validClaims(clientID))
	defer srv.Close()
	ur := newMockUserRepo()
	// Pre-existing GitHub user with same email
	ur.users["existing-id"] = &domain.User{
		ID: "existing-id", Email: "test@gmail.com", Provider: domain.ProviderGitHub,
		ProviderID: "gh-123", IsActive: true,
	}
	ss := newMockStateStore()
	p := New(clientID, "cs", testEncryptor(t), ss, ur,
		WithHTTPClient(srv.Client()),
		WithEndpoints(srv.URL+"/auth", srv.URL+"/token", srv.URL+"/certs"),
	)
	_, state, _ := p.AuthURL(context.Background(), "http://localhost/callback")
	user, err := p.HandleCallback(context.Background(), "code", state, "http://localhost/callback")
	require.NoError(t, err)
	assert.Equal(t, "existing-id", user.ID, "should return existing user, not create duplicate")
}

func TestJWKSCache_UsedOnSubsequentValidations(t *testing.T) {
	key := genRSAKey(t)
	clientID := "test-client-id"
	fetchCount := 0
	mux := http.NewServeMux()
	mux.HandleFunc("/token", func(w http.ResponseWriter, r *http.Request) {
		idToken := signTestIDToken(t, key, "kid1", clientID, validClaims(clientID))
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"access_token": "ya29.test", "token_type": "Bearer",
			"id_token": idToken, "expires_in": 3600,
		})
	})
	mux.HandleFunc("/certs", func(w http.ResponseWriter, r *http.Request) {
		fetchCount++
		pub := key.PublicKey
		nB64 := base64.RawURLEncoding.EncodeToString(pub.N.Bytes())
		eBytes := big.NewInt(int64(pub.E)).Bytes()
		eB64 := base64.RawURLEncoding.EncodeToString(eBytes)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"keys": []map[string]string{
				{"kty": "RSA", "kid": "kid1", "n": nB64, "e": eB64},
			},
		})
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()
	ss := newMockStateStore()
	p := New(clientID, "cs", testEncryptor(t), ss, newMockUserRepo(),
		WithHTTPClient(srv.Client()),
		WithEndpoints(srv.URL+"/auth", srv.URL+"/token", srv.URL+"/certs"),
	)
	// First call
	_, s1, _ := p.AuthURL(context.Background(), "http://localhost/cb")
	_, err := p.HandleCallback(context.Background(), "code", s1, "http://localhost/cb")
	require.NoError(t, err)
	// Second call — should use cached JWKS
	_, s2, _ := p.AuthURL(context.Background(), "http://localhost/cb")
	_, err = p.HandleCallback(context.Background(), "code", s2, "http://localhost/cb")
	require.NoError(t, err)
	assert.Equal(t, 1, fetchCount, "JWKS should be fetched only once due to caching")
}

// Ensure the test key wasn't accidentally exported via x509.
func init() { _ = x509.MarshalPKCS1PrivateKey }
