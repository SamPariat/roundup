package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var allStatuses = []Status{StatusDraft, StatusVoting, StatusConfirmed, StatusActive, StatusClosed}

func TestCanMoveToEveryPair(t *testing.T) {
	allowed := map[[2]Status]bool{
		{StatusDraft, StatusVoting}:     true,
		{StatusVoting, StatusConfirmed}: true,
		{StatusConfirmed, StatusActive}: true,
		{StatusActive, StatusClosed}:    true,
	}

	for _, from := range allStatuses {
		for _, to := range allStatuses {
			t.Run(string(from)+"→"+string(to), func(t *testing.T) {
				assert.Equal(t, allowed[[2]Status{from, to}], from.CanMoveTo(to))
			})
		}
	}
}

func TestWhatEachStatusAllows(t *testing.T) {
	tests := []struct {
		status                                   Status
		editable, rsvps, nominations, takesVotes bool
	}{
		{StatusDraft, true, true, true, false},
		{StatusVoting, true, true, false, true},
		{StatusConfirmed, false, true, false, false},
		{StatusActive, false, true, false, false},
		{StatusClosed, false, false, false, false},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			assert.Equal(t, tt.editable, tt.status.IsEditable(), "IsEditable")
			assert.Equal(t, tt.rsvps, tt.status.TakesRSVPs(), "TakesRSVPs")
			assert.Equal(t, tt.nominations, tt.status.TakesNominations(), "TakesNominations")
			assert.Equal(t, tt.takesVotes, tt.status.TakesVotes(), "TakesVotes")
		})
	}
}

func TestValidValues(t *testing.T) {
	assert.True(t, TypeBoth.IsValid())
	assert.False(t, Type("drinks").IsValid(), "types are case-sensitive")
	assert.True(t, RSVPNotGoing.IsValid())
	assert.False(t, RSVPStatus("PENDING").IsValid(), "pending is the absence of an answer, not an answer")
}

func TestRolesOfAddsCreator(t *testing.T) {
	member := SquadMember{UserID: "u1", Role: string(RoleMember)}

	assert.Equal(t, []Role{RoleMember}, RolesOf(member, nil))
	assert.Equal(t, []Role{RoleMember}, RolesOf(member, &Event{CreatedBy: "u2"}))
	assert.Equal(t, []Role{RoleMember, RoleCreator}, RolesOf(member, &Event{CreatedBy: "u1"}))
}
