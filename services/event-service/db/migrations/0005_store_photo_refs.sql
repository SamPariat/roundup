-- Nominations keep the photo's reference instead of its URL, because the
-- direct URLs venue-service now hands out expire; each response turns the
-- reference into a fresh URL. Existing rows held venue-service's photo path,
-- /api/v1/venues/photos/{ref}, so the reference is what follows the prefix.
-- The photo's author is kept too, since Google requires crediting them.

-- +goose Up
ALTER TABLE nominations RENAME COLUMN photo_url TO photo_ref;

UPDATE nominations
SET photo_ref = CASE
                    WHEN photo_ref LIKE '/api/v1/venues/photos/%'
                        THEN substr(photo_ref, length('/api/v1/venues/photos/') + 1)
                    ELSE ''
    END;

ALTER TABLE nominations
    ADD COLUMN photo_author_name TEXT NOT NULL DEFAULT '',
    ADD COLUMN photo_author_uri  TEXT NOT NULL DEFAULT '';

-- +goose Down
-- Rolling back means running the code from before this change, which reads
-- photo_url as venue-service's old photo path and loads the image from it. So
-- the refs are turned back into those paths, /api/v1/venues/photos/{ref}, the
-- reverse of the Up conversion above. (That route is gone after this change;
-- the Down only matters together with rolling the services back too.)
ALTER TABLE nominations
    DROP COLUMN photo_author_name,
    DROP COLUMN photo_author_uri;

UPDATE nominations
SET photo_ref = '/api/v1/venues/photos/' || photo_ref
WHERE photo_ref <> '';

ALTER TABLE nominations RENAME COLUMN photo_ref TO photo_url;
