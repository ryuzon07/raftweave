// Package consensus implements the Raft consensus algorithm from scratch.
// This is a custom implementation — no external Raft library is used.
package consensus

// Raft is the core consensus engine.
type Raft struct {
	config     *Config
	state      *State
	log        Log
	transport  Transport
	sm         StateMachine
	membership MembershipManager
	election   ElectionManager
}

// NewRaft creates a new Raft consensus engine.
func NewRaft(cfg *Config, log Log, transport Transport, sm StateMachine, membership MembershipManager, election ElectionManager) *Raft {
	return &Raft{
		config:     cfg,
		state:      NewState(cfg.NodeID),
		log:        log,
		transport:  transport,
		sm:         sm,
		membership: membership,
		election:   election,
	}
}

// State holds the volatile and persistent state of a Raft node.
type State struct {
	NodeID      NodeID
	Role        RaftRole
	CurrentTerm Term
	VotedFor    NodeID
	LeaderID    NodeID
	CommitIndex LogIndex
	LastApplied LogIndex
}

// NewState creates a new initial state for a node.
func NewState(nodeID NodeID) *State {
	return &State{
		NodeID: nodeID,
		Role:   Follower,
	}
}
