package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// allowList is an Authorizer that allows exactly the listed actions for every
// role, so these tests check PermissionsFor's wiring, not the policy (that
// lives in adapters/authz).
type allowList map[Action]bool

func (a allowList) Can(_ Role, action Action) bool { return a[action] }

func TestPermissionsFor(t *testing.T) {
	everything := allowList{ActionAddMember: true, ActionRemoveMember: true, ActionChangeRole: true, ActionEditSquad: true}
	member := SquadMember{UserID: "u1", Role: RoleMember}

	tests := []struct {
		name        string
		authz       allowList
		memberCount int
		want        Permissions
	}{
		{name: "EverythingAllowed", authz: everything, memberCount: 3, want: Permissions{CanAddMembers: true, CanRemoveMembers: true, CanChangeRoles: true, CanEditSquad: true}},
		{name: "NothingAllowed", authz: allowList{}, memberCount: 3, want: Permissions{}},
		{name: "OnlyAdding", authz: allowList{ActionAddMember: true}, memberCount: 3, want: Permissions{CanAddMembers: true}},
		{name: "FullSquadCannotAdd", authz: everything, memberCount: MaxSquadMembers, want: Permissions{CanRemoveMembers: true, CanChangeRoles: true, CanEditSquad: true}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, PermissionsFor(tt.authz, member, tt.memberCount))
		})
	}
}

func TestRoleIsValid(t *testing.T) {
	assert.True(t, RoleAdmin.IsValid())
	assert.True(t, RoleMember.IsValid())
	assert.False(t, Role("admin").IsValid(), "roles are case-sensitive")
	assert.False(t, Role("ORGANISER").IsValid())
}
