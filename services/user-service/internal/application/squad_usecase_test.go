package application

import (
	"context"
	"testing"

	"github.com/SamPariat/roundup/services/user-service/internal/domain"
	"github.com/SamPariat/roundup/shared/domainerr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// squad_usecase_test.go tests SquadUseCase with a hand-rolled fake for the
// SquadRepository port, in the same style as fakeUserRepo.

type fakeSquadRepo struct {
	createFn       func(ctx context.Context, cmd domain.CreateSquadCommand) (*domain.Squad, error)
	getForMemberFn func(ctx context.Context, squadID, userID string) (*domain.Squad, error)
	listMembersFn  func(ctx context.Context, squadID string) ([]domain.SquadMember, error)
	getMemberFn    func(ctx context.Context, squadID, userID string) (*domain.SquadMember, error)
	addMemberFn    func(ctx context.Context, squadID, userID string, limit int) (*domain.SquadMember, error)
	removeMemberFn func(ctx context.Context, squadID, userID string) (bool, error)
	changeRoleFn   func(ctx context.Context, squadID, userID string, role domain.Role) error
}

func (f *fakeSquadRepo) Create(ctx context.Context, cmd domain.CreateSquadCommand) (*domain.Squad, error) {
	if f.createFn != nil {
		return f.createFn(ctx, cmd)
	}

	return nil, nil
}

func (f *fakeSquadRepo) GetForMember(ctx context.Context, squadID, userID string) (*domain.Squad, error) {
	if f.getForMemberFn != nil {
		return f.getForMemberFn(ctx, squadID, userID)
	}

	return nil, nil
}

func (f *fakeSquadRepo) ListForUser(_ context.Context, _ string) ([]domain.SquadSummary, error) {
	return nil, nil
}

func (f *fakeSquadRepo) ListMembers(ctx context.Context, squadID string) ([]domain.SquadMember, error) {
	if f.listMembersFn != nil {
		return f.listMembersFn(ctx, squadID)
	}

	return nil, nil
}

func (f *fakeSquadRepo) GetMember(ctx context.Context, squadID, userID string) (*domain.SquadMember, error) {
	if f.getMemberFn != nil {
		return f.getMemberFn(ctx, squadID, userID)
	}

	return nil, nil
}

func (f *fakeSquadRepo) AddMember(ctx context.Context, squadID, userID string, limit int) (*domain.SquadMember, error) {
	if f.addMemberFn != nil {
		return f.addMemberFn(ctx, squadID, userID, limit)
	}

	return nil, nil
}

func (f *fakeSquadRepo) RemoveMember(ctx context.Context, squadID, userID string) (bool, error) {
	if f.removeMemberFn != nil {
		return f.removeMemberFn(ctx, squadID, userID)
	}

	return false, nil
}

func (f *fakeSquadRepo) ChangeRole(ctx context.Context, squadID, userID string, role domain.Role) error {
	if f.changeRoleFn != nil {
		return f.changeRoleFn(ctx, squadID, userID, role)
	}

	return nil
}

const testSquadID = "squad-1"

// memberAs returns a getMemberFn that reports the caller as a member with role.
func memberAs(role domain.Role) func(context.Context, string, string) (*domain.SquadMember, error) {
	return func(_ context.Context, _, userID string) (*domain.SquadMember, error) {
		return &domain.SquadMember{UserID: userID, Role: role}, nil
	}
}

func notMember(_ context.Context, squadID, _ string) (*domain.SquadMember, error) {
	return nil, domainerr.New(domain.ErrSquadNotFound, squadID)
}

func TestCreateSquadRejectsBlankName(t *testing.T) {
	uc := NewSquadUseCase(&fakeSquadRepo{}, &fakeUserRepo{})

	_, err := uc.Create(t.Context(), domain.CreateSquadCommand{CreatorID: "u1", Name: "   "})

	assert.ErrorIs(t, err, domain.ErrInvalidSquadName)
}

func TestCreateSquadTrimsName(t *testing.T) {
	var got domain.CreateSquadCommand

	squads := &fakeSquadRepo{
		createFn: func(_ context.Context, cmd domain.CreateSquadCommand) (*domain.Squad, error) {
			got = cmd

			return &domain.Squad{Name: cmd.Name}, nil
		},
	}

	_, err := NewSquadUseCase(squads, &fakeUserRepo{}).Create(t.Context(), domain.CreateSquadCommand{CreatorID: "u1", Name: "  Friday crew "})
	require.NoError(t, err)

	assert.Equal(t, "Friday crew", got.Name)
}

func TestGetSquadForNonMember(t *testing.T) {
	squads := &fakeSquadRepo{
		getForMemberFn: func(_ context.Context, squadID, _ string) (*domain.Squad, error) {
			return nil, domainerr.New(domain.ErrSquadNotFound, squadID)
		},
		listMembersFn: func(_ context.Context, _ string) ([]domain.SquadMember, error) {
			t.Fatal("members must not be listed for a non-member")

			return nil, nil
		},
	}

	_, err := NewSquadUseCase(squads, &fakeUserRepo{}).Get(t.Context(), testSquadID, "stranger")

	assert.ErrorIs(t, err, domain.ErrSquadNotFound)
}

func TestAddMember(t *testing.T) {
	tests := []struct {
		name        string
		getMember   func(context.Context, string, string) (*domain.SquadMember, error)
		phone       string
		getByPhone  func(context.Context, string) (*domain.User, error)
		wantErr     error
		wantAddedID string
	}{
		{
			name:        "AnyMemberCanAdd",
			getMember:   memberAs(domain.RoleMember),
			phone:       testPhone,
			getByPhone:  func(_ context.Context, _ string) (*domain.User, error) { return &domain.User{ID: "friend"}, nil },
			wantAddedID: "friend",
		},
		{name: "NonMemberCannotAdd", getMember: notMember, phone: testPhone, wantErr: domain.ErrSquadNotFound},
		{name: "InvalidPhone", getMember: memberAs(domain.RoleMember), phone: "12345", wantErr: domain.ErrInvalidPhone},
		{name: "UnknownPhone", getMember: memberAs(domain.RoleMember), phone: testPhone, getByPhone: notFound, wantErr: domain.ErrUserNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var addedID string

			var addedLimit int

			squads := &fakeSquadRepo{
				getMemberFn: tt.getMember,
				addMemberFn: func(_ context.Context, _, userID string, limit int) (*domain.SquadMember, error) {
					addedID, addedLimit = userID, limit

					return &domain.SquadMember{UserID: userID, Role: domain.RoleMember}, nil
				},
			}

			_, err := NewSquadUseCase(squads, &fakeUserRepo{getByPhoneFn: tt.getByPhone}).AddMember(t.Context(), domain.AddMemberCommand{
				SquadID: testSquadID,
				ActorID: "actor",
				Phone:   tt.phone,
			})

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				assert.Empty(t, addedID, "nobody must be added on error")

				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantAddedID, addedID)
			assert.Equal(t, domain.MaxSquadMembers, addedLimit)
		})
	}
}

