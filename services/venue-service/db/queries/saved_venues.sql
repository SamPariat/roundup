-- name: AddFavorite :exec
-- Saves a venue for a user within a squad.
INSERT INTO saved_venues (squad_id, user_id, place_id, name)
VALUES ($1, $2, $3, $4);

-- name: RemoveFavorite :execrows
-- Unsaves a venue. Returns 0 rows if the user had not saved it.
DELETE
FROM saved_venues
WHERE squad_id = $1
  AND user_id = $2
  AND place_id = $3;

-- name: ListFavorites :many
-- Lists every member's saved venues in a squad, newest first.
SELECT id, squad_id, user_id, place_id, name, saved_at
FROM saved_venues
WHERE squad_id = $1
ORDER BY saved_at DESC;

-- name: IsFavorite :one
-- Reports whether this user saved this venue in the squad.
SELECT EXISTS (SELECT 1
               FROM saved_venues
               WHERE squad_id = $1
                 AND user_id = $2
                 AND place_id = $3) AS exists;
