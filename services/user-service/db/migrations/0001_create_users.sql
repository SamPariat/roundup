-- 0001_create_users.sql
-- Stores user profiles. The ID is shared with auth-service, which owns login.
-- Phone uniqueness only applies to active (non-deleted) users, so a deleted
-- account's number can sign up again.

-- +goose Up
CREATE TABLE IF NOT EXISTS users
(
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    phone      TEXT        NOT NULL,
    name       TEXT,
    avatar_url TEXT,
    deleted_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_users_phone_active ON users (phone) WHERE deleted_at IS NULL;

-- +goose Down
DROP TABLE IF EXISTS users;
