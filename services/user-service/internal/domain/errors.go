package domain

import (
	"errors"

	"github.com/SamPariat/roundup/shared/phone"
)

// Sentinel errors are used as stable identities for errors.Is comparisons.
// Callers should match against these values, never against error strings.
// Adapters add context with domainerr.New, which keeps errors.Is working.
var (
	// ErrUserNotFound is returned when no active user matches the lookup.
	ErrUserNotFound = errors.New("user not found")

	// ErrPhoneTaken is returned when creating a user whose phone number already
	// belongs to an active user.
	ErrPhoneTaken = errors.New("phone number already registered")

	// ErrInvalidPhone is returned when a phone number is not in E.164 format.
	ErrInvalidPhone = phone.ErrInvalid

	// ErrInvalidName is returned when a profile update sets an empty name.
	ErrInvalidName = errors.New("name must not be empty")

	// ErrInvalidPlatform is returned when a device token's platform is not iOS or Android.
	ErrInvalidPlatform = errors.New("platform must be iOS or Android")

	// ErrInvalidToken is returned when a device token is empty.
	ErrInvalidToken = errors.New("device token must not be empty")

	// ErrDeviceTokenNotFound is returned when removing a token the user does not own.
	ErrDeviceTokenNotFound = errors.New("device token not found")

	// ErrTooManyPhones is returned when a contact lookup sends more than
	// MaxLookupPhones numbers.
	ErrTooManyPhones = errors.New("too many phone numbers in one lookup")

	// ErrSquadNotFound is returned when the squad does not exist or the caller
	// is not a member. The two are deliberately indistinguishable.
	ErrSquadNotFound = errors.New("squad not found")

	// ErrInvalidSquadName is returned when a squad name is blank.
	ErrInvalidSquadName = errors.New("squad name must not be empty")

	// ErrAlreadyMember is returned when adding someone who is already in the squad.
	ErrAlreadyMember = errors.New("user is already a member of this squad")

	// ErrSquadFull is returned when adding to a squad of MaxSquadMembers.
	ErrSquadFull = errors.New("squad is full")

	// ErrMemberNotFound is returned when the target user is not in the squad.
	ErrMemberNotFound = errors.New("member not found")

	// ErrNotAdmin is returned when a member tries something only admins may do.
	ErrNotAdmin = errors.New("only squad admins can do this")

	// ErrInvalidRole is returned when a role is not ADMIN or MEMBER.
	ErrInvalidRole = errors.New("role must be ADMIN or MEMBER")

	// ErrLastAdmin is returned when a role change would leave a squad without
	// an admin.
	ErrLastAdmin = errors.New("a squad needs at least one admin; make someone else admin first")
)
