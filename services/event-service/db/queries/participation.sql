-- name: SetRSVP :execrows
-- Stores or changes an answer, unless the event is closed.
INSERT INTO event_rsvps (event_id, user_id, status)
SELECT e.id, sqlc.arg('user_id'), sqlc.arg('status')
FROM events e
WHERE e.id = sqlc.arg('event_id')
  AND e.status <> 'CLOSED'
ON CONFLICT (event_id, user_id) DO UPDATE SET status = EXCLUDED.status, updated_at = NOW();

-- name: ListRSVPs :many
SELECT event_id, user_id, status
FROM event_rsvps
WHERE event_id = ANY (sqlc.arg('event_ids')::uuid[]);

-- name: AddNomination :execrows
-- Stores a suggestion while the event is a draft. A repeat is a unique violation.
INSERT INTO nominations (event_id, place_id, name, address, lat, lng, rating, price_level,
                         photo_url, photo_width, photo_height, nominated_by)
SELECT e.id, sqlc.arg('place_id'), sqlc.arg('name'), sqlc.arg('address'), sqlc.arg('lat'),
       sqlc.arg('lng'), sqlc.arg('rating'), sqlc.arg('price_level'), sqlc.arg('photo_url'),
       sqlc.arg('photo_width'), sqlc.arg('photo_height'), sqlc.arg('nominated_by')
FROM events e
WHERE e.id = sqlc.arg('event_id')
  AND e.status = 'DRAFT';

-- name: RemoveNomination :execrows
DELETE
FROM nominations n
USING events e
WHERE n.event_id = e.id
  AND e.id = sqlc.arg('event_id')
  AND e.status = 'DRAFT'
  AND n.place_id = sqlc.arg('place_id');

-- name: ListNominations :many
SELECT place_id, name, address, lat, lng, rating, price_level, photo_url, photo_width,
       photo_height, nominated_by, nominated_at
FROM nominations
WHERE event_id = $1
ORDER BY nominated_at, place_id;

-- name: CountNominations :many
SELECT event_id, COUNT(*) AS count
FROM nominations
WHERE event_id = ANY (sqlc.arg('event_ids')::uuid[])
GROUP BY event_id;

-- name: CastVote :execrows
-- Stores or changes a vote while voting is open, for a suggested place only.
INSERT INTO votes (event_id, user_id, place_id)
SELECT n.event_id, sqlc.arg('user_id'), n.place_id
FROM nominations n
         JOIN events e ON e.id = n.event_id
WHERE n.event_id = sqlc.arg('event_id')
  AND n.place_id = sqlc.arg('place_id')
  AND e.status = 'VOTING'
ON CONFLICT (event_id, user_id) DO UPDATE SET place_id = EXCLUDED.place_id, cast_at = NOW();

-- name: ListVotes :many
SELECT event_id, user_id, place_id
FROM votes
WHERE event_id = ANY (sqlc.arg('event_ids')::uuid[]);
