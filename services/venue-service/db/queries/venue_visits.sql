-- name: RecordVisit :exec
INSERT INTO venue_visits (squad_id, event_id, place_id, name, visited_at, avg_spend_in_paise)
VALUES ($1, $2, $3, $4, $5, $6);

-- name: GetVisitHistory :many
SELECT place_id, name, COUNT(*) AS visit_count, MAX(visited_at) AS last_visited_at
FROM venue_visits
WHERE squad_id = $1
GROUP BY place_id, name
ORDER BY last_visited_at DESC;