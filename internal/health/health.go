// Package health implements health checking and probing across all subsystems.
package health

import "context"

// ProbeAgent manages health probes across all systems.
type ProbeAgent interface {
	// ProbeAll runs all configured probes and returns an aggregate report.
	ProbeAll(ctx context.Context) (*HealthReport, error)
	// ProbeService runs probes for a specific service.
	ProbeService(ctx context.Context, serviceName string) (*ProbeResult, error)
	// RegisterProbe registers a new health probe.
	RegisterProbe(name string, probe Prober)
}
