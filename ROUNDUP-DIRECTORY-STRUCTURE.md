# Roundup — Complete Directory Structure

> **Language policy:** Backend is Go only. Mobile app is Compose Multiplatform (Kotlin). No other languages.

```
roundup/
├── .github/
│   ├── workflows/
│   │   ├── ci.yml                        # gofmt, vet, golangci-lint, go test + docker build on every PR; `ci-ok` gates main
│   │   ├── deploy.yml                    # build + push + deploy on main merge
│   │   └── proto.yml                     # regenerate Go stubs on proto changes
│   ├── CODEOWNERS
│   └── pull_request_template.md
│
├── .githooks/
│   └── pre-commit                        # gofmt + go vet + golangci-lint on staged files
│
├── proto/                                # shared gRPC definitions (Go stubs only)
│   ├── buf.yaml
│   ├── buf.gen.yaml
│   ├── auth/
│   │   └── auth.proto
│   ├── user/
│   │   └── user.proto
│   ├── billing/
│   │   └── billing.proto
│   └── flags/
│       └── flags.proto
│
├── shared/                               # module github.com/SamPariat/roundup/shared
│   ├── env/                              # .env.<APP_ENV> loading + typed getters
│   ├── logger/                           # zap logger factory
│   ├── middleware/                       # RequestID, Logging, UserID (Fiber)
│   ├── postgres/                         # pool setup + Postgres error helpers
│   ├── response/                         # JSON response envelope
│   └── kafka/                            # topic name constants
│
├── infra/
│   ├── docker/
│   │   └── prometheus.yml               # Prometheus scrape config
│   ├── k8s/                             # Kubernetes manifests (future)
│   │   └── .gitkeep
│   └── terraform/                       # GCP infra as code (future)
│       └── .gitkeep
│
├── scripts/
│   ├── setup.sh                         # first-time dev environment setup
│   ├── start-all.sh                     # start all services
│   ├── migrate-all.sh                   # run goose migrations on all DBs
│   └── seed.sh                          # seed dev databases
│
├── docs/
│   ├── ARCHITECTURE.md                  # system overview, service map
│   ├── CONTRIBUTING.md                  # branching, PR, review process
│   ├── RUNBOOK.md                       # ops procedures
│   └── adr/                             # Architecture Decision Records
│       ├── 001-microservices.md
│       ├── 002-kafka-vs-redis-pubsub.md
│       └── 003-go-only-backend.md
│
├── mobile/                              # [Kotlin — Compose Multiplatform]
│   ├── composeApp/
│   │   └── src/
│   │       ├── commonMain/              # shared UI, view models, Ktor API client
│   │       ├── androidMain/             # Android-specific (Keystore, FCM setup)
│   │       └── iosMain/                 # iOS-specific (Keychain)
│   ├── iosApp/                          # Xcode project wrapper
│   ├── build.gradle.kts
│   └── settings.gradle.kts
│
├── services/                            # every service is Go — see template below
│   ├── gateway/
│   ├── auth-service/
│   ├── user-service/
│   ├── event-service/
│   ├── billing-service/
│   ├── venue-service/
│   ├── notification-service/
│   ├── media-service/
│   ├── realtime-service/
│   ├── reservation-service/
│   ├── analytics-service/
│   ├── feature-flag-service/
│   └── admin-service/
│
├── go.work                              # Go workspace — lists every service + shared
├── docker-compose.yml                   # databases + infra
├── docker-compose.override.yml          # local overrides (gitignored)
├── Makefile                             # convenience commands
├── .gitignore
├── .editorconfig
└── README.md
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
│   ├── application/                     # use cases — depend only on domain ports
│   │   ├── <feature>_usecase.go
│   │   └── <feature>_usecase_test.go
│   ├── adapters/                        # implement domain ports
│   │   └── persistence/
│   │       ├── postgres_<name>_repository.go
│   │       └── db/                      # sqlc-generated code
│   └── transport/
│       └── http/
│           ├── router.go                # all routes under /api/v1
│           ├── <feature>_handler.go
│           └── dtos/
├── db/
│   ├── migrations/                      # goose: 0001_<desc>.sql, 0002_<desc>.sql, ...
│   └── queries/                         # sqlc input
├── sqlc.yaml
├── go.mod
├── go.sum
├── .air.toml                            # live reload
├── Dockerfile
├── .env.beta.example                    # template for .env.beta (local + beta)
└── .env.prod.example                    # template for .env.prod
```

---

## Service-specific files

Only the parts that differ from the template are listed.

### gateway

```
internal/
├── config/config.go                     # upstream URLs, JWT_PUBLIC_KEY, timeout
├── proxy/proxy.go                       # forward to an upstream (path, query, X-Request-Id)
├── middleware/identity.go               # StripIdentity + RequireAuth (Ed25519, sets X-User-ID)
└── transport/http/router.go             # public vs protected routes → upstream per prefix
```

### auth-service

```
internal/
├── domain/
│   ├── otp.go                           # OTPCode + OTP limits
│   ├── session.go                       # refresh tokens, TokenPair, LoginResult
│   └── port.go                          # OTPRepository, SessionRepository, OTPSender, UserDirectory
├── application/
│   ├── otp_usecase.go                   # request + verify OTP → login
│   ├── token_usecase.go                 # refresh (rotation + reuse detection), logout
│   └── tokens.go                        # TokenIssuer: Ed25519 access JWTs + refresh tokens
├── adapters/
│   ├── sms/console_sender.go            # beta — logs the code (prod SMS provider: later)
│   ├── users/user_service_client.go     # POST /api/v1/users on user-service
│   └── persistence/                     # OTP + session repositories (sqlc)
└── transport/http/auth_handler.go
db/migrations/
├── 0001_create_otp_codes.sql
└── 0002_create_sessions.sql
```

