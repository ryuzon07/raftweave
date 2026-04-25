// Package github implements the GitHub OAuth 2.0 flow with PKCE and CSRF protection.
package github

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"golang.org/x/oauth2"

	"github.com/google/uuid"
	"github.com/raftweave/raftweave/internal/auth/adapter/crypto"
	oauthutil "github.com/raftweave/raftweave/internal/auth/adapter/oauth"
	"github.com/raftweave/raftweave/internal/auth/domain"
)

// RequiredScopes defines the GitHub OAuth scopes needed by RaftWeave.
var RequiredScopes = []string{
	"read:user", "user:email", "repo", "admin:repo_hook", "read:org",
}

// GitHubUser is the response from GitHub's /user API.
type GitHubUser struct {
	ID        int64  `json:"id"`
	Login     string `json:"login"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatar_url"`
}

// GitHubEmail is a single entry from GitHub's /user/emails API.
type GitHubEmail struct {
	Email    string `json:"email"`
	Primary  bool   `json:"primary"`
	Verified bool   `json:"verified"`
}

// GitHubRepo represents a GitHub repository.
type GitHubRepo struct {
	ID            int64  `json:"id"`
	FullName      string `json:"full_name"`
	HTMLURL       string `json:"html_url"`
	DefaultBranch string `json:"default_branch"`
	Private       bool   `json:"private"`
	CloneURL      string `json:"clone_url"`
}

// Provider implements the GitHub OAuth 2.0 flow.
type Provider interface {
	AuthURL(ctx context.Context, redirectURI string) (url, state string, err error)
	HandleCallback(ctx context.Context, code, state, redirectURI string) (*domain.User, error)
	ListUserRepos(ctx context.Context, userID string) ([]GitHubRepo, error)
}

type provider struct {
	clientID, clientSecret string
	encryptor              crypto.Encryptor
	stateStore             oauthutil.StateStore
	userRepo               domain.UserRepository
	httpClient             *http.Client
	authURL, tokenURL, apiURL string
}

// Option configures the GitHub OAuth provider.
type Option func(*provider)

// WithHTTPClient sets a custom HTTP client (for testing).
func WithHTTPClient(c *http.Client) Option { return func(p *provider) { p.httpClient = c } }

// WithEndpoints overrides GitHub endpoints (for testing).
func WithEndpoints(authURL, tokenURL, apiURL string) Option {
	return func(p *provider) { p.authURL = authURL; p.tokenURL = tokenURL; p.apiURL = apiURL }
}

// New creates a new GitHub OAuth provider.
func New(clientID, clientSecret string, enc crypto.Encryptor, ss oauthutil.StateStore, ur domain.UserRepository, opts ...Option) Provider {
	p := &provider{
		clientID: clientID, clientSecret: clientSecret,
		encryptor: enc, stateStore: ss, userRepo: ur,
		httpClient: http.DefaultClient,
		authURL:    "https://github.com/login/oauth/authorize",
		tokenURL:   "https://github.com/login/oauth/access_token",
		apiURL:     "https://api.github.com",
	}
	for _, o := range opts { o(p) }
	return p
}

func (p *provider) oauthConfig(redirectURI string) *oauth2.Config {
	return &oauth2.Config{
		ClientID: p.clientID, ClientSecret: p.clientSecret, RedirectURL: redirectURI,
		Endpoint: oauth2.Endpoint{AuthURL: p.authURL, TokenURL: p.tokenURL},
		Scopes:   RequiredScopes,
	}
}

func (p *provider) AuthURL(ctx context.Context, redirectURI string) (string, string, error) {
	verifier, challenge, err := oauthutil.GeneratePKCE()
	if err != nil {
		return "", "", fmt.Errorf("github.AuthURL: %w", err)
	}
	state, err := p.stateStore.Create(ctx, verifier, redirectURI)
	if err != nil {
		return "", "", fmt.Errorf("github.AuthURL: %w", err)
	}
	cfg := p.oauthConfig(redirectURI)
	url := cfg.AuthCodeURL(state,
		oauth2.SetAuthURLParam("code_challenge", challenge),
		oauth2.SetAuthURLParam("code_challenge_method", "S256"),
	)
	return url, state, nil
}

