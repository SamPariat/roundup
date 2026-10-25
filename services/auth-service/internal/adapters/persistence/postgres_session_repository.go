package persistence

import (
	"context"
	"fmt"
	"time"

	"github.com/SamPariat/roundup/services/auth-service/internal/adapters/persistence/db"
	"github.com/SamPariat/roundup/services/auth-service/internal/domain"
	"github.com/SamPariat/roundup/shared/domainerr"
	"github.com/SamPariat/roundup/shared/postgres"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresSessionRepository implements domain.SessionRepository on top of the
// sqlc-generated queries. Writes that touch both tables (starting a session,
// rotating a token) run in one transaction.
type PostgresSessionRepository struct {
	pool    *pgxpool.Pool
	queries *db.Queries
}

// NewPostgresSessionRepository constructs a PostgresSessionRepository from a pgxpool.
func NewPostgresSessionRepository(pool *pgxpool.Pool) *PostgresSessionRepository {
	return &PostgresSessionRepository{pool: pool, queries: db.New(pool)}
}

// Create starts a session for userID with its first refresh token, in one
// transaction, and returns the session ID.
func (r *PostgresSessionRepository) Create(ctx context.Context, userID string, tokenHash []byte, expiresAt time.Time) (string, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return "", fmt.Errorf("create session: invalid user ID %q: %w", userID, err)
	}

	var sessionID uuid.UUID

	err = postgres.InTx(ctx, r.pool, func(tx pgx.Tx) error {
		q := r.queries.WithTx(tx)

		sessionID, err = q.CreateSession(ctx, uid)
		if err != nil {
			return err
		}

		return q.CreateRefreshToken(ctx, db.CreateRefreshTokenParams{SessionID: sessionID, TokenHash: tokenHash, ExpiresAt: postgres.Timestamptz(expiresAt)})
	})
	if err != nil {
		return "", err
	}

	return sessionID.String(), nil
}

// FindRefreshToken returns the token with this hash, or ErrInvalidRefreshToken.
func (r *PostgresSessionRepository) FindRefreshToken(ctx context.Context, tokenHash []byte) (*domain.StoredRefreshToken, error) {
	row, err := r.queries.GetRefreshToken(ctx, tokenHash)
	if err != nil {
		return nil, postgres.NoRowsAs(err, domain.ErrInvalidRefreshToken, "unknown token")
	}

	return &domain.StoredRefreshToken{
		ID:             row.ID,
		SessionID:      row.SessionID.String(),
		UserID:         row.UserID.String(),
		ExpiresAt:      row.ExpiresAt.Time,
		UsedAt:         timePtr(row.UsedAt),
		SessionRevoked: row.RevokedAt.Valid,
	}, nil
}

// Rotate marks oldID used and adds a new token to the session, atomically.
// If oldID was already used (a concurrent refresh won), nothing is written and
// ErrInvalidRefreshToken is returned.
func (r *PostgresSessionRepository) Rotate(ctx context.Context, oldID int64, sessionID string, newHash []byte, expiresAt time.Time) error {
	sid, err := uuid.Parse(sessionID)
	if err != nil {
		return domainerr.New(domain.ErrInvalidRefreshToken, "invalid session ID")
	}

	return postgres.InTx(ctx, r.pool, func(tx pgx.Tx) error {
		q := r.queries.WithTx(tx)

		affected, err := q.MarkRefreshTokenUsed(ctx, oldID)
		if err != nil {
			return err
		}

		if affected == 0 {
			return domainerr.New(domain.ErrInvalidRefreshToken, "token already used")
		}

		return q.CreateRefreshToken(ctx, db.CreateRefreshTokenParams{SessionID: sid, TokenHash: newHash, ExpiresAt: postgres.Timestamptz(expiresAt)})
	})
}

// Revoke ends the session. Revoking an already-revoked session is a no-op.
func (r *PostgresSessionRepository) Revoke(ctx context.Context, sessionID string) error {
	sid, err := uuid.Parse(sessionID)
	if err != nil {
		return domainerr.New(domain.ErrInvalidRefreshToken, "invalid session ID")
	}

	return r.queries.RevokeSession(ctx, sid)
}
