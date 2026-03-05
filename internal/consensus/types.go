package consensus

// NodeID uniquely identifies a Raft node.
type NodeID string

// Term is a monotonically increasing election term.
type Term uint64

// LogIndex is a position in the Raft log.
type LogIndex uint64

// RaftRole represents the current role of a Raft node.
type RaftRole int

const (
	Follower RaftRole = iota
	Candidate
	Leader
)

// String returns the string representation of a RaftRole.
func (r RaftRole) String() string {
	switch r {
	case Follower:
		return "follower"
	case Candidate:
		return "candidate"
	case Leader:
		return "leader"
	default:
		return "unknown"
	}
}
