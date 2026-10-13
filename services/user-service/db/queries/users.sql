-- name: CreateUser :one
INSERT INTO users (phone)
VALUES ($1)
RETURNING *;

-- name: GetUserByID :one
SELECT *
FROM users
WHERE id = $1
  AND deleted_at IS NULL;

-- name: GetUserByPhone :one
SELECT *
FROM users
WHERE phone = $1
  AND deleted_at IS NULL;

-- name: UpdateUserProfile :one
UPDATE users
SET name       = COALESCE(sqlc.narg('name'), name),
    avatar_url = COALESCE(sqlc.narg('avatar_url'), avatar_url),
    updated_at = NOW()
WHERE id = sqlc.arg('id')
  AND deleted_at IS NULL
RETURNING *;

-- name: SoftDeleteUser :execrows
UPDATE users
SET deleted_at = NOW(),
    updated_at = NOW()
WHERE id = $1
  AND deleted_at IS NULL;
