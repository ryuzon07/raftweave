CREATE TABLE IF NOT EXISTS build_jobs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workload_id UUID NOT NULL REFERENCES workloads(id) ON DELETE CASCADE,
    commit_sha VARCHAR(40) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'QUEUED',
    image_digest VARCHAR(255),
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    error TEXT
);

CREATE INDEX idx_build_jobs_workload_id ON build_jobs (workload_id);
CREATE INDEX idx_build_jobs_status ON build_jobs (status);
CREATE INDEX idx_build_jobs_commit_sha ON build_jobs (commit_sha);
