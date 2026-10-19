# cmd

Entry point for the auth-service binary.

## `main.go`

This file is responsible for **wiring only** — it contains no business logic. Its job is to:

1. Load configuration via `internal/config` (panics on missing or unsafe settings, such as `OTP_DEV_CODE` in prod)
2. Refuse to start in prod while only the console SMS sender exists
3. Open the PostgreSQL connection pool via `shared/postgres`
4. Instantiate the driven adapters (OTP and session repositories, console sender, user-service client)
5. Inject them into the use cases, and the use cases into the HTTP handler
6. Register shared middleware (`RequestID`, `Logging`) and start the Fiber HTTP server
