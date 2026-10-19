package application

import (
	"testing"
	"time"

	"github.com/SamPariat/roundup/services/auth-service/internal/domain"
	"github.com/SamPariat/roundup/shared/authtoken"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type tokenFixture struct {
	uc       *TokenUseCase
	sessions *fakeSessions
	clock    *clock
	pub      []byte
	login    *domain.TokenPair
}

// newTokenFixture logs a user in and returns the resulting first token pair.
func newTokenFixture(t *testing.T) *tokenFixture {
	t.Helper()

	clk := newClock()
	pub, priv := newKeys(t)
	sessions := newFakeSessions()
	issuer := NewTokenIssuer(priv, 15*time.Minute, 30*24*time.Hour)
	issuer.now = clk.now

	login, err := issuer.startSession(t.Context(), sessions, testUserID)
	require.NoError(t, err)

	uc := NewTokenUseCase(sessions, issuer)
	uc.now = clk.now

	return &tokenFixture{uc: uc, sessions: sessions, clock: clk, pub: pub, login: login}
}

func TestRefreshRotatesTokens(t *testing.T) {
	f := newTokenFixture(t)

	next, err := f.uc.Refresh(t.Context(), f.login.RefreshToken)
	require.NoError(t, err)

	assert.NotEqual(t, f.login.RefreshToken, next.RefreshToken)
	claims, err := authtoken.Verify(f.pub, next.AccessToken)
	require.NoError(t, err)
	assert.Equal(t, testUserID, claims.UserID)

	_, err = f.uc.Refresh(t.Context(), next.RefreshToken)
	assert.NoError(t, err, "the newest token keeps working")
}

func TestRefreshReuseRevokesSession(t *testing.T) {
	f := newTokenFixture(t)
	next, err := f.uc.Refresh(t.Context(), f.login.RefreshToken)
	require.NoError(t, err)

	// The first token is presented again: someone kept a copy.
	_, err = f.uc.Refresh(t.Context(), f.login.RefreshToken)
	require.ErrorIs(t, err, domain.ErrInvalidRefreshToken)

	_, err = f.uc.Refresh(t.Context(), next.RefreshToken)
	assert.ErrorIs(t, err, domain.ErrInvalidRefreshToken, "the legitimate newer token dies with the session")
}

func TestRefreshRejects(t *testing.T) {
	tests := []struct {
		name  string
		token func(f *tokenFixture) string
	}{
		{"Unknown", func(*tokenFixture) string { return "not-a-real-token" }},
		{"Expired", func(f *tokenFixture) string {
			f.clock.advance(30 * 24 * time.Hour)

			return f.login.RefreshToken
		}},
		{"LoggedOut", func(f *tokenFixture) string {
			require.NoError(t, f.uc.Logout(t.Context(), f.login.RefreshToken))

			return f.login.RefreshToken
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newTokenFixture(t)

			_, err := f.uc.Refresh(t.Context(), tt.token(f))

			assert.ErrorIs(t, err, domain.ErrInvalidRefreshToken)
		})
	}
}

func TestLogoutUnknownTokenFails(t *testing.T) {
	f := newTokenFixture(t)

	assert.ErrorIs(t, f.uc.Logout(t.Context(), "nope"), domain.ErrInvalidRefreshToken)
}

func TestLogoutTwiceIsNotAnError(t *testing.T) {
	f := newTokenFixture(t)

	require.NoError(t, f.uc.Logout(t.Context(), f.login.RefreshToken))
	assert.NoError(t, f.uc.Logout(t.Context(), f.login.RefreshToken))
}
