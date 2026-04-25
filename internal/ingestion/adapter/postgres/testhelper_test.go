//go:build integration
// +build integration

package postgres_test

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func setupPostgres(t *testing.T) *pgxpool.Pool {
	ctx := context.Background()

	container, err := postgres.RunContainer(ctx,
		testcontainers.WithImage("docker.io/postgres:16-alpine"),
		postgres.WithDatabase("raftweave_test"),
		postgres.WithUsername("raftweave"),
		postgres.WithPassword("test_password"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)
	require.NoError(t, err)

	t.Cleanup(func() {
		_ = container.Terminate(ctx)
	})

	connStr, err := container.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	runMigrations(t, connStr)

	pool, err := pgxpool.New(ctx, connStr)
	require.NoError(t, err)

	return pool
}

func runMigrations(t *testing.T, connStr string) {
	// Skip migrations for rapid dev, assume tables exist via simple exec
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, connStr)
	require.NoError(t, err)
	defer pool.Close()

	schema := `
	CREATE TABLE IF NOT EXISTS workloads (
		id VARCHAR(36) PRIMARY KEY,
		name VARCHAR(255) NOT NULL UNIQUE,
		descriptor_json JSONB NOT NULL,
		status VARCHAR(50) NOT NULL,
		primary_region VARCHAR(100) NOT NULL,
		primary_provider VARCHAR(50) NOT NULL,
		created_at TIMESTAMP NOT NULL,
		updated_at TIMESTAMP NOT NULL
	);
	CREATE TABLE IF NOT EXISTS credentials (
		id VARCHAR(36) PRIMARY KEY,
		workload_id VARCHAR(36) NOT NULL REFERENCES workloads(id) ON DELETE CASCADE,
		provider VARCHAR(50) NOT NULL,
		credential_type VARCHAR(50) NOT NULL,
		encrypted_payload BYTEA NOT NULL,
		key_version VARCHAR(50) NOT NULL,
		created_at TIMESTAMP NOT NULL,
		rotated_at TIMESTAMP
	);
	`
	_, err = pool.Exec(ctx, schema)
	require.NoError(t, err)
}
