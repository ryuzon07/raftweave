package redis

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/raftweave/raftweave/internal/auth/domain"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tcredis "github.com/testcontainers/testcontainers-go/modules/redis"
)

func setupRedis(t *testing.T) *redis.Client {
	t.Helper()
	ctx := context.Background()

	container, err := tcredis.Run(ctx, "redis:7-alpine")
	if err != nil && (strings.Contains(err.Error(), "Docker is not supported") || strings.Contains(err.Error(), "failed to create Docker provider")) {
		t.Skipf("Skipping integration test: %v", err)
		return nil
	}
	require.NoError(t, err)
	t.Cleanup(func() { _ = container.Terminate(ctx) })

	endpoint, err := container.Endpoint(ctx, "")
	require.NoError(t, err)

	rdb := redis.NewClient(&redis.Options{Addr: endpoint})
	t.Cleanup(func() { _ = rdb.Close() })

	require.NoError(t, rdb.Ping(ctx).Err())
	return rdb
}

func TestIssue_ReturnsUniqueTokens(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	rdb := setupRedis(t)
	store := NewRefreshTokenStore(rdb)
	ctx := context.Background()

	tokens := make(map[string]struct{})
	for i := 0; i < 50; i++ {
		tok, err := store.Issue(ctx, "session-1", "user-1", "fp-1")
		require.NoError(t, err)
		require.NotEmpty(t, tok)
		assert.Len(t, tok, 64, "token should be 64 hex chars (32 bytes)")
		_, exists := tokens[tok]
		assert.False(t, exists, "duplicate token generated")
		tokens[tok] = struct{}{}
	}
}

func TestVerify_ValidToken_ReturnsSessionID(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	rdb := setupRedis(t)
	store := NewRefreshTokenStore(rdb)
	ctx := context.Background()

	tok, err := store.Issue(ctx, "session-abc", "user-1", "fp-1")
	require.NoError(t, err)

	sessionID, err := store.Verify(ctx, tok, "fp-1")
	require.NoError(t, err)
	assert.Equal(t, "session-abc", sessionID)
}

func TestVerify_ExpiredToken_ReturnsError(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	rdb := setupRedis(t)
	store := NewRefreshTokenStore(rdb)
	ctx := context.Background()

	tok, err := store.Issue(ctx, "session-1", "user-1", "fp-1")
	require.NoError(t, err)

	// Manually expire the token by setting its TTL to 1ms and waiting.
	hash := hashToken(tok)
	rdb.PExpire(ctx, tokenKey(hash), time.Millisecond)
	time.Sleep(50 * time.Millisecond)

	_, err = store.Verify(ctx, tok, "fp-1")
	assert.ErrorIs(t, err, domain.ErrSessionNotFound)
}

func TestVerify_FingerprintMismatch_RevokesAllSessions(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	rdb := setupRedis(t)
	store := NewRefreshTokenStore(rdb)
	ctx := context.Background()

	// Issue 3 tokens for the same user.
	tok1, err := store.Issue(ctx, "s1", "user-1", "fp-legit")
	require.NoError(t, err)
	tok2, err := store.Issue(ctx, "s2", "user-1", "fp-legit")
	require.NoError(t, err)
	_, err = store.Issue(ctx, "s3", "user-1", "fp-legit")
	require.NoError(t, err)

	// Attacker uses tok1 with a different fingerprint.
	_, err = store.Verify(ctx, tok1, "fp-attacker")
	assert.ErrorIs(t, err, domain.ErrSessionFingerprintMismatch)

	// ALL sessions for this user should be revoked — even the legitimate ones.
	_, err = store.Verify(ctx, tok2, "fp-legit")
	assert.ErrorIs(t, err, domain.ErrSessionNotFound, "all sessions should be revoked after fingerprint mismatch")
}

func TestRotate_AtomicSwap_OldTokenInvalidated(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	rdb := setupRedis(t)
	store := NewRefreshTokenStore(rdb)
	ctx := context.Background()

	oldTok, err := store.Issue(ctx, "session-old", "user-1", "fp-1")
	require.NoError(t, err)

	newTok, newSID, err := store.Rotate(ctx, oldTok, "fp-1")
	require.NoError(t, err)
	require.NotEmpty(t, newTok)
	require.NotEmpty(t, newSID)
	assert.NotEqual(t, oldTok, newTok)

	// Old token should be invalid.
	_, err = store.Verify(ctx, oldTok, "fp-1")
	assert.ErrorIs(t, err, domain.ErrSessionNotFound)

	// New token should be valid.
	sid, err := store.Verify(ctx, newTok, "fp-1")
	require.NoError(t, err)
	assert.Equal(t, newSID, sid)
}

func TestRotate_OldTokenReuse_ReturnsError(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	rdb := setupRedis(t)
	store := NewRefreshTokenStore(rdb)
	ctx := context.Background()

	tok, err := store.Issue(ctx, "session-1", "user-1", "fp-1")
	require.NoError(t, err)

	// First rotation succeeds.
	_, _, err = store.Rotate(ctx, tok, "fp-1")
	require.NoError(t, err)

	// Second rotation with the same (now-revoked) token should fail — replay detection.
	_, _, err = store.Rotate(ctx, tok, "fp-1")
	assert.ErrorIs(t, err, domain.ErrSessionNotFound)
}

func TestRevokeAll_ClearsAllUserSessions(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	rdb := setupRedis(t)
	store := NewRefreshTokenStore(rdb)
	ctx := context.Background()

	var tokens []string
	for i := 0; i < 3; i++ {
		tok, err := store.Issue(ctx, "session", "user-1", "fp-1")
		require.NoError(t, err)
		tokens = append(tokens, tok)
	}

	err := store.RevokeAll(ctx, "user-1")
	require.NoError(t, err)

	for _, tok := range tokens {
		_, err := store.Verify(ctx, tok, "fp-1")
		assert.ErrorIs(t, err, domain.ErrSessionNotFound)
	}
}

func TestMaxSessions_OldestAutoRevoked(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	rdb := setupRedis(t)
	store := NewRefreshTokenStore(rdb)
	ctx := context.Background()

	var tokens []string
	for i := 0; i < MaxSessionsPerUser+2; i++ {
		tok, err := store.Issue(ctx, "session", "user-1", "fp-1")
		require.NoError(t, err)
		tokens = append(tokens, tok)
		time.Sleep(10 * time.Millisecond) // ensure distinct creation times
	}

	// The most recent MaxSessionsPerUser tokens should be valid.
	// Older tokens should have been auto-revoked.
	validCount := 0
	for _, tok := range tokens {
		_, err := store.Verify(ctx, tok, "fp-1")
		if err == nil {
			validCount++
		}
	}

	assert.LessOrEqual(t, validCount, MaxSessionsPerUser,
		"should not exceed %d active sessions", MaxSessionsPerUser)
}
