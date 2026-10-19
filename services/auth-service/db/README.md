# db

Database schema, migrations, and sqlc query definitions.

| Directory | Purpose |
|---|---|
| `migrations/` | Ordered goose migrations that define the schema. |
| `queries/` | sqlc-annotated SQL used to generate type-safe Go query functions. |

`sqlc.yaml` (service root) writes generated code to `internal/adapters/persistence/db/`. Regenerate after changing a query or migration, from the monorepo root:

```bash
make auth-sqlc
```

## Tables

| Table | Purpose |
|---|---|
| `otp_codes` | Hashed one-time codes with attempts, expiry and consumed time |
| `sessions` | One per login (device); `revoked_at` ends it |
| `refresh_tokens` | Hashed refresh tokens; `used_at` marks a rotated token |

## Running migrations

```bash
docker compose up -d postgres-auth
POSTGRES_DSN="postgres://auth_user:auth_pass@localhost:5433/auth?sslmode=disable" make auth-migrate-up
```
