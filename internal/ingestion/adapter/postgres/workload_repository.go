package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/raftweave/raftweave/internal/ingestion/domain"
)

type WorkloadRepository struct {
	pool   *pgxpool.Pool
	tracer trace.Tracer
}

func NewWorkloadRepository(pool *pgxpool.Pool, tracer trace.Tracer) *WorkloadRepository {
	return &WorkloadRepository{pool: pool, tracer: tracer}
}

func (r *WorkloadRepository) Save(ctx context.Context, w *domain.Workload) error {
	ctx, span := r.tracer.Start(ctx, "postgres.WorkloadRepository.Save")
	defer span.End()
	span.SetAttributes(attribute.String("db.operation", "INSERT"))

	descriptorJSON, err := json.Marshal(w)
	if err != nil {
		return fmt.Errorf("postgres.WorkloadRepository.Save: marshal: %w", err)
	}

	_, err = r.pool.Exec(ctx, queryInsertWorkload,
		w.ID, w.Name, descriptorJSON, w.Status, w.PrimaryRegion.Name, w.PrimaryRegion.Provider, w.CreatedAt, w.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("postgres.WorkloadRepository.Save: %w", err)
	}

	return nil
}

func (r *WorkloadRepository) FindByID(ctx context.Context, id domain.WorkloadID) (*domain.Workload, error) {
	ctx, span := r.tracer.Start(ctx, "postgres.WorkloadRepository.FindByID")
	defer span.End()
	span.SetAttributes(attribute.String("db.operation", "SELECT"))

	var w domain.Workload
	var descriptorJSON []byte
	var primaryRegionName, primaryProvider string

	err := r.pool.QueryRow(ctx, queryFindWorkloadByID, id).Scan(
		&w.ID, &w.Name, &descriptorJSON, &w.Status, &primaryRegionName, &primaryProvider, &w.CreatedAt, &w.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrWorkloadNotFound
		}
		return nil, fmt.Errorf("postgres.WorkloadRepository.FindByID: %w", err)
	}

	if err := json.Unmarshal(descriptorJSON, &w); err != nil {
		return nil, fmt.Errorf("postgres.WorkloadRepository.FindByID: unmarshal: %w", err)
	}

	return &w, nil
}

func (r *WorkloadRepository) FindByName(ctx context.Context, name string) (*domain.Workload, error) {
	ctx, span := r.tracer.Start(ctx, "postgres.WorkloadRepository.FindByName")
	defer span.End()
	span.SetAttributes(attribute.String("db.operation", "SELECT"))

	var w domain.Workload
	var descriptorJSON []byte
	var primaryRegionName, primaryProvider string

	err := r.pool.QueryRow(ctx, queryFindWorkloadByName, name).Scan(
		&w.ID, &w.Name, &descriptorJSON, &w.Status, &primaryRegionName, &primaryProvider, &w.CreatedAt, &w.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrWorkloadNotFound
		}
		return nil, fmt.Errorf("postgres.WorkloadRepository.FindByName: %w", err)
	}

	if err := json.Unmarshal(descriptorJSON, &w); err != nil {
		return nil, fmt.Errorf("postgres.WorkloadRepository.FindByName: unmarshal: %w", err)
	}

	return &w, nil
}

func (r *WorkloadRepository) UpdateStatus(ctx context.Context, id domain.WorkloadID, status domain.WorkloadStatus) error {
	ctx, span := r.tracer.Start(ctx, "postgres.WorkloadRepository.UpdateStatus")
	defer span.End()
	span.SetAttributes(attribute.String("db.operation", "UPDATE"))

	tag, err := r.pool.Exec(ctx, queryUpdateWorkloadStatus, status, id)
	if err != nil {
		return fmt.Errorf("postgres.WorkloadRepository.UpdateStatus: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrWorkloadNotFound
	}

	return nil
}

func (r *WorkloadRepository) FindAll(ctx context.Context, limit, offset int) ([]*domain.Workload, error) {
	return nil, fmt.Errorf("not implemented")
}

func (r *WorkloadRepository) Delete(ctx context.Context, id domain.WorkloadID) error {
	ctx, span := r.tracer.Start(ctx, "postgres.WorkloadRepository.Delete")
	defer span.End()
	span.SetAttributes(attribute.String("db.operation", "DELETE"))

	_, err := r.pool.Exec(ctx, queryDeleteWorkload, id)
	if err != nil {
		return fmt.Errorf("postgres.WorkloadRepository.Delete: %w", err)
	}
	return nil
}
