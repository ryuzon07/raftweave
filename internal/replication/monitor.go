package replication

import "context"

// LagMonitor monitors replication lag between primary and standby databases.
type LagMonitor interface {
	GetLag(ctx context.Context, primaryRegion string, standbyRegion string) (*LagMetrics, error)
	StartMonitoring(ctx context.Context) error
	Stop() error
}

// LagMetrics contains replication lag measurements.
type LagMetrics struct {
	LagSeconds float64
	LagBytes   int64
}