### user-service

```
internal/
├── domain/
│   ├── user.go
│   ├── device_token.go
│   ├── squad.go
│   └── port.go
├── application/
│   ├── user_usecase.go
│   ├── device_token_usecase.go
│   ├── squad_usecase.go
│   └── friend_usecase.go
└── transport/http/
    ├── user_handler.go
    ├── device_token_handler.go
    └── squad_handler.go
db/migrations/
├── 0001_create_users.sql
├── 0002_create_device_tokens.sql
├── 0003_create_squads.sql
├── 0004_create_squad_members.sql
├── 0005_create_user_face_profiles.sql
└── 0006_create_friend_connections.sql
```

### event-service

```
internal/
├── domain/
│   ├── event.go
│   ├── event_state.go                   # DRAFT → VOTING → CONFIRMED → ACTIVE → CLOSED
│   ├── member.go
│   ├── vote.go
│   └── invite.go
├── application/
│   ├── event_usecase.go
│   ├── rsvp_usecase.go
│   ├── voting_usecase.go
│   ├── invite_usecase.go
│   └── guest_usecase.go
└── adapters/kafka/
    ├── publisher.go
    └── consumer.go
db/migrations/
├── 0001_create_events.sql
├── 0002_create_event_members.sql
├── 0003_create_votes.sql
└── 0004_create_invite_links.sql
```

### billing-service

```
internal/
├── domain/
│   ├── bill.go
│   ├── debt.go
│   ├── command.go                       # BillCommand interface
│   ├── split_strategy.go                # SplitStrategy interface
│   └── debt_graph.go                    # min-cash-flow algorithm
├── application/
│   ├── bill_usecase.go
│   ├── debt_usecase.go
│   ├── commands/                        # AddItem, RemoveItem, AssignItem, CloseBill, invoker
│   └── splits/                          # EqualSplit, ByItemSplit, CustomSplit
└── adapters/kafka/
    ├── publisher.go
    └── consumer.go                      # event.closed → close bill
db/migrations/
├── 0001_create_bills.sql
├── 0002_create_bill_items.sql
├── 0003_create_item_assignments.sql
├── 0004_create_debts.sql
├── 0005_create_audit_log.sql
└── 0006_create_split_configs.sql
```

### venue-service (built)

See `services/venue-service/ARCHITECTURE.md`.

### notification-service

```
internal/
├── adapters/
│   ├── fcm/client.go
│   ├── whatsapp/client.go
│   └── kafka/consumer.go                # all Kafka topic listeners
└── application/notification_router.go   # decides FCM vs WhatsApp
db/migrations/
└── 0001_create_notification_log.sql
```

### media-service

```
internal/
├── adapters/
│   ├── cloudinary/client.go
│   ├── vision/client.go                 # Google Vision API
│   └── kafka/tagging_worker.go          # photo.uploaded → Vision → match
├── application/
│   ├── media_usecase.go
│   ├── qr_usecase.go                    # JWT signing + QR PNG
│   └── zip_streamer.go                  # stream ZIP without buffering
db/migrations/
├── 0001_create_photos.sql
├── 0002_create_photo_faces.sql
└── 0003_create_download_tokens.sql
```

### realtime-service

```
internal/
├── hub/
│   ├── hub.go                           # connection registry (sync.RWMutex)
│   └── connection.go
├── presence/presence.go                 # Redis SET per event
├── adapters/kafka/consumer.go           # fan-out Kafka → WebSocket
├── domain/message.go                    # WsMessage + type constants
└── transport/http/ws_handler.go         # WebSocket upgrade endpoint
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

```
internal/
├── adapters/kafka/consumer.go           # ingests domain events
└── application/
    ├── squad_stats_usecase.go
    ├── user_stats_usecase.go
    └── wrapped_usecase.go
db/migrations/
├── 0001_create_analytics_events.sql
├── 0002_squad_stats_view.sql
├── 0003_user_stats_view.sql
├── 0004_venue_stats_view.sql
└── 0005_monthly_summary_view.sql
```

### feature-flag-service

```
internal/
├── application/flag_evaluator.go        # hash-based rollout
├── adapters/cache/flag_cache.go         # Redis cache (60s TTL)
└── transport/http/flag_handler.go       # REST — admin only
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
    ├── venues_handler.go
    └── abuse_handler.go
```

---

## Service → Port map

| Service              | Language | Internal Port | DB Port |
| -------------------- | -------- | ------------- | ------- |
| gateway              | Go       | 8080          | —       |
| auth-service         | Go       | 3001          | 5433    |
| user-service         | Go       | 3002          | 5434    |
| event-service        | Go       | 3003          | 5435    |
| billing-service      | Go       | 3004          | 5436    |
| venue-service        | Go       | 3006          | 5438    |
| notification-service | Go       | 3007          | —       |
| media-service        | Go       | 3008          | 5439    |
| realtime-service     | Go       | 3009          | —       |
| reservation-service  | Go       | 3011          | 5442    |
| analytics-service    | Go       | 3012          | 5441    |
| feature-flag-service | Go       | 3013          | 5440    |
| admin-service        | Go       | 3014          | —       |

_All services share Redis on 6379 and Kafka on 9092. All HTTP routes are versioned under `/api/v1`._

> **Later:** AI features — deferred, will be revisited.
