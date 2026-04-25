// Package oauth contains shared types and utilities for OAuth providers.
package oauth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	// StatePrefix is the Redis key prefix for OAuth CSRF state tokens.
	StatePrefix = "rftw:oauth:state:"

	// StateTTL is the maximum time a state token remains valid.
	StateTTL = 10 * time.Minute
)

// StatePayload is the data associated with an OAuth state token.
type StatePayload struct {
	CodeVerifier string    `json:"code_verifier"`
	RedirectURI  string    `json:"redirect_uri"`
	CreatedAt    time.Time `json:"created_at"`
}

// StateStore manages OAuth CSRF state tokens in Redis.
type StateStore interface {
	// Create generates a cryptographically random state token, stores the associated
	// PKCE code verifier and redirect URI, and returns the state string.
	Create(ctx context.Context, codeVerifier, redirectURI string) (state string, err error)

	// Verify retrieves and deletes the state payload. Returns an error if the state
	// does not exist or has expired (single-use enforcement).
	Verify(ctx context.Context, state string) (*StatePayload, error)
}

type redisStateStore struct {
	rdb *redis.Client
}

// NewStateStore creates a new Redis-backed OAuth state store.
func NewStateStore(rdb *redis.Client) StateStore {
	return &redisStateStore{rdb: rdb}
}

func (s *redisStateStore) Create(ctx context.Context, codeVerifier, redirectURI string) (string, error) {
	state, err := GenerateState()
	if err != nil {
		return "", err
	}

	payload := StatePayload{
		CodeVerifier: codeVerifier,
		RedirectURI:  redirectURI,
		CreatedAt:    time.Now().UTC(),
	}
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("stateStore.Create: marshal: %w", err)
	}

	key := StatePrefix + state
	if err := s.rdb.Set(ctx, key, jsonData, StateTTL).Err(); err != nil {
		return "", fmt.Errorf("stateStore.Create: redis set: %w", err)
	}

	return state, nil
}

func (s *redisStateStore) Verify(ctx context.Context, state string) (*StatePayload, error) {
	key := StatePrefix + state

	// GetDel atomically retrieves and deletes — single-use enforcement.
	jsonData, err := s.rdb.GetDel(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, errors.New("oauth state not found or expired (CSRF protection)")
		}
		return nil, fmt.Errorf("stateStore.Verify: redis getdel: %w", err)
	}

	var payload StatePayload
	if err := json.Unmarshal(jsonData, &payload); err != nil {
		return nil, fmt.Errorf("stateStore.Verify: unmarshal: %w", err)
	}

	return &payload, nil
}

// GenerateState creates a 32-byte cryptographically random, base64url-encoded state token.
func GenerateState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("oauth.GenerateState: %w", err)
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// GeneratePKCE generates a PKCE code verifier and its S256 challenge.
// The verifier is 86 characters (64 random bytes, base64url-encoded without padding).
func GeneratePKCE() (verifier, challenge string, err error) {
	b := make([]byte, 64)
	if _, err = rand.Read(b); err != nil {
		return "", "", fmt.Errorf("oauth.GeneratePKCE: %w", err)
	}
	verifier = base64.RawURLEncoding.EncodeToString(b)
	h := sha256.Sum256([]byte(verifier))
	challenge = base64.RawURLEncoding.EncodeToString(h[:])
	return verifier, challenge, nil
}
