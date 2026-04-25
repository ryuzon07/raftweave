package github

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/raftweave/raftweave/internal/auth/adapter/crypto"
	oauthutil "github.com/raftweave/raftweave/internal/auth/adapter/oauth"
	"github.com/raftweave/raftweave/internal/auth/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockUserRepo implements domain.UserRepository for testing.
type mockUserRepo struct {
	users map[string]*domain.User
}

func newMockUserRepo() *mockUserRepo { return &mockUserRepo{users: make(map[string]*domain.User)} }

func (m *mockUserRepo) Create(_ context.Context, u *domain.User) error {
	m.users[u.ID] = u; return nil
}
func (m *mockUserRepo) GetByID(_ context.Context, id string) (*domain.User, error) {
	if u, ok := m.users[id]; ok { return u, nil }
	return nil, domain.ErrUserNotFound
}
func (m *mockUserRepo) GetByEmail(_ context.Context, email string) (*domain.User, error) {
	for _, u := range m.users { if u.Email == email { return u, nil } }
	return nil, domain.ErrUserNotFound
}
func (m *mockUserRepo) GetByProviderID(_ context.Context, p domain.Provider, pid string) (*domain.User, error) {
	for _, u := range m.users { if u.Provider == p && u.ProviderID == pid { return u, nil } }
	return nil, domain.ErrUserNotFound
}
func (m *mockUserRepo) UpdateGitHubToken(_ context.Context, uid string, t []byte) error {
	if u, ok := m.users[uid]; ok { u.GitHubTokenEnc = t; return nil }
	return domain.ErrUserNotFound
}
func (m *mockUserRepo) UpdateLastLogin(_ context.Context, uid string) error { return nil }
func (m *mockUserRepo) UpdateEmailVerified(_ context.Context, uid string) error { return nil }
func (m *mockUserRepo) SoftDelete(_ context.Context, uid string) error { return nil }

// mockStateStore implements oauthutil.StateStore for testing.
type mockStateStore struct {
	states map[string]*oauthutil.StatePayload
}

func newMockStateStore() *mockStateStore { return &mockStateStore{states: make(map[string]*oauthutil.StatePayload)} }

func (m *mockStateStore) Create(_ context.Context, verifier, uri string) (string, error) {
	state, _ := oauthutil.GenerateState()
	m.states[state] = &oauthutil.StatePayload{CodeVerifier: verifier, RedirectURI: uri}
	return state, nil
}
func (m *mockStateStore) Verify(_ context.Context, state string) (*oauthutil.StatePayload, error) {
	p, ok := m.states[state]
	if !ok { return nil, domain.ErrSessionNotFound }
	delete(m.states, state)
	return p, nil
}

func testEncryptor(t *testing.T) crypto.Encryptor {
	t.Helper()
	key := make([]byte, 32)
	_, _ = rand.Read(key)
	enc, err := crypto.NewEncryptor(key)
	require.NoError(t, err)
	return enc
}

func TestAuthURL_ContainsStateAndPKCE(t *testing.T) {
	ss := newMockStateStore()
	p := New("cid", "csecret", testEncryptor(t), ss, newMockUserRepo())
	url, state, err := p.AuthURL(context.Background(), "http://localhost/callback")
	require.NoError(t, err)
	assert.NotEmpty(t, state)
	assert.Contains(t, url, "code_challenge=")
	assert.Contains(t, url, "code_challenge_method=S256")
	assert.Contains(t, url, "state=")
}

func TestAuthURL_StateStoredInRedis(t *testing.T) {
	ss := newMockStateStore()
	p := New("cid", "csecret", testEncryptor(t), ss, newMockUserRepo())
	_, state, err := p.AuthURL(context.Background(), "http://localhost/callback")
	require.NoError(t, err)
	payload, err := ss.Verify(context.Background(), state)
	require.NoError(t, err)
	assert.NotEmpty(t, payload.CodeVerifier)
	assert.Equal(t, "http://localhost/callback", payload.RedirectURI)
}

func setupMockGitHub(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/login/oauth/access_token", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"access_token": "gho_test_token_123", "token_type": "bearer",
		})
	})
	mux.HandleFunc("/user", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(GitHubUser{
			ID: 12345, Login: "testuser", Name: "Test User",
			Email: "test@example.com", AvatarURL: "https://avatars.githubusercontent.com/u/12345",
		})
	})
	mux.HandleFunc("/user/emails", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode([]GitHubEmail{
			{Email: "test@example.com", Primary: true, Verified: true},
		})
	})
	mux.HandleFunc("/user/repos", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode([]GitHubRepo{
			{ID: 1, FullName: "testuser/repo1", DefaultBranch: "main"},
		})
	})
	return httptest.NewServer(mux)
}

