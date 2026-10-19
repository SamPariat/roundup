package application

import (
	"context"
	"strings"

	"github.com/SamPariat/roundup/services/user-service/internal/domain"
)

// DeviceTokenUseCase registers and removes the FCM push tokens for a user's devices.
type DeviceTokenUseCase struct {
	tokens domain.DeviceTokenRepository
}

// NewDeviceTokenUseCase constructs a DeviceTokenUseCase with the given repository.
func NewDeviceTokenUseCase(tokens domain.DeviceTokenRepository) *DeviceTokenUseCase {
	return &DeviceTokenUseCase{tokens: tokens}
}

// Register saves the token for the user. Registering a token that belongs to
// another user moves it to this user, since a token identifies a device.
func (u *DeviceTokenUseCase) Register(ctx context.Context, cmd domain.RegisterDeviceTokenCommand) (*domain.DeviceToken, error) {
	cmd.Token = strings.TrimSpace(cmd.Token)
	if cmd.Token == "" {
		return nil, domain.ErrInvalidToken
	}

	if !cmd.Platform.IsValid() {
		return nil, domain.ErrInvalidPlatform
	}

	return u.tokens.Upsert(ctx, cmd)
}

// Remove deletes the token if the user owns it.
// Returns domain.ErrDeviceTokenNotFound otherwise.
func (u *DeviceTokenUseCase) Remove(ctx context.Context, cmd domain.RemoveDeviceTokenCommand) error {
	return u.tokens.Delete(ctx, cmd)
}
