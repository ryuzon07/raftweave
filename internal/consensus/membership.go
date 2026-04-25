package consensus

import "context"

// MembershipManager manages the Raft cluster membership.
type MembershipManager interface {
	// AddNode adds a node to the cluster.
	AddNode(ctx context.Context, nodeID NodeID, addr string) error
	// RemoveNode removes a node from the cluster.
	RemoveNode(ctx context.Context, nodeID NodeID) error
	// GetMembers returns all current cluster members.
	GetMembers(ctx context.Context) ([]Member, error)
	// QuorumSize returns the quorum size for the current membership.
	QuorumSize() int
}

// Member represents a cluster member.
type Member struct {
	NodeID NodeID
	Addr   string
}
