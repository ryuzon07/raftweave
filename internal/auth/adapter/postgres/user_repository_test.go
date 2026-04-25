package postgres

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/raftweave/raftweave/internal/auth/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

func setupPostgres(t *testing.T) *pgxpool.Pool {
	t.Helper()
	ctx := context.Background()

	container, err := postgres.Run(ctx, "postgres:16-alpine",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("test"),
		postgres.WithPassword("test"),
	)
	if err != nil && (strings.Contains(err.Error(), "Docker is not supported") || strings.Contains(err.Error(), "failed to create Docker provider")) {
		t.Skipf("Skipping integration test: %v", err)
		return nil
	}
	require.NoError(t, err)
	t.Cleanup(func() { _ = container.Terminate(ctx) })

	connStr, err := container.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	pool, err := pgxpool.New(ctx, connStr)
	require.NoError(t, err)
	t.Cleanup(func() { pool.Close() })

	// Run auth migration.
	_, err = pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS users (
			id TEXT PRIMARY KEY, email TEXT NOT NULL, name TEXT NOT NULL,
			avatar_url TEXT, provider TEXT NOT NULL, provider_id TEXT NOT NULL,
			github_login TEXT, github_token_enc BYTEA,
			is_email_verified BOOLEAN NOT NULL DEFAULT FALSE,
			is_active BOOLEAN NOT NULL DEFAULT TRUE,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			last_login_at TIMESTAMPTZ,
			UNIQUE (email)
		);
		CREATE UNIQUE INDEX IF NOT EXISTS idx_users_provider ON users(provider, provider_id);
	`)
	require.NoError(t, err)

	return pool
}

func newTestUser() *domain.User {
	now := time.Now().UTC().Truncate(time.Microsecond)
	return &domain.User{
		ID: "user-1", Email: "test@raftweave.io", Name: "Test User",
		AvatarURL: "https://example.com/avatar.png",
		Provider: domain.ProviderGitHub, ProviderID: "gh-12345",
		GitHubLogin: "testuser", IsEmailVerified: true, IsActive: true,
		CreatedAt: now, UpdatedAt: now,
	}
}

func TestCreate_Success(t *testing.T) {
	if testing.Short() { t.Skip("skipping integration test") }
	pool := setupPostgres(t)
	repo := NewUserRepository(pool)

	err := repo.Create(context.Background(), newTestUser())
	require.NoError(t, err)

	user, err := repo.GetByID(context.Background(), "user-1")
	require.NoError(t, err)
	assert.Equal(t, "test@raftweave.io", user.Email)
	assert.Equal(t, "testuser", user.GitHubLogin)
}

func TestCreate_DuplicateEmail_ReturnsErrUserAlreadyExists(t *testing.T) {
	if testing.Short() { t.Skip("skipping integration test") }
	pool := setupPostgres(t)
	repo := NewUserRepository(pool)

	u1 := newTestUser()
	require.NoError(t, repo.Create(context.Background(), u1))

	u2 := newTestUser()
	u2.ID = "user-2"
	u2.ProviderID = "gh-99999"
	err := repo.Create(context.Background(), u2)
	assert.ErrorIs(t, err, domain.ErrUserAlreadyExists)
}

func TestGetByEmail_CaseInsensitive(t *testing.T) {
	if testing.Short() { t.Skip("skipping integration test") }
	pool := setupPostgres(t)
	repo := NewUserRepository(pool)

	require.NoError(t, repo.Create(context.Background(), newTestUser()))

	user, err := repo.GetByEmail(context.Background(), "TEST@RAFTWEAVE.IO")
	require.NoError(t, err)
	assert.Equal(t, "user-1", user.ID)

	user, err = repo.GetByEmail(context.Background(), "Test@Raftweave.IO")
	require.NoError(t, err)
	assert.Equal(t, "user-1", user.ID)
}

func TestGetByProviderID_NotFound_ReturnsErrUserNotFound(t *testing.T) {
	if testing.Short() { t.Skip("skipping integration test") }
	pool := setupPostgres(t)
	repo := NewUserRepository(pool)

	_, err := repo.GetByProviderID(context.Background(), domain.ProviderGitHub, "nonexistent")
	assert.ErrorIs(t, err, domain.ErrUserNotFound)
}

func TestUpdateGitHubToken_Success(t *testing.T) {
	if testing.Short() { t.Skip("skipping integration test") }
	pool := setupPostgres(t)
	repo := NewUserRepository(pool)

	require.NoError(t, repo.Create(context.Background(), newTestUser()))

	encToken := []byte{0x01, 0x02, 0x03, 0x04}
	err := repo.UpdateGitHubToken(context.Background(), "user-1", encToken)
	require.NoError(t, err)

	user, err := repo.GetByID(context.Background(), "user-1")
	require.NoError(t, err)
	assert.Equal(t, encToken, user.GitHubTokenEnc)
}

func TestSoftDelete_SetsIsActiveFalse(t *testing.T) {
	if testing.Short() { t.Skip("skipping integration test") }
	pool := setupPostgres(t)
	repo := NewUserRepository(pool)

	require.NoError(t, repo.Create(context.Background(), newTestUser()))
	require.NoError(t, repo.SoftDelete(context.Background(), "user-1"))

	// GetByID filters by is_active=TRUE, so should return not found.
	_, err := repo.GetByID(context.Background(), "user-1")
	assert.ErrorIs(t, err, domain.ErrUserNotFound)
}
