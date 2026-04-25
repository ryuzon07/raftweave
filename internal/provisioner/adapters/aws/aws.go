// Package aws implements the CloudAdapter interface for Amazon Web Services.
package aws

import (
	"context"

	"github.com/raftweave/raftweave/internal/provisioner/adapters"
)

// Adapter implements adapters.CloudAdapter for AWS.
type Adapter struct{}

// NewAdapter creates a new AWS adapter.
func NewAdapter() *Adapter {
	return &Adapter{}
}

func (a *Adapter) ProvisionCompute(ctx context.Context, req *adapters.ComputeRequest) (*adapters.ComputeResult, error) {
	return nil, nil // stub
}

func (a *Adapter) ProvisionDatabase(ctx context.Context, req *adapters.DatabaseRequest) (*adapters.DatabaseResult, error) {
	return nil, nil // stub
}

func (a *Adapter) DestroyCompute(ctx context.Context, resourceID string) error {
	return nil // stub
}

func (a *Adapter) DestroyDatabase(ctx context.Context, resourceID string) error {
	return nil // stub
}

func (a *Adapter) GetHealth(ctx context.Context, resourceID string) (*adapters.HealthResult, error) {
	return nil, nil // stub
}

func (a *Adapter) Provider() string {
	return "aws"
}

// Compile-time interface assertion.
var _ adapters.CloudAdapter = (*Adapter)(nil)
