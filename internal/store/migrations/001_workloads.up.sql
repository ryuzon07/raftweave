CREATE TABLE IF NOT EXISTS workloads (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(63) NOT NULL UNIQUE,
    descriptor_json JSONB NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'PENDING',
    primary_region VARCHAR(100) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_workloads_name ON workloads (name);
CREATE INDEX idx_workloads_status ON workloads (status);
