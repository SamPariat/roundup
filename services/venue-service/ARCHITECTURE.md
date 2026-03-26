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
│    FavouriteHandler  │           │  adapters/cache/          │
│    HistoryHandler    │           │    RedisCacheAdapter      │ ◄─ implements VenueCache
│    server.go (Fiber) │           │  adapters/persistence/    │
│                      │           │    PostgresVenueRepo      │ ◄─ implements VenueRepository
│                      │           └───────────────────────────┘
└──────────┬───────────┘                       ▲
           │ calls                             │ implements
           ▼                                   │
┌──────────────────────────────────────────────┘
│  APPLICATION LAYER  internal/application/
│  VenueUseCase      — depends on PlaceProvider + VenueCache (ports)
│  FavouriteUseCase  — depends on VenueRepository (port)
│  HistoryUseCase    — depends on VenueRepository (port)
└──────────────────────────┬───────────────────
                           │ imports
                           ▼
┌──────────────────────────────────────────────────────────────────┐
│  DOMAIN  internal/domain/   — zero external imports              │
│  Structs: Venue, VenueDetail, SavedVenue, VenueVisit,            │
│           SearchParams, *Command types                           │
│  Ports (interfaces owned here, not in adapters):                 │
│    PlaceProvider   — SearchNearby, GetDetail                     │
│    VenueCache      — GetNearby/Set, GetDetail/Set, Invalidate    │
│    VenueRepository — SaveFavourite, DeleteFavourite,             │
│                      ListFavourites, IsFavourite,                │
│                      RecordVisit, GetVisitHistory                │
│  Errors: sentinel vars + DomainError + constructors              │
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
│   │   ├── venue.go                     # Venue, VenueDetail, SearchParams
│   │   ├── favourite.go                 # SavedVenue, AddFavouriteCommand, RemoveFavouriteCommand
│   │   ├── visit.go                     # VenueVisit, RecordVisitCommand, VisitSummary
│   │   ├── errors.go                    # sentinel errors + DomainError type + constructors
│   │   └── ports.go                     # PlaceProvider, VenueCache, VenueRepository interfaces
│   │
│   ├── application/
│   │   ├── venue_usecase.go             # Search (cache-first), GetDetail
│   │   ├── favourite_usecase.go         # Add, Remove, List
│   │   └── history_usecase.go           # RecordVisit, GetHistory
│   │
│   ├── adapters/
│   │   ├── places/
│   │   │   └── google_maps_adapter.go   # implements domain.PlaceProvider
│   │   ├── cache/
│   │   │   └── redis_cache_adapter.go   # implements domain.VenueCache (24h nearby, 1h detail)
│   │   └── persistence/
│   │       ├── postgres_venue_repo.go   # implements domain.VenueRepository
│   │       └── sqlc/                    # sqlc-generated (db.go, models.go, *_queries.sql.go)
│   │
│   ├── transport/
│   │   └── http/
│   │       ├── server.go                # Fiber app, route registration
│   │       ├── venue_handler.go         # GET /venues/search, GET /venues/:placeId
│   │       ├── favourite_handler.go     # POST/DELETE/GET /venues/favourites
│   │       ├── history_handler.go       # GET /venues/history
│   │       ├── middleware/
│   │       │   ├── request_id.go
│   │       │   └── logging.go
│   │       └── dto/
│   │           ├── request.go           # SearchRequest, FavouriteRequest + Validate()
│   │           └── response.go          # VenueResponse, VenueListResponse, ErrorResponse
│   │
│   └── config/
│       └── config.go                    # env-var loading (PORT, POSTGRES_DSN, REDIS_ADDR, etc.)
│
├── db/
│   ├── migrations/
│   │   ├── 0001_create_saved_venues.sql
│   │   └── 0002_create_venue_visits.sql
│   ├── queries/
│   │   ├── venues.sql
│   │   └── favourites.sql
│   └── sqlc.yaml
│
├── go.mod
├── go.sum
├── Dockerfile
└── .env.example
```

---

## Port Interfaces (`internal/domain/ports.go`)

```go
package domain

import "context"

// PlaceProvider — swap Google Maps for any provider by writing a new adapter
type PlaceProvider interface {
    SearchNearby(ctx context.Context, params SearchParams) ([]Venue, error)
    GetDetail(ctx context.Context, placeID string) (VenueDetail, error)
}

// VenueCache — swap Redis for any cache by writing a new adapter
type VenueCache interface {
    GetNearby(ctx context.Context, key string) (venues []Venue, found bool, err error)
    SetNearby(ctx context.Context, key string, venues []Venue) error
    GetDetail(ctx context.Context, placeID string) (detail VenueDetail, found bool, err error)
    SetDetail(ctx context.Context, placeID string, detail VenueDetail) error
    InvalidateDetail(ctx context.Context, placeID string) error
}

// VenueRepository — swap Postgres for any DB by writing a new adapter
type VenueRepository interface {
    SaveFavourite(ctx context.Context, cmd AddFavouriteCommand) error
    DeleteFavourite(ctx context.Context, cmd RemoveFavouriteCommand) error
    ListFavourites(ctx context.Context, squadID string) ([]SavedVenue, error)
    IsFavourite(ctx context.Context, squadID, userID, placeID string) (bool, error)
    RecordVisit(ctx context.Context, cmd RecordVisitCommand) error
    GetVisitHistory(ctx context.Context, squadID string) ([]VisitSummary, error)
}
```

---

## Domain Models

### `internal/domain/venue.go`
```go
type Venue struct {
    PlaceID    string
    Name       string
    Address    string
    Lat, Lng   float64
    Rating     float32
    PriceLevel int      // 0-4, provider-normalised
    Types      []string
    IsOpen     *bool    // nil = unknown
    PhotoRefs  []string
}

