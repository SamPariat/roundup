package persistence

import (
	"context"
	"strconv"
	"time"

	"github.com/SamPariat/roundup/services/user-service/internal/adapters/persistence/db"
	"github.com/SamPariat/roundup/services/user-service/internal/domain"
	"github.com/SamPariat/roundup/shared/domainerr"
	"github.com/SamPariat/roundup/shared/pagination"
	"github.com/SamPariat/roundup/shared/postgres"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresSquadRepository implements domain.SquadRepository on top of the
// sqlc-generated queries. Membership changes lock the squad row first, so the
// admin-succession and member-cap rules hold under concurrent requests.
//
// IDs that are not valid UUIDs cannot exist in the tables, so they map to the
// not-found sentinel for the operation instead of a parse error.
type PostgresSquadRepository struct {
	pool    *pgxpool.Pool
	queries *db.Queries
}

// squadCursor is the position after which the next page of a squad list
// starts: the last squad's (created_at, id), the list's sort key.
type squadCursor struct {
	CreatedAt time.Time `json:"createdAt"`
	ID        uuid.UUID `json:"id"`
}

// NewPostgresSquadRepository constructs a PostgresSquadRepository from a pgxpool.
func NewPostgresSquadRepository(pool *pgxpool.Pool) *PostgresSquadRepository {
	return &PostgresSquadRepository{pool: pool, queries: db.New(pool)}
}

// Create inserts the squad and makes its creator the first admin.
func (r *PostgresSquadRepository) Create(ctx context.Context, cmd domain.CreateSquadCommand) (*domain.Squad, error) {
	creator, err := uuid.Parse(cmd.CreatorID)
	if err != nil {
		return nil, domainerr.New(domain.ErrUserNotFound, cmd.CreatorID)
	}

	// An empty avatar is stored as NULL.
	var avatar *string
	if cmd.AvatarURL != "" {
		avatar = &cmd.AvatarURL
	}

	var squad db.Squad

	err = postgres.InTx(ctx, r.pool, func(tx pgx.Tx) error {
		q := r.queries.WithTx(tx)

		squad, err = q.CreateSquad(ctx, db.CreateSquadParams{
			Name:      cmd.Name,
			AvatarUrl: avatar,
			CreatedBy: creator,
		})
		if err != nil {
			return err
		}

		_, err = q.AddSquadMember(ctx, db.AddSquadMemberParams{
			SquadID: squad.ID,
			UserID:  creator,
			Role:    string(domain.RoleAdmin),
		})

		return err
	})
	if err != nil {
		return nil, err
	}

	return toDomainSquad(squad), nil
}

// GetForMember returns the squad if userID is a member.
func (r *PostgresSquadRepository) GetForMember(ctx context.Context, squadID, userID string) (*domain.Squad, error) {
	sid, uid, err := parseIDs(squadID, userID)
	if err != nil {
		return nil, domainerr.New(domain.ErrSquadNotFound, squadID)
	}

	row, err := r.queries.GetSquadForMember(ctx, db.GetSquadForMemberParams{ID: sid, UserID: uid})
	if err != nil {
		return nil, postgres.NoRowsAs(err, domain.ErrSquadNotFound, squadID)
	}

	return toDomainSquad(row), nil
}

// ListForUser returns the squads userID belongs to, newest first.
func (r *PostgresSquadRepository) ListForUser(ctx context.Context, userID string, page pagination.Page) ([]domain.SquadSummary, string, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, "", domainerr.New(domain.ErrUserNotFound, userID)
	}

	params := db.ListSquadsForUserParams{UserID: uid, Lim: int32(page.Limit + 1)}

	if page.Cursor != "" {
		var after squadCursor
		if err := pagination.Decode(page.Cursor, &after); err != nil {
			return nil, "", err
		}

		params.CursorAt = postgres.Timestamptz(after.CreatedAt)
		params.CursorID = pgtype.UUID{Bytes: after.ID, Valid: true}
	}

	rows, err := r.queries.ListSquadsForUser(ctx, params)
	if err != nil {
		return nil, "", err
	}

	rows, more := pagination.Trim(rows, page.Limit)

	squads := make([]domain.SquadSummary, 0, len(rows))
	for _, row := range rows {
		squads = append(squads, domain.SquadSummary{
			Squad: *toDomainSquad(db.Squad{
				ID:        row.ID,
				Name:      row.Name,
				AvatarUrl: row.AvatarUrl,
				CreatedBy: row.CreatedBy,
				CreatedAt: row.CreatedAt,
			}),
			Role:        domain.Role(row.Role),
			MemberCount: int(row.MemberCount),
		})
	}

	if err := r.attachPreviews(ctx, squads); err != nil {
		return nil, "", err
	}

	if !more {
		return squads, "", nil
	}

	last := rows[len(rows)-1]

	return squads, pagination.Encode(squadCursor{CreatedAt: last.CreatedAt.Time, ID: last.ID}), nil
}

