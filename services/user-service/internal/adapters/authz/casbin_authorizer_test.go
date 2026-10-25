package authz

import (
	"testing"

	"github.com/SamPariat/roundup/services/user-service/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// casbin_authorizer_test.go checks every role and action against the embedded
// policy. Change this table whenever policy.csv changes.

func TestPolicy(t *testing.T) {
	authz, err := NewCasbinAuthorizer()
	require.NoError(t, err)

	tests := []struct {
		role   domain.Role
		action domain.Action
		want   bool
	}{
		{domain.RoleAdmin, domain.ActionAddMember, true},
		{domain.RoleAdmin, domain.ActionRemoveMember, true},
		{domain.RoleAdmin, domain.ActionLeave, true},
		{domain.RoleAdmin, domain.ActionChangeRole, true},
		{domain.RoleAdmin, domain.ActionEditSquad, true},

		{domain.RoleMember, domain.ActionAddMember, true},
		{domain.RoleMember, domain.ActionRemoveMember, false},
		{domain.RoleMember, domain.ActionLeave, true},
		{domain.RoleMember, domain.ActionChangeRole, false},
		{domain.RoleMember, domain.ActionEditSquad, false},

		{domain.Role("GHOST"), domain.ActionLeave, false},
		{domain.Role("admin"), domain.ActionEditSquad, false},
		{domain.RoleAdmin, domain.Action("squad:delete"), false},
	}

	for _, tt := range tests {
		t.Run(string(tt.role)+"/"+string(tt.action), func(t *testing.T) {
			assert.Equal(t, tt.want, authz.Can(tt.role, tt.action))
		})
	}
}
