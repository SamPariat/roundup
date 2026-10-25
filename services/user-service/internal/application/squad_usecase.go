package application

import (
	"context"
	"strings"

	"github.com/SamPariat/roundup/services/user-service/internal/domain"
	"github.com/SamPariat/roundup/shared/domainerr"
	"github.com/SamPariat/roundup/shared/pagination"
	"github.com/SamPariat/roundup/shared/phone"
)

// SquadUseCase handles creating squads and managing who is in them.
// Which role may take which action comes from the authorization policy
// (domain.Authorizer, see docs/PERMISSIONS.md); the same policy builds the
// permissions shown to clients. Rules that must hold under concurrency (the member cap, admin succession,
// deleting an empty squad) are enforced by the repository in a transaction.
type SquadUseCase struct {
	squads domain.SquadRepository
	users  domain.UserRepository
	authz  domain.Authorizer
}

// NewSquadUseCase constructs a SquadUseCase with the given repositories and
// authorizer.
func NewSquadUseCase(squads domain.SquadRepository, users domain.UserRepository, authz domain.Authorizer) *SquadUseCase {
	return &SquadUseCase{squads: squads, users: users, authz: authz}
}

// Create makes a squad with the creator as its first admin. Returns
// domain.ErrInvalidSquadName if the name is blank.
func (u *SquadUseCase) Create(ctx context.Context, cmd domain.CreateSquadCommand) (*domain.Squad, error) {
	cmd.Name = strings.TrimSpace(cmd.Name)
	if cmd.Name == "" {
		return nil, domain.ErrInvalidSquadName
	}

	return u.squads.Create(ctx, cmd)
}

// List returns one page of the squads userID belongs to, newest first, and
// the next page's cursor ("" on the last page).
func (u *SquadUseCase) List(ctx context.Context, userID string, page pagination.Page) ([]domain.SquadSummary, string, error) {
	return u.squads.ListForUser(ctx, userID, page)
}

// Get returns the squad with its members, the caller's own membership and
// what they may do. Returns domain.ErrSquadNotFound if the squad does not
// exist or userID is not a member.
func (u *SquadUseCase) Get(ctx context.Context, squadID, userID string) (*domain.SquadDetail, error) {
	squad, err := u.squads.GetForMember(ctx, squadID, userID)
	if err != nil {
		return nil, err
	}

	members, err := u.squads.ListMembers(ctx, squadID)
	if err != nil {
		return nil, err
	}

	me, ok := findMember(members, userID)
	if !ok {
		return nil, domainerr.New(domain.ErrSquadNotFound, squadID)
	}

	return &domain.SquadDetail{
		Squad:       *squad,
		Members:     members,
		Me:          me,
		Permissions: domain.PermissionsFor(u.authz, me, len(members)),
	}, nil
}

// Update renames the squad and/or changes its picture. Returns
// domain.ErrSquadNotFound if the actor is not a member, domain.ErrForbidden
// if the policy does not let their role edit, and domain.ErrInvalidSquadName
// for a blank name.
func (u *SquadUseCase) Update(ctx context.Context, cmd domain.UpdateSquadCommand) (*domain.Squad, error) {
	actor, err := u.squads.GetMember(ctx, cmd.SquadID, cmd.ActorID)
	if err != nil {
		return nil, err
	}

	if err := u.authorize(*actor, domain.ActionEditSquad); err != nil {
		return nil, err
	}

	if cmd.Name != nil {
		name := strings.TrimSpace(*cmd.Name)
		if name == "" {
			return nil, domain.ErrInvalidSquadName
		}

		cmd.Name = &name
	}

	return u.squads.Update(ctx, cmd)
}

// AddMember adds the Roundup user with cmd.Phone to the squad. Returns
// domain.ErrSquadNotFound if the actor is not a member, domain.ErrForbidden if
// the policy does not let their role add people, domain.ErrInvalidPhone or
// domain.ErrUserNotFound for a bad or unknown number, and
// domain.ErrAlreadyMember or domain.ErrSquadFull from the repository.
func (u *SquadUseCase) AddMember(ctx context.Context, cmd domain.AddMemberCommand) (*domain.SquadMember, error) {
	actor, err := u.squads.GetMember(ctx, cmd.SquadID, cmd.ActorID)
	if err != nil {
		return nil, err
	}

	if err := u.authorize(*actor, domain.ActionAddMember); err != nil {
		return nil, err
	}

	if !phone.IsValid(cmd.Phone) {
		return nil, domain.ErrInvalidPhone
	}

	user, err := u.users.GetByPhone(ctx, cmd.Phone)
	if err != nil {
		return nil, err
	}

	return u.squads.AddMember(ctx, cmd.SquadID, user.ID, domain.MaxSquadMembers)
}

// RemoveMember removes cmd.UserID from the squad; removing yourself is
// leaving. squadDeleted reports whether that emptied and deleted the squad.
// Removing yourself is domain.ActionLeave; removing anyone else is
// domain.ActionRemoveMember. Returns domain.ErrSquadNotFound if the actor is
// not a member and domain.ErrForbidden if the policy does not allow it.
func (u *SquadUseCase) RemoveMember(ctx context.Context, cmd domain.RemoveMemberCommand) (squadDeleted bool, err error) {
	actor, err := u.squads.GetMember(ctx, cmd.SquadID, cmd.ActorID)
	if err != nil {
		return false, err
	}

	action := domain.ActionRemoveMember
	if cmd.UserID == cmd.ActorID {
		action = domain.ActionLeave
	}

	if err := u.authorize(*actor, action); err != nil {
		return false, err
	}

	return u.squads.RemoveMember(ctx, cmd.SquadID, cmd.UserID)
}

// ChangeRole sets a member's role. Returns domain.ErrSquadNotFound if the
// actor is not a member, domain.ErrForbidden if the policy does not let their
// role change roles,
// domain.ErrInvalidRole for an unknown role, and domain.ErrLastAdmin if the
// squad would be left without an admin.
func (u *SquadUseCase) ChangeRole(ctx context.Context, cmd domain.ChangeRoleCommand) error {
	actor, err := u.squads.GetMember(ctx, cmd.SquadID, cmd.ActorID)
	if err != nil {
		return err
	}

	if err := u.authorize(*actor, domain.ActionChangeRole); err != nil {
		return err
	}

	if !cmd.Role.IsValid() {
		return domain.ErrInvalidRole
	}

	return u.squads.ChangeRole(ctx, cmd.SquadID, cmd.UserID, cmd.Role)
}

// authorize returns domain.ErrForbidden unless the policy lets actor's role
// take action.
func (u *SquadUseCase) authorize(actor domain.SquadMember, action domain.Action) error {
	if !u.authz.Can(actor.Role, action) {
		return domainerr.New(domain.ErrForbidden, string(action))
	}

	return nil
}

// findMember returns the member with userID from members.
func findMember(members []domain.SquadMember, userID string) (domain.SquadMember, bool) {
	for _, m := range members {
		if m.UserID == userID {
			return m, true
		}
	}

	return domain.SquadMember{}, false
}