type VenueDetail struct {
    Venue
    PhoneNumber      string
    Website          string
    OpeningHours     []string
    Reviews          []Review
    EditorialSummary string
}

type SearchParams struct {
    Lat, Lng float64
    RadiusM  int
    Query    string
    Type     string
}
```

### `internal/domain/favourite.go`
```go
type SavedVenue struct {
    ID      int64
    SquadID string
    UserID  string
    PlaceID string
    Name    string     // denormalised for display
    SavedAt time.Time
}

type AddFavouriteCommand    struct { SquadID, UserID, PlaceID, Name string }
type RemoveFavouriteCommand struct { SquadID, UserID, PlaceID string }
```

### `internal/domain/visit.go`
```go
type VenueVisit struct {
    ID            int64
    SquadID       string
    EventID       string
    PlaceID       string
    VisitedAt     time.Time
    AvgSpendPaise int64
}

type RecordVisitCommand struct {
    SquadID, EventID, PlaceID string
    VisitedAt                 time.Time
    AvgSpendPaise             int64
}

type VisitSummary struct {
    PlaceID    string
    Name       string
    VisitCount int
    LastVisit  time.Time
}
```

---

## Adapter Responsibilities

| Adapter | File | Implements |
|---|---|---|
| `GoogleMapsAdapter` | `adapters/places/google_maps_adapter.go` | `domain.PlaceProvider` |
| `RedisCacheAdapter` | `adapters/cache/redis_cache_adapter.go` | `domain.VenueCache` |
| `PostgresVenueRepository` | `adapters/persistence/postgres_venue_repo.go` | `domain.VenueRepository` |

To swap Google Maps: write `adapters/places/foursquare_adapter.go` implementing `PlaceProvider`, change one line in `main.go`. Domain and application layers are untouched.

---

## main.go Wiring (outline)

```go
// 1. Load config
cfg := config.Load()

// 2. Infrastructure clients
pgPool  := pgxpool.New(cfg.PostgresDSN)
redisCl := redis.NewClient(cfg.RedisAddr)

// 3. Driven adapters
googleAdapter := places.NewGoogleMapsAdapter(cfg.GoogleMapsAPIKey)
cacheAdapter  := cache.NewRedisCacheAdapter(redisCl)
repoAdapter   := persistence.NewPostgresVenueRepository(pgPool)

// 4. Use-cases (depend on ports, not concrete adapters)
venueUC     := application.NewVenueUseCase(googleAdapter, cacheAdapter)
favouriteUC := application.NewFavouriteUseCase(repoAdapter)
historyUC   := application.NewHistoryUseCase(repoAdapter)

// 5. Driving adapters
venueH    := http.NewVenueHandler(venueUC)
favH      := http.NewFavouriteHandler(favouriteUC)
historyH  := http.NewHistoryHandler(historyUC)
app       := http.NewServer(venueH, favH, historyH)

// 6. Run
app.Listen(":3006")
```

---

## HTTP Routes

| Method | Path | Handler | Use-case |
|---|---|---|---|
| GET | /venues/search | VenueHandler.Search | VenueUseCase.Search |
| GET | /venues/:placeId | VenueHandler.GetDetail | VenueUseCase.GetDetail |
| POST | /venues/favourites | FavouriteHandler.Add | FavouriteUseCase.Add |
| DELETE | /venues/favourites | FavouriteHandler.Remove | FavouriteUseCase.Remove |
| GET | /venues/favourites | FavouriteHandler.List | FavouriteUseCase.List |
| GET | /venues/history | HistoryHandler.List | HistoryUseCase.GetHistory |
| GET | /health | inline | — |

---

## Error Mapping (transport/http layer only)

| Domain sentinel | HTTP status |
|---|---|
| `ErrVenueNotFound` | 404 |
| `ErrInvalidPlaceID`, `ErrInvalidCoordinates` | 400 |
| `ErrAlreadyFavourited` | 409 |
| `ErrProviderUnavailable` | 502 |
| anything else | 500 |

---

## Database Schema

```sql
-- saved_venues
CREATE TABLE saved_venues (
    id        BIGSERIAL PRIMARY KEY,
    squad_id  TEXT NOT NULL,
    user_id   TEXT NOT NULL,
    place_id  TEXT NOT NULL,
    name      TEXT NOT NULL,
    saved_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (squad_id, user_id, place_id)
);

-- venue_visits
CREATE TABLE venue_visits (
    id               BIGSERIAL PRIMARY KEY,
    squad_id         TEXT NOT NULL,
    event_id         TEXT NOT NULL,
    place_id         TEXT NOT NULL,
    visited_at       TIMESTAMPTZ NOT NULL,
    avg_spend_paise  BIGINT NOT NULL DEFAULT 0
);
```

---

## Key Design Rules

1. **Interfaces live in `domain/ports.go`**, not in adapter packages — the consumer (use-case) owns the contract
2. **DTOs live in `transport/http/dto/`** — domain structs never carry `json:` tags or HTTP concerns
3. **`mapDomainError` lives in the handler** — domain never knows about HTTP status codes
4. **sqlc types never cross into application/domain** — repo adapter maps them at the boundary
5. **Cache writes are best-effort** — `_ = cache.Set(...)` so a cache failure never degrades an API response
