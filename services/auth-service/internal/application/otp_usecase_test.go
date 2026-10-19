package application

import (
	"context"
	"testing"
	"time"

	"github.com/SamPariat/roundup/services/auth-service/internal/domain"
	"github.com/SamPariat/roundup/shared/authtoken"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type otpFixture struct {
	uc       *OTPUseCase
	otps     *fakeOTPRepo
	sender   *fakeSender
	users    *fakeUsers
	sessions *fakeSessions
	clock    *clock
	pub      []byte
}

func newOTPFixture(t *testing.T, devCode string) *otpFixture {
	t.Helper()
	clk := newClock()
	pub, priv := newKeys(t)
	f := &otpFixture{
		otps:     &fakeOTPRepo{clock: clk},
		sender:   &fakeSender{sent: map[string]string{}},
		users:    &fakeUsers{created: true},
		sessions: newFakeSessions(),
		clock:    clk,
		pub:      pub,
	}
	issuer := NewTokenIssuer(priv, 15*time.Minute, 30*24*time.Hour)
	issuer.now = clk.now
	f.uc = NewOTPUseCase(f.otps, f.sender, f.users, f.sessions, issuer, devCode)
	f.uc.now = clk.now
	return f
}

// requestCode requests a code and returns the one that was "sent".
func (f *otpFixture) requestCode(t *testing.T) string {
	t.Helper()
	require.NoError(t, f.uc.Request(context.Background(), testPhone))
	return f.sender.sent[testPhone]
}

// --- Request ---

func TestRequestSendsCodeAndStoresOnlyItsHash(t *testing.T) {
	f := newOTPFixture(t, "")

	code := f.requestCode(t)

	assert.Len(t, code, domain.OTPLength)
	require.Len(t, f.otps.codes, 1)
	assert.Equal(t, hashCode(testPhone, code), f.otps.codes[0].CodeHash)
	assert.NotContains(t, string(f.otps.codes[0].CodeHash), code, "plaintext code must not be stored")
	assert.Equal(t, f.clock.now().Add(domain.OTPTTL), f.otps.codes[0].ExpiresAt)
}

func TestRequestRejectsInvalidPhone(t *testing.T) {
	f := newOTPFixture(t, "")

	assert.ErrorIs(t, f.uc.Request(context.Background(), "98765"), domain.ErrInvalidPhone)
	assert.Empty(t, f.sender.sent)
}

func TestRequestRateLimitPerPhone(t *testing.T) {
	f := newOTPFixture(t, "")
	for range domain.MaxOTPRequestsPerWindow {
		f.requestCode(t)
	}

	assert.ErrorIs(t, f.uc.Request(context.Background(), testPhone), domain.ErrTooManyRequests)

	f.clock.advance(domain.OTPRequestWindow + time.Second)
	assert.NoError(t, f.uc.Request(context.Background(), testPhone), "window has passed")
}

// --- Verify ---

func TestVerifyCorrectCodeLogsIn(t *testing.T) {
	f := newOTPFixture(t, "")
	code := f.requestCode(t)

	result, err := f.uc.Verify(context.Background(), testPhone, code)
	require.NoError(t, err)

	assert.Equal(t, testUserID, result.UserID)
	assert.True(t, result.IsNewUser)
	assert.NotEmpty(t, result.RefreshToken)

	claims, err := authtoken.Verify(f.pub, result.AccessToken)
	require.NoError(t, err, "access token must verify with the matching public key")
	assert.Equal(t, testUserID, claims.UserID)

	_, err = f.uc.Verify(context.Background(), testPhone, code)
	assert.ErrorIs(t, err, domain.ErrInvalidCode, "a code works only once")
}

func TestVerifyExistingUserIsNotNew(t *testing.T) {
	f := newOTPFixture(t, "")
	f.users.created = false

	result, err := f.uc.Verify(context.Background(), testPhone, f.requestCode(t))

	require.NoError(t, err)
	assert.False(t, result.IsNewUser)
}

func TestVerifyRejects(t *testing.T) {
	tests := []struct {
		name  string
		setup func(f *otpFixture) (phone, code string)
		want  error
	}{
		{"WrongCode", func(f *otpFixture) (string, string) { f.requestCode(t); return testPhone, "abcdef" }, domain.ErrInvalidCode},
		{"NeverRequested", func(f *otpFixture) (string, string) { return testPhone, "123456" }, domain.ErrInvalidCode},
		{"Expired", func(f *otpFixture) (string, string) {
			code := f.requestCode(t)
			f.clock.advance(domain.OTPTTL)
			return testPhone, code
		}, domain.ErrInvalidCode},
		{"OlderCodeAfterNewerRequested", func(f *otpFixture) (string, string) {
			old := f.requestCode(t)
			f.requestCode(t)
			return testPhone, old
		}, domain.ErrInvalidCode},
		{"InvalidPhone", func(f *otpFixture) (string, string) { return "98765", "123456" }, domain.ErrInvalidPhone},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newOTPFixture(t, "")
			phone, code := tt.setup(f)

			_, err := f.uc.Verify(context.Background(), phone, code)

			assert.ErrorIs(t, err, tt.want)
			assert.Zero(t, f.users.calls, "user-service must not be called for a rejected code")
		})
	}
}

