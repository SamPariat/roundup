-- 0001_create_otp_codes.sql
-- One-time login codes sent by SMS. Only a hash of each code is stored.
-- Only the newest code per phone is valid; older rows are kept so the
-- per-phone request rate limit can count them.

-- +goose Up
CREATE TABLE IF NOT EXISTS otp_codes
(
    id          BIGSERIAL   PRIMARY KEY,
    phone       TEXT        NOT NULL,
    code_hash   BYTEA       NOT NULL,
    attempts    INT         NOT NULL DEFAULT 0,
    expires_at  TIMESTAMPTZ NOT NULL,
    consumed_at TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_otp_codes_phone_created ON otp_codes (phone, created_at DESC);

-- +goose Down
DROP TABLE IF EXISTS otp_codes;
