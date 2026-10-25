-- name: AddFavorite :exec
-- Saves a venue for a user within a squad, with a snapshot of its display fields.
INSERT INTO saved_venues (squad_id, user_id, place_id, name, address, rating, rating_count,
                          price_level, photo_ref, photo_width, photo_height)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11);

-- name: RemoveFavorite :execrows
-- Unsaves a venue. Returns 0 rows if the user had not saved it.
DELETE
FROM saved_venues
WHERE squad_id = $1
  AND user_id = $2
  AND place_id = $3;

-- name: ListSquadFavorites :many
-- Lists a squad's favorites, one row per venue, most recently saved first.
-- Only saves by member_ids (the squad's current members) count, so a venue
-- saved only by people who left is not listed. Display fields come from the
-- latest save; saved_by lists who saved it, most recent first.
-- The casts give sqlc concrete column types instead of interface{}.
SELECT place_id,
       (ARRAY_AGG(name ORDER BY saved_at DESC))[1]::text            AS name,
       (ARRAY_AGG(address ORDER BY saved_at DESC))[1]::text         AS address,
       (ARRAY_AGG(rating ORDER BY saved_at DESC))[1]::real          AS rating,
       (ARRAY_AGG(rating_count ORDER BY saved_at DESC))[1]::integer AS rating_count,
       (ARRAY_AGG(price_level ORDER BY saved_at DESC))[1]::integer  AS price_level,
       (ARRAY_AGG(photo_ref ORDER BY saved_at DESC))[1]::text       AS photo_ref,
       (ARRAY_AGG(photo_width ORDER BY saved_at DESC))[1]::integer  AS photo_width,
       (ARRAY_AGG(photo_height ORDER BY saved_at DESC))[1]::integer AS photo_height,
       ARRAY_AGG(user_id ORDER BY saved_at DESC)::text[]            AS saved_by,
       MAX(saved_at)::timestamptz                                   AS last_saved_at
FROM saved_venues
WHERE squad_id = $1
  AND user_id = ANY (@member_ids::text[])
GROUP BY place_id
ORDER BY last_saved_at DESC;

-- name: SavedInSquads :many
-- Returns which of the given squads this user saved this venue in.
SELECT squad_id
FROM saved_venues
WHERE user_id = $1
  AND place_id = $2
  AND squad_id = ANY (@squad_ids::text[]);
