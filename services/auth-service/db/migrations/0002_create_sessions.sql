-- 0002_create_sessions.sql
-- A session is one login on one device. Each refresh issues a new refresh
-- token on the same session and marks the old one used, so a reused token
-- (a sign of theft) can be detected and the whole session revoked.
-- Only SHA-256 hashes of refresh tokens are stored.

-- +goose Up
CREATE TABLE IF NOT EXISTS sessions
(
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID        NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    revoked_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_sessions_user ON sessions (user_id);

CREATE TABLE IF NOT EXISTS refresh_tokens
(
    id         BIGSERIAL   PRIMARY KEY,
    session_id UUID        NOT NULL REFERENCES sessions (id) ON DELETE CASCADE,
    token_hash BYTEA       NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    used_at    TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_refresh_tokens_session ON refresh_tokens (session_id);

-- +goose Down
DROP TABLE IF EXISTS refresh_tokens;
DROP TABLE IF EXISTS sessions;
