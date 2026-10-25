package policy

import (
	"strings"
	"testing"

	"github.com/SamPariat/roundup/services/event-service/internal/domain"
	"github.com/SamPariat/roundup/shared/authz"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// policy_test.go checks every role and action against policy.csv. Change this table whenever policy.csv changes.

var allActions = []domain.Action{
	domain.ActionCreate, domain.ActionRSVP, domain.ActionNominate, domain.ActionVote,
	domain.ActionRemoveOwnNomination, domain.ActionRemoveNomination, domain.ActionEdit, domain.ActionManage,
}

func TestPolicy(t *testing.T) {
	authorizer, err := authz.New[domain.Role, domain.Action](CSV)
	require.NoError(t, err)

	everyone := map[domain.Action]bool{
		domain.ActionCreate: true, domain.ActionRSVP: true, domain.ActionNominate: true,
		domain.ActionVote: true, domain.ActionRemoveOwnNomination: true,
	}
	organisers := map[domain.Action]bool{
		domain.ActionRemoveNomination: true, domain.ActionEdit: true, domain.ActionManage: true,
	}

	want := map[domain.Role]func(domain.Action) bool{
		domain.RoleMember:    func(a domain.Action) bool { return everyone[a] },
		domain.RoleAdmin:     func(a domain.Action) bool { return everyone[a] || organisers[a] },
		domain.RoleCreator:   func(a domain.Action) bool { return organisers[a] },
		domain.Role("GHOST"): func(domain.Action) bool { return false },
		domain.Role("admin"): func(domain.Action) bool { return false },
	}

	for role, allowed := range want {
		for _, action := range append(allActions, domain.Action("event:delete")) {
			t.Run(string(role)+"/"+string(action), func(t *testing.T) {
				assert.Equal(t, allowed(action), authorizer.Can(role, action))
			})
		}
	}
}

// Casbin accepts a short line such as "p, ADMIN" without complaint, so a typo
// would silently deny. Every line must name a role and a known action.
func TestPolicyLinesAreWellFormed(t *testing.T) {
	for line := range strings.SplitSeq(strings.TrimSpace(CSV), "\n") {
		fields := strings.Split(line, ",")
		require.Len(t, fields, 3, "line %q", line)

		for i := range fields {
			fields[i] = strings.TrimSpace(fields[i])
		}

		assert.Equal(t, "p", fields[0], "line %q", line)
		assert.Contains(t, []domain.Role{domain.RoleMember, domain.RoleAdmin, domain.RoleCreator}, domain.Role(fields[1]), "line %q", line)
		assert.Contains(t, allActions, domain.Action(fields[2]), "line %q", line)
	}
}
