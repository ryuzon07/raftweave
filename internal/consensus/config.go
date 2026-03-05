package consensus

import "time"

// Config holds all tunables for the Raft consensus engine.
type Config struct {
	NodeID              NodeID
	BindAddr            string
	Peers               []string
	DataDir             string
	ElectionTimeoutMin  time.Duration
	ElectionTimeoutMax  time.Duration
	HeartbeatInterval   time.Duration
	MaxLogEntriesPerReq int
	SnapshotThreshold   uint64
}

// DefaultConfig returns a Config with sensible defaults.
func DefaultConfig(nodeID NodeID) *Config {
	return &Config{
		NodeID:              nodeID,
		ElectionTimeoutMin:  150 * time.Millisecond,
		ElectionTimeoutMax:  300 * time.Millisecond,
		HeartbeatInterval:   50 * time.Millisecond,
		MaxLogEntriesPerReq: 100,
		SnapshotThreshold:   10000,
	}
}
