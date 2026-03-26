# adapters/persistence/sqlc

Auto-generated code. Do not edit manually.

This directory is produced by running `sqlc generate` from the `db/` directory. It is checked into version control so that the build does not require `sqlc` to be installed on CI or in Docker.

## Regenerating

```bash
cd services/venue-service/db
sqlc generate
```

This reads `db/sqlc.yaml`, processes the SQL files in `db/queries/` against the schema in `db/migrations/`, and writes the output here.

## Contents

| File | Contents |
|---|---|
| `db.go` | `DBTX` interface and `Queries` struct with constructor |
| `models.go` | Go structs corresponding to each database table (`SavedVenue`, `VenueVisit`) |
| `venues.sql.go` | Typed Go functions for each query in `db/queries/venues.sql` |
| `favourites.sql.go` | Typed Go functions for each query in `db/queries/favourites.sql` |

## Usage boundary

These types are consumed exclusively by `adapters/persistence/postgres_venue_repo.go`. They must not appear in any method signature visible to the application or domain layers.
