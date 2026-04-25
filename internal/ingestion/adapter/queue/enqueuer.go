package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hibiken/asynq"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"

	"github.com/raftweave/raftweave/internal/ingestion/domain"
)

type AsynqJobEnqueuer struct {
	client *asynq.Client
	tracer trace.Tracer
	logger *zap.Logger
}

func NewAsynqJobEnqueuer(redisOpt asynq.RedisClientOpt, tracer trace.Tracer, logger *zap.Logger) *AsynqJobEnqueuer {
	return &AsynqJobEnqueuer{
		client: asynq.NewClient(redisOpt),
		tracer: tracer,
		logger: logger,
	}
}

func (e *AsynqJobEnqueuer) EnqueueBuildJob(ctx context.Context, workloadID domain.WorkloadID, event *domain.WebhookEvent) error {
	ctx, span := e.tracer.Start(ctx, "queue.EnqueueBuildJob")
	defer span.End()

	payload := BuildJobPayload{
		WorkloadID:   string(workloadID),
		WorkloadName: event.WorkloadName,
		RepoURL:      event.RepoURL,
		Branch:       event.Branch,
		CommitSHA:    event.CommitSHA,
		CommitMsg:    event.CommitMsg,
		TriggeredAt:  event.TriggeredAt.String(),
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal BuildJob payload: %w", err)
	}

	task := asynq.NewTask(TaskTypeBuild, payloadBytes, BuildJobOptions...)
	info, err := e.client.EnqueueContext(ctx, task)
	if err != nil {
		return fmt.Errorf("queue.EnqueueBuildJob failed: %w", err)
	}

	e.logger.Info("enqueued build job",
		zap.String("task_id", info.ID),
		zap.String("queue", info.Queue),
		zap.String("task_type", TaskTypeBuild),
		zap.String("workload_id", string(workloadID)),
		zap.String("trace_id", span.SpanContext().TraceID().String()),
		zap.String("span_id", span.SpanContext().SpanID().String()),
	)

	return nil
}

func (e *AsynqJobEnqueuer) EnqueueProvisionJob(ctx context.Context, workloadID domain.WorkloadID) error {
	ctx, span := e.tracer.Start(ctx, "queue.EnqueueProvisionJob")
	defer span.End()

	payload := ProvisionJobPayload{
		WorkloadID:   string(workloadID),
		WorkloadName: "", // omitted for simplicity in this checkpoint, usually we'd have it or fetch it
		Action:       "PROVISION",
		TriggeredAt:  time.Now().String(),
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal ProvisionJob payload: %w", err)
	}

	task := asynq.NewTask(TaskTypeProvision, payloadBytes, ProvisionJobOptions...)
	info, err := e.client.EnqueueContext(ctx, task)
	if err != nil {
		return fmt.Errorf("queue.EnqueueProvisionJob failed: %w", err)
	}

	e.logger.Info("enqueued provision job",
		zap.String("task_id", info.ID),
		zap.String("queue", info.Queue),
		zap.String("task_type", TaskTypeProvision),
		zap.String("workload_id", string(workloadID)),
		zap.String("trace_id", span.SpanContext().TraceID().String()),
		zap.String("span_id", span.SpanContext().SpanID().String()),
	)

	return nil
}
