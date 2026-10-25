# Roundup — Full Task Breakdown

> Check off tasks as you complete them. Each checkbox is one atomic unit of work.
> **Legend:** `[Go]` `[Infra]` `[App]`
>
> **Language policy:** Backend is Go only. The client is one React Native + Expo app (TypeScript) for web, Android and iOS. No other languages.
>
> **Later:** AI features — deferred, will be revisited.

---

## Table of contents

- [Infra & DevOps](#infra--devops)
- [API Gateway](#api-gateway--go)
- [Auth Service](#auth-service--go)
- [User Service](#user-service--go)
- [Event Service](#event-service--go)
- [Billing Service](#billing-service--go)
- [Venue Service](#venue-service--go)
- [Notification Service](#notification-service--go)
- [Media Service](#media-service--go)
- [Realtime Service](#realtime-service--go)
- [Reservation Service](#reservation-service--go)
- [Analytics Service](#analytics-service--go)
- [Feature Flag Service](#feature-flag-service--go)
- [Admin Service](#admin-service--go)
- [App](#app-react-native--expo-app)
- [Cross-cutting](#cross-cutting)

---

## Infra & DevOps `[Infra]`

### Local environment

- [x] Install Docker Desktop
- [x] Install Go 1.27+
- [x] Install goose (`make install`)
- [x] Install sqlc (`make install`)
- [x] Install golangci-lint
- [ ] Install Node.js 26 LTS (Expo app); Android Studio + Xcode only for local native builds
- [ ] Install protoc (gRPC compiler)
- [ ] Install buf CLI (protobuf linting)

### Monorepo setup

- [x] Initialise git repository
- [x] Create root `docker-compose.yml`
- [x] Add `.gitignore`
- [x] Add `.editorconfig` for consistent formatting
- [x] Create `services/` directory for all microservices
- [ ] Create `proto/` directory for shared gRPC definitions
- [x] Create `scripts/` for dev helper scripts
- [x] Write `scripts/start-all.sh` to boot all services locally

### Docker Compose — local services

- [x] Add PostgreSQL container (one per service, separate ports)
- [x] Add Redis container (shared, single instance)
- [ ] Add Kafka container (Bitnami image)
- [ ] Add Zookeeper container (Kafka dependency)
- [ ] Add Kafka UI container (Provectus) for local debugging
- [ ] Add Redis Commander container for local Redis browsing
- [ ] Add pgAdmin container for DB browsing
- [x] Set up Docker networking so all services can communicate
- [ ] Write `docker-compose.override.yml` for local-only overrides
- [ ] Add health checks to all containers

### CI/CD

- [x] Create GitHub Actions workflow directory `.github/workflows/`
- [x] Write `ci.yml` — runs on every PR
- [x] Add lint step (golangci-lint)
- [x] Add test step for each service
- [x] Add build step for each service (Docker image per service; `ci-ok` is the required check)
- [ ] Write `deploy.yml` — runs on merge to main
- [ ] Add Docker image build per service
- [ ] Push images to Google Artifact Registry
- [ ] Add Cloud Run deploy step per service
- [ ] Set up environment secrets in GitHub (DB URLs, API keys)
- [ ] Write `CODEOWNERS` file

### Kafka setup

- [ ] Define all Kafka topic names as constants in `shared/kafka`
- [ ] Topic: `rsvp.updated`
- [ ] Topic: `event.created`
- [ ] Topic: `event.closed`
- [ ] Topic: `tab.opened`
- [ ] Topic: `tab.closed`
- [ ] Topic: `debt.created`
- [ ] Topic: `debt.settled`
- [ ] Topic: `photo.uploaded`
- [ ] Topic: `photo.tagged`
- [ ] Set retention policy (7 days) for all topics
- [ ] Set replication factor (3) for production topics

### gRPC proto definitions

- [ ] Create `proto/auth.proto` — token validation RPC
- [ ] Create `proto/user.proto` — get user by ID RPC
- [ ] Create `proto/billing.proto` — get bill, get debts RPCs
- [ ] Create `proto/flags.proto` — evaluate flag RPC
- [ ] Generate Go stubs from all protos
- [ ] Add buf lint check to CI

### GCP production setup

- [ ] Create GCP project
- [ ] Enable Cloud Run API
- [ ] Enable Cloud SQL API
- [ ] Enable Memorystore API
- [ ] Enable Artifact Registry API
- [ ] Enable Secret Manager API
- [ ] Enable Pub/Sub API (if replacing Kafka)
- [ ] Create Cloud SQL Postgres instance (asia-south1)
- [ ] Create Memorystore Redis instance
- [ ] Create Artifact Registry repository
- [ ] Set up service accounts per microservice
- [ ] Configure IAM roles (least privilege)
- [ ] Set up Secret Manager for all env vars
- [ ] Configure VPC connector for Cloud Run → Cloud SQL
- [ ] Set up Cloud Armor for DDoS protection on gateway

### Observability

- [x] Graceful shutdown on SIGTERM and panic recovery on every service (`shared/server.Run`, `shared/middleware.Recover`)

- [ ] Set up Grafana Cloud account (free tier)
- [ ] Configure Loki log shipping (Vector sidecar)
- [ ] Create Grafana dashboard — request rate per service
- [ ] Create Grafana dashboard — error rate per service
- [ ] Create Grafana dashboard — p95 latency per service
- [ ] Create Grafana dashboard — Kafka consumer lag
- [ ] Set up Prometheus metrics endpoint per service
- [ ] Configure alert — error rate > 1% for 5min → Discord
- [ ] Configure alert — service down → Discord
- [ ] Set up Jaeger for distributed tracing
- [ ] Add OpenTelemetry SDK to each service
- [ ] Create trace dashboard in Grafana

---

## API Gateway `[Go]`

Access tokens are Ed25519 JWTs: the gateway verifies them locally with auth-service's **public** key (no per-request call to auth, no shared secret).

### Project setup

- [x] `go mod init github.com/SamPariat/roundup/services/gateway`
- [x] Add Fiber v3 dependency (`gofiber/fiber/v3`)
- [x] Create `cmd/main.go` entry point
- [x] Create `internal/config/` — env var loading with validation
- [x] Create `.env.beta.example` and `.env.prod.example` with all required vars
- [x] Write `Dockerfile` (multi-stage, distroless nonroot)
- [ ] Add `prometheus/client_golang` dependency
- [ ] Add OpenTelemetry Go SDK

### Routing

- [x] Proxy `/api/v1/auth`, `/users`, `/venues`, `/squads` to their upstream service URLs (config per service)
- [x] Keep `/events`, `/billing`, `/media`, `/realtime` as 501 until those services exist
- [x] Block external `POST /api/v1/users` (only auth-service creates users, directly)
- [x] Upstream down or slow → 502
- [x] Request ID generation (UUID v4) and forwarding as `X-Request-Id`
- [x] Forward `X-User-ID` after auth validation

### Auth middleware

- [x] Strip any client-sent `X-User-ID` on every request
- [x] Read `Authorization: Bearer <token>`, verify with `shared/authtoken` (EdDSA only, iss/aud/exp)
- [x] Return 401 if the token is missing or invalid
- [x] Public routes: `/api/v1/auth/*`, `/health`

### Rate limiting

- [x] Rate limit `/api/v1/auth/*` by IP: 20 requests per 15 minutes, in memory per gateway instance (protects OTP sending)
- [ ] Rate limit by user ID for authenticated routes
- [x] Return `429 Too Many Requests` with `Retry-After` header
- [x] Public `GET /api/v1/venues/photos/:ref` (image tags can't send a token), rate limited by IP: 600 per 15 minutes

### Load balancing

- [ ] Support multiple upstream instances per service (or rely on the platform's load balancer)

### Observability

- [x] Log every request: method, path, status, duration, requestId
- [x] Expose `/health` (liveness) and `/ready` (readiness) on every service via `shared/server`
- [ ] Expose `/metrics` endpoint for Prometheus
- [ ] Request count, latency histogram and upstream error count per route

---

## Auth Service `[Go]`

Custom auth — phone OTP + JWT. No Firebase Auth.

### Project setup

- [x] `go mod init github.com/SamPariat/roundup/services/auth-service` and add to `go.work`
- [x] Copy the `services/user-service` layout (`cmd/`, `internal/{domain,application,adapters,transport,config}`, `db/`)
- [x] Add `gofiber/fiber/v3`, `jackc/pgx/v5`, `golang-jwt/jwt/v5` (via `shared/authtoken`)
- [x] Create `internal/config/` — `PORT`, `POSTGRES_DSN`, `USER_SERVICE_URL`, `JWT_PRIVATE_KEY`, TTLs, `OTP_DEV_CODE` via `shared/env`
- [x] Write `.env.beta.example` and `.env.prod.example`
- [x] Write `Dockerfile`
- [x] `make jwt-keys` — Ed25519 keypair (private → auth, public → gateway)

### OTP login

- [x] Define `OTPSender` port in `internal/domain/port.go`
- [x] Console adapter for beta — logs the code; `OTP_DEV_CODE` accepted in beta only
- [ ] SMS provider adapter for production (e.g. MSG91) — auth refuses to start in prod until it exists
- [x] Store OTP hashed with 5-minute expiry and an attempt counter (5 wrong guesses max)
- [x] Rate limit OTP requests per phone number (3 per 15 minutes, counted in Postgres)

### Tokens & sessions

- [x] `0002_create_sessions.sql` — `sessions` (one per login) + `refresh_tokens` (hashed, `used_at`)
- [x] Issue short-lived access JWT (15 min, `sub` = user ID), Ed25519-signed
- [x] Issue refresh token (30 days) — store only its SHA-256 hash
- [x] Rotate refresh token on every refresh; reuse of a used token revokes the session
- [x] Revoke session on logout

### Controllers (REST)

- [x] `POST /api/v1/auth/otp/request` — send code to phone
- [x] `POST /api/v1/auth/otp/verify` — verify code, create user via User Service, return tokens
- [x] `POST /api/v1/auth/refresh` — rotate refresh token, return a new pair
- [x] `POST /api/v1/auth/logout` — revoke session

### Tests

- [x] Unit test OTP verify — wrong code, expired code, too many attempts
- [x] Unit test refresh rotation — old refresh token rejected after use, reuse revokes the session
- [x] Repository tests against real Postgres (`postgrestest`)
- [x] E2E request → verify → refresh → logout (manual, through the gateway)

---

## User Service `[Go]`

### Project setup

- [x] `go mod init github.com/SamPariat/roundup/services/user-service` and add to `go.work`
- [x] Copy the `services/venue-service` layout
- [x] Add `gofiber/fiber/v3`, `jackc/pgx/v5`, `stretchr/testify`
- [x] Create `internal/config/` — `PORT=3002`, `POSTGRES_DSN`, `LOG_LEVEL`
- [x] Add `sqlc.yaml`
- [x] Add `postgres-user` (port 5434) to `docker-compose.yml`
- [x] Add `user-*` targets to the `Makefile`
- [x] Write `.env.beta.example` and `.env.prod.example`
- [x] Write `Dockerfile`

### Database schema (goose migrations)

- [x] `0001_create_users.sql` — id, phone, email, name, avatar_url, deleted_at, created_at, updated_at
- [x] `0002_create_device_tokens.sql` — id, user_id, token, platform, created_at
- [x] `0003_create_squads.sql` — id, name, avatar_url, created_by, created_at
- [x] `0004_create_squad_members.sql` — squad_id, user_id, role, joined_at
- [ ] `0005_create_user_face_profiles.sql` — user_id, cloudinary_id, embeddings (jsonb), updated_at
- [ ] `0006_create_friend_connections.sql` — user_id_a, user_id_b, created_at (unique pair)
- [x] Add indexes: users(phone), squad_members(squad_id, user_id)

### Domain & repositories

- [x] `domain/user.go`, `domain/device_token.go`, `domain/squad.go`, `domain/errors.go`
- [x] `domain/port.go` — `UserRepository`, `DeviceTokenRepository`, `SquadRepository`
- [x] sqlc queries in `db/queries/*.sql`
- [x] Postgres repository adapters in `internal/adapters/persistence/`

### Use cases

- [x] `UserUseCase.CreateOrGet(phone)` — idempotent by phone
- [x] `UserUseCase.Get(id)` — excludes soft-deleted
- [x] `UserUseCase.UpdateProfile(id, name, avatarURL)`
- [x] `UserUseCase.Delete(id)` — sets deleted_at
- [x] `DeviceTokenUseCase.Register(userID, token, platform)` — upsert by token
- [x] `DeviceTokenUseCase.Remove(userID, token)`
- [ ] `DeviceTokenUseCase.ListForUsers(userIDs)` — batch for notification service
- [x] `SquadUseCase` — create, get, edit, add/remove member, change role, list my squads (with member previews)
- [x] Squad permissions from a Casbin policy (`internal/policy/policy.csv`, loaded into `shared/authz`); see `services/user-service/docs/PERMISSIONS.md`
- [x] `UserUseCase.Lookup(phones)` — find which contacts are on Roundup
- [ ] `FriendUseCase` — connect, list friends

### Controllers (REST)

- [x] `POST /api/v1/users` — create or return user by phone (called by Auth Service)
- [x] `GET /api/v1/users/me` — get own profile
- [x] `PATCH /api/v1/users/me` — update name, avatar
- [x] `DELETE /api/v1/users/me` — soft delete account
- [ ] `GET /api/v1/users/:id` — get user by ID (internal use)
- [x] `POST /api/v1/users/me/device-tokens` — register FCM token
- [x] `DELETE /api/v1/users/me/device-tokens/:token` — deregister token
- [x] `POST /api/v1/squads` — create squad
- [x] `GET /api/v1/squads` — get my squads
- [x] `GET /api/v1/squads/:id` — get squad with members, my role, `isMe` and `permissions`
- [x] `PATCH /api/v1/squads/:id` — rename or change picture (needs `squad:edit`)
- [x] `POST /api/v1/squads/:id/members` — add member
- [x] `PATCH /api/v1/squads/:id/members/:userId` — make admin or demote
- [x] `DELETE /api/v1/squads/:id/members/:userId` — remove member, or leave with your own ID
- [x] `POST /api/v1/users/lookup` — which of these phone numbers are Roundup users
- [ ] `POST /api/v1/users/me/face-profile` — upload reference photo
- [ ] `GET /api/v1/users/me/friends` — get friend list

### Tests

- [x] Unit test `CreateOrGet` — duplicate calls return the same user
- [x] Test device token upsert — token moves to the new user (repository test)
- [x] Integration test repository against local Postgres
- [ ] E2E test `PATCH /api/v1/users/me` → updates name

---

## Event Service `[Go]`

### Project setup

- [ ] `go mod init github.com/SamPariat/roundup/services/event-service` — venue-service layout
- [ ] Hand-rolled state machine (no external lib)
- [ ] Write `Dockerfile`

### Database schema (goose migrations)

- [ ] `0001_create_events.sql` — id, squad_id, name, type (DRINKS/DINING/BOTH), status, created_by, deadline, created_at
- [ ] `0002_create_event_members.sql` — event_id, user_id, rsvp_status, is_guest, joined_at
- [ ] `0003_create_votes.sql` — event_id, user_id, venue_id, cast_at
- [ ] `0004_create_invite_links.sql` — token (uuid), event_id, created_by, expires_at, max_uses, use_count
- [ ] Add indexes on all foreign keys and status columns

### State machine

- [ ] `EventStatus` type: `DRAFT`, `VOTING`, `CONFIRMED`, `ACTIVE`, `CLOSED`
- [ ] Transition map in `domain/event_state.go`
- [ ] `Transition(current, next)` — returns `ErrInvalidTransition`
- [ ] Unit test every valid transition
- [ ] Unit test every invalid transition errors

### Domain & repositories

- [ ] `Event`, `EventMember`, `Vote`, `InviteLink` domain types
- [ ] `EventRepository` port + Postgres adapter (sqlc)

### Use cases

- [ ] `EventUseCase` — create, open voting, confirm venue, open tab, close, get, list squad events (paginated)
- [ ] `RSVPUseCase` — update RSVP (GOING / NOT_GOING / MAYBE), headcount; publish `rsvp.updated`
- [ ] `VotingUseCase` — cast vote (one per user per event), tally, resolve winner
- [ ] `InviteUseCase` — generate link, validate (not expired, under max uses), redeem as guest
- [ ] `GuestUseCase` — join as guest with phone + name

### Controllers (REST)

- [ ] `POST /api/v1/events` — create outing
- [ ] `GET /api/v1/events/:id` — get event detail
- [ ] `GET /api/v1/events` — get events for my squads
- [ ] `PATCH /api/v1/events/:id/status` — transition status
- [ ] `POST /api/v1/events/:id/rsvp` — update my RSVP
- [ ] `GET /api/v1/events/:id/rsvp` — get all RSVPs
- [ ] `POST /api/v1/events/:id/votes` — cast vote
- [ ] `GET /api/v1/events/:id/votes` — get tally
- [ ] `POST /api/v1/events/:id/invite` — generate invite link
- [ ] `POST /api/v1/invites/:token/redeem` — join via link
- [ ] `GET /api/v1/events/:id/members` — list members + RSVP status

### Kafka events (publish)

- [ ] `event.created` — on new event
- [ ] `rsvp.updated` — on any RSVP change
- [ ] `vote.cast` — on new vote
- [ ] `event.confirmed` — on venue confirmed
- [ ] `event.closed` — on event close

### Tests

- [ ] Unit test state machine — all transitions
- [ ] Unit test vote resolution — tie-breaking
- [ ] Unit test invite expiry validation
- [ ] E2E test full event lifecycle DRAFT → CLOSED

---

## Billing Service `[Go]`

### Project setup

- [ ] `go mod init github.com/SamPariat/roundup/services/billing-service` — venue-service layout
- [ ] Add `segmentio/kafka-go`
- [ ] Write `Dockerfile`

### Database schema (goose migrations)

- [ ] `0001_create_bills.sql` — id, event_id, status (OPEN/CLOSED), split_mode, total_paise, created_at
- [ ] `0002_create_bill_items.sql` — id, bill_id, name, amount_paise, added_by, created_at, deleted_at
- [ ] `0003_create_item_assignments.sql` — item_id, user_id, share_paise
- [ ] `0004_create_debts.sql` — id, bill_id, from_user, to_user, amount_paise, status (PENDING/SETTLED)
- [ ] `0005_create_audit_log.sql` — id, bill_id, action, payload (jsonb), performed_by, created_at
- [ ] `0006_create_split_configs.sql` — bill_id, user_id, percentage

### Command pattern

- [ ] `BillCommand` interface — `Execute(ctx)`, `AuditEntry()`
- [ ] `AddItemCommand` — adds item, writes audit log
- [ ] `RemoveItemCommand` — soft deletes item, writes audit log
- [ ] `AssignItemCommand` — assigns item to user
- [ ] `CloseBillCommand` — runs split, generates debts, closes bill
- [ ] `CommandInvoker` — executes command + persists audit entry in one transaction

### Strategy pattern

- [ ] `SplitStrategy` interface — `Calculate(bill, members) ([]Debt, error)`
- [ ] `EqualSplit` — divide total by member count
- [ ] `ByItemSplit` — sum assigned items per user
- [ ] `CustomSplit` — use explicit percentages from config
- [ ] `SplitStrategyFor(mode)` — map from `SplitMode`

### Debt simplification graph

- [ ] `DebtGraph` type
- [ ] `AddDebt(from, to, amount)` — build net balance map
- [ ] `Simplify()` — min-cash-flow algorithm, returns minimum number of transfers
- [ ] Unit test with 3-person circular debt (A owes B, B owes C, C owes A)
- [ ] Unit test with 5-person complex debt
- [ ] Unit test with all debts already settled

### Use cases

- [ ] `BillUseCase` — create bill, add/remove/assign item (via commands), close bill, get bill
- [ ] `DebtUseCase` — debts for user, mark as paid, total owed, total owing

### Controllers (REST)

- [ ] `POST /api/v1/bills` — create bill for event
- [ ] `GET /api/v1/bills/:id` — get bill with items
- [ ] `POST /api/v1/bills/:id/items` — add line item
- [ ] `DELETE /api/v1/bills/:id/items/:itemId` — remove item
- [ ] `POST /api/v1/bills/:id/items/:itemId/assign` — assign to users
- [ ] `POST /api/v1/bills/:id/close` — close and generate debts
- [ ] `GET /api/v1/bills/:id/debts` — get all debts for bill
- [ ] `GET /api/v1/debts/me` — my debts across all events
- [ ] `POST /api/v1/debts/:id/mark-paid` — mark debt as paid
- [ ] `GET /api/v1/bills/:id/audit` — audit log

### Kafka consumers

- [ ] Listen `event.closed` → auto-close open bill

### Kafka publishers

- [ ] Publish `tab.opened` on bill create
- [ ] Publish `tab.closed` on bill close
- [ ] Publish `debt.created` for each new debt
- [ ] Publish `debt.settled` when a debt is marked paid

### Tests

- [ ] Unit test `EqualSplit` — 3 people, ₹900 → ₹300 each
- [ ] Unit test `ByItemSplit` — uneven item assignments
- [ ] Unit test `DebtGraph.Simplify()` — multiple scenarios
- [ ] Unit test `CommandInvoker` — audit log entry created
- [ ] Integration test full bill lifecycle

---

## Venue Service `[Go]`

### Project setup

- [x] `go mod init github.com/SamPariat/roundup/services/venue-service`
- [x] Add `gofiber/fiber/v3`
- [x] Add `googlemaps.github.io/maps`
- [x] Add `redis/go-redis/v9`
- [x] Write `Dockerfile`

### Google Places integration

- [x] `internal/adapters/places/google_maps_adapter.go` — implements `domain.PlaceProvider`
- [x] Nearby search (lat, lng, radius, keyword, type) → `[]Venue`
- [x] Place details → `VenueDetail` (phone, website, hours, reviews, summary)
- [x] Photos streamed through `GET /api/v1/venues/photos/:ref`, so the API key never reaches clients (`Cache-Control: public, max-age=86400`)
- [x] Map Google responses to domain `Venue`, including `ratingCount` from `user_ratings_total`
- [x] Field mask on details to fetch only needed fields (cost control)
- [ ] Paging (`nextPageToken`): kept in the stash "venue-display full version (backup)"
- [ ] Structured opening hours / open-status labels / distance: kept in the same stash

### Redis cache layer

- [x] `internal/adapters/cache/redis_cache_adapter.go` — implements `domain.VenueCache`, best-effort (errors logged, never returned)
- [x] Search key `venue:search:{lat}:{lng}:{radius}:{query}:{type}` and detail key `venue:detail:{placeID}`, both 10 min TTL (`cache.DefaultTTL`)
- [x] On a cache miss, fetch from Google, then cache

### Database schema

- [x] `0001_create_saved_venues.sql` — id, squad_id, user_id, place_id, name, saved_at
- [x] `0002_create_venue_visits.sql` — id, squad_id, event_id, place_id, name, visited_at, avg_spend_in_paise
- [x] `0003_add_venue_snapshots.sql` — address, rating, rating_count, price_level, photo on both tables, so lists render without a Google call

### Use cases

- [x] `VenueUseCase.Search` and `GetDetail` — cache-first
- [x] `FavoriteUseCase.Add` — stores a venue snapshot from the cached details; the client sends only `placeID`
- [x] `FavoriteUseCase.List` — one entry per venue with `savedBy` names and `savedByMe`; Postgres groups and skips former members
- [x] `FavoriteUseCase.SavedIn` — which of my squads saved this venue (one user-service call, one query)
- [x] `HistoryUseCase.Record` and `List` — per venue, most recently visited first, with average known spend
- [x] Squad membership checked through user-service (`SquadDirectory.Members`) before every squad operation

### Controllers (REST)

- [x] `GET /api/v1/venues?lat=&lng=&radius=&query=&type=` — nearby search
- [x] `GET /api/v1/venues/:placeID` — place details plus `savedIn`
- [x] `GET /api/v1/venues/photos/:ref` — photo bytes (public at the gateway)
- [x] `GET` / `POST /api/v1/squads/:squadID/favorites`, `DELETE /api/v1/squads/:squadID/favorites/:placeID`
- [x] `GET` / `POST /api/v1/squads/:squadID/history`

### Recording visits automatically

- [ ] Record a visit when an event is confirmed (needs event-service; called directly, not through Kafka)

### Tests

- [x] Unit test cache hit — Places API not called
- [x] Unit test cache miss — Places API called, result cached
- [x] Repository tests against Postgres (snapshots, grouping, former members, average spend)
- [x] User-service client tests (membership, squads, timeouts, path escaping)
- [ ] Integration test search with the real Places API (dev key)

---

## Notification Service `[Go]`

### Project setup

- [ ] `go mod init github.com/SamPariat/roundup/services/notification-service` — venue-service layout
- [ ] Add Firebase Admin SDK for Go
- [ ] Add `segmentio/kafka-go`
- [ ] Write `Dockerfile`

### FCM integration

- [ ] Create `internal/fcm/client.go`
- [ ] Implement `Send(token, title, body, data)` → `error`
- [ ] Implement `SendMulticast(tokens, title, body, data)` → `BatchResponse`
- [ ] Handle `messaging/registration-token-not-registered` → signal User Service to remove token
- [ ] Log every notification send (success + failure)

### WhatsApp Business API

- [ ] Create `internal/whatsapp/client.go`
- [ ] Implement `SendTemplate(phone, templateName, params)` → `error`
- [ ] Template: `debt_reminder` — "Hey, you owe ₹X to Y from Friday drinks"
- [ ] Template: `outing_invite` — "Join us at Z this Saturday"
- [ ] Template: `debt_settled` — "Y marked ₹X as paid"

### Kafka consumers

- [ ] Listen `rsvp.updated` → notify organiser
- [ ] Listen `vote.cast` → notify all squad members (tally update)
- [ ] Listen `tab.closed` → notify all members (debts ready)
- [ ] Listen `debt.created` → notify debtor
- [ ] Listen `debt.settled` → notify creditor
- [ ] Listen `photo.tagged` → notify tagged user

### Notification routing

- [ ] Create `NotificationRouter` — decides channel (FCM vs WhatsApp vs both)
- [ ] Route by user preference (from User Service)
- [ ] Route by notification type (debt reminder → WhatsApp, photo tag → FCM)
- [ ] Implement quiet hours check before sending

### Notification log

- [ ] `0001_create_notification_log.sql` — id, user_id, channel, type, status, sent_at
- [ ] Write log entry on every send attempt
- [ ] Update status on delivery receipt (FCM success/failure)

### Tests

- [ ] Unit test `NotificationRouter` — correct channel selection
- [ ] Unit test quiet hours enforcement
- [ ] Integration test FCM send with Firebase test credentials

---

## Media Service `[Go]`

### Project setup

- [ ] `go mod init github.com/SamPariat/roundup/services/media-service` — venue-service layout
- [ ] Add `gofiber/fiber/v3`
- [ ] Add `cloudinary/cloudinary-go`
- [ ] Add `segmentio/kafka-go`
- [ ] Add `golang-jwt/jwt` for download tokens
- [ ] Add `mholt/archiver` for ZIP streaming
- [ ] Write `Dockerfile`

### Database schema

- [ ] `0001_create_photos.sql` — id, event_id, uploaded_by, cloudinary_id, width, height, tagging_status, created_at
- [ ] `0002_create_photo_faces.sql` — id, photo_id, user_id (nullable), confidence, bounding_box (jsonb)
- [ ] `0003_create_download_tokens.sql` — token, user_id, event_id, expires_at, used_at

### Cloudinary integration

- [ ] Create `internal/cloudinary/client.go`
- [ ] Implement `GenerateSignedUploadPreset(eventId, userId)` → `SignedParams`
- [ ] Implement `BuildUrl(cloudinaryId, transformations)` → `string`
- [ ] Implement `DeleteAsset(cloudinaryId)` → `error`
- [ ] Implement `CropFace(cloudinaryId, boundingBox)` → `string` (crop URL)

### Webhook handler

- [ ] `POST /api/v1/webhooks/cloudinary` — receive upload complete event
- [ ] Verify Cloudinary webhook signature
- [ ] Extract `public_id`, `width`, `height`, `event_id` from payload
- [ ] Insert photo record in DB
- [ ] Publish `photo.uploaded` Kafka event
- [ ] Return `200 OK` immediately

### Face tagging pipeline

- [ ] Implement `TaggingWorker` — subscribes to `photo.uploaded` Kafka topic
- [ ] For each photo: call Google Vision API → get face bounding boxes
- [ ] Fetch squad members for the event
- [ ] For each face: call AWS Rekognition `CompareFaces` against each member's reference photo
- [ ] Only create `photo_faces` row if confidence ≥ 0.85
- [ ] Update `tagging_status` → DONE or FAILED
- [ ] Publish `photo.tagged` Kafka event

### QR & download

- [ ] Create `QRService`
- [ ] Implement `GenerateToken(userId, eventId)` → signed JWT (7-day TTL)
- [ ] Store token in `download_tokens` table
- [ ] Implement `GenerateQR(token)` → PNG buffer (using `skip2/go-qrcode`)
- [ ] Create download endpoint `GET /api/v1/media/download/:token`
- [ ] Verify JWT signature
- [ ] Check not expired
- [ ] Fetch all photos tagged with `userId` for `eventId`
- [ ] Stream ZIP: pipe each Cloudinary URL into `archiver` → response
- [ ] Never buffer entire ZIP to disk
- [ ] Mark token `used_at` after download

### Controllers (REST)

- [ ] `GET /api/v1/media/upload-params?eventId=` — returns signed Cloudinary params
- [ ] `POST /api/v1/webhooks/cloudinary` — upload webhook
- [ ] `GET /api/v1/media/qr/:eventId` — generate and return QR PNG
- [ ] `GET /api/v1/media/download/:token` — stream ZIP (no auth guard)
- [ ] `GET /api/v1/events/:id/photos` — list photos for event
- [ ] `GET /api/v1/events/:id/photos/tagged` — my tagged photos in event
- [ ] `POST /api/v1/photos/:id/tag` — manual self-tag

### Kafka consumers / publishers

- [ ] Consume `photo.uploaded` → trigger tagging worker
- [ ] Publish `photo.tagged` → notify tagged users

### Tests

- [ ] Unit test ZIP streaming — no disk buffering
- [ ] Unit test JWT signing + verification
- [ ] Unit test confidence threshold — below 0.85 not saved
- [ ] Integration test Cloudinary webhook parsing

---

## Realtime Service `[Go]`

### Project setup

- [ ] `go mod init github.com/SamPariat/roundup/services/realtime-service` — venue-service layout
- [ ] Add `gorilla/websocket`
- [ ] Add `segmentio/kafka-go`
- [ ] Add `redis/go-redis`
- [ ] Write `Dockerfile`

### Connection management

- [ ] Create `Hub` — manages all active WebSocket connections
- [ ] Store connections in `map[userID]map[connID]*Conn`
- [ ] Handle concurrent access with `sync.RWMutex`
- [ ] Implement `Register(conn)` — add to hub
- [ ] Implement `Unregister(conn)` — remove from hub
- [ ] Implement `BroadcastToEvent(eventId, message)` — send to all users in event
- [ ] Implement `SendToUser(userId, message)` — send to specific user

### WebSocket server

- [ ] `GET /ws` — WebSocket upgrade endpoint
- [ ] Authenticate on connect: read `?token=` query param, validate via Auth Service
- [ ] Subscribe user to their events on connect
- [ ] Handle client ping/pong to detect dead connections
- [ ] Auto-remove dead connections from hub
- [ ] Implement graceful shutdown — drain connections

### Presence tracking

- [ ] Store active users per event in Redis SET `presence:{eventId}`
- [ ] Add user on connect (if they're viewing an event)
- [ ] Remove user on disconnect
- [ ] TTL on presence keys (5min) as fallback
- [ ] `GET /api/v1/events/:id/presence` — REST endpoint for current viewers

### Kafka consumers (fan out to WebSocket clients)

- [ ] Listen `rsvp.updated` → broadcast to event room
- [ ] Listen `vote.cast` → broadcast tally update to event room
- [ ] Listen `tab.opened` → broadcast to event members
- [ ] Listen `debt.created` → send to specific user
- [ ] Listen `debt.settled` → send to both parties
- [ ] Listen `photo.tagged` → send to tagged user
- [ ] Listen `event.confirmed` → broadcast venue to event room

### Message format

- [ ] Define `WsMessage` struct — `{ type, payload, eventId, timestamp }`
- [ ] Types: `RSVP_UPDATE`, `VOTE_UPDATE`, `TAB_ITEM_ADDED`, `DEBT_CREATED`, `DEBT_SETTLED`, `PHOTO_TAGGED`, `VENUE_CONFIRMED`, `USER_TYPING`, `USER_PRESENCE`
- [ ] Client sends `TYPING` message → server broadcasts to event members

### Tests

- [ ] Unit test Hub concurrent register/unregister
- [ ] Unit test broadcast to event — only connected users receive
- [ ] Integration test WebSocket connect + receive Kafka-triggered message

---

## Reservation Service `[Go]`

### Project setup

- [ ] `go mod init github.com/SamPariat/roundup/services/reservation-service` — venue-service layout
- [ ] Add `gofiber/fiber/v3`
- [ ] Add `segmentio/kafka-go`
- [ ] Write `Dockerfile`
- [ ] Research Dineout / EazyDiner partner API access — submit request

### Database schema

- [ ] `0001_create_reservations.sql` — id, event_id, venue_place_id, provider, reservation_ref, status, party_size, reservation_time, created_at
- [ ] `0002_create_reservation_events.sql` — id, reservation_id, event_type, payload, created_at

### Reservation state machine

- [ ] States: `PENDING`, `CONFIRMED`, `CANCELLED`, `NO_SHOW`
- [ ] Implement transitions with guard conditions
- [ ] Unit test all transitions

### Dineout/EazyDiner client

- [ ] Create `internal/dineout/client.go`
- [ ] Implement `SearchAvailability(placeId, date, partySize)` → `[]Slot`
- [ ] Implement `CreateReservation(placeId, slot, contactPhone, partySize)` → `ReservationRef`
- [ ] Implement `CancelReservation(ref)` → `error`
- [ ] Handle partner API auth (API key + HMAC signature)
- [ ] Map partner response to internal `Reservation` struct

### Service layer

- [ ] Create `ReservationService`
- [ ] Implement `CheckAvailability(eventId, venueId, date, partySize)`
- [ ] Implement `Book(eventId, venueId, slot, partySize)`
- [ ] Implement `Cancel(reservationId)`
- [ ] Implement `GetReservation(eventId)`

### Webhook handler

- [ ] `POST /api/v1/webhooks/dineout` — receive confirmation / cancellation
- [ ] Verify provider webhook signature
- [ ] Update reservation status
- [ ] Publish Kafka event on status change

### Controllers (REST)

- [ ] `GET /api/v1/reservations/availability?eventId=&venueId=&date=&partySize=`
- [ ] `POST /api/v1/reservations` — book a table
- [ ] `DELETE /api/v1/reservations/:id` — cancel
- [ ] `GET /api/v1/reservations/:eventId` — get event reservation

### Tests

- [ ] Unit test state machine transitions
- [ ] Unit test Dineout client response mapping
- [ ] Integration test booking flow with sandbox

---

## Analytics Service `[Go]`

### Project setup

- [ ] `go mod init github.com/SamPariat/roundup/services/analytics-service` — venue-service layout
- [ ] Add `segmentio/kafka-go`
- [ ] Write `Dockerfile`

### Database schema (goose migrations)

- [ ] `0001_create_analytics_events.sql` — id, squad_id, user_id, event_type, payload (jsonb), occurred_at
- [ ] `0002_squad_stats_view.sql` — materialized view: total outings, total spend, avg per outing per squad
- [ ] `0003_user_stats_view.sql` — materialized view: per-user stats across squads
- [ ] `0004_venue_stats_view.sql` — most visited venues per squad
- [ ] `0005_monthly_summary_view.sql` — monthly aggregates
- [ ] Add `pg_cron` job: refresh all views nightly at 2am

### Kafka consumers (event ingestion)

- [ ] Listen `event.created` → store analytics event
- [ ] Listen `event.closed` → store analytics event
- [ ] Listen `rsvp.updated` → store analytics event
- [ ] Listen `debt.created` → store analytics event
- [ ] Listen `debt.settled` → store analytics event
- [ ] Listen `photo.uploaded` → store analytics event

### Use cases

- [ ] `SquadStatsUseCase` — squad stats, leaderboard, venue history, spend by month
- [ ] `UserStatsUseCase` — total outings, total spent, squads
- [ ] `WrappedUseCase` — full year summary

### Controllers (REST)

- [ ] `GET /api/v1/analytics/squads/:id/stats` — squad summary stats
- [ ] `GET /api/v1/analytics/squads/:id/leaderboard` — rankings
- [ ] `GET /api/v1/analytics/squads/:id/venues` — venue history
- [ ] `GET /api/v1/analytics/squads/:id/spend?year=` — monthly spend
- [ ] `GET /api/v1/analytics/users/:id/stats` — user stats
- [ ] `GET /api/v1/analytics/squads/:id/wrapped?year=` — yearly wrapped

### Tests

- [ ] Unit test leaderboard calculation
- [ ] Unit test monthly spend aggregation
- [ ] Integration test materialized view refresh

---

## Feature Flag Service `[Go]`

### Project setup

- [ ] `go mod init github.com/SamPariat/roundup/services/feature-flag-service` — venue-service layout
- [ ] Add `gofiber/fiber/v3`
- [ ] Add `redis/go-redis`
- [ ] Add `google.golang.org/grpc` for gRPC server
- [ ] Write `Dockerfile`

### Database schema

- [ ] `0001_create_feature_flags.sql` — key, enabled, rollout (0-100), metadata (jsonb), updated_at
- [ ] `0002_create_user_overrides.sql` — flag_key, user_id, enabled, created_at

### Flag evaluation

- [ ] Create `FlagEvaluator`
- [ ] Implement `IsEnabled(key, userId)` → `bool`
- [ ] Check user override first
- [ ] Check global enabled
- [ ] Apply rollout percentage — deterministic hash of userId % 100
- [ ] Cache evaluation result in Redis (60s TTL)
- [ ] Invalidate cache on flag update

### gRPC server

- [ ] Implement `EvaluateFlag(key, userId)` RPC — for service-to-service
- [ ] Implement `EvaluateBulk(keys, userId)` RPC — evaluate multiple flags at once

### Controllers (REST — admin only)

- [ ] `GET /api/v1/flags` — list all flags
- [ ] `POST /api/v1/flags` — create flag
- [ ] `PATCH /api/v1/flags/:key` — update enabled, rollout
- [ ] `DELETE /api/v1/flags/:key` — delete flag
- [ ] `POST /api/v1/flags/:key/overrides` — set user-level override
- [ ] `DELETE /api/v1/flags/:key/overrides/:userId` — remove override

### Tests

- [ ] Unit test rollout — userId always maps to same bucket
- [ ] Unit test user override takes precedence over global
- [ ] Unit test cache invalidation on flag update

---

## Admin Service `[Go]`

### Project setup

- [ ] `go mod init github.com/SamPariat/roundup/services/admin-service` — venue-service layout
- [ ] Add `golang-jwt/jwt/v5`, `golang.org/x/crypto/bcrypt`
- [ ] Admin-only JWT separate from user JWT
- [ ] Write `Dockerfile`

### Auth

- [ ] Admin login — email + password (bcrypt hashed)
- [ ] Issue admin JWT (short TTL, 1hr)
- [ ] Admin JWT middleware on all routes
- [ ] Role-based: `SUPER_ADMIN`, `OPS`, `SUPPORT`

### User management

- [ ] `GET /api/v1/admin/users?q=` — search users by phone, email, name
- [ ] `GET /api/v1/admin/users/:id` — full user detail
- [ ] `POST /api/v1/admin/users/:id/ban` — soft ban (sets banned_at)
- [ ] `POST /api/v1/admin/users/:id/unban`
- [ ] `POST /api/v1/admin/users/:id/impersonate` — generate token for support
- [ ] `GET /api/v1/admin/users/:id/squads` — user's squads
- [ ] `GET /api/v1/admin/users/:id/debts` — open debts

### Webhook ops

- [ ] `POST /api/v1/admin/webhooks/:id/replay` — replay failed webhook

### Feature flag admin

- [ ] `GET /api/v1/admin/flags` — all flags with metadata
- [ ] `PATCH /api/v1/admin/flags/:key` — update flag (proxies to Flag Service)
- [ ] `POST /api/v1/admin/flags/:key/overrides` — per-user override

### Abuse detection

- [ ] `GET /api/v1/admin/abuse/suspicious-debts` — unusually large debts
- [ ] `GET /api/v1/admin/abuse/fake-invites` — invite links with abnormal redemption rate
- [ ] `POST /api/v1/admin/abuse/users/:id/flag` — flag user for review

### Venue partner ops

- [ ] `GET /api/v1/admin/venues/partners` — all partner venues
- [ ] `POST /api/v1/admin/venues/partners` — add partner
- [ ] `PATCH /api/v1/admin/venues/partners/:id` — update deal, rollout %
- [ ] `GET /api/v1/admin/venues/partners/:id/analytics` — visits, revenue

### Tests

- [ ] Unit test role-based access — OPS cannot access SUPER_ADMIN routes
- [ ] E2E test impersonation token — expires in 1hr
- [ ] E2E test webhook replay — re-publishes to Kafka

---

## App (React Native + Expo) `[App]`

One TypeScript codebase for web, Android and iOS. The web build ships first.

### Project setup

- [ ] Create the Expo app (TypeScript, Expo Router) in `app/`, with web, Android and iOS targets
- [ ] Expo Router for navigation; routes double as web URLs (`/squads/[id]/members`)
- [ ] TanStack Query for server state; React context for the session
- [ ] Generate a typed API client from `shared/api/openapi.yaml` (openapi-typescript + openapi-fetch)
- [ ] Firebase JS SDK — Firestore (live bill tab)
- [ ] `expo-image` for images, `react-native-svg` for icons and the logo
- [ ] `react-native-qrcode-svg` for QR display
- [ ] `expo-image-picker` for avatars and photos
- [ ] Configure environments: `APP_ENV=dev|prod` in `app.config.ts`, EAS Build profiles in `eas.json`

### Design system

- [ ] Generate `src/theme/tokens.ts` from the design system's `tokens.json` (Mocha + Latte, type, spacing, radius)
- [ ] Load Onest with `@expo-google-fonts/onest`
- [ ] Theme provider: follow the system (`useColorScheme`), with a System / Light / Dark override saved per device
- [ ] Build the design-system components in React Native with the same prop names and variant values
- [ ] Web: hover states via `Pressable`, SideNav from 1024px wide, BottomNav below

### Web

- [ ] Static export with `npx expo export --platform web`; host behind the gateway's domain
- [ ] Welcome page at `/` for signed-out visitors; return to the original link after sign-in
- [ ] Keep the refresh token in an httpOnly cookie on web (needs a gateway change); access token in memory

### Auth screens

- [ ] Phone number input screen
- [ ] OTP verification screen — calls Auth Service `/api/v1/auth/otp/*`
- [ ] Profile setup screen (name, avatar)
- [ ] Store tokens securely on native — `expo-secure-store` (Keychain / Keystore)
- [ ] Refresh access token automatically on 401
- [ ] Auto-login on app launch if refresh token valid

### Planning screens

- [ ] Squad list screen — all my squads
- [ ] Create squad screen
- [ ] Outing list screen — upcoming + past
- [ ] Create outing screen — name, type, deadline
- [ ] Venue search screen — map + list view
- [ ] Venue detail screen — photos, rating, price level
- [ ] Voting screen — see options, cast vote, live tally via WebSocket
- [ ] RSVP screen — going / not going / maybe, live headcount

### Bill tab screens

- [ ] Live bill tab screen — Firestore real-time
- [ ] Add item bottom sheet — name + amount
- [ ] Assign item to people sheet
- [ ] Split mode selector — equal, by item, custom
- [ ] Close tab confirmation screen

### Debt screens

- [ ] Debts overview screen — who owes me, what I owe
- [ ] Mark as paid action on a debt
- [ ] Running balance screen — across all squads

### Memory wall screens

- [ ] Event album screen — grid of photos
- [ ] Photo viewer screen — full screen, swipe
- [ ] Face tagging screen — tap untagged face → "that's me"
- [ ] QR code screen — show download QR

### Notification handling

- [ ] Request push permission on first launch (native; no web push in v1)
- [ ] Get the FCM / APNs device token with `expo-notifications` (`getDevicePushTokenAsync`)
- [ ] Register device token with User Service (`POST /api/v1/users/me/device-tokens`)
- [ ] Handle foreground notification — show in-app banner
- [ ] Handle background notification tap — deep link to relevant screen
- [ ] Notification preferences screen

### Realtime WebSocket

- [ ] Connect WebSocket on app launch (authenticated)
- [ ] Handle `RSVP_UPDATE` — update headcount on planning screen
- [ ] Handle `VOTE_UPDATE` — update tally on voting screen
- [ ] Handle `TAB_ITEM_ADDED` — append item on bill tab screen
- [ ] Handle `DEBT_SETTLED` — update debt status live
- [ ] Handle `PHOTO_TAGGED` — show notification in-app
- [ ] Reconnect on disconnect (exponential backoff)

### Settings

- [ ] Profile settings screen
- [ ] Theme setting — System / Light / Dark
- [ ] Notification preferences screen
- [ ] Linked squads screen
- [ ] Delete account flow (with confirmation)
- [ ] GDPR data export request button

---

## Cross-cutting

### API contracts

- [x] Document all REST endpoints in OpenAPI 3.0 YAML
- [x] Host Swagger UI for dev environment
- [ ] Write contract tests between gateway and each service (Pact)
- [x] Version all APIs (`/api/v1/`)

### Security

- [ ] Rotate all API keys every 90 days (automated)
- [ ] Enable HTTPS only on all Cloud Run services
- [ ] Add Content Security Policy headers
- [ ] Add CORS — whitelist only the web app origin and the admin domain (native apps send no Origin)
- [ ] Penetration test auth flow (JWT forgery, OTP brute force)
- [ ] Scan Docker images for vulnerabilities (Trivy in CI)

### Data & compliance

- [ ] Implement GDPR data export — all user data as JSON
- [ ] Implement account deletion — purge from all services
- [ ] Add data retention policy — delete inactive data after 2 years
- [ ] Encrypt PII at rest (Postgres column-level encryption for phone numbers)
- [ ] Add audit log for all admin actions

### Performance

- [ ] Load test API gateway — target 1000 req/s
- [ ] Load test Billing Service debt simplification — 100 concurrent bills
- [ ] Load test Realtime Service — 500 concurrent WebSocket connections
- [ ] DB query optimisation — EXPLAIN ANALYZE on all N+1 suspects
- [ ] Add connection pooling (PgBouncer) in front of all Postgres instances

### Documentation

- [ ] Write `README.md` for each service — setup, run, env vars
- [ ] Write `ARCHITECTURE.md` — system overview, service map
- [ ] Write `CONTRIBUTING.md` — branching, PR, code review process
- [ ] Write `RUNBOOK.md` — common ops procedures (restart service, replay webhook)
- [ ] Write `ADR/` folder — Architecture Decision Records for key choices

---

_Total services: 13 | Backend language: Go | Client: React Native + Expo (web, Android, iOS) | Estimated team: 4 people × 6 months_
