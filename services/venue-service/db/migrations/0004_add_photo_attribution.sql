-- Stores who took each snapshot's photo, since Google requires crediting the
-- author wherever the photo is shown. Only the first author is kept, as a
-- place photo has one. Rows written before this migration have no credit.

-- +goose Up
ALTER TABLE saved_venues
    ADD COLUMN photo_author_name TEXT NOT NULL DEFAULT '',
    ADD COLUMN photo_author_uri  TEXT NOT NULL DEFAULT '';

ALTER TABLE venue_visits
    ADD COLUMN photo_author_name TEXT NOT NULL DEFAULT '',
    ADD COLUMN photo_author_uri  TEXT NOT NULL DEFAULT '';

-- +goose Down
ALTER TABLE venue_visits
    DROP COLUMN photo_author_name,
    DROP COLUMN photo_author_uri;

ALTER TABLE saved_venues
    DROP COLUMN photo_author_name,
    DROP COLUMN photo_author_uri;
