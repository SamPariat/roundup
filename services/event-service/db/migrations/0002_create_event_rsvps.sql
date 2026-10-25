-- 0002_create_event_rsvps.sql
-- Each member's answer to an event. Only answers are stored: a squad member
-- with no row is still pending.

-- +goose Up
CREATE TABLE IF NOT EXISTS event_rsvps
(
    event_id   UUID        NOT NULL REFERENCES events (id) ON DELETE CASCADE,
    user_id    UUID        NOT NULL,
    status     TEXT        NOT NULL CHECK (status IN ('GOING', 'MAYBE', 'NOT_GOING')),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    PRIMARY KEY (event_id, user_id)
);

-- +goose Down
DROP TABLE IF EXISTS event_rsvps;
