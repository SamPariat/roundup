package application

import (
	"context"
	"errors"
	"testing"

	"github.com/SamPariat/roundup/services/user-service/internal/domain"
	"github.com/SamPariat/roundup/shared/domainerr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// user_usecase_test.go tests UserUseCase using hand-rolled fakes for the
// UserRepository and DeviceTokenRepository ports. No real database calls are made.
// device_token_usecase_test.go reuses the same fakes.

type fakeUserRepo struct {
	createFn        func(ctx context.Context, phone string) (*domain.User, error)
	getByIDFn       func(ctx context.Context, id string) (*domain.User, error)
	getByPhoneFn    func(ctx context.Context, phone string) (*domain.User, error)
	updateProfileFn func(ctx context.Context, cmd domain.UpdateProfileCommand) (*domain.User, error)
	softDeleteFn    func(ctx context.Context, id string) error
	listByPhonesFn  func(ctx context.Context, phones []string) ([]*domain.User, error)
}

func (f *fakeUserRepo) Create(ctx context.Context, phone string) (*domain.User, error) {
	if f.createFn != nil {
		return f.createFn(ctx, phone)
	}

	return nil, nil
}

func (f *fakeUserRepo) GetByID(ctx context.Context, id string) (*domain.User, error) {
	if f.getByIDFn != nil {
		return f.getByIDFn(ctx, id)
	}

	return nil, nil
}

func (f *fakeUserRepo) GetByPhone(ctx context.Context, phone string) (*domain.User, error) {
	if f.getByPhoneFn != nil {
		return f.getByPhoneFn(ctx, phone)
	}

	return nil, nil
}

func (f *fakeUserRepo) UpdateProfile(ctx context.Context, cmd domain.UpdateProfileCommand) (*domain.User, error) {
	if f.updateProfileFn != nil {
		return f.updateProfileFn(ctx, cmd)
	}

	return nil, nil
}

func (f *fakeUserRepo) SoftDelete(ctx context.Context, id string) error {
	if f.softDeleteFn != nil {
		return f.softDeleteFn(ctx, id)
	}

	return nil
}

func (f *fakeUserRepo) ListByPhones(ctx context.Context, phones []string) ([]*domain.User, error) {
	if f.listByPhonesFn != nil {
		return f.listByPhonesFn(ctx, phones)
	}

	return nil, nil
}

type fakeDeviceTokenRepo struct {
	upsertFn           func(ctx context.Context, cmd domain.RegisterDeviceTokenCommand) (*domain.DeviceToken, error)
	deleteFn           func(ctx context.Context, cmd domain.RemoveDeviceTokenCommand) error
	deleteAllForUserFn func(ctx context.Context, userID string) error
}

func (f *fakeDeviceTokenRepo) Upsert(ctx context.Context, cmd domain.RegisterDeviceTokenCommand) (*domain.DeviceToken, error) {
	if f.upsertFn != nil {
		return f.upsertFn(ctx, cmd)
	}

	return nil, nil
}

func (f *fakeDeviceTokenRepo) Delete(ctx context.Context, cmd domain.RemoveDeviceTokenCommand) error {
	if f.deleteFn != nil {
		return f.deleteFn(ctx, cmd)
	}

	return nil
}

func (f *fakeDeviceTokenRepo) DeleteAllForUser(ctx context.Context, userID string) error {
	if f.deleteAllForUserFn != nil {
		return f.deleteAllForUserFn(ctx, userID)
	}

	return nil
}

const testPhone = "+919876543210"

func notFound(_ context.Context, phone string) (*domain.User, error) {
	return nil, domainerr.New(domain.ErrUserNotFound, phone)
}

func TestCreateOrGetRejectsInvalidPhone(t *testing.T) {
	uc := NewUserUseCase(&fakeUserRepo{}, &fakeDeviceTokenRepo{})

	_, _, err := uc.CreateOrGet(t.Context(), "98765")

	assert.ErrorIs(t, err, domain.ErrInvalidPhone)
}

func TestCreateOrGetReturnsExistingUser(t *testing.T) {
	existing := &domain.User{ID: "user-1", Phone: testPhone}
	repo := &fakeUserRepo{
		getByPhoneFn: func(_ context.Context, _ string) (*domain.User, error) { return existing, nil },
		createFn: func(_ context.Context, _ string) (*domain.User, error) {
			t.Fatal("Create must not be called when the user already exists")

			return nil, nil
		},
	}
	uc := NewUserUseCase(repo, &fakeDeviceTokenRepo{})

	user, created, err := uc.CreateOrGet(t.Context(), testPhone)

	require.NoError(t, err)
	assert.False(t, created)
	assert.Equal(t, "user-1", user.ID)
}

func TestCreateOrGetCreatesNewUser(t *testing.T) {
	repo := &fakeUserRepo{
		getByPhoneFn: notFound,
		createFn: func(_ context.Context, phone string) (*domain.User, error) {
			return &domain.User{ID: "new-user", Phone: phone}, nil
		},
	}
	uc := NewUserUseCase(repo, &fakeDeviceTokenRepo{})

	user, created, err := uc.CreateOrGet(t.Context(), testPhone)

	require.NoError(t, err)
	assert.True(t, created)
	assert.Equal(t, "new-user", user.ID)
}

// Two concurrent signups for the same phone: both miss the lookup, one insert
// wins, the other hits the unique index. The loser should still return the user.
func TestCreateOrGetHandlesConcurrentCreate(t *testing.T) {
	lookups := 0
	repo := &fakeUserRepo{
		getByPhoneFn: func(ctx context.Context, phone string) (*domain.User, error) {
			lookups++
			if lookups == 1 {
				return notFound(ctx, phone)
			}

			return &domain.User{ID: "winner", Phone: phone}, nil
		},
		createFn: func(_ context.Context, phone string) (*domain.User, error) {
			return nil, domainerr.New(domain.ErrPhoneTaken, phone)
		},
	}
	uc := NewUserUseCase(repo, &fakeDeviceTokenRepo{})

	user, created, err := uc.CreateOrGet(t.Context(), testPhone)

	require.NoError(t, err)
	assert.False(t, created)
	assert.Equal(t, "winner", user.ID)
}

func TestCreateOrGetPropagatesLookupFailure(t *testing.T) {
	dbDown := errors.New("connection refused")
	repo := &fakeUserRepo{
		getByPhoneFn: func(_ context.Context, _ string) (*domain.User, error) { return nil, dbDown },
	}
	uc := NewUserUseCase(repo, &fakeDeviceTokenRepo{})

	_, _, err := uc.CreateOrGet(t.Context(), testPhone)

	assert.ErrorIs(t, err, dbDown)
}

func TestUpdateProfileRejectsBlankName(t *testing.T) {
	uc := NewUserUseCase(&fakeUserRepo{}, &fakeDeviceTokenRepo{})
	blank := "   "

	_, err := uc.UpdateProfile(t.Context(), domain.UpdateProfileCommand{UserID: "user-1", Name: &blank})

	assert.ErrorIs(t, err, domain.ErrInvalidName)
}

func TestUpdateProfileTrimsName(t *testing.T) {
	var got domain.UpdateProfileCommand

	repo := &fakeUserRepo{
		updateProfileFn: func(_ context.Context, cmd domain.UpdateProfileCommand) (*domain.User, error) {
			got = cmd

			return &domain.User{ID: cmd.UserID, Name: *cmd.Name}, nil
		},
	}
	uc := NewUserUseCase(repo, &fakeDeviceTokenRepo{})
	name := "  Priya  "

	user, err := uc.UpdateProfile(t.Context(), domain.UpdateProfileCommand{UserID: "user-1", Name: &name})

	require.NoError(t, err)
	assert.Equal(t, "Priya", *got.Name)
	assert.Equal(t, "Priya", user.Name)
}

func TestDeleteRemovesDeviceTokens(t *testing.T) {
	var clearedFor string

	tokens := &fakeDeviceTokenRepo{
		deleteAllForUserFn: func(_ context.Context, userID string) error {
			clearedFor = userID

			return nil
		},
	}
	uc := NewUserUseCase(&fakeUserRepo{}, tokens)

	err := uc.Delete(t.Context(), "user-1")

	require.NoError(t, err)
	assert.Equal(t, "user-1", clearedFor)
}

func TestDeleteSkipsTokensWhenUserMissing(t *testing.T) {
	repo := &fakeUserRepo{
		softDeleteFn: func(_ context.Context, id string) error {
			return domainerr.New(domain.ErrUserNotFound, id)
		},
	}
	tokens := &fakeDeviceTokenRepo{
		deleteAllForUserFn: func(_ context.Context, _ string) error {
			t.Fatal("tokens must not be cleared when the user does not exist")

			return nil
		},
	}
	uc := NewUserUseCase(repo, tokens)

	err := uc.Delete(t.Context(), "missing")

	assert.ErrorIs(t, err, domain.ErrUserNotFound)
}

func TestLookupSkipsInvalidAndDuplicatePhones(t *testing.T) {
	var got []string

	repo := &fakeUserRepo{
		listByPhonesFn: func(_ context.Context, phones []string) ([]*domain.User, error) {
			got = phones

			return nil, nil
		},
	}

	_, err := NewUserUseCase(repo, &fakeDeviceTokenRepo{}).Lookup(t.Context(), "caller", []string{testPhone, "12345", testPhone, "+14155550123"})
	require.NoError(t, err)

	assert.Equal(t, []string{testPhone, "+14155550123"}, got)
}

func TestLookupExcludesCaller(t *testing.T) {
	repo := &fakeUserRepo{
		listByPhonesFn: func(_ context.Context, _ []string) ([]*domain.User, error) {
			return []*domain.User{{ID: "caller"}, {ID: "friend"}}, nil
		},
	}

	users, err := NewUserUseCase(repo, &fakeDeviceTokenRepo{}).Lookup(t.Context(), "caller", []string{testPhone})
	require.NoError(t, err)

	require.Len(t, users, 1)
	assert.Equal(t, "friend", users[0].ID)
}

func TestLookupWithNoValidPhonesSkipsRepository(t *testing.T) {
	repo := &fakeUserRepo{
		listByPhonesFn: func(_ context.Context, _ []string) ([]*domain.User, error) {
			t.Fatal("repository must not be queried without valid phones")

			return nil, nil
		},
	}

	users, err := NewUserUseCase(repo, &fakeDeviceTokenRepo{}).Lookup(t.Context(), "caller", []string{"", "not a phone"})
	require.NoError(t, err)

	assert.Empty(t, users)
	assert.NotNil(t, users, "an empty result must still be a list, not null")
}

func TestLookupRejectsTooManyPhones(t *testing.T) {
	phones := make([]string, domain.MaxLookupPhones+1)

	_, err := NewUserUseCase(&fakeUserRepo{}, &fakeDeviceTokenRepo{}).Lookup(t.Context(), "caller", phones)

	assert.ErrorIs(t, err, domain.ErrTooManyPhones)
}
