<p align="center">
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset="assets/brand/banner-dark.svg">
    <img alt="Roundup: round up the squad, pick the place, split the bill" src="assets/brand/banner-light.svg" width="100%">
  </picture>
</p>

<p align="center">
  <a href="https://github.com/SamPariat/roundup/actions/workflows/ci.yml"><img alt="CI" src="https://img.shields.io/github/actions/workflow/status/SamPariat/roundup/ci.yml?branch=main&style=for-the-badge&label=CI&logo=githubactions&logoColor=cdd6f4&labelColor=1e1e2e&color=a6e3a1"></a>
  <img alt="Go 1.27" src="https://img.shields.io/badge/Go-1.27-cba6f7?style=for-the-badge&logo=go&logoColor=cdd6f4&labelColor=1e1e2e">
  <a href="shared/api/openapi.yaml"><img alt="OpenAPI /api/v1" src="https://img.shields.io/badge/OpenAPI-%2Fapi%2Fv1-fab387?style=for-the-badge&logo=openapiinitiative&logoColor=cdd6f4&labelColor=1e1e2e"></a>
  <img alt="PostgreSQL 18" src="https://img.shields.io/badge/PostgreSQL-18-b4befe?style=for-the-badge&logo=postgresql&logoColor=cdd6f4&labelColor=1e1e2e">
  <img alt="Client: Expo, planned" src="https://img.shields.io/badge/client-Expo%20(planned)-f5c2e7?style=for-the-badge&logo=expo&logoColor=cdd6f4&labelColor=1e1e2e">
  <img alt="Made with love" src="https://img.shields.io/badge/made%20with-%F0%9F%92%9C-cba6f7?style=for-the-badge&labelColor=1e1e2e">
</p>

<p align="center">
  <b>The weekend planner for your squad</b>: pick where to go, see who's coming, split the bill, settle up and remember the night.
</p>

<p align="center">
  <a href="#-getting-started">🚀 Getting started</a> ·
  <a href="#-services">🧭 Services</a> ·
  <a href="#-api-docs">📖 API docs</a> ·
  <a href="ROUNDUP-TASKS.md">🗺️ Roadmap</a>
</p>

<p align="center"><img src="assets/brand/divider.svg" width="100%" alt=""></p>

## ✨ What Roundup does

<table>
  <tr>
    <td width="25%" valign="top"><h3>🗳️ Plan &amp; vote</h3>Suggest bars and restaurants from Google Places, then vote in the open. Most votes wins; a tie goes to the first suggestion.</td>
    <td width="25%" valign="top"><h3>🙋 RSVP</h3>Going, maybe or not going, with a live headcount and the faces of who's in.</td>
    <td width="25%" valign="top"><h3>🧾 Split the bill</h3>Scan the receipt, split it evenly and net every bill of the night into one debt per pair. <i>Coming next.</i></td>
    <td width="25%" valign="top"><h3>📸 Remember the night</h3>Every night closes into the squad's history: where you went and how often. Photos come later.</td>
  </tr>
</table>

> **Language policy:** the backend is Go only. The client will be one React Native + Expo app (TypeScript) for web, Android and iOS. No other languages.

<p align="center"><img src="assets/brand/divider.svg" width="100%" alt=""></p>

## 🏗️ Architecture

Go microservices behind one gateway, each with its own Postgres. Every route is described in a single OpenAPI document.

