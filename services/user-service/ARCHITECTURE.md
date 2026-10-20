# user-service: Hexagonal Architecture LLD

## Context
The user-service owns user profiles and the FCM device tokens used for push notifications. Login lives in auth-service: after a phone OTP is verified, auth-service calls `POST /api/v1/users`, and both services share the resulting user ID. The service follows the same hexagonal layout as venue-service — the domain sits at the centre, port interfaces are declared in the domain package, and Postgres and HTTP are pushed out to adapter packages.

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
│  transport/http/     │           │  adapters/persistence/    │
│    UserHandler       │           │    PostgresUserRepository │ ◄─ implements UserRepository
│    DeviceTokenHandler│           │                           │    and DeviceTokenRepository
│    router.go (Fiber) │           └───────────────────────────┘
└──────────┬───────────┘                       ▲
           │ calls                             │ implements
           ▼                                   │
┌──────────────────────────────────────────────┘
│  APPLICATION LAYER  internal/application/
│  UserUseCase        — depends on UserRepository + DeviceTokenRepository (ports)
│  DeviceTokenUseCase — depends on DeviceTokenRepository (port)
└──────────────────────────┬───────────────────
                           │ imports
                           ▼
┌──────────────────────────────────────────────────────────────────┐
│  DOMAIN  internal/domain/   — imports only the standard library  │
│  Structs: User, DeviceToken, Platform, *Command types            │
│  Ports (interfaces owned here, not in adapters):                 │
│    UserRepository        — Create, GetByID, GetByPhone,          │
│                            UpdateProfile, SoftDelete             │
│    DeviceTokenRepository — Upsert, Delete, DeleteAllForUser      │
│  Errors: sentinel vars, wrapped with shared/domainerr            │
└──────────────────────────────────────────────────────────────────┘
All dependency arrows point inward. Domain has no outward arrows.
```

---

## Directory Tree

```
services/user-service/
├── cmd/
│   └── main.go                          # wiring only
│
├── internal/
│   ├── domain/
│   │   ├── user.go                      # User, UpdateProfileCommand, IsValidPhone
│   │   ├── device_token.go              # DeviceToken, Platform, Register/RemoveDeviceTokenCommand
│   │   ├── errors.go                    # sentinel errors (context added with shared/domainerr)
│   │   └── port.go                      # UserRepository, DeviceTokenRepository interfaces
│   │
│   ├── application/
│   │   ├── user_usecase.go              # CreateOrGet (idempotent), Get, UpdateProfile, Delete
│   │   └── device_token_usecase.go      # Register (upsert), Remove
│   │
│   ├── adapters/
│   │   └── persistence/
│   │       ├── postgres_user_repository.go  # implements both domain repositories
│   │       └── db/                      # sqlc-generated (db.go, models.go, *.sql.go)
│   │
│   ├── transport/
│   │   └── http/
│   │       ├── router.go                # route registration under /api/v1
│   │       ├── user_handler.go          # POST /users, GET/PATCH/DELETE /users/me
│   │       ├── device_token_handler.go  # POST/DELETE /users/me/device-tokens
│   │       └── dtos/
│   │           ├── user_dtos.go
│   │           ├── device_token_dtos.go
│   │           └── mappers.go           # domain → response DTOs
│   │
│   └── config/
│       └── config.go                    # env-var loading (APP_ENV, PORT, POSTGRES_DSN, LOG_LEVEL)
│
├── db/
│   ├── migrations/
│   │   ├── 0001_create_users.sql
│   │   └── 0002_create_device_tokens.sql
│   └── queries/
│       ├── users.sql
│       └── device_tokens.sql
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

// UserRepository — all lookups ignore soft-deleted users
type UserRepository interface {
    Create(ctx context.Context, phone string) (*User, error)                 // ErrPhoneTaken on duplicate
    GetByID(ctx context.Context, id string) (*User, error)                   // ErrUserNotFound
    GetByPhone(ctx context.Context, phone string) (*User, error)             // ErrUserNotFound
    UpdateProfile(ctx context.Context, cmd UpdateProfileCommand) (*User, error)
    SoftDelete(ctx context.Context, id string) error
}

