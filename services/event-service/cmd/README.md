# cmd

Entry point for the event-service binary.

## `main.go`

This file is responsible for **wiring only** — it contains no business logic. Its job is to:

1. Load configuration from environment variables via `internal/config`
2. Construct infrastructure clients (PostgreSQL pool via `shared/postgres`, the Casbin enforcer via `shared/authz`)
3. Instantiate all driven adapters (Postgres repository, `shared/userclient`, venue-service client)
4. Inject those adapters into the use-cases (application layer)
5. Inject the use-cases into the driving adapters (HTTP handlers)
6. Register shared middleware (`RequestID`, `Logging`) and start the Fiber HTTP server

Nothing in this file should contain conditional logic, data transformation, or domain rules. If you find yourself writing an `if` statement here that isn't about startup, it belongs in the application or domain layer instead.
