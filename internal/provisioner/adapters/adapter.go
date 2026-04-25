// Package adapters defines the cloud adapter interface and
// provider-specific implementations.
package adapters

import "context"

// CloudAdapter is the interface that all cloud providers must implement.
type CloudAdapter interface {
	// ProvisionCompute creates compute resources in the cloud.
	ProvisionCompute(ctx context.Context, req *ComputeRequest) (*ComputeResult, error)
	// ProvisionDatabase creates a managed database in the cloud.
	ProvisionDatabase(ctx context.Context, req *DatabaseRequest) (*DatabaseResult, error)
	// DestroyCompute destroys compute resources.
	DestroyCompute(ctx context.Context, resourceID string) error
	// DestroyDatabase destroys a managed database.
	DestroyDatabase(ctx context.Context, resourceID string) error
	// GetHealth returns the health status of a provisioned resource.
	GetHealth(ctx context.Context, resourceID string) (*HealthResult, error)
	// Provider returns the cloud provider name.
	Provider() string
}

// ComputeRequest describes a compute resource to provision.
type ComputeRequest struct {
	WorkloadName string
	Region       string
	CPU          string
	Memory       string
	Replicas     int32
	ImageRef     string
	Port         int32
}

// ComputeResult describes a provisioned compute resource.
type ComputeResult struct {
	ResourceID string
	Endpoint   string
	Status     string
}

// DatabaseRequest describes a database to provision.
type DatabaseRequest struct {
	WorkloadName string
	Region       string
	Engine       string
	Version      string
	StorageGB    int32
}

// DatabaseResult describes a provisioned database.
type DatabaseResult struct {
	ResourceID string
	Endpoint   string
	Status     string
}

// HealthResult describes the health of a resource.
type HealthResult struct {
	Healthy bool
	Status  string
}
