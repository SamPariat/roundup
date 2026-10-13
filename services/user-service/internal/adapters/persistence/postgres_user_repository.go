package persistence

import (
	"context"

	"github.com/SamPariat/roundup/services/user-service/internal/adapters/persistence/db"
	"github.com/SamPariat/roundup/services/user-service/internal/domain"
	"github.com/SamPariat/roundup/shared/postgres"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// postgres_user_repository.go implements domain.UserRepository and
// domain.DeviceTokenRepository backed by PostgreSQL. It wraps the
// sqlc-generated Queries and maps between db types and domain types.

// PostgresUserRepository implements domain.UserRepository and
// domain.DeviceTokenRepository using PostgreSQL.
type PostgresUserRepository struct {
	queries *db.Queries
}

// NewPostgresUserRepository constructs a PostgresUserRepository from a pgxpool.
func NewPostgresUserRepository(pool *pgxpool.Pool) *PostgresUserRepository {
	return &PostgresUserRepository{queries: db.New(pool)}
}

// Create inserts a new user. Returns ErrPhoneTaken if an active user already
// has this phone number.
func (r *PostgresUserRepository) Create(ctx context.Context, phone string) (*domain.User, error) {
	row, err := r.queries.CreateUser(ctx, phone)
	if err != nil {
		if postgres.IsUniqueViolation(err) {
			return nil, domain.NewDomainError(domain.ErrPhoneTaken, phone)
		}
		return nil, err
	}

	return toDomainUser(row), nil
}

// GetByID returns the active user with the given ID.
func (r *PostgresUserRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrUserNotFound, id)
	}

	row, err := r.queries.GetUserByID(ctx, uid)
	if err != nil {
		return nil, notFoundOr(err, id)
	}

	return toDomainUser(row), nil
}

// GetByPhone returns the active user with the given phone number.
func (r *PostgresUserRepository) GetByPhone(ctx context.Context, phone string) (*domain.User, error) {
	row, err := r.queries.GetUserByPhone(ctx, phone)
	if err != nil {
		return nil, notFoundOr(err, phone)
	}

	return toDomainUser(row), nil
}

// UpdateProfile applies the non-nil fields of cmd and returns the updated user.
func (r *PostgresUserRepository) UpdateProfile(ctx context.Context, cmd domain.UpdateProfileCommand) (*domain.User, error) {
	uid, err := uuid.Parse(cmd.UserID)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrUserNotFound, cmd.UserID)
	}

	row, err := r.queries.UpdateUserProfile(ctx, db.UpdateUserProfileParams{
		ID:        uid,
		Name:      cmd.Name,
		AvatarUrl: cmd.AvatarURL,
	})
	if err != nil {
		return nil, notFoundOr(err, cmd.UserID)
	}

	return toDomainUser(row), nil
}

// SoftDelete marks the user as deleted.
func (r *PostgresUserRepository) SoftDelete(ctx context.Context, id string) error {
	uid, err := uuid.Parse(id)
	if err != nil {
		return domain.NewDomainError(domain.ErrUserNotFound, id)
	}

	affected, err := r.queries.SoftDeleteUser(ctx, uid)
	if err != nil {
		return err
	}
	if affected == 0 {
		return domain.NewDomainError(domain.ErrUserNotFound, id)
	}

	return nil
}

// Upsert registers the token to the user, moving it from any previous owner.
func (r *PostgresUserRepository) Upsert(ctx context.Context, cmd domain.RegisterDeviceTokenCommand) (*domain.DeviceToken, error) {
	uid, err := uuid.Parse(cmd.UserID)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrUserNotFound, cmd.UserID)
	}

	row, err := r.queries.UpsertDeviceToken(ctx, db.UpsertDeviceTokenParams{
		UserID:   uid,
		Token:    cmd.Token,
		Platform: string(cmd.Platform),
	})
	if err != nil {
		return nil, err
	}

	return &domain.DeviceToken{
		ID:        row.ID,
		UserID:    row.UserID.String(),
		Token:     row.Token,
		Platform:  domain.Platform(row.Platform),
		CreatedAt: row.CreatedAt.Time,
	}, nil
}

// Delete removes the token if the user owns it.
func (r *PostgresUserRepository) Delete(ctx context.Context, cmd domain.RemoveDeviceTokenCommand) error {
	uid, err := uuid.Parse(cmd.UserID)
	if err != nil {
		return domain.NewDomainError(domain.ErrDeviceTokenNotFound, cmd.Token)
	}

	affected, err := r.queries.DeleteDeviceToken(ctx, db.DeleteDeviceTokenParams{
		UserID: uid,
		Token:  cmd.Token,
	})
	if err != nil {
		return err
	}
	if affected == 0 {
		return domain.NewDomainError(domain.ErrDeviceTokenNotFound, cmd.Token)
	}

	return nil
}

// DeleteAllForUser removes every token registered to the user.
func (r *PostgresUserRepository) DeleteAllForUser(ctx context.Context, userID string) error {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil
	}

	return r.queries.DeleteDeviceTokensForUser(ctx, uid)
}

// toDomainUser maps a sqlc row to a domain.User, turning NULL columns into empty strings.
func toDomainUser(row db.User) *domain.User {
	user := &domain.User{
		ID:        row.ID.String(),
		Phone:     row.Phone,
		CreatedAt: row.CreatedAt.Time,
		UpdatedAt: row.UpdatedAt.Time,
	}
	if row.Name != nil {
		user.Name = *row.Name
	}
	if row.AvatarUrl != nil {
		user.AvatarURL = *row.AvatarUrl
	}

	return user
}

// notFoundOr maps pgx.ErrNoRows to ErrUserNotFound and passes any other error through.
func notFoundOr(err error, detail string) error {
	if postgres.IsNoRows(err) {
		return domain.NewDomainError(domain.ErrUserNotFound, detail)
	}

	return err
}
