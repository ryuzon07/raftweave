package observability

// Named metric instruments used across the system.
const (
	MetricRaftElectionTotal      = "raftweave.raft.election_total"
	MetricRaftLeaderChangesTotal = "raftweave.raft.leader_changes_total"
	MetricReplicationLagSeconds  = "raftweave.replication.lag_seconds"
	MetricBuildDurationSeconds   = "raftweave.build.duration_seconds"
	MetricFailoverTotal          = "raftweave.failover.total"
	MetricFailoverRTOSeconds     = "raftweave.failover.rto_seconds"
	MetricHealthProbeDurationMs  = "raftweave.health.probe_duration_ms"
)
