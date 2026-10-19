package domain

import "time"

// OTP rules. Codes are short-lived and rate-limited per phone number, so a
// six-digit code is safe even though it is easy to type.
const (
	// OTPLength is the number of digits in a code.
	OTPLength = 6

	// OTPTTL is how long a code stays valid after it is sent.
	OTPTTL = 5 * time.Minute

	// OTPRequestWindow and MaxOTPRequestsPerWindow limit how many codes one
	// phone number can request, which also caps SMS spend.
	OTPRequestWindow        = 15 * time.Minute
	MaxOTPRequestsPerWindow = 3

	// MaxOTPAttempts is how many wrong guesses one code allows. With the
	// request limit, that caps guessing at 15 tries per 15 minutes per phone.
	MaxOTPAttempts = 5
)

// OTPCode is a one-time login code sent to a phone number. Only the newest
// code for a phone is valid.
type OTPCode struct {
	// ID is the database-assigned primary key.
	ID int64

	// Phone is the E.164 number the code was sent to.
	Phone string

	// CodeHash is the hash of the code; the code itself is never stored.
	CodeHash []byte

	// Attempts counts wrong guesses against this code.
	Attempts int

	// ExpiresAt is when the code stops being accepted.
	ExpiresAt time.Time

	// ConsumedAt is set once the code has been used to log in; nil until then.
	ConsumedAt *time.Time
}
