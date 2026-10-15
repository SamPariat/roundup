package application

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"fmt"
	"math/big"
	"time"

	"github.com/SamPariat/roundup/services/auth-service/internal/domain"
	"github.com/SamPariat/roundup/shared/phone"
)

// OTPUseCase handles phone login: sending a one-time code, then exchanging a
// correct code for a session and tokens.
type OTPUseCase struct {
	otps     domain.OTPRepository
	sender   domain.OTPSender
	users    domain.UserDirectory
	sessions domain.SessionRepository
	issuer   *TokenIssuer
	devCode  string
	now      func() time.Time
}

// NewOTPUseCase constructs an OTPUseCase. devCode, when non-empty, is accepted
// in place of the real code; config only allows it outside prod.
func NewOTPUseCase(
	otps domain.OTPRepository,
	sender domain.OTPSender,
	users domain.UserDirectory,
	sessions domain.SessionRepository,
	issuer *TokenIssuer,
	devCode string,
) *OTPUseCase {
	return &OTPUseCase{otps: otps, sender: sender, users: users, sessions: sessions, issuer: issuer, devCode: devCode, now: time.Now}
}

// Request sends a new code to number. Returns domain.ErrInvalidPhone for a
// malformed number and domain.ErrTooManyRequests when the per-phone limit is hit.
func (u *OTPUseCase) Request(ctx context.Context, number string) error {
	if !phone.IsValid(number) {
		return domain.ErrInvalidPhone
	}

	recent, err := u.otps.CountSince(ctx, number, u.now().Add(-domain.OTPRequestWindow))
	if err != nil {
		return err
	}
	if recent >= domain.MaxOTPRequestsPerWindow {
		return domain.ErrTooManyRequests
	}

	code, err := randomCode()
	if err != nil {
		return err
	}

	if err := u.otps.Create(ctx, number, hashCode(number, code), u.now().Add(domain.OTPTTL)); err != nil {
		return err
	}

	return u.sender.Send(ctx, number, code)
}

// Verify checks code against the newest code sent to number and, if it
// matches, logs the user in: the user is found or created in user-service, the
// code is consumed, and a new session's tokens are returned.
//
// Returns domain.ErrInvalidCode for a wrong, expired, used, or unrequested
// code, and domain.ErrUserServiceUnavailable if the user cannot be looked up —
// the code is left unused in that case so the user can retry.
func (u *OTPUseCase) Verify(ctx context.Context, number, code string) (*domain.LoginResult, error) {
	if !phone.IsValid(number) {
		return nil, domain.ErrInvalidPhone
	}

	otp, err := u.otps.Latest(ctx, number)
	if err != nil {
		return nil, err
	}
	if otp.ConsumedAt != nil || !u.now().Before(otp.ExpiresAt) {
		return nil, domain.ErrInvalidCode
	}

	// Checked before comparing, so once the limit is hit even the correct code
	// is refused; otherwise the guess after the limit could still succeed.
	if otp.Attempts >= domain.MaxOTPAttempts {
		return nil, domain.ErrTooManyAttempts
	}

	if !u.codeMatches(number, code, otp.CodeHash) {
		if err := u.otps.IncrementAttempts(ctx, otp.ID); err != nil {
			return nil, err
		}
		return nil, domain.ErrInvalidCode
	}

	userID, created, err := u.users.EnsureUser(ctx, number)
	if err != nil {
		return nil, err
	}

	// Consume only after the user exists, so a user-service outage leaves the
	// code usable for a retry.
	if err := u.otps.Consume(ctx, otp.ID); err != nil {
		return nil, err
	}

	pair, err := u.issuer.startSession(ctx, u.sessions, userID)
	if err != nil {
		return nil, err
	}

	return &domain.LoginResult{TokenPair: *pair, UserID: userID, IsNewUser: created}, nil
}

// codeMatches compares in constant time, so response timing reveals nothing
// about how close a guess was.
func (u *OTPUseCase) codeMatches(number, code string, storedHash []byte) bool {
	if u.devCode != "" && subtle.ConstantTimeCompare([]byte(code), []byte(u.devCode)) == 1 {
		return true
	}

	return subtle.ConstantTimeCompare(hashCode(number, code), storedHash) == 1
}

// hashCode binds the code to the phone it was sent to, so equal codes for
// different phones never share a hash.
func hashCode(number, code string) []byte {
	sum := sha256.Sum256([]byte(number + ":" + code))
	return sum[:]
}

// randomCode returns a uniformly random OTPLength-digit code, leading zeros kept.
func randomCode() (string, error) {
	limit := big.NewInt(1)
	for range domain.OTPLength {
		limit.Mul(limit, big.NewInt(10))
	}

	n, err := rand.Int(rand.Reader, limit)
	if err != nil {
		return "", fmt.Errorf("generate code: %w", err)
	}

	return fmt.Sprintf("%0*d", domain.OTPLength, n), nil
}