```mermaid
flowchart LR
    app(["📱 Expo app<br/>(planned)"]):::planned --> gw
    gw["🚪 gateway<br/>:8080"]:::gateway

    gw --> auth["🔐 auth-service<br/>:3001"]:::live
    gw --> user["👥 user-service<br/>:3002"]:::live
    gw --> event["🗳️ event-service<br/>:3003"]:::live
    gw --> venue["📍 venue-service<br/>:3006"]:::live

    auth --> authdb[("auth db")]:::db
    user --> userdb[("users db")]:::db
    event --> eventdb[("events db")]:::db
    venue --> venuedb[("venues db")]:::db
    venue --> redis[("Redis cache")]:::db
    venue --> places{{"Google Places<br/>API (New)"}}:::external

    auth -. "create user" .-> user
    event -. "members" .-> user
    event -. "place details, visits" .-> venue
    venue -. "members" .-> user

    classDef gateway fill:#cba6f7,stroke:#8839ef,color:#1e1e2e,font-weight:bold
    classDef live fill:#fab387,stroke:#fe640b,color:#1e1e2e
    classDef db fill:#b4befe,stroke:#7287fd,color:#1e1e2e
    classDef external fill:#f5c2e7,stroke:#ea76cb,color:#1e1e2e
    classDef planned fill:#313244,stroke:#a6adc8,color:#cdd6f4,stroke-dasharray:5 4
```

Services call each other directly over HTTP for reads they need right away (dotted lines); scheduled work such as voting deadlines will live in a scheduler-service that calls each service's internal endpoints.

### 🗓️ An event's life

```mermaid
%%{init: {'themeVariables': {'noteBkgColor': '#f5c2e7', 'noteTextColor': '#1e1e2e', 'noteBorderColor': '#ea76cb'}}}%%
stateDiagram-v2
    direction LR
    classDef draft fill:#b4befe,stroke:#7287fd,color:#1e1e2e
    classDef voting fill:#cba6f7,stroke:#8839ef,color:#1e1e2e,font-weight:bold
    classDef confirmed fill:#fab387,stroke:#fe640b,color:#1e1e2e,font-weight:bold
    classDef active fill:#a6e3a1,stroke:#40a02b,color:#1e1e2e,font-weight:bold
    classDef closed fill:#45475a,stroke:#a6adc8,color:#cdd6f4
    class DRAFT draft
    class VOTING voting
    class CONFIRMED confirmed
    class ACTIVE active
    class CLOSED closed

    [*] --> DRAFT: create
    DRAFT --> VOTING: start voting
    VOTING --> CONFIRMED: close voting<br/>or the deadline passes
    CONFIRMED --> ACTIVE: start the night
    ACTIVE --> CLOSED: close<br/>or 12h after the start
    CLOSED --> [*]

    note right of DRAFT: RSVP and suggest places
    note right of VOTING: one open vote each
    note right of CONFIRMED: the winner is set
    note right of CLOSED: added to the squad's history
```

<p align="center"><img src="assets/brand/divider.svg" width="100%" alt=""></p>

## 🧱 Tech stack

| | In use | Planned |
| --- | --- | --- |
| **Services** | Go 1.27, Fiber v3 | |
| **Data** | PostgreSQL 18 (one per service), sqlc + pgx, goose migrations | |
| **Cache** | Redis (venue search and details) | |
| **Auth** | Custom phone OTP + Ed25519 JWTs | an SMS provider for production |
| **Permissions** | Casbin, from a `policy.csv` per service | |
| **Places** | Google Places API (New) | |
| **Client** | | React Native + Expo (web first, then Android and iOS) |
| **Live bill tab** | | Firestore |
| **Media** | | Cloudinary |
| **Push** | | Firebase Cloud Messaging |

## 📁 Repository structure

```
roundup/
├── services/           # one Go module per service (see the services table)
├── shared/             # Go packages every service uses: api spec, auth, pagination, clients…
├── assets/brand/       # logo, banner and mark files
├── scripts/            # dev helper scripts
├── .github/workflows/  # CI
├── docker-compose.yml  # databases, Redis and the services for local development
├── go.work             # Go workspace listing every module
└── Makefile            # convenience commands
```

<p align="center"><img src="assets/brand/divider.svg" width="100%" alt=""></p>

## 🧰 Prerequisites

