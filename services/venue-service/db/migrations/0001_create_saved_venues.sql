-- 0001_create_saved_venues.sql
-- Stores venue favorites saved by squad members.
-- The unique constraint on (squad_id, user_id, place_id) enforces the
-- "already saved" invariant at the database level.

-- +goose Up
CREATE TABLE IF NOT EXISTS saved_venues
(
    id       BIGSERIAL   PRIMARY KEY,
    squad_id TEXT        NOT NULL,
    user_id  TEXT        NOT NULL,
    place_id TEXT        NOT NULL,
    name     TEXT        NOT NULL,
    saved_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_saved_venues UNIQUE (squad_id, user_id, place_id)
);

CREATE INDEX IF NOT EXISTS idx_saved_venues_squad ON saved_venues (squad_id);

-- +goose Down
DROP TABLE IF EXISTS saved_venues;
