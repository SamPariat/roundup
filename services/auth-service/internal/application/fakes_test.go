package application

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

// fakes_test.go holds in-memory fakes for the auth ports, shared by
// otp_usecase_test.go and token_usecase_test.go.

const (
	testPhone  = "+919876543210"
	testUserID = "7dc92d84-1950-43d6-b676-debacf84d830"
)

// clock is a controllable time source injected into use cases as `now`.
type clock struct{ t time.Time }

func (c *clock) now() time.Time          { return c.t }
func (c *clock) advance(d time.Duration) { c.t = c.t.Add(d) }
func newClock() *clock                   { return &clock{t: time.Date(2026, 10, 15, 12, 0, 0, 0, time.UTC)} }

func newKeys(t *testing.T) (ed25519.PublicKey, ed25519.PrivateKey) {
	t.Helper()

	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	return pub, priv
}

// --- OTPRepository ---

type fakeOTPRepo struct {
	codes []*domain.OTPCode
	clock *clock
}

func (f *fakeOTPRepo) CountSince(_ context.Context, phone string, since time.Time) (int, error) {
	n := 0

	for _, c := range f.codes {
		if c.Phone == phone && !c.ExpiresAt.Add(-domain.OTPTTL).Before(since) {
			n++
		}
	}

	return n, nil
}

func (f *fakeOTPRepo) Create(_ context.Context, phone string, hash []byte, expiresAt time.Time) error {
	f.codes = append(f.codes, &domain.OTPCode{ID: int64(len(f.codes) + 1), Phone: phone, CodeHash: hash, ExpiresAt: expiresAt})

	return nil
}

func (f *fakeOTPRepo) Latest(_ context.Context, phone string) (*domain.OTPCode, error) {
	for _, v := range slices.Backward(f.codes) {
		if v.Phone == phone {
			c := *v

			return &c, nil
		}
	}

	return nil, domain.ErrInvalidCode
}

func (f *fakeOTPRepo) IncrementAttempts(_ context.Context, id int64) error {
	f.codes[id-1].Attempts++

	return nil
}

func (f *fakeOTPRepo) Consume(_ context.Context, id int64) error {
	if f.codes[id-1].ConsumedAt != nil {
		return domain.ErrInvalidCode
	}

	now := f.clock.now()
	f.codes[id-1].ConsumedAt = &now

	return nil
}

// --- OTPSender ---

type fakeSender struct{ sent map[string]string }

func (f *fakeSender) Send(_ context.Context, phone, code string) error {
	f.sent[phone] = code

	return nil
}

// --- UserDirectory ---

type fakeUsers struct {
	created bool
	err     error
	calls   int
}

func (f *fakeUsers) EnsureUser(context.Context, string) (string, bool, error) {
	f.calls++
	if f.err != nil {
		return "", false, f.err
	}

	return testUserID, f.created, nil
}

// --- SessionRepository ---

type fakeSessions struct {
	tokens  map[string]*domain.StoredRefreshToken // keyed by string(hash)
	revoked map[string]bool
	nextID  int64
}

func newFakeSessions() *fakeSessions {
	return &fakeSessions{tokens: map[string]*domain.StoredRefreshToken{}, revoked: map[string]bool{}}
}

func (f *fakeSessions) add(sessionID, userID string, hash []byte, expiresAt time.Time) {
	f.nextID++
	f.tokens[string(hash)] = &domain.StoredRefreshToken{ID: f.nextID, SessionID: sessionID, UserID: userID, ExpiresAt: expiresAt}
}

func (f *fakeSessions) Create(_ context.Context, userID string, hash []byte, expiresAt time.Time) (string, error) {
	sessionID := fmt.Sprintf("session-%d", len(f.revoked)+len(f.tokens)+1)
	f.add(sessionID, userID, hash, expiresAt)

	return sessionID, nil
}

func (f *fakeSessions) FindRefreshToken(_ context.Context, hash []byte) (*domain.StoredRefreshToken, error) {
	t, ok := f.tokens[string(hash)]
	if !ok {
		return nil, domain.ErrInvalidRefreshToken
	}

	c := *t
	c.SessionRevoked = f.revoked[t.SessionID]

	return &c, nil
}

func (f *fakeSessions) Rotate(_ context.Context, oldID int64, sessionID string, newHash []byte, expiresAt time.Time) error {
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

func (f *fakeSessions) Revoke(_ context.Context, sessionID string) error {
	f.revoked[sessionID] = true

	return nil
}
