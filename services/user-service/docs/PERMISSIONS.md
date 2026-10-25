# Squad permissions

Who may do what in a squad is decided by a policy, not by Go code. The policy is two files, read by [Casbin](https://casbin.org) when the service starts:

| File | What it holds |
|---|---|
| [`internal/adapters/authz/model.conf`](../internal/adapters/authz/model.conf) | The *shape* of a permission check: what a question looks like and how it is matched. Rarely changes. |
| [`internal/adapters/authz/policy.csv`](../internal/adapters/authz/policy.csv) | The *rules*: which role may take which action. This is the file you edit. |

Both are embedded into the binary with `go:embed`, so a change needs a rebuild and redeploy, and goes through code review like any other change.

Code asks one question, `Authorizer.Can(role, action)`, in two places:

1. **Enforcement:** `SquadUseCase` asks before every squad change and returns `ErrForbidden` (HTTP 403) on a "no".
2. **Display:** `domain.PermissionsFor` asks the same questions to build the `permissions` object on `GET /squads/:squadID`. The app shows a button only when the matching flag is true.

Because both use the same policy, a button the app shows never meets a 403.

---

## Reading `model.conf`

```ini
[request_definition]
r = role, act

[policy_definition]
p = role, act

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = r.role == p.role && r.act == p.act
```

| Line | Symbol | Meaning |
|---|---|---|
| `r = role, act` | `r` is the **request** | The question code asks: "may a member with this `role` take this `act`ion?" For example, `("MEMBER", "squad:edit")`. |
| `p = role, act` | `p` is a **policy rule** | The shape of each line in `policy.csv`: a role and an action it is allowed. |
| `e = some(where (p.eft == allow))` | `e` is the **effect** | How matching rules combine. `some(...)` means "allowed if at least one rule matches". `p.eft` is a rule's effect, which defaults to `allow`. There are no deny rules, so **anything not listed is denied**. |
| `m = r.role == p.role && r.act == p.act` | `m` is the **matcher** | When a rule matches a request: the role and the action must both be equal. Comparisons are exact and case-sensitive, so `admin` doesn't match `ADMIN`. |

## Reading `policy.csv`

```csv
p, ADMIN, squad:add_member
p, MEMBER, squad:add_member
p, ADMIN, squad:remove_member
p, ADMIN, squad:change_role
p, ADMIN, squad:edit
p, ADMIN, squad:leave
p, MEMBER, squad:leave
```

Each line is `p, <role>, <action>`. It reads as "a `<role>` may take `<action>`".

| Column | Meaning | Values |
|---|---|---|
| 1 | Line type. `p` means a policy rule (it matches `p` in `model.conf`). | always `p` |
| 2 | The member's role in the squad, as stored in `squad_members.role`. | `ADMIN`, `MEMBER` (`domain.Role`) |
| 3 | The action being attempted. | see [Actions](#actions) |

**Default deny:** if no line matches, the answer is "no". That includes an unknown role, an unknown action or a typo.

## Actions

Defined in `internal/domain/squad.go` as `domain.Action` constants.

| Action | Constant | Guards | `permissions` flag | Allowed today |
|---|---|---|---|---|
| `squad:add_member` | `ActionAddMember` | `POST /squads/:id/members` | `canAddMembers` | ADMIN, MEMBER |
| `squad:remove_member` | `ActionRemoveMember` | `DELETE /squads/:id/members/:userID` for someone else | `canRemoveMembers` | ADMIN |
| `squad:leave` | `ActionLeave` | `DELETE /squads/:id/members/:userID` with your own ID | (none: everyone sees "Leave") | ADMIN, MEMBER |
| `squad:change_role` | `ActionChangeRole` | `PATCH /squads/:id/members/:userID` | `canChangeRoles` | ADMIN |
| `squad:edit` | `ActionEditSquad` | `PATCH /squads/:id` | `canEditSquad` | ADMIN |

## What is decided outside the policy

The policy answers only "may this role take this action?". These rules are not permissions, so they stay in code:

| Rule | Where | Why it isn't in the policy |
|---|---|---|
| Not a member → 404 | `SquadRepository.GetMember` and `GetForMember` | Checked *before* the policy is asked: there is no role to check yet. It also hides squads you aren't in. |
| Removing yourself is `squad:leave` | `SquadUseCase.RemoveMember` | Same endpoint, two actions; the use case picks one by comparing IDs. |
| `canAddMembers` is false when the squad has `MaxSquadMembers` (50) | `domain.PermissionsFor` | A capacity limit, not a role rule. Adding to a full squad returns `ErrSquadFull` (409). |
| A squad always has an admin; the earliest joiner is promoted when the last admin leaves | `PostgresSquadRepository` (in a transaction) | An invariant on the data, enforced under concurrency. |
| A valid role is `ADMIN` or `MEMBER` | `domain.Role.IsValid` | Validates input to `PATCH .../members/:userID`. |

## Worked example: a MEMBER tries to remove someone

```
DELETE /api/v1/squads/s1/members/rohan        (caller: meera, X-User-ID)
  │
  ├─ SquadUseCase.RemoveMember
  │    ├─ GetMember(s1, meera)      → {role: MEMBER}       (not a member? → 404)
  │    ├─ rohan ≠ meera             → action = squad:remove_member
  │    └─ authz.Can(MEMBER, squad:remove_member)
  │         r = (MEMBER, squad:remove_member)
  │         no line in policy.csv has role MEMBER and that action
  │         → false
  │
  └─ ErrForbidden  →  403 "you are not allowed to do this in this squad"
```

If Meera had sent her own ID, the action would be `squad:leave`. The line `p, MEMBER, squad:leave` matches, so she leaves the squad.

## Changing permissions

Always update the table in `internal/adapters/authz/casbin_authorizer_test.go` in the same change. It checks every role and action pair, so an accidental grant or removal fails the build.

- **Let a role do something:** add a line, e.g. `p, MEMBER, squad:edit` to let every member rename the squad. Nothing else changes: the 403 stops, and `canEditSquad` becomes true for members.
- **Take something away:** delete the line.
- **Add a new action:**
  1. Add a `domain.Action` constant.
  2. Add policy lines.
  3. Call `u.authorize(actor, domain.ActionX)` in the use case method.
  4. If the app needs to know, add a flag to `domain.Permissions`, `PermissionsFor`, the DTO and `openapi.yaml`.
  5. Add a row to the [Actions](#actions) table.
- **Add a new role** (e.g. `MODERATOR`):
  1. Add a `domain.Role` constant and accept it in `Role.IsValid`.
  2. Give it policy lines.
  3. Add it to the `Role` enum in `openapi.yaml`.
