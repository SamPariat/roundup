package application

import (
	"context"
	"strings"

	"github.com/SamPariat/roundup/services/user-service/internal/domain"
	"github.com/SamPariat/roundup/shared/phone"
)

// SquadUseCase handles creating squads and managing who is in them.
// Permission rules live in the domain (domain.CanRemove, domain.CanChangeRole);
// rules that must hold under concurrency (the member cap, admin succession,
// deleting an empty squad) are enforced by the repository in a transaction.
type SquadUseCase struct {
	squads domain.SquadRepository
	users  domain.UserRepository
}

// NewSquadUseCase constructs a SquadUseCase with the given repositories.
func NewSquadUseCase(squads domain.SquadRepository, users domain.UserRepository) *SquadUseCase {
	return &SquadUseCase{squads: squads, users: users}
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

// List returns the squads userID belongs to, newest first.
func (u *SquadUseCase) List(ctx context.Context, userID string) ([]domain.SquadSummary, error) {
	return u.squads.ListForUser(ctx, userID)
}

// Get returns the squad with its members. Returns domain.ErrSquadNotFound if
// the squad does not exist or userID is not a member.
func (u *SquadUseCase) Get(ctx context.Context, squadID, userID string) (*domain.SquadDetail, error) {
	squad, err := u.squads.GetForMember(ctx, squadID, userID)
	if err != nil {
		return nil, err
	}

	members, err := u.squads.ListMembers(ctx, squadID)
	if err != nil {
		return nil, err
	}

	return &domain.SquadDetail{Squad: *squad, Members: members}, nil
}

// AddMember adds the Roundup user with cmd.Phone to the squad. Any member may
// add people. Returns domain.ErrSquadNotFound if the actor is not a member,
// domain.ErrInvalidPhone or domain.ErrUserNotFound for a bad or unknown
// number, and domain.ErrAlreadyMember or domain.ErrSquadFull from the
// repository.
func (u *SquadUseCase) AddMember(ctx context.Context, cmd domain.AddMemberCommand) (*domain.SquadMember, error) {
	if _, err := u.squads.GetMember(ctx, cmd.SquadID, cmd.ActorID); err != nil {
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
// Returns domain.ErrSquadNotFound if the actor is not a member and
// domain.ErrNotAdmin if a member tries to remove someone else.
func (u *SquadUseCase) RemoveMember(ctx context.Context, cmd domain.RemoveMemberCommand) (squadDeleted bool, err error) {
	actor, err := u.squads.GetMember(ctx, cmd.SquadID, cmd.ActorID)
	if err != nil {
		return false, err
	}

	if err := domain.CanRemove(*actor, cmd.UserID); err != nil {
		return false, err
	}

	return u.squads.RemoveMember(ctx, cmd.SquadID, cmd.UserID)
}

// ChangeRole sets a member's role. Returns domain.ErrSquadNotFound if the
// actor is not a member, domain.ErrNotAdmin if they are not an admin,
// domain.ErrInvalidRole for an unknown role, and domain.ErrLastAdmin if the
// squad would be left without an admin.
func (u *SquadUseCase) ChangeRole(ctx context.Context, cmd domain.ChangeRoleCommand) error {
	actor, err := u.squads.GetMember(ctx, cmd.SquadID, cmd.ActorID)
	if err != nil {
		return err
	}

	if err := domain.CanChangeRole(*actor); err != nil {
		return err
	}

	if !cmd.Role.IsValid() {
		return domain.ErrInvalidRole
	}

	return u.squads.ChangeRole(ctx, cmd.SquadID, cmd.UserID, cmd.Role)
}
