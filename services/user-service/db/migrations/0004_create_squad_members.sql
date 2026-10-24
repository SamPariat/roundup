-- 0004_create_squad_members.sql
-- Who belongs to which squad, and as what. A squad always has at least one
-- ADMIN while it has members; when the last member leaves the squad row is
-- deleted and these rows cascade with it.

-- +goose Up
CREATE TABLE IF NOT EXISTS squad_members
(
    squad_id  UUID        NOT NULL REFERENCES squads (id) ON DELETE CASCADE,
    user_id   UUID        NOT NULL REFERENCES users (id),
    role      TEXT        NOT NULL DEFAULT 'MEMBER' CHECK (role IN ('ADMIN', 'MEMBER')),
    joined_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (squad_id, user_id)
);

-- The primary key serves lookups by squad; this one serves "my squads".
CREATE INDEX IF NOT EXISTS idx_squad_members_user ON squad_members (user_id);

-- +goose Down
DROP TABLE IF EXISTS squad_members;
