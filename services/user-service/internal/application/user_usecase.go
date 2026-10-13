package application

import (
	"context"
	"errors"
	"strings"

	"github.com/SamPariat/roundup/services/user-service/internal/domain"
)

// user_usecase.go implements UserUseCase, which handles creating, reading,
// updating, and deleting user profiles.

// UserUseCase handles user profile operations.
type UserUseCase struct {
	users  domain.UserRepository
	tokens domain.DeviceTokenRepository
}

// NewUserUseCase constructs a UserUseCase with the given repositories.
func NewUserUseCase(users domain.UserRepository, tokens domain.DeviceTokenRepository) *UserUseCase {
	return &UserUseCase{users: users, tokens: tokens}
}

// CreateOrGet returns the active user with the given phone number, creating
// one if none exists. It is idempotent: calling it twice with the same phone
// returns the same user. created reports whether a new user was inserted.
// Returns domain.ErrInvalidPhone if phone is not in E.164 format.
func (u *UserUseCase) CreateOrGet(ctx context.Context, phone string) (user *domain.User, created bool, err error) {
	if !domain.IsValidPhone(phone) {
		return nil, false, domain.ErrInvalidPhone
	}

	user, err = u.users.GetByPhone(ctx, phone)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			user, err = u.users.Create(ctx, phone)
			if err != nil {
				// Another request inserted this phone between our lookup and our insert.
				// The user exists now, so fetch it instead of failing.
				if errors.Is(err, domain.ErrPhoneTaken) {
					user, err = u.users.GetByPhone(ctx, phone)
					if err != nil {
						return nil, false, err
					}

					return user, false, nil
				}

				return nil, false, err
			}

			return user, true, nil
		}

		return nil, false, err
	}

	return user, false, nil
}

// Get returns the active user with the given ID.
// Returns domain.ErrUserNotFound if the user does not exist or was deleted.
func (u *UserUseCase) Get(ctx context.Context, id string) (*domain.User, error) {
	return u.users.GetByID(ctx, id)
}

// UpdateProfile applies a partial profile update and returns the updated user.
// Returns domain.ErrInvalidName if the update sets a blank name.
func (u *UserUseCase) UpdateProfile(ctx context.Context, cmd domain.UpdateProfileCommand) (*domain.User, error) {
	if cmd.Name != nil {
		trimmed := strings.TrimSpace(*cmd.Name)
		if trimmed == "" {
			return nil, domain.ErrInvalidName
		}
		cmd.Name = &trimmed
	}

	return u.users.UpdateProfile(ctx, cmd)
}

// Delete soft-deletes the user and removes their device tokens so they stop
// receiving push notifications.
func (u *UserUseCase) Delete(ctx context.Context, id string) error {
	if err := u.users.SoftDelete(ctx, id); err != nil {
		return err
	}

	return u.tokens.DeleteAllForUser(ctx, id)
}
