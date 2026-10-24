-- name: CreateSquad :one
INSERT INTO squads (name, avatar_url, created_by)
VALUES ($1, $2, $3)
RETURNING *;

-- name: AddSquadMember :one
INSERT INTO squad_members (squad_id, user_id, role)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetSquadForMember :one
-- Returns no rows for non-members, so a squad you are not in looks the same
-- as one that does not exist.
SELECT s.*
FROM squads s
         JOIN squad_members m ON m.squad_id = s.id
WHERE s.id = $1
  AND m.user_id = $2;

-- name: ListSquadsForUser :many
SELECT s.*,
       m.role,
       (SELECT COUNT(*) FROM squad_members c WHERE c.squad_id = s.id) AS member_count
FROM squads s
         JOIN squad_members m ON m.squad_id = s.id
WHERE m.user_id = $1
ORDER BY s.created_at DESC;

-- name: ListSquadMembers :many
SELECT m.user_id, m.role, m.joined_at, u.name, u.avatar_url
FROM squad_members m
         JOIN users u ON u.id = m.user_id
WHERE m.squad_id = $1
  AND u.deleted_at IS NULL
ORDER BY m.joined_at, m.user_id;

-- name: GetSquadMember :one
SELECT m.user_id, m.role, m.joined_at, u.name, u.avatar_url
FROM squad_members m
         JOIN users u ON u.id = m.user_id
WHERE m.squad_id = $1
  AND m.user_id = $2;

-- name: LockSquad :one
-- Serialises membership changes to one squad for the rest of the transaction.
SELECT id FROM squads WHERE id = $1 FOR UPDATE;

-- name: CountSquadMembers :one
SELECT COUNT(*) FROM squad_members WHERE squad_id = $1;

-- name: CountSquadAdmins :one
SELECT COUNT(*) FROM squad_members WHERE squad_id = $1 AND role = 'ADMIN';

-- name: RemoveSquadMember :execrows
DELETE FROM squad_members WHERE squad_id = $1 AND user_id = $2;

-- name: UpdateSquadMemberRole :execrows
UPDATE squad_members SET role = $3 WHERE squad_id = $1 AND user_id = $2;

-- name: PromoteEarliestMember :exec
-- Makes the longest-standing member admin. user_id breaks ties so the choice
-- is deterministic when two people joined at the same instant.
UPDATE squad_members t
SET role = 'ADMIN'
WHERE t.squad_id = sqlc.arg('squad_id')
  AND t.user_id = (SELECT e.user_id
                   FROM squad_members e
                   WHERE e.squad_id = sqlc.arg('squad_id')
                   ORDER BY e.joined_at, e.user_id
                   LIMIT 1);

-- name: DeleteSquad :exec
DELETE FROM squads WHERE id = $1;