// ListMembers returns the squad's active members, earliest joiner first.
func (r *PostgresSquadRepository) ListMembers(ctx context.Context, squadID string) ([]domain.SquadMember, error) {
	sid, err := uuid.Parse(squadID)
	if err != nil {
		return nil, domainerr.New(domain.ErrSquadNotFound, squadID)
	}

	rows, err := r.queries.ListSquadMembers(ctx, sid)
	if err != nil {
		return nil, err
	}

	members := make([]domain.SquadMember, 0, len(rows))
	for _, row := range rows {
		members = append(members, toDomainMember(db.GetSquadMemberRow(row)))
	}

	return members, nil
}

// GetMember returns userID's membership in the squad.
func (r *PostgresSquadRepository) GetMember(ctx context.Context, squadID, userID string) (*domain.SquadMember, error) {
	sid, uid, err := parseIDs(squadID, userID)
	if err != nil {
		return nil, domainerr.New(domain.ErrSquadNotFound, squadID)
	}

	row, err := r.queries.GetSquadMember(ctx, db.GetSquadMemberParams{SquadID: sid, UserID: uid})
	if err != nil {
		return nil, postgres.NoRowsAs(err, domain.ErrSquadNotFound, squadID)
	}

	member := toDomainMember(row)

	return &member, nil
}

// AddMember adds userID as a member, unless the squad already has limit members.
func (r *PostgresSquadRepository) AddMember(ctx context.Context, squadID, userID string, limit int) (*domain.SquadMember, error) {
	sid, uid, err := parseIDs(squadID, userID)
	if err != nil {
		return nil, domainerr.New(domain.ErrSquadNotFound, squadID)
	}

	var row db.GetSquadMemberRow

	err = postgres.InTx(ctx, r.pool, func(tx pgx.Tx) error {
		q := r.queries.WithTx(tx)

		if _, err := q.LockSquad(ctx, sid); err != nil {
			return postgres.NoRowsAs(err, domain.ErrSquadNotFound, squadID)
		}

		count, err := q.CountSquadMembers(ctx, sid)
		if err != nil {
			return err
		}

		if count >= int64(limit) {
			return domainerr.New(domain.ErrSquadFull, strconv.Itoa(limit)+" members")
		}

		_, err = q.AddSquadMember(ctx, db.AddSquadMemberParams{SquadID: sid, UserID: uid, Role: string(domain.RoleMember)})
		if err != nil {
			if postgres.IsUniqueViolation(err) {
				return domainerr.New(domain.ErrAlreadyMember, userID)
			}

			return err
		}

		row, err = q.GetSquadMember(ctx, db.GetSquadMemberParams{SquadID: sid, UserID: uid})

		return err
	})
	if err != nil {
		return nil, err
	}

	member := toDomainMember(row)

	return &member, nil
}

// RemoveMember removes userID, then promotes the earliest joiner if no admin
// remains, or deletes the squad if no member remains.
func (r *PostgresSquadRepository) RemoveMember(ctx context.Context, squadID, userID string) (bool, error) {
	sid, uid, err := parseIDs(squadID, userID)
	if err != nil {
		return false, domainerr.New(domain.ErrMemberNotFound, userID)
	}

	deleted := false

	err = postgres.InTx(ctx, r.pool, func(tx pgx.Tx) error {
		q := r.queries.WithTx(tx)

		if _, err := q.LockSquad(ctx, sid); err != nil {
			return postgres.NoRowsAs(err, domain.ErrMemberNotFound, userID)
		}

		affected, err := q.RemoveSquadMember(ctx, db.RemoveSquadMemberParams{SquadID: sid, UserID: uid})
		if err != nil {
			return err
		}

		if affected == 0 {
			return domainerr.New(domain.ErrMemberNotFound, userID)
		}

		members, err := q.CountSquadMembers(ctx, sid)
		if err != nil {
			return err
		}

		if members == 0 {
			deleted = true

			return q.DeleteSquad(ctx, sid)
		}

		// Someone is left; if no admin is, the earliest joiner takes over.
		admins, err := q.CountSquadAdmins(ctx, sid)
		if err != nil || admins > 0 {
			return err
		}

		return q.PromoteEarliestMember(ctx, sid)
	})

	return deleted, err
}

