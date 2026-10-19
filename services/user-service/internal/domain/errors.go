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
)
