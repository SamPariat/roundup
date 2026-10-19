// Package users provides a client for user-service, implementing
// domain.UserDirectory.
package users

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/SamPariat/roundup/services/auth-service/internal/domain"
	"github.com/SamPariat/roundup/shared/api"
	"github.com/SamPariat/roundup/shared/domainerr"
	sharedmw "github.com/SamPariat/roundup/shared/middleware"
	"github.com/gofiber/fiber/v3"
)

// requestTimeout bounds each call so a slow user-service fails a login fast
// instead of hanging it.
const requestTimeout = 3 * time.Second

// createUserResponse is the part of user-service's response envelope we read.
type createUserResponse struct {
	Data struct {
		ID string `json:"id"`
	} `json:"data"`
}

// UserServiceClient implements domain.UserDirectory by calling user-service's
// internal POST /api/v1/users endpoint directly (not through the gateway,
// which blocks that route).
type UserServiceClient struct {
	baseURL string
	http    *http.Client
}

// NewUserServiceClient constructs a client for the user-service at baseURL,
// e.g. "http://localhost:3002".
func NewUserServiceClient(baseURL string) *UserServiceClient {
	return &UserServiceClient{baseURL: baseURL, http: &http.Client{Timeout: requestTimeout}}
}

// EnsureUser finds or creates the user with this phone. user-service answers
// 201 for a new user and 200 for an existing one; anything else, including a
// timeout or a response without an ID, is ErrUserServiceUnavailable.
func (c *UserServiceClient) EnsureUser(ctx context.Context, phone string) (string, bool, error) {
	body, err := json.Marshal(map[string]string{"phone": phone})
	if err != nil {
		return "", false, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+api.UsersPath, bytes.NewReader(body))
	if err != nil {
		return "", false, err
	}

	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	// Pass the login's request ID on so user-service logs the same ID.
	if id := sharedmw.RequestIDFromContext(ctx); id != "" {
		req.Header.Set(sharedmw.RequestIDHeader, id)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return "", false, unavailable(err.Error())
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return "", false, unavailable(fmt.Sprintf("unexpected status %d", resp.StatusCode))
	}

	var decoded createUserResponse
	if err := json.NewDecoder(resp.Body).Decode(&decoded); err != nil || decoded.Data.ID == "" {
		return "", false, unavailable("response has no user ID")
	}

	return decoded.Data.ID, resp.StatusCode == http.StatusCreated, nil
}

func unavailable(detail string) error {
	return domainerr.New(domain.ErrUserServiceUnavailable, detail)
}
