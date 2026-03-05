// Package ingestion handles workload submission, credential management,
// webhook processing, and job enqueuing.
package ingestion

import "context"

// Service defines the ingestion system's business logic.
type Service interface {
	SubmitWorkload(ctx context.Context, req *SubmitWorkloadInput) (*SubmitWorkloadOutput, error)
	AddCredentials(ctx context.Context, req *AddCredentialsInput) (*AddCredentialsOutput, error)
	GetWorkloadStatus(ctx context.Context, name string) (*WorkloadStatusOutput, error)
	ListWorkloads(ctx context.Context, req *ListWorkloadsInput) (*ListWorkloadsOutput, error)
}
