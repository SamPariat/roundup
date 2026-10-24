package domain

import "time"

// MaxSquadMembers caps how many people a squad can hold.
const MaxSquadMembers = 50

// Role is what a member may do in a squad. Like a WhatsApp group, a squad can
// have several admins.
type Role string

// Supported squad roles.
const (
	// RoleAdmin can remove members and change roles.
	RoleAdmin Role = "ADMIN"

	// RoleMember can add people and leave.
	RoleMember Role = "MEMBER"
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

// SquadDetail is a squad with its members, ordered by when they joined.
type SquadDetail struct {
	Squad

	// Members are the squad's members, earliest joiner first.
	Members []SquadMember
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

	// ActorID is the admin making the change.
	ActorID string

	// UserID is the member whose role changes.
	UserID string

	// Role is the new role.
	Role Role
}

// CanRemove reports whether actor may remove the member targetUserID from the
// squad. Admins may remove anyone, themselves included; members may only
// remove themselves (leave). Returns ErrNotAdmin otherwise.
func CanRemove(actor SquadMember, targetUserID string) error {
	if actor.UserID == targetUserID || actor.Role == RoleAdmin {
		return nil
	}

	return ErrNotAdmin
}

// CanChangeRole reports whether actor may change members' roles: admins only.
func CanChangeRole(actor SquadMember) error {
	if actor.Role != RoleAdmin {
		return ErrNotAdmin
	}

	return nil
}
