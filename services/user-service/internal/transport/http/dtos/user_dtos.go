package dtos

import "time"

// user_dtos.go defines request and response structs for the user endpoints.
// Domain types are mapped into these structs at the handler boundary — JSON
// tags never appear in the domain layer.

// CreateUserRequest carries the JSON body for POST /api/v1/users.
type CreateUserRequest struct {
	Phone string `json:"phone"`
}

// UpdateProfileRequest carries the JSON body for PATCH /api/v1/users/me.
// Omitted fields are left unchanged.
type UpdateProfileRequest struct {
	Name      *string `json:"name"`
	AvatarURL *string `json:"avatarURL"`
}

// UserResponse is the JSON representation of a domain.User.
type UserResponse struct {
	ID        string    `json:"id"`
	Phone     string    `json:"phone"`
	Name      string    `json:"name"`
	AvatarURL string    `json:"avatarURL"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}
