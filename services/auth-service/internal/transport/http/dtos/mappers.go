package dtos

import "github.com/SamPariat/roundup/services/auth-service/internal/domain"

// TokenPairToResponse maps a domain.TokenPair to a TokenPairResponse.
func TokenPairToResponse(p *domain.TokenPair) TokenPairResponse {
	return TokenPairResponse{
		AccessToken:           p.AccessToken,
		AccessTokenExpiresAt:  p.AccessTokenExpiresAt,
		RefreshToken:          p.RefreshToken,
		RefreshTokenExpiresAt: p.RefreshTokenExpiresAt,
	}
}

// LoginToResponse maps a domain.LoginResult to a LoginResponse.
func LoginToResponse(r *domain.LoginResult) LoginResponse {
	return LoginResponse{
		TokenPairResponse: TokenPairToResponse(&r.TokenPair),
		UserID:            r.UserID,
		IsNewUser:         r.IsNewUser,
	}
}
