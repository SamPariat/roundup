# db/migrations

goose migration files for the user-service database. Applied in filename order.

## Naming convention

```
{sequence}_{description}.sql
```

Examples:
- `0001_create_users.sql`
- `0002_create_device_tokens.sql`
- `0003_add_index_users_name.sql`

Always use a 4-digit zero-padded sequence number. Never rename or reorder existing migration files — goose tracks which files have been applied by version number.

## Rules

- Every file has a `-- +goose Up` section and a `-- +goose Down` section that undoes it.
- Each file should be idempotent where possible (use `CREATE TABLE IF NOT EXISTS`, `CREATE INDEX IF NOT EXISTS`).
- Never modify an already-applied migration. Create a new migration to alter a table.
- Keep migrations small and focused — one logical change per file.
