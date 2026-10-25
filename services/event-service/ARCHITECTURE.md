# event-service architecture

event-service plans nights out for a squad: create an event, answer whether you're coming, suggest places, vote, then run the night. It follows the same hexagonal layout as venue-service.

```
cmd/main.go                      wiring only
internal/
├── domain/                      types, statuses, roles, ports, sentinel errors
├── policy/                      policy.csv (who may do what), embedded
├── application/                 EventUseCase: event_usecase.go, participation.go, lifecycle.go, view.go
├── adapters/
│   ├── persistence/             Postgres via sqlc (db/queries → persistence/db)
│   └── venues/                  venue-service client (shared/serviceclient)
└── transport/http/              Fiber handlers, router, DTOs, error table
```

Squad membership comes from user-service through `shared/userclient`; the Casbin enforcer comes from `shared/authz`. Both are built in `cmd/main.go` and passed in directly.

## Lifecycle

```
DRAFT ──start voting──▶ VOTING ──close voting / deadline──▶ CONFIRMED ──start the night──▶ ACTIVE ──close / +12h──▶ CLOSED
```

| Status | RSVPs | Suggest / withdraw | Vote | Edit |
|---|---|---|---|---|
| `DRAFT` | ✅ | ✅ | | ✅ |
| `VOTING` | ✅ | | ✅ | ✅ |
| `CONFIRMED` | ✅ | | | |
| `ACTIVE` | ✅ | | | |
| `CLOSED` | | | | |

- **One step at a time:** `Status.CanMoveTo` allows only the next status; anything else is `ErrInvalidTransition` (409).
- **Starting voting** needs at least one suggestion (`ErrNoNominations`, 409).
- **The winner** is picked by the database when voting closes: the most votes, a tie (or no votes) going to the earliest suggestion. The same SQL runs for a manual close and for the deadline sweep.
- **`ACTIVE`** is "Start the night". billing-service will open the bill here; until it exists, the step only changes the status.
- **Closing** adds the winning place to the squad's history in venue-service. That call is best-effort: a failure is logged and the night still closes.

## Status checks happen in the database too

Every write that depends on the status checks it in the same SQL statement (`UPDATE … WHERE status = 'VOTING'`, `INSERT … SELECT … WHERE e.status = 'DRAFT'`). The use case checks first to give a clear error; the guarded statement makes the check race-free. If the statement changes nothing, the status moved in between and the caller gets 409.

## Deadlines without a cron

event-service has no scheduler of its own; scheduled jobs live in scheduler-service. It exposes two internal routes, which the gateway never routes:

| Route | Does |
|---|---|
| `POST /internal/v1/events/confirm-due` | Confirms every `VOTING` event past its `votingDeadline`. |
| `POST /internal/v1/events/close-due` | Closes every `CONFIRMED` or `ACTIVE` event 12 hours after it started, and records the visits as the event's creator. |

Each is a single guarded `UPDATE`. Postgres locks each row and re-checks the `WHERE` clause, so overlapping calls (a retry, two scheduler instances) move each event exactly once. `TestSweepsMoveEachEventOnceUnderConcurrency` checks this with four concurrent calls.

## Who sees what

- **Membership:** every route checks squad membership through user-service. A non-member gets 404 for the squad, and 404 for any event in it, so they can't tell an event exists.
- **Permissions:** see [docs/PERMISSIONS.md](docs/PERMISSIONS.md). `GET /events/:eventID` returns `permissions` so the app shows only the buttons that will work.
- **Counts:** only current members' answers and votes count toward RSVP counts, the going preview and `votedCount`. `totalVotes` and each place's `voteCount` count every vote cast, so "won with 6 of 11" doesn't change when someone leaves.

## Responses

Every route that changes an event answers with the full event detail, so the app redraws from one response. Errors map to statuses in one table, `transport/http/errors.go`:

| Status | Errors |
|---|---|
| 400 | `ErrInvalidEvent`, `ErrInvalidRSVP`, `ErrNotNominated`, bad paging |
| 403 | `ErrForbidden` |
| 404 | `ErrEventNotFound`, `ErrSquadNotFound`, `ErrNominationNotFound`, `ErrVenueNotFound` |
| 409 | `ErrInvalidTransition`, `ErrNoNominations`, `ErrAlreadyNominated` |
| 503 | user-service or venue-service unavailable |

## Tests

| Layer | How |
|---|---|
| domain, policy | Pure unit tests: every status pair, every role × action, policy line shape |
| application | In-memory fakes with the real `policy.csv` |
| persistence | Real Postgres via `shared/postgres/postgrestest` (set `TEST_POSTGRES_DSN`) |
| venues client | `httptest` server |
| router | `TestRoutesAreDocumented` against `shared/api/openapi.yaml` |