// DeviceTokenRepository — a token identifies a device, so it is unique across users
type DeviceTokenRepository interface {
    Upsert(ctx context.Context, cmd RegisterDeviceTokenCommand) (*DeviceToken, error)
    Delete(ctx context.Context, cmd RemoveDeviceTokenCommand) error          // ErrDeviceTokenNotFound
    DeleteAllForUser(ctx context.Context, userID string) error
}
```

---

## Domain Models

### `internal/domain/user.go`

| Field | Type | Notes |
|---|---|---|
| `ID` | `string` | UUID, shared with auth-service |
| `Phone` | `string` | E.164 format, e.g. `+919876543210` |
| `Name` | `string` | Empty until profile setup |
| `AvatarURL` | `string` | Empty until an avatar is uploaded |
| `CreatedAt`, `UpdatedAt` | `time.Time` | UTC |

`UpdateProfileCommand` uses `*string` fields: `nil` means "leave unchanged".

### `internal/domain/device_token.go`

| Field | Type | Notes |
|---|---|---|
| `ID` | `int64` | Database primary key |
| `UserID` | `string` | User currently logged in on the device |
| `Token` | `string` | FCM registration token |
| `Platform` | `Platform` | `PlatformIOS` or `PlatformAndroid` |
| `CreatedAt` | `time.Time` | UTC |

---

## Adapter Responsibilities

| Adapter | File | Implements |
|---|---|---|
| `PostgresUserRepository` | `adapters/persistence/postgres_user_repository.go` | `domain.UserRepository`, `domain.DeviceTokenRepository` |

Domain IDs are plain strings. The adapter parses them into `uuid.UUID` for sqlc and maps sqlc rows (with nullable `*string` columns) back into domain structs, so sqlc and pgx types never reach the application layer.

---

## main.go Wiring (outline)

```go
// 1. Load config
cfg := config.Load()

// 2. Infrastructure clients
pool, err := postgres.Connect(ctx, cfg.PostgresDSN) // shared/postgres — pings at startup

// 3. Driven adapters
repo := persistence.NewPostgresUserRepository(pool)

// 4. Use-cases (depend on ports, not concrete adapters)
userUC        := application.NewUserUseCase(repo, repo)
deviceTokenUC := application.NewDeviceTokenUseCase(repo)

// 5. Driving adapters
app := fiber.New()
app.Use(middleware.RequestID(), middleware.Logging(log))
transporthttp.RegisterRoutes(app, &transporthttp.Handlers{...})

