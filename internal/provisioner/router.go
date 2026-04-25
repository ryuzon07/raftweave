package provisioner

import "context"

// TrafficRouter manages DNS-based traffic routing across clouds.
type TrafficRouter interface {
	UpdateRoute(ctx context.Context, workloadName string, targetRegion string, targetEndpoint string) error
	GetCurrentRoute(ctx context.Context, workloadName string) (*RouteInfo, error)
}

// RouteInfo describes the current traffic routing state.
type RouteInfo struct {
	WorkloadName string
	Region       string
	Endpoint     string
}
