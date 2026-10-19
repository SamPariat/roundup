# venue-service: Hexagonal Architecture LLD

## Context
The venue-service wraps the Google Places API but must allow swapping the provider in the future (Foursquare, HERE, etc.) with zero changes to business logic. Hexagonal architecture (ports & adapters) achieves this by keeping the domain at the centre, defining provider contracts as interfaces inside the domain package, and pushing all external concerns (HTTP, Redis, Postgres, Google Maps) out to adapter packages.

---

## Layer Diagram

```
┌──────────────────────────────────────────────────────────────────┐
│  cmd/main.go — wiring only, no logic                             │
│  constructs adapters → injects into use-cases → into handlers    │
└──────────┬───────────────────────────────────┬───────────────────┘
           │                                   │
           ▼                                   ▼
┌──────────────────────┐           ┌───────────────────────────┐
│  DRIVING ADAPTERS    │           │  DRIVEN ADAPTERS          │
│  transport/http/     │           │  adapters/places/         │
│    VenueHandler      │           │    GoogleMapsAdapter      │ ◄─ implements PlaceProvider
│    FavoriteHandler   │           │  adapters/cache/          │
│    HistoryHandler    │           │    RedisCacheAdapter      │ ◄─ implements VenueCache
│    router.go (Fiber) │           │  adapters/persistence/    │
│                      │           │    PostgresVenueRepository│ ◄─ implements VenueRepository
└──────────┬───────────┘           └───────────────────────────┘
           │ calls                             ▲
           ▼                                   │ implements
┌──────────────────────────────────────────────┘
│  APPLICATION LAYER  internal/application/
│  VenueUseCase     — depends on PlaceProvider + VenueCache (ports)
│  FavoriteUseCase  — depends on VenueRepository (port)
│  HistoryUseCase   — depends on VenueRepository (port)
└──────────────────────────┬───────────────────
                           │ imports
                           ▼
┌──────────────────────────────────────────────────────────────────┐
│  DOMAIN  internal/domain/   — imports only the standard library  │
│  Structs: Venue, VenueDetail, SavedVenue, VisitSummary,          │
│           SearchParams, *Command types                           │
│  Ports (interfaces owned here, not in adapters):                 │
│    PlaceProvider   — SearchNearby, GetDetail                     │
│    VenueCache      — GetNearby/Set, GetDetail/Set, Invalidate    │
│    VenueRepository — AddFavorite, RemoveFavorite, ListFavorites, │
│                      IsFavorite, RecordVisit, GetVisitHistory    │
│  Errors: sentinel vars + DomainError + constructor               │
└──────────────────────────────────────────────────────────────────┘
All dependency arrows point inward. Domain has no outward arrows.
```

---

## Directory Tree

```
services/venue-service/
├── cmd/
│   └── main.go                          # wiring only
│
├── internal/
│   ├── domain/
│   │   ├── venue.go                     # Venue, VenueDetail, PhotoRef, Review, SearchParams
│   │   ├── favorite.go                  # SavedVenue, AddFavoriteCommand, RemoveFavoriteCommand
│   │   ├── visit.go                     # RecordVisitCommand, VisitSummary
│   │   ├── errors.go                    # sentinel errors + DomainError type + constructor
│   │   └── port.go                      # PlaceProvider, VenueCache, VenueRepository interfaces
│   │
│   ├── application/
│   │   ├── venue_usecase.go             # Search (cache-first), GetDetail (cache-first)
│   │   ├── favorite_usecase.go          # AddFavorite, RemoveFavorite, ListFavorites, IsFavorite
│   │   └── history_usecase.go           # RecordVisit, GetVisitHistory
│   │
│   ├── adapters/
│   │   ├── places/
│   │   │   └── google_maps_adapter.go   # implements domain.PlaceProvider
│   │   ├── cache/
│   │   │   └── redis_cache_adapter.go   # implements domain.VenueCache
│   │   └── persistence/
│   │       ├── postgres_venue_repository.go  # implements domain.VenueRepository
│   │       └── db/                      # sqlc-generated (db.go, models.go, *.sql.go)
│   │
│   ├── transport/
│   │   └── http/
│   │       ├── router.go                # route registration under /api/v1
│   │       ├── venue_handler.go
│   │       ├── favorite_handler.go
│   │       ├── history_handler.go
│   │       └── dtos/
│   │           ├── venue_dtos.go
│   │           ├── favorite_dtos.go
│   │           ├── history_dtos.go
│   │           └── mappers.go           # domain → response DTOs
│   │
│   └── config/
│       └── config.go                    # env-var loading (APP_ENV, PORT, POSTGRES_DSN, REDIS_ADDR, GOOGLE_MAPS_API_KEY, LOG_LEVEL)
│
├── db/
│   ├── migrations/
│   │   ├── 0001_create_saved_venues.sql
│   │   └── 0002_create_venue_visits.sql
│   └── queries/
│       ├── saved_venues.sql
│       └── venue_visits.sql
│
├── sqlc.yaml
├── go.mod
├── go.sum
├── Dockerfile
├── .env.beta.example                # template for .env.beta (local + beta)
└── .env.prod.example                # template for .env.prod
```

