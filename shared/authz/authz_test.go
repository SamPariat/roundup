package authz

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const policy = `p, ADMIN, thing:edit
p, ADMIN, thing:read
p, MEMBER, thing:read
`

func TestCan(t *testing.T) {
	e, err := New[string, string](policy)
	require.NoError(t, err)

	tests := []struct {
		name         string
		role, action string
		want         bool
	}{
		{name: "AdminEdits", role: "ADMIN", action: "thing:edit", want: true},
		{name: "MemberReads", role: "MEMBER", action: "thing:read", want: true},
		{name: "MemberCannotEdit", role: "MEMBER", action: "thing:edit"},
		{name: "UnknownRole", role: "GUEST", action: "thing:read"},
		{name: "UnknownAction", role: "ADMIN", action: "thing:delete"},
		{name: "Empty"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, e.Can(tt.role, tt.action))
		})
	}
}

func TestNewRejectsEmptyPolicy(t *testing.T) {
	_, err := New[string, string]("")

	assert.Error(t, err, "a service with no rules is a mistake, not an allow-nothing policy")
}
