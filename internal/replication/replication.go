// Package replication manages WAL streaming, cloud DB translation,
// lag monitoring, and standby promotion.
package replication

import "context"

// ReplicationManager orchestrates database replication across cloud regions.
type ReplicationManager interface {
	GetStatus(ctx context.Context, workloadName string) ([]*Status, error)
	PromoteStandby(ctx context.Context, workloadName string, standbyRegion string, reason string) (*PromotionResult, error)
}
