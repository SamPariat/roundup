package application

import (
	"context"
	"testing"

	"github.com/SamPariat/roundup/services/user-service/internal/domain"
	"github.com/SamPariat/roundup/services/user-service/internal/policy"
	"github.com/SamPariat/roundup/shared/authz"
	"github.com/SamPariat/roundup/shared/domainerr"
	"github.com/SamPariat/roundup/shared/pagination"
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
	updateFn       func(ctx context.Context, cmd domain.UpdateSquadCommand) (*domain.Squad, error)
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

func (f *fakeSquadRepo) ListForUser(_ context.Context, _ string, _ pagination.Page) ([]domain.SquadSummary, string, error) {
	return nil, "", nil
}

func (f *fakeSquadRepo) Update(ctx context.Context, cmd domain.UpdateSquadCommand) (*domain.Squad, error) {
	if f.updateFn != nil {
		return f.updateFn(ctx, cmd)
	}

	return &domain.Squad{ID: cmd.SquadID}, nil
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

// rules returns an authorizer over the real policy, so these tests run
// against the shipped rules instead of a copy of them.
func rules(t *testing.T) *authz.Enforcer[domain.Role, domain.Action] {
	t.Helper()

	a, err := authz.New[domain.Role, domain.Action](policy.CSV)
	require.NoError(t, err)

	return a
}

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
	uc := NewSquadUseCase(&fakeSquadRepo{}, &fakeUserRepo{}, rules(t))

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

	_, err := NewSquadUseCase(squads, &fakeUserRepo{}, rules(t)).Create(t.Context(), domain.CreateSquadCommand{CreatorID: "u1", Name: "  Friday crew "})
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

	_, err := NewSquadUseCase(squads, &fakeUserRepo{}, rules(t)).Get(t.Context(), testSquadID, "stranger")

	assert.ErrorIs(t, err, domain.ErrSquadNotFound)
}

func TestGetSquadDescribesTheCaller(t *testing.T) {
	squads := &fakeSquadRepo{
		getForMemberFn: func(_ context.Context, squadID, _ string) (*domain.Squad, error) {
			return &domain.Squad{ID: squadID, CreatedBy: "meera"}, nil
		},
		listMembersFn: func(_ context.Context, _ string) ([]domain.SquadMember, error) {
			return []domain.SquadMember{
				{UserID: "meera", Name: "Meera", Role: domain.RoleAdmin},
				{UserID: "rohan", Name: "Rohan", Role: domain.RoleMember},
			}, nil
		},
	}

	got, err := NewSquadUseCase(squads, &fakeUserRepo{}, rules(t)).Get(t.Context(), testSquadID, "rohan")
	require.NoError(t, err)

	assert.Equal(t, "rohan", got.Me.UserID)
	assert.Equal(t, domain.RoleMember, got.Me.Role)
	assert.Equal(t, domain.Permissions{CanAddMembers: true}, got.Permissions, "a member may only add people")
}

func TestGetSquadWhenCallerMissingFromMembers(t *testing.T) {
	squads := &fakeSquadRepo{
		getForMemberFn: func(_ context.Context, squadID, _ string) (*domain.Squad, error) {
			return &domain.Squad{ID: squadID}, nil
		},
	}

	_, err := NewSquadUseCase(squads, &fakeUserRepo{}, rules(t)).Get(t.Context(), testSquadID, "left-just-now")

	assert.ErrorIs(t, err, domain.ErrSquadNotFound, "someone who left between the two reads is no longer a member")
}

func TestUpdateSquad(t *testing.T) {
	name := func(s string) *string { return &s }

	tests := []struct {
		name      string
		getMember func(context.Context, string, string) (*domain.SquadMember, error)
		newName   *string
		wantErr   error
		wantName  *string
	}{
		{name: "AdminRenames", getMember: memberAs(domain.RoleAdmin), newName: name("  Saturday crew "), wantName: name("Saturday crew")},
		{name: "AdminChangesOnlyAvatar", getMember: memberAs(domain.RoleAdmin)},
		{name: "BlankName", getMember: memberAs(domain.RoleAdmin), newName: name("   "), wantErr: domain.ErrInvalidSquadName},
		{name: "MemberCannotEdit", getMember: memberAs(domain.RoleMember), newName: name("Mine now"), wantErr: domain.ErrForbidden},
		{name: "NonMember", getMember: notMember, newName: name("Mine now"), wantErr: domain.ErrSquadNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got *domain.UpdateSquadCommand

			squads := &fakeSquadRepo{
				getMemberFn: tt.getMember,
				updateFn: func(_ context.Context, cmd domain.UpdateSquadCommand) (*domain.Squad, error) {
					got = &cmd

					return &domain.Squad{ID: cmd.SquadID}, nil
				},
			}

			_, err := NewSquadUseCase(squads, &fakeUserRepo{}, rules(t)).Update(t.Context(), domain.UpdateSquadCommand{
				SquadID: testSquadID, ActorID: "u1", Name: tt.newName, AvatarURL: name("https://img/new.png"),
			})
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, got, "the squad must not be updated")

				return
			}

			require.NoError(t, err)
			require.NotNil(t, got)
			assert.Equal(t, tt.wantName, got.Name)
			assert.Equal(t, "https://img/new.png", *got.AvatarURL)
		})
	}
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

			_, err := NewSquadUseCase(squads, &fakeUserRepo{getByPhoneFn: tt.getByPhone}, rules(t)).AddMember(t.Context(), domain.AddMemberCommand{
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
		{name: "MemberRemovesOther", getMember: memberAs(domain.RoleMember), target: "other", wantErr: domain.ErrForbidden},
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

			_, err := NewSquadUseCase(squads, &fakeUserRepo{}, rules(t)).RemoveMember(t.Context(), domain.RemoveMemberCommand{
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
		{name: "MemberCannotChangeRoles", getMember: memberAs(domain.RoleMember), role: domain.RoleAdmin, wantErr: domain.ErrForbidden},
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

			err := NewSquadUseCase(squads, &fakeUserRepo{}, rules(t)).ChangeRole(t.Context(), domain.ChangeRoleCommand{
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
