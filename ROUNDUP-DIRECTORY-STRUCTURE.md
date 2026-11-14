# Roundup — Directory Structure

> **Language policy:** Backend is Go only. The client is one React Native + Expo app (TypeScript) for web, Android and iOS, built side by side in the separate `roundup-ui` repo.
>
> Services talk over REST, and each owns its own Postgres database. No Kafka and no gRPC.

```
roundup/
├── .github/
│   └── workflows/
│       ├── ci.yml                        # gofmt, vet, golangci-lint, go test + docker build on every PR; `ci-ok` gates main
│       └── deploy.yml                    # (planned) build + push + deploy to Cloud Run on main merge
│
├── .githooks/
│   └── pre-commit                        # gofmt + go vet + golangci-lint on staged files
│
├── shared/                               # module github.com/SamPariat/roundup/shared
│   ├── api/                              # cross-service contract: route prefixes, shared bodies, openapi.yaml (served as Swagger UI)
│   ├── authtoken/                        # Ed25519 access JWTs: sign (auth) and verify (gateway)
│   ├── authz/                            # Casbin model + Enforcer; each service supplies its policy.csv
│   ├── domainerr/                        # wrap a service's sentinel errors with context, errors.Is still matches
│   ├── env/                              # .env.<APP_ENV> loading + typed getters
│   ├── logger/                           # zap logger factory
│   ├── middleware/                       # RequestID, Logging, Recover, UserID (Fiber)
│   ├── pagination/                       # cursor pagination
│   ├── phone/                            # phone number normalisation
│   ├── postgres/                         # pool setup + Postgres error helpers
│   ├── response/                         # JSON response envelope
│   ├── server/                           # Run: health/ready probes, graceful shutdown
│   ├── serviceclient/                    # HTTP client for service-to-service calls
│   └── userclient/                       # user-service client: squad members (with role) and squads
│
├── scripts/
│   └── setup.sh                          # first-time dev environment setup (`make setup`)
│
├── assets/brand/                         # logo, banner and mark files
│
├── services/                             # every service is Go — see template below
│   ├── gateway/                          # built
│   ├── auth-service/                     # built
│   ├── user-service/                     # built
│   ├── event-service/                    # built
│   ├── venue-service/                    # built
│   ├── billing-service/                  # next
│   ├── scheduler-service/                # next
│   ├── notification-service/             # planned
│   ├── media-service/                    # planned
│   ├── realtime-service/                 # planned
│   ├── reservation-service/              # planned
│   ├── analytics-service/                # planned
│   ├── feature-flag-service/             # planned
│   └── admin-service/                    # planned
│
├── go.work                               # Go workspace — lists every service + shared
├── docker-compose.yml                    # one Postgres per service, Redis, and the services (`services` profile)
├── Makefile                              # convenience commands
├── .golangci.yml
├── .editorconfig
├── .gitignore
├── README.md
├── SCOPE.md                              # product + technical scope
└── ROUNDUP-TASKS.md                      # backend task list
```

---

## Go service template

Every service follows the hexagonal layout already used by `services/venue-service` (see its `ARCHITECTURE.md`).

```
services/<name>/
├── cmd/
│   └── main.go                          # wiring only — config → adapters → use cases → handlers
├── internal/
│   ├── config/
│   │   └── config.go                    # env-var loading via shared/env
│   ├── domain/                          # structs, errors, port interfaces — no external imports
│   │   ├── <entity>.go
│   │   ├── errors.go
│   │   └── port.go
│   ├── policy/policy.csv                # Casbin permissions, when the service has roles
│   ├── application/                     # use cases — depend only on domain ports
│   │   ├── <feature>_usecase.go
│   │   └── <feature>_usecase_test.go
│   ├── adapters/                        # implement domain ports
│   │   └── persistence/
│   │       ├── postgres_<name>_repository.go
│   │       └── db/                      # sqlc-generated code
│   └── transport/
│       └── http/
│           ├── router.go                # public routes under /api/v1, service-to-service under /internal/v1
│           ├── <feature>_handler.go
│           └── dtos/
├── db/
│   ├── migrations/                      # goose: 0001_<desc>.sql, 0002_<desc>.sql, ...
│   └── queries/                         # sqlc input
├── sqlc.yaml
├── go.mod
├── go.sum
├── ARCHITECTURE.md
├── Dockerfile
├── .env.beta.example                    # template for .env.beta (local + beta)
└── .env.prod.example                    # template for .env.prod
```

---

## Service-specific files

Only the parts that differ from the template are listed. Built services are documented in their own `ARCHITECTURE.md`.

### gateway (built)

