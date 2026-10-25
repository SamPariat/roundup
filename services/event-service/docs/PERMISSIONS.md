# Event permissions

Who may do what with an event is decided by a policy, not by Go code, the same way as squads in user-service. How to read the shared Casbin model is explained in [user-service's PERMISSIONS.md](../../user-service/docs/PERMISSIONS.md#reading-modelconf).

| File | What it holds |
|---|---|
| [`shared/authz/model.conf`](../../../shared/authz/model.conf) | The shape of a permission check. Shared by every service. |
| [`internal/policy/policy.csv`](../internal/policy/policy.csv) | The rules: which role may take which action. This is the file you edit. |

## Roles

A caller holds **more than one role** for an event:

- their **squad role**, `ADMIN` or `MEMBER`, from user-service;
- **`CREATOR`**, on the events they created.

An action is allowed if **any** of their roles allows it (`domain.Allowed`). That is how the creator can manage their own event without being a squad admin.

## Rules

| Action | `MEMBER` | `ADMIN` | `CREATOR` |
|---|---|---|---|
| `event:create` | ✅ | ✅ | |
| `event:rsvp` | ✅ | ✅ | |
| `event:nominate` | ✅ | ✅ | |
| `event:vote` | ✅ | ✅ | |
| `event:remove_own_nomination` | ✅ | ✅ | |
| `event:remove_nomination` (anyone's) | | ✅ | ✅ |
| `event:edit` | | ✅ | ✅ |
| `event:manage` (start/close voting, start/close the night) | | ✅ | ✅ |

`CREATOR` lists only what it adds: a creator is always also a member.

## Policy and status together

The policy says *who* may act; the event's status says *when*. Both are checked:

1. **Enforcement:** every use case checks the policy (`ErrForbidden`, 403), then the status (`ErrInvalidTransition`, 409).
2. **Display:** `GET /events/:eventID` returns `permissions` and each nomination's `canRemove`, computed from the same policy and status, so a button the app shows never meets a 403 or 409.

## Changing the rules

Edit `policy.csv`, then update the table in `internal/policy/policy_test.go` and this page. `TestPolicyLinesAreWellFormed` fails on a line that isn't `p, <known role>, <known action>`. Casbin itself accepts a short line such as `p, ADMIN` silently, which would deny without any error.
