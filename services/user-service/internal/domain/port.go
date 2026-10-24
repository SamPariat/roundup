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

	// ListByPhones returns the active users with any of these phone numbers,
	// in no particular order. Numbers with no user are skipped.
	ListByPhones(ctx context.Context, phones []string) ([]*User, error)
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

// SquadRepository abstracts persistence of squads and their members. Methods
// that change membership run in a transaction that locks the squad, so
// concurrent changes to one squad are applied one at a time.
type SquadRepository interface {
	// Create inserts the squad and makes its creator the first admin.
	Create(ctx context.Context, cmd CreateSquadCommand) (*Squad, error)

	// GetForMember returns the squad if userID is a member, and
	// ErrSquadNotFound otherwise.
	GetForMember(ctx context.Context, squadID, userID string) (*Squad, error)

	// ListForUser returns the squads userID belongs to, newest first.
	ListForUser(ctx context.Context, userID string) ([]SquadSummary, error)

	// ListMembers returns the squad's active members, earliest joiner first.
	ListMembers(ctx context.Context, squadID string) ([]SquadMember, error)

	// GetMember returns userID's membership, or ErrSquadNotFound if they are
	// not in the squad.
	GetMember(ctx context.Context, squadID, userID string) (*SquadMember, error)

	// AddMember adds userID as a member. Returns ErrSquadFull if the squad
	// already has limit members, ErrAlreadyMember if they are in it, and
	// ErrSquadNotFound if the squad does not exist.
	AddMember(ctx context.Context, squadID, userID string, limit int) (*SquadMember, error)

	// RemoveMember removes userID. If no admin remains, the earliest joiner is
	// promoted; if no member remains, the squad is deleted and squadDeleted is
	// true. Returns ErrMemberNotFound if userID is not in the squad.
	RemoveMember(ctx context.Context, squadID, userID string) (squadDeleted bool, err error)

	// ChangeRole sets userID's role. Returns ErrMemberNotFound if they are not
	// in the squad and ErrLastAdmin if the squad would be left without an admin.
	ChangeRole(ctx context.Context, squadID, userID string, role Role) error
}
