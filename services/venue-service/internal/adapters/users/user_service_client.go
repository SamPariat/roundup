// Package users provides a client for user-service, implementing
// domain.SquadDirectory.
package users

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/SamPariat/roundup/services/venue-service/internal/domain"
	"github.com/SamPariat/roundup/shared/api"
	"github.com/SamPariat/roundup/shared/domainerr"
	sharedmw "github.com/SamPariat/roundup/shared/middleware"
)

// requestTimeout bounds each membership check so a slow user-service fails
// the request fast instead of hanging it.
const requestTimeout = 3 * time.Second

// UserServiceClient implements domain.SquadDirectory by calling user-service's
// GET /api/v1/squads/{squadID} directly (not through the gateway) as the
// caller. That route answers 404 to non-members, so membership is its status.
type UserServiceClient struct {
	baseURL string
	http    *http.Client
}

// NewUserServiceClient constructs a client for the user-service at baseURL,
// e.g. "http://localhost:3002".
func NewUserServiceClient(baseURL string) *UserServiceClient {
	return &UserServiceClient{baseURL: baseURL, http: &http.Client{Timeout: requestTimeout}}
}

// EnsureMember returns nil if userID is a member of squadID. user-service
// answers 200 for members and 404 otherwise; anything else, including a
// timeout, is ErrUserServiceUnavailable.
func (c *UserServiceClient) EnsureMember(ctx context.Context, squadID, userID string) error {
	endpoint := c.baseURL + api.SquadsPath + "/" + url.PathEscape(squadID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return err
	}

	req.Header.Set(sharedmw.UserIDHeader, userID)
	// Pass the request ID on so user-service logs the same ID.
	if id := sharedmw.RequestIDFromContext(ctx); id != "" {
		req.Header.Set(sharedmw.RequestIDHeader, id)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return unavailable(err.Error())
	}
	defer func() { _ = resp.Body.Close() }()

	switch resp.StatusCode {
	case http.StatusOK:
		return nil
	case http.StatusNotFound:
		return domainerr.New(domain.ErrSquadNotFound, squadID)
	default:
		return unavailable(fmt.Sprintf("unexpected status %d", resp.StatusCode))
	}
}

// unavailable wraps detail in ErrUserServiceUnavailable.
func unavailable(detail string) error {
	return domainerr.New(domain.ErrUserServiceUnavailable, detail)
}
