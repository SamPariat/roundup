# Roundup

Weekend squad outing planner. Plan where to go, track who's coming, split the bill, settle up, remember the night.

---

## What's this?

Roundup is a mobile app (Flutter) backed by a microservices architecture. It handles the full lifecycle of a group outing — from venue voting and RSVP tracking to live bill splitting and UPI payments.

This repository is a monorepo containing all backend services, shared proto definitions, and infrastructure config.

---

## Tech stack

| Layer                                                | Technology                     |
| ---------------------------------------------------- | ------------------------------ |
| Mobile                                               | Flutter                        |
| API Gateway                                          | Go (Fiber)                     |
| Auth                                                 | NestJS + TypeScript + Firebase |
| User, Event, Billing, Analytics                      | Java + Spring Boot 3           |
| Payment, Venue, Notification, Media, Realtime, Flags | Go                             |
| AI                                                   | Python + FastAPI + LangChain   |
| Message broker                                       | Apache Kafka                   |
| Cache                                                | Redis                          |
| Database                                             | PostgreSQL (one per service)   |
| Real-time sync                                       | Firestore (live bill tab only) |
| Media storage                                        | Cloudinary                     |
| Push                                                 | Firebase Cloud Messaging       |
| Payments                                             | Razorpay                       |

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
| Go             | 1.22+    | https://go.dev/dl                            |
| Java           | 21 (LTS) | https://adoptium.net                         |
| Node.js        | 20 LTS   | https://nodejs.org                           |
| pnpm           | 9+       | `npm i -g pnpm`                              |
| Python         | 3.11+    | https://python.org                           |
| Flutter        | stable   | https://flutter.dev/docs/get-started/install |
| nest-cli       | latest   | `npm i -g @nestjs/cli`                       |
| buf            | latest   | https://buf.build/docs/installation          |
| golangci-lint  | latest   | https://golangci-lint.run/usage/install      |

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

This copies `.env.example` to `.env` for every service. The `.env` files are gitignored — they live only on your machine.

### 3. Fill in environment variables

