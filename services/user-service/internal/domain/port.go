package domain

import "context"

// UserRepository abstracts persistence of user profiles. All lookups ignore
// soft-deleted users. Implemented by adapters/persistence.
type UserRepository interface {
	// Create inserts a new user. Returns ErrPhoneTaken if an active user
	// already has this phone number.
	Create(ctx context.Context, phone string) (*User, error)
	// GetByID returns ErrUserNotFound if no active user has this ID.
	GetByID(ctx context.Context, id string) (*User, error)
	// GetByPhone returns ErrUserNotFound if no active user has this phone.
	GetByPhone(ctx context.Context, phone string) (*User, error)
	// UpdateProfile returns ErrUserNotFound if no active user has this ID.
	UpdateProfile(ctx context.Context, cmd UpdateProfileCommand) (*User, error)
	// SoftDelete returns ErrUserNotFound if no active user has this ID.
	SoftDelete(ctx context.Context, id string) error
}

// DeviceTokenRepository abstracts persistence of FCM device tokens.
type DeviceTokenRepository interface {
	// Upsert registers the token to the user, moving it from any previous owner.
	Upsert(ctx context.Context, cmd RegisterDeviceTokenCommand) (*DeviceToken, error)
	// Delete returns ErrDeviceTokenNotFound if the user does not own the token.
	Delete(ctx context.Context, cmd RemoveDeviceTokenCommand) error
	// DeleteAllForUser removes every token registered to the user.
	DeleteAllForUser(ctx context.Context, userID string) error
}
