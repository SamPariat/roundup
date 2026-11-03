package http

import (
	"context"
	"errors"

	"github.com/SamPariat/roundup/services/auth-service/internal/application"
	"github.com/SamPariat/roundup/services/auth-service/internal/domain"
	"github.com/SamPariat/roundup/services/auth-service/internal/transport/http/dtos"
	"github.com/SamPariat/roundup/shared/response"
	"github.com/gofiber/fiber/v3"
)

// otpService is what the handler needs from the OTP use case.
type otpService interface {
	Request(ctx context.Context, number string) error
	Verify(ctx context.Context, number, code string) (*domain.LoginResult, error)
}

// tokenService is what the handler needs from the token use case.
type tokenService interface {
	Refresh(ctx context.Context, token string) (*domain.TokenPair, error)
	Logout(ctx context.Context, token string) error
}

// AuthHandler handles HTTP requests for login, refresh, and logout. All of its
// routes are public: they are how a client gets a token in the first place.
//
// Web clients (X-Client: web) keep the refresh token in an httpOnly cookie
// instead of the response body; every other client gets it in the body.
type AuthHandler struct {
	otp          otpService
	tokens       tokenService
	cookieSecure bool
}

// NewAuthHandler constructs an AuthHandler with the given use cases.
// cookieSecure sets the Secure attribute of the web refresh cookie; it is
// false only for plain-http local development.
func NewAuthHandler(otp *application.OTPUseCase, tokens *application.TokenUseCase, cookieSecure bool) *AuthHandler {
	return &AuthHandler{otp: otp, tokens: tokens, cookieSecure: cookieSecure}
}

// RequestOTP handles POST /api/v1/auth/otp/request. Returns 202: the code is
// on its way, and the response says nothing more.
func (h *AuthHandler) RequestOTP(c fiber.Ctx) error {
	var req dtos.RequestOTPRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, err)
	}

	if err := h.otp.Request(c.Context(), req.Phone); err != nil {
		return response.Error(c, authErrorStatus(err), err)
	}

	return c.SendStatus(fiber.StatusAccepted)
}

// VerifyOTP handles POST /api/v1/auth/otp/verify.
func (h *AuthHandler) VerifyOTP(c fiber.Ctx) error {
	var req dtos.VerifyOTPRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, err)
	}

	result, err := h.otp.Verify(c.Context(), req.Phone, req.Code)
	if err != nil {
		return response.Error(c, authErrorStatus(err), err)
	}

	login := dtos.LoginToResponse(result)
	if isWebClient(c) {
		c.Cookie(refreshCookie(login.RefreshToken, login.RefreshTokenExpiresAt, h.cookieSecure))
		login.RefreshToken = ""
	}

	return c.Status(fiber.StatusOK).JSON(response.OK(login))
}

// Refresh handles POST /api/v1/auth/refresh. The refresh token comes from the
// body or, when the body has none, from the web refresh cookie.
func (h *AuthHandler) Refresh(c fiber.Ctx) error {
	token, err := refreshTokenFrom(c)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, err)
	}

	pair, err := h.tokens.Refresh(c.Context(), token)
	if err != nil {
		return response.Error(c, authErrorStatus(err), err)
	}

	res := dtos.TokenPairToResponse(pair)
	if isWebClient(c) {
		c.Cookie(refreshCookie(res.RefreshToken, res.RefreshTokenExpiresAt, h.cookieSecure))
		res.RefreshToken = ""
	}

	return c.Status(fiber.StatusOK).JSON(response.OK(res))
}

// Logout handles POST /api/v1/auth/logout. A web client always gets its cookie
// cleared, even when the token was rejected, so a stale cookie cannot trap it.
func (h *AuthHandler) Logout(c fiber.Ctx) error {
	if isWebClient(c) {
		c.Cookie(clearedRefreshCookie(h.cookieSecure))
	}

	token, err := refreshTokenFrom(c)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, err)
	}

	if err := h.tokens.Logout(c.Context(), token); err != nil {
		return response.Error(c, authErrorStatus(err), err)
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// errMissingRefreshToken is returned when a request carries no refresh token
// in either the body or the cookie.
var errMissingRefreshToken = errors.New("refresh token is required")

// refreshTokenFrom reads the refresh token from the JSON body, falling back to
// the web refresh cookie. The body is optional, so an empty one is not an error.
func refreshTokenFrom(c fiber.Ctx) (string, error) {
	var req dtos.RefreshTokenRequest

	if len(c.Body()) > 0 {
		if err := c.Bind().Body(&req); err != nil {
			return "", err
		}
	}

	if req.RefreshToken != "" {
		return req.RefreshToken, nil
	}

	if token := c.Cookies(refreshCookieName); token != "" {
		return token, nil
	}

	return "", errMissingRefreshToken
}

// authErrorStatus maps domain sentinel errors to HTTP status codes for auth.
func authErrorStatus(err error) int {
	switch {
	case errors.Is(err, domain.ErrInvalidPhone):
		return fiber.StatusBadRequest
	case errors.Is(err, domain.ErrInvalidCode), errors.Is(err, domain.ErrInvalidRefreshToken):
		return fiber.StatusUnauthorized
	case errors.Is(err, domain.ErrTooManyRequests), errors.Is(err, domain.ErrTooManyAttempts):
		return fiber.StatusTooManyRequests
	case errors.Is(err, domain.ErrUserServiceUnavailable):
		return fiber.StatusServiceUnavailable
	default:
		return fiber.StatusInternalServerError
	}
}