| Tool           | Version | Install                                  |
| -------------- | ------- | ---------------------------------------- |
| Docker Desktop | latest  | https://docs.docker.com/get-docker       |
| Git            | 2.x+    | https://git-scm.com                      |
| Make           | any     | pre-installed on macOS / Linux           |
| Go             | 1.27+   | https://go.dev/dl                        |
| goose          | latest  | `make install`                           |
| sqlc           | latest  | `make install`                           |
| golangci-lint  | latest  | https://golangci-lint.run/usage/install  |

## 🚀 Getting started

### 1. Clone

```bash
git clone git@github.com:SamPariat/roundup.git
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

Open each `services/*/.env.beta` file and fill in the required values. See [Environment variables](#-environment-variables) below for what you need.

### 4. Start the databases

```bash
docker compose up -d
docker compose ps   # every container should show as healthy
```

Starts one Postgres per service (auth, user, event, venue) and Redis. The Go services are not started by this.

### 5. Run migrations

Each migrate target takes the database's DSN:

```bash
make auth-migrate-up  POSTGRES_DSN="postgres://auth_user:auth_pass@localhost:5433/auth?sslmode=disable"
make user-migrate-up  POSTGRES_DSN="postgres://user_user:user_pass@localhost:5434/users?sslmode=disable"
make event-migrate-up POSTGRES_DSN="postgres://event_user:event_pass@localhost:5435/events?sslmode=disable"
make venue-migrate-up POSTGRES_DSN="postgres://venue_user:venue_pass@localhost:5438/venues?sslmode=disable"
```

### 6. Start the services

Either in Docker (all five, logs in Docker Desktop):

```bash
make services-up
```

Or natively, for active development; see [Running a service locally](#-running-a-service-locally-outside-docker).

### 7. Verify

```bash
curl http://localhost:8080/ready
# {"success":true,"data":{"status":"ready"}}
```

## 🔐 Environment variables

All third-party services have free sandbox or test modes. No real money changes hands during development.

<details>
<summary><b>🔑 Auth: custom phone OTP + JWT</b></summary>

<br>

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

</details>

<details>
<summary><b>📍 Google Places API (New)</b></summary>

<br>

1. [console.cloud.google.com](https://console.cloud.google.com) → APIs & Services → enable **Places API (New)**
2. Create an API key → API restrictions → allow **Places API (New)** only
3. Billing → Budgets & Alerts → set an alert at $0.01 so you're notified before any charge
4. Free tier: 10,000 calls/month per SKU, more than enough for development

```env
# services/venue-service/.env.beta
GOOGLE_MAPS_API_KEY=AIza...
```

Restart venue-service after changing the key; a running container keeps the old one.

</details>

<details>
<summary><b>🖼️ Cloudinary (media, planned)</b></summary>

<br>

For media-service, once it exists.

1. [cloudinary.com](https://cloudinary.com) → sign up, no credit card needed
2. Dashboard → copy Cloud Name, API Key, API Secret

```env
# services/media-service/.env.beta
CLOUDINARY_CLOUD_NAME=...
CLOUDINARY_API_KEY=...
CLOUDINARY_API_SECRET=...
```

Free plan: 25 credits/month, enough for development and small-scale testing.

</details>

<p align="center"><img src="assets/brand/divider.svg" width="100%" alt=""></p>

## 🛠️ Daily workflow

```bash
# databases only (always needed)
docker compose up -d

# all five services in Docker, rebuilt from the current code
make services-up

# follow their logs (also in Docker Desktop: Containers → roundup-monorepo → a service → Logs)
make services-logs

# stop the services, keep the databases running
make services-down

# run tests for a service
make auth-test        # also user-test, event-test, venue-test, gateway-test

# stop everything (data is kept in volumes)
docker compose --profile services down
```

Services in Docker read the same `services/*/.env.<APP_ENV>` files as native runs (`make services-up APP_ENV=prod` for prod; beta by default); `docker-compose.yml` only swaps `localhost` addresses for container names. Docker and native runs use the same ports, so run `make services-down` before `make *-run`.

### 💻 Running a service locally (outside Docker)

<details>
<summary>Faster for active development; click for how.</summary>

<br>

Faster for active development: run the service you're working on natively, with only the databases in Docker.

```bash
docker compose up -d   # databases only
```

Every service is Go and runs the same way:

```bash
cd services/venue-service
APP_ENV=beta go run ./cmd
```

Or use the Makefile shortcuts: `make auth-run`, `make user-run`, `make event-run`, `make venue-run`, `make gateway-run`.

</details>

## 🧭 Services

| | Service | Port | What it does |
| --- | --- | --- | --- |
| 🟢 | gateway | 8080 | Routes `/api/v1`, verifies access tokens, rate-limits sign-in |
| 🟢 | auth-service | 3001 | Phone OTP sign-in, Ed25519 access and refresh tokens |
| 🟢 | user-service | 3002 | Profiles, squads and members, device tokens |
| 🟡 | event-service | 3003 | Events, RSVPs, place suggestions, open votes ([#36](https://github.com/SamPariat/roundup/pull/36)) |
| 🟢 | venue-service | 3006 | Places API (New) search and details with a Redis cache; favorites and squad history |
| ⚪ | scheduler-service | | Runs scheduled jobs, such as closing voting at its deadline |
| ⚪ | billing-service | | Receipt scanning, bill splitting, debts |
| ⚪ | notification-service | | Push notifications |
| ⚪ | media-service | | Photos |
| ⚪ | realtime-service | | WebSocket updates |

🟢 live · 🟡 in review · ⚪ planned

Every HTTP route is versioned under `/api/v1` (e.g. `/api/v1/users/me`). Postgres runs one container per service (auth 5433, user 5434, event 5435, venue 5438) and Redis is on 6379.

### 📖 API docs

Every route is described in one OpenAPI 3 document, [`shared/api/openapi.yaml`](shared/api/openapi.yaml). With the services running (`make services-up`), Swagger UI is at **http://localhost:8080/docs**: use "Authorize" with an `accessToken` from `/auth/otp/verify` to try the protected routes. The gateway serves it only when `DOCS_ENABLED=true`, as in `.env.beta`; prod leaves it off.

Each service's `TestRoutesAreDocumented` compares its router with the spec, so adding, removing or renaming a route fails CI until `openapi.yaml` is updated to match.

### 📄 Pagination

Every list whose size has no limit is paged with a cursor:

- `GET /squads`
- `GET /squads/:id/favorites`
- `GET /squads/:id/history`
- `GET /squads/:id/events`
- `GET /venues` (search)

```
GET /api/v1/squads?limit=25                 → { "squads": [...25], "nextCursor": "eyJj..." }
GET /api/v1/squads?limit=25&cursor=eyJj...  → { "squads": [...7],  "nextCursor": null }
```

- **`limit`:** 1–100, default 25. Outside that range is a 400. Search has no `limit`: Google fixes its page at 20, up to 60 results per search.
- **`cursor`:** opaque. Send the previous page's `nextCursor` back exactly as received, and stop when it is `null`. A cursor the API did not issue is a 400.
- **Why a cursor, not page numbers:** pages continue from the last row's sort key (keyset paging), not from an offset. Rows added while someone scrolls don't shift later pages, so nothing is skipped or shown twice.
- **Bounded lists are not paged:** squad members (at most 50) and contact lookup (at most 500 numbers per request).
- **For service code:** the helper is [`shared/pagination`](shared/pagination/pagination.go). Repositories fetch `limit + 1` rows to know whether another page exists.

<p align="center"><img src="assets/brand/divider.svg" width="100%" alt=""></p>

## 🗄️ Database migrations

Each service owns its own database and migrations.

All services use **goose** for migrations and **sqlc** for typed queries. Files live in `services/<name>/db/migrations/`, named `0001_description.sql`, `0002_description.sql`, etc.

```bash
make venue-migrate-up POSTGRES_DSN=...   # apply pending migrations
make venue-sqlc                          # regenerate Go code from db/queries/*.sql
```

Repository tests run against real Postgres. Each test creates its own throwaway database, migrates it and drops it afterwards (`shared/postgres/postgrestest`). They are skipped unless `TEST_POSTGRES_DSN` points at a server whose user can create databases, such as any compose container:

```bash
docker compose up -d postgres-user
TEST_POSTGRES_DSN="postgres://user_user:user_pass@localhost:5434/users?sslmode=disable" make user-test
```

## 🌿 Branching

```
main   ← always deployable, protected: PRs only, and the ci-ok check must pass
  └── feat/service-name/what-it-does
