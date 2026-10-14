package persistence

import (
	"context"
	"testing"
	"time"

	"github.com/SamPariat/roundup/services/user-service/internal/domain"
	"github.com/SamPariat/roundup/shared/postgres/postgrestest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// postgres_user_repository_test.go runs the repository against a real,
// freshly migrated Postgres database. Tests are skipped unless
// TEST_POSTGRES_DSN is set (see shared/postgres/postgrestest).

const migrationsDir = "../../../db/migrations"

func newRepo(t *testing.T) *PostgresUserRepository {
	t.Helper()
	return NewPostgresUserRepository(postgrestest.New(t, migrationsDir))
}

func mustCreateUser(t *testing.T, repo *PostgresUserRepository, phone string) *domain.User {
	t.Helper()
	user, err := repo.Create(context.Background(), phone)
	require.NoError(t, err)
	return user
}

// --- users ---

func TestCreateAndGetUser(t *testing.T) {
	ctx := context.Background()
	repo := newRepo(t)

	created := mustCreateUser(t, repo, "+919876543210")
	assert.NotEmpty(t, created.ID)
	assert.Equal(t, "", created.Name, "NULL name maps to empty string")

	byID, err := repo.GetByID(ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, created, byID)

	byPhone, err := repo.GetByPhone(ctx, "+919876543210")
	require.NoError(t, err)
	assert.Equal(t, created.ID, byPhone.ID)
}

func TestCreateDuplicatePhoneReturnsErrPhoneTaken(t *testing.T) {
	repo := newRepo(t)
	mustCreateUser(t, repo, "+919876543210")

	_, err := repo.Create(context.Background(), "+919876543210")

	assert.ErrorIs(t, err, domain.ErrPhoneTaken)
}

func TestGetByIDNotFound(t *testing.T) {
	repo := newRepo(t)

	for _, id := range []string{"00000000-0000-0000-0000-000000000000", "not-a-uuid"} {
		_, err := repo.GetByID(context.Background(), id)
		assert.ErrorIs(t, err, domain.ErrUserNotFound, id)
	}
}

func TestUpdateProfileOnlyChangesNonNilFields(t *testing.T) {
	ctx := context.Background()
	repo := newRepo(t)
	user := mustCreateUser(t, repo, "+919876543210")

	name, avatar := "Sam", "https://cdn.example/a.png"
	_, err := repo.UpdateProfile(ctx, domain.UpdateProfileCommand{UserID: user.ID, Name: &name, AvatarURL: &avatar})
	require.NoError(t, err)

	newName := "Sam"
	updated, err := repo.UpdateProfile(ctx, domain.UpdateProfileCommand{UserID: user.ID, Name: &newName})
	require.NoError(t, err)
	assert.Equal(t, "Sam", updated.Name)
	assert.Equal(t, avatar, updated.AvatarURL, "nil AvatarURL must leave it unchanged")
	assert.True(t, updated.UpdatedAt.After(user.UpdatedAt))

	empty := ""
	updated, err = repo.UpdateProfile(ctx, domain.UpdateProfileCommand{UserID: user.ID, AvatarURL: &empty})
	require.NoError(t, err)
	assert.Equal(t, "", updated.AvatarURL, "empty string clears the avatar")
}

func TestSoftDeleteHidesUser(t *testing.T) {
	ctx := context.Background()
	repo := newRepo(t)
	user := mustCreateUser(t, repo, "+919876543210")

	require.NoError(t, repo.SoftDelete(ctx, user.ID))

	_, err := repo.GetByID(ctx, user.ID)
	assert.ErrorIs(t, err, domain.ErrUserNotFound)
	_, err = repo.GetByPhone(ctx, user.Phone)
	assert.ErrorIs(t, err, domain.ErrUserNotFound)
	assert.ErrorIs(t, repo.SoftDelete(ctx, user.ID), domain.ErrUserNotFound, "deleting twice")
}

func TestSoftDeletedPhoneCanSignUpAgain(t *testing.T) {
	ctx := context.Background()
	repo := newRepo(t)
	user := mustCreateUser(t, repo, "+919876543210")

	require.NoError(t, repo.SoftDelete(ctx, user.ID))

	again, err := repo.Create(ctx, "+919876543210")
	require.NoError(t, err, "a soft-deleted user's phone must be free to sign up again")
	assert.NotEqual(t, user.ID, again.ID, "re-signup creates a new user, not a revived one")

	found, err := repo.GetByPhone(ctx, "+919876543210")
	require.NoError(t, err)
	assert.Equal(t, again.ID, found.ID, "lookup by phone returns the active user")
}

// --- device tokens ---

func TestUpsertMovesTokenToNewOwnerAndKeepsCreatedAt(t *testing.T) {
	ctx := context.Background()
	repo := newRepo(t)
	alice := mustCreateUser(t, repo, "+919800000001")
	bob := mustCreateUser(t, repo, "+919800000002")

	first, err := repo.Upsert(ctx, domain.RegisterDeviceTokenCommand{UserID: alice.ID, Token: "tok-1", Platform: domain.PlatformIOS})
	require.NoError(t, err)

	time.Sleep(10 * time.Millisecond) // make a changed created_at detectable
	moved, err := repo.Upsert(ctx, domain.RegisterDeviceTokenCommand{UserID: bob.ID, Token: "tok-1", Platform: domain.PlatformAndroid})
	require.NoError(t, err)

	assert.Equal(t, first.ID, moved.ID, "same token is updated in place, not duplicated")
	assert.Equal(t, bob.ID, moved.UserID)
	assert.Equal(t, domain.PlatformAndroid, moved.Platform)
	assert.True(t, first.CreatedAt.Equal(moved.CreatedAt), "re-registering keeps the original created_at")

	assert.ErrorIs(t, repo.Delete(ctx, domain.RemoveDeviceTokenCommand{UserID: alice.ID, Token: "tok-1"}),
		domain.ErrDeviceTokenNotFound, "previous owner no longer owns the token")
}

func TestUpsertRejectsUnknownPlatform(t *testing.T) {
	// The use case validates platforms first; the CHECK constraint is the backstop.
	repo := newRepo(t)
	user := mustCreateUser(t, repo, "+919876543210")

	_, err := repo.Upsert(context.Background(), domain.RegisterDeviceTokenCommand{UserID: user.ID, Token: "tok-1", Platform: "ios"})

	assert.Error(t, err)
}

func TestDeleteTokenRequiresOwner(t *testing.T) {
	ctx := context.Background()
	repo := newRepo(t)
	owner := mustCreateUser(t, repo, "+919800000001")
	other := mustCreateUser(t, repo, "+919800000002")
	_, err := repo.Upsert(ctx, domain.RegisterDeviceTokenCommand{UserID: owner.ID, Token: "tok-1", Platform: domain.PlatformIOS})
	require.NoError(t, err)

	assert.ErrorIs(t, repo.Delete(ctx, domain.RemoveDeviceTokenCommand{UserID: other.ID, Token: "tok-1"}), domain.ErrDeviceTokenNotFound)
	assert.NoError(t, repo.Delete(ctx, domain.RemoveDeviceTokenCommand{UserID: owner.ID, Token: "tok-1"}))
	assert.ErrorIs(t, repo.Delete(ctx, domain.RemoveDeviceTokenCommand{UserID: owner.ID, Token: "tok-1"}), domain.ErrDeviceTokenNotFound, "already deleted")
}

func TestDeleteAllForUserOnlyRemovesThatUsersTokens(t *testing.T) {
	ctx := context.Background()
	repo := newRepo(t)
	alice := mustCreateUser(t, repo, "+919800000001")
	bob := mustCreateUser(t, repo, "+919800000002")
	for _, cmd := range []domain.RegisterDeviceTokenCommand{
		{UserID: alice.ID, Token: "alice-ios", Platform: domain.PlatformIOS},
		{UserID: alice.ID, Token: "alice-android", Platform: domain.PlatformAndroid},
		{UserID: bob.ID, Token: "bob-ios", Platform: domain.PlatformIOS},
	} {
		_, err := repo.Upsert(ctx, cmd)
		require.NoError(t, err)
	}

	require.NoError(t, repo.DeleteAllForUser(ctx, alice.ID))

	assert.ErrorIs(t, repo.Delete(ctx, domain.RemoveDeviceTokenCommand{UserID: alice.ID, Token: "alice-ios"}), domain.ErrDeviceTokenNotFound)
	assert.NoError(t, repo.Delete(ctx, domain.RemoveDeviceTokenCommand{UserID: bob.ID, Token: "bob-ios"}), "bob's token survives")
	assert.NoError(t, repo.DeleteAllForUser(ctx, "not-a-uuid"), "invalid ID is a no-op")
}
