package consensus

import "context"

// ElectionManager manages leader election logic.
type ElectionManager interface {
	// StartElection begins a new election for the given term.
	StartElection(ctx context.Context, term Term) error
	// ResetElectionTimer resets the randomised election timeout.
	ResetElectionTimer(ctx context.Context) error
	// IsElectionTimedOut returns true if the election timer has expired.
	IsElectionTimedOut() bool
}
