// Package domaintest holds in-memory fakes of the auth ports, so the
// application tests and the HTTP handler tests drive the real use cases
// against the same fakes instead of each package keeping its own.
package domaintest

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"fmt"
	"slices"
	"testing"
	"time"

	"github.com/SamPariat/roundup/services/auth-service/internal/domain"
	"github.com/stretchr/testify/require"
)

const (
	// Phone is the number every test logs in with.
	Phone = "+919876543210"

	// UserID is the user the fake directory returns for Phone.
	UserID = "7dc92d84-1950-43d6-b676-debacf84d830"
)

// Clock is a controllable time source injected into use cases as `now`.
type Clock struct{ T time.Time }

// Now returns the clock's current time.
func (c *Clock) Now() time.Time { return c.T }

// Advance moves the clock forward by d.
func (c *Clock) Advance(d time.Duration) { c.T = c.T.Add(d) }

// NewClock returns a clock set to a fixed instant.
func NewClock() *Clock { return &Clock{T: time.Date(2026, 10, 15, 12, 0, 0, 0, time.UTC)} }

// NewKeys returns a fresh Ed25519 key pair for signing test tokens.
func NewKeys(t *testing.T) (ed25519.PublicKey, ed25519.PrivateKey) {
	t.Helper()

	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	return pub, priv
}

// --- OTPRepository ---

// OTPRepo is an in-memory domain.OTPRepository. Clock stamps consumed codes.
type OTPRepo struct {
	Codes []*domain.OTPCode
	Clock *Clock
}

// CountSince counts the codes for phone created at or after since.
func (f *OTPRepo) CountSince(_ context.Context, phone string, since time.Time) (int, error) {
	n := 0

	for _, c := range f.Codes {
		if c.Phone == phone && !c.ExpiresAt.Add(-domain.OTPTTL).Before(since) {
			n++
		}
	}

	return n, nil
}

// Create stores a new code hash for phone.
func (f *OTPRepo) Create(_ context.Context, phone string, hash []byte, expiresAt time.Time) error {
	f.Codes = append(f.Codes, &domain.OTPCode{ID: int64(len(f.Codes) + 1), Phone: phone, CodeHash: hash, ExpiresAt: expiresAt})

	return nil
}

// Latest returns a copy of the newest code for phone.
func (f *OTPRepo) Latest(_ context.Context, phone string) (*domain.OTPCode, error) {
	for _, v := range slices.Backward(f.Codes) {
		if v.Phone == phone {
			c := *v

			return &c, nil
		}
	}

	return nil, domain.ErrInvalidCode
}

// IncrementAttempts counts one wrong guess on the code.
func (f *OTPRepo) IncrementAttempts(_ context.Context, id int64) error {
	f.Codes[id-1].Attempts++

	return nil
}

// Consume marks the code used, once.
func (f *OTPRepo) Consume(_ context.Context, id int64) error {
	if f.Codes[id-1].ConsumedAt != nil {
		return domain.ErrInvalidCode
	}

	now := f.Clock.Now()
	f.Codes[id-1].ConsumedAt = &now

	return nil
}

// --- OTPSender ---

// OTPSender records the last code sent to each phone instead of sending it.
type OTPSender struct{ Sent map[string]string }

// NewOTPSender returns a sender with nothing sent yet.
func NewOTPSender() *OTPSender { return &OTPSender{Sent: map[string]string{}} }

// Send records code as sent to phone.
func (f *OTPSender) Send(_ context.Context, phone, code string) error {
	f.Sent[phone] = code

	return nil
}

// --- UserDirectory ---

// UserDirectory always returns UserID, reporting Created, unless Err is set.
type UserDirectory struct {
	Created bool
	Err     error
	Calls   int
}

// EnsureUser returns UserID and Created, or Err.
func (f *UserDirectory) EnsureUser(context.Context, string) (string, bool, error) {
	f.Calls++
	if f.Err != nil {
		return "", false, f.Err
	}

	return UserID, f.Created, nil
}

// --- SessionRepository ---

// SessionRepo is an in-memory domain.SessionRepository.
type SessionRepo struct {
	tokens  map[string]*domain.StoredRefreshToken // keyed by string(hash)
	revoked map[string]bool
	nextID  int64
}

// NewSessionRepo returns a repository with no sessions.
func NewSessionRepo() *SessionRepo {
	return &SessionRepo{tokens: map[string]*domain.StoredRefreshToken{}, revoked: map[string]bool{}}
}

func (f *SessionRepo) add(sessionID, userID string, hash []byte, expiresAt time.Time) {
	f.nextID++
	f.tokens[string(hash)] = &domain.StoredRefreshToken{ID: f.nextID, SessionID: sessionID, UserID: userID, ExpiresAt: expiresAt}
}

// Create starts a session with its first refresh token.
func (f *SessionRepo) Create(_ context.Context, userID string, hash []byte, expiresAt time.Time) (string, error) {
	sessionID := fmt.Sprintf("session-%d", len(f.revoked)+len(f.tokens)+1)
	f.add(sessionID, userID, hash, expiresAt)

	return sessionID, nil
}

// FindRefreshToken returns a copy of the token with this hash.
func (f *SessionRepo) FindRefreshToken(_ context.Context, hash []byte) (*domain.StoredRefreshToken, error) {
	t, ok := f.tokens[string(hash)]
	if !ok {
		return nil, domain.ErrInvalidRefreshToken
	}

	c := *t
	c.SessionRevoked = f.revoked[t.SessionID]

	return &c, nil
}

// Rotate marks the old token used and adds the new one to the same session.
func (f *SessionRepo) Rotate(_ context.Context, oldID int64, sessionID string, newHash []byte, expiresAt time.Time) error {
	for _, t := range f.tokens {
		if t.ID == oldID {
			if t.UsedAt != nil {
				return domain.ErrInvalidRefreshToken
			}

			now := time.Now()
			t.UsedAt = &now
			f.add(sessionID, t.UserID, newHash, expiresAt)

			return nil
		}
	}

	return domain.ErrInvalidRefreshToken
}

// Revoke ends the session.
func (f *SessionRepo) Revoke(_ context.Context, sessionID string) error {
	f.revoked[sessionID] = true

	return nil
}
