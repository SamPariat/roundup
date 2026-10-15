package authtoken

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const userID = "7dc92d84-1950-43d6-b676-debacf84d830"

func newKeyPair(t *testing.T) (ed25519.PublicKey, ed25519.PrivateKey) {
	t.Helper()
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)
	return pub, priv
}

// signRaw signs arbitrary claims with any method, for building tokens that Sign would never produce.
func signRaw(t *testing.T, method jwt.SigningMethod, key any, claims jwt.RegisteredClaims) string {
	t.Helper()
	token, err := jwt.NewWithClaims(method, claims).SignedString(key)
	require.NoError(t, err)
	return token
}

func validClaims() jwt.RegisteredClaims {
	return jwt.RegisteredClaims{
		Issuer:    Issuer,
		Subject:   userID,
		Audience:  jwt.ClaimStrings{Audience},
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute)),
	}
}

func TestSignThenVerify(t *testing.T) {
	pub, priv := newKeyPair(t)

	token, expiresAt, err := Sign(priv, userID, 15*time.Minute)
	require.NoError(t, err)

	claims, err := Verify(pub, token)
	require.NoError(t, err)
	assert.Equal(t, userID, claims.UserID)
	assert.True(t, expiresAt.Equal(claims.ExpiresAt))
	assert.WithinDuration(t, time.Now().Add(15*time.Minute), claims.ExpiresAt, 2*time.Second)
}

func TestVerifyRejects(t *testing.T) {
	pub, priv := newKeyPair(t)
	_, otherPriv := newKeyPair(t)

	expired := validClaims()
	expired.ExpiresAt = jwt.NewNumericDate(time.Now().Add(-time.Minute))
	wrongIssuer := validClaims()
	wrongIssuer.Issuer = "someone-else"
	wrongAudience := validClaims()
	wrongAudience.Audience = jwt.ClaimStrings{"other-api"}
	noExpiry := validClaims()
	noExpiry.ExpiresAt = nil
	badSubject := validClaims()
	badSubject.Subject = "admin"

	valid := signRaw(t, jwt.SigningMethodEdDSA, priv, validClaims())
	parts := strings.Split(valid, ".")

	tests := []struct {
		name  string
		token string
	}{
		{"Expired", signRaw(t, jwt.SigningMethodEdDSA, priv, expired)},
		{"SignedWithAnotherKey", signRaw(t, jwt.SigningMethodEdDSA, otherPriv, validClaims())},
		{"WrongIssuer", signRaw(t, jwt.SigningMethodEdDSA, priv, wrongIssuer)},
		{"WrongAudience", signRaw(t, jwt.SigningMethodEdDSA, priv, wrongAudience)},
		{"NoExpiry", signRaw(t, jwt.SigningMethodEdDSA, priv, noExpiry)},
		{"SubjectNotAUserID", signRaw(t, jwt.SigningMethodEdDSA, priv, badSubject)},
		// "alg: none" carries no signature at all.
		{"AlgNone", signRaw(t, jwt.SigningMethodNone, jwt.UnsafeAllowNoneSignatureType, validClaims())},
		// Algorithm confusion: an attacker HMAC-signs with the (public) key bytes as the secret.
		{"HS256WithPublicKeyAsSecret", signRaw(t, jwt.SigningMethodHS256, []byte(pub), validClaims())},
		// Payload swapped after signing: signature no longer matches.
		{"TamperedPayload", parts[0] + "." + strings.Split(signRaw(t, jwt.SigningMethodEdDSA, priv, badSubject), ".")[1] + "." + parts[2]},
		{"Garbage", "not.a.jwt"},
		{"Empty", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Verify(pub, tt.token)
			assert.ErrorIs(t, err, ErrInvalidToken)
		})
	}
}

func TestParseKeyPEMRoundTrip(t *testing.T) {
	pub, priv := newKeyPair(t)

	privDER, err := x509.MarshalPKCS8PrivateKey(priv)
	require.NoError(t, err)
	pubDER, err := x509.MarshalPKIXPublicKey(pub)
	require.NoError(t, err)
	privPEM := string(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: privDER}))
	pubPEM := string(pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubDER}))

	for name, escape := range map[string]bool{"RealNewlines": false, "EscapedNewlines": true} {
		t.Run(name, func(t *testing.T) {
			p, q := privPEM, pubPEM
			if escape {
				p, q = strings.ReplaceAll(p, "\n", `\n`), strings.ReplaceAll(q, "\n", `\n`)
			}

			parsedPriv, err := ParsePrivateKeyPEM(p)
			require.NoError(t, err)
			parsedPub, err := ParsePublicKeyPEM(q)
			require.NoError(t, err)

			token, _, err := Sign(parsedPriv, userID, time.Minute)
			require.NoError(t, err)
			_, err = Verify(parsedPub, token)
			assert.NoError(t, err)
		})
	}

	_, err = ParsePrivateKeyPEM("not a key")
	assert.Error(t, err)
	_, err = ParsePublicKeyPEM(privPEM)
	assert.Error(t, err, "a private key is not a public key")
}
