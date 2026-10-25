package dtos

import "time"

// CreateSquadRequest carries the JSON body for POST /api/v1/squads.
type CreateSquadRequest struct {
	// Name is the squad's display name. Trimmed; a blank name is rejected with 400.
	Name string `json:"name"`

	// AvatarURL is the squad picture URL. Optional.
	AvatarURL string `json:"avatarURL"`
}

// UpdateSquadRequest carries the JSON body for PATCH /api/v1/squads/:squadID.
// Omitted fields are left unchanged.
type UpdateSquadRequest struct {
	// Name is the new display name. Trimmed; a blank name is rejected with 400.
	Name *string `json:"name"`

	// AvatarURL is the new picture URL; "" removes the picture.
	AvatarURL *string `json:"avatarURL"`
}

// AddMemberRequest carries the JSON body for POST /api/v1/squads/:squadID/members.
type AddMemberRequest struct {
	// Phone is the E.164 number of the Roundup user to add.
	Phone string `json:"phone"`
}

// ChangeRoleRequest carries the JSON body for
// PATCH /api/v1/squads/:squadID/members/:userID.
type ChangeRoleRequest struct {
	// Role is "ADMIN" or "MEMBER".
	Role string `json:"role"`
}

// SquadResponse is the JSON representation of a domain.Squad.
type SquadResponse struct {
	// ID is the squad's UUID.
	ID string `json:"id"`

	// Name is the squad's display name.
	Name string `json:"name"`

	// AvatarURL is "" when no picture is set.
	AvatarURL string `json:"avatarURL"`

	// CreatedBy is the user who created the squad.
	CreatedBy string `json:"createdBy"`

	// CreatedAt is when the squad was created, as an RFC 3339 timestamp.
	CreatedAt time.Time `json:"createdAt"`
}

// SquadSummaryResponse is a squad in the caller's squad list.
type SquadSummaryResponse struct {
	SquadResponse

	// Role is the caller's role in the squad.
	Role string `json:"role"`

	// MemberCount is the number of members, the caller included.
	MemberCount int `json:"memberCount"`

	// MemberPreview is the first 4 members, earliest joiner first. Never null.
	MemberPreview []SquadMemberResponse `json:"memberPreview"`
}

// SquadsResponse lists the caller's squads.
type SquadsResponse struct {
	// Squads are newest first. Never null.
	Squads []SquadSummaryResponse `json:"squads"`

	// NextCursor fetches the next page as ?cursor=; null on the last page.
	NextCursor *string `json:"nextCursor"`
}

// SquadMemberResponse is the JSON representation of a domain.SquadMember.
type SquadMemberResponse struct {
	// UserID is the member's user ID.
	UserID string `json:"userID"`

	// Name is "" until the member completes profile setup.
	Name string `json:"name"`

	// AvatarURL is "" when no avatar is set.
	AvatarURL string `json:"avatarURL"`

	// Role is "ADMIN" or "MEMBER".
	Role string `json:"role"`

	// JoinedAt is when the member joined, as an RFC 3339 timestamp.
	JoinedAt time.Time `json:"joinedAt"`
}

// SquadDetailMemberResponse is a member as listed on the squad detail.
type SquadDetailMemberResponse struct {
	SquadMemberResponse

	// IsMe is true for the caller's own entry.
	IsMe bool `json:"isMe"`
}

// PermissionsResponse says which squad actions the caller may take.
type PermissionsResponse struct {
	// CanAddMembers is true while the squad has room.
	CanAddMembers bool `json:"canAddMembers"`

	// CanRemoveMembers is whether the caller may remove other members.
	CanRemoveMembers bool `json:"canRemoveMembers"`

	// CanChangeRoles is whether the caller may promote or demote members.
	CanChangeRoles bool `json:"canChangeRoles"`

	// CanEditSquad is whether the caller may rename the squad or change its picture.
	CanEditSquad bool `json:"canEditSquad"`
}

// SquadDetailResponse is a squad with its members, as the caller sees it.
type SquadDetailResponse struct {
	SquadResponse

	// MyRole is the caller's role in the squad.
	MyRole string `json:"myRole"`

	// MemberCount is the number of members, the caller included.
	MemberCount int `json:"memberCount"`

	// CreatedByName is the creator's profile name; null if they have left
	// the squad, "" if they have not set a name.
	CreatedByName *string `json:"createdByName"`

	// Members are earliest joiner first. Never null.
	Members []SquadDetailMemberResponse `json:"members"`

	// Permissions are what the caller may do in the squad.
	Permissions PermissionsResponse `json:"permissions"`
}
