package domain

import "time"

// device_token.go defines the DeviceToken type and the command structs for
// registering and removing FCM push tokens.

// Platform is the mobile OS a device token belongs to.
type Platform string

const (
	// PlatformIOS is an Apple device.
	PlatformIOS Platform = "iOS"
	// PlatformAndroid is an Android device.
	PlatformAndroid Platform = "Android"
)

// IsValid reports whether p is a supported platform.
func (p Platform) IsValid() bool {
	return p == PlatformIOS || p == PlatformAndroid
}

// DeviceToken is an FCM push token registered to a user's device.
type DeviceToken struct {
	// ID is the database-assigned primary key.
	ID int64
	// UserID is the user currently logged in on the device.
	UserID string
	// Token is the FCM registration token.
	Token string
	// Platform is the device's OS.
	Platform Platform
	// CreatedAt is the UTC timestamp when the token was first registered.
	CreatedAt time.Time
}

// RegisterDeviceTokenCommand carries the data needed to register a device token.
type RegisterDeviceTokenCommand struct {
	// UserID is the user registering the token.
	UserID string
	// Token is the FCM registration token.
	Token string
	// Platform is the device's OS.
	Platform Platform
}

// RemoveDeviceTokenCommand carries the data needed to remove a device token.
type RemoveDeviceTokenCommand struct {
	// UserID is the user removing the token. Only their own tokens can be removed.
	UserID string
	// Token is the FCM registration token.
	Token string
}