```
internal/
├── config/config.go                     # upstream URLs, JWT_PUBLIC_KEY, timeout, CORS origins
├── proxy/proxy.go                       # forward to an upstream (path, query, X-Request-Id)
├── middleware/identity.go               # StripIdentity + RequireAuth (Ed25519, sets X-User-ID)
└── transport/http/router.go             # public vs protected routes → upstream per prefix; only /api/v1 is mounted
```

### auth-service, user-service, event-service, venue-service (built)

See each service's `ARCHITECTURE.md`.

### billing-service (next)

```
internal/
├── domain/
│   ├── bill.go                          # one payer, items, who shares it
│   ├── split.go                         # equal split in paise, leftover paise in a fixed order
│   ├── netting.go                       # all of a night's bills → one debt per pair
│   └── debt.go                          # PENDING → SETTLED
├── application/
│   ├── bill_usecase.go
│   └── debt_usecase.go
└── adapters/
    ├── persistence/
    └── receipts/                        # ReceiptScanner (OCR provider still to be chosen)
db/migrations/
├── 0001_create_bills.sql
├── 0002_create_bill_items.sql
├── 0003_create_bill_shares.sql
└── 0004_create_debts.sql
```

### scheduler-service (next)

No database and no domain logic.

```
internal/
├── config/config.go                     # target URLs + interval per job
└── jobs/                                # each job POSTs to one /internal/v1 endpoint
    ├── event_confirm_due.go
    └── event_close_due.go
```

### notification-service

```
internal/
├── adapters/
│   ├── fcm/client.go
│   └── whatsapp/client.go
├── application/notification_router.go   # decides FCM vs WhatsApp, quiet hours
└── transport/http/                      # POST /internal/v1/notifications
db/migrations/
└── 0001_create_notification_log.sql
```

### media-service

Face tagging is deferred.

```
internal/
├── adapters/cloudinary/client.go
├── application/
│   ├── media_usecase.go
│   ├── download_usecase.go              # signed download token + QR PNG
│   └── zip_streamer.go                  # stream ZIP without buffering
db/migrations/
├── 0001_create_photos.sql
└── 0002_create_download_tokens.sql
```

### realtime-service

```
internal/
├── hub/
│   ├── hub.go                           # connection registry (sync.RWMutex)
│   └── connection.go
├── domain/message.go                    # WsMessage + type constants
└── transport/http/
    ├── ws_handler.go                    # WebSocket upgrade endpoint
    └── publish_handler.go               # POST /internal/v1/realtime/publish
```

### reservation-service

```
internal/
├── domain/reservation_state.go          # PENDING → CONFIRMED → CANCELLED → NO_SHOW
├── adapters/dineout/client.go           # Dineout / EazyDiner API client
└── application/reservation_usecase.go
db/migrations/
├── 0001_create_reservations.sql
└── 0002_create_reservation_events.sql
```

### analytics-service

Reads from the other services' internal endpoints; a scheduler-service job refreshes it nightly.

```
internal/
└── application/
    ├── squad_stats_usecase.go
    ├── user_stats_usecase.go
    └── wrapped_usecase.go
```

### feature-flag-service

```
internal/
├── application/flag_evaluator.go        # override → global → hash-based rollout
└── transport/http/
    ├── flag_handler.go                  # REST — admin only
    └── evaluate_handler.go              # POST /internal/v1/flags/evaluate
db/migrations/
├── 0001_create_feature_flags.sql
└── 0002_create_user_overrides.sql
```

### admin-service

```
internal/
├── middleware/
│   ├── admin_auth.go                    # admin JWT
│   └── roles.go                         # SUPER_ADMIN / OPS / SUPPORT
└── transport/http/
    ├── users_handler.go
    ├── flags_handler.go
    └── abuse_handler.go
```

---

## Service → Port map

| Service              | Status  | Internal Port | DB Port |
| -------------------- | ------- | ------------- | ------- |
| gateway              | built   | 8080          | —       |
| auth-service         | built   | 3001          | 5433    |
| user-service         | built   | 3002          | 5434    |
| event-service        | built   | 3003          | 5435    |
| billing-service      | next    | 3004          | 5436    |
| scheduler-service    | next    | 3005          | —       |
| venue-service        | built   | 3006          | 5438    |
| notification-service | planned | 3007          | 5437    |
| media-service        | planned | 3008          | 5439    |
| realtime-service     | planned | 3009          | —       |
| reservation-service  | planned | 3011          | 5442    |
| analytics-service    | planned | 3012          | —       |
| feature-flag-service | planned | 3013          | 5440    |
| admin-service        | planned | 3014          | —       |

_Redis on 6379 is used only by venue-service's best-effort cache locally; production launches without it. All public HTTP routes are versioned under `/api/v1`; service-to-service routes live under `/internal/v1` and are never exposed by the gateway._

> **Later:** AI features are deferred, except receipt scanning for bills.