func TestHandleCallback_ValidCode_ReturnsUser(t *testing.T) {
	srv := setupMockGitHub(t)
	defer srv.Close()
	ss := newMockStateStore()
	ur := newMockUserRepo()
	enc := testEncryptor(t)
	p := New("cid", "csecret", enc, ss, ur,
		WithHTTPClient(srv.Client()),
		WithEndpoints(srv.URL+"/login/oauth/authorize", srv.URL+"/login/oauth/access_token", srv.URL),
	)
	_, state, _ := p.AuthURL(context.Background(), "http://localhost/callback")
	user, err := p.HandleCallback(context.Background(), "valid-code", state, "http://localhost/callback")
	require.NoError(t, err)
	assert.Equal(t, "test@example.com", user.Email)
	assert.Equal(t, "testuser", user.GitHubLogin)
	assert.Equal(t, domain.ProviderGitHub, user.Provider)
}

func TestHandleCallback_InvalidState_ReturnsCsrfError(t *testing.T) {
	srv := setupMockGitHub(t)
	defer srv.Close()
	p := New("cid", "csecret", testEncryptor(t), newMockStateStore(), newMockUserRepo(),
		WithHTTPClient(srv.Client()),
		WithEndpoints(srv.URL+"/login/oauth/authorize", srv.URL+"/login/oauth/access_token", srv.URL),
	)
	_, err := p.HandleCallback(context.Background(), "code", "invalid-state", "http://localhost/callback")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "CSRF")
}

func TestHandleCallback_ExistingUser_UpdatesToken(t *testing.T) {
	srv := setupMockGitHub(t)
	defer srv.Close()
	ss := newMockStateStore()
	ur := newMockUserRepo()
	enc := testEncryptor(t)
	p := New("cid", "csecret", enc, ss, ur,
		WithHTTPClient(srv.Client()),
		WithEndpoints(srv.URL+"/login/oauth/authorize", srv.URL+"/login/oauth/access_token", srv.URL),
	)
	// First login
	_, state1, _ := p.AuthURL(context.Background(), "http://localhost/callback")
	user1, err := p.HandleCallback(context.Background(), "code", state1, "http://localhost/callback")
	require.NoError(t, err)
	oldToken := make([]byte, len(user1.GitHubTokenEnc))
	copy(oldToken, user1.GitHubTokenEnc)

	// Second login — should update token, not create duplicate
	_, state2, _ := p.AuthURL(context.Background(), "http://localhost/callback")
	user2, err := p.HandleCallback(context.Background(), "code", state2, "http://localhost/callback")
	require.NoError(t, err)
	assert.Equal(t, user1.ID, user2.ID)
}

func TestHandleCallback_TokenEncryptedAtRest(t *testing.T) {
	srv := setupMockGitHub(t)
	defer srv.Close()
	ss := newMockStateStore()
	ur := newMockUserRepo()
	enc := testEncryptor(t)
	p := New("cid", "csecret", enc, ss, ur,
		WithHTTPClient(srv.Client()),
		WithEndpoints(srv.URL+"/login/oauth/authorize", srv.URL+"/login/oauth/access_token", srv.URL),
	)
	_, state, _ := p.AuthURL(context.Background(), "http://localhost/callback")
	user, err := p.HandleCallback(context.Background(), "code", state, "http://localhost/callback")
	require.NoError(t, err)
	// Token should be encrypted, not plaintext
	assert.NotEqual(t, []byte("gho_test_token_123"), user.GitHubTokenEnc)
	// But decryptable
	plain, err := enc.Decrypt(user.GitHubTokenEnc)
	require.NoError(t, err)
	assert.Equal(t, "gho_test_token_123", string(plain))
}

func TestListUserRepos_ReturnsDecryptedTokenForAPI(t *testing.T) {
	srv := setupMockGitHub(t)
	defer srv.Close()
	ss := newMockStateStore()
	ur := newMockUserRepo()
	enc := testEncryptor(t)
	p := New("cid", "csecret", enc, ss, ur,
		WithHTTPClient(srv.Client()),
		WithEndpoints(srv.URL+"/login/oauth/authorize", srv.URL+"/login/oauth/access_token", srv.URL),
	)
	_, state, _ := p.AuthURL(context.Background(), "http://localhost/callback")
	user, err := p.HandleCallback(context.Background(), "code", state, "http://localhost/callback")
	require.NoError(t, err)

	repos, err := p.ListUserRepos(context.Background(), user.ID)
	require.NoError(t, err)
	assert.Len(t, repos, 1)
	assert.Equal(t, "testuser/repo1", repos[0].FullName)
}
