# Roundup

Weekend squad outing planner. Plan where to go, track who's coming, split the bill, settle up, remember the night.

> **Language policy:** Backend is Go only. Mobile app is Compose Multiplatform (Kotlin). No other languages.

---

## What's this?

Roundup is a mobile app (Compose Multiplatform — Android + iOS) backed by a Go microservices architecture. It handles the full lifecycle of a group outing — from venue voting and RSVP tracking to live bill splitting and keeping track of who owes whom.

This repository is a monorepo containing all backend services, shared proto definitions, and infrastructure config.

---

## Tech stack

| Layer                          | Technology                     |
| ------------------------------ | ------------------------------ |
| Mobile                         | Compose Multiplatform (Kotlin) |
| All backend services + gateway | Go (Fiber v3)                  |
| Auth                           | Custom — phone OTP + JWT (Go)  |
| DB access / migrations         | sqlc + pgx / goose             |
| Message broker                 | Apache Kafka                   |
| Cache                          | Redis                          |
| Database                       | PostgreSQL (one per service)   |
| Real-time sync                 | Firestore (live bill tab only) |
| Media storage                  | Cloudinary                     |
| Push                           | Firebase Cloud Messaging       |

---

## Repository structure

```
roundup/
├── services/           # one folder per microservice
├── proto/              # shared gRPC definitions
├── shared/             # constants shared across services
├── infra/              # docker, k8s, terraform configs
├── scripts/            # dev helper scripts (added as needed)
├── docs/               # architecture docs, ADRs
├── docker-compose.yml  # local dev environment
└── Makefile            # convenience commands
```

---

## Prerequisites

Install these before anything else.

| Tool           | Version  | Install                                      |
| -------------- | -------- | -------------------------------------------- |
| Docker Desktop | latest   | https://docs.docker.com/get-docker           |
| Git            | 2.x+     | https://git-scm.com                          |
| Make           | any      | pre-installed on macOS / Linux               |
| Go             | 1.27+    | https://go.dev/dl                            |
| goose          | latest   | `make install`                               |
| sqlc           | latest   | `make install`                               |
| buf            | latest   | https://buf.build/docs/installation          |
| golangci-lint  | latest   | https://golangci-lint.run/usage/install      |
| Android Studio | latest   | https://developer.android.com/studio (mobile only) |
| Xcode          | latest   | Mac App Store (mobile, iOS builds only)      |

---

## Getting started

### 1. Clone

```bash
git clone git@github.com:your-org/roundup.git
cd roundup
```

### 2. Copy environment files

```bash
make setup
```

This creates `.env.beta` from `.env.beta.example` for every service (existing files are kept). Real env files are gitignored; they live only on your machine.

Each service reads `.env.<APP_ENV>` from its own folder at startup. `APP_ENV` is `beta` or `prod`, and the Makefile defaults it to `beta`:

| File | Committed? | Used for |
|---|---|---|
| `.env.beta.example`, `.env.prod.example` | yes | templates listing every variable |
| `.env.beta` | no | local development and the beta environment |
| `.env.prod` | no | running locally against production resources; copy `.env.prod.example` by hand |

```bash
make user-run                 # APP_ENV=beta → services/user-service/.env.beta
make user-run APP_ENV=prod    # → services/user-service/.env.prod
```

Deployed containers don't use env files (they are excluded from images by `.dockerignore`). Set `APP_ENV` and every variable as real environment variables on the platform; those always take precedence over a file.

`APP_ENV` is required: a service refuses to start if it is unset or anything other than `beta` or `prod` (exact, lowercase). Running `go run` directly instead of through `make` therefore needs it too, e.g. `APP_ENV=beta go run ./cmd`.

### 3. Fill in environment variables

