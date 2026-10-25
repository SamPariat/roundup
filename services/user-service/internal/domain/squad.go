package domain

import "time"

// MaxSquadMembers caps how many people a squad can hold.
const MaxSquadMembers = 50

// MemberPreviewSize is how many members a squad list entry previews.
const MemberPreviewSize = 4

// Role is what a member may do in a squad. Like a WhatsApp group, a squad can
// have several admins.
type Role string

// Supported squad roles. What each may do is set in the authorization
// policy (see docs/PERMISSIONS.md), not here.
const (
	// RoleAdmin is a squad admin; the creator is the first.
	RoleAdmin Role = "ADMIN"

	// RoleMember is a regular member.
	RoleMember Role = "MEMBER"
)

// Action is something a member may try to do in a squad. Authorizer decides
// which roles may take each action.
type Action string

// Squad actions checked against the authorization policy.
const (
	// ActionAddMember adds a Roundup user to the squad.
	ActionAddMember Action = "squad:add_member"

	// ActionRemoveMember removes someone other than yourself.
	ActionRemoveMember Action = "squad:remove_member"

	// ActionLeave removes yourself from the squad.
	ActionLeave Action = "squad:leave"

	// ActionChangeRole promotes or demotes a member.
	ActionChangeRole Action = "squad:change_role"

	// ActionEditSquad renames the squad or changes its picture.
	ActionEditSquad Action = "squad:edit"
)

// IsValid reports whether r is a supported role.
func (r Role) IsValid() bool {
	return r == RoleAdmin || r == RoleMember
}

// Squad is a standing group of friends who go out together.
type Squad struct {
	// ID is the squad's UUID.
	ID string

	// Name is the display name. Never empty.
	Name string

	// AvatarURL is the squad picture, or empty if unset.
	AvatarURL string

	// CreatedBy is the user who created the squad. They may since have left.
	CreatedBy string

	// CreatedAt is the UTC timestamp when the squad was created.
	CreatedAt time.Time
}

// SquadSummary is a squad as it appears in a user's squad list.
type SquadSummary struct {
	Squad

	// Role is the listing user's role in the squad.
	Role Role

	// MemberCount is the number of members, the listing user included.
	MemberCount int

	// MemberPreview is the first MemberPreviewSize members, earliest joiner first.
	MemberPreview []SquadMember
}

// SquadMember is a user's membership in a squad.
type SquadMember struct {
	// UserID is the member's user ID.
	UserID string

	// Name is the member's profile name, or empty if unset.
	Name string

	// AvatarURL is the member's profile picture, or empty if unset.
	AvatarURL string

	// Role is what the member may do in the squad.
	Role Role

	// JoinedAt is the UTC timestamp when the member joined. The earliest
	// joiner becomes admin when the last admin leaves.
	JoinedAt time.Time
}

// SquadDetail is a squad with its members, as the member Me sees it.
type SquadDetail struct {
	Squad

	// Members are the squad's members, earliest joiner first.
	Members []SquadMember

	// Me is the caller's own membership.
	Me SquadMember

	// Permissions are what Me may do in the squad.
	Permissions Permissions
}

// Permissions says which squad actions a member may take, so clients show
// only the controls that will work. Built by PermissionsFor from the same
// policy the use case enforces.
type Permissions struct {
	// CanAddMembers is true while the squad has room; any member may add.
	CanAddMembers bool

	// CanRemoveMembers is whether they may remove someone other than themselves.
	CanRemoveMembers bool

	// CanChangeRoles is whether they may promote or demote members.
	CanChangeRoles bool

	// CanEditSquad is whether they may rename the squad or change its picture.
	CanEditSquad bool
}

// CreateSquadCommand carries the data needed to create a squad.
type CreateSquadCommand struct {
	// CreatorID is the user creating the squad. They become its first admin.
	CreatorID string

	// Name is the squad's display name. Trimmed; must not be blank.
	Name string

	// AvatarURL is the squad picture, or empty for none.
	AvatarURL string
}

// AddMemberCommand carries the data needed to add someone to a squad.
type AddMemberCommand struct {
	// SquadID is the squad being added to.
	SquadID string

	// ActorID is the member doing the adding.
	ActorID string

	// Phone is the E.164 phone number of the Roundup user to add.
	Phone string
}

// UpdateSquadCommand carries the changes to a squad's name and picture. A nil
// field is left unchanged.
type UpdateSquadCommand struct {
	// SquadID is the squad being edited.
	SquadID string

	// ActorID is the member making the change.
	ActorID string

	// Name is the new display name. Trimmed; must not be blank.
	Name *string

	// AvatarURL is the new picture URL; empty removes the picture.
	AvatarURL *string
}

// RemoveMemberCommand carries the data needed to remove a member. Leaving a
// squad is removing yourself: ActorID == UserID.
type RemoveMemberCommand struct {
	// SquadID is the squad being removed from.
	SquadID string

	// ActorID is the member doing the removing.
	ActorID string

	// UserID is the member being removed.
	UserID string
}

// ChangeRoleCommand carries the data needed to change a member's role.
type ChangeRoleCommand struct {
	// SquadID is the squad the member belongs to.
	SquadID string

	// ActorID is the member making the change.
	ActorID string

	// UserID is the member whose role changes.
	UserID string

	// Role is the new role.
	Role Role
}

// PermissionsFor returns what actor may do in a squad of memberCount members.
// Which role may take which action is decided by authz (the policy), never
// here; only the member cap, a capacity limit, is checked in code.
func PermissionsFor(authz Authorizer, actor SquadMember, memberCount int) Permissions {
	return Permissions{
		CanAddMembers:    memberCount < MaxSquadMembers && authz.Can(actor.Role, ActionAddMember),
		CanRemoveMembers: authz.Can(actor.Role, ActionRemoveMember),
		CanChangeRoles:   authz.Can(actor.Role, ActionChangeRole),
		CanEditSquad:     authz.Can(actor.Role, ActionEditSquad),
	}
}
