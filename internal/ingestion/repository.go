package ingestion

import "context"

// Repository defines the data access layer for ingestion.
type Repository interface {
	CreateWorkload(ctx context.Context, workload *Workload) error
	GetWorkload(ctx context.Context, name string) (*Workload, error)
	ListWorkloads(ctx context.Context, opts ListOptions) ([]*Workload, int, error)
	UpdateWorkloadStatus(ctx context.Context, name string, status string) error
	StoreCredentials(ctx context.Context, cred *Credential) error
	GetCredentials(ctx context.Context, workloadID string) ([]*Credential, error)
}
