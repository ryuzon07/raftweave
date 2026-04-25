package ingestion

import "time"

// Workload is the internal domain representation.
type Workload struct {
	ID            string
	Name          string
	Descriptor    []byte // JSON-encoded workload descriptor
	Status        string
	PrimaryRegion string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// Credential holds encrypted cloud provider credentials.
type Credential struct {
	ID               string
	WorkloadID       string
	Provider         string
	EncryptedPayload []byte
	CreatedAt        time.Time
}

// SubmitWorkloadInput is the service-layer input for workload submission.
type SubmitWorkloadInput struct {
	Name           string
	DescriptorJSON []byte
	PrimaryRegion  string
}

// SubmitWorkloadOutput is the service-layer output for workload submission.
type SubmitWorkloadOutput struct {
	WorkloadID string
	Status     string
}

// AddCredentialsInput is the service-layer input for credential addition.
type AddCredentialsInput struct {
	WorkloadID       string
	Provider         string
	EncryptedPayload []byte
}

// AddCredentialsOutput is the service-layer output for credential addition.
type AddCredentialsOutput struct {
	CredentialID string
}

// WorkloadStatusOutput is the service-layer output for workload status.
type WorkloadStatusOutput struct {
	Name          string
	Status        string
	PrimaryRegion string
}

// ListWorkloadsInput is the service-layer input for listing workloads.
type ListWorkloadsInput struct {
	PageSize  int
	PageToken string
	Labels    map[string]string
}

// ListWorkloadsOutput is the service-layer output for listing workloads.
type ListWorkloadsOutput struct {
	Workloads     []*WorkloadStatusOutput
	NextPageToken string
	TotalCount    int
}

// ListOptions holds query options for listing workloads.
type ListOptions struct {
	PageSize int
	Offset   int
	Labels   map[string]string
}
