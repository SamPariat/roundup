-- name: RecordVisit :exec
-- Records one squad visit to a venue during an event, with a snapshot of its display fields.
INSERT INTO venue_visits (squad_id, event_id, place_id, name, visited_at, avg_spend_in_paise,
                          address, rating, rating_count, price_level, photo_ref, photo_width,
                          photo_height)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13);

-- name: GetVisitHistory :many
-- Summarises a squad's visits per venue, most recently visited first.
-- Display fields come from the latest visit. The average spend skips
-- unknown (NULL) values and is 0 when none are known.
-- The casts give sqlc concrete column types instead of interface{}.
SELECT place_id,
       (ARRAY_AGG(name ORDER BY visited_at DESC))[1]::text            AS name,
       (ARRAY_AGG(address ORDER BY visited_at DESC))[1]::text         AS address,
       (ARRAY_AGG(photo_ref ORDER BY visited_at DESC))[1]::text       AS photo_ref,
       (ARRAY_AGG(photo_width ORDER BY visited_at DESC))[1]::integer  AS photo_width,
       (ARRAY_AGG(photo_height ORDER BY visited_at DESC))[1]::integer AS photo_height,
       COUNT(*)                                                       AS visit_count,
       MAX(visited_at)::timestamptz                                   AS last_visited_at,
       COALESCE(ROUND(AVG(avg_spend_in_paise)), 0)::bigint            AS avg_spend_in_paise
FROM venue_visits
WHERE squad_id = $1
GROUP BY place_id
ORDER BY last_visited_at DESC;
