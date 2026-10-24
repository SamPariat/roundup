-- 0003_create_squads.sql
-- A squad is a standing group of friends who go out together. Events, saved
-- venues and visit history all belong to a squad.

-- +goose Up
CREATE TABLE IF NOT EXISTS squads
(
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    name       TEXT        NOT NULL,
    avatar_url TEXT,
    created_by UUID        NOT NULL REFERENCES users (id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- +goose Down
DROP TABLE IF EXISTS squads;
