// Package provisioner manages cloud resource provisioning and failover
// across AWS, Azure, and GCP.
package provisioner

import "context"

// CloudProvisioner defines the top-level provisioner interface.
type CloudProvisioner interface {
	Provision(ctx context.Context, req *ProvisionInput) (*ProvisionOutput, error)
	Deprovision(ctx context.Context, workloadName string, provider string, region string) error
	ExecuteFailover(ctx context.Context, req *FailoverInput) (*FailoverOutput, error)
	GetResourceStatus(ctx context.Context, workloadName string, provider string) ([]*CloudResource, error)
}
