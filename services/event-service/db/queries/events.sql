-- name: CreateEvent :one
INSERT INTO events (squad_id, name, type, created_by, starts_at, voting_deadline)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, created_at;

-- name: GetEvent :one
-- One event with its winner's snapshot (all NULL until voting closes).
SELECT e.id, e.squad_id, e.name, e.type, e.status, e.created_by, e.starts_at,
       e.voting_deadline, e.closed_at, e.created_at,
       w.place_id AS winner_place_id, w.name AS winner_name, w.address AS winner_address,
       w.lat AS winner_lat, w.lng AS winner_lng, w.rating AS winner_rating,
       w.price_level AS winner_price_level, w.photo_url AS winner_photo_url,
       w.photo_width AS winner_photo_width, w.photo_height AS winner_photo_height,
       w.nominated_by AS winner_nominated_by, w.nominated_at AS winner_nominated_at
FROM events e
         LEFT JOIN nominations w ON w.event_id = e.id AND w.place_id = e.winner_place_id
WHERE e.id = $1;

-- name: UpdateEvent :execrows
-- Edits are allowed only before the place is confirmed.
UPDATE events
SET name = $2, type = $3, starts_at = $4, voting_deadline = $5, updated_at = NOW()
WHERE id = $1
  AND status IN ('DRAFT', 'VOTING');

-- name: ListSquadEvents :many
-- One page of a squad's events: open ones first, soonest first; then closed
-- ones, most recent first. The cursor is the last row's (closed, starts_at,
-- id). A NULL cursor starts at the top. Ties on starts_at are broken by id in
-- the same direction as the time.
SELECT e.id, e.squad_id, e.name, e.type, e.status, e.created_by, e.starts_at,
       e.voting_deadline, e.closed_at, e.created_at,
       w.place_id AS winner_place_id, w.name AS winner_name, w.address AS winner_address,
       w.lat AS winner_lat, w.lng AS winner_lng, w.rating AS winner_rating,
       w.price_level AS winner_price_level, w.photo_url AS winner_photo_url,
       w.photo_width AS winner_photo_width, w.photo_height AS winner_photo_height,
       w.nominated_by AS winner_nominated_by, w.nominated_at AS winner_nominated_at
FROM events e
         LEFT JOIN nominations w ON w.event_id = e.id AND w.place_id = e.winner_place_id
WHERE e.squad_id = sqlc.arg('squad_id')
  AND (sqlc.narg('cursor_closed')::boolean IS NULL
    -- after an open event: later open events, then every closed one
    OR (NOT sqlc.narg('cursor_closed')::boolean AND (
            (e.status <> 'CLOSED' AND (e.starts_at, e.id) > (sqlc.narg('cursor_at')::timestamptz, sqlc.narg('cursor_id')::uuid))
            OR e.status = 'CLOSED'))
    -- after a closed event: earlier closed events
    OR (sqlc.narg('cursor_closed')::boolean AND e.status = 'CLOSED'
        AND (e.starts_at, e.id) < (sqlc.narg('cursor_at')::timestamptz, sqlc.narg('cursor_id')::uuid)))
ORDER BY (e.status = 'CLOSED'),
         CASE WHEN e.status <> 'CLOSED' THEN e.starts_at END,
         CASE WHEN e.status <> 'CLOSED' THEN e.id END,
         CASE WHEN e.status = 'CLOSED' THEN e.starts_at END DESC,
         CASE WHEN e.status = 'CLOSED' THEN e.id END DESC
LIMIT sqlc.arg('lim');

-- name: CountOpenEvents :one
SELECT COUNT(*)
FROM events
WHERE squad_id = $1
  AND status <> 'CLOSED';

-- name: StartVoting :execrows
UPDATE events
SET status = 'VOTING', updated_at = NOW()
WHERE id = $1
  AND status = 'DRAFT'
  AND EXISTS (SELECT 1 FROM nominations n WHERE n.event_id = $1);

-- name: ConfirmVoting :execrows
-- Closes voting and sets the winner in one statement: the nomination with the
-- most votes, a tie (or no votes at all) going to the earliest suggestion.
-- Pass an event ID to confirm that event now, or NULL to confirm every
-- VOTING event whose deadline is at or before due_at. Concurrent calls can't
-- confirm an event twice: the row lock makes the second re-check the status.
UPDATE events e
SET status = 'CONFIRMED', winner_place_id = w.place_id, updated_at = NOW()
FROM (SELECT DISTINCT ON (n.event_id) n.event_id, n.place_id
      FROM nominations n
               JOIN events v ON v.id = n.event_id AND v.status = 'VOTING'
               LEFT JOIN votes vt ON vt.event_id = n.event_id AND vt.place_id = n.place_id
      GROUP BY n.event_id, n.place_id, n.nominated_at
      ORDER BY n.event_id, COUNT(vt.user_id) DESC, n.nominated_at, n.place_id) w
WHERE e.id = w.event_id
  AND e.status = 'VOTING'
  AND CASE
          WHEN sqlc.narg('event_id')::uuid IS NULL THEN e.voting_deadline <= sqlc.arg('due_at')::timestamptz
          ELSE e.id = sqlc.narg('event_id')::uuid
    END;

-- name: StartNight :execrows
UPDATE events
SET status = 'ACTIVE', updated_at = NOW()
WHERE id = $1
  AND status = 'CONFIRMED';

-- name: CloseEvent :one
UPDATE events
SET status = 'CLOSED', closed_at = NOW(), updated_at = NOW()
WHERE id = $1
  AND status = 'ACTIVE'
RETURNING id, squad_id, created_by, starts_at, winner_place_id;

-- name: CloseDueEvents :many
-- Closes nights nobody closed by hand, auto_close_after after they started.
-- A confirmed night nobody started closes too.
UPDATE events
SET status = 'CLOSED', closed_at = NOW(), updated_at = NOW()
WHERE status IN ('CONFIRMED', 'ACTIVE')
  AND starts_at + sqlc.arg('auto_close_after')::interval <= sqlc.arg('due_at')::timestamptz
RETURNING id, squad_id, created_by, starts_at, winner_place_id;
