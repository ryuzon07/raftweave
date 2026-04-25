package di

import (
	"context"
	"net/http"

	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"

	"github.com/raftweave/raftweave/internal/gen/raftweave/v1/raftweavev1connect"
	"github.com/raftweave/raftweave/internal/ingestion"
	"github.com/raftweave/raftweave/internal/ingestion/adapter/crypto"
	"github.com/raftweave/raftweave/internal/ingestion/adapter/postgres"
	"github.com/raftweave/raftweave/internal/ingestion/adapter/queue"
	"github.com/raftweave/raftweave/internal/ingestion/handler"
	"github.com/raftweave/raftweave/internal/ingestion/usecase"
)

type Config struct {
	// Database and Message Queue connections
	DBPool *pgxpool.Pool
	Redis  asynq.RedisClientOpt

	// Cryptography settings
	CryptoKeys          map[string][]byte
	ActiveCryptoVersion string

	// Observability
	Tracer trace.Tracer
	Meter  metric.Meter
	Logger *zap.Logger

	// External Dependencies
	WebhookSecretStore usecase.WebhookSecretStore
}

type Module struct {
	// The gRPC / Connect-RPC HTTP mount path and handler (e.g. "/raftweave.v1.IngestionService/")
	RPCHandlerPath string
	RPCHandler     http.Handler

	// The RESTful webhook receiver handler
	WebhookHandler http.Handler
}

// Bootstrap wires the entire Ingestion clean architecture layer gracefully
func Bootstrap(ctx context.Context, cfg Config) (*Module, error) {
	// 1. Initialize Adapters
	encryptor, err := crypto.NewAESEncryptor(cfg.CryptoKeys, cfg.ActiveCryptoVersion)
	if err != nil {
		return nil, err
	}

	workloadRepo := postgres.NewWorkloadRepository(cfg.DBPool, cfg.Tracer)
	credentialRepo := postgres.NewCredentialRepository(cfg.DBPool, cfg.Tracer)
	jobEnqueuer := queue.NewAsynqJobEnqueuer(cfg.Redis, cfg.Tracer, cfg.Logger)

	// 2. Initialize Telemetry Instruments
	metrics := ingestion.NewIngestionMetrics(cfg.Meter)

	// 3. Construct Domain Dependencies Registry
	deps := usecase.Dependencies{
		WorkloadRepo:   workloadRepo,
		CredentialRepo: credentialRepo,
		JobEnqueuer:    jobEnqueuer,
		Encryptor:      encryptor,
		Tracer:         cfg.Tracer,
		Logger:         cfg.Logger,
		Metrics:        metrics,
	}

	// 4. Initialize UseCases
	submitWorkload := usecase.NewSubmitWorkloadUseCase(deps)
	addCredential := usecase.NewAddCredentialUseCase(deps)
	processWebhook := usecase.NewProcessWebhookUseCase(deps, cfg.WebhookSecretStore)

	// 5. Initialize Handlers (Presentation Layer)
	rpcHandler := handler.NewIngestionHandler(cfg.Logger, submitWorkload, addCredential, workloadRepo)
	webhookHandler := handler.NewWebhookHandler(cfg.Logger, processWebhook)

	// 6. Wrap Protobuf Connect Service
	rpcPath, rpcSvc := raftweavev1connect.NewIngestionServiceHandler(rpcHandler)

	return &Module{
		RPCHandlerPath: rpcPath,
		RPCHandler:     rpcSvc,
		WebhookHandler: webhookHandler,
	}, nil
}