```

```bash
# create a branch
git checkout main && git pull origin main
git checkout -b feat/billing-service/debt-simplification

# before opening a PR, rebase onto the latest main
git fetch origin
git rebase origin/main
git push origin feat/billing-service/debt-simplification

# open a PR → main
```

**CI:** every PR runs `.github/workflows/ci.yml`: a `lint` job (gofmt, go vet, golangci-lint with the root `.golangci.yml`, which also enforces gofumpt formatting and blank-line spacing) and a `test` job (`go test -race` against a Postgres 18 service container, coverage in the run summary) per module, plus a Docker build per service. Merging into `main` requires the `ci-ok` check to pass.

**Commit format:** `type(scope): description`

```
feat(billing): implement min-cash-flow debt graph
fix(auth): reject expired OTP codes
chore(infra): add postgres-event to docker compose
test(event): add state machine transition unit tests
refactor(venue): extract places client into adapter
```

Types: `feat` `fix` `chore` `refactor` `test` `docs` `perf` `ci`

Scope: the service name (`auth` `gateway` `user` `event` `venue` `billing`), or `shared` / `infra`.

## ➕ Adding a new service

1. Create the directory under `services/`
2. `go mod init github.com/SamPariat/roundup/services/<name>` and copy the `services/venue-service` layout
3. Add the module to `go.work`
4. Add a `Dockerfile`
5. Add `.env.beta.example` and `.env.prod.example` with all required variables documented, and add the service folder to `SERVICE_DIRS` in the `Makefile`
6. Add the service's database to `docker-compose.yml`
7. Add `<name>-run` / `<name>-test` / migration targets to the `Makefile`
8. Add the module to the `lint` and `test` matrices (and the service to the `docker` matrix) in `.github/workflows/ci.yml`
9. Update the services table in this README

## 👥 Team

| Person   | Services                                                  |
| -------- | --------------------------------------------------------- |
| Person A | gateway, auth-service, admin-service                      |
| Person B | user-service, event-service, billing-service              |
| Person C | venue-service, notification-service, media-service        |
| Person D | realtime-service, analytics-service, feature-flag-service |

## 📚 Further reading

- [SCOPE.md](SCOPE.md): what Roundup does and doesn't do
- [ROUNDUP-TASKS.md](ROUNDUP-TASKS.md): the roadmap, service by service
- [ROUNDUP-DIRECTORY-STRUCTURE.md](ROUNDUP-DIRECTORY-STRUCTURE.md): the planned layout of every service
- Each service's `ARCHITECTURE.md`, e.g. [venue-service](services/venue-service/ARCHITECTURE.md) and [event-service](services/event-service/ARCHITECTURE.md)
- [assets/brand](assets/brand/README.md): the logo files and how to use them

<p align="center"><img src="assets/brand/divider.svg" width="100%" alt=""></p>

<p align="center">
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset="assets/brand/roundup-mark.svg">
    <img alt="" src="assets/brand/roundup-mark-latte.svg" width="40">
  </picture>
  <br>
  <sub>Built with Go 🐹, Postgres 🐘 and good company.</sub>
</p>
