package persistence

import (
	"context"
	"time"

	"github.com/SamPariat/roundup/services/event-service/internal/adapters/persistence/db"
	"github.com/SamPariat/roundup/services/event-service/internal/domain"
	"github.com/SamPariat/roundup/shared/domainerr"
	"github.com/SamPariat/roundup/shared/pagination"
	"github.com/SamPariat/roundup/shared/postgres"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresEventRepository implements domain.EventRepository on top of the
// sqlc-generated queries. Status-dependent writes check the status in the
// same statement, so they report false instead of racing a concurrent change.
//
// IDs that are not valid UUIDs cannot exist in the tables, so they map to
// ErrEventNotFound instead of a parse error.
type PostgresEventRepository struct {
	queries *db.Queries
}

// eventCursor is the position after which the next page of a squad's events
// starts: the last event's (closed, starts_at, id), the list's sort key.
type eventCursor struct {
	Closed   bool      `json:"closed"`
	StartsAt time.Time `json:"startsAt"`
	ID       uuid.UUID `json:"id"`
}

// NewPostgresEventRepository constructs a PostgresEventRepository from a pgxpool.
func NewPostgresEventRepository(pool *pgxpool.Pool) *PostgresEventRepository {
	return &PostgresEventRepository{queries: db.New(pool)}
}

// Create stores a new DRAFT event.
func (r *PostgresEventRepository) Create(ctx context.Context, e domain.NewEvent) (*domain.Event, error) {
	squadID, err := uuid.Parse(e.SquadID)
	if err != nil {
		return nil, domainerr.New(domain.ErrSquadNotFound, e.SquadID)
	}

	createdBy, err := uuid.Parse(e.CreatedBy)
	if err != nil {
		return nil, domainerr.New(domain.ErrSquadNotFound, e.SquadID)
	}

	row, err := r.queries.CreateEvent(ctx, db.CreateEventParams{
		SquadID:        squadID,
		Name:           e.Name,
		Type:           string(e.Type),
		CreatedBy:      createdBy,
		StartsAt:       postgres.Timestamptz(e.Schedule.StartsAt),
		VotingDeadline: postgres.Timestamptz(e.Schedule.VotingDeadline),
	})
	if err != nil {
		return nil, err
	}

	return &domain.Event{
		ID:             row.ID.String(),
		SquadID:        e.SquadID,
		Name:           e.Name,
		Type:           e.Type,
		Status:         domain.StatusDraft,
		CreatedBy:      e.CreatedBy,
		StartsAt:       e.Schedule.StartsAt,
		VotingDeadline: e.Schedule.VotingDeadline,
		CreatedAt:      row.CreatedAt.Time,
	}, nil
}

// Get returns the event with its winner, or ErrEventNotFound.
func (r *PostgresEventRepository) Get(ctx context.Context, eventID string) (*domain.Event, error) {
	id, err := uuid.Parse(eventID)
	if err != nil {
		return nil, domainerr.New(domain.ErrEventNotFound, eventID)
	}

	row, err := r.queries.GetEvent(ctx, id)
	if err != nil {
		return nil, postgres.NoRowsAs(err, domain.ErrEventNotFound, eventID)
	}

	event := toDomainEvent(row)

	return &event, nil
}

// Update saves the name, type and times if the event is still DRAFT or VOTING.
func (r *PostgresEventRepository) Update(ctx context.Context, e *domain.Event) (bool, error) {
	id, err := uuid.Parse(e.ID)
	if err != nil {
		return false, domainerr.New(domain.ErrEventNotFound, e.ID)
	}

	n, err := r.queries.UpdateEvent(ctx, db.UpdateEventParams{
		ID:             id,
		Name:           e.Name,
		Type:           string(e.Type),
		StartsAt:       postgres.Timestamptz(e.StartsAt),
		VotingDeadline: postgres.Timestamptz(e.VotingDeadline),
	})

	return n == 1, err
}

// ListForSquad returns one page of the squad's events and the next page's
// cursor ("" on the last page).
func (r *PostgresEventRepository) ListForSquad(ctx context.Context, squadID string, page pagination.Page) ([]domain.Event, string, error) {
	sid, err := uuid.Parse(squadID)
	if err != nil {
		return nil, "", domainerr.New(domain.ErrSquadNotFound, squadID)
	}

	params := db.ListSquadEventsParams{SquadID: sid, Lim: int32(page.Limit + 1)}

	if page.Cursor != "" {
		var after eventCursor
		if err := pagination.Decode(page.Cursor, &after); err != nil {
			return nil, "", err
		}

		params.CursorClosed = &after.Closed
		params.CursorAt = postgres.Timestamptz(after.StartsAt)
		params.CursorID = pgtype.UUID{Bytes: after.ID, Valid: true}
	}

	rows, err := r.queries.ListSquadEvents(ctx, params)
	if err != nil {
		return nil, "", err
	}

	rows, more := pagination.Trim(rows, page.Limit)

	events := make([]domain.Event, 0, len(rows))
	for _, row := range rows {
		events = append(events, toDomainEvent(db.GetEventRow(row)))
	}

	if !more {
		return events, "", nil
	}

	last := events[len(events)-1]

	return events, pagination.Encode(eventCursor{
		Closed:   last.Status == domain.StatusClosed,
		StartsAt: last.StartsAt,
		ID:       uuid.MustParse(last.ID),
	}), nil
}

// CountOpen counts the squad's events that aren't CLOSED.
func (r *PostgresEventRepository) CountOpen(ctx context.Context, squadID string) (int, error) {
	sid, err := uuid.Parse(squadID)
	if err != nil {
		return 0, domainerr.New(domain.ErrSquadNotFound, squadID)
	}

	n, err := r.queries.CountOpenEvents(ctx, sid)

	return int(n), err
}

// RSVPs returns every answer to the given events.
func (r *PostgresEventRepository) RSVPs(ctx context.Context, eventIDs []string) ([]domain.RSVP, error) {
	rows, err := r.queries.ListRSVPs(ctx, parseUUIDs(eventIDs))
	if err != nil {
		return nil, err
	}

	rsvps := make([]domain.RSVP, 0, len(rows))
	for _, row := range rows {
		rsvps = append(rsvps, domain.RSVP{
			EventID: row.EventID.String(),
			UserID:  row.UserID.String(),
			Status:  domain.RSVPStatus(row.Status),
		})
	}

	return rsvps, nil
}

// Votes returns every vote cast in the given events.
func (r *PostgresEventRepository) Votes(ctx context.Context, eventIDs []string) ([]domain.Vote, error) {
	rows, err := r.queries.ListVotes(ctx, parseUUIDs(eventIDs))
	if err != nil {
		return nil, err
	}

	votes := make([]domain.Vote, 0, len(rows))
	for _, row := range rows {
		votes = append(votes, domain.Vote{
			EventID: row.EventID.String(),
			UserID:  row.UserID.String(),
			PlaceID: row.PlaceID,
		})
	}

	return votes, nil
}

// Nominations returns the event's suggested places, earliest first.
func (r *PostgresEventRepository) Nominations(ctx context.Context, eventID string) ([]domain.Nomination, error) {
	id, err := uuid.Parse(eventID)
	if err != nil {
		return nil, domainerr.New(domain.ErrEventNotFound, eventID)
	}

	rows, err := r.queries.ListNominations(ctx, id)
	if err != nil {
		return nil, err
	}

	nominations := make([]domain.Nomination, 0, len(rows))
	for _, row := range rows {
		nominations = append(nominations, domain.Nomination{
			PlaceID:     row.PlaceID,
			Name:        row.Name,
			Address:     row.Address,
			Lat:         row.Lat,
			Lng:         row.Lng,
			Rating:      row.Rating,
			PriceLevel:  int(row.PriceLevel),
			PhotoURL:    row.PhotoUrl,
			PhotoWidth:  int(row.PhotoWidth),
			PhotoHeight: int(row.PhotoHeight),
			NominatedBy: row.NominatedBy.String(),
			NominatedAt: row.NominatedAt.Time,
		})
	}

	return nominations, nil
}

// NominationCounts returns how many places each event has, by event ID.
// Events with none are absent from the map.
func (r *PostgresEventRepository) NominationCounts(ctx context.Context, eventIDs []string) (map[string]int, error) {
	rows, err := r.queries.CountNominations(ctx, parseUUIDs(eventIDs))
	if err != nil {
		return nil, err
	}

	counts := make(map[string]int, len(rows))
	for _, row := range rows {
		counts[row.EventID.String()] = int(row.Count)
	}

	return counts, nil
}

// SetRSVP stores or changes userID's answer unless the event is CLOSED.
func (r *PostgresEventRepository) SetRSVP(ctx context.Context, eventID, userID string, status domain.RSVPStatus) (bool, error) {
	eid, uid, err := parseIDs(eventID, userID)
	if err != nil {
		return false, err
	}

	n, err := r.queries.SetRSVP(ctx, db.SetRSVPParams{EventID: eid, UserID: uid, Status: string(status)})

	return n == 1, err
}

// AddNomination stores n if the event is DRAFT. Returns ErrAlreadyNominated
// if the place was already suggested.
func (r *PostgresEventRepository) AddNomination(ctx context.Context, eventID string, n domain.Nomination) (bool, error) {
	eid, uid, err := parseIDs(eventID, n.NominatedBy)
	if err != nil {
		return false, err
	}

	affected, err := r.queries.AddNomination(ctx, db.AddNominationParams{
		EventID:     eid,
		PlaceID:     n.PlaceID,
		Name:        n.Name,
		Address:     n.Address,
		Lat:         n.Lat,
		Lng:         n.Lng,
		Rating:      n.Rating,
		PriceLevel:  int32(n.PriceLevel),
		PhotoUrl:    n.PhotoURL,
		PhotoWidth:  int32(n.PhotoWidth),
		PhotoHeight: int32(n.PhotoHeight),
		NominatedBy: uid,
	})
	if err != nil {
		if postgres.IsUniqueViolation(err) {
			return false, domainerr.New(domain.ErrAlreadyNominated, n.PlaceID)
		}

		return false, err
	}

	return affected == 1, nil
}

// RemoveNomination deletes the place if the event is DRAFT.
func (r *PostgresEventRepository) RemoveNomination(ctx context.Context, eventID, placeID string) (bool, error) {
	id, err := uuid.Parse(eventID)
	if err != nil {
		return false, domainerr.New(domain.ErrEventNotFound, eventID)
	}

	n, err := r.queries.RemoveNomination(ctx, db.RemoveNominationParams{EventID: id, PlaceID: placeID})

	return n == 1, err
}

// CastVote stores or changes userID's vote if the event is VOTING and placeID
// is one of its nominations.
func (r *PostgresEventRepository) CastVote(ctx context.Context, eventID, userID, placeID string) (bool, error) {
	eid, uid, err := parseIDs(eventID, userID)
	if err != nil {
		return false, err
	}

	n, err := r.queries.CastVote(ctx, db.CastVoteParams{EventID: eid, UserID: uid, PlaceID: placeID})

	return n == 1, err
}

// StartVoting moves a DRAFT event with at least one nomination to VOTING.
func (r *PostgresEventRepository) StartVoting(ctx context.Context, eventID string) (bool, error) {
	return r.transition(eventID, func(id uuid.UUID) (int64, error) { return r.queries.StartVoting(ctx, id) })
}

// ConfirmVoting moves a VOTING event to CONFIRMED and sets its winner.
func (r *PostgresEventRepository) ConfirmVoting(ctx context.Context, eventID string) (bool, error) {
	return r.transition(eventID, func(id uuid.UUID) (int64, error) {
		return r.queries.ConfirmVoting(ctx, db.ConfirmVotingParams{EventID: pgtype.UUID{Bytes: id, Valid: true}})
	})
}

// ConfirmDue confirms every VOTING event whose deadline is at or before now.
func (r *PostgresEventRepository) ConfirmDue(ctx context.Context, now time.Time) (int, error) {
	n, err := r.queries.ConfirmVoting(ctx, db.ConfirmVotingParams{DueAt: postgres.Timestamptz(now)})

	return int(n), err
}

// StartNight moves a CONFIRMED event to ACTIVE.
func (r *PostgresEventRepository) StartNight(ctx context.Context, eventID string) (bool, error) {
	return r.transition(eventID, func(id uuid.UUID) (int64, error) { return r.queries.StartNight(ctx, id) })
}

// Close moves an ACTIVE event to CLOSED and returns it, or nil if it wasn't ACTIVE.
func (r *PostgresEventRepository) Close(ctx context.Context, eventID string) (*domain.Event, error) {
	id, err := uuid.Parse(eventID)
	if err != nil {
		return nil, domainerr.New(domain.ErrEventNotFound, eventID)
	}

	row, err := r.queries.CloseEvent(ctx, id)
	if err != nil {
		if postgres.IsNoRows(err) {
			return nil, nil
		}

		return nil, err
	}

	closed := closedEvent(db.CloseDueEventsRow(row))

	return &closed, nil
}

// CloseDue closes every CONFIRMED or ACTIVE event that started
// domain.AutoCloseAfter or more before now.
func (r *PostgresEventRepository) CloseDue(ctx context.Context, now time.Time) ([]domain.Event, error) {
	rows, err := r.queries.CloseDueEvents(ctx, db.CloseDueEventsParams{
		AutoCloseAfter: pgtype.Interval{Microseconds: domain.AutoCloseAfter.Microseconds(), Valid: true},
		DueAt:          postgres.Timestamptz(now),
	})
	if err != nil {
		return nil, err
	}

	events := make([]domain.Event, 0, len(rows))
	for _, row := range rows {
		events = append(events, closedEvent(row))
	}

	return events, nil
}

// transition runs a status-guarded update on eventID and reports whether it
// changed the row.
func (r *PostgresEventRepository) transition(eventID string, update func(uuid.UUID) (int64, error)) (bool, error) {
	id, err := uuid.Parse(eventID)
	if err != nil {
		return false, domainerr.New(domain.ErrEventNotFound, eventID)
	}

	n, err := update(id)

	return n == 1, err
}

// toDomainEvent maps an event row, with its winner's columns, to a domain.Event.
func toDomainEvent(row db.GetEventRow) domain.Event {
	event := domain.Event{
		ID:             row.ID.String(),
		SquadID:        row.SquadID.String(),
		Name:           row.Name,
		Type:           domain.Type(row.Type),
		Status:         domain.Status(row.Status),
		CreatedBy:      row.CreatedBy.String(),
		StartsAt:       row.StartsAt.Time,
		VotingDeadline: row.VotingDeadline.Time,
		CreatedAt:      row.CreatedAt.Time,
	}

	if row.ClosedAt.Valid {
		event.ClosedAt = &row.ClosedAt.Time
	}

	if row.WinnerPlaceID != nil {
		event.Winner = &domain.Nomination{
			PlaceID:     *row.WinnerPlaceID,
			Name:        *row.WinnerName,
			Address:     *row.WinnerAddress,
			Lat:         *row.WinnerLat,
			Lng:         *row.WinnerLng,
			Rating:      *row.WinnerRating,
			PriceLevel:  int(*row.WinnerPriceLevel),
			PhotoURL:    *row.WinnerPhotoUrl,
			PhotoWidth:  int(*row.WinnerPhotoWidth),
			PhotoHeight: int(*row.WinnerPhotoHeight),
			NominatedBy: uuid.UUID(row.WinnerNominatedBy.Bytes).String(),
			NominatedAt: row.WinnerNominatedAt.Time,
		}
	}

	return event
}

// closedEvent maps a just-closed row to the fields recording a visit needs.
// The winner carries only its place ID.
func closedEvent(row db.CloseDueEventsRow) domain.Event {
	event := domain.Event{
		ID:        row.ID.String(),
		SquadID:   row.SquadID.String(),
		Status:    domain.StatusClosed,
		CreatedBy: row.CreatedBy.String(),
		StartsAt:  row.StartsAt.Time,
	}

	if row.WinnerPlaceID != nil {
		event.Winner = &domain.Nomination{PlaceID: *row.WinnerPlaceID}
	}

	return event
}

// parseIDs parses an event ID and a user ID, mapping either failure to
// ErrEventNotFound.
func parseIDs(eventID, userID string) (uuid.UUID, uuid.UUID, error) {
	eid, err := uuid.Parse(eventID)
	if err != nil {
		return uuid.Nil, uuid.Nil, domainerr.New(domain.ErrEventNotFound, eventID)
	}

	uid, err := uuid.Parse(userID)
	if err != nil {
		return uuid.Nil, uuid.Nil, domainerr.New(domain.ErrEventNotFound, eventID)
	}

	return eid, uid, nil
}

// parseUUIDs keeps the IDs that parse; the others cannot match any row.
func parseUUIDs(ids []string) []uuid.UUID {
	parsed := make([]uuid.UUID, 0, len(ids))
	for _, id := range ids {
		if u, err := uuid.Parse(id); err == nil {
			parsed = append(parsed, u)
		}
	}

	return parsed
}
