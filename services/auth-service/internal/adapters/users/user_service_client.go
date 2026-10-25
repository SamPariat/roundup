// Package users provides a client for user-service, implementing
// domain.UserDirectory.
package users

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/SamPariat/roundup/services/auth-service/internal/domain"
	"github.com/SamPariat/roundup/shared/api"
	"github.com/SamPariat/roundup/shared/domainerr"
	"github.com/SamPariat/roundup/shared/serviceclient"
)

// requestTimeout bounds each call so a slow user-service fails a login fast
// instead of hanging it.
const requestTimeout = 3 * time.Second

// createdUser is the part of user-service's response data we read.
type createdUser struct {
	// ID is the new (or existing) user's UUID.
	ID string `json:"id"`
}

// UserServiceClient implements domain.UserDirectory by calling user-service's
// internal POST /api/v1/users endpoint directly (not through the gateway,
// which blocks that route).
type UserServiceClient struct {
	client *serviceclient.Client
}

// NewUserServiceClient constructs a client for the user-service at baseURL,
// e.g. "http://localhost:3002".
func NewUserServiceClient(baseURL string) *UserServiceClient {
	return &UserServiceClient{client: serviceclient.New(baseURL, requestTimeout)}
}

// EnsureUser finds or creates the user with this phone. user-service answers
// 201 for a new user and 200 for an existing one; anything else, including a
// timeout or a response without an ID, is ErrUserServiceUnavailable.
func (c *UserServiceClient) EnsureUser(ctx context.Context, phone string) (string, bool, error) {
	resp, err := c.client.Do(ctx, http.MethodPost, api.UsersPath, "", map[string]string{"phone": phone})
	if err != nil {
		return "", false, unavailable(err.Error())
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return "", false, unavailable(fmt.Sprintf("unexpected status %d", resp.StatusCode))
	}

	var user createdUser
	if err := serviceclient.DecodeData(resp.Body, &user); err != nil || user.ID == "" {
		return "", false, unavailable("response has no user ID")
	}

	return user.ID, resp.StatusCode == http.StatusCreated, nil
}

func unavailable(detail string) error {
	return domainerr.New(domain.ErrUserServiceUnavailable, detail)
}
