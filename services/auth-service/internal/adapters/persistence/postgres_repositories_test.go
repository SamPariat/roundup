package persistence

import (
	"context"
	"crypto/sha256"
	"testing"
	"time"

	"github.com/SamPariat/roundup/services/auth-service/internal/domain"
	"github.com/SamPariat/roundup/shared/postgres/postgrestest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// postgres_repositories_test.go runs both repositories against a real,
// freshly migrated Postgres database. Tests are skipped unless
// TEST_POSTGRES_DSN is set (see shared/postgres/postgrestest).

const (
	migrationsDir = "../../../db/migrations"
	phone         = "+919876543210"
	userID        = "7dc92d84-1950-43d6-b676-debacf84d830"
)

func hash(s string) []byte {
	sum := sha256.Sum256([]byte(s))
	return sum[:]
}

// --- OTP codes ---

func TestOTPCountSinceAndLatest(t *testing.T) {
	ctx := context.Background()
	repo := NewPostgresOTPRepository(postgrestest.New(t, migrationsDir))
	before := time.Now().Add(-time.Second)

	require.NoError(t, repo.Create(ctx, phone, hash("111111"), time.Now().Add(5*time.Minute)))
	require.NoError(t, repo.Create(ctx, phone, hash("222222"), time.Now().Add(5*time.Minute)))
	require.NoError(t, repo.Create(ctx, "+919800000002", hash("333333"), time.Now().Add(5*time.Minute)))

	n, err := repo.CountSince(ctx, phone, before)
	require.NoError(t, err)
	assert.Equal(t, 2, n, "counts only this phone's codes")

	n, err = repo.CountSince(ctx, phone, time.Now().Add(time.Minute))
	require.NoError(t, err)
	assert.Zero(t, n, "codes before the window are not counted")

	latest, err := repo.Latest(ctx, phone)
	require.NoError(t, err)
	assert.Equal(t, hash("222222"), latest.CodeHash, "the newest code wins")
	assert.Nil(t, latest.ConsumedAt)
}

func TestOTPLatestNeverRequested(t *testing.T) {
	repo := NewPostgresOTPRepository(postgrestest.New(t, migrationsDir))

	_, err := repo.Latest(context.Background(), phone)

	assert.ErrorIs(t, err, domain.ErrInvalidCode)
}

func TestOTPAttemptsAndConsumeOnce(t *testing.T) {
	ctx := context.Background()
	repo := NewPostgresOTPRepository(postgrestest.New(t, migrationsDir))
	require.NoError(t, repo.Create(ctx, phone, hash("111111"), time.Now().Add(5*time.Minute)))
	code, err := repo.Latest(ctx, phone)
	require.NoError(t, err)

	require.NoError(t, repo.IncrementAttempts(ctx, code.ID))
	require.NoError(t, repo.IncrementAttempts(ctx, code.ID))
	require.NoError(t, repo.Consume(ctx, code.ID))

	code, err = repo.Latest(ctx, phone)
	require.NoError(t, err)
	assert.Equal(t, 2, code.Attempts)
	assert.NotNil(t, code.ConsumedAt)

	assert.ErrorIs(t, repo.Consume(ctx, code.ID), domain.ErrInvalidCode, "a code can be consumed once")
}

// --- sessions and refresh tokens ---

func TestSessionCreateAndFind(t *testing.T) {
	ctx := context.Background()
	repo := NewPostgresSessionRepository(postgrestest.New(t, migrationsDir))
	expires := time.Now().Add(time.Hour).Truncate(time.Microsecond)

	sessionID, err := repo.Create(ctx, userID, hash("refresh-1"), expires)
	require.NoError(t, err)

	stored, err := repo.FindRefreshToken(ctx, hash("refresh-1"))
	require.NoError(t, err)
	assert.Equal(t, sessionID, stored.SessionID)
	assert.Equal(t, userID, stored.UserID)
	assert.True(t, expires.Equal(stored.ExpiresAt))
	assert.Nil(t, stored.UsedAt)
	assert.False(t, stored.SessionRevoked)

	_, err = repo.FindRefreshToken(ctx, hash("unknown"))
	assert.ErrorIs(t, err, domain.ErrInvalidRefreshToken)
}

func TestSessionRotateMarksOldUsedOnce(t *testing.T) {
	ctx := context.Background()
	repo := NewPostgresSessionRepository(postgrestest.New(t, migrationsDir))
	sessionID, err := repo.Create(ctx, userID, hash("refresh-1"), time.Now().Add(time.Hour))
	require.NoError(t, err)
	first, err := repo.FindRefreshToken(ctx, hash("refresh-1"))
	require.NoError(t, err)

	require.NoError(t, repo.Rotate(ctx, first.ID, sessionID, hash("refresh-2"), time.Now().Add(time.Hour)))

	first, err = repo.FindRefreshToken(ctx, hash("refresh-1"))
	require.NoError(t, err)
	assert.NotNil(t, first.UsedAt)
	second, err := repo.FindRefreshToken(ctx, hash("refresh-2"))
	require.NoError(t, err)
	assert.Equal(t, sessionID, second.SessionID, "rotation stays on the same session")

	err = repo.Rotate(ctx, first.ID, sessionID, hash("refresh-3"), time.Now().Add(time.Hour))
	assert.ErrorIs(t, err, domain.ErrInvalidRefreshToken, "a used token cannot rotate again")
	_, err = repo.FindRefreshToken(ctx, hash("refresh-3"))
	assert.ErrorIs(t, err, domain.ErrInvalidRefreshToken, "the failed rotation wrote nothing")
}

func TestSessionRevoke(t *testing.T) {
	ctx := context.Background()
	repo := NewPostgresSessionRepository(postgrestest.New(t, migrationsDir))
	sessionID, err := repo.Create(ctx, userID, hash("refresh-1"), time.Now().Add(time.Hour))
	require.NoError(t, err)

	require.NoError(t, repo.Revoke(ctx, sessionID))
	require.NoError(t, repo.Revoke(ctx, sessionID), "revoking twice is a no-op")

	stored, err := repo.FindRefreshToken(ctx, hash("refresh-1"))
	require.NoError(t, err)
	assert.True(t, stored.SessionRevoked)
}
