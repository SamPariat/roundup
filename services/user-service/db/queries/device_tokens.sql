-- name: UpsertDeviceToken :one
INSERT INTO device_tokens (user_id, token, platform)
VALUES ($1, $2, $3)
ON CONFLICT (token) DO UPDATE
    SET user_id  = EXCLUDED.user_id,
        platform = EXCLUDED.platform
RETURNING *;

-- name: DeleteDeviceToken :execrows
DELETE
FROM device_tokens
WHERE user_id = $1
  AND token = $2;

-- name: DeleteDeviceTokensForUser :exec
DELETE
FROM device_tokens
WHERE user_id = $1;
