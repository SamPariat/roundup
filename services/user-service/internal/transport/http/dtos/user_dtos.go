package dtos

import "time"

// CreateUserRequest carries the JSON body for POST /api/v1/users.
type CreateUserRequest struct {
	// Phone is the verified phone number in E.164 format, e.g. "+919876543210".
	Phone string `json:"phone"`
}

// UpdateProfileRequest carries the JSON body for PATCH /api/v1/users/me.
// Omitted or null fields are left unchanged.
type UpdateProfileRequest struct {
	// Name is the new display name. Surrounding whitespace is trimmed; a blank
	// name is rejected with 400.
	Name *string `json:"name"`

	// AvatarURL is the new avatar image URL. An empty string clears it.
	AvatarURL *string `json:"avatarURL"`
}

// UserResponse is the JSON representation of a domain.User.
type UserResponse struct {
	// ID is the user's UUID, shared with auth-service.
	ID string `json:"id"`

	// Phone is the phone number in E.164 format.
	Phone string `json:"phone"`

	// Name is "" until the user completes profile setup.
	Name string `json:"name"`

	// AvatarURL is "" when no avatar is set.
	AvatarURL string `json:"avatarURL"`

	// CreatedAt is when the user signed up, as an RFC 3339 timestamp.
	CreatedAt time.Time `json:"createdAt"`

	// UpdatedAt is when the profile last changed, as an RFC 3339 timestamp.
	UpdatedAt time.Time `json:"updatedAt"`
}
