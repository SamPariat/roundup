# adapters/persistence

Implements `domain.VenueRepository` — the port that abstracts all persistent storage operations for venue favourites and visit history.

## Files

| File | Purpose |
|---|---|
| `postgres_venue_repo.go` | Production adapter. Uses `pgx/v5` connection pool and delegates to sqlc-generated queries. Maps sqlc model types to domain types at the boundary. |
| `sqlc/` | Auto-generated Go code produced by running `sqlc generate` from the `db/` directory. Do not edit these files by hand. |

## sqlc

All SQL lives in `db/queries/` as annotated `.sql` files. Running `sqlc generate` (configured by `db/sqlc.yaml`) produces:

- `sqlc/db.go` — `*Queries` struct and constructor
- `sqlc/models.go` — Go structs mirroring each database table row
- `sqlc/*.sql.go` — one typed method per named SQL query

The repository adapter calls these generated methods and maps the resulting `sqlc.SavedVenue` / `sqlc.VenueVisit` structs into `domain.SavedVenue` / `domain.VenueVisit` before returning them. **sqlc types must never appear in method signatures that cross the adapter boundary.**

## Schema overview

| Table | Purpose |
|---|---|
| `saved_venues` | Venues saved as favourites by a user within a squad. Unique per `(squad_id, user_id, place_id)`. |
| `venue_visits` | Record of every confirmed outing at a venue. Written when an event is confirmed. |

See `db/migrations/` for the full DDL.
