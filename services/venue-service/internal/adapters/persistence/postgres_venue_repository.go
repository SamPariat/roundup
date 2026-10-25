package persistence

import (
	"context"
	"time"

	"github.com/SamPariat/roundup/services/venue-service/internal/adapters/persistence/db"
	"github.com/SamPariat/roundup/services/venue-service/internal/domain"
	"github.com/SamPariat/roundup/shared/domainerr"
	"github.com/SamPariat/roundup/shared/pagination"
	"github.com/SamPariat/roundup/shared/postgres"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresVenueRepository implements domain.VenueRepository on top of the
// sqlc-generated queries.
type PostgresVenueRepository struct {
	queries *db.Queries
}

// placeCursor is the position after which the next page of favorites or
// history starts: the last row's (time, place_id), both lists' sort key.
type placeCursor struct {
	At      time.Time `json:"at"`
	PlaceID string    `json:"placeID"`
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

// ListSquadFavorites returns one page of the squad's favorites, grouped per
// venue by the database and counting only saves by memberIDs, and the next
// page's cursor ("" on the last page).
func (r *PostgresVenueRepository) ListSquadFavorites(ctx context.Context, squadID string, memberIDs []string, page pagination.Page) ([]domain.FavoriteGroup, string, error) {
	cursorAt, cursorPlace, err := decodePlaceCursor(page)
	if err != nil {
		return nil, "", err
	}

	rows, err := r.queries.ListSquadFavorites(ctx, db.ListSquadFavoritesParams{
		SquadID:     squadID,
		MemberIds:   memberIDs,
		CursorAt:    cursorAt,
		CursorPlace: cursorPlace,
		Lim:         int32(page.Limit + 1),
	})
	if err != nil {
		return nil, "", err
	}

	rows, more := pagination.Trim(rows, page.Limit)

	result := make([]domain.FavoriteGroup, 0, len(rows))
	for _, row := range rows {
		result = append(result, domain.FavoriteGroup{
			PlaceID:       row.PlaceID,
			Name:          row.Name,
			Address:       row.Address,
			AverageRating: row.Rating,
			RatingCount:   int(row.RatingCount),
			PriceLevel:    int(row.PriceLevel),
			Photo:         photoRef(row.PhotoRef, row.PhotoWidth, row.PhotoHeight),
			SavedBy:       row.SavedBy,
			LastSavedAt:   row.LastSavedAt.Time,
		})
	}

	if !more {
		return result, "", nil
	}

	last := result[len(result)-1]

	return result, pagination.Encode(placeCursor{At: last.LastSavedAt, PlaceID: last.PlaceID}), nil
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
		VisitedAt:       postgres.Timestamptz(cmd.VisitedAt),
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

// GetVisitHistory returns one page of visit summaries per venue the squad
// visited, most recent first, and the next page's cursor ("" on the last page).
func (r *PostgresVenueRepository) GetVisitHistory(ctx context.Context, squadID string, page pagination.Page) ([]domain.VisitSummary, string, error) {
	cursorAt, cursorPlace, err := decodePlaceCursor(page)
	if err != nil {
		return nil, "", err
	}

	rows, err := r.queries.GetVisitHistory(ctx, db.GetVisitHistoryParams{
		SquadID:     squadID,
		CursorAt:    cursorAt,
		CursorPlace: cursorPlace,
		Lim:         int32(page.Limit + 1),
	})
	if err != nil {
		return nil, "", err
	}

	rows, more := pagination.Trim(rows, page.Limit)

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

	if !more {
		return result, "", nil
	}

	last := result[len(result)-1]

	return result, pagination.Encode(placeCursor{At: last.LastVisitedAt, PlaceID: last.PlaceID}), nil
}

// decodePlaceCursor turns page's cursor into the query's keyset arguments;
// both are NULL for the first page.
func decodePlaceCursor(page pagination.Page) (pgtype.Timestamptz, pgtype.Text, error) {
	if page.Cursor == "" {
		return pgtype.Timestamptz{}, pgtype.Text{}, nil
	}

	var after placeCursor
	if err := pagination.Decode(page.Cursor, &after); err != nil {
		return pgtype.Timestamptz{}, pgtype.Text{}, err
	}

	return postgres.Timestamptz(after.At), pgtype.Text{String: after.PlaceID, Valid: true}, nil
}

// photoRef builds a domain.PhotoRef from the photo_ref, photo_width and
// photo_height columns.
func photoRef(reference string, width, height int32) domain.PhotoRef {
	return domain.PhotoRef{Reference: reference, Width: int(width), Height: int(height)}
}
