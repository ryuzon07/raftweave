CREATE TABLE IF NOT EXISTS credentials (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workload_id UUID NOT NULL REFERENCES workloads(id) ON DELETE CASCADE,
    provider VARCHAR(20) NOT NULL,
    encrypted_payload BYTEA NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_credentials_workload_id ON credentials (workload_id);
CREATE INDEX idx_credentials_provider ON credentials (provider);
