// Package authtoken signs and verifies Roundup access tokens: short-lived JWTs
// signed with Ed25519. auth-service holds the private key and signs; the
// gateway holds only the public key and verifies, so it can check tokens
// locally without being able to mint them.
//
// This is the only place that knows the token format. Keep issuer, audience
// and algorithm checks here so signer and verifier can never drift apart.
package authtoken

import (
	"crypto/ed25519"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const (
	// Issuer is the iss claim: tokens are only ever minted by auth-service.
	Issuer = "roundup-auth"
	// Audience is the aud claim: tokens are only valid for the Roundup API.
	Audience = "roundup"
)

// ErrInvalidToken is returned by Verify for any token that must be rejected:
// bad signature, wrong algorithm, wrong issuer or audience, expired, or a
// subject that is not a user ID.
var ErrInvalidToken = errors.New("invalid access token")

// Claims is what a verified access token says about the caller.
type Claims struct {
	// UserID is the authenticated user's UUID (the sub claim).
	UserID string
	// ExpiresAt is when the token stops being accepted.
	ExpiresAt time.Time
}

// Sign returns an access token for userID that expires after ttl, along with
// its expiry time (truncated to whole seconds, as stored in the token).
func Sign(priv ed25519.PrivateKey, userID string, ttl time.Duration) (string, time.Time, error) {
	now := time.Now()
	claims := jwt.RegisteredClaims{
		Issuer:    Issuer,
		Subject:   userID,
		Audience:  jwt.ClaimStrings{Audience},
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
	}

	token, err := jwt.NewWithClaims(jwt.SigningMethodEdDSA, claims).SignedString(priv)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("sign access token: %w", err)
	}

	return token, claims.ExpiresAt.Time, nil
}

// Verify checks token against pub and returns its claims. Only EdDSA is
// accepted, which rules out "alg: none" and HMAC algorithm-confusion attacks.
// Every failure wraps ErrInvalidToken.
func Verify(pub ed25519.PublicKey, token string) (Claims, error) {
	var claims jwt.RegisteredClaims
	_, err := jwt.ParseWithClaims(token, &claims,
		func(*jwt.Token) (any, error) { return pub, nil },
		jwt.WithValidMethods([]string{jwt.SigningMethodEdDSA.Alg()}),
		jwt.WithIssuer(Issuer),
		jwt.WithAudience(Audience),
		jwt.WithExpirationRequired(),
	)
	if err != nil {
		return Claims{}, fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}

	if _, err := uuid.Parse(claims.Subject); err != nil {
		return Claims{}, fmt.Errorf("%w: subject is not a user ID", ErrInvalidToken)
	}

	return Claims{UserID: claims.Subject, ExpiresAt: claims.ExpiresAt.Time}, nil
}

// ParsePrivateKeyPEM parses a PKCS #8 Ed25519 private key, as written by
// `openssl genpkey -algorithm ed25519`. Literal "\n" sequences are accepted
// in place of newlines, for platforms that store env vars on one line.
func ParsePrivateKeyPEM(pemText string) (ed25519.PrivateKey, error) {
	key, err := jwt.ParseEdPrivateKeyFromPEM([]byte(unescapeNewlines(pemText)))
	if err != nil {
		return nil, fmt.Errorf("parse Ed25519 private key: %w", err)
	}

	priv, ok := key.(ed25519.PrivateKey)
	if !ok {
		return nil, errors.New("parse Ed25519 private key: not an Ed25519 key")
	}

	return priv, nil
}

// ParsePublicKeyPEM parses a PKIX Ed25519 public key, as written by
// `openssl pkey -pubout`. Literal "\n" sequences are accepted in place of newlines.
func ParsePublicKeyPEM(pemText string) (ed25519.PublicKey, error) {
	key, err := jwt.ParseEdPublicKeyFromPEM([]byte(unescapeNewlines(pemText)))
	if err != nil {
		return nil, fmt.Errorf("parse Ed25519 public key: %w", err)
	}

	pub, ok := key.(ed25519.PublicKey)
	if !ok {
		return nil, errors.New("parse Ed25519 public key: not an Ed25519 key")
	}

	return pub, nil
}

func unescapeNewlines(s string) string {
	return strings.ReplaceAll(s, `\n`, "\n")
}
