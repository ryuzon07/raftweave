package store

import "time"

// WorkloadRow is the database representation of a workload.
type WorkloadRow struct {
	ID             string
	Name           string
	DescriptorJSON []byte
	Status         string
	PrimaryRegion  string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// CredentialRow is the database representation of a cloud credential.
type CredentialRow struct {
	ID               string
	WorkloadID       string
	Provider         string
	EncryptedPayload []byte
	CreatedAt        time.Time
}

// BuildJobRow is the database representation of a build job.
type BuildJobRow struct {
	ID          string
	WorkloadID  string
	CommitSHA   string
	Status      string
	ImageDigest string
	StartedAt   time.Time
	CompletedAt time.Time
	Error       string
}

// RaftStateRow is the database representation of Raft node state.
type RaftStateRow struct {
	NodeID      string
	Role        string
	Term        uint64
	VotedFor    string
	LogIndex    uint64
	CommitIndex uint64
	UpdatedAt   time.Time
}

// FailoverEventRow is the database representation of a failover event.
type FailoverEventRow struct {
	ID              string
	WorkloadID      string
	FromRegion      string
	ToRegion        string
	InitiatedAt     time.Time
	CompletedAt     time.Time
	RTOSeconds      int64
	RPOSeconds      int64
	DataLossSeconds int64
	Status          string
	Reason          string
}
