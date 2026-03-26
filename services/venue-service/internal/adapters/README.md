# adapters

Driven adapters — concrete implementations of the port interfaces defined in `internal/domain/ports.go`. Each sub-package wraps one external system and translates between its API and the domain's types.

## Sub-packages

| Package | Implements | External system |
|---|---|---|
| `places/` | `domain.PlaceProvider` | Google Maps Places API (swappable) |
| `cache/` | `domain.VenueCache` | Redis via `go-redis/v9` |
| `persistence/` | `domain.VenueRepository` | PostgreSQL via `pgx/v5` + sqlc |

## Adapter contract

Each adapter must:
- Accept only infrastructure clients (e.g. `*redis.Client`, `*pgxpool.Pool`) in its constructor — not domain types or other adapters.
- Map external types (SDK structs, sqlc models, JSON responses) to domain types at the boundary. Nothing from the external SDK should leak into the return values.
- Return domain sentinel errors (e.g. `domain.ErrVenueNotFound`) rather than raw SDK or database errors, so callers can use `errors.Is()` without knowing the adapter's internals.

## Swapping a provider

To replace Google Maps with a different places provider:
1. Create a new file in `places/` (e.g. `foursquare_adapter.go`) that implements `domain.PlaceProvider`.
2. Change one line in `cmd/main.go` to inject the new adapter.
3. Delete (or keep) `google_maps_adapter.go` — no other file needs to change.
