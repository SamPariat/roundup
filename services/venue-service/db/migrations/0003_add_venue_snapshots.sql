-- 0003_add_venue_snapshots.sql
-- Stores a snapshot of each venue's display fields when it is saved or
-- visited, so favorites and history render without a provider call.
-- Defaults keep rows written before this migration valid; they show as
-- unrated with no address or photo.

-- +goose Up
ALTER TABLE saved_venues
    ADD COLUMN address      TEXT    NOT NULL DEFAULT '',
    ADD COLUMN rating       REAL    NOT NULL DEFAULT 0,
    ADD COLUMN rating_count INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN price_level  INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN photo_ref    TEXT    NOT NULL DEFAULT '',
    ADD COLUMN photo_width  INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN photo_height INTEGER NOT NULL DEFAULT 0;

ALTER TABLE venue_visits
    ADD COLUMN address      TEXT    NOT NULL DEFAULT '',
    ADD COLUMN rating       REAL    NOT NULL DEFAULT 0,
    ADD COLUMN rating_count INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN price_level  INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN photo_ref    TEXT    NOT NULL DEFAULT '',
    ADD COLUMN photo_width  INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN photo_height INTEGER NOT NULL DEFAULT 0;

-- +goose Down
ALTER TABLE venue_visits
    DROP COLUMN address,
    DROP COLUMN rating,
    DROP COLUMN rating_count,
    DROP COLUMN price_level,
    DROP COLUMN photo_ref,
    DROP COLUMN photo_width,
    DROP COLUMN photo_height;

ALTER TABLE saved_venues
    DROP COLUMN address,
    DROP COLUMN rating,
    DROP COLUMN rating_count,
    DROP COLUMN price_level,
    DROP COLUMN photo_ref,
    DROP COLUMN photo_width,
    DROP COLUMN photo_height;
