# db/queries

sqlc-annotated SQL query files. Each file groups queries for one table.

## Files

| File | Tables touched | Purpose |
|---|---|---|
| `users.sql` | `users` | Create, get by ID, get by phone, partial profile update, soft delete |
| `device_tokens.sql` | `device_tokens` | Upsert by token, delete one (owner-checked), delete all for a user |

## sqlc annotation format

Every query must have a sqlc annotation comment immediately above it:

```sql
-- name: GetUserByPhone :one
SELECT *
FROM users
WHERE phone = $1
  AND deleted_at IS NULL;
```

The annotation specifies:
- **name** — the Go method name generated for this query
- **return type** — `:exec` (no rows), `:execrows` (rows affected), `:one` (single row), `:many` (multiple rows)

After editing these files, run `make user-sqlc` from the monorepo root to regenerate the Go code in `internal/adapters/persistence/db/`.
