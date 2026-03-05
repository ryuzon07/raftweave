package consensus

import "context"

// Transport handles inter-node RPC communication.
type Transport interface {
	// SendRequestVote sends a vote request to a peer.
	SendRequestVote(ctx context.Context, target NodeID, req *VoteRequest) (*VoteResponse, error)
	// SendAppendEntries sends an append entries request to a peer.
	SendAppendEntries(ctx context.Context, target NodeID, req *AppendEntriesRequest) (*AppendEntriesResponse, error)
	// Listen starts listening for incoming RPC requests.
	Listen(ctx context.Context, addr string) error
}

// VoteRequest is the internal domain type for vote requests.
type VoteRequest struct {
	Term         Term
	CandidateID  NodeID
	LastLogIndex LogIndex
	LastLogTerm  Term
}

// VoteResponse is the internal domain type for vote responses.
type VoteResponse struct {
	Term        Term
	VoteGranted bool
}

// AppendEntriesRequest is the internal domain type for append entries.
type AppendEntriesRequest struct {
	Term         Term
	LeaderID     NodeID
	PrevLogIndex LogIndex
	PrevLogTerm  Term
	Entries      []LogEntry
	LeaderCommit LogIndex
}

// AppendEntriesResponse is the internal domain type for append entries responses.
type AppendEntriesResponse struct {
	Term       Term
	Success    bool
	MatchIndex LogIndex
}
