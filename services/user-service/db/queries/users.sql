-- name: CreateUser :one
-- Creates a user with just a phone number; the profile is filled in later.
INSERT INTO users (phone)
VALUES ($1)
RETURNING *;

-- name: GetUserByID :one
-- Returns an active (not soft-deleted) user by ID.
SELECT *
FROM users
WHERE id = $1
  AND deleted_at IS NULL;

-- name: GetUserByPhone :one
-- Returns the active user with this phone number.
SELECT *
FROM users
WHERE phone = $1
  AND deleted_at IS NULL;

-- name: UpdateUserProfile :one
-- Updates the given profile fields; a NULL argument leaves that column as is.
UPDATE users
SET name       = COALESCE(sqlc.narg('name'), name),
    avatar_url = COALESCE(sqlc.narg('avatar_url'), avatar_url),
    updated_at = NOW()
WHERE id = sqlc.arg('id')
  AND deleted_at IS NULL
RETURNING *;

-- name: SoftDeleteUser :execrows
-- Marks a user deleted. Their phone number becomes free to sign up again.
UPDATE users
SET deleted_at = NOW(),
    updated_at = NOW()
WHERE id = $1
  AND deleted_at IS NULL;

-- name: ListUsersByPhones :many
-- Returns the active users with any of these phone numbers, for contact lookup.
SELECT *
FROM users
WHERE phone = ANY (sqlc.arg('phones')::text[])
  AND deleted_at IS NULL;
