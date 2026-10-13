package application

import (
	"context"
	"testing"

	"github.com/SamPariat/roundup/services/user-service/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// device_token_usecase_test.go tests DeviceTokenUseCase using the fakeTokenRepo
// defined in user_usecase_test.go.

func TestRegisterRejectsEmptyToken(t *testing.T) {
	uc := NewDeviceTokenUseCase(&fakeTokenRepo{})

	_, err := uc.Register(context.Background(), domain.RegisterDeviceTokenCommand{
		UserID: "user-1", Token: "  ", Platform: domain.PlatformIOS,
	})

	assert.ErrorIs(t, err, domain.ErrInvalidToken)
}

func TestRegisterRejectsUnknownPlatform(t *testing.T) {
	uc := NewDeviceTokenUseCase(&fakeTokenRepo{})

	_, err := uc.Register(context.Background(), domain.RegisterDeviceTokenCommand{
		UserID: "user-1", Token: "fcm-token", Platform: "WINDOWS",
	})

	assert.ErrorIs(t, err, domain.ErrInvalidPlatform)
}

func TestRegisterSavesToken(t *testing.T) {
	repo := &fakeTokenRepo{
		upsertFn: func(ctx context.Context, cmd domain.RegisterDeviceTokenCommand) (*domain.DeviceToken, error) {
			return &domain.DeviceToken{ID: 1, UserID: cmd.UserID, Token: cmd.Token, Platform: cmd.Platform}, nil
		},
	}
	uc := NewDeviceTokenUseCase(repo)

	token, err := uc.Register(context.Background(), domain.RegisterDeviceTokenCommand{
		UserID: "user-1", Token: " fcm-token ", Platform: domain.PlatformAndroid,
	})

	require.NoError(t, err)
	assert.Equal(t, "fcm-token", token.Token)
	assert.Equal(t, domain.PlatformAndroid, token.Platform)
}

func TestRemovePropagatesNotFound(t *testing.T) {
	repo := &fakeTokenRepo{
		deleteFn: func(ctx context.Context, cmd domain.RemoveDeviceTokenCommand) error {
			return domain.NewDomainError(domain.ErrDeviceTokenNotFound, cmd.Token)
		},
	}
	uc := NewDeviceTokenUseCase(repo)

	err := uc.Remove(context.Background(), domain.RemoveDeviceTokenCommand{UserID: "user-1", Token: "x"})

	assert.ErrorIs(t, err, domain.ErrDeviceTokenNotFound)
}
