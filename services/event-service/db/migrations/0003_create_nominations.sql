-- 0003_create_nominations.sql
-- Places suggested for an event, with a snapshot of their display fields
-- taken from venue-service when suggested, so the voting screen needs no
-- provider call. The winning nomination is the confirmed venue.

-- +goose Up
CREATE TABLE IF NOT EXISTS nominations
(
    event_id     UUID        NOT NULL REFERENCES events (id) ON DELETE CASCADE,
    place_id     TEXT        NOT NULL,
    name         TEXT        NOT NULL,
    address      TEXT        NOT NULL DEFAULT '',
    lat          DOUBLE PRECISION NOT NULL DEFAULT 0,
    lng          DOUBLE PRECISION NOT NULL DEFAULT 0,
    rating       REAL        NOT NULL DEFAULT 0,
    price_level  INTEGER     NOT NULL DEFAULT 0,
    photo_url    TEXT        NOT NULL DEFAULT '',
    photo_width  INTEGER     NOT NULL DEFAULT 0,
    photo_height INTEGER     NOT NULL DEFAULT 0,
    nominated_by UUID        NOT NULL,
    nominated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    PRIMARY KEY (event_id, place_id)
);

ALTER TABLE events
    ADD CONSTRAINT fk_events_winner FOREIGN KEY (id, winner_place_id)
        REFERENCES nominations (event_id, place_id);

-- +goose Down
ALTER TABLE events DROP CONSTRAINT IF EXISTS fk_events_winner;
DROP TABLE IF EXISTS nominations;
