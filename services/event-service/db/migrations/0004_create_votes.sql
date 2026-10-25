-- 0004_create_votes.sql
-- One vote per member per event, for a nominated place. Votes are open
-- (everyone sees who voted for what) and kept after voting closes, for
-- "won with 6 of 11 votes".

-- +goose Up
CREATE TABLE IF NOT EXISTS votes
(
    event_id UUID        NOT NULL,
    user_id  UUID        NOT NULL,
    place_id TEXT        NOT NULL,
    cast_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    PRIMARY KEY (event_id, user_id),
    FOREIGN KEY (event_id, place_id) REFERENCES nominations (event_id, place_id) ON DELETE CASCADE
);

-- +goose Down
DROP TABLE IF EXISTS votes;
