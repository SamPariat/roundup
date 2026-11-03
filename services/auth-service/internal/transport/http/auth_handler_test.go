package http

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/SamPariat/roundup/services/auth-service/internal/application"
	"github.com/SamPariat/roundup/services/auth-service/internal/domain/domaintest"
	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type call struct {
	path, body string
	web        bool
	cookie     string
}

func send(t *testing.T, h *AuthHandler, in call) *http.Response {
	t.Helper()

	app := newApp(h)
	req := httptest.NewRequest(http.MethodPost, in.path, strings.NewReader(in.body))

	if in.body != "" {
		req.Header.Set("Content-Type", "application/json")
	}

	if in.web {
		req.Header.Set("X-Client", "web")
	}

	if in.cookie != "" {
		req.Header.Set("Cookie", "roundup_refresh="+in.cookie)
	}

	resp, err := app.Test(req)
	require.NoError(t, err)

	return resp
}

// tokens decodes data.refreshToken and data.accessToken from a success envelope.
func tokens(t *testing.T, resp *http.Response) (access, refresh string, hasRefreshKey bool) {
	t.Helper()

	var body struct {
		Data map[string]any `json:"data"`
	}

	require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))

	access, _ = body.Data["accessToken"].(string)
	refresh, hasRefreshKey = body.Data["refreshToken"].(string)

	return access, refresh, hasRefreshKey
}

func refreshCookieOf(t *testing.T, resp *http.Response) *http.Cookie {
	t.Helper()

	for _, c := range resp.Cookies() {
		if c.Name == "roundup_refresh" {
			return c
		}
	}

	return nil
}

func newApp(h *AuthHandler) *fiber.App {
	app := fiber.New()
	RegisterRoutes(app, &Handlers{Auth: h})

	return app
}

// newTestHandler builds the handler on the real OTP and token use cases, over
// the in-memory fakes, the same way cmd/main.go wires the Postgres adapters.
func newTestHandler(t *testing.T, secure bool) (*AuthHandler, *application.OTPUseCase, *domaintest.OTPSender) {
	t.Helper()

	_, priv := domaintest.NewKeys(t)
	sender := domaintest.NewOTPSender()
	sessions := domaintest.NewSessionRepo()
	issuer := application.NewTokenIssuer(priv, 15*time.Minute, 720*time.Hour)
	otp := application.NewOTPUseCase(
		&domaintest.OTPRepo{Clock: domaintest.NewClock()},
		sender,
		&domaintest.UserDirectory{Created: true},
		sessions,
		issuer,
		"",
	)

	return NewAuthHandler(otp, application.NewTokenUseCase(sessions, issuer), secure), otp, sender
}

const (
	verifyPath  = "/api/v1/auth/otp/verify"
	refreshPath = "/api/v1/auth/refresh"
	logoutPath  = "/api/v1/auth/logout"
)

// verifyBody requests a code for domaintest.Phone and returns the verify body
// that sends it back.
func verifyBody(t *testing.T, otp *application.OTPUseCase, sender *domaintest.OTPSender) string {
	t.Helper()

	require.NoError(t, otp.Request(t.Context(), domaintest.Phone))

	return `{"phone":"` + domaintest.Phone + `","code":"` + sender.Sent[domaintest.Phone] + `"}`
}

// login signs in as a native client and returns the refresh token it got.
func login(t *testing.T, h *AuthHandler, otp *application.OTPUseCase, sender *domaintest.OTPSender) string {
	t.Helper()

	resp := send(t, h, call{path: verifyPath, body: verifyBody(t, otp, sender)})
	require.Equal(t, http.StatusOK, resp.StatusCode)

	_, refresh, _ := tokens(t, resp)
	require.NotEmpty(t, refresh)

	return refresh
}

func refreshBody(token string) string { return `{"refreshToken":"` + token + `"}` }

func TestVerifyNativeGetsTokenInBodyAndNoCookie(t *testing.T) {
	h, otp, sender := newTestHandler(t, true)

	resp := send(t, h, call{path: verifyPath, body: verifyBody(t, otp, sender)})

	require.Equal(t, http.StatusOK, resp.StatusCode)
	access, refresh, _ := tokens(t, resp)
	assert.NotEmpty(t, access)
	assert.NotEmpty(t, refresh)
	assert.Nil(t, refreshCookieOf(t, resp))
}

func TestVerifyWebGetsCookieAndNoTokenInBody(t *testing.T) {
	h, otp, sender := newTestHandler(t, true)

	resp := send(t, h, call{path: verifyPath, body: verifyBody(t, otp, sender), web: true})

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var body struct {
		Data map[string]any `json:"data"`
	}

	require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
	assert.NotContains(t, body.Data, "refreshToken")
	assert.NotEmpty(t, body.Data["accessToken"])
	assert.Equal(t, true, body.Data["isNewUser"])

	expiresAt, err := time.Parse(time.RFC3339, body.Data["refreshTokenExpiresAt"].(string))
	require.NoError(t, err)

	cookie := refreshCookieOf(t, resp)
	require.NotNil(t, cookie)
	assert.NotEmpty(t, cookie.Value)
	assert.Equal(t, "/api/v1/auth", cookie.Path)
	assert.True(t, cookie.HttpOnly)
	assert.True(t, cookie.Secure)
	assert.Equal(t, http.SameSiteStrictMode, cookie.SameSite)
	assert.WithinDuration(t, expiresAt, cookie.Expires, time.Second, "expires with the refresh token")
}

