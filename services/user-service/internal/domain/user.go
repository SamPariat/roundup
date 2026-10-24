package domain

import "time"

// MaxLookupPhones caps how many phone numbers one contact lookup may send.
const MaxLookupPhones = 500

// User is a Roundup user's profile. Login credentials live in auth-service;
// both services share the same ID.
type User struct {
	// ID is the user's UUID, shared with auth-service.
	ID string

	// Phone is the user's phone number in E.164 format (e.g. +919876543210).
	Phone string

	// Name is empty until the user completes profile setup.
	Name string

	// AvatarURL is empty until the user uploads an avatar.
	AvatarURL string

	// CreatedAt is the UTC timestamp when the user was created.
	CreatedAt time.Time

	// UpdatedAt is the UTC timestamp of the last profile change.
	UpdatedAt time.Time
}

// UpdateProfileCommand carries a partial profile update. Nil fields are left unchanged.
type UpdateProfileCommand struct {
	// UserID is the user being updated.
	UserID string

	// Name replaces the display name when non-nil.
	Name *string

	// AvatarURL replaces the avatar URL when non-nil.
	AvatarURL *string
}
