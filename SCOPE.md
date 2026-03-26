# Roundup — Full Product & Technical Scope

> Complete reference document covering features, architecture, services, communication patterns, design patterns, and database schemas.

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
- [Build phases](#build-phases)

---

## Product overview

Roundup is a weekend squad outing planner. It handles the full arc of a group night out — from deciding where to go, to splitting the bill, settling up, and remembering the night.

**Core loop**

```
Plan → Vote → RSVP → Go out → Bill tab → Settle → Remember
```

**Target users** — friend groups of 5–15 people who go out regularly and currently coordinate over WhatsApp, split bills on Splitwise, and lose photos across everyone's camera rolls.

---

## Feature scope

### Planning & discovery

| Feature             | Description                                                                     |
| ------------------- | ------------------------------------------------------------------------------- |
| Venue voting        | Create a poll with deadline. Everyone votes. Highest score wins.                |
| Place discovery     | Google Places search filtered by vibe — rooftop, cocktail bar, pub, fine dining |
| Saved favourites    | Squad's go-to spots saved for quick access                                      |
| Venue history       | Every venue the squad has visited, with avg spend                               |
| AI recommendations  | Ranked suggestions based on past behaviour and current request                  |
| Table reservation   | Direct booking via Dineout / EazyDiner                                          |
| Deals & offers      | Venue-specific discounts surfaced during planning                               |
| Event templates     | Reuse a past outing's config (same squad, same split mode)                      |
| Budget cap          | Set max spend per person before the outing                                      |
| Recurring outings   | Every Friday auto-created with same squad                                       |
| Best day suggestion | AI analyses RSVP patterns to suggest highest-turnout day                        |
| Calendar sync       | Export outing to Google / Apple Calendar                                        |
| Invite links        | Shareable link — join without being in the squad                                |
| Guest mode          | RSVP and settle bill without creating an account                                |
| Cross-squad outings | Merge two squads for one event                                                  |

### Billing & payments

| Feature             | Description                                                         |
| ------------------- | ------------------------------------------------------------------- |
| Live bill tab       | Firestore real-time sync — multiple people add items simultaneously |
| Receipt OCR         | Photograph the bill — AI extracts line items automatically          |
| Split strategies    | Equal, by item, custom percentage                                   |
| Item assignment     | Tap a line item to claim it                                         |
| Tax & tip split     | Proportional or equal                                               |
| Multi-bill support  | Bar tab + dinner bill on the same outing                            |
| Debt simplification | Min-cash-flow graph — reduces N pairwise debts to minimum transfers |
| UPI settle-up       | Razorpay collect requests sent inside the app                       |
| Running balance     | Total owed / owing across all squads                                |
| Partial payments    | Pay a debt in instalments                                           |
| Cash settlement     | Mark a debt paid offline                                            |
| Payment proof       | Upload screenshot as confirmation                                   |
| Dispute resolution  | Flag an incorrect line item                                         |
| Bill audit log      | Every add / remove / assign recorded with timestamp and user        |

### Social & memory

| Feature            | Description                                                      |
| ------------------ | ---------------------------------------------------------------- |
| Photo album        | One album per outing, auto-collected from all squad members      |
| Face detection     | Google Vision API — bounding box per face per photo              |
| Face tagging       | Auto-match against squad profiles (≥0.85 confidence) or self-tag |
| QR download        | Scan QR to download ZIP of your tagged photos only               |
| Outing recap       | AI-generated summary after event closes                          |
| Shareable card     | Instagram-ready highlight card for the outing                    |
| On this day        | Anniversary reminder for past outings                            |
| Video reel         | Auto-compilation of photos into a short reel                     |
| Reactions          | Emoji reactions on photos                                        |
| Caption generation | AI suggests captions for photos                                  |
| Squad stats        | Most visited spot, biggest spender, most frequent RSVPer         |
| Leaderboard        | Rankings per squad — fun data to roast each other with           |
| Streak tracking    | Consecutive weekly outings                                       |
| Vibes rating       | Rate the outing after it closes                                  |
| Yearly wrapped     | Spotify-style annual recap                                       |

### Notifications

| Feature             | Description                                             |
| ------------------- | ------------------------------------------------------- |
| RSVP nudge          | Reminder if not responded before deadline               |
| Debt reminder       | Gentle nudge 48 hours after debt created                |
| Friday prompt       | Weekly push if no outing planned for the weekend        |
| Vote deadline alert | Notification 1 hour before voting closes                |
| AI nudges           | Personalised, context-aware message generation          |
| Photo tagged        | Notification when you appear in a new photo             |
| Payment confirmed   | Instant notification when debt settled                  |
| Download alert      | Notifies photo uploader when someone downloads          |
| Quiet hours         | User-set window where no notifications are sent         |
| Channels            | FCM push (primary), WhatsApp Business API, SMS fallback |

### Real-time

| Feature                | Description                                          |
| ---------------------- | ---------------------------------------------------- |
| Live RSVP count        | Headcount updates on planning screen without refresh |
| Live vote tally        | See votes come in as friends cast them               |
| Typing indicators      | "Priya is adding an item..." on bill tab             |
| Presence               | Who is currently viewing the outing                  |
| Payment confirmed live | Debt status clears instantly on both screens         |

### AI features (LangChain + Gemini 1.5 Pro)

| Feature               | Chain type | Description                                                   |
| --------------------- | ---------- | ------------------------------------------------------------- |
| Receipt OCR           | LLM chain  | Photo → Vision API → LLM extraction → structured JSON         |
| Venue recommendations | Agent      | Tools: venue search, squad history, preferences → ranked list |
| RAG queries           | RAG chain  | "Where did we go most?" over pgvector-stored squad history    |
| Smart nudges          | LLM chain  | Context-aware debt reminder generation                        |
| Outing recap          | LLM chain  | Auto-summary on event close                                   |
| Planning assistant    | Agent      | Conversational — "plan something for 8 of us Saturday"        |
| Spend prediction      | Predictor  | "This venue averages ₹X for your group size"                  |
| Caption generation    | LLM chain  | 3 Instagram-style caption options per photo                   |
| Yearly wrapped        | LLM chain  | Narrative annual summary                                      |

### Monetisation

| Feature                | Description                                        |
| ---------------------- | -------------------------------------------------- |
| Venue partnerships     | Featured placement in discovery                    |
| Commission on bookings | Revenue on Dineout reservations                    |
| Roundup Pro            | Unlimited history, advanced analytics, priority AI |
| Referral program       | Invite friends, earn Pro credits                   |
| Deals marketplace      | Exclusive squad offers from partner venues         |

---

## Architecture

### Layers

```
┌─────────────────────────────────────────────────────────┐
│  Flutter app                                            │
└────────────────────────┬────────────────────────────────┘
                         │ HTTPS
┌────────────────────────▼────────────────────────────────┐
│  API Gateway (Go)                                       │
│  routing · rate limiting · auth forwarding              │
└──┬──────┬──────┬──────┬──────┬──────┬──────┬───────────┘
   │      │      │      │      │      │      │
   ▼      ▼      ▼      ▼      ▼      ▼      ▼
 Auth   User  Event Billing Payment Venue  Media ...
 (Nest) (Java)(Java)(Java)  (Go)    (Go)   (Go)

                   │ Kafka (async events)
         ┌─────────┴──────────┐
         ▼                    ▼
   Notification           Analytics
      (Go)                 (Java)

┌──────────────────────────────────────────────────────┐
│  Data layer                                          │
│  PostgreSQL (per service) · Redis · Firestore        │
└──────────────────────────────────────────────────────┘

┌──────────────────────────────────────────────────────┐
│  External                                            │
│  Firebase · Razorpay · Cloudinary · Google Places    │
│  Google Vision · FCM · WhatsApp Business             │
└──────────────────────────────────────────────────────┘
```

### Communication strategy

**REST over HTTP** — all synchronous calls. Simple, debuggable, no tooling overhead.

**Kafka** — all asynchronous domain events. Services are fully decoupled — publishers never know who consumes.

> gRPC was considered and dropped. At Roundup's scale the latency benefit is single-digit milliseconds. REST over an internal network is adequate and far easier to debug.

---

## Services

| Service              | Language   | Port | DB              |
| -------------------- | ---------- | ---- | --------------- |
| API Gateway          | Go (Fiber) | 8080 | —               |
| Auth Service         | NestJS     | 3001 | 5433            |
| User Service         | Java       | 3002 | 5434            |
| Event Service        | Java       | 3003 | 5435            |
| Billing Service      | Java       | 3004 | 5436            |
| Payment Service      | Go         | 3005 | 5437            |
| Venue Service        | Go         | 3006 | 5438            |
| Notification Service | Go         | 3007 | —               |
| Media Service        | Go         | 3008 | 5439            |
| Realtime Service     | Go         | 3009 | —               |
| AI Service           | Python     | 3010 | 5441 (pgvector) |
| Reservation Service  | Go         | 3011 | 5442            |
| Analytics Service    | Java       | 3012 | 5441            |
| Feature Flag Service | Go         | 3013 | 5440            |
| Admin Service        | NestJS     | 3014 | —               |

All services share Redis on `6379` and Kafka on `9092`.

---

## Communication

### Kafka topics

| Topic                    | Publisher | Consumers                         |
| ------------------------ | --------- | --------------------------------- |
| `rsvp.updated`           | Event     | Notification, Realtime, Analytics |
| `event.created`          | Event     | Analytics                         |
| `event.closed`           | Event     | Billing, AI, Analytics            |
| `vote.cast`              | Event     | Realtime, Analytics               |
| `event.confirmed`        | Event     | Realtime, Reservation             |
| `tab.opened`             | Billing   | Realtime                          |
| `tab.closed`             | Billing   | Notification, AI                  |
| `debt.created`           | Billing   | Notification, Realtime            |
| `debt.settled`           | Billing   | Analytics, Realtime               |
| `payment.initiated`      | Payment   | Analytics                         |
| `payment.confirmed`      | Payment   | Billing, Notification, Realtime   |
| `payment.failed`         | Payment   | Billing, Notification             |
| `photo.uploaded`         | Media     | Media (tagging worker)            |
| `photo.tagged`           | Media     | Notification, Realtime            |
| `outing.recap.requested` | AI        | Notification                      |
| `notification.send`      | AI        | Notification                      |

### Synchronous REST calls

| Caller          | Callee               | Endpoint                  | When                                           |
| --------------- | -------------------- | ------------------------- | ---------------------------------------------- |
| Gateway         | Auth Service         | `POST /auth/validate`     | Every authenticated request (Redis cached)     |
| Billing Service | Payment Service      | `POST /payments/initiate` | On debt settlement request                     |
| AI Service      | Venue Service        | `GET /venues/search`      | Venue recommendation agent tool call           |
| Any service     | Feature Flag Service | `GET /flags/:key?userId=` | On feature-gated code paths (Redis cached 60s) |

---

## Design patterns

| Pattern        | Where                      | What it does                                                                                                        |
| -------------- | -------------------------- | ------------------------------------------------------------------------------------------------------------------- |
| Strategy       | Billing — split modes      | `EqualSplit`, `ByItemSplit`, `CustomSplit` all implement `SplitStrategy`. Swapped at runtime based on `split_mode`. |
| Strategy       | Auth — providers           | `FirebaseStrategy`, `JwtStrategy`, `ApiKeyStrategy` all implement `IAuthStrategy`. Factory picks at request time.   |
| Factory        | Auth                       | `AuthStrategyFactory` holds a `Map<AuthProvider, IAuthStrategy>` and resolves via `resolve(provider)`.              |
| Repository     | All Java + NestJS services | All DB queries behind an interface. Services never touch the ORM directly.                                          |
| Port / Adapter | Auth — Firebase            | `AuthProviderPort` interface. `FirebaseAuthAdapter` implements it. Swap provider without touching strategy.         |
| Port / Adapter | Venue — Places             | `VenueProvider` interface. `GooglePlacesAdapter` implements it. Future: `ZomatoAdapter`, `FoursquareAdapter`.       |
| Command        | Billing — bill tab         | Every mutation (`AddItem`, `RemoveItem`, `AssignItem`, `CloseBill`) is a Command. Enables audit log and undo.       |
| State machine  | Event Service              | `DRAFT → VOTING → CONFIRMED → ACTIVE → CLOSED`. Invalid transitions throw.                                          |
| State machine  | Billing — debts            | `PENDING → REQUESTED → SETTLED`.                                                                                    |
| State machine  | Reservation Service        | `PENDING → CONFIRMED → CANCELLED → NO_SHOW`.                                                                        |
| Observer       | All services               | `EventEmitter2` domain events. `rsvp.updated` triggers notification without Billing knowing Notification exists.    |
| Facade         | Venue Service              | Wraps Google Places API + Redis cache. Callers call `venueService.search()` — no knowledge of Places API.           |
| Facade         | Payment Service            | Wraps Razorpay SDK. Callers call `paymentService.initiate()` — no knowledge of Razorpay internals.                  |
| Decorator      | NestJS common              | `AuthGuard`, `TransformInterceptor`, `LoggingInterceptor`, `ZodValidationPipe`, `GlobalExceptionFilter`.            |
| Singleton      | All services               | DB pool, Redis client, Firebase Admin SDK — one instance, DI-managed.                                               |
| Feature flags  | All services               | `feature_flags` table + Redis cache. Kill switch for every major feature.                                           |

---

## Database schemas

### Auth Service (PostgreSQL)

```sql
-- sessions
CREATE TABLE sessions (
  id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id      UUID NOT NULL,
  firebase_uid TEXT NOT NULL,
  device_token TEXT,
  ip_address   TEXT,
  created_at   TIMESTAMP NOT NULL DEFAULT now(),
  expires_at   TIMESTAMP NOT NULL,
  revoked_at   TIMESTAMP
);

CREATE INDEX idx_sessions_user_id ON sessions(user_id);
CREATE INDEX idx_sessions_firebase_uid ON sessions(firebase_uid);
```

---

### User Service (PostgreSQL)

```sql
-- users
CREATE TABLE users (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  uid         TEXT UNIQUE NOT NULL,          -- Firebase UID
  phone       TEXT UNIQUE,
  email       TEXT UNIQUE,
  name        TEXT NOT NULL,
  avatar_url  TEXT,
  deleted_at  TIMESTAMP,                     -- soft delete
  created_at  TIMESTAMP NOT NULL DEFAULT now(),
  updated_at  TIMESTAMP NOT NULL DEFAULT now()
);

-- squads
CREATE TABLE squads (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name        TEXT NOT NULL,
  avatar_url  TEXT,
  created_by  UUID NOT NULL REFERENCES users(id),
  created_at  TIMESTAMP NOT NULL DEFAULT now()
);

-- squad_members
CREATE TABLE squad_members (
  squad_id    UUID NOT NULL REFERENCES squads(id),
  user_id     UUID NOT NULL REFERENCES users(id),
  role        TEXT NOT NULL DEFAULT 'MEMBER', -- ORGANISER | MEMBER
  joined_at   TIMESTAMP NOT NULL DEFAULT now(),
  PRIMARY KEY (squad_id, user_id)
);

-- device_tokens (FCM)
CREATE TABLE device_tokens (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id     UUID NOT NULL REFERENCES users(id),
  token       TEXT NOT NULL UNIQUE,
  platform    TEXT NOT NULL,                 -- IOS | ANDROID
  created_at  TIMESTAMP NOT NULL DEFAULT now()
);

-- user_face_profiles (for face recognition matching)
CREATE TABLE user_face_profiles (
  user_id      UUID PRIMARY KEY REFERENCES users(id),
  cloudinary_id TEXT NOT NULL,              -- reference photo
  embeddings   JSONB,                       -- face embedding vector
  updated_at   TIMESTAMP NOT NULL DEFAULT now()
);

-- friend_connections
CREATE TABLE friend_connections (
  user_id_a   UUID NOT NULL REFERENCES users(id),
  user_id_b   UUID NOT NULL REFERENCES users(id),
  created_at  TIMESTAMP NOT NULL DEFAULT now(),
  PRIMARY KEY (user_id_a, user_id_b),
  CHECK (user_id_a < user_id_b)             -- prevent duplicate pairs
);

CREATE INDEX idx_users_uid ON users(uid);
CREATE INDEX idx_users_phone ON users(phone);
CREATE INDEX idx_squad_members_squad_id ON squad_members(squad_id);
CREATE INDEX idx_squad_members_user_id ON squad_members(user_id);
CREATE INDEX idx_device_tokens_user_id ON device_tokens(user_id);
```

---

### Event Service (PostgreSQL)

```sql
-- event type and status enums
CREATE TYPE event_type AS ENUM ('DRINKS', 'DINING', 'BOTH');
CREATE TYPE event_status AS ENUM ('DRAFT', 'VOTING', 'CONFIRMED', 'ACTIVE', 'CLOSED');
CREATE TYPE rsvp_status AS ENUM ('PENDING', 'GOING', 'NOT_GOING', 'MAYBE');

-- events
CREATE TABLE events (
  id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  squad_id     UUID NOT NULL,
  name         TEXT NOT NULL,
  type         event_type NOT NULL,
  status       event_status NOT NULL DEFAULT 'DRAFT',
  venue_id     TEXT,                          -- Google Place ID, set after voting
  venue_name   TEXT,
  created_by   UUID NOT NULL,
  deadline     TIMESTAMP,                     -- voting deadline
  event_date   TIMESTAMP,
  max_budget_paise BIGINT,                   -- optional budget cap
  created_at   TIMESTAMP NOT NULL DEFAULT now(),
  updated_at   TIMESTAMP NOT NULL DEFAULT now()
);

-- event_members
CREATE TABLE event_members (
  event_id    UUID NOT NULL REFERENCES events(id),
  user_id     UUID NOT NULL,
  rsvp_status rsvp_status NOT NULL DEFAULT 'PENDING',
  is_guest    BOOLEAN NOT NULL DEFAULT false,
  joined_at   TIMESTAMP NOT NULL DEFAULT now(),
  PRIMARY KEY (event_id, user_id)
);

-- votes
CREATE TABLE votes (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  event_id    UUID NOT NULL REFERENCES events(id),
  user_id     UUID NOT NULL,
  venue_id    TEXT NOT NULL,                 -- Google Place ID
  venue_name  TEXT NOT NULL,
  cast_at     TIMESTAMP NOT NULL DEFAULT now(),
  UNIQUE (event_id, user_id)                -- one vote per user per event
);

-- invite_links
CREATE TABLE invite_links (
  token       UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  event_id    UUID NOT NULL REFERENCES events(id),
  created_by  UUID NOT NULL,
  max_uses    INT NOT NULL DEFAULT 10,
  use_count   INT NOT NULL DEFAULT 0,
  expires_at  TIMESTAMP NOT NULL,
  created_at  TIMESTAMP NOT NULL DEFAULT now()
);

CREATE INDEX idx_events_squad_id ON events(squad_id);
CREATE INDEX idx_events_status ON events(status);
CREATE INDEX idx_event_members_event_id ON event_members(event_id);
CREATE INDEX idx_event_members_user_id ON event_members(user_id);
CREATE INDEX idx_votes_event_id ON votes(event_id);
CREATE INDEX idx_invite_links_token ON invite_links(token);
```

---

### Billing Service (PostgreSQL)

```sql
CREATE TYPE split_mode AS ENUM ('EQUAL', 'BY_ITEM', 'CUSTOM');
CREATE TYPE bill_status AS ENUM ('OPEN', 'CLOSED');
CREATE TYPE debt_status AS ENUM ('PENDING', 'REQUESTED', 'SETTLED');

-- bills
CREATE TABLE bills (
  id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  event_id     UUID NOT NULL,
  status       bill_status NOT NULL DEFAULT 'OPEN',
  split_mode   split_mode NOT NULL DEFAULT 'EQUAL',
  total_paise  BIGINT NOT NULL DEFAULT 0,
  created_at   TIMESTAMP NOT NULL DEFAULT now(),
  closed_at    TIMESTAMP
);

-- bill_items
CREATE TABLE bill_items (
  id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  bill_id      UUID NOT NULL REFERENCES bills(id),
  name         TEXT NOT NULL,
  amount_paise BIGINT NOT NULL,
  added_by     UUID NOT NULL,
  created_at   TIMESTAMP NOT NULL DEFAULT now(),
  deleted_at   TIMESTAMP                         -- soft delete
);

-- item_assignments (who is responsible for which item)
CREATE TABLE item_assignments (
  item_id      UUID NOT NULL REFERENCES bill_items(id),
  user_id      UUID NOT NULL,
  share_paise  BIGINT NOT NULL,
  PRIMARY KEY (item_id, user_id)
);

-- debts (generated after bill closes + simplification runs)
CREATE TABLE debts (
  id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  bill_id      UUID NOT NULL REFERENCES bills(id),
  from_user    UUID NOT NULL,                   -- owes money
  to_user      UUID NOT NULL,                   -- receives money
  amount_paise BIGINT NOT NULL,
  status       debt_status NOT NULL DEFAULT 'PENDING',
  created_at   TIMESTAMP NOT NULL DEFAULT now(),
  settled_at   TIMESTAMP
);

-- split_configs (for CUSTOM mode)
CREATE TABLE split_configs (
  bill_id     UUID NOT NULL REFERENCES bills(id),
  user_id     UUID NOT NULL,
  percentage  NUMERIC(5,2) NOT NULL,
  PRIMARY KEY (bill_id, user_id)
);

-- audit_log (Command pattern — every mutation recorded)
CREATE TABLE audit_log (
  id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  bill_id      UUID NOT NULL REFERENCES bills(id),
  action       TEXT NOT NULL,                  -- ADD_ITEM | REMOVE_ITEM | ASSIGN | CLOSE
  payload      JSONB NOT NULL,
  performed_by UUID NOT NULL,
  created_at   TIMESTAMP NOT NULL DEFAULT now()
);

CREATE INDEX idx_bills_event_id ON bills(event_id);
CREATE INDEX idx_bill_items_bill_id ON bill_items(bill_id);
CREATE INDEX idx_debts_bill_id ON debts(bill_id);
CREATE INDEX idx_debts_from_user ON debts(from_user);
CREATE INDEX idx_debts_to_user ON debts(to_user);
CREATE INDEX idx_debts_status ON debts(status);
CREATE INDEX idx_audit_log_bill_id ON audit_log(bill_id);
```

---

### Payment Service (PostgreSQL)

```sql
CREATE TYPE payment_status AS ENUM ('INITIATED', 'CAPTURED', 'FAILED', 'REFUNDED');

-- payment_requests
CREATE TABLE payment_requests (
  id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  debt_id             UUID NOT NULL UNIQUE,         -- idempotency key
  razorpay_payment_id TEXT UNIQUE,
  razorpay_order_id   TEXT,
  from_user           UUID NOT NULL,
  to_user             UUID NOT NULL,
  amount_paise        BIGINT NOT NULL,
  status              payment_status NOT NULL DEFAULT 'INITIATED',
  created_at          TIMESTAMP NOT NULL DEFAULT now(),
  captured_at         TIMESTAMP
);

-- payment_events (raw webhook payloads for replay)
CREATE TABLE payment_events (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  payment_id  UUID REFERENCES payment_requests(id),
  event_type  TEXT NOT NULL,                        -- payment.captured | payment.failed
  raw_payload JSONB NOT NULL,
  processed   BOOLEAN NOT NULL DEFAULT false,
  created_at  TIMESTAMP NOT NULL DEFAULT now()
);

-- refunds
CREATE TABLE refunds (
  id                 UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  payment_id         UUID NOT NULL REFERENCES payment_requests(id),
  razorpay_refund_id TEXT UNIQUE,
  amount_paise       BIGINT NOT NULL,
  status             TEXT NOT NULL DEFAULT 'PENDING',
  created_at         TIMESTAMP NOT NULL DEFAULT now()
);

CREATE INDEX idx_payment_requests_debt_id ON payment_requests(debt_id);
CREATE INDEX idx_payment_requests_status ON payment_requests(status);
CREATE INDEX idx_payment_events_payment_id ON payment_events(payment_id);
CREATE INDEX idx_payment_events_processed ON payment_events(processed);
```

---

### Venue Service (PostgreSQL)

```sql
-- saved_venues (squad favourites)
CREATE TABLE saved_venues (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  squad_id    UUID NOT NULL,
  user_id     UUID NOT NULL,                    -- who saved it
  place_id    TEXT NOT NULL,                    -- Google Place ID
  name        TEXT NOT NULL,
  saved_at    TIMESTAMP NOT NULL DEFAULT now(),
  UNIQUE (squad_id, place_id)
);

-- venue_visits (recorded when event.confirmed fires)
CREATE TABLE venue_visits (
  id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  squad_id         UUID NOT NULL,
  event_id         UUID NOT NULL,
  place_id         TEXT NOT NULL,
  name             TEXT NOT NULL,
  avg_spend_paise  BIGINT,
  visited_at       TIMESTAMP NOT NULL DEFAULT now()
);

CREATE INDEX idx_saved_venues_squad_id ON saved_venues(squad_id);
CREATE INDEX idx_venue_visits_squad_id ON venue_visits(squad_id);
CREATE INDEX idx_venue_visits_place_id ON venue_visits(place_id);
```

---

### Media Service (PostgreSQL)

```sql
CREATE TYPE tagging_status AS ENUM ('PENDING', 'PROCESSING', 'DONE', 'FAILED');

-- photos
CREATE TABLE photos (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  event_id        UUID NOT NULL,
  uploaded_by     UUID NOT NULL,
  cloudinary_id   TEXT NOT NULL UNIQUE,         -- public_id — not the full URL
  width           INT,
  height          INT,
  tagging_status  tagging_status NOT NULL DEFAULT 'PENDING',
  created_at      TIMESTAMP NOT NULL DEFAULT now()
);

-- photo_faces (one row per detected face per photo)
CREATE TABLE photo_faces (
  id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  photo_id      UUID NOT NULL REFERENCES photos(id),
  user_id       UUID,                           -- null = unknown face
  confidence    REAL,                           -- 0.0 – 1.0
  bounding_box  JSONB NOT NULL,                 -- { x, y, width, height } normalised
  tagged_by     TEXT NOT NULL DEFAULT 'AUTO',   -- AUTO | SELF
  created_at    TIMESTAMP NOT NULL DEFAULT now()
);

-- download_tokens (QR-gated photo downloads)
CREATE TABLE download_tokens (
  token       TEXT PRIMARY KEY,                 -- signed JWT stored for revocation
  user_id     UUID NOT NULL,
  event_id    UUID NOT NULL,
  expires_at  TIMESTAMP NOT NULL,
  used_at     TIMESTAMP                         -- null = not yet downloaded
);

CREATE INDEX idx_photos_event_id ON photos(event_id);
CREATE INDEX idx_photos_uploaded_by ON photos(uploaded_by);
CREATE INDEX idx_photo_faces_photo_id ON photo_faces(photo_id);
CREATE INDEX idx_photo_faces_user_id ON photo_faces(user_id);
CREATE INDEX idx_download_tokens_user_event ON download_tokens(user_id, event_id);
```

---

### Notification Service (no persistent DB — logs only)

```sql
-- notification_log
CREATE TABLE notification_log (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id     UUID NOT NULL,
  channel     TEXT NOT NULL,                   -- FCM | WHATSAPP | SMS
  type        TEXT NOT NULL,                   -- RSVP_NUDGE | DEBT_REMINDER | etc.
  title       TEXT,
  body        TEXT NOT NULL,
  status      TEXT NOT NULL DEFAULT 'SENT',    -- SENT | DELIVERED | FAILED
  sent_at     TIMESTAMP NOT NULL DEFAULT now()
);

CREATE INDEX idx_notification_log_user_id ON notification_log(user_id);
CREATE INDEX idx_notification_log_type ON notification_log(type);
```

---

### Analytics Service (PostgreSQL + materialized views)

```sql
-- raw event ingestion
CREATE TABLE analytics_events (
  id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  squad_id     UUID,
  user_id      UUID,
  event_type   TEXT NOT NULL,
  payload      JSONB NOT NULL,
  occurred_at  TIMESTAMP NOT NULL DEFAULT now()
);

CREATE INDEX idx_analytics_events_squad_id ON analytics_events(squad_id);
CREATE INDEX idx_analytics_events_user_id ON analytics_events(user_id);
CREATE INDEX idx_analytics_events_type ON analytics_events(event_type);
CREATE INDEX idx_analytics_events_occurred_at ON analytics_events(occurred_at);

-- materialized view: squad summary stats
CREATE MATERIALIZED VIEW squad_stats AS
SELECT
  squad_id,
  COUNT(DISTINCT CASE WHEN event_type = 'event.closed' THEN payload->>'event_id' END) AS total_outings,
  SUM(CASE WHEN event_type = 'debt.settled' THEN (payload->>'amount_paise')::BIGINT ELSE 0 END) AS total_spend_paise,
  MAX(occurred_at) AS last_outing_at
FROM analytics_events
GROUP BY squad_id;

-- materialized view: per-user stats
CREATE MATERIALIZED VIEW user_stats AS
SELECT
  user_id,
  COUNT(DISTINCT CASE WHEN event_type = 'rsvp.updated'
    AND payload->>'status' = 'GOING' THEN payload->>'event_id' END) AS outings_attended,
  SUM(CASE WHEN event_type = 'debt.settled'
    AND payload->>'direction' = 'paid'
    THEN (payload->>'amount_paise')::BIGINT ELSE 0 END) AS total_paid_paise
FROM analytics_events
GROUP BY user_id;

-- materialized view: venue visit counts per squad
CREATE MATERIALIZED VIEW venue_stats AS
SELECT
  squad_id,
  payload->>'place_id'   AS place_id,
  payload->>'venue_name' AS venue_name,
  COUNT(*)               AS visit_count,
  AVG((payload->>'avg_spend_paise')::BIGINT) AS avg_spend_paise
FROM analytics_events
WHERE event_type = 'event.confirmed'
GROUP BY squad_id, payload->>'place_id', payload->>'venue_name';

-- refresh all views nightly (via pg_cron at 2am)
-- SELECT cron.schedule('0 2 * * *', 'REFRESH MATERIALIZED VIEW CONCURRENTLY squad_stats');
-- SELECT cron.schedule('0 2 * * *', 'REFRESH MATERIALIZED VIEW CONCURRENTLY user_stats');
-- SELECT cron.schedule('0 2 * * *', 'REFRESH MATERIALIZED VIEW CONCURRENTLY venue_stats');
```

---

### Feature Flag Service (PostgreSQL)

```sql
-- feature_flags
CREATE TABLE feature_flags (
  key         TEXT PRIMARY KEY,
  enabled     BOOLEAN NOT NULL DEFAULT false,
  rollout     INT NOT NULL DEFAULT 100,          -- 0–100 percentage
  description TEXT,
  metadata    JSONB,
  updated_at  TIMESTAMP NOT NULL DEFAULT now()
);

-- user_overrides (per-user flag value, takes precedence over global)
CREATE TABLE user_overrides (
  flag_key    TEXT NOT NULL REFERENCES feature_flags(key),
  user_id     UUID NOT NULL,
  enabled     BOOLEAN NOT NULL,
  created_at  TIMESTAMP NOT NULL DEFAULT now(),
  PRIMARY KEY (flag_key, user_id)
);

-- seed flags
INSERT INTO feature_flags (key, enabled, rollout, description) VALUES
  ('memory_wall',         false, 100, 'Photo upload and memory wall'),
  ('new_bill_tab_v2',     false, 100, 'Redesigned live bill tab UI'),
  ('squad_stats',         false, 100, 'Squad leaderboard and analytics'),
  ('razorpay_payments',   false, 100, 'UPI settle-up via Razorpay'),
  ('ai_recommendations',  false, 10,  'AI venue recommendations (10% rollout)'),
  ('receipt_ocr',         false, 100, 'Receipt photo to line items'),
  ('table_reservation',   false, 100, 'Direct table booking via Dineout');
```

---

### Reservation Service (PostgreSQL)

```sql
CREATE TYPE reservation_status AS ENUM ('PENDING', 'CONFIRMED', 'CANCELLED', 'NO_SHOW');

-- reservations
CREATE TABLE reservations (
  id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  event_id         UUID NOT NULL,
  venue_place_id   TEXT NOT NULL,
  provider         TEXT NOT NULL DEFAULT 'DINEOUT',
  reservation_ref  TEXT UNIQUE,                -- provider's booking reference
  status           reservation_status NOT NULL DEFAULT 'PENDING',
  party_size       INT NOT NULL,
  contact_phone    TEXT NOT NULL,
  reservation_time TIMESTAMP NOT NULL,
  created_at       TIMESTAMP NOT NULL DEFAULT now(),
  updated_at       TIMESTAMP NOT NULL DEFAULT now()
);

-- reservation_events (webhook log + replay)
CREATE TABLE reservation_events (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  reservation_id  UUID NOT NULL REFERENCES reservations(id),
  event_type      TEXT NOT NULL,
  raw_payload     JSONB NOT NULL,
  created_at      TIMESTAMP NOT NULL DEFAULT now()
);

CREATE INDEX idx_reservations_event_id ON reservations(event_id);
CREATE UNIQUE INDEX idx_reservations_event_id_unique ON reservations(event_id)
  WHERE status NOT IN ('CANCELLED');
```

---

### AI Service (PostgreSQL + pgvector)

```sql
-- requires: CREATE EXTENSION vector;

-- squad_memories (embeddings of past outings for RAG)
CREATE TABLE squad_memories (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  squad_id    UUID NOT NULL,
  event_id    UUID NOT NULL UNIQUE,
  content     TEXT NOT NULL,                   -- text representation of the outing
  embedding   vector(768) NOT NULL,            -- Gemini embedding dimension
  metadata    JSONB,                           -- venue, spend, members, date
  created_at  TIMESTAMP NOT NULL DEFAULT now()
);

-- ai_request_log
CREATE TABLE ai_request_log (
  id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  squad_id     UUID,
  user_id      UUID,
  request_type TEXT NOT NULL,                  -- RECEIPT_OCR | RECOMMEND | QUERY | NUDGE
  input        JSONB NOT NULL,
  output       JSONB,
  tokens_used  INT,
  duration_ms  INT,
  created_at   TIMESTAMP NOT NULL DEFAULT now()
);

CREATE INDEX idx_squad_memories_squad_id ON squad_memories(squad_id);
CREATE INDEX idx_squad_memories_embedding ON squad_memories
  USING ivfflat (embedding vector_cosine_ops) WITH (lists = 100);
```

---

### Firestore (live bill tab — not PostgreSQL)

Firestore is used only for the live bill tab. When the tab closes, the final state is synced to the Billing Service's PostgreSQL.

```
bill_tabs/                              (collection)
  {eventId}/                            (document)
    status: 'OPEN' | 'CLOSED'
    createdAt: timestamp
    closedAt: timestamp | null

    items/                              (subcollection)
      {itemId}/                         (document)
        name: string
        amount_paise: number
        added_by: string (userId)
        added_at: timestamp
        deleted: boolean

    presence/                           (subcollection)
      {userId}/                         (document)
        name: string
        last_seen: timestamp
        typing: boolean
```

---

## SDKs & third-party

| Category  | SDK                                  | Free tier                        | Used by              |
| --------- | ------------------------------------ | -------------------------------- | -------------------- |
| Auth      | `firebase-admin`                     | 10,000 OTPs/month free           | Auth Service         |
| Push      | Firebase FCM                         | Unlimited forever                | Notification Service |
| Real-time | Firestore                            | 50k reads/day, 20k writes/day    | Media + Billing      |
| Maps      | `googlemaps/google-maps-services-go` | 10,000 calls/month per SKU       | Venue Service        |
| Payments  | `razorpay` (npm)                     | Test mode free, no KYC           | Payment Service      |
| Media     | `cloudinary-go`                      | 25 credits/month, no card        | Media Service        |
| Vision    | `@google-cloud/vision`               | 1,000 units/month free           | Media Service        |
| Queue     | BullMQ + Redis                       | Self-hosted on Redis             | Notification, Media  |
| AI        | LangChain Python                     | Pay per token (Gemini free tier) | AI Service           |
| QR        | `skip2/go-qrcode`                    | Free open source                 | Media Service        |

---

## Build phases

### Phase 1 — Foundation (Month 1)

| Person | Work                                                                    |
| ------ | ----------------------------------------------------------------------- |
| A      | API Gateway (Go) + Auth Service (NestJS) + Firebase integration         |
| B      | User Service (Java) — profiles, squads, device tokens                   |
| C      | Infra — Docker Compose, Kafka, Redis, Postgres per service, CI skeleton |
| D      | Realtime Service (Go) — WebSocket hub, connection registry, presence    |

**Done when:** a user can sign up with phone OTP, create a squad, and the gateway routes authenticated requests.

### Phase 2 — Core domain (Month 2–3)

| Person | Work                                                                         |
| ------ | ---------------------------------------------------------------------------- |
| A      | Event Service (Java) — state machine, RSVP, voting, invite links, guest mode |
| B      | Billing Service (Java) — tab, strategies, debt graph, command pattern        |
| C      | Payment Service (Go) + Venue Service (Go) + Feature Flag Service (Go)        |
| D      | Realtime wired — live RSVP, live votes, live bill tab typing indicators      |

**Done when:** a squad can plan an outing, vote on a venue, open a bill tab, and one person can settle a debt.

### Phase 3 — Media + notifications + reservation (Month 4)

| Person | Work                                                                    |
| ------ | ----------------------------------------------------------------------- |
| A      | Media Service (Go) — Cloudinary, face detection pipeline, QR download   |
| B      | Notification Service (Go) + Analytics Service (Java)                    |
| C      | Reservation Service (Go) — Dineout integration, booking state machine   |
| D      | AI Service (Python) — LangChain setup, receipt OCR, pgvector embeddings |

### Phase 4 — AI features + admin (Month 5)

| Person | Work                                                                     |
| ------ | ------------------------------------------------------------------------ |
| A      | Admin Service (NestJS) — user mgmt, payment ops, webhook replay          |
| B      | Venue partnerships — deals, featured placement, partner analytics        |
| C      | Friend graph, cross-squad outings, yearly wrapped                        |
| D      | AI — venue recommendation agent, smart nudges, outing recap, RAG queries |

### Phase 5 — Hardening (Month 6)

| Person | Work                                                                |
| ------ | ------------------------------------------------------------------- |
| A      | E2E tests, contract tests, security audit, pen test auth + payments |
| B      | Java service profiling, GC tuning, DB index audit                   |
| C      | GCP production deploy — Cloud Run, Cloud SQL, Memorystore           |
| D      | Grafana + Loki + Prometheus, distributed tracing, alerting          |

---

_15 services · 4 languages (Go, Java, NestJS, Python) · 4 people · 6 months_
