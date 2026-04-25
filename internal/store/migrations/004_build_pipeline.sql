-- +goose Up
CREATE TABLE IF NOT EXISTS builds (
    id              TEXT PRIMARY KEY,
    workload_id     TEXT NOT NULL,
    workspace_id    TEXT NOT NULL,
    git_commit_sha  TEXT NOT NULL,
    git_branch      TEXT NOT NULL,
    status          TEXT NOT NULL DEFAULT 'QUEUED',
    language        TEXT,
    image_ref       TEXT,
    image_digest    TEXT,
    size_bytes      BIGINT,
    error_message   TEXT,
    started_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at    TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_builds_workload_id ON builds(workload_id);
CREATE INDEX idx_builds_status ON builds(status);
CREATE INDEX idx_builds_created_at ON builds(created_at DESC);

CREATE TABLE IF NOT EXISTS build_log_lines (
    id          BIGSERIAL PRIMARY KEY,
    build_id    TEXT NOT NULL REFERENCES builds(id) ON DELETE CASCADE,
    sequence    BIGINT NOT NULL,
    stream      TEXT NOT NULL CHECK (stream IN ('stdout', 'stderr')),
    text        TEXT NOT NULL,
    ts          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (build_id, sequence)
);

CREATE INDEX idx_build_log_lines_build_seq ON build_log_lines(build_id, sequence);

-- +goose Down
DROP TABLE IF EXISTS build_log_lines;
DROP TABLE IF EXISTS builds;
