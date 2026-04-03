# db/queries

sqlc-annotated SQL query files. Each file groups queries for a related set of database operations.

## Files

| File | Tables touched | Purpose |
|---|---|---|
| `venues.sql` | `venue_visits` | Insert a visit record; query visit history grouped by place |
| `favourites.sql` | `saved_venues` | Insert, delete, list, and existence-check for saved venues |

## sqlc annotation format

Every query must have a sqlc annotation comment immediately above it:

```sql
-- name: InsertSavedVenue :exec
INSERT INTO saved_venues (squad_id, user_id, place_id, name)
VALUES ($1, $2, $3, $4);
```

The annotation specifies:
- **name** — the Go method name generated for this query
- **return type** — `:exec` (no rows), `:one` (single row), `:many` (multiple rows)

After editing these files, run `sqlc generate` from the `db/` directory to regenerate the Go code in `internal/adapters/persistence/sqlc/`.
