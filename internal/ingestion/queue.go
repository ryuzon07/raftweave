package ingestion

import "context"

// Queue defines the interface for enqueuing async jobs.
type Queue interface {
	EnqueueBuildJob(ctx context.Context, workloadName string, repoURL string, commitSHA string) error
	EnqueueProvisionJob(ctx context.Context, workloadName string, provider string, region string) error
}
