package dtos

import "time"

// RequestOTPRequest carries the JSON body for POST /api/v1/auth/otp/request.
type RequestOTPRequest struct {
	// Phone is the number to send the code to, in E.164 format, e.g. "+919876543210".
	Phone string `json:"phone"`
}

// VerifyOTPRequest carries the JSON body for POST /api/v1/auth/otp/verify.
type VerifyOTPRequest struct {
	// Phone is the number the code was sent to, in E.164 format.
	Phone string `json:"phone"`
	// Code is the 6-digit code from the SMS.
	Code string `json:"code"`
}

// RefreshTokenRequest carries the JSON body for POST /api/v1/auth/refresh and
// POST /api/v1/auth/logout.
type RefreshTokenRequest struct {
	// RefreshToken is the refresh token from the last login or refresh.
	RefreshToken string `json:"refreshToken"`
}

// TokenPairResponse is returned by refresh, and embedded in LoginResponse.
type TokenPairResponse struct {
	// AccessToken is a JWT to send as "Authorization: Bearer <token>".
	AccessToken string `json:"accessToken"`
	// AccessTokenExpiresAt is when the access token expires (RFC 3339);
	// refresh shortly before this.
	AccessTokenExpiresAt time.Time `json:"accessTokenExpiresAt"`
	// RefreshToken replaces the previous one, which no longer works. Store it
	// securely (Keychain / Keystore); sending an old one logs the session out.
	RefreshToken string `json:"refreshToken"`
	// RefreshTokenExpiresAt is when the refresh token expires (RFC 3339);
	// after it, the user must log in again.
	RefreshTokenExpiresAt time.Time `json:"refreshTokenExpiresAt"`
}

// LoginResponse is returned by POST /api/v1/auth/otp/verify. The embedded
// token fields appear at the top level of the JSON object.
type LoginResponse struct {
	TokenPairResponse
	// UserID is the logged-in user's UUID.
	UserID string `json:"userID"`
	// IsNewUser is true when this login created the account; show profile setup.
	IsNewUser bool `json:"isNewUser"`
}
