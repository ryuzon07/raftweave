package replication

import "time"

// Status describes replication state between two regions.
type Status struct {
	PrimaryRegion string
	StandbyRegion string
	LagSeconds    float64
	LagBytes      int64
	LastSyncedAt  time.Time
	Status        string
}

// WALEntry represents a single Write-Ahead Log entry.
type WALEntry struct {
	LSN       string
	Timestamp time.Time
	Relation  string
	Operation string
	Data      []byte
}

// PromotionResult captures the outcome of a standby promotion.
type PromotionResult struct {
	Status             string
	PromotedAt         time.Time
	NewPrimaryEndpoint string
}
