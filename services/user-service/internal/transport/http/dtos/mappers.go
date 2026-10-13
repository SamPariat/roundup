package dtos

import "github.com/SamPariat/roundup/services/user-service/internal/domain"

// mappers.go contains functions that translate domain types into DTO response types.
// Nothing outside this file should construct response DTOs manually.

// UserToResponse maps a domain.User to a UserResponse.
func UserToResponse(u *domain.User) UserResponse {
	return UserResponse{
		ID:        u.ID,
		Phone:     u.Phone,
		Name:      u.Name,
		AvatarURL: u.AvatarURL,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}

// DeviceTokenToResponse maps a domain.DeviceToken to a DeviceTokenResponse.
func DeviceTokenToResponse(t *domain.DeviceToken) DeviceTokenResponse {
	return DeviceTokenResponse{
		ID:        t.ID,
		Token:     t.Token,
		Platform:  string(t.Platform),
		CreatedAt: t.CreatedAt,
	}
}
