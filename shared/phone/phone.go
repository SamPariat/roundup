// Package phone validates phone numbers. Shared so every service that accepts
// a phone number (auth-service at login, user-service at signup) agrees on
// what a valid one is.
package phone

import (
	"regexp"

	"github.com/SamPariat/roundup/shared/domainerr"
)

// ErrInvalid is the error for a number IsValid rejects. Services expose it as
// their own domain.ErrInvalidPhone, so clients get the same message from each.
var ErrInvalid = domainerr.Sentinel("INVALID_PHONE", "phone must be in E.164 format, e.g. +919876543210")

// e164 matches phone numbers in E.164 format: a leading +, then 8–15 digits.
var e164 = regexp.MustCompile(`^\+[1-9]\d{7,14}$`)

// IsValid reports whether p is in E.164 format, e.g. "+919876543210".
func IsValid(p string) bool {
	return e164.MatchString(p)
}
