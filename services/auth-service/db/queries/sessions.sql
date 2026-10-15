-- name: CreateSession :one
INSERT INTO sessions (user_id)
VALUES ($1)
RETURNING id;

-- name: CreateRefreshToken :exec
INSERT INTO refresh_tokens (session_id, token_hash, expires_at)
VALUES ($1, $2, $3);

-- name: GetRefreshToken :one
SELECT rt.id, rt.session_id, s.user_id, rt.expires_at, rt.used_at, s.revoked_at
FROM refresh_tokens rt
         JOIN sessions s ON s.id = rt.session_id
WHERE rt.token_hash = $1;

-- name: MarkRefreshTokenUsed :execrows
UPDATE refresh_tokens
SET used_at = NOW()
WHERE id = $1
  AND used_at IS NULL;

-- name: RevokeSession :exec
UPDATE sessions
SET revoked_at = NOW()
WHERE id = $1
  AND revoked_at IS NULL;
