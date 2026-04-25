//go:build integration
// +build integration

package postgres_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace/noop"

	"github.com/raftweave/raftweave/internal/ingestion/adapter/postgres"
	"github.com/raftweave/raftweave/internal/ingestion/domain"
)

func TestWorkloadRepository_Save_Success(t *testing.T) {
	pool := setupPostgres(t)
	repo := postgres.NewWorkloadRepository(pool, noop.NewTracerProvider().Tracer("test"))

	w := createTestWorkload()

	err := repo.Save(context.Background(), w)
	require.NoError(t, err)

	// Verify
	fetched, err := repo.FindByID(context.Background(), w.ID)
	require.NoError(t, err)
	assert.Equal(t, w.Name, fetched.Name)
	assert.Equal(t, w.PrimaryRegion.Name, fetched.PrimaryRegion.Name)
}

func TestWorkloadRepository_Save_DuplicateName(t *testing.T) {
	pool := setupPostgres(t)
	repo := postgres.NewWorkloadRepository(pool, noop.NewTracerProvider().Tracer("test"))

	w1 := createTestWorkload()
	err := repo.Save(context.Background(), w1)
	require.NoError(t, err)

	w2 := createTestWorkload()
	w2.Name = w1.Name // duplicate name
	
	err = repo.Save(context.Background(), w2)
	assert.Error(t, err)
	// In a complete implementation we might parse the pgerr to map to domain.ErrWorkloadAlreadyExists
}

func TestWorkloadRepository_FindByID_NotFound(t *testing.T) {
	pool := setupPostgres(t)
	repo := postgres.NewWorkloadRepository(pool, noop.NewTracerProvider().Tracer("test"))

	_, err := repo.FindByID(context.Background(), "non-existent")
	assert.ErrorIs(t, err, domain.ErrWorkloadNotFound)
}

func createTestWorkload() *domain.Workload {
	now := time.Now().UTC()
	return &domain.Workload{
		ID:     domain.WorkloadID("test-id-1234"),
		Name:   "test-workload",
		Status: domain.WorkloadStatusPending,
		Source: domain.SourceConfig{
			Type:    "git",
			RepoURL: "https://github.com/test",
			Branch:  "main",
		},
		PrimaryRegion: domain.Region{Name: "us-east-1", Provider: domain.CloudProviderAWS},
		StandbyRegions: []domain.Region{
			{Name: "us-west-2", Provider: domain.CloudProviderAWS},
		},
		CreatedAt: now,
		UpdatedAt: now,
	}
}