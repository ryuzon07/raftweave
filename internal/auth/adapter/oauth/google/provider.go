// Package google implements the Google OAuth 2.0 + OpenID Connect flow.
package google

import (
	"context"
	"crypto/rsa"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"sync"
	"time"

	"golang.org/x/oauth2"
	googleoauth "golang.org/x/oauth2/google"

	jwtlib "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/raftweave/raftweave/internal/auth/adapter/crypto"
	oauthutil "github.com/raftweave/raftweave/internal/auth/adapter/oauth"
	"github.com/raftweave/raftweave/internal/auth/domain"
)

// RequiredScopes for Google OAuth.
var RequiredScopes = []string{"openid", "profile", "email"}

// GoogleIDTokenClaims from Google's id_token.
type GoogleIDTokenClaims struct {
	jwtlib.RegisteredClaims
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
}

// Provider implements the Google OAuth 2.0 + OIDC flow.
type Provider interface {
	AuthURL(ctx context.Context, redirectURI string) (url, state string, err error)
	HandleCallback(ctx context.Context, code, state, redirectURI string) (*domain.User, error)
}

type provider struct {
	clientID, clientSecret string
	encryptor              crypto.Encryptor
	stateStore             oauthutil.StateStore
	userRepo               domain.UserRepository
	httpClient             *http.Client
	jwksURL                string
	jwksCache              *jwksCache
	// Override for testing
	useGoogleEndpoint bool
	authURL, tokenURL string
}

// Option configures the Google OAuth provider.
type Option func(*provider)

// WithHTTPClient sets a custom HTTP client.
func WithHTTPClient(c *http.Client) Option { return func(p *provider) { p.httpClient = c } }

// WithEndpoints overrides Google endpoints (for testing).
func WithEndpoints(authURL, tokenURL, jwksURL string) Option {
	return func(p *provider) {
		p.authURL = authURL; p.tokenURL = tokenURL; p.jwksURL = jwksURL
		p.useGoogleEndpoint = false
	}
}

// New creates a new Google OAuth provider.
func New(clientID, clientSecret string, enc crypto.Encryptor, ss oauthutil.StateStore, ur domain.UserRepository, opts ...Option) Provider {
	p := &provider{
		clientID: clientID, clientSecret: clientSecret,
		encryptor: enc, stateStore: ss, userRepo: ur,
		httpClient: http.DefaultClient, useGoogleEndpoint: true,
		jwksURL:   "https://www.googleapis.com/oauth2/v3/certs",
		jwksCache: &jwksCache{},
	}
	for _, o := range opts { o(p) }
	return p
}

func (p *provider) oauthConfig(redirectURI string) *oauth2.Config {
	cfg := &oauth2.Config{
		ClientID: p.clientID, ClientSecret: p.clientSecret,
		RedirectURL: redirectURI, Scopes: RequiredScopes,
	}
	if p.useGoogleEndpoint {
		cfg.Endpoint = googleoauth.Endpoint
	} else {
		cfg.Endpoint = oauth2.Endpoint{AuthURL: p.authURL, TokenURL: p.tokenURL}
	}
	return cfg
}

func (p *provider) AuthURL(ctx context.Context, redirectURI string) (string, string, error) {
	verifier, challenge, err := oauthutil.GeneratePKCE()
	if err != nil { return "", "", fmt.Errorf("google.AuthURL: %w", err) }
	state, err := p.stateStore.Create(ctx, verifier, redirectURI)
	if err != nil { return "", "", fmt.Errorf("google.AuthURL: %w", err) }
	cfg := p.oauthConfig(redirectURI)
	url := cfg.AuthCodeURL(state,
		oauth2.SetAuthURLParam("code_challenge", challenge),
		oauth2.SetAuthURLParam("code_challenge_method", "S256"),
	)
	return url, state, nil
}

func (p *provider) HandleCallback(ctx context.Context, code, state, redirectURI string) (*domain.User, error) {
	payload, err := p.stateStore.Verify(ctx, state)
	if err != nil { return nil, fmt.Errorf("google.HandleCallback: CSRF invalid: %w", err) }

	cfg := p.oauthConfig(redirectURI)
	token, err := cfg.Exchange(ctx, code, oauth2.SetAuthURLParam("code_verifier", payload.CodeVerifier))
	if err != nil { return nil, fmt.Errorf("google.HandleCallback: exchange: %w", err) }

	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok || rawIDToken == "" {
		return nil, errors.New("google.HandleCallback: no id_token in response")
	}

	claims, err := p.validateIDToken(ctx, rawIDToken)
	if err != nil { return nil, fmt.Errorf("google.HandleCallback: %w", err) }

	if !claims.EmailVerified {
		return nil, errors.New("google.HandleCallback: email not verified by Google")
	}

	sub := claims.Subject
	existing, err := p.userRepo.GetByProviderID(ctx, domain.ProviderGoogle, sub)
	if err == nil {
		_ = p.userRepo.UpdateLastLogin(ctx, existing.ID)
		now := time.Now().UTC()
		existing.LastLoginAt = &now
		return existing, nil
	}

	// Check if email already exists (account linking with GitHub).
	if existingByEmail, err := p.userRepo.GetByEmail(ctx, claims.Email); err == nil {
		_ = p.userRepo.UpdateLastLogin(ctx, existingByEmail.ID)
		now := time.Now().UTC()
		existingByEmail.LastLoginAt = &now
		return existingByEmail, nil
	}

	now := time.Now().UTC()
	user := &domain.User{
		ID: uuid.New().String(), Email: claims.Email, Name: claims.Name,
		AvatarURL: claims.Picture, Provider: domain.ProviderGoogle,
		ProviderID: sub, IsEmailVerified: true, IsActive: true,
		CreatedAt: now, UpdatedAt: now, LastLoginAt: &now,
	}
	if err := p.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("google.HandleCallback: create: %w", err)
	}
	return user, nil
}

