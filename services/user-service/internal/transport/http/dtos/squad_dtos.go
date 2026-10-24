package dtos

import "time"

// CreateSquadRequest carries the JSON body for POST /api/v1/squads.
type CreateSquadRequest struct {
	// Name is the squad's display name. Trimmed; a blank name is rejected with 400.
	Name string `json:"name"`

	// AvatarURL is the squad picture URL. Optional.
	AvatarURL string `json:"avatarURL"`
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
}

// SquadsResponse lists the caller's squads.
type SquadsResponse struct {
	// Squads are newest first. Never null.
	Squads []SquadSummaryResponse `json:"squads"`
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

// SquadDetailResponse is a squad with its members.
type SquadDetailResponse struct {
	SquadResponse

	// Members are earliest joiner first. Never null.
	Members []SquadMemberResponse `json:"members"`
}