func TestWebCookieIsNotSecureWhenConfiguredOff(t *testing.T) {
	h, otp, sender := newTestHandler(t, false)

	cookie := refreshCookieOf(t, send(t, h, call{path: verifyPath, body: verifyBody(t, otp, sender), web: true}))

	require.NotNil(t, cookie)
	assert.False(t, cookie.Secure)
}

func TestRefreshNativeUsesBodyAndReturnsTokenInBody(t *testing.T) {
	h, otp, sender := newTestHandler(t, true)
	token := login(t, h, otp, sender)

	resp := send(t, h, call{path: refreshPath, body: refreshBody(token)})

	require.Equal(t, http.StatusOK, resp.StatusCode)
	_, refresh, _ := tokens(t, resp)
	assert.NotEmpty(t, refresh)
	assert.NotEqual(t, token, refresh, "the token rotates")
	assert.Nil(t, refreshCookieOf(t, resp))
}

func TestRefreshWebReadsCookieAndRotatesIt(t *testing.T) {
	h, otp, sender := newTestHandler(t, true)
	token := login(t, h, otp, sender)

	resp := send(t, h, call{path: refreshPath, web: true, cookie: token})

	require.Equal(t, http.StatusOK, resp.StatusCode)

	access, _, hasRefresh := tokens(t, resp)
	assert.NotEmpty(t, access)
	assert.False(t, hasRefresh, "the rotated token is only in the cookie")

	cookie := refreshCookieOf(t, resp)
	require.NotNil(t, cookie)
	assert.NotEmpty(t, cookie.Value)
	assert.NotEqual(t, token, cookie.Value, "the cookie rotates")
	assert.True(t, cookie.HttpOnly)
	assert.Equal(t, http.SameSiteStrictMode, cookie.SameSite)
}

func TestRefreshBodyTokenWinsOverCookie(t *testing.T) {
	h, otp, sender := newTestHandler(t, true)
	token := login(t, h, otp, sender)

	resp := send(t, h, call{path: refreshPath, body: refreshBody(token), web: true, cookie: "stale"})

	assert.Equal(t, http.StatusOK, resp.StatusCode, "the stale cookie would have been a 401")
}

func TestRefreshWithoutAnyTokenIsBadRequest(t *testing.T) {
	tests := []struct {
		name string
		in   call
	}{
		{"NoBodyNoCookie", call{path: refreshPath, web: true}},
		{"EmptyJSONNoCookie", call{path: refreshPath, body: `{}`}},
		{"BrokenJSON", call{path: refreshPath, body: `{`}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h, _, _ := newTestHandler(t, true)

			resp := send(t, h, tt.in)

			assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		})
	}
}

func TestRefreshWithBadCookieIsUnauthorized(t *testing.T) {
	h, _, _ := newTestHandler(t, true)

	resp := send(t, h, call{path: refreshPath, web: true, cookie: "stale"})

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestLogoutWebClearsCookieOnSuccess(t *testing.T) {
	h, otp, sender := newTestHandler(t, true)
	token := login(t, h, otp, sender)

	resp := send(t, h, call{path: logoutPath, web: true, cookie: token})

	require.Equal(t, http.StatusNoContent, resp.StatusCode)
	assertCookieCleared(t, resp)
	assertLoggedOut(t, h, token)
}

func TestLogoutWebClearsCookieOnUnauthorized(t *testing.T) {
	h, _, _ := newTestHandler(t, true)

	resp := send(t, h, call{path: logoutPath, web: true, cookie: "stale"})

	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	assertCookieCleared(t, resp)
}

func TestLogoutWebClearsCookieWhenNothingWasSent(t *testing.T) {
	h, _, _ := newTestHandler(t, true)

	resp := send(t, h, call{path: logoutPath, web: true})

	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	assertCookieCleared(t, resp)
}

func TestLogoutNativeUsesBodyAndSetsNoCookie(t *testing.T) {
	h, otp, sender := newTestHandler(t, true)
	token := login(t, h, otp, sender)

	resp := send(t, h, call{path: logoutPath, body: refreshBody(token)})

	require.Equal(t, http.StatusNoContent, resp.StatusCode)
	assert.Nil(t, refreshCookieOf(t, resp))
	assertLoggedOut(t, h, token)
}

// assertLoggedOut checks the session is over: its refresh token no longer works.
func assertLoggedOut(t *testing.T, h *AuthHandler, token string) {
	t.Helper()

	resp := send(t, h, call{path: refreshPath, body: refreshBody(token)})

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode, "the logged-out session's token is rejected")
}

func assertCookieCleared(t *testing.T, resp *http.Response) {
	t.Helper()

	cookie := refreshCookieOf(t, resp)
	require.NotNil(t, cookie, "a Set-Cookie clears the stored token")
	assert.Empty(t, cookie.Value)
	assert.Equal(t, "/api/v1/auth", cookie.Path)
	assert.Negative(t, cookie.MaxAge, "Max-Age=0 on the wire")
	assert.True(t, cookie.HttpOnly)
}
