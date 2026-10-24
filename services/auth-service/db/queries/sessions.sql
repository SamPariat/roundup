-- name: CreateSession :one
-- Starts a login session for a user and returns its ID.
INSERT INTO sessions (user_id)
VALUES ($1)
RETURNING id;

-- name: CreateRefreshToken :exec
-- Stores the hash of a refresh token issued for a session.
INSERT INTO refresh_tokens (session_id, token_hash, expires_at)
VALUES ($1, $2, $3);

-- name: GetRefreshToken :one
-- Looks a refresh token up by hash, with its session's user and revocation
-- state, so a refresh can be checked in one query.
SELECT rt.id, rt.session_id, s.user_id, rt.expires_at, rt.used_at, s.revoked_at
FROM refresh_tokens rt
         JOIN sessions s ON s.id = rt.session_id
WHERE rt.token_hash = $1;

-- name: MarkRefreshTokenUsed :execrows
-- Marks a refresh token spent. Returns 0 rows if it was already used, which
-- signals token reuse.
UPDATE refresh_tokens
SET used_at = NOW()
WHERE id = $1
  AND used_at IS NULL;

-- name: RevokeSession :exec
-- Ends a session; every refresh token in it stops working.
UPDATE sessions
SET revoked_at = NOW()
WHERE id = $1
  AND revoked_at IS NULL;
