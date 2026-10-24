-- name: CountOTPCodesSince :one
-- Counts codes sent to a phone since a time, for the per-phone request limit.
SELECT COUNT(*)
FROM otp_codes
WHERE phone = $1
  AND created_at >= $2;

-- name: CreateOTPCode :exec
-- Stores a newly sent code. Only its hash is kept.
INSERT INTO otp_codes (phone, code_hash, expires_at)
VALUES ($1, $2, $3);

-- name: GetLatestOTPCode :one
-- Returns the most recent code for a phone; only that one can be verified.
-- id breaks ties between codes created in the same instant.
SELECT *
FROM otp_codes
WHERE phone = $1
ORDER BY created_at DESC, id DESC
LIMIT 1;

-- name: IncrementOTPAttempts :exec
-- Counts one wrong guess against a code.
UPDATE otp_codes
SET attempts = attempts + 1
WHERE id = $1;

-- name: ConsumeOTPCode :execrows
-- Marks a code used. Returns 0 rows if it was already consumed, so two
-- concurrent verifications cannot both succeed.
UPDATE otp_codes
SET consumed_at = NOW()
WHERE id = $1
  AND consumed_at IS NULL;
