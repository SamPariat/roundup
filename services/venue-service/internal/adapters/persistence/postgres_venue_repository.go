package persistence

import (
	"context"

	"github.com/SamPariat/roundup/services/venue-service/internal/adapters/persistence/db"
	"github.com/SamPariat/roundup/services/venue-service/internal/domain"
	"github.com/SamPariat/roundup/shared/domainerr"
	"github.com/SamPariat/roundup/shared/postgres"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresVenueRepository implements domain.VenueRepository on top of the
// sqlc-generated queries.
type PostgresVenueRepository struct {
	queries *db.Queries
}

// NewPostgresVenueRepository constructs a PostgresVenueRepository from a pgxpool.
func NewPostgresVenueRepository(pool *pgxpool.Pool) *PostgresVenueRepository {
	return &PostgresVenueRepository{queries: db.New(pool)}
}

// AddFavorite persists a new favorite for the user. Returns ErrAlreadySaved if
// the (squad_id, user_id, place_id) combination already exists.
func (r *PostgresVenueRepository) AddFavorite(ctx context.Context, cmd domain.AddFavoriteCommand, snapshot domain.VenueSnapshot) error {
	err := r.queries.AddFavorite(ctx, db.AddFavoriteParams{
		SquadID:     cmd.SquadID,
		UserID:      cmd.UserID,
		PlaceID:     cmd.PlaceID,
		Name:        snapshot.Name,
		Address:     snapshot.Address,
		Rating:      snapshot.AverageRating,
		RatingCount: int32(snapshot.RatingCount),
		PriceLevel:  int32(snapshot.PriceLevel),
		PhotoRef:    snapshot.Photo.Reference,
		PhotoWidth:  int32(snapshot.Photo.Width),
		PhotoHeight: int32(snapshot.Photo.Height),
	})
	if err != nil {
		if postgres.IsUniqueViolation(err) {
			return domainerr.New(domain.ErrAlreadySaved, cmd.PlaceID)
		}

		return err
	}

	return nil
}

// RemoveFavorite deletes a saved favorite in one statement. Returns
// ErrNotSaved if the user has not saved this venue in the squad.
func (r *PostgresVenueRepository) RemoveFavorite(ctx context.Context, cmd domain.RemoveFavoriteCommand) error {
	affected, err := r.queries.RemoveFavorite(ctx, db.RemoveFavoriteParams{
		SquadID: cmd.SquadID,
		UserID:  cmd.UserID,
		PlaceID: cmd.PlaceID,
	})
	if err != nil {
		return err
	}

	if affected == 0 {
		return domainerr.New(domain.ErrNotSaved, cmd.PlaceID)
	}

	return nil
}

// ListSquadFavorites returns the squad's favorites grouped per venue by the
// database, counting only saves by memberIDs.
func (r *PostgresVenueRepository) ListSquadFavorites(ctx context.Context, squadID string, memberIDs []string) ([]domain.FavoriteGroup, error) {
	rows, err := r.queries.ListSquadFavorites(ctx, db.ListSquadFavoritesParams{
		SquadID:   squadID,
		MemberIds: memberIDs,
	})
	if err != nil {
		return nil, err
	}

	result := make([]domain.FavoriteGroup, 0, len(rows))
	for _, row := range rows {
		result = append(result, domain.FavoriteGroup{
			PlaceID: row.PlaceID,
			VenueSnapshot: domain.VenueSnapshot{
				Name:          row.Name,
				Address:       row.Address,
				AverageRating: row.Rating,
				RatingCount:   int(row.RatingCount),
				PriceLevel:    int(row.PriceLevel),
				Photo:         photoRef(row.PhotoRef, row.PhotoWidth, row.PhotoHeight),
			},
			SavedBy:     row.SavedBy,
			LastSavedAt: row.LastSavedAt.Time,
		})
	}

	return result, nil
}

// SavedInSquads returns the squads, among squadIDs, in which the user saved
// placeID, in one query however many squads there are.
func (r *PostgresVenueRepository) SavedInSquads(ctx context.Context, userID, placeID string, squadIDs []string) ([]string, error) {
	return r.queries.SavedInSquads(ctx, db.SavedInSquadsParams{
		UserID:   userID,
		PlaceID:  placeID,
		SquadIds: squadIDs,
	})
}

// RecordVisit inserts a new venue visit record for the squad. A zero
// AvgSpendInPaise is stored as NULL (unknown) rather than as a real 0.
// Returns ErrVisitAlreadyRecorded if the (squad, event, place) row exists.
func (r *PostgresVenueRepository) RecordVisit(ctx context.Context, cmd domain.RecordVisitCommand, snapshot domain.VenueSnapshot) error {
	err := r.queries.RecordVisit(ctx, db.RecordVisitParams{
		SquadID:         cmd.SquadID,
		EventID:         cmd.EventID,
		PlaceID:         cmd.PlaceID,
		Name:            snapshot.Name,
		VisitedAt:       pgtype.Timestamptz{Time: cmd.VisitedAt, Valid: true},
		AvgSpendInPaise: pgtype.Int8{Int64: cmd.AvgSpendInPaise, Valid: cmd.AvgSpendInPaise > 0},
		Address:         snapshot.Address,
		Rating:          snapshot.AverageRating,
		RatingCount:     int32(snapshot.RatingCount),
		PriceLevel:      int32(snapshot.PriceLevel),
		PhotoRef:        snapshot.Photo.Reference,
		PhotoWidth:      int32(snapshot.Photo.Width),
		PhotoHeight:     int32(snapshot.Photo.Height),
	})
	if err != nil {
		if postgres.IsUniqueViolation(err) {
			return domainerr.New(domain.ErrVisitAlreadyRecorded, cmd.EventID+"/"+cmd.PlaceID)
		}

		return err
	}

	return nil
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
			PlaceID:         row.PlaceID,
			Name:            row.Name,
			Address:         row.Address,
			Photo:           photoRef(row.PhotoRef, row.PhotoWidth, row.PhotoHeight),
			VisitCount:      row.VisitCount,
			LastVisitedAt:   row.LastVisitedAt.Time,
			AvgSpendInPaise: row.AvgSpendInPaise,
		})
	}

	return result, nil
}

// photoRef builds a domain.PhotoRef from the photo_ref, photo_width and
// photo_height columns.
func photoRef(reference string, width, height int32) domain.PhotoRef {
	return domain.PhotoRef{Reference: reference, Width: int(width), Height: int(height)}
}
