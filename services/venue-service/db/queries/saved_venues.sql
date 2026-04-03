-- name: AddFavorite :exec
INSERT INTO saved_venues (squad_id, user_id, place_id, name)
VALUES ($1, $2, $3, $4);

-- name: RemoveFavorite :exec
DELETE
FROM saved_venues
WHERE squad_id = $1
  AND user_id = $2
  AND place_id = $3;

-- name: ListFavorites :many
SELECT id, squad_id, user_id, place_id, name, saved_at
FROM saved_venues
WHERE squad_id = $1
ORDER BY saved_at DESC;

-- name: IsFavorite :one
SELECT EXISTS (SELECT 1
               FROM saved_venues
               WHERE squad_id = $1
                 AND user_id = $2
                 AND place_id = $3) AS exists;