// 6. Run
app.Listen(":" + cfg.Port) // 3002
```

---

## HTTP Routes

The full request/response contract is in [`shared/api/openapi.yaml`](../../shared/api/openapi.yaml) (Swagger UI at `localhost:8080/docs`); this table is an overview.

All routes are versioned under `/api/v1`. Routes under `/users/me` require the `X-User-ID` header (set by the gateway after token validation) and are wrapped in `middleware.UserID()`.

| Method | Path | Handler | Use-case | Success |
|---|---|---|---|---|
| POST | /api/v1/users | UserHandler.Create | UserUseCase.CreateOrGet | 201 new / 200 existing |
| GET | /api/v1/users/me | UserHandler.GetMe | UserUseCase.Get | 200 |
| PATCH | /api/v1/users/me | UserHandler.UpdateMe | UserUseCase.UpdateProfile | 200 |
| DELETE | /api/v1/users/me | UserHandler.DeleteMe | UserUseCase.Delete | 204 |
| POST | /api/v1/users/me/device-tokens | DeviceTokenHandler.Register | DeviceTokenUseCase.Register | 201 |
| DELETE | /api/v1/users/me/device-tokens/:token | DeviceTokenHandler.Remove | DeviceTokenUseCase.Remove | 204 |

`POST /api/v1/users` is internal — called by auth-service, never exposed by the gateway to the mobile app.

---

## Error Mapping (transport/http layer only)

| Domain sentinel | HTTP status |
|---|---|
| `ErrInvalidPhone`, `ErrInvalidName`, `ErrInvalidToken`, `ErrInvalidPlatform` | 400 |
| missing / invalid `X-User-ID` header (middleware) | 401 |
| `ErrUserNotFound`, `ErrDeviceTokenNotFound` | 404 |
| `ErrPhoneTaken` | 409 |
| anything else | 500 |

---

## Database Schema

```sql
-- users
CREATE TABLE users (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    phone      TEXT        NOT NULL,
    name       TEXT,
    avatar_url TEXT,
    deleted_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
-- phone is unique among active users only, so a deleted account's number can sign up again
CREATE UNIQUE INDEX uq_users_phone_active ON users (phone) WHERE deleted_at IS NULL;

-- device_tokens
CREATE TABLE device_tokens (
    id         BIGSERIAL   PRIMARY KEY,
    user_id    UUID        NOT NULL REFERENCES users (id),
    token      TEXT        NOT NULL UNIQUE,
    platform   TEXT        NOT NULL CHECK (platform IN ('iOS', 'Android')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

---

## Error Propagation Chain

Library errors are translated into domain sentinels at the adapter boundary. The use-case layer never sees `pgx` types.

| Adapter | Library error | Mapped domain sentinel |
|---|---|---|
| `PostgresUserRepository` | `pgx.ErrNoRows` (`postgres.IsNoRows`) | `domain.ErrUserNotFound` |
| `PostgresUserRepository` | unique violation (`postgres.IsUniqueViolation`) | `domain.ErrPhoneTaken` |
| `PostgresUserRepository` | 0 rows affected on soft delete | `domain.ErrUserNotFound` |
| `PostgresUserRepository` | 0 rows affected on token delete | `domain.ErrDeviceTokenNotFound` |
| `PostgresUserRepository` | ID that is not a valid UUID | `domain.ErrUserNotFound` |

`UserUseCase.CreateOrGet` recovers from one of these itself: if `Create` returns `ErrPhoneTaken`, a concurrent request inserted the same phone first, so it looks the user up again and returns it with `created = false`.

---

## Middleware Chain

```
request ──► RequestID ──► Logging ──► [UserID on /users/me/*] ──► route handler ──► response
```

| Middleware | Source | Responsibility |
|---|---|---|
| `RequestID` | `shared/middleware` | Reads or generates `X-Request-Id`, stores it in locals, echoes it on the response |
| `Logging` | `shared/middleware` | Logs method, path, status, latency, and request ID |
| `UserID` | `shared/middleware` | Requires a UUID in `X-User-ID`, returns 401 otherwise, stores it for `middleware.UserIDFrom(c)` |

---

## Testing Strategy

| Layer | Test type | Technique | External deps |
|---|---|---|---|
| `domain/` | Unit | Pure structs and validation helpers | None |
| `application/` | Unit | Hand-rolled fakes implementing `UserRepository` and `DeviceTokenRepository` | None |
| `adapters/persistence/` | Integration (future) | Real Postgres with migrations applied | Docker |
| `transport/http/` | Unit (future) | `fiber.Test()` with use-cases backed by fakes | None |

Unit tests live alongside source (`user_usecase_test.go` next to `user_usecase.go`). Run with `make user-test`.

---

## Key Design Rules

1. **Interfaces live in `domain/port.go`**, not in adapter packages — the consumer (use-case) owns the contract
2. **DTOs live in `transport/http/dtos/`** — domain structs never carry `json:` tags or HTTP concerns
3. **Status-code mapping lives in the handlers** — domain never knows about HTTP status codes
4. **sqlc types never cross into application/domain** — the repository adapter maps them at the boundary
5. **Identity comes from the gateway** — handlers read the caller via `middleware.UserIDFrom(c)`, never from the request body
