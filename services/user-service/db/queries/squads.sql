-- name: CreateSquad :one
-- Creates a squad. The caller adds the creator as its first admin.
INSERT INTO squads (name, avatar_url, created_by)
VALUES ($1, $2, $3)
RETURNING *;

-- name: AddSquadMember :one
-- Adds a user to a squad with the given role.
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
-- Lists the squads a user belongs to, newest first, with their role and the
-- member count.
SELECT s.*,
       m.role,
       (SELECT COUNT(*) FROM squad_members c WHERE c.squad_id = s.id) AS member_count
FROM squads s
         JOIN squad_members m ON m.squad_id = s.id
WHERE m.user_id = $1
ORDER BY s.created_at DESC;

-- name: ListSquadMembers :many
-- Lists a squad's active members with their profile, earliest joiner first.
SELECT m.user_id, m.role, m.joined_at, u.name, u.avatar_url
FROM squad_members m
         JOIN users u ON u.id = m.user_id
WHERE m.squad_id = $1
  AND u.deleted_at IS NULL
ORDER BY m.joined_at, m.user_id;

-- name: ListMemberPreviews :many
-- Returns the first per_squad active members (earliest joiner first) of each
-- of the given squads in one query, for the avatar row on the squad list.
SELECT p.squad_id, p.user_id, p.role, p.joined_at, p.name, p.avatar_url
FROM (SELECT m.squad_id,
             m.user_id,
             m.role,
             m.joined_at,
             u.name,
             u.avatar_url,
             ROW_NUMBER() OVER (PARTITION BY m.squad_id ORDER BY m.joined_at, m.user_id) AS position
      FROM squad_members m
               JOIN users u ON u.id = m.user_id
      WHERE m.squad_id = ANY (@squad_ids::uuid[])
        AND u.deleted_at IS NULL) p
WHERE p.position <= @per_squad::bigint
ORDER BY p.squad_id, p.position;

-- name: GetSquadMember :one
-- Returns one member's role and profile. No rows means not a member.
SELECT m.user_id, m.role, m.joined_at, u.name, u.avatar_url
FROM squad_members m
         JOIN users u ON u.id = m.user_id
WHERE m.squad_id = $1
  AND m.user_id = $2;

-- name: UpdateSquad :one
-- Sets the name and/or avatar of a squad; a NULL argument leaves that column
-- as it is, and an empty avatar URL clears the picture.
UPDATE squads
SET name       = COALESCE(sqlc.narg('name'), name),
    avatar_url = CASE
                     WHEN sqlc.narg('avatar_url')::text IS NULL THEN avatar_url
                     ELSE NULLIF(sqlc.narg('avatar_url')::text, '')
        END
WHERE id = sqlc.arg('id')
RETURNING *;

-- name: LockSquad :one
-- Serialises membership changes to one squad for the rest of the transaction.
SELECT id
FROM squads
WHERE id = $1
    FOR UPDATE;

-- name: CountSquadMembers :one
-- Counts a squad's members, for the member cap and deleting an empty squad.
SELECT COUNT(*)
FROM squad_members
WHERE squad_id = $1;

-- name: CountSquadAdmins :one
-- Counts a squad's admins, so it is never left without one.
SELECT COUNT(*)
FROM squad_members
WHERE squad_id = $1
  AND role = 'ADMIN';

-- name: RemoveSquadMember :execrows
-- Removes a user from a squad. Returns 0 rows if they were not in it.
DELETE
FROM squad_members
WHERE squad_id = $1
  AND user_id = $2;

-- name: UpdateSquadMemberRole :execrows
-- Sets a member's role. Returns 0 rows if they are not in the squad.
UPDATE squad_members
SET role = $3
WHERE squad_id = $1
  AND user_id = $2;

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
-- Deletes a squad; its members are removed by ON DELETE CASCADE.
DELETE
FROM squads
WHERE id = $1;
