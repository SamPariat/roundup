# db

Database schema, migrations, and sqlc query definitions.

## Sub-directories

| Directory | Purpose |
|---|---|
| `migrations/` | Ordered goose migration files that define and evolve the database schema. |
| `queries/` | sqlc-annotated SQL files used to generate type-safe Go query functions. |

## `sqlc.yaml`

Lives in the service root (`services/user-service/sqlc.yaml`). Points sqlc at `db/queries/` and `db/migrations/` and writes generated Go code to `internal/adapters/persistence/db/`. Postgres `uuid` columns map to `uuid.UUID`, and nullable columns map to pointers (`*string`).

To regenerate Go code after modifying a query or schema, from the monorepo root:

```bash
make user-sqlc
```

## Tables

| Table | Purpose |
|---|---|
| `users` | User profiles. The ID is shared with auth-service. Soft-deleted via `deleted_at`. |
| `device_tokens` | FCM push tokens, one row per device, owned by the user currently logged in on it |

## Running migrations

From the monorepo root:

```bash
POSTGRES_DSN="postgres://user_user:user_pass@localhost:5434/users?sslmode=disable" make user-migrate-up
```

`user-migrate-down` rolls back the latest migration and `user-migrate-status` lists what has been applied.
