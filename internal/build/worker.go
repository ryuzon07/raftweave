package build

import (
	"context"

	"go.opentelemetry.io/otel"
)

// Worker processes build jobs from the async queue.
type Worker struct {
	svc     Service
	builder Builder
	reg     Registry
}

// NewWorker creates a new build worker.
func NewWorker(svc Service, builder Builder, reg Registry) *Worker {
	return &Worker{
		svc:     svc,
		builder: builder,
		reg:     reg,
	}
}

// ProcessBuildJob processes a single build job from the queue.
func (w *Worker) ProcessBuildJob(ctx context.Context, jobID string) error {
	ctx, span := otel.Tracer("internal/build").Start(ctx, "build.Worker.ProcessBuildJob")
	defer span.End()

	_ = jobID
	return nil // stub — to be implemented
}
