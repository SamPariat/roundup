# config

Environment variable loading and service configuration.

## Files

| File | Purpose |
|---|---|
| `config.go` | Defines the `Config` struct and the `Load()` function that reads all required and optional environment variables. |

## Variables

| Env var | Required | Default | Purpose |
|---|---|---|---|
| `PORT` | No | `3006` | HTTP server listen port |
| `POSTGRES_DSN` | Yes | — | PostgreSQL connection string (e.g. `postgres://user:pass@localhost:5438/venue`) |
| `REDIS_ADDR` | No | `localhost:6379` | Redis server address |
| `GOOGLE_MAPS_API_KEY` | Yes | — | Google Maps Places API key |

## Behaviour

- Optional variables fall back to their defaults when unset.
- Required variables (`POSTGRES_DSN`, `GOOGLE_MAPS_API_KEY`) cause the service to `panic` at startup if missing. Fail-fast is preferable to a partially-started service that errors at request time.
- `Load()` is called once in `cmd/main.go` and the resulting `Config` struct is passed into constructors. Config is never read from environment variables outside this package.

## Local development

Copy `.env.example` to `.env` and fill in the values. The `.env` file is gitignored. Use `make infra` from the monorepo root to start the required infrastructure containers before running the service locally.