---

## Port Interfaces (`internal/domain/port.go`)

```go
package domain

import "context"

// PlaceProvider — swap Google Maps for any provider by writing a new adapter
type PlaceProvider interface {
    SearchNearby(ctx context.Context, params SearchParams) ([]Venue, error)
    GetDetail(ctx context.Context, placeID string) (*VenueDetail, error)
}

// VenueCache — all cache operations are best-effort; callers ignore write errors
type VenueCache interface {
    GetNearby(ctx context.Context, key string) ([]Venue, error)
    SetNearby(ctx context.Context, key string, venues []Venue) error
    GetDetail(ctx context.Context, placeID string) (*VenueDetail, error)
    SetDetail(ctx context.Context, placeID string, detail *VenueDetail) error
    InvalidateDetail(ctx context.Context, placeID string) error
}

// VenueRepository — swap Postgres for any DB by writing a new adapter
type VenueRepository interface {
    AddFavorite(ctx context.Context, cmd AddFavoriteCommand) error
    RemoveFavorite(ctx context.Context, cmd RemoveFavoriteCommand) error
    ListFavorites(ctx context.Context, squadID string) ([]SavedVenue, error)
    IsFavorite(ctx context.Context, squadID, userID, placeID string) (bool, error)
    RecordVisit(ctx context.Context, cmd RecordVisitCommand) error
    GetVisitHistory(ctx context.Context, squadID string) ([]VisitSummary, error)
}
```

---

## Domain Models

### `internal/domain/venue.go`
```go
type Venue struct {
    PlaceID       string
    Name          string
    Address       string
    Latitude      float64
    Longitude     float64
    AverageRating float32
    PriceLevel    int        // 0-4, provider-normalised
    Types         []string
    IsOpen        *bool      // nil = unknown
    PhotoRefs     []PhotoRef // Reference, Width, Height
}

type VenueDetail struct {
    Venue
    PhoneNumber      string
    Website          string
    OpeningHours     []string
    Reviews          []Review // AuthorName, Rating, Text, PublishedAt
    EditorialSummary string
}

type SearchParams struct {
    Latitude       float64
    Longitude      float64
    RadiusInMeters int
    Query          string
    Type           string
}
```

### `internal/domain/favorite.go`
```go
type SavedVenue struct {
    ID      int64
    SquadID string
    UserID  string
    PlaceID string
    Name    string     // denormalised for display
    SavedAt time.Time
}

type AddFavoriteCommand    struct { SquadID, UserID, PlaceID, Name string }
type RemoveFavoriteCommand struct { SquadID, UserID, PlaceID string }
```

### `internal/domain/visit.go`
```go
type RecordVisitCommand struct {
    SquadID, EventID, PlaceID, Name string
    VisitedAt                       time.Time
    AvgSpendInPaise                 int64
}

type VisitSummary struct {
    PlaceID       string
    Name          string
    VisitCount    int64
    LastVisitedAt time.Time
}
```

---

## Adapter Responsibilities

| Adapter | File | Implements |
|---|---|---|
| `GoogleMapsAdapter` | `adapters/places/google_maps_adapter.go` | `domain.PlaceProvider` |
| `RedisCacheAdapter` | `adapters/cache/redis_cache_adapter.go` | `domain.VenueCache` |
| `PostgresVenueRepository` | `adapters/persistence/postgres_venue_repository.go` | `domain.VenueRepository` |

To swap Google Maps: write `adapters/places/foursquare_adapter.go` implementing `PlaceProvider`, change one line in `main.go`. Domain and application layers are untouched.

---

## main.go Wiring (outline)

