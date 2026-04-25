package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/hibiken/asynq"
	"github.com/raftweave/raftweave/internal/build/domain"
	"github.com/raftweave/raftweave/internal/build/usecase"
)

const TaskTypeBuild = "build:execute"

// Handler implements asynq.Handler for build jobs.
type Handler struct {
	uc *usecase.BuildUseCase
}

func NewHandler(uc *usecase.BuildUseCase) *Handler {
	return &Handler{uc: uc}
}

func (h *Handler) ProcessTask(ctx context.Context, t *asynq.Task) error {
	// Simulated deserialization
	var job usecase.BuildJob
	if err := json.Unmarshal(t.Payload(), &job); err != nil {
		return fmt.Errorf("could not unmarshal job: %v: %w", err, asynq.SkipRetry)
	}

	err := h.uc.Execute(ctx, &job)
	if err != nil {
		// Map domain errors to Asynq directives
		if domain.ErrDetectionFailed.Error() == err.Error() || domain.ErrDockerfileInvalid.Error() == err.Error() {
			return fmt.Errorf("fatal user error, skipping retry: %w: %w", err, asynq.SkipRetry)
		}
		if domain.ErrKanikoJobFailed.Error() == err.Error() {
			return fmt.Errorf("kaniko job failed, skipping retry: %w: %w", err, asynq.SkipRetry)
		}
		return fmt.Errorf("transient build error, retrying: %w", err)
	}

	return nil
}

// StartWorker starts the Asynq server with the build handler registered.
func StartWorker(redisAddr string, handler *Handler) error {
	srv := asynq.NewServer(
		asynq.RedisClientOpt{Addr: redisAddr},
		asynq.Config{
			Concurrency:    5,
			Queues:         map[string]int{"builds:high": 6, "builds:default": 3, "builds:low": 1},
			StrictPriority: true,
			RetryDelayFunc: asynq.DefaultRetryDelayFunc,
			IsFailure: func(err error) bool {
				return err != asynq.SkipRetry
			},
		},
	)

	mux := asynq.NewServeMux()
	mux.Handle(TaskTypeBuild, handler)

	log.Printf("Starting build worker listening on %s...", redisAddr)
	return srv.Run(mux)
}
