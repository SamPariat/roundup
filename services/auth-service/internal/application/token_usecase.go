package application

import (
	"context"
	"time"

	"github.com/SamPariat/roundup/services/auth-service/internal/domain"
	"github.com/SamPariat/roundup/shared/domainerr"
)

// TokenUseCase refreshes and revokes sessions using refresh tokens.
type TokenUseCase struct {
	sessions domain.SessionRepository
	issuer   *TokenIssuer
	now      func() time.Time
}

// NewTokenUseCase constructs a TokenUseCase.
func NewTokenUseCase(sessions domain.SessionRepository, issuer *TokenIssuer) *TokenUseCase {
	return &TokenUseCase{sessions: sessions, issuer: issuer, now: time.Now}
}

// Refresh exchanges a refresh token for a new pair, rotating it: the old token
// is marked used and a new one is issued on the same session.
//
// Presenting a token that was already used means it was copied — the client
// that rotated it holds the newer one. Since we cannot tell which party is the
// thief, the whole session is revoked and both must log in again.
// Returns domain.ErrInvalidRefreshToken in every rejection case.
func (u *TokenUseCase) Refresh(ctx context.Context, token string) (*domain.TokenPair, error) {
	stored, err := u.sessions.FindRefreshToken(ctx, hashRefreshToken(token))
	if err != nil {
		return nil, err
	}

	if stored.SessionRevoked || !u.now().Before(stored.ExpiresAt) {
		return nil, domain.ErrInvalidRefreshToken
	}

	if stored.UsedAt != nil {
		if err := u.sessions.Revoke(ctx, stored.SessionID); err != nil {
			return nil, err
		}

		return nil, domainerr.New(domain.ErrInvalidRefreshToken, "token reuse detected; session revoked")
	}

	next, hash, err := newRefreshToken()
	if err != nil {
		return nil, err
	}

	expiresAt := u.now().Add(u.issuer.refreshTTL)

	if err := u.sessions.Rotate(ctx, stored.ID, stored.SessionID, hash, expiresAt); err != nil {
		return nil, err
	}

	return u.issuer.pair(stored.UserID, next, expiresAt)
}

// Logout revokes the session the refresh token belongs to. Access tokens
// already issued keep working until they expire (minutes), the trade-off for
// verifying them without a database lookup. Logging out twice is not an error.
func (u *TokenUseCase) Logout(ctx context.Context, token string) error {
	stored, err := u.sessions.FindRefreshToken(ctx, hashRefreshToken(token))
	if err != nil {
		return err
	}

	return u.sessions.Revoke(ctx, stored.SessionID)
}
