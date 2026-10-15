package domain

import "time"

// StoredRefreshToken is a refresh token record joined with its session, as
// needed to decide whether the token may be used.
type StoredRefreshToken struct {
	// ID is the database-assigned primary key of the token record.
	ID int64
	// SessionID is the login session the token belongs to.
	SessionID string
	// UserID is the user who owns the session.
	UserID string
	// ExpiresAt is when the token stops being accepted.
	ExpiresAt time.Time
	// UsedAt is set once the token has been exchanged; nil while unused.
	UsedAt *time.Time
	// SessionRevoked reports whether the session was logged out or revoked.
	SessionRevoked bool
}

// TokenPair is what a client receives at login and on every refresh.
type TokenPair struct {
	// AccessToken is the short-lived JWT sent as "Authorization: Bearer".
	AccessToken string
	// AccessTokenExpiresAt is when the access token expires.
	AccessTokenExpiresAt time.Time
	// RefreshToken is the long-lived opaque token used to get a new pair.
	RefreshToken string
	// RefreshTokenExpiresAt is when the refresh token expires.
	RefreshTokenExpiresAt time.Time
}

// LoginResult is returned by a successful OTP verification.
type LoginResult struct {
	TokenPair
	// UserID is the logged-in user's UUID.
	UserID string
	// IsNewUser is true when this login created the account, so the app can
	// show profile setup.
	IsNewUser bool
}
