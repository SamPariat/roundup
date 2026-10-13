package dtos

import "time"

// device_token_dtos.go defines request and response structs for the device token endpoints.

// RegisterDeviceTokenRequest carries the JSON body for POST /api/v1/users/me/device-tokens.
type RegisterDeviceTokenRequest struct {
	Token    string `json:"token"`
	Platform string `json:"platform"`
}

// DeviceTokenResponse is the JSON representation of a domain.DeviceToken.
type DeviceTokenResponse struct {
	ID        int64     `json:"id"`
	Token     string    `json:"token"`
	Platform  string    `json:"platform"`
	CreatedAt time.Time `json:"createdAt"`
}
