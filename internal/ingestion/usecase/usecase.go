package usecase

import (
	"context"

	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"

	"github.com/raftweave/raftweave/internal/ingestion"
	"github.com/raftweave/raftweave/internal/ingestion/domain"
)

type Dependencies struct {
	WorkloadRepo   domain.WorkloadRepository
	CredentialRepo domain.CredentialRepository
	JobEnqueuer    domain.JobEnqueuer
	Encryptor      domain.Encryptor
	Tracer         trace.Tracer
	Logger         *zap.Logger
	Metrics        *ingestion.IngestionMetrics
}

type WebhookSecretStore interface {
	GetSecret(ctx context.Context, workloadName string) (string, error)
}
