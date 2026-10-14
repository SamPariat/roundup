# db

Database schema, migrations, and sqlc query definitions.

## Sub-directories

| Directory | Purpose |
|---|---|
| `migrations/` | Ordered goose migration files that define and evolve the database schema. |
| `queries/` | sqlc-annotated SQL files used to generate type-safe Go query functions. |

## `sqlc.yaml`

Lives in the service root (`services/venue-service/sqlc.yaml`). Points sqlc at `db/queries/` and `db/migrations/` and writes generated Go code to `internal/adapters/persistence/db/`.

To regenerate Go code after modifying a query or schema, from the monorepo root:

```bash
make venue-sqlc
```

## Tables

| Table | Purpose |
|---|---|
| `saved_venues` | Venues saved as favorites by a user within a squad |
| `venue_visits` | Historical record of squad outings at a confirmed venue |

## Running migrations

From the monorepo root:

```bash
POSTGRES_DSN="postgres://venue_user:venue_pass@localhost:5438/venues?sslmode=disable" make venue-migrate-up
```

`venue-migrate-down` rolls back the latest migration and `venue-migrate-status` lists what has been applied.
