-- 0002_create_device_tokens.sql
-- Stores FCM push tokens. A token identifies a device, so it is unique across
-- all users: when someone else logs in on the same device the row moves to them.

-- +goose Up
CREATE TABLE IF NOT EXISTS device_tokens
(
    id         BIGSERIAL   PRIMARY KEY,
    user_id    UUID        NOT NULL REFERENCES users (id),
    token      TEXT        NOT NULL UNIQUE,
    platform   TEXT        NOT NULL CHECK (platform IN ('iOS', 'Android')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_device_tokens_user ON device_tokens (user_id);

-- +goose Down
DROP TABLE IF EXISTS device_tokens;