```go
// 1. Load config
cfg := config.Load()

// 2. Infrastructure clients
redisClient := redis.NewClient(&redis.Options{Addr: cfg.RedisAddr})
pool, err := postgres.Connect(ctx, cfg.PostgresDSN) // shared/postgres — pings at startup

// 3. Driven adapters
venueRepo    := persistence.NewPostgresVenueRepository(pool)
mapsAdapter  := places.NewGoogleMapsAdapter(cfg.GoogleMapsAPIKey)
cacheAdapter := cache.NewRedisCacheAdapter(redisClient, 10*time.Minute, log)

// 4. Use-cases (depend on ports, not concrete adapters)
venueUC    := application.NewVenueUseCase(mapsAdapter, cacheAdapter)
favoriteUC := application.NewFavoriteUseCase(venueRepo)
historyUC  := application.NewHistoryUseCase(venueRepo)

// 5. Driving adapters
app := fiber.New()
app.Use(middleware.RequestID(), middleware.Logging(log))
transporthttp.RegisterRoutes(app, &transporthttp.Handlers{...})

// 6. Run
app.Listen(":" + cfg.Port) // 3006
```

---

## HTTP Routes

All routes are versioned under `/api/v1`.

| Method | Path | Handler | Use-case | Success |
|---|---|---|---|---|
| GET | /api/v1/venues | VenueHandler.Search | VenueUseCase.Search | 200 |
| GET | /api/v1/venues/:placeID | VenueHandler.GetDetail | VenueUseCase.GetDetail | 200 |
| GET | /api/v1/squads/:squadID/favorites | FavoriteHandler.ListFavorites | FavoriteUseCase.ListFavorites | 200 |
| GET | /api/v1/squads/:squadID/favorites/:placeID/check | FavoriteHandler.IsFavorite | FavoriteUseCase.IsFavorite | 200 |
| POST | /api/v1/squads/:squadID/favorites | FavoriteHandler.AddFavorite | FavoriteUseCase.AddFavorite | 201 |
| DELETE | /api/v1/squads/:squadID/favorites/:placeID | FavoriteHandler.RemoveFavorite | FavoriteUseCase.RemoveFavorite | 204 |
| GET | /api/v1/squads/:squadID/history | HistoryHandler.GetVisitHistory | HistoryUseCase.GetVisitHistory | 200 |
| POST | /api/v1/squads/:squadID/history | HistoryHandler.RecordVisit | HistoryUseCase.RecordVisit | 201 |

Search takes query params `lat`, `lng`, `radius`, `query`, `type`.

All `/squads/:squadID/*` routes require a UUID `X-User-ID` header (`shared/middleware.UserID()`); without it they return 401. Handlers read the caller with `middleware.UserIDFrom(c)` — the user ID is never taken from the body or query string.

---

## Error Mapping (transport/http layer only)

| Domain sentinel | HTTP status |
|---|---|
| request body / query fails to bind, `ErrInvalidSearch` | 400 |
| missing or invalid `X-User-ID` on squad routes (middleware) | 401 |
| `ErrVenueNotFound`, `ErrNotSaved` | 404 |
| `ErrAlreadySaved`, `ErrVisitAlreadyRecorded` | 409 |
| `ErrProviderUnavailable` | 503 |
| anything else | 500 |

---

## Database Schema

```sql
-- saved_venues
CREATE TABLE saved_venues (
    id       BIGSERIAL   PRIMARY KEY,
    squad_id TEXT        NOT NULL,
    user_id  TEXT        NOT NULL,
    place_id TEXT        NOT NULL,
    name     TEXT        NOT NULL,
    saved_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_saved_venues UNIQUE (squad_id, user_id, place_id)
);

-- venue_visits — see db/migrations/0002_create_venue_visits.sql for the full definition
```

---

## Error Propagation Chain

Library errors are translated into domain sentinels at the adapter boundary. The use-case layer never sees `pgx`, `redis`, or `googlemaps` types.

```
┌───────────────┐   wrap    ┌──────────────────────┐  pass-through  ┌────────────┐  *ErrorStatus  ┌──────────┐
│ External lib  │ ────────► │ Adapter (driven)     │ ─────────────► │ Use-case   │ ─────────────► │ Handler  │
│ pgx / redis / │           │ returns domain error │                │ returns    │                │ → HTTP   │
│ googlemaps    │           └──────────────────────┘                └────────────┘                └──────────┘
└───────────────┘
```

