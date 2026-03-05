package replication

import "context"

// StandbyPromoter handles promoting a standby database to primary.
type StandbyPromoter interface {
	Promote(ctx context.Context, workloadName string, standbyRegion string) (*PromotionResult, error)
}
