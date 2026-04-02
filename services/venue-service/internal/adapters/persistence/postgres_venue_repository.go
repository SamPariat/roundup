package persistence

import (
	"context"
	"errors"
	"time"

	"github.com/SamPariatIL/roundup/services/venue-service/internal/adapters/persistence/db"
	"github.com/SamPariatIL/roundup/services/venue-service/internal/domain"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

const pgUniqueViolation = "23505"

type PostgresVenueRepository struct {
	queries *db.Queries
}

func NewPostgresVenueRepository(pool *pgxpool.Pool) *PostgresVenueRepository {
	return &PostgresVenueRepository{queries: db.New(pool)}
}

func (r *PostgresVenueRepository) AddFavorite(ctx context.Context, cmd domain.AddFavoriteCommand) error {
	err := r.queries.AddFavorite(ctx, db.AddFavoriteParams{
		SquadID: cmd.SquadID,
		UserID:  cmd.UserID,
		PlaceID: cmd.PlaceID,
		Name:    cmd.Name,
	})
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok && pgErr.Code == pgUniqueViolation {
			return domain.NewDomainError(domain.ErrAlreadySaved, cmd.PlaceID)
		}

		return err
	}

	return nil
}

func (r *PostgresVenueRepository) RemoveFavorite(ctx context.Context, cmd domain.RemoveFavoriteCommand) error {
	err := r.queries.RemoveFavorite(ctx, db.RemoveFavoriteParams{
		SquadID: cmd.SquadID,
		UserID:  cmd.UserID,
		PlaceID: cmd.PlaceID,
	})
	if err != nil {
		return domain.NewDomainError(domain.ErrNotSaved, cmd.PlaceID)
	}

	return nil
}

func (r *PostgresVenueRepository) ListFavorites(ctx context.Context, squadID string) ([]domain.SavedVenue, error) {
	rows, err := r.queries.ListFavorites(ctx, squadID)
	if err != nil {
		return nil, err
	}

	result := make([]domain.SavedVenue, 0, len(rows))
	for _, row := range rows {
		result = append(result, domain.SavedVenue{
			ID:      row.ID,
			SquadID: row.SquadID,
			UserID:  row.UserID,
			PlaceID: row.PlaceID,
			Name:    row.Name,
			SavedAt: row.SavedAt.Time,
		})
	}

	return result, nil
}

func (r *PostgresVenueRepository) IsFavorite(ctx context.Context, squadID, userID, placeID string) (bool, error) {
	//TODO implement me
	panic("implement me")
}

func (r *PostgresVenueRepository) RecordVisit(ctx context.Context, cmd domain.RecordVisitCommand) error {
	return nil
}

func (r *PostgresVenueRepository) GetVisitHistory(ctx context.Context, squadID string) ([]domain.VisitSummary, error) {
	return nil, nil
}

func toTime(v interface{}) time.Time {
	if t, ok := v.(time.Time); ok {
		return t
	}

	if ts, ok := v.(pgtype.Timestamptz); ok {
		return ts.Time
	}

	return time.Time{}
}
