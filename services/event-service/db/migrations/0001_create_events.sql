-- 0001_create_events.sql
-- An event is one night out planned by a squad. It moves through
-- DRAFT → VOTING → CONFIRMED → ACTIVE → CLOSED; the application enforces the
-- order, the CHECK only the allowed values. squad_id and user IDs belong to
-- user-service, so they have no foreign keys here.

-- +goose Up
CREATE TABLE IF NOT EXISTS events
(
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    squad_id        UUID        NOT NULL,
    name            TEXT        NOT NULL,
    type            TEXT        NOT NULL CHECK (type IN ('DRINKS', 'DINING', 'BOTH')),
    status          TEXT        NOT NULL DEFAULT 'DRAFT'
        CHECK (status IN ('DRAFT', 'VOTING', 'CONFIRMED', 'ACTIVE', 'CLOSED')),
    created_by      UUID        NOT NULL,
    starts_at       TIMESTAMPTZ NOT NULL,
    voting_deadline TIMESTAMPTZ NOT NULL,
    -- Set when voting closes; the winner's details live in its nomination.
    winner_place_id TEXT,
    closed_at       TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_events_deadline CHECK (voting_deadline <= starts_at)
);

-- A squad's list: open events by start time, then closed ones.
CREATE INDEX IF NOT EXISTS idx_events_squad ON events (squad_id, starts_at DESC, id);

-- The confirm-due and close-due sweeps.
CREATE INDEX IF NOT EXISTS idx_events_status_deadline ON events (status, voting_deadline);
CREATE INDEX IF NOT EXISTS idx_events_status_starts ON events (status, starts_at);

-- +goose Down
DROP TABLE IF EXISTS events;
