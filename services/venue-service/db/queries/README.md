# db/queries

sqlc-annotated SQL query files. Each file groups queries for one table.

## Files

| File | Tables touched | Purpose |
|---|---|---|
| `saved_venues.sql` | `saved_venues` | Add, remove, list, and existence-check for saved venues |
| `venue_visits.sql` | `venue_visits` | Record a visit; query visit history grouped by place |

## sqlc annotation format

Every query must have a sqlc annotation comment immediately above it:

```sql
-- name: AddFavorite :exec
INSERT INTO saved_venues (squad_id, user_id, place_id, name)
VALUES ($1, $2, $3, $4);
```

The annotation specifies:
- **name** — the Go method name generated for this query
- **return type** — `:exec` (no rows), `:execrows` (rows affected), `:one` (single row), `:many` (multiple rows)

After editing these files, run `make venue-sqlc` from the monorepo root to regenerate the Go code in `internal/adapters/persistence/db/`.
