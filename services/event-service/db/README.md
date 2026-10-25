# db

Database schema, migrations, and sqlc query definitions.

## Sub-directories

| Directory | Purpose |
|---|---|
| `migrations/` | Ordered goose migration files that define and evolve the database schema. |
| `queries/` | sqlc-annotated SQL files used to generate type-safe Go query functions. |

## `sqlc.yaml`

Lives in the service root (`services/event-service/sqlc.yaml`). Points sqlc at `db/queries/` and `db/migrations/` and writes generated Go code to `internal/adapters/persistence/db/`.

To regenerate Go code after modifying a query or schema, from the monorepo root:

```bash
make event-sqlc
```

## Tables

| Table | Purpose |
|---|---|
| `events` | One night out per row: status, times, and the winning place once voting closes |
| `event_rsvps` | Each member's answer. No row means pending |
| `nominations` | Suggested places, with a snapshot of their details from venue-service |
| `votes` | One vote per member per event, kept after voting closes |

## Running migrations

```bash
make event-migrate-up POSTGRES_DSN="postgres://event_user:event_pass@localhost:5435/events?sslmode=disable"
```
