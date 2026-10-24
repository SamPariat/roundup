-- name: UpsertDeviceToken :one
-- Registers a push token to a user. A token identifies a device, so if it
-- already exists it moves to this user; created_at is kept.
INSERT INTO device_tokens (user_id, token, platform)
VALUES ($1, $2, $3)
ON CONFLICT (token) DO UPDATE
    SET user_id  = EXCLUDED.user_id,
        platform = EXCLUDED.platform
RETURNING *;

-- name: DeleteDeviceToken :execrows
-- Removes a token, only if this user owns it.
DELETE
FROM device_tokens
WHERE user_id = $1
  AND token = $2;

-- name: DeleteDeviceTokensForUser :exec
-- Removes every token a user registered, when their account is deleted.
DELETE
FROM device_tokens
WHERE user_id = $1;
