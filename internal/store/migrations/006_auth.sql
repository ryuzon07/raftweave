-- +goose Up

CREATE TABLE IF NOT EXISTS users (
    id                TEXT PRIMARY KEY,
    email             TEXT NOT NULL,
    name              TEXT NOT NULL,
    avatar_url        TEXT,
    provider          TEXT NOT NULL,
    provider_id       TEXT NOT NULL,
    github_login      TEXT,
    github_token_enc  BYTEA,
    is_email_verified BOOLEAN NOT NULL DEFAULT FALSE,
    is_active         BOOLEAN NOT NULL DEFAULT TRUE,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_login_at     TIMESTAMPTZ,
    UNIQUE (email)
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_provider ON users(provider, provider_id);
CREATE INDEX IF NOT EXISTS idx_users_email_lower ON users(LOWER(email));

CREATE TABLE IF NOT EXISTS user_providers (
    user_id      TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    provider     TEXT NOT NULL,
    provider_id  TEXT NOT NULL,
    linked_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, provider)
);

CREATE TABLE IF NOT EXISTS sessions (
    id                  TEXT PRIMARY KEY,
    user_id             TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    refresh_token_hash  TEXT NOT NULL,
    fingerprint         TEXT NOT NULL,
    expires_at          TIMESTAMPTZ NOT NULL,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    revoked_at          TIMESTAMPTZ,
    is_revoked          BOOLEAN NOT NULL DEFAULT FALSE
);
CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_sessions_expires_at ON sessions(expires_at);
CREATE INDEX IF NOT EXISTS idx_sessions_is_revoked ON sessions(is_revoked);

CREATE TABLE IF NOT EXISTS workspace_memberships (
    user_id       TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    workspace_id  TEXT NOT NULL,
    role          TEXT NOT NULL,
    invited_by    TEXT REFERENCES users(id),
    joined_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, workspace_id)
);
CREATE INDEX IF NOT EXISTS idx_memberships_workspace ON workspace_memberships(workspace_id);

-- +goose Down
DROP TABLE IF EXISTS workspace_memberships;
DROP TABLE IF EXISTS sessions;
DROP TABLE IF EXISTS user_providers;
DROP TABLE IF EXISTS users;
