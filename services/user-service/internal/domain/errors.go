package domain

import (
	"github.com/SamPariat/roundup/shared/domainerr"
	"github.com/SamPariat/roundup/shared/phone"
)

// Sentinel errors are used as stable identities for errors.Is comparisons.
// Callers should match against these values, never against error strings.
// Adapters add context with domainerr.New, which keeps errors.Is working.
var (
	// ErrUserNotFound is returned when no active user matches the lookup.
	ErrUserNotFound = domainerr.Sentinel("USER_NOT_FOUND", "user not found")

	// ErrPhoneTaken is returned when creating a user whose phone number already
	// belongs to an active user.
	ErrPhoneTaken = domainerr.Sentinel("PHONE_TAKEN", "phone number already registered")

	// ErrInvalidPhone is returned when a phone number is not in E.164 format.
	ErrInvalidPhone = phone.ErrInvalid

	// ErrInvalidName is returned when a profile update sets an empty name.
	ErrInvalidName = domainerr.Sentinel("INVALID_NAME", "name must not be empty")

	// ErrInvalidPlatform is returned when a device token's platform is not iOS or Android.
	ErrInvalidPlatform = domainerr.Sentinel("INVALID_PLATFORM", "platform must be iOS or Android")

	// ErrInvalidToken is returned when a device token is empty.
	ErrInvalidToken = domainerr.Sentinel("INVALID_DEVICE_TOKEN", "device token must not be empty")

	// ErrDeviceTokenNotFound is returned when removing a token the user does not own.
	ErrDeviceTokenNotFound = domainerr.Sentinel("DEVICE_TOKEN_NOT_FOUND", "device token not found")

	// ErrTooManyPhones is returned when a contact lookup sends more than
	// MaxLookupPhones numbers.
	ErrTooManyPhones = domainerr.Sentinel("TOO_MANY_PHONES", "too many phone numbers in one lookup")

	// ErrSquadNotFound is returned when the squad does not exist or the caller
	// is not a member. The two are deliberately indistinguishable.
	ErrSquadNotFound = domainerr.Sentinel("SQUAD_NOT_FOUND", "squad not found")

	// ErrInvalidSquadName is returned when a squad name is blank.
	ErrInvalidSquadName = domainerr.Sentinel("INVALID_SQUAD_NAME", "squad name must not be empty")

	// ErrAlreadyMember is returned when adding someone who is already in the squad.
	ErrAlreadyMember = domainerr.Sentinel("ALREADY_MEMBER", "user is already a member of this squad")

	// ErrSquadFull is returned when adding to a squad of MaxSquadMembers.
	ErrSquadFull = domainerr.Sentinel("SQUAD_FULL", "squad is full")

	// ErrMemberNotFound is returned when the target user is not in the squad.
	ErrMemberNotFound = domainerr.Sentinel("MEMBER_NOT_FOUND", "member not found")

	// ErrNotAdmin is returned when a member tries something only admins may do.
	ErrNotAdmin = domainerr.Sentinel("NOT_ADMIN", "only squad admins can do this")

	// ErrInvalidRole is returned when a role is not ADMIN or MEMBER.
	ErrInvalidRole = domainerr.Sentinel("INVALID_ROLE", "role must be ADMIN or MEMBER")

	// ErrLastAdmin is returned when a role change would leave a squad without
	// an admin.
	ErrLastAdmin = domainerr.Sentinel("LAST_ADMIN", "a squad needs at least one admin; make someone else admin first")
)
