package provisioner

import "context"

// FencingController implements STONITH-style fencing to prevent split-brain.
type FencingController interface {
	// Fence isolates a failed region's resources.
	Fence(ctx context.Context, workloadName string, region string) error
	// Unfence restores access to a previously fenced region.
	Unfence(ctx context.Context, workloadName string, region string) error
	// IsFenced returns true if the given region is currently fenced.
	IsFenced(ctx context.Context, workloadName string, region string) (bool, error)
}
