package domain

import "context"

// BuildRepository is the persistence contract for build metadata.
// Implementations live in the adapter/postgres package.
type BuildRepository interface {
	// Create persists a new build record.
	Create(ctx context.Context, b *Build) error
	// GetByID retrieves a build by its unique identifier.
	GetByID(ctx context.Context, id string) (*Build, error)
	// ListByWorkload returns a paginated list of builds for a specific workload.
	ListByWorkload(ctx context.Context, workloadID string, limit, offset int) ([]*Build, error)
	// UpdateStatus updates the lifecycle state and optional error message of a build.
	UpdateStatus(ctx context.Context, id string, status BuildStatus, errMsg string) error
	// UpdateImageRef populates the registry reference and digest after a successful push.
	UpdateImageRef(ctx context.Context, id, imageRef, imageDigest string, sizeBytes int64) error
	// MarkCompleted sets the completion timestamp and terminal status.
	MarkCompleted(ctx context.Context, id string) error
}

// LogRepository stores and retrieves streaming log lines for a build.
type LogRepository interface {
	// AppendLine persists a single log line to the storage backend.
	AppendLine(ctx context.Context, line *LogLine) error
	// GetLines retrieves all log lines for a build starting from a specific sequence number.
	GetLines(ctx context.Context, buildID string, fromSeq int64) ([]*LogLine, error)
}