Open each `services/*/.env` file and fill in the required values.
See [Environment variables](#environment-variables) below for what you need.

### 4. Start infrastructure only

```bash
make infra
```

Starts Postgres (one per service), Redis, Kafka, and Zookeeper. No application services yet.

```bash
make ps   # all infra containers should show as healthy
```

### 5. Run migrations

```bash
make migrate-all
```

### 6. Start everything

```bash
make up
```

### 7. Verify

```bash
curl http://localhost:8080/health
# {"status":"ok"}
```

---

## Environment variables

All third-party services have free sandbox / test modes. No real money changes hands during development.

### Firebase

1. Go to [console.firebase.google.com](https://console.firebase.google.com) → create a project
2. Project Settings → Service Accounts → Generate new private key → download JSON
3. Authentication → Sign-in method → enable **Phone**
4. Authentication → Phone → **Test phone numbers** → add your dev numbers here (no SMS cost, no OTP sent)

```env
# services/auth-service/.env
FIREBASE_PROJECT_ID=your-project-id
FIREBASE_CLIENT_EMAIL=firebase-adminsdk-xxx@your-project.iam.gserviceaccount.com
FIREBASE_PRIVATE_KEY="-----BEGIN PRIVATE KEY-----\n...\n-----END PRIVATE KEY-----\n"
```

> The `\n` in `FIREBASE_PRIVATE_KEY` must be literal backslash-n in the `.env` file, not actual newlines. The service handles the replacement.

### Google Maps / Places

1. [console.cloud.google.com](https://console.cloud.google.com) → enable **Places API (New)**
2. Create an API key → restrict it to Places API only
3. Billing → Budgets & Alerts → set an alert at $0.01 so you're notified before any charge
4. Free tier: 10,000 calls/month per SKU — more than enough for development

```env
# services/venue-service/.env
GOOGLE_MAPS_API_KEY=AIza...
```

### Razorpay (payments)

1. Go to [razorpay.com](https://razorpay.com) → sign up — no KYC needed to access test mode
2. Dashboard → toggle **Test Mode** on (top-right switch)
3. Settings → API Keys → **Generate Test Key**
4. Copy the Key ID and Key Secret

```env
# services/payment-service/.env
RAZORPAY_KEY_ID=rzp_test_...
RAZORPAY_KEY_SECRET=...
RAZORPAY_WEBHOOK_SECRET=...
```

**Test credentials for local development**

| Method       | Value                   | Behaviour       |
| ------------ | ----------------------- | --------------- |
| UPI ID       | `success@razorpay`      | Always succeeds |
| UPI ID       | `failure@razorpay`      | Always fails    |
| UPI ID       | `pending@razorpay`      | Stays pending   |
| Card number  | `4111 1111 1111 1111`   | Always succeeds |
| Card number  | `4242 4242 4242 4242`   | Always succeeds |
| CVV / expiry | any valid-looking value | Works in test   |

**Webhook testing locally** — use [ngrok](https://ngrok.com) to expose your local payment service:

```bash
ngrok http 3005
# copy the https URL → Razorpay Dashboard → Webhooks → Add new webhook
# e.g. https://abc123.ngrok.io/webhooks/razorpay
```

Subscribe to these events in the Razorpay dashboard:

- `payment.captured`
- `payment.failed`
- `refund.created`

> Live mode requires KYC (Aadhaar + PAN + business details). Stay in test mode until you are ready to go live.

### Cloudinary (media)

1. [cloudinary.com](https://cloudinary.com) → sign up — no credit card needed
2. Dashboard → copy Cloud Name, API Key, API Secret

```env
# services/media-service/.env
CLOUDINARY_CLOUD_NAME=...
CLOUDINARY_API_KEY=...
CLOUDINARY_API_SECRET=...
```

Free plan: 25 credits/month. Sufficient for development and small-scale testing.

---

## Daily workflow

```bash
# start everything
make up

# tail logs for a specific service
make logs s=auth-service

# restart after a code change
make restart s=billing-service

# run tests for a service
make test s=billing-service

# open a shell inside a running container
make shell s=redis

# open psql for a service's database
make psql s=billing

# stop everything
make down
```

---

## Running a service locally (outside Docker)

Faster for active development — run the service you are working on natively, with only infra in Docker.

```bash
make infra   # start Postgres, Redis, Kafka only
```

**NestJS (auth-service, admin-service)**

```bash
cd services/auth-service
pnpm install
pnpm start:dev
```

**Go (gateway, payment-service, venue-service, etc.)**

```bash
cd services/gateway
go run cmd/main.go
```

**Java (user-service, event-service, billing-service, analytics-service)**

```bash
cd services/billing-service
./gradlew bootRun
```

**Python (ai-service)**

```bash
cd services/ai-service
python -m venv .venv
source .venv/bin/activate        # Windows: .venv\Scripts\activate
pip install -r requirements.txt
uvicorn app.main:app --reload
```

---

## Services

| Service              | Language | Port | Description                             |
| -------------------- | -------- | ---- | --------------------------------------- |
| gateway              | Go       | 8080 | Routes all traffic, rate limiting, auth |
| auth-service         | NestJS   | 3001 | Firebase token verification, JWT        |
| user-service         | Java     | 3002 | Profiles, squads, device tokens         |
| event-service        | Java     | 3003 | Outing lifecycle, RSVP, voting          |
| billing-service      | Java     | 3004 | Bill tab, split strategies, debt graph  |
| payment-service      | Go       | 3005 | Razorpay integration, webhooks          |
| venue-service        | Go       | 3006 | Google Places facade, Redis cache       |
| notification-service | Go       | 3007 | FCM push, WhatsApp                      |
| media-service        | Go       | 3008 | Photos, face detection, QR download     |
| realtime-service     | Go       | 3009 | WebSocket hub, presence                 |
| ai-service           | Python   | 3010 | LangChain, receipt OCR, recommendations |
| reservation-service  | Go       | 3011 | Table booking via Dineout               |
| analytics-service    | Java     | 3012 | Squad stats, leaderboards, wrapped      |
| feature-flag-service | Go       | 3013 | Flag evaluation, rollout control        |
| admin-service        | NestJS   | 3014 | Internal ops dashboard                  |

All services share Redis on `6379` and Kafka on `9092`.

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

Inter-service events flow through Kafka. Topic name constants are in `shared/constants/kafka-topics.*`.

| Topic                    | Publisher       | Consumers                         |
| ------------------------ | --------------- | --------------------------------- |
| `rsvp.updated`           | event-service   | notification, realtime, analytics |
| `event.created`          | event-service   | analytics                         |
| `event.closed`           | event-service   | billing, ai, analytics            |
| `tab.opened`             | billing-service | realtime                          |
| `tab.closed`             | billing-service | notification, ai                  |
| `debt.created`           | billing-service | notification, realtime            |
| `debt.settled`           | billing-service | analytics, realtime               |
| `payment.initiated`      | payment-service | analytics                         |
| `payment.confirmed`      | payment-service | billing, notification, realtime   |
| `payment.failed`         | payment-service | billing, notification             |
| `photo.uploaded`         | media-service   | media (tagging worker)            |
| `photo.tagged`           | media-service   | notification, realtime            |
| `outing.recap.requested` | ai-service      | notification                      |
| `notification.send`      | ai-service      | notification-service              |

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
| `billing.proto` | billing → payment-service           |
| `flags.proto`   | any service → feature-flag-service  |

---

## Database migrations

Each service owns its own database and migrations independently.

**Java services** — Flyway runs automatically on startup. Files live in `src/main/resources/db/migration/` named `V1__description.sql`, `V2__description.sql`, etc.

**Go services** — run manually:

```bash
make migrate s=payment-service
```

**NestJS services** — Drizzle Kit:

```bash
cd services/auth-service
pnpm drizzle-kit generate   # generate migration from schema change
pnpm drizzle-kit migrate    # apply pending migrations
pnpm drizzle-kit studio     # open DB browser UI
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

**Commit format:** `type(scope): description`

```
feat(billing): implement min-cash-flow debt graph
fix(auth): handle expired firebase token gracefully
chore(infra): add kafka-ui to docker compose
test(event): add state machine transition unit tests
refactor(payment): extract razorpay client into adapter
```

Types: `feat` `fix` `chore` `refactor` `test` `docs` `perf` `ci`

Scope: the service name — `billing` `auth` `gateway` `payment` `infra` `proto`

---

## Adding a new service

1. Create the directory under `services/`
2. Initialise the project (`nest new`, `go mod init`, `spring init`, `fastapi`, etc.)
3. Add a `Dockerfile`
4. Add `.env.example` with all required variables documented
5. Add the service to `docker-compose.yml`
6. Add a shortcut to the `Makefile` if useful
7. Update the services table in this README
8. Add ownership to `.github/CODEOWNERS`

---

## Team

| Person   | Services                                                              |
| -------- | --------------------------------------------------------------------- |
| Person A | gateway, auth-service, admin-service                                  |
| Person B | user-service, event-service, billing-service                          |
| Person C | payment-service, venue-service, notification-service, media-service   |
| Person D | realtime-service, ai-service, analytics-service, feature-flag-service |

---

## Further reading

- [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) — system design, service map, communication patterns
- [docs/CONTRIBUTING.md](docs/CONTRIBUTING.md) — detailed contribution guide
- [docs/RUNBOOK.md](docs/RUNBOOK.md) — ops procedures (restart service, replay webhook, manual refund)
- [docs/adr/](docs/adr/) — architecture decision records
