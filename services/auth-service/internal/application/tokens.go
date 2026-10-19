package application

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/SamPariat/roundup/services/auth-service/internal/domain"
	"github.com/SamPariat/roundup/shared/authtoken"
)

// refreshTokenBytes is the refresh token length before encoding: 256 bits,
// far beyond guessing range.
const refreshTokenBytes = 32

// TokenIssuer mints access tokens (Ed25519 JWTs via shared/authtoken) and
// opaque refresh tokens. Shared by the login and refresh use cases.
type TokenIssuer struct {
	privateKey ed25519.PrivateKey
	accessTTL  time.Duration
	refreshTTL time.Duration
	now        func() time.Time
}

// NewTokenIssuer constructs a TokenIssuer that signs with privateKey.
func NewTokenIssuer(privateKey ed25519.PrivateKey, accessTTL, refreshTTL time.Duration) *TokenIssuer {
	return &TokenIssuer{privateKey: privateKey, accessTTL: accessTTL, refreshTTL: refreshTTL, now: time.Now}
}

// startSession creates a session for userID and returns its first token pair.
func (i *TokenIssuer) startSession(ctx context.Context, sessions domain.SessionRepository, userID string) (*domain.TokenPair, error) {
	refresh, hash, err := newRefreshToken()
	if err != nil {
		return nil, err
	}

	refreshExpiresAt := i.now().Add(i.refreshTTL)

	if _, err := sessions.Create(ctx, userID, hash, refreshExpiresAt); err != nil {
		return nil, err
	}

	return i.pair(userID, refresh, refreshExpiresAt)
}

// pair signs an access token for userID and bundles it with refresh.
func (i *TokenIssuer) pair(userID, refresh string, refreshExpiresAt time.Time) (*domain.TokenPair, error) {
	access, accessExpiresAt, err := authtoken.Sign(i.privateKey, userID, i.accessTTL)
	if err != nil {
		return nil, err
	}

	return &domain.TokenPair{
		AccessToken:           access,
		AccessTokenExpiresAt:  accessExpiresAt,
		RefreshToken:          refresh,
		RefreshTokenExpiresAt: refreshExpiresAt,
	}, nil
}

// newRefreshToken returns 32 random bytes, base64url-encoded, and the hash to
// store. The token has full entropy, so a fast SHA-256 is enough; slow hashes
// like bcrypt only matter for guessable secrets such as passwords.
func newRefreshToken() (token string, hash []byte, err error) {
	raw := make([]byte, refreshTokenBytes)
	if _, err := rand.Read(raw); err != nil {
		return "", nil, fmt.Errorf("generate refresh token: %w", err)
	}

	token = base64.RawURLEncoding.EncodeToString(raw)

	return token, hashRefreshToken(token), nil
}

// hashRefreshToken is the lookup key a refresh token is stored under.
func hashRefreshToken(token string) []byte {
	sum := sha256.Sum256([]byte(token))

	return sum[:]
}