func TestVerifyWrongCodeCountsAnAttempt(t *testing.T) {
	f := newOTPFixture(t, "")
	f.requestCode(t)

	_, _ = f.uc.Verify(context.Background(), testPhone, "not-it")

	assert.Equal(t, 1, f.otps.codes[0].Attempts)
}

func TestVerifyUserServiceDownLeavesCodeUsable(t *testing.T) {
	f := newOTPFixture(t, "")
	code := f.requestCode(t)
	f.users.err = domain.ErrUserServiceUnavailable

	_, err := f.uc.Verify(context.Background(), testPhone, code)
	require.ErrorIs(t, err, domain.ErrUserServiceUnavailable)
	assert.Nil(t, f.otps.codes[0].ConsumedAt, "code must stay usable for a retry")

	f.users.err = nil
	_, err = f.uc.Verify(context.Background(), testPhone, code)
	assert.NoError(t, err, "retry with the same code succeeds")
}

func TestVerifyAcceptsDevCodeWhenConfigured(t *testing.T) {
	f := newOTPFixture(t, "123456")
	f.requestCode(t)

	_, err := f.uc.Verify(context.Background(), testPhone, "123456")

	assert.NoError(t, err)
}

func TestVerifyDevCodeStillNeedsARequestedCode(t *testing.T) {
	f := newOTPFixture(t, "123456")

	_, err := f.uc.Verify(context.Background(), testPhone, "123456")

	assert.ErrorIs(t, err, domain.ErrInvalidCode, "the dev code skips the SMS, not the request step")
}

func TestRandomCodeIsAlwaysSixDigits(t *testing.T) {
	for range 1000 {
		code, err := randomCode()
		require.NoError(t, err)
		require.Len(t, code, domain.OTPLength)
		for _, r := range code {
			require.True(t, r >= '0' && r <= '9', code)
		}
	}
}

func TestVerifyAttemptLimit(t *testing.T) {
	tests := []struct {
		name         string
		wrongGuesses int
		want         error
	}{
		{name: "UnderLimitCorrectCodeWorks", wrongGuesses: domain.MaxOTPAttempts - 1, want: nil},
		{name: "AtLimitCorrectCodeRefused", wrongGuesses: domain.MaxOTPAttempts, want: domain.ErrTooManyAttempts},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newOTPFixture(t, "")
			code := f.requestCode(t)
			for range tt.wrongGuesses {
				_, err := f.uc.Verify(context.Background(), testPhone, "abcdef")
				require.ErrorIs(t, err, domain.ErrInvalidCode)
			}

			_, err := f.uc.Verify(context.Background(), testPhone, code)

			if tt.want == nil {
				assert.NoError(t, err)
			} else {
				assert.ErrorIs(t, err, tt.want)
			}
		})
	}
}

func TestVerifyNewCodeResetsAttemptLimit(t *testing.T) {
	f := newOTPFixture(t, "")
	f.requestCode(t)
	for range domain.MaxOTPAttempts {
		_, _ = f.uc.Verify(context.Background(), testPhone, "abcdef")
	}

	fresh := f.requestCode(t)
	_, err := f.uc.Verify(context.Background(), testPhone, fresh)

	assert.NoError(t, err, "a newly requested code has its own attempt budget")
}
