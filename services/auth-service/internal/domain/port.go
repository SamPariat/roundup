package domain

import (
	"context"
	"time"
)

// OTPRepository stores one-time login codes. Implemented by adapters/persistence.
type OTPRepository interface {
	// CountSince returns how many codes were created for phone at or after since.
	CountSince(ctx context.Context, phone string, since time.Time) (int, error)
	// Create stores a new code hash for phone.
	Create(ctx context.Context, phone string, codeHash []byte, expiresAt time.Time) error
	// Latest returns the newest code for phone, consumed or not.
	// Returns ErrInvalidCode if the phone has never requested one.
	Latest(ctx context.Context, phone string) (*OTPCode, error)
	// IncrementAttempts records one wrong guess against the code.
	IncrementAttempts(ctx context.Context, id int64) error
	// Consume marks the code used. Returns ErrInvalidCode if it already was,
	// so two concurrent logins with one code cannot both succeed.
	Consume(ctx context.Context, id int64) error
}

// SessionRepository stores login sessions and their refresh tokens.
type SessionRepository interface {
	// Create starts a session for userID with its first refresh token and
	// returns the session ID.
	Create(ctx context.Context, userID string, tokenHash []byte, expiresAt time.Time) (string, error)
	// FindRefreshToken returns the token with this hash.
	// Returns ErrInvalidRefreshToken if there is none.
	FindRefreshToken(ctx context.Context, tokenHash []byte) (*StoredRefreshToken, error)
	// Rotate marks token oldID used and adds a new token to the same session,
	// atomically. Returns ErrInvalidRefreshToken if oldID was already used.
	Rotate(ctx context.Context, oldID int64, sessionID string, newHash []byte, expiresAt time.Time) error
	// Revoke ends the session; none of its refresh tokens work afterwards.
	Revoke(ctx context.Context, sessionID string) error
}

// OTPSender delivers a code to a phone, by SMS in production.
type OTPSender interface {
	Send(ctx context.Context, phone, code string) error
}

// UserDirectory finds or creates the user for a verified phone number.
// Implemented by a client for user-service.
type UserDirectory interface {
	// EnsureUser returns the user ID for phone, creating the user if needed.
	// created reports whether a new user was made. Returns
	// ErrUserServiceUnavailable if user-service cannot answer.
	EnsureUser(ctx context.Context, phone string) (userID string, created bool, err error)
}
