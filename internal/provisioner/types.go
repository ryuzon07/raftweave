package provisioner

import "time"

// ProvisionInput is the service-layer input for provisioning.
type ProvisionInput struct {
	WorkloadName string
	Provider     string
	Region       string
	Action       string
	Compute      *ResourceSpec
	Database     *DatabaseSpec
}

// ResourceSpec mirrors the proto ResourceSpec for internal use.
type ResourceSpec struct {
	CPU      string
	Memory   string
	Replicas int32
}

// DatabaseSpec mirrors the proto DatabaseSpec for internal use.
type DatabaseSpec struct {
	Engine    string
	Version   string
	StorageGB int32
}

// ProvisionOutput is the service-layer output for provisioning.
type ProvisionOutput struct {
	JobID  string
	Status string
}

// FailoverInput is the service-layer input for failover.
type FailoverInput struct {
	WorkloadName string
	FromRegion   string
	ToRegion     string
	Reason       string
}

// FailoverOutput is the service-layer output for failover.
type FailoverOutput struct {
	CommandID             string
	Status                string
	RTOSeconds            int64
	RPOSeconds            int64
	DataLossWindowSeconds int64
	CompletedAt           time.Time
}

// CloudResource represents a provisioned cloud resource.
type CloudResource struct {
	ResourceID string
	Provider   string
	Region     string
	Type       string
	Status     string
	Endpoint   string
	CreatedAt  time.Time
}
