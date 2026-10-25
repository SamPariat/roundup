package domain

// Role is one of the subjects the authorization policy grants actions to. A
// caller holds their squad role (ADMIN or MEMBER), plus CREATOR on events
// they created; an action is allowed if any of their roles allows it.
type Role string

// Roles named in the policy. What each may do is set in policy.csv (see
// docs/PERMISSIONS.md), not here.
const (
	RoleAdmin   Role = "ADMIN"
	RoleMember  Role = "MEMBER"
	RoleCreator Role = "CREATOR"
)

// Action is something a member may try to do with an event.
type Action string

// Event actions checked against the authorization policy.
const (
	// ActionCreate creates an event for the squad.
	ActionCreate Action = "event:create"

	// ActionRSVP answers going, maybe or not going.
	ActionRSVP Action = "event:rsvp"

	// ActionNominate suggests a place.
	ActionNominate Action = "event:nominate"

	// ActionVote votes for a suggested place.
	ActionVote Action = "event:vote"

	// ActionRemoveOwnNomination withdraws a place you suggested.
	ActionRemoveOwnNomination Action = "event:remove_own_nomination"

	// ActionRemoveNomination withdraws anyone's suggestion.
	ActionRemoveNomination Action = "event:remove_nomination"

	// ActionEdit changes the name, type or times.
	ActionEdit Action = "event:edit"

	// ActionManage moves the event along: start and close voting, start
	// the night, close it.
	ActionManage Action = "event:manage"
)

// Authorizer decides which roles may take which actions. Implemented by the
// shared Casbin enforcer loaded with policy.csv.
type Authorizer interface {
	Can(role Role, action Action) bool
}

// RolesOf returns the roles member holds for event: their squad role, plus
// CREATOR if they created it. event is nil for actions on the squad, such as
// creating an event.
func RolesOf(member SquadMember, event *Event) []Role {
	roles := []Role{Role(member.Role)}
	if event != nil && event.CreatedBy == member.UserID {
		roles = append(roles, RoleCreator)
	}

	return roles
}

// Allowed reports whether any of roles may take action.
func Allowed(authorizer Authorizer, roles []Role, action Action) bool {
	for _, role := range roles {
		if authorizer.Can(role, action) {
			return true
		}
	}

	return false
}