| Adapter | Library error | Mapped domain sentinel |
|---|---|---|
| `PostgresVenueRepository` | unique violation (`postgres.IsUniqueViolation`) | `domain.ErrAlreadySaved` |
| `PostgresVenueRepository` | favorite does not exist on remove | `domain.ErrNotSaved` |
| `PostgresVenueRepository` | unique violation on visits | `domain.ErrVisitAlreadyRecorded` |
| `RedisCacheAdapter` | `redis.Nil` | `nil` return — a cache miss, not an error |
| `RedisCacheAdapter` | timeout / conn refused / bad JSON | logged with zap, swallowed (best-effort) |
| `GoogleMapsAdapter` | `NOT_FOUND`; `INVALID_REQUEST` on place details | `domain.ErrVenueNotFound` |
| `GoogleMapsAdapter` | `ZERO_RESULTS` | not an error — empty result list |
| `GoogleMapsAdapter` | `REQUEST_DENIED` / `OVER_QUERY_LIMIT` / `UNKNOWN_ERROR` | `domain.ErrProviderUnavailable` |

**Rule:** If an adapter cannot map a library error to a known sentinel, it returns it as-is so the handler's default branch returns 500.

---

## Middleware Chain

```
request ──► RequestID ──► Logging ──► [UserID — squad routes only] ──► route handler ──► response
```

| Middleware | Source | Responsibility |
|---|---|---|
| `RequestID` | `shared/middleware` | Reads or generates `X-Request-Id`, stores it in locals, echoes it on the response |
| `Logging` | `shared/middleware` | Logs method, path, status, latency, and request ID |
| `UserID` | `shared/middleware` | Squad routes only: requires a UUID `X-User-ID` header (401 otherwise) and stores it for `UserIDFrom` |

**Token validation is not in this chain.** It happens in the API gateway, which sets `X-User-ID` from the verified access token. The service trusts that header, so it must only be reachable through the gateway.

---

## Caching Strategy

The `RedisCacheAdapter` implements `domain.VenueCache`. All cache operations are best-effort — see Key Design Rule 5.

| Data | Key | TTL |
|---|---|---|
| Nearby search | `venue-search: {lat}:{lng}:{radius}:{query}:{type}` | 10 min |
| Venue detail | `{placeID}` | 10 min |

The TTL is a single value passed to `NewRedisCacheAdapter` in `main.go` and applied to every write.

### Cache Miss Flow

```
Client ──► VenueUseCase.Search()
               │
               ├── cached, _ := cache.GetNearby(key)
               │       ├── cached != nil  → return cached
               │       └── cached == nil  ↓
               │
               ├── provider.SearchNearby(params)      // Google Maps call
               │
               ├── _ = cache.SetNearby(key, results)  // best-effort write
               │
               └── return results
```

There is no active invalidation — entries expire via TTL. `InvalidateDetail` exists on the port for future use. Values are stored as JSON via `encoding/json`.

---

## Testing Strategy

| Layer | Test type | Technique | External deps |
|---|---|---|---|
| `domain/` | Unit | Pure structs and error constructors | None |
| `application/` | Unit | Hand-rolled fakes implementing `PlaceProvider`, `VenueCache`, `VenueRepository` | None |
| `adapters/places/` | Unit | Response-mapping tests in `google_maps_adapter_test.go` | None |
| `adapters/cache/`, `adapters/persistence/` | Integration (future) | Real Redis / Postgres | Docker |
| `transport/http/` | Unit (future) | `fiber.Test()` with use-cases backed by fakes | None |

Unit tests live alongside source (`venue_usecase_test.go` next to `venue_usecase.go`). Run with `make venue-test`.

---

## Lifecycle

- `GET /health` is liveness and never touches dependencies. `GET /ready` pings Postgres only: Redis is a cache and the use case ignores its errors, so a Redis outage slows searches but doesn't make the service unready.
- On SIGTERM/SIGINT, `shared/server.Run` stops accepting connections and waits up to 8s for in-flight requests, then `main` closes the pool and Redis client.
- `shared/middleware.Recover` turns a handler panic into a logged 500.

## Not Yet Implemented

- Tracing and Prometheus metrics

---

## Key Design Rules

1. **Interfaces live in `domain/port.go`**, not in adapter packages — the consumer (use-case) owns the contract
2. **DTOs live in `transport/http/dtos/`** — domain structs never carry `json:` tags or HTTP concerns
3. **Status-code mapping lives in the handlers** — domain never knows about HTTP status codes
4. **sqlc types never cross into application/domain** — the repository adapter maps them at the boundary
5. **Cache writes are best-effort** — `_ = cache.SetNearby(...)` so a cache failure never degrades an API response
