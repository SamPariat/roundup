# application

Use-case orchestration layer. Each file corresponds to one cohesive area of business functionality and contains a single use-case struct.

## Files

| File | Struct | Responsibilities |
|---|---|---|
| `venue_usecase.go` | `VenueUseCase` | Cache-first nearby search; cache-first place detail fetch |
| `favourite_usecase.go` | `FavouriteUseCase` | Add a venue to a squad's favourites (with duplicate check); remove; list |
| `history_usecase.go` | `HistoryUseCase` | Record a venue visit triggered by a confirmed event; retrieve visit history for a squad |

## How use-cases work

Each use-case struct holds references to **port interfaces** from `internal/domain/ports.go` — never to concrete adapter types. This means:

- Unit tests can inject simple in-memory stubs with no Redis or Postgres running.
- Swapping an adapter (e.g. replacing Google Maps with Foursquare) requires no changes here.

A use-case method should:
1. Validate or prepare inputs using domain types
2. Coordinate one or more port calls (cache lookup, provider fetch, persistence write)
3. Return domain types or domain errors — never HTTP status codes or raw SQL errors

## What does NOT belong here

- HTTP request parsing or response serialisation → `transport/http/`
- SQL queries or Redis commands → `adapters/`
- Business rules expressed only as data structure definitions → `domain/`
