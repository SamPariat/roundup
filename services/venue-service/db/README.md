# db

Database schema, migrations, and sqlc query definitions.

## Sub-directories

| Directory | Purpose |
|---|---|
| `migrations/` | Ordered SQL migration files that define and evolve the database schema. |
| `queries/` | sqlc-annotated SQL files used to generate type-safe Go query functions. |

## `sqlc.yaml`

Configuration file for [sqlc](https://sqlc.dev). Points sqlc at the `queries/` and `migrations/` directories and specifies the output location for generated Go code (`internal/adapters/persistence/sqlc/`).

To regenerate Go code after modifying a query or schema:

```bash
cd services/venue-service/db
sqlc generate
```

## Tables

| Table | Purpose |
|---|---|
| `saved_venues` | Venues saved as favourites by a user within a squad |
| `venue_visits` | Historical record of squad outings at a confirmed venue |

## Running migrations

Migrations are run manually using the `make migrate` target from the monorepo root:

```bash
make migrate s=venue-service
```

This applies all pending migration files in `migrations/` in filename order against the venue-service's PostgreSQL database.
