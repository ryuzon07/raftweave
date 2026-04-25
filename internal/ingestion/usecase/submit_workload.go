package usecase

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/raftweave/raftweave/internal/ingestion/domain"
)

type SubmitWorkloadUseCase struct {
	deps Dependencies
}

type SubmitWorkloadInput struct {
	Name           string
	Source         domain.SourceConfig
	PrimaryRegion  domain.Region
	StandbyRegions []domain.Region
	Compute        domain.ResourceSpec
	Database       domain.DatabaseSpec
	Failover       domain.FailoverConfig
	Compliance     domain.ComplianceConfig
}

type SubmitWorkloadOutput struct {
	WorkloadID domain.WorkloadID
	Status     domain.WorkloadStatus
}

func NewSubmitWorkloadUseCase(deps Dependencies) *SubmitWorkloadUseCase {
	return &SubmitWorkloadUseCase{deps: deps}
}

func (uc *SubmitWorkloadUseCase) Execute(ctx context.Context, input SubmitWorkloadInput) (*SubmitWorkloadOutput, error) {
	ctx, span := uc.deps.Tracer.Start(ctx, "ingestion.SubmitWorkload")
	defer span.End()

	w, err := domain.NewWorkload(input.Name, input.Source, input.PrimaryRegion, input.StandbyRegions, input.Compute, input.Database, input.Failover)
	if err != nil {
		return nil, err // Domain errors directly returned
	}

	_, err = uc.deps.WorkloadRepo.FindByName(ctx, input.Name)
	if err == nil {
		return nil, domain.ErrWorkloadAlreadyExists
	}
	if err != domain.ErrWorkloadNotFound {
		return nil, fmt.Errorf("usecase.SubmitWorkload: FindByName: %w", err)
	}

	if err := uc.deps.WorkloadRepo.Save(ctx, w); err != nil {
		return nil, fmt.Errorf("usecase.SubmitWorkload: Save: %w", err)
	}

	if err := uc.deps.JobEnqueuer.EnqueueProvisionJob(ctx, w.ID); err != nil {
		_ = uc.deps.WorkloadRepo.Delete(ctx, w.ID) // rollback
		return nil, fmt.Errorf("usecase.SubmitWorkload: Enqueue: %w", err)
	}

	uc.deps.Metrics.WorkloadsSubmitted.Add(ctx, 1)

	uc.deps.Logger.Info("workload_submitted",
		zap.String("workload_id", string(w.ID)),
		zap.String("workload_name", w.Name),
		zap.String("primary_region", w.PrimaryRegion.Name),
		zap.String("component", "ingestion"),
		zap.String("trace_id", span.SpanContext().TraceID().String()),
		zap.String("span_id", span.SpanContext().SpanID().String()),
	)

	return &SubmitWorkloadOutput{
		WorkloadID: w.ID,
		Status:     w.Status,
	}, nil
}