Open each `services/*/.env.beta` file and fill in the required values.
See [Environment variables](#environment-variables) below for what you need.

### 4. Start the databases

```bash
docker compose up -d
```

Starts one Postgres per service (auth, user, venue) and Redis. The Go services are not started by this.

```bash
docker compose ps   # all containers should show as healthy
```

### 5. Run migrations

```bash
make auth-migrate-up user-migrate-up venue-migrate-up
```

### 6. Start the services

Either in Docker (all four, logs in Docker Desktop):

```bash
make services-up
```

Or natively, for active development; see [Running a service locally](#running-a-service-locally-outside-docker).

### 7. Verify

```bash
curl http://localhost:8080/ready
# {"success":true,"data":{"status":"ready"}}
```

---

## Environment variables

All third-party services have free sandbox / test modes. No real money changes hands during development.

### Auth (custom — phone OTP + JWT)

No Firebase Auth. auth-service sends a one-time code by SMS and issues its own tokens: a 15-minute access JWT signed with **Ed25519**, and a 30-day refresh token that rotates on every use. In beta, codes are written to auth-service's log instead of sent, and `OTP_DEV_CODE` is also accepted (auth-service refuses to start if it is set with `APP_ENV=prod`).

```bash
make setup      # creates services/*/.env.beta
make jwt-keys   # private key → auth-service/.env.beta, public key → gateway/.env.beta
```

```env
# services/auth-service/.env.beta
JWT_PRIVATE_KEY="-----BEGIN PRIVATE KEY-----\n…"   # written by make jwt-keys
OTP_DEV_CODE="123456"
```

Only auth-service holds the private key; the gateway verifies tokens with the public key and can't mint them. Use a separate keypair per environment.

### Google Maps / Places

1. [console.cloud.google.com](https://console.cloud.google.com) → enable **Places API (New)**
2. Create an API key → restrict it to Places API only
3. Billing → Budgets & Alerts → set an alert at $0.01 so you're notified before any charge
4. Free tier: 10,000 calls/month per SKU — more than enough for development

```env
# services/venue-service/.env.beta
GOOGLE_MAPS_API_KEY=AIza...
```

### Cloudinary (media)

1. [cloudinary.com](https://cloudinary.com) → sign up — no credit card needed
2. Dashboard → copy Cloud Name, API Key, API Secret

```env
# services/media-service/.env.beta
CLOUDINARY_CLOUD_NAME=...
CLOUDINARY_API_KEY=...
CLOUDINARY_API_SECRET=...
```

Free plan: 25 credits/month. Sufficient for development and small-scale testing.

---

## Daily workflow

```bash
# databases only (always needed)
docker compose up -d

# all four services in Docker, rebuilt from the current code
make services-up

# follow their logs (also in Docker Desktop: Containers → roundup-monorepo → a service → Logs)
make services-logs

# stop the services, keep the databases running
make services-down

# run tests for a service
make auth-test        # also user-test, venue-test, gateway-test

# stop everything (data is kept in volumes)
docker compose --profile services down
```

Services in Docker read the same `services/*/.env.<APP_ENV>` files as native runs (`make services-up APP_ENV=prod` for prod; beta by default); `docker-compose.yml` only swaps `localhost` addresses for container names. Docker and native runs use the same ports, so run `make services-down` before `make *-run`.

---

## Running a service locally (outside Docker)

Faster for active development — run the service you are working on natively, with only infra in Docker.

```bash
docker compose up -d   # databases only
```

Every service is Go and runs the same way:

```bash
cd services/venue-service
go run ./cmd
```

Or use the Makefile shortcuts, e.g. `make auth-run`, `make user-run`, `make venue-run`, `make gateway-run`.

---

## Services

| Service              | Language | Port | Description                             |
| -------------------- | -------- | ---- | --------------------------------------- |
| gateway              | Go       | 8080 | Routes /api/v1, verifies access tokens |
| auth-service         | Go       | 3001 | Phone OTP login, JWT issuing            |
| user-service         | Go       | 3002 | Profiles, squads, device tokens         |
| event-service        | Go       | 3003 | Outing lifecycle, RSVP, voting          |
| billing-service      | Go       | 3004 | Bill tab, split strategies, debt graph  |
| venue-service        | Go       | 3006 | Google Places facade, Redis cache       |
| notification-service | Go       | 3007 | FCM push, WhatsApp                      |
| media-service        | Go       | 3008 | Photos, face detection, QR download     |
| realtime-service     | Go       | 3009 | WebSocket hub, presence                 |
| reservation-service  | Go       | 3011 | Table booking via Dineout               |
| analytics-service    | Go       | 3012 | Squad stats, leaderboards, wrapped      |
| feature-flag-service | Go       | 3013 | Flag evaluation, rollout control        |
| admin-service        | Go       | 3014 | Internal ops dashboard                  |

All services share Redis on `6379` and Kafka on `9092`. All HTTP routes are versioned under `/api/v1` (e.g. `/api/v1/users/me`).

> **Later:** AI features — deferred, will be revisited.

---

## Local tooling

Started automatically with `make infra`.

| Tool            | URL                   | Purpose                      |
| --------------- | --------------------- | ---------------------------- |
| Kafka UI        | http://localhost:8090 | Browse topics, messages, lag |
| Redis Commander | http://localhost:8091 | Browse Redis keys            |
| Gateway         | http://localhost:8080 | All API traffic entry point  |

---

## Kafka topics

Inter-service events flow through Kafka. Topic name constants live in `shared/kafka`.

| Topic                    | Publisher       | Consumers                         |
| ------------------------ | --------------- | --------------------------------- |
| `rsvp.updated`           | event-service   | notification, realtime, analytics |
| `event.created`          | event-service   | analytics                         |
| `event.closed`           | event-service   | billing, analytics                |
| `tab.opened`             | billing-service | realtime                          |
| `tab.closed`             | billing-service | notification                      |
| `debt.created`           | billing-service | notification, realtime            |
| `debt.settled`           | billing-service | analytics, realtime               |
| `photo.uploaded`         | media-service   | media (tagging worker)            |
| `photo.tagged`           | media-service   | notification, realtime            |

---

## gRPC

Synchronous service-to-service calls use gRPC. Proto definitions live in `proto/`.

```bash
# regenerate stubs after changing a .proto file
make proto
```

| Proto           | Used by                             |
| --------------- | ----------------------------------- |
| `auth.proto`    | gateway → auth-service              |
| `user.proto`    | auth, event, billing → user-service |
| `flags.proto`   | any service → feature-flag-service  |

---

## Database migrations

Each service owns its own database and migrations independently.

All services use **goose** for migrations and **sqlc** for typed queries. Files live in `services/<name>/db/migrations/` named `0001_description.sql`, `0002_description.sql`, etc.

```bash
POSTGRES_DSN=... make venue-migrate-up   # apply pending migrations
make venue-sqlc                          # regenerate Go code from db/queries/*.sql
```

Repository tests run against real Postgres. Each test creates its own throwaway database, migrates it and drops it afterwards (`shared/postgres/postgrestest`). They are skipped unless `TEST_POSTGRES_DSN` points at a server whose user can create databases, such as any compose container:

```bash
docker compose up -d postgres-user
TEST_POSTGRES_DSN="postgres://user_user:user_pass@localhost:5434/users?sslmode=disable" make user-test
```

---

## Branching

```
main   ← always deployable, protected — never commit directly
  └── dev ← integration branch — all PRs target here
        └── feat/service-name/what-it-does
```

```bash
# create a branch
git checkout dev && git pull origin dev
git checkout -b feat/billing-service/debt-simplification

# before opening a PR — rebase onto latest dev
git fetch origin
git rebase origin/dev
git push origin feat/billing-service/debt-simplification

# open PR → dev  (not main)
```

**CI:** every PR runs `.github/workflows/ci.yml`: a `lint` job (gofmt, go vet, golangci-lint) and a `test` job (`go test -race` against a Postgres 18 service container, coverage in the run summary) per module, plus a Docker build per service. Merging into `main` requires the `ci-ok` check to pass.

**Commit format:** `type(scope): description`

```
feat(billing): implement min-cash-flow debt graph
fix(auth): reject expired OTP codes
chore(infra): add kafka-ui to docker compose
test(event): add state machine transition unit tests
refactor(venue): extract places client into adapter
```

Types: `feat` `fix` `chore` `refactor` `test` `docs` `perf` `ci`

Scope: the service name — `billing` `auth` `gateway` `user` `venue` `infra` `proto`

---

## Adding a new service

1. Create the directory under `services/`
2. `go mod init github.com/SamPariat/roundup/services/<name>` and copy the `services/venue-service` layout
3. Add the module to `go.work`
4. Add a `Dockerfile`
5. Add `.env.beta.example` and `.env.prod.example` with all required variables documented, and add the service folder to `SERVICE_DIRS` in the `Makefile`
6. Add the service's database to `docker-compose.yml`
7. Add `<name>-run` / `<name>-test` / migration targets to the `Makefile`
8. Add the module to the `lint` and `test` matrices (and the service to the `docker` matrix) in `.github/workflows/ci.yml`
9. Update the services table in this README
10. Add ownership to `.github/CODEOWNERS`

---

## Team

| Person   | Services                                                              |
| -------- | --------------------------------------------------------------------- |
| Person A | gateway, auth-service, admin-service                                  |
| Person B | user-service, event-service, billing-service                          |
| Person C | venue-service, notification-service, media-service                    |
| Person D | realtime-service, analytics-service, feature-flag-service             |

---

## Further reading

- [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) — system design, service map, communication patterns
- [docs/CONTRIBUTING.md](docs/CONTRIBUTING.md) — detailed contribution guide
- [docs/RUNBOOK.md](docs/RUNBOOK.md) — ops procedures (restart service, replay webhook)
- [docs/adr/](docs/adr/) — architecture decision records
