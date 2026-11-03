package application

import (
	"testing"
	"time"

	"github.com/SamPariat/roundup/services/auth-service/internal/domain"
	"github.com/SamPariat/roundup/services/auth-service/internal/domain/domaintest"
	"github.com/SamPariat/roundup/shared/authtoken"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type otpFixture struct {
	uc       *OTPUseCase
	otps     *domaintest.OTPRepo
	sender   *domaintest.OTPSender
	users    *domaintest.UserDirectory
	sessions *domaintest.SessionRepo
	clock    *domaintest.Clock
	pub      []byte
}

func newOTPFixture(t *testing.T, devCode string) *otpFixture {
	t.Helper()

	clk := domaintest.NewClock()
	pub, priv := domaintest.NewKeys(t)
	f := &otpFixture{
		otps:     &domaintest.OTPRepo{Clock: clk},
		sender:   domaintest.NewOTPSender(),
		users:    &domaintest.UserDirectory{Created: true},
		sessions: domaintest.NewSessionRepo(),
		clock:    clk,
		pub:      pub,
	}
	issuer := NewTokenIssuer(priv, 15*time.Minute, 30*24*time.Hour)
	issuer.now = clk.Now
	f.uc = NewOTPUseCase(f.otps, f.sender, f.users, f.sessions, issuer, devCode)
	f.uc.now = clk.Now

	return f
}

// requestCode requests a code and returns the one that was "sent".
func (f *otpFixture) requestCode(t *testing.T) string {
	t.Helper()
	require.NoError(t, f.uc.Request(t.Context(), domaintest.Phone))

	return f.sender.Sent[domaintest.Phone]
}

// --- Request ---

func TestRequestSendsCodeAndStoresOnlyItsHash(t *testing.T) {
	f := newOTPFixture(t, "")

	code := f.requestCode(t)

	assert.Len(t, code, domain.OTPLength)
	require.Len(t, f.otps.Codes, 1)
	assert.Equal(t, hashCode(domaintest.Phone, code), f.otps.Codes[0].CodeHash)
	assert.NotContains(t, string(f.otps.Codes[0].CodeHash), code, "plaintext code must not be stored")
	assert.Equal(t, f.clock.Now().Add(domain.OTPTTL), f.otps.Codes[0].ExpiresAt)
}

func TestRequestRejectsInvalidPhone(t *testing.T) {
	f := newOTPFixture(t, "")

	require.ErrorIs(t, f.uc.Request(t.Context(), "98765"), domain.ErrInvalidPhone)
	assert.Empty(t, f.sender.Sent)
}

func TestRequestRateLimitPerPhone(t *testing.T) {
	f := newOTPFixture(t, "")
	for range domain.MaxOTPRequestsPerWindow {
		f.requestCode(t)
	}

	require.ErrorIs(t, f.uc.Request(t.Context(), domaintest.Phone), domain.ErrTooManyRequests)

	f.clock.Advance(domain.OTPRequestWindow + time.Second)
	assert.NoError(t, f.uc.Request(t.Context(), domaintest.Phone), "window has passed")
}

// --- Verify ---

func TestVerifyCorrectCodeLogsIn(t *testing.T) {
	f := newOTPFixture(t, "")
	code := f.requestCode(t)

	result, err := f.uc.Verify(t.Context(), domaintest.Phone, code)
	require.NoError(t, err)

	assert.Equal(t, domaintest.UserID, result.UserID)
	assert.True(t, result.IsNewUser)
	assert.NotEmpty(t, result.RefreshToken)

	claims, err := authtoken.Verify(f.pub, result.AccessToken)
	require.NoError(t, err, "access token must verify with the matching public key")
	assert.Equal(t, domaintest.UserID, claims.UserID)

	_, err = f.uc.Verify(t.Context(), domaintest.Phone, code)
	assert.ErrorIs(t, err, domain.ErrInvalidCode, "a code works only once")
}

func TestVerifyExistingUserIsNotNew(t *testing.T) {
	f := newOTPFixture(t, "")
	f.users.Created = false

	result, err := f.uc.Verify(t.Context(), domaintest.Phone, f.requestCode(t))

	require.NoError(t, err)
	assert.False(t, result.IsNewUser)
}

func TestVerifyRejects(t *testing.T) {
	tests := []struct {
		name  string
		setup func(f *otpFixture) (phone, code string)
		want  error
	}{
		{"WrongCode", func(f *otpFixture) (string, string) {
			f.requestCode(t)

			return domaintest.Phone, "abcdef"
		}, domain.ErrInvalidCode},
		{"NeverRequested", func(_ *otpFixture) (string, string) { return domaintest.Phone, "123456" }, domain.ErrInvalidCode},
		{"Expired", func(f *otpFixture) (string, string) {
			code := f.requestCode(t)
			f.clock.Advance(domain.OTPTTL)

			return domaintest.Phone, code
		}, domain.ErrInvalidCode},
		{"OlderCodeAfterNewerRequested", func(f *otpFixture) (string, string) {
			old := f.requestCode(t)
			f.requestCode(t)

			return domaintest.Phone, old
		}, domain.ErrInvalidCode},
		{"InvalidPhone", func(_ *otpFixture) (string, string) { return "98765", "123456" }, domain.ErrInvalidPhone},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newOTPFixture(t, "")
			phone, code := tt.setup(f)

			_, err := f.uc.Verify(t.Context(), phone, code)

			require.ErrorIs(t, err, tt.want)
			assert.Zero(t, f.users.Calls, "user-service must not be called for a rejected code")
		})
	}
}

func TestVerifyWrongCodeCountsAnAttempt(t *testing.T) {
	f := newOTPFixture(t, "")
	f.requestCode(t)

	_, _ = f.uc.Verify(t.Context(), domaintest.Phone, "not-it")

	assert.Equal(t, 1, f.otps.Codes[0].Attempts)
}

func TestVerifyUserServiceDownLeavesCodeUsable(t *testing.T) {
	f := newOTPFixture(t, "")
	code := f.requestCode(t)
	f.users.Err = domain.ErrUserServiceUnavailable

	_, err := f.uc.Verify(t.Context(), domaintest.Phone, code)
	require.ErrorIs(t, err, domain.ErrUserServiceUnavailable)
	assert.Nil(t, f.otps.Codes[0].ConsumedAt, "code must stay usable for a retry")

	f.users.Err = nil
	_, err = f.uc.Verify(t.Context(), domaintest.Phone, code)
	assert.NoError(t, err, "retry with the same code succeeds")
}

func TestVerifyAcceptsDevCodeWhenConfigured(t *testing.T) {
	f := newOTPFixture(t, "123456")
	f.requestCode(t)

	_, err := f.uc.Verify(t.Context(), domaintest.Phone, "123456")

	assert.NoError(t, err)
}

func TestVerifyDevCodeStillNeedsARequestedCode(t *testing.T) {
	f := newOTPFixture(t, "123456")

	_, err := f.uc.Verify(t.Context(), domaintest.Phone, "123456")

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
				_, err := f.uc.Verify(t.Context(), domaintest.Phone, "abcdef")
				require.ErrorIs(t, err, domain.ErrInvalidCode)
			}

			_, err := f.uc.Verify(t.Context(), domaintest.Phone, code)

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
		_, _ = f.uc.Verify(t.Context(), domaintest.Phone, "abcdef")
	}

	fresh := f.requestCode(t)
	_, err := f.uc.Verify(t.Context(), domaintest.Phone, fresh)

	assert.NoError(t, err, "a newly requested code has its own attempt budget")
}
