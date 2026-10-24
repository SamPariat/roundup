package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCanRemove(t *testing.T) {
	admin := SquadMember{UserID: "admin", Role: RoleAdmin}
	member := SquadMember{UserID: "member", Role: RoleMember}

	tests := []struct {
		name    string
		actor   SquadMember
		target  string
		wantErr error
	}{
		{name: "AdminRemovesMember", actor: admin, target: "member"},
		{name: "AdminRemovesAnotherAdmin", actor: admin, target: "other-admin"},
		{name: "AdminLeaves", actor: admin, target: "admin"},
		{name: "MemberLeaves", actor: member, target: "member"},
		{name: "MemberRemovesSomeoneElse", actor: member, target: "admin", wantErr: ErrNotAdmin},
		{name: "UnknownRoleRemovesSomeoneElse", actor: SquadMember{UserID: "ghost"}, target: "member", wantErr: ErrNotAdmin},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CanRemove(tt.actor, tt.target)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)

				return
			}

			assert.NoError(t, err)
		})
	}
}

func TestCanChangeRole(t *testing.T) {
	require.NoError(t, CanChangeRole(SquadMember{Role: RoleAdmin}))
	assert.ErrorIs(t, CanChangeRole(SquadMember{Role: RoleMember}), ErrNotAdmin)
}

func TestRoleIsValid(t *testing.T) {
	assert.True(t, RoleAdmin.IsValid())
	assert.True(t, RoleMember.IsValid())
	assert.False(t, Role("admin").IsValid(), "roles are case-sensitive")
	assert.False(t, Role("ORGANISER").IsValid())
}
