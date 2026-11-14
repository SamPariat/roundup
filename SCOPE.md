# Roundup — Product & Technical Scope

> What Roundup does, how the backend is put together, and in what order it gets built. For the task-level checklist see `ROUNDUP-TASKS.md`; for how a built service works inside, see its `ARCHITECTURE.md`.

> **Language policy:** Backend is Go only. The client is one React Native + Expo app (TypeScript) for web, Android and iOS, built side by side in the separate `roundup-ui` repo.

---

## Table of contents

- [Product overview](#product-overview)
- [Feature scope](#feature-scope)
- [Architecture](#architecture)
- [Services](#services)
- [Communication](#communication)
- [Design patterns](#design-patterns)
- [Database schemas](#database-schemas)
- [SDKs & third-party](#sdks--third-party)
- [Deployment](#deployment)
- [Build order](#build-order)

---

## Product overview

Roundup is a weekend squad outing planner. It handles the full arc of a group night out — from deciding where to go, to splitting the bill, settling up, and remembering the night.

**Core loop**

```
Plan → Vote → RSVP → Go out → Scan the bill → Settle → Remember
```

**Target users** — friend groups of 5–15 people who go out regularly and currently coordinate over WhatsApp, split bills on Splitwise, and lose photos across everyone's camera rolls.

---

## Feature scope

### Planning & discovery

| Feature             | Description                                                                     | Status  |
| ------------------- | ------------------------------------------------------------------------------- | ------- |
| Venue voting        | Suggest places, open voting with a deadline. Most votes wins; a tie goes to the first suggestion. | Built   |
| Place discovery     | Google Places search filtered by vibe — rooftop, cocktail bar, pub, fine dining | Built   |
| Saved favourites    | Squad's go-to spots saved for quick access                                      | Built   |
| Venue history       | Every venue the squad has visited, with average spend                           | Built   |
| RSVP                | Going / maybe / not going, per event                                            | Built   |
| Table reservation   | Direct booking via Dineout / EazyDiner                                          | Planned |
| Deals & offers      | Venue-specific discounts surfaced during planning                               | Planned |
| Event templates     | Reuse a past outing's setup (same squad, same venue shortlist)                  | Planned |
| Budget cap          | Set max spend per person before the outing                                      | Planned |
| Recurring outings   | Every Friday auto-created with the same squad (a scheduler-service job)         | Planned |
| Calendar sync       | Export outing to Google / Apple Calendar                                        | Planned |
| Invite links        | Shareable link — join without being in the squad                                | Deferred |
| Guest mode          | RSVP and settle bill without creating an account                                | Deferred |
| Cross-squad outings | Merge two squads for one event                                                  | Planned |

### Billing & debts

The bill is the reason to open an event: `ACTIVE` means the bill is open. Money never moves through Roundup.

| Feature             | Description                                                                 | Status  |
| ------------------- | --------------------------------------------------------------------------- | ------- |
| Receipt scanning    | Photograph the bill; OCR proposes the items and total, the payer confirms   | Next    |
| One payer per bill  | Whoever paid adds the bill and picks who shares it                          | Next    |
| Multi-bill nights   | Bar tab + dinner on the same night, possibly paid by different people       | Next    |
| Equal split         | Split evenly in paise; leftover paise go to members in a fixed order        | Next    |
| Netting             | All of a night's bills net into one debt per pair of people                 | Next    |
| Mark as paid        | Settle outside the app, then the person owed marks it paid                  | Next    |
| Running balance     | Total owed / owing across all squads                                        | Next    |
| By-item and custom % split | Claim line items, or split by percentage                             | Later   |
| Tax & tip split     | Proportional or equal                                                       | Later   |
| Dispute resolution  | Flag an incorrect line item                                                 | Later   |

### Social & memory

| Feature            | Description                                                      | Status   |
| ------------------ | ---------------------------------------------------------------- | -------- |
| Photo album        | One album per outing, collected from all squad members           | Planned  |
| QR download        | Scan a QR to download a ZIP of the outing's photos               | Planned  |
| Squad stats        | Most visited spot, biggest spender, most frequent RSVPer         | Planned  |
| Leaderboard        | Rankings per squad — fun data to roast each other with           | Planned  |
| Yearly wrapped     | Spotify-style annual recap                                       | Planned  |
| Streak tracking    | Consecutive weekly outings                                       | Planned  |
| Vibes rating       | Rate the outing after it closes                                  | Planned  |
| Reactions          | Emoji reactions on photos                                        | Planned  |
| Shareable card     | Instagram-ready highlight card for the outing                    | Planned  |
| On this day        | Anniversary reminder for past outings                            | Planned  |
| Face tagging       | Detect faces and match them to squad members                     | Deferred (AI) |
| Video reel         | Auto-compilation of photos into a short reel                     | Deferred (AI) |

### Notifications

| Feature             | Description                                             |
| ------------------- | ------------------------------------------------------- |
| RSVP nudge          | Reminder if not responded before deadline               |
| Vote deadline alert | Notification 1 hour before voting closes                |
| Debt reminder       | Gentle nudge 48 hours after a debt is created           |
| Debt settled        | Instant notification when a debt is marked paid         |
| Friday prompt       | Weekly push if no outing is planned for the weekend     |
| Quiet hours         | User-set window where no notifications are sent         |
| Channels            | FCM push (primary), WhatsApp Business API               |

Time-based notifications (nudges, reminders, the Friday prompt) are scheduler-service jobs.

### Real-time

| Feature           | Description                                          |
| ----------------- | ---------------------------------------------------- |
| Live RSVP count   | Headcount updates on the planning screen without refresh |
| Live vote tally   | See votes come in as friends cast them               |
| Live bills        | A new or edited bill shows up for everyone at the table |
| Debt settled live | Debt status clears instantly on both screens         |

### Monetisation

| Feature                | Description                                        |
| ---------------------- | -------------------------------------------------- |
| Venue partnerships     | Featured placement in discovery                    |
| Commission on bookings | Revenue on Dineout reservations                    |
| Roundup Pro            | Unlimited history, advanced analytics              |
| Referral program       | Invite friends, earn Pro credits                   |
| Deals marketplace      | Exclusive squad offers from partner venues         |

> **Later:** AI features are deferred, except receipt scanning for bills.

---

## Architecture

### Layers

```
┌─────────────────────────────────────────────────────────┐
│  React Native + Expo app — roundup-ui (web · Android · iOS) │
└────────────────────────┬────────────────────────────────┘
                         │ HTTPS  /api/v1/*   (WebSocket for realtime)
┌────────────────────────▼────────────────────────────────┐
│  API Gateway (Go)                                       │
│  Ed25519 JWT check · rate limiting · CORS · request ID  │
└──┬──────┬──────┬──────┬──────┬──────┬──────┬───────────┘
   ▼      ▼      ▼      ▼      ▼      ▼      ▼
 Auth   User  Event Billing Venue  Media  Realtime ...
   │      ▲      │      │      │
   └──────┴──────┴──────┴──────┘   direct REST calls on /internal/v1 (and a few /api/v1)

 Scheduler ──► POST /internal/v1/... on event, billing, notification, analytics

┌──────────────────────────────────────────────────────┐
│  Data                                                │
│  PostgreSQL — one database per service               │
│  Redis — venue cache only, best-effort, local        │
└──────────────────────────────────────────────────────┘

┌──────────────────────────────────────────────────────┐
│  External                                            │
│  Google Places · SMS provider · OCR provider          │
│  Cloudinary · FCM · WhatsApp Business                │
└──────────────────────────────────────────────────────┘
```

### Communication strategy

**REST over HTTP, and nothing else.** A service that needs something from another calls it directly. Routes only other services call are mounted under `/internal/v1`, which the gateway never exposes.

**Scheduled work lives in scheduler-service.** Domain services never run their own cron or ticker. They expose idempotent internal endpoints (for example `POST /internal/v1/events/confirm-due`) and scheduler-service calls them. Duplicate runs are harmless.

> Kafka and gRPC were considered and dropped. At Roundup's scale a message bus adds infrastructure and failure modes without a real benefit, and REST over an internal network is fast enough and far easier to debug.

---

## Services

| Service              | Status  | Port | DB   | Owns |
| -------------------- | ------- | ---- | ---- | ---- |
| API Gateway          | Built   | 8080 | —    | Routing, token check, rate limits, CORS |
| Auth Service         | Built   | 3001 | 5433 | Phone OTP, sessions, access + refresh tokens |
| User Service         | Built   | 3002 | 5434 | Profiles, squads, members, device tokens |
| Event Service        | Built   | 3003 | 5435 | Events, RSVPs, suggestions, votes, the event lifecycle |
| Billing Service      | Next    | 3004 | 5436 | Bills, splits, debts |
| Scheduler Service    | Next    | 3005 | —    | Calls other services' internal endpoints on a schedule |
| Venue Service        | Built   | 3006 | 5438 | Places search and details, favourites, history |
| Notification Service | Planned | 3007 | 5437 | FCM + WhatsApp sends, notification log |
| Media Service        | Planned | 3008 | 5439 | Photos, Cloudinary, QR download |
| Realtime Service     | Planned | 3009 | —    | WebSocket fan-out |
| Reservation Service  | Planned | 3011 | 5442 | Table bookings |
| Analytics Service    | Planned | 3012 | —    | Squad and user stats, wrapped |
| Feature Flag Service | Planned | 3013 | 5440 | Flags, rollout, overrides |
| Admin Service        | Planned | 3014 | —    | Admin login, user and flag management |

All public HTTP routes are versioned under `/api/v1`.

---

## Communication

### Direct calls between services

| Caller       | Callee        | Call                                          | When                              | Status  |
| ------------ | ------------- | --------------------------------------------- | --------------------------------- | ------- |
| Auth         | User          | `POST /api/v1/users`                          | Login — create or get user        | Built   |
| Venue        | User          | squad members / my squads (`shared/userclient`) | Before every squad operation    | Built   |
| Event        | User          | squad members (`shared/userclient`)           | Before every event operation      | Built   |
| Event        | Venue         | place details, `POST /internal/v1/venues/photo-urls`, record visit | Suggest, show, close the night | Built   |
| Scheduler    | Event         | `POST /internal/v1/events/confirm-due`, `…/close-due` | Voting deadline, night over | Next    |
| Event        | Billing       | `POST /internal/v1/events/:id/bills/close`    | The night closes                  | Next    |
| Billing      | Venue         | spend for the visit                           | Bills close                       | Next    |
| Event, Billing | Notification | `POST /internal/v1/notifications`            | RSVPs, votes, debts               | Planned |
| Event, Billing | Realtime    | `POST /internal/v1/realtime/publish`          | Anything a screen shows live      | Planned |
| Notification | User          | device tokens for a list of users             | Before sending push               | Planned |
| Any service  | Feature Flag  | `POST /internal/v1/flags/evaluate`            | Feature-gated code paths          | Planned |

The gateway verifies access tokens itself with auth-service's public key; it never calls auth-service per request.

---

## Design patterns

| Pattern        | Where                      | What it does |
| -------------- | -------------------------- | ------------ |
| Hexagonal (ports & adapters) | Every service | Domain at the centre; port interfaces in `domain/port.go`; Postgres, HTTP and providers are adapters. |
| Repository     | Every service              | All DB queries behind a domain interface. Use cases never touch sqlc/pgx directly. |
| Port / Adapter | Auth — SMS                 | `OTPSender` interface; console adapter in beta, an SMS provider in prod. |
| Port / Adapter | Venue — Places             | `PlaceProvider` interface; `GooglePlacesAdapter` implements it. |
| Port / Adapter | Billing — receipts         | `ReceiptScanner` interface; the OCR provider is an adapter. |
| State machine  | Event Service              | `DRAFT → VOTING → CONFIRMED → ACTIVE → CLOSED`; only the next step is allowed, re-checked inside each SQL write. |
| State machine  | Billing — debts            | `PENDING → SETTLED` (marked paid by hand). |
| State machine  | Reservation Service        | `PENDING → CONFIRMED → CANCELLED → NO_SHOW`. |
| Policy (Casbin) | User, Event               | Roles and permissions in a `policy.csv` per service, loaded into `shared/authz`. |
| Facade         | Venue Service              | Wraps Google Places + the cache. Callers never see the Places API. |
| Idempotent sweep | Event (and later Billing) | Internal endpoints that move every due row exactly once: a single guarded `UPDATE` re-checks the status under the row lock, so calling twice is harmless. |
| Middleware     | `shared/middleware`        | `RequestID`, `Logging`, `Recover`, `UserID` — Fiber middleware shared by every service. |
| Built once, injected | Every service        | DB pool, cache client, enforcer — built in `cmd/main.go` and passed in. No per-service wrappers around shared code. |

---

## Database schemas

### Built services

The goose migrations are the source of truth. Read them rather than a copy here:

- auth-service — `services/auth-service/db/migrations/` (OTP codes, sessions, refresh tokens)
- user-service — `services/user-service/db/migrations/` (users, device tokens, squads, squad members)
- event-service — `services/event-service/db/migrations/` (events, RSVPs, nominations, votes)
- venue-service — `services/venue-service/db/migrations/` (saved venues, visits, venue snapshots)

### User Service — planned addition

```sql
-- friend_connections
CREATE TABLE friend_connections (
  user_id_a   UUID NOT NULL REFERENCES users(id),
  user_id_b   UUID NOT NULL REFERENCES users(id),
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (user_id_a, user_id_b),
  CHECK (user_id_a < user_id_b)             -- prevent duplicate pairs
);
```

---

### Billing Service (next)

```sql
CREATE TYPE debt_status AS ENUM ('PENDING', 'SETTLED');

-- bills: one payer each; a night can have several
CREATE TABLE bills (
  id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  event_id     UUID NOT NULL,
  paid_by      UUID NOT NULL,
  total_paise  BIGINT NOT NULL CHECK (total_paise > 0),
  receipt_ref  TEXT,                              -- the scanned photo, if any
  created_by   UUID NOT NULL,
  created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
  deleted_at   TIMESTAMPTZ
);

-- bill_items: from the scanned receipt, editable before saving
CREATE TABLE bill_items (
  id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  bill_id      UUID NOT NULL REFERENCES bills(id),
  name         TEXT NOT NULL,
  amount_paise BIGINT NOT NULL
);

-- bill_shares: who splits this bill (equal split)
CREATE TABLE bill_shares (
  bill_id  UUID NOT NULL REFERENCES bills(id),
  user_id  UUID NOT NULL,
  PRIMARY KEY (bill_id, user_id)
);

-- debts: generated when the night closes, one per pair after netting
CREATE TABLE debts (
  id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  event_id     UUID NOT NULL,
  from_user    UUID NOT NULL,                     -- owes money
  to_user      UUID NOT NULL,                     -- is owed
  amount_paise BIGINT NOT NULL CHECK (amount_paise > 0),
  status       debt_status NOT NULL DEFAULT 'PENDING',
  created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
  settled_at   TIMESTAMPTZ,
  UNIQUE (event_id, from_user, to_user)
);

CREATE INDEX idx_bills_event_id ON bills(event_id);
CREATE INDEX idx_debts_from_user ON debts(from_user) WHERE status = 'PENDING';
CREATE INDEX idx_debts_to_user ON debts(to_user) WHERE status = 'PENDING';
```

---

### Media Service (planned)

```sql
-- photos
CREATE TABLE photos (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  event_id        UUID NOT NULL,
  uploaded_by     UUID NOT NULL,
  cloudinary_id   TEXT NOT NULL UNIQUE,         -- public_id — not the full URL
  width           INT,
  height          INT,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- download_tokens (QR-gated photo downloads)
CREATE TABLE download_tokens (
  token       TEXT PRIMARY KEY,
  user_id     UUID NOT NULL,
  event_id    UUID NOT NULL,
  expires_at  TIMESTAMPTZ NOT NULL,
  used_at     TIMESTAMPTZ
);

CREATE INDEX idx_photos_event_id ON photos(event_id);
```

---

### Notification Service (planned)

```sql
CREATE TABLE notification_log (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id     UUID NOT NULL,
  channel     TEXT NOT NULL,                   -- FCM | WHATSAPP
  type        TEXT NOT NULL,                   -- RSVP_NUDGE | DEBT_REMINDER | etc.
  status      TEXT NOT NULL DEFAULT 'SENT',    -- SENT | FAILED
  sent_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_notification_log_user_id ON notification_log(user_id);
```

---

### Feature Flag Service (planned)

```sql
CREATE TABLE feature_flags (
  key         TEXT PRIMARY KEY,
  enabled     BOOLEAN NOT NULL DEFAULT false,
  rollout     INT NOT NULL DEFAULT 100 CHECK (rollout BETWEEN 0 AND 100),
  description TEXT,
  updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- per-user value, takes precedence over the global one
CREATE TABLE user_overrides (
  flag_key    TEXT NOT NULL REFERENCES feature_flags(key),
  user_id     UUID NOT NULL,
  enabled     BOOLEAN NOT NULL,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (flag_key, user_id)
);
```

---

### Reservation Service (planned)

```sql
CREATE TYPE reservation_status AS ENUM ('PENDING', 'CONFIRMED', 'CANCELLED', 'NO_SHOW');

CREATE TABLE reservations (
  id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  event_id         UUID NOT NULL,
  venue_place_id   TEXT NOT NULL,
  provider         TEXT NOT NULL DEFAULT 'DINEOUT',
  reservation_ref  TEXT UNIQUE,                -- provider's booking reference
  status           reservation_status NOT NULL DEFAULT 'PENDING',
  party_size       INT NOT NULL,
  contact_phone    TEXT NOT NULL,
  reservation_time TIMESTAMPTZ NOT NULL,
  created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- webhook log + replay
CREATE TABLE reservation_events (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  reservation_id  UUID NOT NULL REFERENCES reservations(id),
  event_type      TEXT NOT NULL,
  raw_payload     JSONB NOT NULL,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX idx_reservations_event_id_unique ON reservations(event_id)
  WHERE status <> 'CANCELLED';
```

### Analytics Service (planned)

No event store. It reads from the other services' internal endpoints, and a scheduler-service job refreshes its numbers nightly.

---

## SDKs & third-party

| Category | SDK / API                               | Free tier                        | Used by              | Status  |
| -------- | --------------------------------------- | -------------------------------- | -------------------- | ------- |
| Maps     | Places API (New), plain HTTP            | 10,000 calls/month per SKU       | Venue Service        | Built   |
| Auth     | SMS provider (e.g. MSG91)               | Beta uses a console sender       | Auth Service         | Needed for launch |
| Receipts | OCR provider (to be chosen)             | —                                | Billing Service      | Next    |
| Push     | Firebase FCM (`firebase.google.com/go`) | Unlimited                        | Notification Service | Planned |
| Messages | WhatsApp Business API                   | Per-conversation pricing         | Notification Service | Planned |
| Media    | `cloudinary-go`                         | 25 credits/month, no card        | Media Service        | Planned |
| QR       | `skip2/go-qrcode`                       | Open source                      | Media Service        | Planned |

---

## Deployment

- GCP, `asia-south1`. Every Go service on **Cloud Run, scaling to zero** — no min-instances; cold starts are acceptable.
- **One Cloud SQL Postgres instance** (db-g1-small) with one database per service, reached over Direct VPC egress.
- **No Redis at launch.** The venue cache is best-effort, so venue-service runs without it.
- scheduler-service's jobs are triggered by Cloud Scheduler, so nothing has to stay awake.
- The web build is hosted on Firebase Hosting; Android and iOS builds go through EAS (both from `roundup-ui`).
- Rough hosting cost ≈ $30/month; the Google Places API is the larger variable cost.

---

## Build order

### Done

Gateway, auth-service, user-service (profiles, squads, permissions), venue-service (search, details, favourites, history) and event-service (lifecycle, RSVP, suggestions, voting). CI on every PR. The client in `roundup-ui` is being built alongside, web and native together.

### Next — finish the core loop

**Done when:** a squad can plan an outing, vote on a venue, add the night's bills, and one person can mark a debt as paid.

1. **billing-service**: bills, receipt scanning, equal split, netting, mark as paid
2. **scheduler-service**: voting deadlines and closing the night happen on their own
3. **Launch blockers**: an SMS provider for auth, then `deploy.yml` to Cloud Run

### After that

- Notification service (FCM first, then WhatsApp) and realtime service
- Media service (album, QR download)
- Analytics (squad stats, wrapped), friends, cross-squad outings
- Reservation service, feature flags, admin
- Hardening: metrics, tracing, alerts, pen test of auth, GDPR export and deletion

---

_14 services · 1 backend language (Go) · one React Native + Expo client in `roundup-ui` (web, Android, iOS)_
