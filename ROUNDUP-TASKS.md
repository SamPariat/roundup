# Roundup — Backend Task Breakdown

> Check off tasks as you complete them. Each checkbox is one atomic unit of work.
> **Legend:** `[Go]` `[Infra]`
>
> **Language policy:** Backend is Go only. The client is one React Native + Expo app (TypeScript) for web, Android and iOS, built side by side in the separate `roundup-ui` repo. Its tasks live there.
>
> **Keep it simple:** services talk over REST (`/internal/v1/...` for service-to-service calls), each with its own Postgres database. No Kafka, gRPC or Pub/Sub. Scheduled work lives only in scheduler-service.
>
> **Later:** AI features are deferred, except receipt scanning for bills. Face recognition is still deferred.

---

## Table of contents

- [Infra & DevOps](#infra--devops)
- [API Gateway](#api-gateway--go)
- [Auth Service](#auth-service--go)
- [User Service](#user-service--go)
- [Event Service](#event-service--go)
- [Billing Service](#billing-service--go)
- [Scheduler Service](#scheduler-service--go)
- [Venue Service](#venue-service--go)
- [Notification Service](#notification-service--go)
- [Media Service](#media-service--go)
- [Realtime Service](#realtime-service--go)
- [Reservation Service](#reservation-service--go)
- [Analytics Service](#analytics-service--go)
- [Feature Flag Service](#feature-flag-service--go)
- [Admin Service](#admin-service--go)
- [Client contract](#client-contract)
- [Cross-cutting](#cross-cutting)

---

## Infra & DevOps `[Infra]`

### Local environment

- [x] Install Docker Desktop
- [x] Install Go 1.27+
- [x] Install goose (`make install`)
- [x] Install sqlc (`make install`)
- [x] Install golangci-lint

### Monorepo setup

- [x] Initialise git repository
- [x] Create root `docker-compose.yml`
- [x] Add `.gitignore`
- [x] Add `.editorconfig` for consistent formatting
- [x] Create `services/` directory for all microservices
- [x] Create `scripts/` for dev helper scripts
- [x] `make services-up` boots every service in Docker (compose `services` profile)

### Docker Compose — local services

- [x] Add PostgreSQL container (one per service, separate ports)
- [x] Add Redis container (local venue cache only)
- [x] Set up Docker networking so all services can communicate
- [ ] Add health checks to all containers

### CI/CD

- [x] Create GitHub Actions workflow directory `.github/workflows/`
- [x] Write `ci.yml` — runs on every PR
- [x] Add lint step (golangci-lint)
- [x] Add test step for each service
- [x] Add build step for each service (Docker image per service; `ci-ok` is the required check)
- [ ] Write `deploy.yml` — runs on merge to main
- [ ] Push images to Google Artifact Registry
- [ ] Add Cloud Run deploy step per service, plus goose migrations before rollout
- [ ] Set up environment secrets in GitHub (DB URLs, API keys)
- [ ] Write `CODEOWNERS` file

### GCP production setup

Cloud Run in asia-south1, every service scales to zero (no min-instances; cold starts are fine). One Cloud SQL instance with one database per service. No Memorystore at launch: the venue cache is best-effort, so venue-service runs without Redis.

- [ ] Create GCP project
- [ ] Enable Cloud Run, Cloud SQL, Artifact Registry and Secret Manager APIs
- [ ] Create one Cloud SQL Postgres instance (asia-south1, db-g1-small) with a database per service
- [ ] Create Artifact Registry repository
- [ ] Set up service accounts per microservice
- [ ] Configure IAM roles (least privilege), including who may call each service's `/internal/v1` routes
- [ ] Set up Secret Manager for all env vars
- [ ] Direct VPC egress from Cloud Run to Cloud SQL

### Observability

- [x] Graceful shutdown on SIGTERM and panic recovery on every service (`shared/server.Run`, `shared/middleware.Recover`)
- [x] Structured request/response logs with the gateway's request ID on every service

- [ ] Set up Grafana Cloud account (free tier)
- [ ] Ship Cloud Run logs to Grafana Loki
- [ ] Set up Prometheus metrics endpoint per service
- [ ] Create Grafana dashboards — request rate, error rate and p95 latency per service
- [ ] Configure alert — error rate > 1% for 5min → Discord
- [ ] Configure alert — service down → Discord
- [ ] Add OpenTelemetry tracing to each service, keyed on the request ID

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

### Routing

- [x] Proxy `/api/v1/auth`, `/users`, `/venues`, `/squads`, `/events` to their upstream service URLs (config per service)
- [x] Answer 501 for route groups whose service isn't built yet
- [ ] Proxy the billing routes once billing-service exists
- [x] Block external `POST /api/v1/users` (only auth-service creates users, directly)
- [x] `/internal/v1/*` is unreachable through the gateway (it only mounts `/api/v1`)
- [x] Upstream down or slow → 502
- [x] Request ID generation (UUID v4) and forwarding as `X-Request-Id`
- [x] Forward `X-User-ID` after auth validation
- [x] Refresh token in an httpOnly `roundup_refresh` cookie for web clients, CORS from `CORS_ALLOWED_ORIGINS`

### Auth middleware

- [x] Strip any client-sent `X-User-ID` on every request
- [x] Read `Authorization: Bearer <token>`, verify with `shared/authtoken` (EdDSA only, iss/aud/exp)
- [x] Return 401 if the token is missing or invalid
- [x] Public routes: `/api/v1/auth/*`, `/health`

### Rate limiting

- [x] Rate limit `/api/v1/auth/*` by IP: 20 requests per 15 minutes, in memory per gateway instance (protects OTP sending)
- [ ] Rate limit by user ID for authenticated routes
- [x] Return `429 Too Many Requests` with `Retry-After` header

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
- [ ] SMS provider adapter for production (e.g. MSG91) — auth refuses to start in prod until it exists (**blocks launch**)
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
- [ ] `0005_create_friend_connections.sql` — user_id_a, user_id_b, created_at (unique pair)
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
- [ ] `POST /internal/v1/device-tokens/batch` — device tokens for a list of users (notification service)
- [ ] `GET /api/v1/users/me/friends` — get friend list

### Tests

- [x] Unit test `CreateOrGet` — duplicate calls return the same user
- [x] Test device token upsert — token moves to the new user (repository test)
- [x] Integration test repository against local Postgres
- [ ] E2E test `PATCH /api/v1/users/me` → updates name

---

## Event Service `[Go]`

See `services/event-service/ARCHITECTURE.md`. Other services are called directly; invites and guests are deferred.

### Project setup

- [x] `go mod init github.com/SamPariat/roundup/services/event-service` — venue-service layout
- [x] Hand-rolled state machine (no external lib)
- [x] Write `Dockerfile`

### Database schema (goose migrations)

- [x] `0001_create_events.sql` — status, times, winner place
- [x] `0002_create_event_rsvps.sql` — one answer per member; no row = pending
- [x] `0003_create_nominations.sql` — suggested places with a venue snapshot
- [x] `0004_create_votes.sql` — one vote per member, kept after close
- [x] Indexes for the squad list and the due sweeps

### State machine

- [x] `Status`: `DRAFT`, `VOTING`, `CONFIRMED`, `ACTIVE`, `CLOSED`; `CanMoveTo` allows only the next step
- [x] Unit test every status pair
- [x] Status checks repeated inside each guarded SQL write

### Use cases

- [x] Create, list (paged, open first), get, edit
- [x] RSVP (GOING / MAYBE / NOT_GOING)
- [x] Suggest and withdraw places (snapshot from venue-service)
- [x] Open votes; winner = most votes, tie → earliest suggestion
- [x] Start/close voting, start the night, close the night (records the visit)
- [x] Internal `confirm-due` and `close-due` sweeps for scheduler-service
- [x] Permissions from `policy.csv` (MEMBER, ADMIN, CREATOR)
- [ ] `ACTIVE` = bill open: closing the night closes the bills in billing-service (direct call)
- [ ] Invite links and guests — deferred

### Controllers (REST)

- [x] `GET/POST /api/v1/squads/:squadID/events`
- [x] `GET/PATCH /api/v1/events/:eventID`
- [x] `PUT /api/v1/events/:eventID/rsvp`
- [x] `POST /api/v1/events/:eventID/nominations`, `DELETE …/nominations/:placeID`
- [x] `PUT /api/v1/events/:eventID/vote`
- [x] `POST /api/v1/events/:eventID/voting/start`, `…/voting/close`, `…/start`, `…/close`
- [x] `POST /internal/v1/events/confirm-due`, `…/close-due`

### Tests

- [x] Domain, policy, use cases (fakes), Postgres repository, venue client, routes vs spec
- [x] Concurrent sweeps move each event exactly once

---

## Billing Service `[Go]`

The bill is why you open an event. Scan the receipt, one payer per bill, several bills per night, equal split, and the bills are netted into one debt per pair. People settle outside the app and mark the debt paid. No payments.

### Project setup

- [ ] `go mod init github.com/SamPariat/roundup/services/billing-service` — venue-service layout
- [ ] Write `Dockerfile`, `.env.*.example`, compose Postgres, `billing-*` Makefile targets

### Database schema (goose migrations)

- [ ] `0001_create_bills.sql` — id, event_id, paid_by, total_paise, receipt_ref, created_at, deleted_at
- [ ] `0002_create_bill_items.sql` — id, bill_id, name, amount_paise (from the scanned receipt, editable)
- [ ] `0003_create_bill_shares.sql` — bill_id, user_id (who splits this bill)
- [ ] `0004_create_debts.sql` — id, event_id, from_user, to_user, amount_paise, status (PENDING/SETTLED), settled_at

### Domain

- [ ] Equal split in paise; the leftover paise go to a fixed order of members so the totals always match
- [ ] Net all of a night's bills into one debt per pair of people
- [ ] Unit tests: ₹900 across 3 → ₹300 each; ₹1000 across 3 (leftover paise); two payers in one night net out

### Receipt scanning

- [ ] `ReceiptScanner` port — photo → line items + total
- [ ] OCR adapter (provider still to be chosen)
- [ ] User confirms or edits items before the bill is saved

### Use cases

- [ ] `BillUseCase` — add, edit and remove a bill while the event is `ACTIVE`, choose who shares it
- [ ] Close the event's bills → generate the netted debts (idempotent)
- [ ] `DebtUseCase` — my debts across events, mark as paid (only the person owed can confirm), totals owed / owing

### Controllers (REST)

- [ ] `GET/POST /api/v1/events/:eventID/bills`
- [ ] `GET/PATCH/DELETE /api/v1/bills/:billID`
- [ ] `POST /api/v1/events/:eventID/bills/scan` — receipt photo → proposed items
- [ ] `GET /api/v1/events/:eventID/debts`
- [ ] `GET /api/v1/debts/me`
- [ ] `POST /api/v1/debts/:debtID/mark-paid`
- [ ] `POST /internal/v1/events/:eventID/bills/close` — called by event-service when the night closes
- [ ] Add every route to `shared/api/openapi.yaml`

### Tests

- [ ] Use cases with fakes, Postgres repository, routes vs spec
- [ ] Integration test full bill lifecycle

---

## Scheduler Service `[Go]`

No database and no domain logic. It calls each service's idempotent `/internal/v1` endpoint on a schedule.

- [ ] `go mod init github.com/SamPariat/roundup/services/scheduler-service`
- [ ] Job: event `confirm-due` (voting deadline passed)
- [ ] Job: event `close-due` (night over)
- [ ] Config: target URLs and interval per job via `shared/env`
- [ ] Log each run with its outcome; one failed job never stops the others
- [ ] Write `Dockerfile`
- [ ] Prod: Cloud Scheduler hits the jobs (scale to zero) instead of an always-on ticker

---

## Venue Service `[Go]`

### Project setup

- [x] `go mod init github.com/SamPariat/roundup/services/venue-service`
- [x] Add `gofiber/fiber/v3`
- [x] Places API (New) over plain HTTP (no Google client library)
- [x] Add `redis/go-redis/v9`
- [x] Write `Dockerfile`

### Google Places integration

- [x] `internal/adapters/places/google_places_adapter.go` — implements `domain.PlaceProvider` with Places API (New)
- [x] Search (lat, lng, radius, query, type) → `[]Venue`, as Text Search (New) inside the square around the radius
- [x] Place details → `VenueDetail` (phone, website, hours, reviews, summary)
- [x] Photos as direct, keyless Google image URLs in each response (`skipHttpRedirect`), looked up per response since they expire, with author credits
- [x] Map Google responses to domain `Venue`, including `ratingCount` from `userRatingCount`
- [x] Field masks on search and details to fetch only needed fields (cost control)
- [x] Paging with Google's `nextPageToken` as the cursor (20 per page, 60 per search)
- [ ] Structured opening hours / open-status labels / distance: kept in the same stash

### Redis cache layer

- [x] `internal/adapters/cache/redis_cache_adapter.go` — implements `domain.VenueCache`, best-effort (errors logged, never returned)
- [x] Search key `venue:search:v1:{lat}:{lng}:{radius}:{query}:{type}:{pageToken}` and detail key `venue:detail:v1:{placeID}`, both 10 min TTL (`cache.DefaultTTL`)
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
- [ ] Fill a visit's spend from the night's bills once billing-service exists

### Controllers (REST)

- [x] `GET /api/v1/venues?lat=&lng=&radius=&query=&type=` — nearby search
- [x] `GET /api/v1/venues/:placeID` — place details plus `savedIn`
- [x] `POST /internal/v1/venues/photo-urls` — fresh URLs for stored photo refs (event-service)
- [x] `GET` / `POST /api/v1/squads/:squadID/favorites`, `DELETE /api/v1/squads/:squadID/favorites/:placeID`
- [x] `GET` / `POST /api/v1/squads/:squadID/history`

### Recording visits automatically

- [x] Record a visit when an event closes (event-service calls it directly)

### Tests

- [x] Unit test cache hit — Places API not called
- [x] Unit test cache miss — Places API called, result cached
- [x] Repository tests against Postgres (snapshots, grouping, former members, average spend)
- [x] User-service client tests (membership, squads, timeouts, path escaping)
- [ ] Integration test search with the real Places API (dev key)

---

## Notification Service `[Go]`

Other services call `POST /internal/v1/notifications` directly when something happens.

### Project setup

- [ ] `go mod init github.com/SamPariat/roundup/services/notification-service` — venue-service layout
- [ ] Add Firebase Admin SDK for Go (FCM only)
- [ ] Write `Dockerfile`

### FCM integration

- [ ] `Send(token, title, body, data)` and `SendMulticast(tokens, …)`
- [ ] Unregistered token → ask user-service to remove it
- [ ] Log every send (success + failure)

### WhatsApp Business API

- [ ] `SendTemplate(phone, templateName, params)`
- [ ] Templates: `debt_reminder`, `outing_invite`, `debt_settled`

### Triggers (direct calls)

- [ ] Event-service: RSVP changed → organiser; voting opened / venue confirmed → squad
- [ ] Billing-service: debts ready → everyone in the night; debt created → debtor; debt marked paid → the other person
- [ ] Scheduler-service: debt reminders

### Routing & log

- [ ] `NotificationRouter` — FCM vs WhatsApp by notification type and user preference
- [ ] Quiet hours check before sending
- [ ] `0001_create_notification_log.sql` — id, user_id, channel, type, status, sent_at

### Tests

- [ ] Unit test `NotificationRouter` — correct channel selection
- [ ] Unit test quiet hours enforcement

---

## Media Service `[Go]`

Face tagging is deferred with the other AI features.

### Project setup

- [ ] `go mod init github.com/SamPariat/roundup/services/media-service` — venue-service layout
- [ ] Add `cloudinary/cloudinary-go`
- [ ] Write `Dockerfile`

### Database schema

- [ ] `0001_create_photos.sql` — id, event_id, uploaded_by, cloudinary_id, width, height, created_at
- [ ] `0002_create_download_tokens.sql` — token, user_id, event_id, expires_at, used_at

### Cloudinary & upload

- [ ] Signed upload params per event and user
- [ ] `POST /api/v1/webhooks/cloudinary` — verify signature, insert photo row, return 200
- [ ] Delete asset when a photo is removed

### QR & download

- [ ] Signed download token (7-day TTL), stored in `download_tokens`
- [ ] QR PNG for the token (`skip2/go-qrcode`)
- [ ] `GET /api/v1/media/download/:token` — stream a ZIP of the event's photos, never buffered to disk

### Controllers (REST)

- [ ] `GET /api/v1/media/upload-params?eventId=`
- [ ] `GET /api/v1/media/qr/:eventId`
- [ ] `GET /api/v1/events/:id/photos`

### Tests

- [ ] Unit test ZIP streaming — no disk buffering
- [ ] Unit test download token signing + expiry
- [ ] Integration test Cloudinary webhook parsing

---

## Realtime Service `[Go]`

Other services POST to `/internal/v1/realtime/publish` and realtime fans the update out to connected clients.

### Project setup

- [ ] `go mod init github.com/SamPariat/roundup/services/realtime-service` — venue-service layout
- [ ] Add a WebSocket library
- [ ] Write `Dockerfile`

### Connection management

- [ ] `Hub` — connections per user, guarded by `sync.RWMutex`
- [ ] `Register`, `Unregister`, `BroadcastToEvent`, `SendToUser`
- [ ] `GET /ws` — upgrade, authenticate with the access token via `shared/authtoken`
- [ ] Ping/pong; drop dead connections; drain on shutdown

### Messages

- [ ] `WsMessage` — `{ type, payload, eventId, timestamp }`
- [ ] Types: `RSVP_UPDATE`, `VOTE_UPDATE`, `VENUE_CONFIRMED`, `BILL_UPDATED`, `DEBT_SETTLED`
- [ ] Decide how fan-out works across several Cloud Run instances before launch

### Tests

- [ ] Unit test Hub concurrent register/unregister
- [ ] Unit test broadcast to event — only connected users receive

---

## Reservation Service `[Go]`

- [ ] Research Dineout / EazyDiner partner API access — submit request
- [ ] `go mod init github.com/SamPariat/roundup/services/reservation-service` — venue-service layout, `Dockerfile`
- [ ] `0001_create_reservations.sql` — id, event_id, venue_place_id, provider, reservation_ref, status, party_size, reservation_time, created_at
- [ ] State machine: `PENDING`, `CONFIRMED`, `CANCELLED`, `NO_SHOW` (unit test every transition)
- [ ] Partner client — availability, book, cancel (API key + HMAC)
- [ ] `POST /api/v1/webhooks/dineout` — verify signature, update status
- [ ] REST: availability, book, cancel, get for event

---

## Analytics Service `[Go]`

Reads from the other services' internal endpoints. No event stream.

- [ ] `go mod init github.com/SamPariat/roundup/services/analytics-service` — venue-service layout, `Dockerfile`
- [ ] Squad stats — outings, total spend, average per outing, most visited venues
- [ ] User stats across squads
- [ ] Yearly wrapped
- [ ] Nightly refresh via a scheduler-service job (not `pg_cron`)
- [ ] REST: `GET /api/v1/analytics/squads/:id/stats`, `…/wrapped?year=`, `GET /api/v1/analytics/users/me/stats`

---

## Feature Flag Service `[Go]`

- [ ] `go mod init github.com/SamPariat/roundup/services/feature-flag-service` — venue-service layout, `Dockerfile`
- [ ] `0001_create_feature_flags.sql` — key, enabled, rollout (0-100), updated_at
- [ ] `0002_create_user_overrides.sql` — flag_key, user_id, enabled
- [ ] `IsEnabled(key, userId)` — override, then global, then deterministic rollout bucket
- [ ] `POST /internal/v1/flags/evaluate` — several keys for one user
- [ ] Admin REST: list, create, update, delete flags; set/remove overrides
- [ ] Unit test: same user always lands in the same bucket; override beats global

---

## Admin Service `[Go]`

- [ ] `go mod init github.com/SamPariat/roundup/services/admin-service` — venue-service layout, `Dockerfile`
- [ ] Admin login — email + password (bcrypt), separate short-lived admin JWT
- [ ] Roles: `SUPER_ADMIN`, `OPS`, `SUPPORT`
- [ ] Users: search, detail, ban/unban, squads, open debts
- [ ] Flags: manage through feature-flag-service
- [ ] Abuse: unusually large debts, flag user for review
- [ ] Audit log for every admin action
- [ ] Unit test role-based access — OPS cannot reach SUPER_ADMIN routes

---

## Client contract

The client (`roundup-ui`) generates its typed API client from `shared/api/openapi.yaml`. Web and native are built side by side there, so the spec must ship with each backend change.

- [x] Every public route in `shared/api/openapi.yaml`, served by Swagger UI at the gateway
- [x] `TestRoutesAreDocumented` in auth, user, event and venue services (CI fails until the spec matches)
- [ ] Same test in every new service
- [ ] Land the billing spec before the handlers so the bill screens can start early

---

## Cross-cutting

### Security

- [ ] HTTPS only on all Cloud Run services; internal services not publicly invokable
- [ ] Add Content Security Policy headers
- [x] CORS — gateway `CORS_ALLOWED_ORIGINS`; set the real web origin in prod, and the admin domain once it exists
- [ ] Penetration test auth flow (JWT forgery, OTP brute force)
- [ ] Scan Docker images for vulnerabilities (Trivy in CI)

### Data & compliance

- [ ] GDPR data export — all user data as JSON
- [ ] Account deletion — purge from every service
- [ ] Data retention policy — delete inactive data after 2 years
- [ ] Encrypt PII at rest (phone numbers)

### Performance

- [ ] Load test the gateway
- [ ] `EXPLAIN ANALYZE` on list queries and N+1 suspects
- [ ] Check Cloud SQL connection limits with every service scaled out

### Documentation

- [x] `ARCHITECTURE.md` for auth, user, venue and event services
- [ ] `ARCHITECTURE.md` for each new service
- [ ] `RUNBOOK.md` — deploy, roll back, run a migration, rerun a scheduler job
- [ ] `CONTRIBUTING.md` — branching, PR, code review process

---

_Backend: 14 Go services (gateway, auth, user, event, billing, scheduler, venue, notification, media, realtime, reservation, analytics, flags, admin) · Client: React Native + Expo in `roundup-ui` (web, Android, iOS side by side)_
