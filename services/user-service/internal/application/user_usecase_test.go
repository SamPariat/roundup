package application

import (
	"context"
	"errors"
	"testing"

	"github.com/SamPariat/roundup/services/user-service/internal/domain"
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

type fakeTokenRepo struct {
	upsertFn           func(ctx context.Context, cmd domain.RegisterDeviceTokenCommand) (*domain.DeviceToken, error)
	deleteFn           func(ctx context.Context, cmd domain.RemoveDeviceTokenCommand) error
	deleteAllForUserFn func(ctx context.Context, userID string) error
}

func (f *fakeTokenRepo) Upsert(ctx context.Context, cmd domain.RegisterDeviceTokenCommand) (*domain.DeviceToken, error) {
	if f.upsertFn != nil {
		return f.upsertFn(ctx, cmd)
	}

	return nil, nil
}

func (f *fakeTokenRepo) Delete(ctx context.Context, cmd domain.RemoveDeviceTokenCommand) error {
	if f.deleteFn != nil {
		return f.deleteFn(ctx, cmd)
	}

	return nil
}

func (f *fakeTokenRepo) DeleteAllForUser(ctx context.Context, userID string) error {
	if f.deleteAllForUserFn != nil {
		return f.deleteAllForUserFn(ctx, userID)
	}

	return nil
}

const testPhone = "+919876543210"

func notFound(ctx context.Context, phone string) (*domain.User, error) {
	return nil, domain.NewDomainError(domain.ErrUserNotFound, phone)
}

func TestCreateOrGetRejectsInvalidPhone(t *testing.T) {
	uc := NewUserUseCase(&fakeUserRepo{}, &fakeTokenRepo{})

	_, _, err := uc.CreateOrGet(context.Background(), "98765")

	assert.ErrorIs(t, err, domain.ErrInvalidPhone)
}

func TestCreateOrGetReturnsExistingUser(t *testing.T) {
	existing := &domain.User{ID: "user-1", Phone: testPhone}
	repo := &fakeUserRepo{
		getByPhoneFn: func(ctx context.Context, phone string) (*domain.User, error) { return existing, nil },
		createFn: func(ctx context.Context, phone string) (*domain.User, error) {
			t.Fatal("Create must not be called when the user already exists")
			return nil, nil
		},
	}
	uc := NewUserUseCase(repo, &fakeTokenRepo{})

	user, created, err := uc.CreateOrGet(context.Background(), testPhone)

	require.NoError(t, err)
	assert.False(t, created)
	assert.Equal(t, "user-1", user.ID)
}

func TestCreateOrGetCreatesNewUser(t *testing.T) {
	repo := &fakeUserRepo{
		getByPhoneFn: notFound,
		createFn: func(ctx context.Context, phone string) (*domain.User, error) {
			return &domain.User{ID: "new-user", Phone: phone}, nil
		},
	}
	uc := NewUserUseCase(repo, &fakeTokenRepo{})

	user, created, err := uc.CreateOrGet(context.Background(), testPhone)

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
		createFn: func(ctx context.Context, phone string) (*domain.User, error) {
			return nil, domain.NewDomainError(domain.ErrPhoneTaken, phone)
		},
	}
	uc := NewUserUseCase(repo, &fakeTokenRepo{})

	user, created, err := uc.CreateOrGet(context.Background(), testPhone)

	require.NoError(t, err)
	assert.False(t, created)
	assert.Equal(t, "winner", user.ID)
}

func TestCreateOrGetPropagatesLookupFailure(t *testing.T) {
	dbDown := errors.New("connection refused")
	repo := &fakeUserRepo{
		getByPhoneFn: func(ctx context.Context, phone string) (*domain.User, error) { return nil, dbDown },
	}
	uc := NewUserUseCase(repo, &fakeTokenRepo{})

	_, _, err := uc.CreateOrGet(context.Background(), testPhone)

	assert.ErrorIs(t, err, dbDown)
}

func TestUpdateProfileRejectsBlankName(t *testing.T) {
	uc := NewUserUseCase(&fakeUserRepo{}, &fakeTokenRepo{})
	blank := "   "

	_, err := uc.UpdateProfile(context.Background(), domain.UpdateProfileCommand{UserID: "user-1", Name: &blank})

	assert.ErrorIs(t, err, domain.ErrInvalidName)
}

func TestUpdateProfileTrimsName(t *testing.T) {
	var got domain.UpdateProfileCommand
	repo := &fakeUserRepo{
		updateProfileFn: func(ctx context.Context, cmd domain.UpdateProfileCommand) (*domain.User, error) {
			got = cmd
			return &domain.User{ID: cmd.UserID, Name: *cmd.Name}, nil
		},
	}
	uc := NewUserUseCase(repo, &fakeTokenRepo{})
	name := "  Priya  "

	user, err := uc.UpdateProfile(context.Background(), domain.UpdateProfileCommand{UserID: "user-1", Name: &name})

	require.NoError(t, err)
	assert.Equal(t, "Priya", *got.Name)
	assert.Equal(t, "Priya", user.Name)
}

func TestDeleteRemovesDeviceTokens(t *testing.T) {
	var clearedFor string
	tokens := &fakeTokenRepo{
		deleteAllForUserFn: func(ctx context.Context, userID string) error {
			clearedFor = userID
			return nil
		},
	}
	uc := NewUserUseCase(&fakeUserRepo{}, tokens)

	err := uc.Delete(context.Background(), "user-1")

	require.NoError(t, err)
	assert.Equal(t, "user-1", clearedFor)
}

func TestDeleteSkipsTokensWhenUserMissing(t *testing.T) {
	repo := &fakeUserRepo{
		softDeleteFn: func(ctx context.Context, id string) error {
			return domain.NewDomainError(domain.ErrUserNotFound, id)
		},
	}
	tokens := &fakeTokenRepo{
		deleteAllForUserFn: func(ctx context.Context, userID string) error {
			t.Fatal("tokens must not be cleared when the user does not exist")
			return nil
		},
	}
	uc := NewUserUseCase(repo, tokens)

	err := uc.Delete(context.Background(), "missing")

	assert.ErrorIs(t, err, domain.ErrUserNotFound)
}
