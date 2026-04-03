# domain

The innermost layer of the hexagonal architecture. This package defines **what the venue-service is** — independently of how it is deployed, which database it uses, or which maps provider it talks to.

## Files

| File | Purpose |
|---|---|
| `venue.go` | Core read models: `Venue` (search result), `VenueDetail` (enriched single place), `Review`, `SearchParams` |
| `favourite.go` | Persistence model `SavedVenue` and command types `AddFavouriteCommand`, `RemoveFavouriteCommand` |
| `visit.go` | Persistence model `VenueVisit`, `RecordVisitCommand`, `VisitSummary` (read model for history endpoint) |
| `errors.go` | Sentinel errors (`ErrVenueNotFound`, `ErrProviderUnavailable`, etc.), the `DomainError` wrapper type, and constructor functions |
| `ports.go` | **Port interfaces** — `PlaceProvider`, `VenueCache`, `VenueRepository`. These are the contracts that all driven adapters must satisfy. |

## Rules

- **Zero external imports** — only the Go standard library is allowed here. No third-party SDKs, no framework types.
- **No `json:` struct tags** on domain models — JSON serialisation is an HTTP concern and belongs in `transport/http/dto/`.
- **No SQL types** — domain structs use plain Go types (`string`, `int64`, `time.Time`), never `pgtype.Text` or `sql.NullString`.
- Interfaces are defined here, not in the adapter packages — the consumer (use-case) owns the contract.

## Why ports live here

Go's idiomatic interface placement is at the point of use. The use-cases in `application/` consume `PlaceProvider`, `VenueCache`, and `VenueRepository` — so those interfaces live in `domain/ports.go`, not inside the adapter packages that implement them. This means adding a new adapter (e.g. Foursquare) requires zero changes to this package.
