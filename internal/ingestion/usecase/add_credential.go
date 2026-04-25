package usecase

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel/metric"
	"go.uber.org/zap"

	"github.com/raftweave/raftweave/internal/ingestion/domain"
)

type AddCredentialUseCase struct {
	deps Dependencies
}

type AddCredentialInput struct {
	WorkloadID domain.WorkloadID
	Provider   domain.CloudProvider
	CredType   domain.CredentialType
	RawPayload []byte
}

type AddCredentialOutput struct {
	CredentialID domain.CredentialID
}

func NewAddCredentialUseCase(deps Dependencies) *AddCredentialUseCase {
	return &AddCredentialUseCase{deps: deps}
}

func (uc *AddCredentialUseCase) Execute(ctx context.Context, input AddCredentialInput) (*AddCredentialOutput, error) {
	ctx, span := uc.deps.Tracer.Start(ctx, "ingestion.AddCredential")
	defer span.End()

	w, err := uc.deps.WorkloadRepo.FindByID(ctx, input.WorkloadID)
	if err != nil {
		return nil, fmt.Errorf("usecase.AddCredential: FindByID: %w", err)
	}

	c, err := domain.NewCloudCredential(w.ID, input.Provider, input.CredType, input.RawPayload, uc.deps.Encryptor)
	if err != nil {
		return nil, fmt.Errorf("usecase.AddCredential: domain.NewCloudCredential: %w", err)
	}

	if err := uc.deps.CredentialRepo.Save(ctx, c); err != nil {
		return nil, fmt.Errorf("usecase.AddCredential: Save: %w", err)
	}

	uc.deps.Metrics.CredentialsAdded.Add(ctx, 1, metric.WithAttributes())

	uc.deps.Logger.Info("credential_added",
		zap.String("workload_id", string(w.ID)),
		zap.String("workload_name", w.Name),
		zap.String("provider", string(c.Provider)),
		zap.String("component", "ingestion"),
		zap.String("trace_id", span.SpanContext().TraceID().String()),
		zap.String("span_id", span.SpanContext().SpanID().String()),
	)

	return &AddCredentialOutput{
		CredentialID: c.ID,
	}, nil
}
