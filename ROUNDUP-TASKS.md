# Roundup — Full Task Breakdown

> Check off tasks as you complete them. Each checkbox is one atomic unit of work.
> **Legend:** `[Go]` `[NestJS]` `[Java]` `[Python]` `[Infra]` `[Flutter]`

---

## Table of contents

- [Infra & DevOps](#infra--devops)
- [API Gateway](#api-gateway--go)
- [Auth Service](#auth-service--nestjs)
- [User Service](#user-service--java)
- [Event Service](#event-service--java)
- [Billing Service](#billing-service--java)
- [Payment Service](#payment-service--go)
- [Venue Service](#venue-service--go)
- [Notification Service](#notification-service--go)
- [Media Service](#media-service--go)
- [Realtime Service](#realtime-service--go)
- [Reservation Service](#reservation-service--go)
- [AI Service](#ai-service--python)
- [Analytics Service](#analytics-service--java)
- [Feature Flag Service](#feature-flag-service--go)
- [Admin Service](#admin-service--nestjs)
- [Flutter App](#flutter-app)
- [Cross-cutting](#cross-cutting)

---

## Infra & DevOps `[Infra]`

### Local environment

- [x] Install Docker Desktop
- [x] Install pnpm globally
- [x] Install Java 21 (Temurin / Adoptium)
- [x] Install Go 1.22+
- [x] Install Python 3.11+
- [x] Install Node.js 20 LTS
- [x] Install Flutter SDK
- [x] Install nest-cli globally (`npm i -g @nestjs/cli`)
- [ ] Install drizzle-kit globally
- [ ] Install sqlc (Go SQL codegen)
- [ ] Install protoc (gRPC compiler)
- [ ] Install buf CLI (protobuf linting)

### Monorepo setup

- [ ] Initialise git repository
- [ ] Create root `docker-compose.yml`
- [x] Add `.gitignore` for all languages
- [x] Add `.editorconfig` for consistent formatting
- [x] Create `services/` directory for all microservices
- [x] Create `proto/` directory for shared gRPC definitions
- [x] Create `scripts/` for dev helper scripts
- [x] Write `scripts/start-all.sh` to boot all services locally

### Docker Compose — local services

- [ ] Add PostgreSQL container (one per service, separate ports)
- [ ] Add Redis container (shared, single instance)
- [ ] Add Kafka container (Bitnami image)
- [ ] Add Zookeeper container (Kafka dependency)
- [ ] Add Kafka UI container (Provectus) for local debugging
- [ ] Add Redis Commander container for local Redis browsing
- [ ] Add pgAdmin container for DB browsing
- [ ] Set up Docker networking so all services can communicate
- [ ] Write `docker-compose.override.yml` for local-only overrides
- [ ] Add health checks to all containers

### CI/CD

- [ ] Create GitHub Actions workflow directory `.github/workflows/`
- [ ] Write `ci.yml` — runs on every PR
- [ ] Add lint step for each language (golangci-lint, eslint, checkstyle, ruff)
- [ ] Add test step for each service
- [ ] Add build step for each service
- [ ] Write `deploy.yml` — runs on merge to main
- [ ] Add Docker image build per service
- [ ] Push images to Google Artifact Registry
- [ ] Add Cloud Run deploy step per service
- [ ] Set up environment secrets in GitHub (DB URLs, API keys)
- [ ] Write `CODEOWNERS` file

### Kafka setup

- [ ] Define all Kafka topic names as constants in `proto/topics.ts` / shared constants
- [ ] Topic: `rsvp.updated`
- [ ] Topic: `event.created`
- [ ] Topic: `event.closed`
- [ ] Topic: `tab.opened`
- [ ] Topic: `tab.closed`
- [ ] Topic: `debt.created`
- [ ] Topic: `debt.settled`
- [ ] Topic: `photo.uploaded`
- [ ] Topic: `photo.tagged`
- [ ] Topic: `payment.initiated`
- [ ] Topic: `payment.confirmed`
- [ ] Topic: `payment.failed`
- [ ] Topic: `outing.recap.requested`
- [ ] Topic: `notification.send`
- [ ] Set retention policy (7 days) for all topics
- [ ] Set replication factor (3) for production topics

### gRPC proto definitions

- [ ] Create `proto/auth.proto` — token validation RPC
- [ ] Create `proto/user.proto` — get user by ID RPC
- [ ] Create `proto/billing.proto` — get bill, get debts RPCs
- [ ] Create `proto/payment.proto` — initiate payment RPC
- [ ] Create `proto/flags.proto` — evaluate flag RPC
- [ ] Generate Go stubs from all protos
- [ ] Generate Java stubs from all protos
- [ ] Generate Python stubs from all protos
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

### Project setup

- [ ] `go mod init roundup/gateway`
- [ ] Add Fiber v2 dependency (`gofiber/fiber`)
- [ ] Add `golang-jwt/jwt` dependency
- [ ] Add `redis/go-redis` dependency
- [ ] Add `prometheus/client_golang` dependency
- [ ] Add OpenTelemetry Go SDK
- [ ] Create `cmd/main.go` entry point
- [ ] Create `internal/config/` — env var loading with validation
- [ ] Create `.env.example` with all required vars
- [ ] Write `Dockerfile` (multi-stage, scratch final image)

### Routing

- [ ] Define route groups: `/api/v1/auth`, `/api/v1/users`, `/api/v1/events`, `/api/v1/billing`, `/api/v1/venues`, `/api/v1/media`, `/api/v1/realtime`
- [ ] Implement dynamic proxy to upstream service URLs
- [ ] Add service URL config per route group
- [ ] Handle routing errors (upstream down → 503)
- [ ] Add request ID generation middleware (UUID v4)
- [ ] Forward `X-Request-ID` header to all upstream services
- [ ] Forward `X-User-ID` header after auth validation

### Auth middleware

- [ ] Create auth middleware — reads `Authorization: Bearer <token>`
- [ ] Call Auth Service gRPC `ValidateToken` RPC
- [ ] Cache valid tokens in Redis (TTL = token expiry)
- [ ] Return 401 if token invalid
- [ ] Attach user ID to request context
- [ ] Whitelist public routes (skip auth): `POST /auth/login`, `GET /venues/search`

### Rate limiting

- [ ] Implement Redis-backed rate limiter
- [ ] Rate limit by IP: 100 req/min for unauthenticated
- [ ] Rate limit by user ID: 500 req/min for authenticated
- [ ] Return `429 Too Many Requests` with `Retry-After` header
- [ ] Exempt internal service calls from rate limiting

### Load balancing

- [ ] Support multiple upstream instances per service
- [ ] Implement round-robin selection
- [ ] Add health check poller per upstream (every 10s)
- [ ] Remove unhealthy upstreams from rotation
- [ ] Re-add upstreams when health check recovers

### Observability

- [ ] Log every request: method, path, status, duration, requestId
- [ ] Expose `/metrics` endpoint for Prometheus
- [ ] Add request count metric per route
- [ ] Add latency histogram per route
- [ ] Add upstream error count metric
- [ ] Expose `/health` endpoint — returns 200 if healthy

---

## Auth Service `[NestJS]`

### Project setup

- [ ] `nest new auth-service` with pnpm
- [ ] Add `@nestjs/config` dependency
- [ ] Add `firebase-admin` dependency
- [ ] Add `@nestjs/jwt` dependency
- [ ] Add `zod` dependency
- [ ] Add `ioredis` dependency
- [ ] Add `@grpc/grpc-js` and `@grpc/proto-loader` for gRPC server
- [ ] Configure `tsconfig.json` (strict mode, path aliases)
- [ ] Create `src/config/` — typed env config with Zod validation
- [ ] Write `.env.example`
- [ ] Write `Dockerfile`

### Firebase module

- [ ] Create `src/infrastructure/firebase/firebase.provider.ts`
- [ ] Implement `admin.initializeApp()` with service account from env
- [ ] Handle already-initialised guard (`admin.apps.length`)
- [ ] Create `src/infrastructure/firebase/firebase.module.ts` (`@Global()`)
- [ ] Create `src/infrastructure/firebase/firebase-auth.adapter.ts`
- [ ] Implement `verifyToken(token)` → `VerifiedToken | null`
- [ ] Implement `revokeTokens(uid)` → `void`
- [ ] Implement `getUser(uid)` → `ProviderUser | null`
- [ ] Handle `auth/id-token-expired` gracefully (return null)
- [ ] Handle `auth/argument-error` gracefully (return null)
- [ ] Add test phone numbers in dev environment

### Port definition

- [ ] Create `src/modules/auth/ports/auth-provider.port.ts`
- [ ] Define `VerifiedToken` interface
- [ ] Define `ProviderUser` interface
- [ ] Define `AuthProviderPort` interface

### Strategy pattern

- [ ] Create `src/modules/auth/strategies/auth-strategy.interface.ts`
- [ ] Define `AuthUser` interface
- [ ] Define `AuthProvider` enum (`FIREBASE`, `JWT`, `API_KEY`)
- [ ] Define `IAuthStrategy` interface
- [ ] Create `firebase-auth.strategy.ts` — implements `IAuthStrategy`
- [ ] Implement `validate(token)` — calls port, maps to `AuthUser`
- [ ] Create `jwt.strategy.ts` — for internal service tokens
- [ ] Implement JWT `validate(token)` using `JwtService`
- [ ] Create `api-key.strategy.ts` — for webhook endpoints
- [ ] Implement API key lookup with Redis cache (5min TTL)

### Factory

- [ ] Create `src/modules/auth/factories/auth-strategy.factory.ts`
- [ ] Implement `static create(...)` — returns configured factory
- [ ] Build strategy map `Map<AuthProvider, IAuthStrategy>`
- [ ] Implement `resolve(provider)` — throws if unknown

### Guard & decorators

- [ ] Create `src/modules/auth/auth.guard.ts`
- [ ] Implement `canActivate()` — reads reflector metadata
- [ ] Handle `@Public()` decorator — skip guard
- [ ] Handle `@UseAuthProvider()` decorator — pin strategy
- [ ] Extract Bearer token from `Authorization` header
- [ ] Extract API key from `X-API-Key` header
- [ ] Attach `AuthUser` to `request.user`
- [ ] Create `src/common/decorators/public.decorator.ts`
- [ ] Create `src/common/decorators/auth-provider.decorator.ts`
- [ ] Create `src/common/decorators/current-user.decorator.ts`

### Auth module wiring

- [ ] Create `src/modules/auth/auth.module.ts`
- [ ] Wire `useClass: FirebaseAuthAdapter` for `AUTH_PROVIDER` token
- [ ] Register all strategies as providers
- [ ] Register `AuthStrategyFactory` with `useFactory`
- [ ] Export `AuthGuard` for global registration

### Auth controller & service

- [ ] Create `src/modules/auth/dto/login.dto.ts` (Zod schema)
- [ ] Create `src/modules/auth/dto/refresh.dto.ts`
- [ ] Create `src/modules/auth/auth.service.ts`
- [ ] Implement `validateAndSync(token)` — verify + upsert user in User Service
- [ ] Implement `logout(uid)` — revoke Firebase tokens
- [ ] Implement `generateInternalJwt(userId)` — for service-to-service calls
- [ ] Create `src/modules/auth/auth.controller.ts`
- [ ] `POST /auth/validate` — validate token, return user
- [ ] `POST /auth/logout` — revoke tokens
- [ ] `POST /auth/refresh` — issue new internal JWT

### gRPC server

- [ ] Configure Nest gRPC microservice transport
- [ ] Implement `ValidateToken` RPC handler
- [ ] Implement `GetCurrentUser` RPC handler
- [ ] Add gRPC health check

### Tests

- [ ] Unit test `FirebaseAuthStrategy.validate()` with mock port
- [ ] Unit test `AuthStrategyFactory.resolve()` — valid and invalid providers
- [ ] Unit test `AuthGuard.canActivate()` — public, pinned, default
- [ ] E2E test `POST /auth/validate` with valid Firebase token
- [ ] E2E test `POST /auth/validate` with expired token → 401

---

## User Service `[Java]`

### Project setup

- [ ] Create Spring Boot 3 project (Spring Initializr)
- [ ] Add dependencies: Spring Web, Spring Data JPA, Lombok, Validation, Actuator
- [ ] Add PostgreSQL driver dependency
- [ ] Add Flyway migration dependency
- [ ] Add gRPC Spring Boot starter
- [ ] Add Kafka client dependency
- [ ] Add OpenTelemetry Java agent
- [ ] Configure `application.yml` with profiles (dev, prod)
- [ ] Write `.env.example`
- [ ] Write `Dockerfile`

### Database schema (Flyway migrations)

- [ ] `V1__create_users.sql` — id, uid, phone, email, name, avatar_url, deleted_at, created_at, updated_at
- [ ] `V2__create_squads.sql` — id, name, avatar_url, created_by, created_at
- [ ] `V3__create_squad_members.sql` — squad_id, user_id, role, joined_at
- [ ] `V4__create_device_tokens.sql` — id, user_id, token, platform, created_at
- [ ] `V5__create_user_face_profiles.sql` — user_id, cloudinary_id, embeddings (jsonb), updated_at
- [ ] `V6__create_friend_connections.sql` — user_id_a, user_id_b, created_at (unique pair)
- [ ] Add indexes: users(uid), users(phone), squad_members(squad_id, user_id)

### Entities & repositories

- [ ] Create `User` JPA entity
- [ ] Create `Squad` JPA entity
- [ ] Create `SquadMember` JPA entity with composite key
- [ ] Create `DeviceToken` JPA entity
- [ ] Create `UserFaceProfile` JPA entity
- [ ] Create `FriendConnection` JPA entity
- [ ] Create `UserRepository` — `findByUid`, `findByPhone`, `findById`, `softDelete`
- [ ] Create `SquadRepository` — `findById`, `findByCreatedBy`, `findSquadsForUser`
- [ ] Create `SquadMemberRepository` — `findBySquadId`, `findByUserId`
- [ ] Create `DeviceTokenRepository` — `findByUserId`, `deleteByToken`
- [ ] Create `FriendConnectionRepository` — `findFriendsOfUser`, `existsByPair`

### Service layer

- [ ] Create `UserService`
- [ ] Implement `createOrSync(uid, phone, name)` — idempotent upsert
- [ ] Implement `findById(id)` — excludes soft-deleted
- [ ] Implement `findByUid(uid)`
- [ ] Implement `update(id, dto)` — name, avatarUrl
- [ ] Implement `softDelete(id)` — sets deleted_at
- [ ] Create `SquadService`
- [ ] Implement `create(name, createdBy)`
- [ ] Implement `addMember(squadId, userId, role)`
- [ ] Implement `removeMember(squadId, userId)`
- [ ] Implement `getMembers(squadId)`
- [ ] Implement `getUserSquads(userId)`
- [ ] Implement `mergeSquads(squadId1, squadId2, eventId)`
- [ ] Create `DeviceTokenService`
- [ ] Implement `register(userId, token, platform)`
- [ ] Implement `getTokensForUser(userId)`
- [ ] Implement `getTokensForUsers(userIds)` — batch for notification service
- [ ] Implement `removeToken(token)` — on FCM delivery failure
- [ ] Create `FriendService`
- [ ] Implement `connect(userIdA, userIdB)`
- [ ] Implement `getFriends(userId)`
- [ ] Implement `getSuggestedSquads(userId)` — based on friend graph

### Controllers (REST)

- [ ] `POST /users` — create/sync user (called after Firebase signup)
- [ ] `GET /users/me` — get own profile
- [ ] `PATCH /users/me` — update name, avatar
- [ ] `DELETE /users/me` — soft delete account
- [ ] `GET /users/:id` — get user by ID (internal use)
- [ ] `POST /users/me/device-tokens` — register FCM token
- [ ] `DELETE /users/me/device-tokens/:token` — deregister token
- [ ] `POST /squads` — create squad
- [ ] `GET /squads` — get my squads
- [ ] `GET /squads/:id` — get squad with members
- [ ] `POST /squads/:id/members` — add member
- [ ] `DELETE /squads/:id/members/:userId` — remove member
- [ ] `GET /squads/:id/members` — list members
- [ ] `POST /users/me/face-profile` — upload reference photo
- [ ] `GET /users/me/friends` — get friend list
- [ ] `GET /users/me/friends/suggestions` — suggested squad members

### gRPC server

- [ ] Implement `GetUser(id)` RPC
- [ ] Implement `GetUsersByIds(ids)` batch RPC
- [ ] Implement `GetSquadMembers(squadId)` RPC
- [ ] Implement `GetDeviceTokens(userIds)` RPC

### Kafka events

- [ ] Publish `user.created` on new user sync
- [ ] Publish `user.deleted` on soft delete
- [ ] Listen `payment.confirmed` → emit user stat update

### Tests

- [ ] Unit test `UserService.createOrSync()` — duplicate calls are idempotent
- [ ] Unit test `SquadService.mergeSquads()`
- [ ] Integration test `UserRepository.findByUid()`
- [ ] E2E test `POST /users` → creates user in DB
- [ ] E2E test `PATCH /users/me` → updates name

---

## Event Service `[Java]`

### Project setup

- [ ] Spring Boot 3 project with same base deps as User Service
- [ ] Add `fsm` / hand-rolled state machine (no external lib)
- [ ] Configure Flyway, Kafka, gRPC
- [ ] Write `Dockerfile`

### Database schema

- [ ] `V1__create_events.sql` — id, squad_id, name, type (DRINKS/DINING/BOTH), status, created_by, deadline, created_at
- [ ] `V2__create_event_members.sql` — event_id, user_id, rsvp_status, is_guest, joined_at
- [ ] `V3__create_votes.sql` — event_id, user_id, venue_id, cast_at
- [ ] `V4__create_invite_links.sql` — token (uuid), event_id, created_by, expires_at, max_uses, use_count
- [ ] Add indexes on all foreign keys and status columns

### State machine

- [ ] Create `EventStatus` enum: `DRAFT`, `VOTING`, `CONFIRMED`, `ACTIVE`, `CLOSED`
- [ ] Create `EventStateMachine` — transition map
- [ ] Implement `transition(current, next)` — throws `InvalidTransitionException`
- [ ] Unit test every valid transition
- [ ] Unit test every invalid transition throws

### Entities & repositories

- [ ] Create `Event` JPA entity
- [ ] Create `EventMember` JPA entity
- [ ] Create `Vote` JPA entity
- [ ] Create `InviteLink` JPA entity
- [ ] Create `EventRepository`
- [ ] Create `EventMemberRepository`
- [ ] Create `VoteRepository`
- [ ] Create `InviteLinkRepository`

### Service layer

- [ ] Create `EventService`
- [ ] Implement `create(squadId, name, type, deadline, createdBy)`
- [ ] Implement `openVoting(eventId)` — transition DRAFT → VOTING
- [ ] Implement `confirmVenue(eventId, venueId)` — VOTING → CONFIRMED
- [ ] Implement `openTab(eventId)` — CONFIRMED → ACTIVE
- [ ] Implement `closeEvent(eventId)` — ACTIVE → CLOSED
- [ ] Implement `getEvent(id)`
- [ ] Implement `getSquadEvents(squadId)` — paginated
- [ ] Create `RsvpService`
- [ ] Implement `updateRsvp(eventId, userId, status)` — GOING / NOT_GOING / MAYBE
- [ ] Implement `getHeadcount(eventId)` — count of GOING members
- [ ] Publish `rsvp.updated` Kafka event on change
- [ ] Create `VotingService`
- [ ] Implement `castVote(eventId, userId, venueId)`
- [ ] Implement `getVoteTally(eventId)` — grouped by venue
- [ ] Implement `resolveWinner(eventId)` — called on deadline or all voted
- [ ] Enforce one vote per user per event
- [ ] Create `InviteService`
- [ ] Implement `generateInviteLink(eventId, createdBy, maxUses, expiresInHours)`
- [ ] Implement `redeemInvite(token, userId)` — adds user as guest member
- [ ] Implement `validateInvite(token)` — not expired, not over max uses
- [ ] Create `GuestService`
- [ ] Implement `joinAsGuest(inviteToken, phone, name)` — creates lightweight guest user

### Controllers (REST)

- [ ] `POST /events` — create outing
- [ ] `GET /events/:id` — get event detail
- [ ] `GET /events` — get events for my squads
- [ ] `PATCH /events/:id/status` — transition status
- [ ] `POST /events/:id/rsvp` — update my RSVP
- [ ] `GET /events/:id/rsvp` — get all RSVPs
- [ ] `POST /events/:id/votes` — cast vote
- [ ] `GET /events/:id/votes` — get tally
- [ ] `POST /events/:id/invite` — generate invite link
- [ ] `POST /invites/:token/redeem` — join via link
- [ ] `GET /events/:id/members` — list members + RSVP status

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

## Billing Service `[Java]`

### Project setup

- [ ] Spring Boot 3 project
- [ ] Add `kafka-clients` dependency
- [ ] Add gRPC for Payment Service communication
- [ ] Configure Flyway, Postgres
- [ ] Write `Dockerfile`

### Database schema

- [ ] `V1__create_bills.sql` — id, event_id, status (OPEN/CLOSED), split_mode, total_paise, created_at
- [ ] `V2__create_bill_items.sql` — id, bill_id, name, amount_paise, added_by, created_at, deleted_at
- [ ] `V3__create_item_assignments.sql` — item_id, user_id, share_paise
- [ ] `V4__create_debts.sql` — id, bill_id, from_user, to_user, amount_paise, status (PENDING/REQUESTED/SETTLED)
- [ ] `V5__create_audit_log.sql` — id, bill_id, action, payload (jsonb), performed_by, created_at
- [ ] `V6__create_split_configs.sql` — bill_id, mode (EQUAL/BY_ITEM/CUSTOM), config (jsonb)

### Command pattern

- [ ] Create `BillCommand` interface — `execute()`, `toAuditEntry()`
- [ ] Create `AddItemCommand` — adds item, writes audit log
- [ ] Create `RemoveItemCommand` — soft deletes item, writes audit log
- [ ] Create `AssignItemCommand` — assigns item to user
- [ ] Create `CloseBillCommand` — runs split, generates debts, closes bill
- [ ] Create `CommandInvoker` — executes command + persists audit entry atomically

### Strategy pattern

- [ ] Create `SplitStrategy` interface — `calculate(bill, members): List<Debt>`
- [ ] Create `EqualSplitStrategy` — divide total by member count
- [ ] Create `ByItemSplitStrategy` — sum assigned items per user
- [ ] Create `CustomSplitStrategy` — use explicit percentages from config
- [ ] Create `SplitStrategyFactory` — map from `SplitMode` enum

### Debt simplification graph

- [ ] Create `DebtGraph` class
- [ ] Implement `addDebt(from, to, amount)` — build net balance map
- [ ] Implement `simplify()` — min-cash-flow algorithm
- [ ] Return `List<SimplifiedDebt>` — minimum number of transfers
- [ ] Unit test with 3-person circular debt (A owes B, B owes C, C owes A)
- [ ] Unit test with 5-person complex debt
- [ ] Unit test with all debts already settled

### Service layer

- [ ] Create `BillService`
- [ ] Implement `createBill(eventId, splitMode)` — called when tab opened
- [ ] Implement `addItem(billId, name, amount, addedBy)` — via command
- [ ] Implement `removeItem(billId, itemId, removedBy)` — via command
- [ ] Implement `assignItem(billId, itemId, userIds)` — via command
- [ ] Implement `closeBill(billId)` — run strategy, simplify, create debts
- [ ] Implement `getBill(billId)` — with items and debts
- [ ] Create `DebtService`
- [ ] Implement `getDebtsForUser(userId, eventId)`
- [ ] Implement `requestSettlement(debtId)` — calls Payment Service gRPC
- [ ] Implement `markSettled(debtId)` — called by Kafka consumer
- [ ] Implement `getTotalOwed(userId)` — across all open debts
- [ ] Implement `getTotalOwing(userId)` — across all open debts

### Controllers (REST)

- [ ] `POST /bills` — create bill for event
- [ ] `GET /bills/:id` — get bill with items
- [ ] `POST /bills/:id/items` — add line item
- [ ] `DELETE /bills/:id/items/:itemId` — remove item
- [ ] `POST /bills/:id/items/:itemId/assign` — assign to users
- [ ] `POST /bills/:id/close` — close and generate debts
- [ ] `GET /bills/:id/debts` — get all debts for bill
- [ ] `GET /debts/me` — my debts across all events
- [ ] `POST /debts/:id/settle` — initiate settlement
- [ ] `GET /bills/:id/audit` — audit log

### Kafka consumers

- [ ] Listen `event.closed` → auto-close open bill
- [ ] Listen `payment.confirmed` → mark debt settled

### Kafka publishers

- [ ] Publish `tab.opened` on bill create
- [ ] Publish `tab.closed` on bill close
- [ ] Publish `debt.created` for each new debt
- [ ] Publish `debt.settled` when debt status → SETTLED

### Tests

- [ ] Unit test `EqualSplitStrategy` — 3 people, ₹900 → ₹300 each
- [ ] Unit test `ByItemSplitStrategy` — uneven item assignments
- [ ] Unit test `DebtGraph.simplify()` — multiple scenarios
- [ ] Unit test `CommandInvoker` — audit log entry created
- [ ] Integration test full bill lifecycle

---

## Payment Service `[Go]`

### Project setup

- [ ] `go mod init roundup/payment`
- [ ] Add `gin-gonic/gin` dependency
- [ ] Add `google.golang.org/grpc` dependency
- [ ] Add `segmentio/kafka-go` dependency
- [ ] Add `redis/go-redis` dependency
- [ ] Configure `internal/config/` — env loading
- [ ] Write `Dockerfile`

### Juspay / Hyperswitch integration

- [ ] Create `internal/juspay/client.go` — HTTP client with base URL, API key
- [ ] Implement `CreateCollectRequest(payorId, payeeId, amount, currency)` → `CollectRequest`
- [ ] Implement `GetPaymentStatus(paymentId)` → `PaymentStatus`
- [ ] Implement `CreateRefund(paymentId, amount)` → `Refund`
- [ ] Implement `VerifyWebhookSignature(payload, signature)` → `bool`
- [ ] Add retry logic (3 attempts, exponential backoff) for all API calls
- [ ] Add circuit breaker (open after 5 failures in 10s)

### Idempotency

- [ ] Create `internal/idempotency/store.go` — Redis-backed
- [ ] Implement `Check(key)` → `(existingResponse, bool)`
- [ ] Implement `Store(key, response, ttl)` → `error`
- [ ] Use `debt_id` as idempotency key for collect requests
- [ ] Return cached response if key already exists (no duplicate API call)

### Webhook handling

- [ ] Create `POST /webhooks/juspay` endpoint
- [ ] Verify signature on every webhook
- [ ] Parse `payment.captured` event → publish `payment.confirmed` Kafka
- [ ] Parse `payment.failed` event → publish `payment.failed` Kafka
- [ ] Parse `refund.created` event → update refund record
- [ ] Respond `200 OK` immediately (process async)
- [ ] Store raw webhook payload in `payment_events` table for replay

### Database schema

- [ ] `V1__create_payment_requests.sql` — id, debt_id, juspay_payment_id, status, amount_paise, created_at
- [ ] `V2__create_payment_events.sql` — id, payment_id, event_type, raw_payload (jsonb), created_at
- [ ] `V3__create_refunds.sql` — id, payment_id, amount_paise, status, created_at
- [ ] Add `sqlc` config and generate Go types from schema

### gRPC server

- [ ] Implement `InitiatePayment(debtId, fromUserId, toUserId, amount)` RPC
- [ ] Implement `GetPaymentStatus(debtId)` RPC

### Kafka consumers / publishers

- [ ] Publish `payment.initiated` on new collect request
- [ ] Publish `payment.confirmed` on successful webhook
- [ ] Publish `payment.failed` on failed webhook

### Service layer

- [ ] Create `PaymentService` — orchestrates Juspay client + idempotency + DB
- [ ] Implement `InitiateCollect(debtId, ...)` — idempotency check first
- [ ] Implement `HandleWebhook(payload)` — verify + store + publish
- [ ] Implement `ReplayWebhook(eventId)` — for ops / admin

### Tests

- [ ] Unit test idempotency store — same key returns same response
- [ ] Unit test webhook signature verification
- [ ] Unit test circuit breaker — opens after threshold failures
- [ ] Integration test `InitiateCollect` → Juspay sandbox

---

## Venue Service `[Go]`

### Project setup

- [ ] `go mod init roundup/venue`
- [ ] Add `gin-gonic/gin`
- [ ] Add `googlemaps/google-maps-services-go`
- [ ] Add `redis/go-redis`
- [ ] Write `Dockerfile`

### Google Places integration

- [ ] Create `internal/places/client.go` — wraps Google Maps SDK
- [ ] Implement `NearbySearch(lat, lng, radius, type, keyword)` → `[]Place`
- [ ] Implement `GetPlaceDetails(placeId, fields)` → `PlaceDetail`
- [ ] Implement `GetPlacePhotos(placeId)` → `[]PhotoRef`
- [ ] Map Google Place response to internal `Venue` struct
- [ ] Add field mask to only fetch needed fields (cost control)

### Redis cache layer

- [ ] Create `internal/cache/venue_cache.go`
- [ ] Cache key: `venues:nearby:{lat}:{lng}:{type}:{keyword}` — 24hr TTL
- [ ] Cache key: `venue:detail:{placeId}` — 1hr TTL
- [ ] Implement `GetCached(key)` → `([]Venue, bool)`
- [ ] Implement `SetCached(key, venues, ttl)` → `error`
- [ ] On cache miss — fetch from Places API, then cache

### Database schema (venue history)

- [ ] `V1__create_saved_venues.sql` — id, squad_id, user_id, place_id, name, saved_at
- [ ] `V2__create_venue_visits.sql` — id, squad_id, event_id, place_id, visited_at, avg_spend_paise

### Service layer

- [ ] Create `VenueService`
- [ ] Implement `Search(lat, lng, query, filters)` — cache-first
- [ ] Implement `GetDetail(placeId)` — cache-first
- [ ] Implement `SaveFavourite(squadId, userId, placeId)`
- [ ] Implement `GetFavourites(squadId)`
- [ ] Implement `RecordVisit(squadId, eventId, placeId, avgSpend)`
- [ ] Implement `GetVisitHistory(squadId)` — sorted by visit count

### Controllers (REST)

- [ ] `GET /venues/search?lat=&lng=&q=&type=` — nearby search
- [ ] `GET /venues/:placeId` — place details
- [ ] `POST /venues/favourites` — save favourite
- [ ] `GET /venues/favourites?squadId=` — get favourites
- [ ] `GET /venues/history?squadId=` — visit history

### Kafka consumer

- [ ] Listen `event.confirmed` → call `RecordVisit`

### Tests

- [ ] Unit test cache hit — Places API not called
- [ ] Unit test cache miss — Places API called, result cached
- [ ] Integration test search with real Places API (dev key)

---

## Notification Service `[Go]`

### Project setup

- [ ] `go mod init roundup/notification`
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
- [ ] Template: `payment_confirmed` — "₹X received from Y"

### Kafka consumers

- [ ] Listen `rsvp.updated` → notify organiser
- [ ] Listen `vote.cast` → notify all squad members (tally update)
- [ ] Listen `tab.closed` → notify all members (debts ready)
- [ ] Listen `debt.created` → notify debtor
- [ ] Listen `debt.settled` → notify creditor
- [ ] Listen `payment.confirmed` → notify both parties
- [ ] Listen `photo.tagged` → notify tagged user
- [ ] Listen `outing.recap.requested` → send recap when AI responds
- [ ] Listen `notification.send` — generic topic for AI-generated messages

### Notification routing

- [ ] Create `NotificationRouter` — decides channel (FCM vs WhatsApp vs both)
- [ ] Route by user preference (from User Service)
- [ ] Route by notification type (payment → WhatsApp, photo tag → FCM)
- [ ] Implement quiet hours check before sending

### Notification log

- [ ] `V1__create_notification_log.sql` — id, user_id, channel, type, status, sent_at
- [ ] Write log entry on every send attempt
- [ ] Update status on delivery receipt (FCM success/failure)

### Tests

- [ ] Unit test `NotificationRouter` — correct channel selection
- [ ] Unit test quiet hours enforcement
- [ ] Integration test FCM send with Firebase test credentials

---

## Media Service `[Go]`

### Project setup

- [ ] `go mod init roundup/media`
- [ ] Add `gin-gonic/gin`
- [ ] Add `cloudinary/cloudinary-go`
- [ ] Add `segmentio/kafka-go`
- [ ] Add `golang-jwt/jwt` for download tokens
- [ ] Add `mholt/archiver` for ZIP streaming
- [ ] Write `Dockerfile`

### Database schema

- [ ] `V1__create_photos.sql` — id, event_id, uploaded_by, cloudinary_id, width, height, tagging_status, created_at
- [ ] `V2__create_photo_faces.sql` — id, photo_id, user_id (nullable), confidence, bounding_box (jsonb)
- [ ] `V3__create_download_tokens.sql` — token, user_id, event_id, expires_at, used_at

### Cloudinary integration

- [ ] Create `internal/cloudinary/client.go`
- [ ] Implement `GenerateSignedUploadPreset(eventId, userId)` → `SignedParams`
- [ ] Implement `BuildUrl(cloudinaryId, transformations)` → `string`
- [ ] Implement `DeleteAsset(cloudinaryId)` → `error`
- [ ] Implement `CropFace(cloudinaryId, boundingBox)` → `string` (crop URL)

### Webhook handler

- [ ] `POST /webhooks/cloudinary` — receive upload complete event
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
- [ ] Create download endpoint `GET /media/download/:token`
- [ ] Verify JWT signature
- [ ] Check not expired
- [ ] Fetch all photos tagged with `userId` for `eventId`
- [ ] Stream ZIP: pipe each Cloudinary URL into `archiver` → response
- [ ] Never buffer entire ZIP to disk
- [ ] Mark token `used_at` after download

### Controllers (REST)

- [ ] `GET /media/upload-params?eventId=` — returns signed Cloudinary params
- [ ] `POST /webhooks/cloudinary` — upload webhook
- [ ] `GET /media/qr/:eventId` — generate and return QR PNG
- [ ] `GET /media/download/:token` — stream ZIP (no auth guard)
- [ ] `GET /events/:id/photos` — list photos for event
- [ ] `GET /events/:id/photos/tagged` — my tagged photos in event
- [ ] `POST /photos/:id/tag` — manual self-tag

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

- [ ] `go mod init roundup/realtime`
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
- [ ] `GET /events/:id/presence` — REST endpoint for current viewers

### Kafka consumers (fan out to WebSocket clients)

- [ ] Listen `rsvp.updated` → broadcast to event room
- [ ] Listen `vote.cast` → broadcast tally update to event room
- [ ] Listen `tab.opened` → broadcast to event members
- [ ] Listen `debt.created` → send to specific user
- [ ] Listen `payment.confirmed` → send to both parties
- [ ] Listen `photo.tagged` → send to tagged user
- [ ] Listen `event.confirmed` → broadcast venue to event room

### Message format

- [ ] Define `WsMessage` struct — `{ type, payload, eventId, timestamp }`
- [ ] Types: `RSVP_UPDATE`, `VOTE_UPDATE`, `TAB_ITEM_ADDED`, `DEBT_CREATED`, `PAYMENT_CONFIRMED`, `PHOTO_TAGGED`, `VENUE_CONFIRMED`, `USER_TYPING`, `USER_PRESENCE`
- [ ] Client sends `TYPING` message → server broadcasts to event members

### Tests

- [ ] Unit test Hub concurrent register/unregister
- [ ] Unit test broadcast to event — only connected users receive
- [ ] Integration test WebSocket connect + receive Kafka-triggered message

---

## Reservation Service `[Go]`

### Project setup

- [ ] `go mod init roundup/reservation`
- [ ] Add `gin-gonic/gin`
- [ ] Add `segmentio/kafka-go`
- [ ] Write `Dockerfile`
- [ ] Research Dineout / EazyDiner partner API access — submit request

### Database schema

- [ ] `V1__create_reservations.sql` — id, event_id, venue_place_id, provider, reservation_ref, status, party_size, reservation_time, created_at
- [ ] `V2__create_reservation_events.sql` — id, reservation_id, event_type, payload, created_at

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

- [ ] `POST /webhooks/dineout` — receive confirmation / cancellation
- [ ] Verify provider webhook signature
- [ ] Update reservation status
- [ ] Publish Kafka event on status change

### Controllers (REST)

- [ ] `GET /reservations/availability?eventId=&venueId=&date=&partySize=`
- [ ] `POST /reservations` — book a table
- [ ] `DELETE /reservations/:id` — cancel
- [ ] `GET /reservations/:eventId` — get event reservation

### Tests

- [ ] Unit test state machine transitions
- [ ] Unit test Dineout client response mapping
- [ ] Integration test booking flow with sandbox

---

## AI Service `[Python]`

### Project setup

- [ ] `pip install fastapi uvicorn langchain langchain-google-genai python-dotenv sqlalchemy asyncpg kafka-python`
- [ ] Create `pyproject.toml`
- [ ] Create `app/main.py` — FastAPI app entry
- [ ] Create `app/config.py` — Pydantic settings from env
- [ ] Configure pgvector extension in Postgres (`CREATE EXTENSION vector`)
- [ ] Write `Dockerfile` (python:3.11-slim)
- [ ] Write `.env.example`

### LangChain setup

- [ ] Configure `ChatGoogleGenerativeAI` (Gemini 1.5 Pro) as default LLM
- [ ] Configure `GoogleGenerativeAIEmbeddings` for embeddings
- [ ] Configure `PGVector` vector store (connect to same Postgres)
- [ ] Create `app/llm/base.py` — shared LLM + embedding instances

### Receipt OCR chain

- [ ] Create `app/chains/receipt_ocr.py`
- [ ] Implement chain: image URL → Vision API → LLM extraction → JSON
- [ ] Prompt: extract `[{name, quantity, unit_price_paise, total_paise}]`
- [ ] Validate output with Pydantic model
- [ ] Handle extraction failures gracefully (return empty list)
- [ ] `POST /ai/receipt/extract` — accepts `{ cloudinary_url, event_id }`
- [ ] Return extracted line items for Billing Service to create

### Venue recommendation agent

- [ ] Create `app/agents/venue_recommender.py`
- [ ] Define tools: `search_venues(lat, lng, query)`, `get_visit_history(squad_id)`, `get_squad_preferences(squad_id)`
- [ ] Implement LangChain agent with tool calling
- [ ] Agent prompt: "Recommend venues given past behaviour and current request"
- [ ] `POST /ai/venues/recommend` — accepts squad context, returns ranked venues

### RAG over squad history

- [ ] Create `app/rag/squad_history.py`
- [ ] Embed past outings on event close: venue, spend, members, date
- [ ] Store embeddings in pgvector `squad_memories` table
- [ ] Implement retriever — top 5 most relevant past outings
- [ ] Create RAG chain: query → retrieve → LLM answer
- [ ] `POST /ai/query` — natural language question over squad history
- [ ] Listen Kafka `event.closed` → embed and store outing

### Smart nudge generation

- [ ] Create `app/chains/nudge_generator.py`
- [ ] Context: debt amount, relationship (who owes whom), outing name, days overdue
- [ ] Generate personalised nudge message (casual, friendly tone)
- [ ] `POST /ai/nudges/generate` — returns message text for Notification Service

### Outing recap

- [ ] Create `app/chains/recap_generator.py`
- [ ] Input: event name, venue, members, total spend, photo count, debts settled
- [ ] Output: 2-3 sentence human recap + shareable card text
- [ ] `POST /ai/recap/generate`
- [ ] Listen Kafka `event.closed` → auto-generate and publish `outing.recap.requested`

### Caption generation

- [ ] Create `app/chains/caption_generator.py`
- [ ] Input: photo description from Vision API scene detection
- [ ] Generate Instagram-style caption options (3 variations)
- [ ] `POST /ai/captions/generate`

### Spend prediction

- [ ] Create `app/models/spend_predictor.py`
- [ ] Input: squad visit history at venue, time of year, group size
- [ ] Output: predicted spend range
- [ ] `POST /ai/predict/spend`

### Best day suggestion

- [ ] Analyse squad RSVP acceptance patterns by day of week
- [ ] Return day with highest historical turnout
- [ ] `GET /ai/suggest/day?squadId=`

### Tests

- [ ] Unit test receipt OCR extraction with sample bill image
- [ ] Unit test nudge generator — tone is casual not robotic
- [ ] Unit test RAG retrieval — relevant outings returned
- [ ] Integration test full recommendation agent with mock tools

---

## Analytics Service `[Java]`

### Project setup

- [ ] Spring Boot 3 project
- [ ] Add Spring Data JPA, Flyway, Kafka, Actuator
- [ ] Write `Dockerfile`

### Database schema

- [ ] `V1__create_analytics_events.sql` — id, squad_id, user_id, event_type, payload (jsonb), occurred_at
- [ ] `V2__squad_stats_view.sql` — materialized view: total outings, total spend, avg per outing per squad
- [ ] `V3__user_stats_view.sql` — materialized view: per-user stats across squads
- [ ] `V4__venue_stats_view.sql` — most visited venues per squad
- [ ] `V5__monthly_summary_view.sql` — monthly aggregates
- [ ] Add `pg_cron` job: refresh all views nightly at 2am

### Kafka consumers (event ingestion)

- [ ] Listen `event.created` → store analytics event
- [ ] Listen `event.closed` → store analytics event
- [ ] Listen `rsvp.updated` → store analytics event
- [ ] Listen `debt.created` → store analytics event
- [ ] Listen `debt.settled` → store analytics event
- [ ] Listen `photo.uploaded` → store analytics event
- [ ] Listen `payment.confirmed` → store analytics event

### Service layer

- [ ] Create `SquadStatsService`
- [ ] Implement `getSquadStats(squadId)` — from materialized view
- [ ] Implement `getLeaderboard(squadId)` — biggest spender, most RSVPs, most photos
- [ ] Implement `getVenueHistory(squadId)` — sorted by visit count
- [ ] Implement `getSpendByMonth(squadId, year)` — monthly breakdown
- [ ] Create `UserStatsService`
- [ ] Implement `getUserStats(userId)` — total outings, total spent, squads
- [ ] Create `WrappedService`
- [ ] Implement `generateWrapped(squadId, year)` — full year summary

### Controllers (REST)

- [ ] `GET /analytics/squads/:id/stats` — squad summary stats
- [ ] `GET /analytics/squads/:id/leaderboard` — rankings
- [ ] `GET /analytics/squads/:id/venues` — venue history
- [ ] `GET /analytics/squads/:id/spend?year=` — monthly spend
- [ ] `GET /analytics/users/:id/stats` — user stats
- [ ] `GET /analytics/squads/:id/wrapped?year=` — yearly wrapped

### Tests

- [ ] Unit test leaderboard calculation
- [ ] Unit test monthly spend aggregation
- [ ] Integration test materialized view refresh

---

## Feature Flag Service `[Go]`

### Project setup

- [ ] `go mod init roundup/flags`
- [ ] Add `gin-gonic/gin`
- [ ] Add `redis/go-redis`
- [ ] Add `google.golang.org/grpc` for gRPC server
- [ ] Write `Dockerfile`

### Database schema

- [ ] `V1__create_feature_flags.sql` — key, enabled, rollout (0-100), metadata (jsonb), updated_at
- [ ] `V2__create_user_overrides.sql` — flag_key, user_id, enabled, created_at

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

- [ ] `GET /flags` — list all flags
- [ ] `POST /flags` — create flag
- [ ] `PATCH /flags/:key` — update enabled, rollout
- [ ] `DELETE /flags/:key` — delete flag
- [ ] `POST /flags/:key/overrides` — set user-level override
- [ ] `DELETE /flags/:key/overrides/:userId` — remove override

### Tests

- [ ] Unit test rollout — userId always maps to same bucket
- [ ] Unit test user override takes precedence over global
- [ ] Unit test cache invalidation on flag update

---

## Admin Service `[NestJS]`

### Project setup

- [ ] `nest new admin-service`
- [ ] Add `@nestjs/jwt`, `zod`, `ioredis`
- [ ] Admin-only JWT separate from user JWT
- [ ] Write `Dockerfile`

### Auth

- [ ] Admin login — email + password (bcrypt hashed)
- [ ] Issue admin JWT (short TTL, 1hr)
- [ ] Admin JWT guard on all routes
- [ ] Role-based: `SUPER_ADMIN`, `OPS`, `SUPPORT`

### User management

- [ ] `GET /admin/users?q=` — search users by phone, email, name
- [ ] `GET /admin/users/:id` — full user detail
- [ ] `POST /admin/users/:id/ban` — soft ban (sets banned_at)
- [ ] `POST /admin/users/:id/unban`
- [ ] `POST /admin/users/:id/impersonate` — generate token for support
- [ ] `GET /admin/users/:id/squads` — user's squads
- [ ] `GET /admin/users/:id/debts` — open debts

### Payment ops

- [ ] `GET /admin/payments?status=&from=&to=` — payment log
- [ ] `POST /admin/payments/:id/refund` — trigger refund via Payment Service
- [ ] `POST /admin/webhooks/:id/replay` — replay failed webhook
- [ ] `GET /admin/payments/disputes` — flagged disputes

### Feature flag admin

- [ ] `GET /admin/flags` — all flags with metadata
- [ ] `PATCH /admin/flags/:key` — update flag (proxies to Flag Service)
- [ ] `POST /admin/flags/:key/overrides` — per-user override

### Abuse detection

- [ ] `GET /admin/abuse/suspicious-debts` — unusually large debts
- [ ] `GET /admin/abuse/fake-invites` — invite links with abnormal redemption rate
- [ ] `POST /admin/abuse/users/:id/flag` — flag user for review

### Venue partner ops

- [ ] `GET /admin/venues/partners` — all partner venues
- [ ] `POST /admin/venues/partners` — add partner
- [ ] `PATCH /admin/venues/partners/:id` — update deal, rollout %
- [ ] `GET /admin/venues/partners/:id/analytics` — visits, revenue

### Tests

- [ ] Unit test role-based access — OPS cannot access SUPER_ADMIN routes
- [ ] E2E test impersonation token — expires in 1hr
- [ ] E2E test webhook replay — re-publishes to Kafka

---

## Flutter App

### Project setup

- [ ] `flutter create roundup_app`
- [ ] Add `go_router` dependency
- [ ] Add `riverpod` (or `bloc`) for state management
- [ ] Add `firebase_core`, `firebase_auth`, `firebase_messaging` dependencies
- [ ] Add `cloud_firestore` dependency (live bill tab)
- [ ] Add `dio` for HTTP client
- [ ] Add `qr_flutter` for QR display
- [ ] Add `image_picker` for photo uploads
- [ ] Add `web_socket_channel` for realtime
- [ ] Configure flavors: dev, prod

### Auth screens

- [ ] Phone number input screen
- [ ] OTP verification screen
- [ ] Profile setup screen (name, avatar)
- [ ] Firebase Auth integration — `signInWithPhoneNumber`
- [ ] Store returned user token — send to API on every request
- [ ] Auto-login on app launch if token valid

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
- [ ] Receipt camera screen — upload to Cloudinary, show OCR result
- [ ] Split mode selector — equal, by item, custom
- [ ] Close tab confirmation screen

### Payment screens

- [ ] Debts overview screen — who owes me, what I owe
- [ ] Settle up screen — show UPI collect QR
- [ ] Payment history screen per event
- [ ] Running balance screen — across all squads

### Memory wall screens

- [ ] Event album screen — grid of photos
- [ ] Photo viewer screen — full screen, swipe
- [ ] Face tagging screen — tap untagged face → "that's me"
- [ ] QR code screen — show download QR
- [ ] Outing recap card screen — AI-generated summary

### Notification handling

- [ ] Request FCM permission on first launch
- [ ] Register device token with User Service
- [ ] Handle foreground notification — show in-app banner
- [ ] Handle background notification tap — deep link to relevant screen
- [ ] Notification preferences screen

### Realtime WebSocket

- [ ] Connect WebSocket on app launch (authenticated)
- [ ] Handle `RSVP_UPDATE` — update headcount on planning screen
- [ ] Handle `VOTE_UPDATE` — update tally on voting screen
- [ ] Handle `TAB_ITEM_ADDED` — append item on bill tab screen
- [ ] Handle `PAYMENT_CONFIRMED` — update debt status live
- [ ] Handle `PHOTO_TAGGED` — show notification in-app
- [ ] Reconnect on disconnect (exponential backoff)

### AI features (Flutter side)

- [ ] Chat input screen — planning assistant
- [ ] AI venue recommendation list screen
- [ ] "Ask your history" screen — natural language query
- [ ] Spend prediction chip on venue detail screen

### Settings

- [ ] Profile settings screen
- [ ] Notification preferences screen
- [ ] Linked squads screen
- [ ] Delete account flow (with confirmation)
- [ ] GDPR data export request button

---

## Cross-cutting

### API contracts

- [ ] Document all REST endpoints in OpenAPI 3.0 YAML
- [ ] Host Swagger UI for dev environment
- [ ] Write contract tests between gateway and each service (Pact)
- [ ] Version all APIs (`/api/v1/`)

### Security

- [ ] Rotate all API keys every 90 days (automated)
- [ ] Enable HTTPS only on all Cloud Run services
- [ ] Add Content Security Policy headers
- [ ] Add CORS — whitelist only app bundle ID and admin domain
- [ ] Penetration test auth flow (Firebase token forgery)
- [ ] Penetration test payment webhook (signature bypass)
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
- [ ] Write `RUNBOOK.md` — common ops procedures (restart service, replay webhook, manual refund)
- [ ] Write `ADR/` folder — Architecture Decision Records for key choices

---

_Total services: 15 | Languages: Go, Java, NestJS, Python | Estimated team: 4 people × 6 months_
