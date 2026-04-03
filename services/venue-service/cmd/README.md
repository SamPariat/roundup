# cmd

Entry point for the venue-service binary.

## `main.go`

This file is responsible for **wiring only** — it contains no business logic. Its job is to:

1. Load configuration from environment variables via `internal/config`
2. Construct infrastructure clients (PostgreSQL connection pool, Redis client)
3. Instantiate all driven adapters (Google Maps, Redis cache, Postgres repository)
4. Inject those adapters into the use-cases (application layer)
5. Inject the use-cases into the driving adapters (HTTP handlers)
6. Start the Fiber HTTP server
7. Listen for OS signals and perform a graceful shutdown

Nothing in this file should contain conditional logic, data transformation, or domain rules. If you find yourself writing an `if` statement here that isn't about startup/shutdown, it belongs in the application or domain layer instead.
