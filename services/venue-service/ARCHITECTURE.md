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
    GetDetail(ctx context.Context, placeID string) (*VenueDetail, error)
}

// VenueCache — swap Redis for any cache by writing a new adapter
// nil return = cache miss; non-nil error = cache failure (best-effort, callers ignore)
type VenueCache interface {
    GetNearby(ctx context.Context, key string) ([]Venue, error)
    SetNearby(ctx context.Context, key string, venues []Venue) error
    GetDetail(ctx context.Context, placeID string) (*VenueDetail, error)
    SetDetail(ctx context.Context, placeID string, detail *VenueDetail) error
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

## Graceful Shutdown

The service traps SIGTERM and SIGINT, then tears down resources in reverse-construction order. Kubernetes sends SIGTERM during pod eviction; the 10 s budget below fits inside the default 30 s `terminationGracePeriodSeconds`.

```
signal ──► cancel root ctx ──► Fiber ShutdownWithTimeout(10s) ──► pgxpool.Close() ──► redis.Close() ──► log.Sync()
```

```go
// cmd/main.go (shutdown section)
ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
defer stop()

go func() { _ = app.Listen(":" + cfg.Port) }()

<-ctx.Done()
log.Info("shutting down")

_ = app.ShutdownWithTimeout(10 * time.Second) // drain in-flight HTTP
pgPool.Close()                                 // close idle + active conns
_ = redisCl.Close()                            // close Redis conn pool
_ = log.Sync()                                 // flush buffered zap entries
```

**Key points:**
- `ShutdownWithTimeout` lets Fiber finish in-flight requests before closing listeners.
- Postgres pool is closed **after** Fiber so that draining requests can still hit the DB.
- The root `context.Context` from `signal.NotifyContext` is propagated to use-cases so long-running provider calls abort promptly on shutdown.

---

## Health Checks

Two probes serve different Kubernetes needs:

| Endpoint | Probe type | What it checks | Failure consequence |
|---|---|---|---|
| `GET /healthz` | **Liveness** | Process is up, Fiber is accepting | kubelet restarts the pod |
| `GET /health` | **Readiness** | Postgres `pool.Ping()` + Redis `client.Ping()` | pod removed from Service endpoints (no traffic) |

```go
// Liveness — always 200 if the process is running
app.Get("/healthz", func(c fiber.Ctx) error {
    return c.SendStatus(fiber.StatusOK)
})

// Readiness — 200 only when both dependencies are reachable
app.Get("/health", func(c fiber.Ctx) error {
    ctx, cancel := context.WithTimeout(c.Context(), 2*time.Second)
    defer cancel()

    if err := pgPool.Ping(ctx); err != nil {
        return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"pg": "down"})
    }
    if err := redisCl.Ping(ctx).Err(); err != nil {
        return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"redis": "down"})
    }
    return c.SendStatus(fiber.StatusOK)
})
```

Kubernetes manifest snippet:
```yaml
readinessProbe:
  httpGet: { path: /health, port: 3006 }
  periodSeconds: 5
  failureThreshold: 3
livenessProbe:
  httpGet: { path: /healthz, port: 3006 }
  periodSeconds: 10
  failureThreshold: 3
```

Liveness is deliberately trivial — a deep check in the liveness probe causes unnecessary restarts when only a dependency is down.

---

## Error Propagation Chain

Adapter-level errors are translated into domain sentinels at the adapter boundary. The use-case layer never sees `pgx`, `redis`, or `googlemaps` types.

```
┌───────────────┐   wrap    ┌──────────────────────┐  pass-through  ┌────────────┐  mapDomainError  ┌──────────┐
│ External lib  │ ────────► │ Adapter (driven)     │ ─────────────► │ Use-case   │ ───────────────► │ Handler  │
│ pgx / redis / │           │ returns domain error │                │ returns    │                  │ → HTTP   │
│ googlemaps    │           └──────────────────────┘                └────────────┘                  └──────────┘
└───────────────┘
```

| Adapter | Library error | Mapped domain sentinel |
|---|---|---|
| `PostgresVenueRepository` | `pgx.ErrNoRows` | `domain.ErrVenueNotFound` |
| `PostgresVenueRepository` | unique-violation (SQLSTATE 23505) | `domain.ErrAlreadySaved` |
| `RedisCacheAdapter` | `redis.Nil` | `nil` return — not an error, just a miss |
| `RedisCacheAdapter` | timeout / conn refused | logged, swallowed (best-effort) |
| `GoogleMapsAdapter` | HTTP 429 / 5xx | `domain.ErrProviderUnavailable` |
| `GoogleMapsAdapter` | ZERO_RESULTS | `domain.ErrVenueNotFound` |

```go
// Example: inside PostgresVenueRepository.SaveFavourite
_, err := q.InsertSavedVenue(ctx, params)
if err != nil {
    if pgerrcode.IsUniqueViolation(err) {
        return domain.NewDomainError(domain.ErrAlreadySaved, cmd.PlaceID)
    }
    return fmt.Errorf("save favourite: %w", err) // unexpected → 500
}
```

**Rule:** If an adapter cannot map a library error to a known sentinel, it wraps with `fmt.Errorf` so the handler's catch-all branch returns 500.

---

## Middleware Chain

Middlewares are registered in `server.go` in this exact order — execution flows top-to-bottom on request, bottom-to-top on response:

```
request ──► request_id ──► logging ──► recover ──► route handler ──► response
```

| Order | Middleware | File | Responsibility |
|---|---|---|---|
| 1 | `RequestID` | `middleware/request_id.go` | Reads `X-Request-Id` from the incoming header; generates `uuid.New()` if absent. Stores it in `c.Locals("request_id")` and sets it on the response header. |
| 2 | `Logging` | `middleware/logging.go` | Logs method, path, status, and latency as structured zap fields on response. Reads `request_id` from locals. |
| 3 | `Recover` | Fiber built-in `recover.New()` | Catches panics in handlers. Logs the stack trace and returns `500` with a generic error response. |

```go
app := fiber.New(fiber.Config{ErrorHandler: mapDomainError})

app.Use(middleware.RequestID())
app.Use(middleware.Logging(log))
app.Use(recover.New())

registerRoutes(app, venueH, favH, historyH)
```

**Authentication is not in this chain.** JWT validation is handled upstream by the API gateway. The gateway injects two trusted headers that handlers read directly:
- `X-User-ID` — the authenticated user
- `X-Squad-ID` — the active squad context

---

## Observability

### Structured Logging

All log output is JSON via `zap.NewProduction()`. Every HTTP request log line includes:

| Field | Source | Example |
|---|---|---|
| `request_id` | `middleware.RequestID` | `"b7e4c2a1-..."` |
| `squad_id` | `X-Squad-ID` header | `"squad_abc"` |
| `user_id` | `X-User-ID` header | `"user_123"` |
| `method` | Fiber context | `"GET"` |
| `path` | Fiber context | `"/venues/search"` |
| `status` | Fiber context (post-handler) | `200` |
| `latency_ms` | `time.Since(start)` | `42` |
| `error` | Only on 4xx/5xx | `"venue not found"` |

Adapter-level logs use the same logger with `log.With(zap.String("adapter", "google_maps"))` so they inherit request-scoped fields when the logger is passed via constructor.

### Distributed Tracing (future)

No tracing SDK is integrated yet. To preserve trace continuity through the API gateway, handlers propagate the `X-Trace-Id` header as-is when calling external services. When OpenTelemetry is added, this header becomes the parent span ID.

### Metrics (future)

Not yet instrumented. Planned: Prometheus `/metrics` endpoint via `fiberprometheus` with RED metrics (rate, errors, duration) per route.

---

## Caching Strategy

The `RedisCacheAdapter` implements `domain.VenueCache`. All cache operations are best-effort — see Key Design Rule 5.

### Cache Keys and TTLs

| Data | Key pattern | TTL | Rationale |
|---|---|---|---|
| Nearby search | `nearby:{lat}:{lng}:{radiusM}:{query}` | 24 h | Venue lists change slowly; avoids redundant Google API calls |
| Venue detail | `detail:{placeID}` | 1 h | Hours/reviews change more often than search results |

Coordinates in the key are truncated to 4 decimal places (~11 m precision) to improve hit rates for nearly-identical requests.

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

### Invalidation

There is no active invalidation — caches expire via TTL. `InvalidateDetail` exists on the port for future use (e.g. when a visit is recorded and the detail may have stale metadata).

### Serialisation

Values are stored as JSON via `encoding/json`. Keys are unnamespaced beyond their type prefix because this Redis instance is dedicated to venue-service.

---

## Testing Strategy

Hexagonal architecture makes each layer independently testable by controlling the boundary:

| Layer | Test type | Technique | External deps |
|---|---|---|---|
| `domain/` | Unit | Standard `go test`. Pure structs and error constructors, zero imports beyond `fmt`. | None |
| `application/` | Unit | Inject **in-memory fakes** implementing `PlaceProvider`, `VenueCache`, `VenueRepository`. Assert output and side-effects. | None |
| `adapters/places/` | Integration | Stub HTTP server or live sandbox key behind `//go:build integration`. | Google Maps API |
| `adapters/cache/` | Integration | `testcontainers-go` spins up a Redis container per test suite. | Docker |
| `adapters/persistence/` | Integration | `testcontainers-go` spins up Postgres, applies migrations, runs sqlc queries. | Docker |
| `transport/http/` | Unit / E2E | `fiber.Test()` against the Fiber app with fake use-cases injected. Asserts status codes, JSON shape, headers. | None |

```go
// Example: testing VenueUseCase with fakes
func TestSearch_CacheMiss_CallsProvider(t *testing.T) {
    fakeCache    := &FakeVenueCache{GetNearbyFn: func(...) ([]domain.Venue, error) { return nil, nil }}
    fakeProvider := &FakeProvider{SearchNearbyFn: func(...) ([]domain.Venue, error) { return testVenues, nil }}

    uc := application.NewVenueUseCase(fakeProvider, fakeCache)
    result, err := uc.Search(ctx, params)

    require.NoError(t, err)
    assert.Equal(t, testVenues, result)
    assert.True(t, fakeCache.SetNearbyCalled) // verify cache was populated
}
```

**Conventions:**
- Unit tests live alongside source: `venue_usecase_test.go` next to `venue_usecase.go`.
- Integration tests use `//go:build integration` so `go test ./...` skips them by default.
- CI runs `go test ./...` first, then `go test -tags=integration ./...` with Docker available.

---

## Key Design Rules

1. **Interfaces live in `domain/ports.go`**, not in adapter packages — the consumer (use-case) owns the contract
2. **DTOs live in `transport/http/dto/`** — domain structs never carry `json:` tags or HTTP concerns
3. **`mapDomainError` lives in the handler** — domain never knows about HTTP status codes
4. **sqlc types never cross into application/domain** — repo adapter maps them at the boundary
5. **Cache writes are best-effort** — `_ = cache.Set(...)` so a cache failure never degrades an API response
