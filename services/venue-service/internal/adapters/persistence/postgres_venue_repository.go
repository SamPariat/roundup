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

// postgres_venue_repository.go implements domain.VenueRepository backed by PostgreSQL.
// It wraps the sqlc-generated Queries and maps between db types and domain types.

// pgUniqueViolation is the Postgres error code for unique constraint violations.
const pgUniqueViolation = "23505"

// PostgresVenueRepository implements domain.VenueRepository using PostgreSQL.
type PostgresVenueRepository struct {
	queries *db.Queries
}

// NewPostgresVenueRepository constructs a PostgresVenueRepository from a pgxpool.
func NewPostgresVenueRepository(pool *pgxpool.Pool) *PostgresVenueRepository {
	return &PostgresVenueRepository{queries: db.New(pool)}
}

// AddFavorite persists a new favorite for the user. Returns ErrAlreadySaved if
// the (squad_id, user_id, place_id) combination already exists.
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

// RemoveFavorite deletes a saved favorite. Returns ErrNotSaved if the record
// does not exist before attempting to delete.
func (r *PostgresVenueRepository) RemoveFavorite(ctx context.Context, cmd domain.RemoveFavoriteCommand) error {
	exists, err := r.queries.IsFavorite(ctx, db.IsFavoriteParams{
		SquadID: cmd.SquadID,
		UserID:  cmd.UserID,
		PlaceID: cmd.PlaceID,
	})
	if err != nil {
		return err
	}
	if !exists {
		return domain.NewDomainError(domain.ErrNotSaved, cmd.PlaceID)
	}

	return r.queries.RemoveFavorite(ctx, db.RemoveFavoriteParams{
		SquadID: cmd.SquadID,
		UserID:  cmd.UserID,
		PlaceID: cmd.PlaceID,
	})
}

// ListFavorites returns all saved venues for the given squad, ordered by most recently saved.
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

// IsFavorite reports whether the given (squadID, userID, placeID) combination is saved.
func (r *PostgresVenueRepository) IsFavorite(ctx context.Context, squadID, userID, placeID string) (bool, error) {
	isFavorite, err := r.queries.IsFavorite(ctx, db.IsFavoriteParams{
		SquadID: squadID,
		UserID:  userID,
		PlaceID: placeID,
	})
	if err != nil {
		return false, err
	}

	return isFavorite, nil
}

// RecordVisit inserts a new venue visit record for the squad.
func (r *PostgresVenueRepository) RecordVisit(ctx context.Context, cmd domain.RecordVisitCommand) error {
	return r.queries.RecordVisit(ctx, db.RecordVisitParams{
		SquadID:         cmd.SquadID,
		EventID:         cmd.EventID,
		PlaceID:         cmd.PlaceID,
		Name:            cmd.Name,
		VisitedAt:       pgtype.Timestamptz{Time: cmd.VisitedAt, Valid: true},
		AvgSpendInPaise: pgtype.Int8{Int64: cmd.AvgSpendInPaise, Valid: cmd.AvgSpendInPaise > 0},
	})
}

// GetVisitHistory returns aggregated visit summaries for all venues visited by the squad.
func (r *PostgresVenueRepository) GetVisitHistory(ctx context.Context, squadID string) ([]domain.VisitSummary, error) {
	rows, err := r.queries.GetVisitHistory(ctx, squadID)
	if err != nil {
		return nil, err
	}

	result := make([]domain.VisitSummary, 0, len(rows))
	for _, row := range rows {
		result = append(result, domain.VisitSummary{
			PlaceID:       row.PlaceID,
			Name:          row.Name,
			VisitCount:    row.VisitCount,
			LastVisitedAt: toTime(row.LastVisitedAt),
		})
	}

	return result, nil
}

// toTime safely asserts an interface{} scanned from a pgx aggregate to time.Time.
// pgx v5 returns time.Time for timestamptz columns, but aggregate functions like
// MAX() produce interface{} — this helper handles both cases gracefully.
func toTime(v interface{}) time.Time {
	if t, ok := v.(time.Time); ok {
		return t
	}

	if ts, ok := v.(pgtype.Timestamptz); ok {
		return ts.Time
	}

	return time.Time{}
}