// validateIDToken validates Google's id_token JWT against Google's JWKS.
func (p *provider) validateIDToken(ctx context.Context, rawToken string) (*GoogleIDTokenClaims, error) {
	keys, err := p.jwksCache.getKeys(ctx, p.httpClient, p.jwksURL)
	if err != nil { return nil, fmt.Errorf("validateIDToken: jwks: %w", err) }

	token, err := jwtlib.ParseWithClaims(rawToken, &GoogleIDTokenClaims{}, func(t *jwtlib.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwtlib.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		kid, _ := t.Header["kid"].(string)
		if key, ok := keys[kid]; ok { return key, nil }
		return nil, fmt.Errorf("unknown kid: %s", kid)
	}, jwtlib.WithValidMethods([]string{"RS256"}))

	if err != nil {
		if errors.Is(err, jwtlib.ErrTokenExpired) { return nil, domain.ErrTokenExpired }
		return nil, fmt.Errorf("validateIDToken: %w", err)
	}

	claims, ok := token.Claims.(*GoogleIDTokenClaims)
	if !ok || !token.Valid { return nil, domain.ErrTokenInvalid }

	iss := claims.Issuer
	if iss != "accounts.google.com" && iss != "https://accounts.google.com" {
		return nil, fmt.Errorf("validateIDToken: invalid issuer: %s", iss)
	}
	audValid := false
	for _, a := range claims.Audience {
		if a == p.clientID {
			audValid = true
			break
		}
	}
	if !audValid {
		return nil, fmt.Errorf("validateIDToken: invalid audience")
	}

	return claims, nil
}

// jwksCache caches Google's JWKS keys with a 1-hour TTL.
type jwksCache struct {
	mu      sync.RWMutex
	keys    map[string]*rsa.PublicKey
	fetchAt time.Time
	ttl     time.Duration
	fetches int
}

type jwksResponse struct {
	Keys []jwksKey `json:"keys"`
}

type jwksKey struct {
	Kty string `json:"kty"`
	Kid string `json:"kid"`
	N   string `json:"n"`
	E   string `json:"e"`
}

func (c *jwksCache) getKeys(ctx context.Context, client *http.Client, url string) (map[string]*rsa.PublicKey, error) {
	c.mu.RLock()
	if c.keys != nil && time.Since(c.fetchAt) < c.ttl {
		defer c.mu.RUnlock()
		return c.keys, nil
	}
	c.mu.RUnlock()

	c.mu.Lock()
	defer c.mu.Unlock()
	// Double-check after acquiring write lock.
	if c.keys != nil && time.Since(c.fetchAt) < c.ttl {
		return c.keys, nil
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil { return nil, err }
	resp, err := client.Do(req)
	if err != nil { return nil, err }
	defer resp.Body.Close()

	var jwks jwksResponse
	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil { return nil, err }

	keys := make(map[string]*rsa.PublicKey)
	for _, k := range jwks.Keys {
		if k.Kty != "RSA" { continue }
		pub, err := parseRSAPublicKey(k.N, k.E)
		if err != nil { continue }
		keys[k.Kid] = pub
	}

	c.keys = keys
	c.fetchAt = time.Now()
	if c.ttl == 0 { c.ttl = 1 * time.Hour }
	c.fetches++
	return keys, nil
}

// FetchCount returns the number of JWKS fetches (for testing cache behavior).
func (c *jwksCache) FetchCount() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.fetches
}

func parseRSAPublicKey(nB64, eB64 string) (*rsa.PublicKey, error) {
	nBytes, err := jwtlib.NewParser().DecodeSegment(nB64)
	if err != nil { return nil, err }
	eBytes, err := jwtlib.NewParser().DecodeSegment(eB64)
	if err != nil { return nil, err }
	n := new(big.Int).SetBytes(nBytes)
	e := 0
	for _, b := range eBytes { e = e<<8 + int(b) }
	return &rsa.PublicKey{N: n, E: e}, nil
}