func (p *provider) HandleCallback(ctx context.Context, code, state, redirectURI string) (*domain.User, error) {
	payload, err := p.stateStore.Verify(ctx, state)
	if err != nil {
		return nil, fmt.Errorf("github.HandleCallback: CSRF invalid: %w", err)
	}
	cfg := p.oauthConfig(redirectURI)
	token, err := cfg.Exchange(ctx, code, oauth2.SetAuthURLParam("code_verifier", payload.CodeVerifier))
	if err != nil {
		return nil, fmt.Errorf("github.HandleCallback: exchange: %w", err)
	}
	ghUser, err := p.fetchUser(ctx, token.AccessToken)
	if err != nil {
		return nil, err
	}
	email := ghUser.Email
	if email == "" {
		email, err = p.fetchPrimaryEmail(ctx, token.AccessToken)
		if err != nil {
			return nil, err
		}
	}
	encToken, err := p.encryptor.Encrypt([]byte(token.AccessToken))
	if err != nil {
		return nil, fmt.Errorf("github.HandleCallback: encrypt: %w", err)
	}
	providerID := strconv.FormatInt(ghUser.ID, 10)

	existing, err := p.userRepo.GetByProviderID(ctx, domain.ProviderGitHub, providerID)
	if err == nil {
		_ = p.userRepo.UpdateGitHubToken(ctx, existing.ID, encToken)
		_ = p.userRepo.UpdateLastLogin(ctx, existing.ID)
		existing.GitHubTokenEnc = encToken
		now := time.Now().UTC()
		existing.LastLoginAt = &now
		return existing, nil
	}

	now := time.Now().UTC()
	user := &domain.User{
		ID: uuid.New().String(), Email: email, Name: ghUser.Name,
		AvatarURL: ghUser.AvatarURL, Provider: domain.ProviderGitHub,
		ProviderID: providerID, GitHubLogin: ghUser.Login,
		GitHubTokenEnc: encToken, IsEmailVerified: true, IsActive: true,
		CreatedAt: now, UpdatedAt: now, LastLoginAt: &now,
	}
	if err := p.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("github.HandleCallback: create: %w", err)
	}
	return user, nil
}

func (p *provider) ListUserRepos(ctx context.Context, userID string) ([]GitHubRepo, error) {
	user, err := p.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("github.ListUserRepos: %w", err)
	}
	if user.GitHubTokenEnc == nil {
		return nil, fmt.Errorf("github.ListUserRepos: no GitHub token")
	}
	plain, err := p.encryptor.Decrypt(user.GitHubTokenEnc)
	if err != nil {
		return nil, fmt.Errorf("github.ListUserRepos: decrypt: %w", err)
	}
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, p.apiURL+"/user/repos?per_page=100&sort=updated", nil)
	req.Header.Set("Authorization", "Bearer "+string(plain))
	req.Header.Set("Accept", "application/vnd.github+json")
	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("github.ListUserRepos: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("github.ListUserRepos: status %d", resp.StatusCode)
	}
	var repos []GitHubRepo
	if err := json.NewDecoder(resp.Body).Decode(&repos); err != nil {
		return nil, fmt.Errorf("github.ListUserRepos: decode: %w", err)
	}
	return repos, nil
}

func (p *provider) fetchUser(ctx context.Context, accessToken string) (*GitHubUser, error) {
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, p.apiURL+"/user", nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/vnd.github+json")
	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("github.fetchUser: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("github.fetchUser: status %d: %s", resp.StatusCode, body)
	}
	var u GitHubUser
	if err := json.NewDecoder(resp.Body).Decode(&u); err != nil {
		return nil, fmt.Errorf("github.fetchUser: decode: %w", err)
	}
	return &u, nil
}

func (p *provider) fetchPrimaryEmail(ctx context.Context, accessToken string) (string, error) {
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, p.apiURL+"/user/emails", nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/vnd.github+json")
	resp, err := p.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("github.fetchPrimaryEmail: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("github.fetchPrimaryEmail: status %d", resp.StatusCode)
	}
	var emails []GitHubEmail
	if err := json.NewDecoder(resp.Body).Decode(&emails); err != nil {
		return "", fmt.Errorf("github.fetchPrimaryEmail: decode: %w", err)
	}
	for _, e := range emails {
		if e.Primary && e.Verified {
			return e.Email, nil
		}
	}
	return "", fmt.Errorf("github.fetchPrimaryEmail: no verified primary email")
}
