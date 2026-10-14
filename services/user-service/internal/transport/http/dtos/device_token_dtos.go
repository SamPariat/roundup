package dtos

import "time"

// RegisterDeviceTokenRequest carries the JSON body for POST /api/v1/users/me/device-tokens.
type RegisterDeviceTokenRequest struct {
	// Token is the FCM registration token for this device.
	Token string `json:"token"`
	// Platform is "iOS" or "Android" (case-sensitive).
	Platform string `json:"platform"`
}

// DeviceTokenResponse is the JSON representation of a domain.DeviceToken.
type DeviceTokenResponse struct {
	// ID is the server-assigned ID of the token record.
	ID int64 `json:"id"`
	// Token is the FCM registration token.
	Token string `json:"token"`
	// Platform is "iOS" or "Android".
	Platform string `json:"platform"`
	// CreatedAt is when the token was first registered, as an RFC 3339 timestamp.
	// Registering the same token again keeps the original time.
	CreatedAt time.Time `json:"createdAt"`
}
