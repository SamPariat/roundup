package persistence

import (
	"fmt"
	"testing"

	"github.com/SamPariat/roundup/services/user-service/internal/domain"
	"github.com/SamPariat/roundup/shared/pagination"
	"github.com/SamPariat/roundup/shared/postgres/postgrestest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// firstPage is the default first page, for tests that list fewer squads than a page.
var firstPage = pagination.Page{Limit: pagination.DefaultLimit}

// postgres_squad_repository_test.go runs the squad repository against a real,
// freshly migrated Postgres database, including the transactional rules
// (member cap, admin succession, deleting an empty squad).

type squadFixture struct {
	users  *PostgresUserRepository
	squads *PostgresSquadRepository
	phones int
}

func newSquadFixture(t *testing.T) *squadFixture {
	t.Helper()

	pool := postgrestest.New(t, migrationsDir)

	return &squadFixture{users: NewPostgresUserRepository(pool), squads: NewPostgresSquadRepository(pool)}
}

// user creates a user with a fresh phone number.
func (f *squadFixture) user(t *testing.T) string {
	t.Helper()

	f.phones++

	return mustCreateUser(t, f.users, fmt.Sprintf("+9198765%05d", f.phones)).ID
}

// squad creates a squad whose admin is creator.
func (f *squadFixture) squad(t *testing.T, creator string) string {
	t.Helper()

	squad, err := f.squads.Create(t.Context(), domain.CreateSquadCommand{CreatorID: creator, Name: "Friday crew"})
	require.NoError(t, err)

	return squad.ID
}

func (f *squadFixture) add(t *testing.T, squadID, userID string) {
	t.Helper()

	_, err := f.squads.AddMember(t.Context(), squadID, userID, domain.MaxSquadMembers)
	require.NoError(t, err)
}

func (f *squadFixture) role(t *testing.T, squadID, userID string) domain.Role {
	t.Helper()

	member, err := f.squads.GetMember(t.Context(), squadID, userID)
	require.NoError(t, err)

	return member.Role
}

func TestCreateSquadMakesCreatorAdmin(t *testing.T) {
	f := newSquadFixture(t)
	creator := f.user(t)

	squad, err := f.squads.Create(t.Context(), domain.CreateSquadCommand{CreatorID: creator, Name: "Friday crew", AvatarURL: "https://img/x.png"})
	require.NoError(t, err)

	assert.Equal(t, "Friday crew", squad.Name)
	assert.Equal(t, "https://img/x.png", squad.AvatarURL)
	assert.Equal(t, creator, squad.CreatedBy)
	assert.Equal(t, domain.RoleAdmin, f.role(t, squad.ID, creator))
}

func TestGetForMemberHidesSquadFromNonMembers(t *testing.T) {
	f := newSquadFixture(t)
	creator, stranger := f.user(t), f.user(t)
	squadID := f.squad(t, creator)

	_, err := f.squads.GetForMember(t.Context(), squadID, creator)
	require.NoError(t, err)

	_, err = f.squads.GetForMember(t.Context(), squadID, stranger)
	require.ErrorIs(t, err, domain.ErrSquadNotFound)

	_, err = f.squads.GetForMember(t.Context(), "not-a-uuid", creator)
	assert.ErrorIs(t, err, domain.ErrSquadNotFound)
}

func TestListForUser(t *testing.T) {
	f := newSquadFixture(t)
	alice, bob := f.user(t), f.user(t)
	squadID := f.squad(t, alice)
	f.add(t, squadID, bob)

	squads, _, err := f.squads.ListForUser(t.Context(), bob, firstPage)
	require.NoError(t, err)

	require.Len(t, squads, 1)
	assert.Equal(t, squadID, squads[0].ID)
	assert.Equal(t, domain.RoleMember, squads[0].Role)
	assert.Equal(t, 2, squads[0].MemberCount)
}

func TestListForUserPreviewsFirstMembersPerSquad(t *testing.T) {
	f := newSquadFixture(t)

	users := make([]string, 6)
	for i := range users {
		users[i] = f.user(t)
	}

	big := f.squad(t, users[0])
	for _, u := range users[1:] {
		f.add(t, big, u)
	}

	small := f.squad(t, users[5])
	require.NoError(t, f.users.SoftDelete(t.Context(), users[2]))

	squads, _, err := f.squads.ListForUser(t.Context(), users[5], firstPage)
	require.NoError(t, err)

	require.Len(t, squads, 2)

	previews := map[string][]string{}

	for _, s := range squads {
		for _, m := range s.MemberPreview {
			previews[s.ID] = append(previews[s.ID], m.UserID)
		}
	}

	assert.Equal(t, []string{users[0], users[1], users[3], users[4]}, previews[big],
		"earliest joiners first, at most MemberPreviewSize, deleted users hidden")
	assert.Equal(t, []string{users[5]}, previews[small])
}

func TestListForUserWithNoSquads(t *testing.T) {
	f := newSquadFixture(t)

	squads, _, err := f.squads.ListForUser(t.Context(), f.user(t), firstPage)

	require.NoError(t, err)
	assert.Empty(t, squads)
}

func TestUpdateSquad(t *testing.T) {
	f := newSquadFixture(t)
	squadID := f.squad(t, f.user(t))
	text := func(s string) *string { return &s }

	squad, err := f.squads.Update(t.Context(), domain.UpdateSquadCommand{SquadID: squadID, AvatarURL: text("https://img/new.png")})
	require.NoError(t, err)
	assert.Equal(t, "Friday crew", squad.Name, "a nil name is left unchanged")
	assert.Equal(t, "https://img/new.png", squad.AvatarURL)

	squad, err = f.squads.Update(t.Context(), domain.UpdateSquadCommand{SquadID: squadID, Name: text("Saturday crew")})
	require.NoError(t, err)
	assert.Equal(t, "Saturday crew", squad.Name)
	assert.Equal(t, "https://img/new.png", squad.AvatarURL, "a nil avatar is left unchanged")

	squad, err = f.squads.Update(t.Context(), domain.UpdateSquadCommand{SquadID: squadID, AvatarURL: text("")})
	require.NoError(t, err)
	assert.Empty(t, squad.AvatarURL, "an empty avatar removes the picture")
}

func TestUpdateMissingSquad(t *testing.T) {
	f := newSquadFixture(t)
	name := "Ghost crew"

	_, err := f.squads.Update(t.Context(), domain.UpdateSquadCommand{SquadID: "6f1c2a9e-1111-4c2b-9d7e-2b1f0a9c4e11", Name: &name})
	require.ErrorIs(t, err, domain.ErrSquadNotFound)

	_, err = f.squads.Update(t.Context(), domain.UpdateSquadCommand{SquadID: "not-a-uuid", Name: &name})
	assert.ErrorIs(t, err, domain.ErrSquadNotFound)
}

func TestListForUserPagesWithoutGapsOrRepeats(t *testing.T) {
	f := newSquadFixture(t)
	owner := f.user(t)

	for range 25 {
		f.squad(t, owner)
	}

	var (
		seen   = map[string]bool{}
		sizes  []int
		cursor string
	)

	for {
		squads, next, err := f.squads.ListForUser(t.Context(), owner, pagination.Page{Limit: 10, Cursor: cursor})
		require.NoError(t, err)

		sizes = append(sizes, len(squads))

		for _, s := range squads {
			assert.False(t, seen[s.ID], "squad %s repeated", s.ID)
			seen[s.ID] = true
		}

		if next == "" {
			break
		}

		if len(sizes) == 1 {
			f.squad(t, owner) // created mid-scroll: newer than the cursor, so it must not shift later pages
		}

		cursor = next
	}

	assert.Equal(t, []int{10, 10, 5}, sizes)
	assert.Len(t, seen, 25)
}

func TestListForUserRejectsForeignCursor(t *testing.T) {
	f := newSquadFixture(t)

	_, _, err := f.squads.ListForUser(t.Context(), f.user(t), pagination.Page{Limit: 10, Cursor: "not-ours"})

	assert.ErrorIs(t, err, pagination.ErrInvalid)
}

func TestAddMemberTwiceReturnsErrAlreadyMember(t *testing.T) {
	f := newSquadFixture(t)
	creator, friend := f.user(t), f.user(t)
	squadID := f.squad(t, creator)
	f.add(t, squadID, friend)

	_, err := f.squads.AddMember(t.Context(), squadID, friend, domain.MaxSquadMembers)

	assert.ErrorIs(t, err, domain.ErrAlreadyMember)
}

func TestAddMemberToMissingSquad(t *testing.T) {
	f := newSquadFixture(t)

	_, err := f.squads.AddMember(t.Context(), "7dc92d84-1950-43d6-b676-debacf84d830", f.user(t), domain.MaxSquadMembers)

	assert.ErrorIs(t, err, domain.ErrSquadNotFound)
}

func TestAddMemberRespectsLimit(t *testing.T) {
	f := newSquadFixture(t)
	creator := f.user(t)
	squadID := f.squad(t, creator)

	_, err := f.squads.AddMember(t.Context(), squadID, f.user(t), 2)
	require.NoError(t, err)

	_, err = f.squads.AddMember(t.Context(), squadID, f.user(t), 2)
	assert.ErrorIs(t, err, domain.ErrSquadFull)
}

func TestListMembersOrdersByJoinAndHidesDeletedUsers(t *testing.T) {
	f := newSquadFixture(t)
	alice, bob, carol := f.user(t), f.user(t), f.user(t)
	squadID := f.squad(t, alice)
	f.add(t, squadID, bob)
	f.add(t, squadID, carol)
	require.NoError(t, f.users.SoftDelete(t.Context(), bob))

	members, err := f.squads.ListMembers(t.Context(), squadID)
	require.NoError(t, err)

	require.Len(t, members, 2)
	assert.Equal(t, alice, members[0].UserID)
	assert.Equal(t, carol, members[1].UserID)
}

func TestLastAdminLeavingPromotesEarliestJoiner(t *testing.T) {
	f := newSquadFixture(t)
	alice, bob, carol := f.user(t), f.user(t), f.user(t)
	squadID := f.squad(t, alice)
	f.add(t, squadID, bob)
	f.add(t, squadID, carol)

	deleted, err := f.squads.RemoveMember(t.Context(), squadID, alice)
	require.NoError(t, err)

	assert.False(t, deleted)
	assert.Equal(t, domain.RoleAdmin, f.role(t, squadID, bob), "bob joined first after alice")
	assert.Equal(t, domain.RoleMember, f.role(t, squadID, carol))
}

func TestAdminLeavingWithAnotherAdminPromotesNobody(t *testing.T) {
	f := newSquadFixture(t)
	alice, bob, carol := f.user(t), f.user(t), f.user(t)
	squadID := f.squad(t, alice)
	f.add(t, squadID, bob)
	f.add(t, squadID, carol)
	require.NoError(t, f.squads.ChangeRole(t.Context(), squadID, carol, domain.RoleAdmin))

	_, err := f.squads.RemoveMember(t.Context(), squadID, alice)
	require.NoError(t, err)

	assert.Equal(t, domain.RoleMember, f.role(t, squadID, bob))
	assert.Equal(t, domain.RoleAdmin, f.role(t, squadID, carol))
}

func TestLastMemberLeavingDeletesSquad(t *testing.T) {
	f := newSquadFixture(t)
	alice := f.user(t)
	squadID := f.squad(t, alice)

	deleted, err := f.squads.RemoveMember(t.Context(), squadID, alice)
	require.NoError(t, err)

	assert.True(t, deleted)

	_, err = f.squads.AddMember(t.Context(), squadID, f.user(t), domain.MaxSquadMembers)
	assert.ErrorIs(t, err, domain.ErrSquadNotFound, "the squad row must be gone")
}

func TestRemoveMissingMember(t *testing.T) {
	f := newSquadFixture(t)
	squadID := f.squad(t, f.user(t))

	_, err := f.squads.RemoveMember(t.Context(), squadID, f.user(t))

	assert.ErrorIs(t, err, domain.ErrMemberNotFound)
}

func TestOnlyAdminCannotDemoteThemselves(t *testing.T) {
	f := newSquadFixture(t)
	alice, bob := f.user(t), f.user(t)
	squadID := f.squad(t, alice)
	f.add(t, squadID, bob)

	err := f.squads.ChangeRole(t.Context(), squadID, alice, domain.RoleMember)

	require.ErrorIs(t, err, domain.ErrLastAdmin)
	assert.Equal(t, domain.RoleAdmin, f.role(t, squadID, alice), "the failed change must roll back")
}

func TestPromotingGivesSquadTwoAdmins(t *testing.T) {
	f := newSquadFixture(t)
	alice, bob := f.user(t), f.user(t)
	squadID := f.squad(t, alice)
	f.add(t, squadID, bob)

	require.NoError(t, f.squads.ChangeRole(t.Context(), squadID, bob, domain.RoleAdmin))
	require.NoError(t, f.squads.ChangeRole(t.Context(), squadID, alice, domain.RoleMember), "bob is still admin")

	assert.Equal(t, domain.RoleMember, f.role(t, squadID, alice))
	assert.Equal(t, domain.RoleAdmin, f.role(t, squadID, bob))
}

func TestChangeRoleOfNonMember(t *testing.T) {
	f := newSquadFixture(t)
	squadID := f.squad(t, f.user(t))

	err := f.squads.ChangeRole(t.Context(), squadID, f.user(t), domain.RoleAdmin)

	assert.ErrorIs(t, err, domain.ErrMemberNotFound)
}
