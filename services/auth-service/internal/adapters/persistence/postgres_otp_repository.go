package persistence

import (
	"context"
	"time"

	"github.com/SamPariat/roundup/services/auth-service/internal/adapters/persistence/db"
	"github.com/SamPariat/roundup/services/auth-service/internal/domain"
	"github.com/SamPariat/roundup/shared/domainerr"
	"github.com/SamPariat/roundup/shared/postgres"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresOTPRepository implements domain.OTPRepository on top of the
// sqlc-generated queries.
type PostgresOTPRepository struct {
	queries *db.Queries
}

// NewPostgresOTPRepository constructs a PostgresOTPRepository from a pgxpool.
func NewPostgresOTPRepository(pool *pgxpool.Pool) *PostgresOTPRepository {
	return &PostgresOTPRepository{queries: db.New(pool)}
}

// CountSince returns how many codes were created for phone at or after since.
func (r *PostgresOTPRepository) CountSince(ctx context.Context, phone string, since time.Time) (int, error) {
	n, err := r.queries.CountOTPCodesSince(ctx, db.CountOTPCodesSinceParams{Phone: phone, CreatedAt: timestamptz(since)})

	return int(n), err
}

// Create stores a new code hash for phone.
func (r *PostgresOTPRepository) Create(ctx context.Context, phone string, codeHash []byte, expiresAt time.Time) error {
	return r.queries.CreateOTPCode(ctx, db.CreateOTPCodeParams{Phone: phone, CodeHash: codeHash, ExpiresAt: timestamptz(expiresAt)})
}

// Latest returns the newest code for phone, or ErrInvalidCode if there is none.
func (r *PostgresOTPRepository) Latest(ctx context.Context, phone string) (*domain.OTPCode, error) {
	row, err := r.queries.GetLatestOTPCode(ctx, phone)
	if err != nil {
		if postgres.IsNoRows(err) {
			return nil, domainerr.New(domain.ErrInvalidCode, "no code requested")
		}

		return nil, err
	}

	return &domain.OTPCode{
		ID:         row.ID,
		Phone:      row.Phone,
		CodeHash:   row.CodeHash,
		Attempts:   int(row.Attempts),
		ExpiresAt:  row.ExpiresAt.Time,
		ConsumedAt: timePtr(row.ConsumedAt),
	}, nil
}

// IncrementAttempts records one wrong guess against the code.
func (r *PostgresOTPRepository) IncrementAttempts(ctx context.Context, id int64) error {
	return r.queries.IncrementOTPAttempts(ctx, id)
}

// Consume marks the code used, or returns ErrInvalidCode if it already was.
func (r *PostgresOTPRepository) Consume(ctx context.Context, id int64) error {
	affected, err := r.queries.ConsumeOTPCode(ctx, id)
	if err != nil {
		return err
	}

	if affected == 0 {
		return domainerr.New(domain.ErrInvalidCode, "code already used")
	}

	return nil
}

func timestamptz(t time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: t, Valid: true}
}

func timePtr(ts pgtype.Timestamptz) *time.Time {
	if !ts.Valid {
		return nil
	}

	return &ts.Time
}
