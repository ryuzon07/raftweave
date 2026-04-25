package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/raftweave/raftweave/internal/ingestion/domain"
)

type CredentialRepository struct {
	pool   *pgxpool.Pool
	tracer trace.Tracer
}

func NewCredentialRepository(pool *pgxpool.Pool, tracer trace.Tracer) *CredentialRepository {
	return &CredentialRepository{pool: pool, tracer: tracer}
}

func (r *CredentialRepository) Save(ctx context.Context, c *domain.CloudCredential) error {
	ctx, span := r.tracer.Start(ctx, "postgres.CredentialRepository.Save")
	defer span.End()
	span.SetAttributes(attribute.String("db.operation", "INSERT"))

	_, err := r.pool.Exec(ctx, queryInsertCredential,
		c.ID, c.WorkloadID, c.Provider, c.Type, c.EncryptedPayload, c.KeyVersion, c.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("postgres.CredentialRepository.Save: %w", err)
	}

	return nil
}

func (r *CredentialRepository) FindByWorkloadAndProvider(ctx context.Context, workloadID domain.WorkloadID, provider domain.CloudProvider) (*domain.CloudCredential, error) {
	ctx, span := r.tracer.Start(ctx, "postgres.CredentialRepository.FindByWorkloadAndProvider")
	defer span.End()
	span.SetAttributes(attribute.String("db.operation", "SELECT"))

	var c domain.CloudCredential
	err := r.pool.QueryRow(ctx, queryFindCredentialByWorkloadAndProvider, workloadID, provider).Scan(
		&c.ID, &c.WorkloadID, &c.Provider, &c.Type, &c.EncryptedPayload, &c.KeyVersion, &c.CreatedAt, &c.RotatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrCredentialNotFound
		}
		return nil, fmt.Errorf("postgres.CredentialRepository.FindByWorkloadAndProvider: %w", err)
	}

	return &c, nil
}

func (r *CredentialRepository) FindAllByWorkload(ctx context.Context, workloadID domain.WorkloadID) ([]*domain.CloudCredential, error) {
	return nil, fmt.Errorf("not implemented")
}

func (r *CredentialRepository) Delete(ctx context.Context, id domain.CredentialID) error {
	ctx, span := r.tracer.Start(ctx, "postgres.CredentialRepository.Delete")
	defer span.End()
	span.SetAttributes(attribute.String("db.operation", "DELETE"))

	_, err := r.pool.Exec(ctx, queryDeleteCredential, id)
	if err != nil {
		return fmt.Errorf("postgres.CredentialRepository.Delete: %w", err)
	}
	return nil
}
