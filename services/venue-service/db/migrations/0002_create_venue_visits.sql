-- 0002_create_venue_visits.sql
-- Stores venue visits made by squads during events.
-- The unique constraint on (squad_id, event_id, place_id) enforces that
-- a squad can only record one visit per venue per event.

-- +goose Up
CREATE TABLE IF NOT EXISTS venue_visits
(
    id                 BIGSERIAL   PRIMARY KEY,
    squad_id           TEXT        NOT NULL,
    event_id           TEXT        NOT NULL,
    place_id           TEXT        NOT NULL,
    name               TEXT        NOT NULL,
    visited_at         TIMESTAMPTZ NOT NULL,
    avg_spend_in_paise BIGINT,

    CONSTRAINT uq_venue_visits UNIQUE (squad_id, event_id, place_id)
);

CREATE INDEX IF NOT EXISTS idx_venue_visits_squad ON venue_visits (squad_id);

-- +goose Down
DROP TABLE IF EXISTS venue_visits;