// ChangeRole sets userID's role, refusing to leave the squad without an admin.
func (r *PostgresSquadRepository) ChangeRole(ctx context.Context, squadID, userID string, role domain.Role) error {
	sid, uid, err := parseIDs(squadID, userID)
	if err != nil {
		return domainerr.New(domain.ErrMemberNotFound, userID)
	}

	return postgres.InTx(ctx, r.pool, func(tx pgx.Tx) error {
		q := r.queries.WithTx(tx)

		if _, err := q.LockSquad(ctx, sid); err != nil {
			return postgres.NoRowsAs(err, domain.ErrMemberNotFound, userID)
		}

		affected, err := q.UpdateSquadMemberRole(ctx, db.UpdateSquadMemberRoleParams{
			SquadID: sid,
			UserID:  uid,
			Role:    string(role),
		})
		if err != nil {
			return err
		}

		if affected == 0 {
			return domainerr.New(domain.ErrMemberNotFound, userID)
		}

		admins, err := q.CountSquadAdmins(ctx, sid)
		if err != nil {
			return err
		}

		if admins == 0 {
			// Returning an error rolls the role change back.
			return domainerr.New(domain.ErrLastAdmin, squadID)
		}

		return nil
	})
}

// Update sets the squad's name and/or avatar. A nil field is left unchanged.
func (r *PostgresSquadRepository) Update(ctx context.Context, cmd domain.UpdateSquadCommand) (*domain.Squad, error) {
	sid, err := uuid.Parse(cmd.SquadID)
	if err != nil {
		return nil, domainerr.New(domain.ErrSquadNotFound, cmd.SquadID)
	}

	row, err := r.queries.UpdateSquad(ctx, db.UpdateSquadParams{ID: sid, Name: cmd.Name, AvatarUrl: cmd.AvatarURL})
	if err != nil {
		return nil, postgres.NoRowsAs(err, domain.ErrSquadNotFound, cmd.SquadID)
	}

	return toDomainSquad(row), nil
}

// attachPreviews fills each squad's MemberPreview with one query for all of
// them, so listing n squads costs two queries instead of n+1.
func (r *PostgresSquadRepository) attachPreviews(ctx context.Context, squads []domain.SquadSummary) error {
	if len(squads) == 0 {
		return nil
	}

	ids := make([]uuid.UUID, 0, len(squads))
	index := make(map[string]int, len(squads))

	for i, s := range squads {
		ids = append(ids, uuid.MustParse(s.ID))
		index[s.ID] = i
		squads[i].MemberPreview = []domain.SquadMember{}
	}

	rows, err := r.queries.ListMemberPreviews(ctx, db.ListMemberPreviewsParams{SquadIds: ids, PerSquad: domain.MemberPreviewSize})
	if err != nil {
		return err
	}

	for _, row := range rows {
		i := index[row.SquadID.String()]
		squads[i].MemberPreview = append(squads[i].MemberPreview, toDomainMember(db.GetSquadMemberRow{
			UserID:    row.UserID,
			Role:      row.Role,
			JoinedAt:  row.JoinedAt,
			Name:      row.Name,
			AvatarUrl: row.AvatarUrl,
		}))
	}

	return nil
}

// parseIDs parses a squad ID and a user ID, failing if either is not a UUID.
func parseIDs(squadID, userID string) (uuid.UUID, uuid.UUID, error) {
	sid, err := uuid.Parse(squadID)
	if err != nil {
		return uuid.Nil, uuid.Nil, err
	}

	uid, err := uuid.Parse(userID)
	if err != nil {
		return uuid.Nil, uuid.Nil, err
	}

	return sid, uid, nil
}

// toDomainSquad maps a sqlc row to a domain.Squad, turning a NULL avatar into an empty string.
func toDomainSquad(row db.Squad) *domain.Squad {
	squad := &domain.Squad{
		ID:        row.ID.String(),
		Name:      row.Name,
		CreatedBy: row.CreatedBy.String(),
		CreatedAt: row.CreatedAt.Time,
	}
	if row.AvatarUrl != nil {
		squad.AvatarURL = *row.AvatarUrl
	}

	return squad
}

// toDomainMember maps a sqlc row to a domain.SquadMember, turning NULL profile
// columns into empty strings.
func toDomainMember(row db.GetSquadMemberRow) domain.SquadMember {
	member := domain.SquadMember{
		UserID:   row.UserID.String(),
		Role:     domain.Role(row.Role),
		JoinedAt: row.JoinedAt.Time,
	}
	if row.Name != nil {
		member.Name = *row.Name
	}

	if row.AvatarUrl != nil {
		member.AvatarURL = *row.AvatarUrl
	}

	return member
}