func TestRemoveMemberChecksPermission(t *testing.T) {
	tests := []struct {
		name      string
		getMember func(context.Context, string, string) (*domain.SquadMember, error)
		target    string
		wantErr   error
	}{
		{name: "AdminRemovesOther", getMember: memberAs(domain.RoleAdmin), target: "other"},
		{name: "MemberLeaves", getMember: memberAs(domain.RoleMember), target: "actor"},
		{name: "MemberRemovesOther", getMember: memberAs(domain.RoleMember), target: "other", wantErr: domain.ErrNotAdmin},
		{name: "NonMember", getMember: notMember, target: "other", wantErr: domain.ErrSquadNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			removed := false

			squads := &fakeSquadRepo{
				getMemberFn: tt.getMember,
				removeMemberFn: func(_ context.Context, _, _ string) (bool, error) {
					removed = true

					return false, nil
				},
			}

			_, err := NewSquadUseCase(squads, &fakeUserRepo{}).RemoveMember(t.Context(), domain.RemoveMemberCommand{
				SquadID: testSquadID,
				ActorID: "actor",
				UserID:  tt.target,
			})

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				assert.False(t, removed, "nobody must be removed when the check fails")

				return
			}

			require.NoError(t, err)
			assert.True(t, removed)
		})
	}
}

func TestChangeRole(t *testing.T) {
	tests := []struct {
		name      string
		getMember func(context.Context, string, string) (*domain.SquadMember, error)
		role      domain.Role
		wantErr   error
	}{
		{name: "AdminPromotes", getMember: memberAs(domain.RoleAdmin), role: domain.RoleAdmin},
		{name: "MemberCannotChangeRoles", getMember: memberAs(domain.RoleMember), role: domain.RoleAdmin, wantErr: domain.ErrNotAdmin},
		{name: "UnknownRole", getMember: memberAs(domain.RoleAdmin), role: "OWNER", wantErr: domain.ErrInvalidRole},
		{name: "NonMember", getMember: notMember, role: domain.RoleAdmin, wantErr: domain.ErrSquadNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			changed := false

			squads := &fakeSquadRepo{
				getMemberFn: tt.getMember,
				changeRoleFn: func(_ context.Context, _, _ string, _ domain.Role) error {
					changed = true

					return nil
				},
			}

			err := NewSquadUseCase(squads, &fakeUserRepo{}).ChangeRole(t.Context(), domain.ChangeRoleCommand{
				SquadID: testSquadID,
				ActorID: "actor",
				UserID:  "other",
				Role:    tt.role,
			})

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				assert.False(t, changed)

				return
			}

			require.NoError(t, err)
			assert.True(t, changed)
		})
	}
}
