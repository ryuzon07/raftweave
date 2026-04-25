// Package store provides the data access layer backed by PostgreSQL.
package store

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Store is the interface for database operations.
type Store interface {
	// Pool returns the underlying connection pool.
	Pool() *pgxpool.Pool
	// Close closes all connections.
	Close()
	// Ping verifies the database connection is alive.
	Ping(ctx context.Context) error
}

// PGStore implements Store using pgx.
type PGStore struct {
	pool *pgxpool.Pool
}

// NewPGStore creates a new PostgreSQL store.
func NewPGStore(pool *pgxpool.Pool) *PGStore {
	return &PGStore{pool: pool}
}

func (s *PGStore) Pool() *pgxpool.Pool {
	return s.pool
}

func (s *PGStore) Close() {
	s.pool.Close()
}

func (s *PGStore) Ping(ctx context.Context) error {
	return s.pool.Ping(ctx)
}
