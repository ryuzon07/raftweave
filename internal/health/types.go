package health

import "time"

// HealthReport is an aggregate health status across all subsystems.
type HealthReport struct {
	Overall string
	Results map[string]*ProbeResult
}

// ProbeResult contains the outcome of a single health probe.
type ProbeResult struct {
	Name      string
	Healthy   bool
	Latency   time.Duration
	Error     string
	CheckedAt time.Time
}
