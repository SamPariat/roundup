-- name: RecordVisit :exec
-- Records one squad visit to a venue during an event, with a snapshot of its display fields.
INSERT INTO venue_visits (squad_id, event_id, place_id, name, visited_at, avg_spend_in_paise,
                          address, rating, rating_count, price_level, photo_ref, photo_width,
                          photo_height)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13);

-- name: GetVisitHistory :many
-- Summarises one page of a squad's visits per venue, most recently visited
-- first. Display fields come from the latest visit. The average spend skips
-- unknown (NULL) values and is 0 when none are known.
-- A NULL cursor starts at the most recent; otherwise the page starts after
-- the (last_visited_at, place_id) the cursor names, filtered outside the
-- grouping because it compares grouped values.
-- The casts give sqlc concrete column types instead of interface{}.
SELECT h.place_id, h.name, h.address, h.photo_ref, h.photo_width, h.photo_height,
       h.visit_count, h.last_visited_at, h.avg_spend_in_paise
FROM (SELECT place_id,
             (ARRAY_AGG(name ORDER BY visited_at DESC))[1]::text            AS name,
             (ARRAY_AGG(address ORDER BY visited_at DESC))[1]::text         AS address,
             (ARRAY_AGG(photo_ref ORDER BY visited_at DESC))[1]::text       AS photo_ref,
             (ARRAY_AGG(photo_width ORDER BY visited_at DESC))[1]::integer  AS photo_width,
             (ARRAY_AGG(photo_height ORDER BY visited_at DESC))[1]::integer AS photo_height,
             COUNT(*)                                                       AS visit_count,
             MAX(visited_at)::timestamptz                                   AS last_visited_at,
             COALESCE(ROUND(AVG(avg_spend_in_paise)), 0)::bigint            AS avg_spend_in_paise
      FROM venue_visits
      WHERE squad_id = sqlc.arg('squad_id')
      GROUP BY place_id) h
WHERE sqlc.narg('cursor_at')::timestamptz IS NULL
   OR (h.last_visited_at, h.place_id) < (sqlc.narg('cursor_at')::timestamptz, sqlc.narg('cursor_place')::text)
ORDER BY h.last_visited_at DESC, h.place_id DESC
LIMIT sqlc.arg('lim');
