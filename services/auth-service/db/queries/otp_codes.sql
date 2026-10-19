-- name: CountOTPCodesSince :one
SELECT COUNT(*)
FROM otp_codes
WHERE phone = $1
  AND created_at >= $2;

-- name: CreateOTPCode :exec
INSERT INTO otp_codes (phone, code_hash, expires_at)
VALUES ($1, $2, $3);

-- name: GetLatestOTPCode :one
SELECT *
FROM otp_codes
WHERE phone = $1
ORDER BY created_at DESC, id DESC
LIMIT 1;

-- name: IncrementOTPAttempts :exec
UPDATE otp_codes
SET attempts = attempts + 1
WHERE id = $1;

-- name: ConsumeOTPCode :execrows
UPDATE otp_codes
SET consumed_at = NOW()
WHERE id = $1
  AND consumed_at IS NULL;
