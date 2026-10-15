# auth-service Architecture

## Context

auth-service logs users in with their phone number and issues the tokens every other service relies on. A user requests a one-time code (sent by SMS; logged in beta), verifies it, and receives:

- an **access token**: a 15-minute JWT signed with **Ed25519** (`shared/authtoken`). The gateway verifies it with the public key and turns its `sub` into the `X-User-ID` header for upstream services.
- a **refresh token**: 32 random bytes, valid 30 days, **rotated on every use**. Only its SHA-256 hash is stored.

Users themselves live in user-service. At login, auth-service calls user-service's internal `POST /api/v1/users` (create-or-get) to turn a verified phone into a user ID. All routes are public; they are how a client gets a token in the first place.

---

## Layer Diagram

```
┌──────────────────────────────────────────────────────────────────┐
│  TRANSPORT   internal/transport/http/                            │
│  AuthHandler: otp/request, otp/verify, refresh, logout           │
└──────────────────────────┬───────────────────────────────────────┘
                           ▼
┌──────────────────────────────────────────────────────────────────┐
│  APPLICATION internal/application/                               │
│  OTPUseCase (Request, Verify) · TokenUseCase (Refresh, Logout)   │
│  TokenIssuer (access JWTs + refresh tokens)                      │
└──────────────────────────┬───────────────────────────────────────┘
                           ▼
┌──────────────────────────────────────────────────────────────────┐
│  DOMAIN      internal/domain/   — imports only the standard lib  │
│  OTPCode, StoredRefreshToken, TokenPair, LoginResult, errors     │
│  Ports: OTPRepository, SessionRepository, OTPSender, UserDirectory│
└──────────────────────────▲───────────────────────────────────────┘
                           │ implemented by
┌──────────────────────────┴───────────────────────────────────────┐
│  ADAPTERS    internal/adapters/                                  │
│  persistence/ PostgresOTPRepository, PostgresSessionRepository   │
│  sms/         ConsoleSender (beta)                               │
│  users/       UserServiceClient (HTTP, 3 s timeout)              │
└──────────────────────────────────────────────────────────────────┘
```

---

## Login Flow

```
POST /otp/request {phone}
  ├─ phone valid (shared/phone)?               no → 400
  ├─ < 3 codes for this phone in 15 min?       no → 429
  └─ store sha256(phone:code), expires 5 min → OTPSender.Send → 202

POST /otp/verify {phone, code}
  ├─ newest code for phone, unused, unexpired? no → 401
  ├─ 5 wrong guesses already on this code?    yes → 429 (even for the right code)
  ├─ code matches (constant time; or OTP_DEV_CODE in beta)?  no → attempts++, 401
  ├─ user-service create-or-get                fails → 503 (code NOT consumed; retry works)
  ├─ consume code (only once, enforced in SQL)
  └─ new session + first refresh token → 200 {tokens, userID, isNewUser}
```

---

## Refresh Rotation and Reuse Detection

Each refresh marks the presented token **used** and issues a new one on the same session, in one transaction. A token is only ever exchanged once.

If a **used** token is presented again, it was copied: the legitimate client already holds the newer one. We can't tell which party is the thief, so the **whole session is revoked**; both must log in again. Logout revokes the session the same way.

Access tokens are not revoked by logout. They expire within 15 minutes; that is the cost of the gateway verifying them without a database lookup.

---

## HTTP Routes

| Method | Path | Success | Errors |
|---|---|---|---|
| POST | `/api/v1/auth/otp/request` | 202 | 400 bad phone, 429 too many requests |
| POST | `/api/v1/auth/otp/verify` | 200 `LoginResponse` | 400, 401 invalid code, 429 too many attempts, 503 user-service down |
| POST | `/api/v1/auth/refresh` | 200 `TokenPairResponse` | 401 invalid / expired / reused |
| POST | `/api/v1/auth/logout` | 204 | 401 unknown token |

## Error Mapping

| Domain sentinel | HTTP status |
|---|---|
| `ErrInvalidPhone`, body fails to bind | 400 |
| `ErrInvalidCode`, `ErrInvalidRefreshToken` | 401 |
| `ErrTooManyRequests`, `ErrTooManyAttempts` | 429 |
| `ErrUserServiceUnavailable` | 503 |
| anything else | 500 |

---

## Database Schema

| Table | Purpose |
|---|---|
| `otp_codes` | Hash of each code sent, with attempts, expiry and consumed time. Only the newest per phone is valid; older rows are kept for the request rate limit. |
| `sessions` | One per login (device): `user_id`, `revoked_at`. |
| `refresh_tokens` | `token_hash` (unique), `expires_at`, `used_at`; belongs to a session. |

## Configuration

| Variable | Default | Notes |
|---|---|---|
| `APP_ENV` | — | Required: `beta` or `prod` |
| `PORT` | 3001 | |
| `POSTGRES_DSN` | — | Required |
| `USER_SERVICE_URL` | — | Required; called directly, not via the gateway |
| `JWT_PRIVATE_KEY` | — | Required; Ed25519 PEM from `make jwt-keys` |
| `ACCESS_TOKEN_TTL` / `REFRESH_TOKEN_TTL` | 15m / 720h | Go durations |
| `OTP_DEV_CODE` | — | Beta only; startup panics if set in prod |

In prod, `cmd/main.go` refuses to start until a real SMS provider adapter exists; the console sender is beta-only.

## Testing Strategy

| Layer | Tests |
|---|---|
| application | In-memory fakes and an injected clock: rate limit, wrong/expired/old/used codes, attempt counting, dev code, user-service outage, rotation, reuse revoking the session |
| persistence | `postgrestest` against real Postgres (needs `TEST_POSTGRES_DSN`): counting, newest code, consume-once, rotation atomicity, revoke |
| users client | `httptest` user-service: 201/200, error statuses, bad bodies, unreachable, timeout |
| config | Defaults, and panics for dev code in prod, bad key, bad TTLs |
