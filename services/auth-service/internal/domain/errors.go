package domain

import (
	"errors"

	"github.com/SamPariat/roundup/shared/phone"
)

// Sentinel errors are used as stable identities for errors.Is comparisons.
// Callers should match against these values, never against error strings.
// Adapters add context with domainerr.New, which keeps errors.Is working.
var (
	// ErrInvalidPhone is returned when a phone number is not in E.164 format.
	ErrInvalidPhone = phone.ErrInvalid

	// ErrTooManyRequests is returned when a phone has requested too many codes
	// within the rate-limit window.
	ErrTooManyRequests = errors.New("too many code requests, try again later")

	// ErrInvalidCode is returned when a code is wrong, expired, already used,
	// or was never requested. One error for all cases avoids telling an
	// attacker which part was wrong.
	ErrInvalidCode = errors.New("invalid or expired code")

	// ErrTooManyAttempts is returned when the current code has been guessed
	// wrongly too many times. The user must request a new code.
	ErrTooManyAttempts = errors.New("too many wrong attempts, request a new code")

	// ErrUserServiceUnavailable is returned when user-service cannot be reached
	// or answers unexpectedly during login. The code is left unused so the
	// user can retry.
	ErrUserServiceUnavailable = errors.New("user service unavailable")

	// ErrInvalidRefreshToken is returned for a refresh token that is unknown,
	// expired, already used, or belongs to a revoked session.
	ErrInvalidRefreshToken = errors.New("invalid refresh token")
)
