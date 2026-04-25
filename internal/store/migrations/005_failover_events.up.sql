CREATE TABLE IF NOT EXISTS failover_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workload_id UUID NOT NULL REFERENCES workloads(id) ON DELETE CASCADE,
    from_region VARCHAR(100) NOT NULL,
    to_region VARCHAR(100) NOT NULL,
    initiated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ,
    rto_seconds BIGINT,
    rpo_seconds BIGINT,
    data_loss_seconds BIGINT,
    status VARCHAR(50) NOT NULL DEFAULT 'IN_PROGRESS',
    reason TEXT
);

CREATE INDEX idx_failover_events_workload_id ON failover_events (workload_id);
CREATE INDEX idx_failover_events_status ON failover_events (status);
CREATE INDEX idx_failover_events_initiated_at ON failover_events (initiated_at DESC);
