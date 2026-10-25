// Package users provides a client for user-service, implementing
// domain.SquadDirectory.
package users

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/SamPariat/roundup/services/venue-service/internal/domain"
	"github.com/SamPariat/roundup/shared/api"
	"github.com/SamPariat/roundup/shared/domainerr"
	sharedmw "github.com/SamPariat/roundup/shared/middleware"
	"github.com/SamPariat/roundup/shared/pagination"
	"github.com/SamPariat/roundup/shared/response"
)

// requestTimeout bounds each call so a slow user-service fails the request
// fast instead of hanging it.
const requestTimeout = 3 * time.Second

// UserServiceClient implements domain.SquadDirectory by calling user-service
// directly (not through the gateway) as the caller.
type UserServiceClient struct {
	baseURL string
	http    *http.Client
}

// NewUserServiceClient constructs a client for the user-service at baseURL,
// e.g. "http://localhost:3002".
func NewUserServiceClient(baseURL string) *UserServiceClient {
	return &UserServiceClient{baseURL: baseURL, http: &http.Client{Timeout: requestTimeout}}
}

// squadDetail is the part of user-service's GET /squads/{id} body we read.
type squadDetail struct {
	Members []struct {
		UserID    string `json:"userID"`
		Name      string `json:"name"`
		AvatarURL string `json:"avatarURL"`
	} `json:"members"`
}

// squadList is the part of user-service's GET /squads body we read.
type squadList struct {
	Squads []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"squads"`

	NextCursor *string `json:"nextCursor"`
}

// Members calls GET /api/v1/squads/{squadID}. That route answers 404 to
// non-members, so membership is its status; anything else but 200,
// including a timeout, is ErrUserServiceUnavailable.
func (c *UserServiceClient) Members(ctx context.Context, squadID, userID string) ([]domain.SquadMember, error) {
	var body squadDetail

	status, err := c.get(ctx, api.SquadsPath+"/"+url.PathEscape(squadID), userID, &body)
	if err != nil {
		return nil, err
	}

	if status == http.StatusNotFound {
		return nil, domainerr.New(domain.ErrSquadNotFound, squadID)
	}

	members := make([]domain.SquadMember, 0, len(body.Members))
	for _, m := range body.Members {
		members = append(members, domain.SquadMember{UserID: m.UserID, Name: m.Name, AvatarURL: m.AvatarURL})
	}

	return members, nil
}

// Squads calls GET /api/v1/squads, which lists the caller's squads, and
// follows nextCursor with the largest page size until the last page, so the
// result holds every squad.
func (c *UserServiceClient) Squads(ctx context.Context, userID string) ([]domain.Squad, error) {
	squads := []domain.Squad{}
	query := url.Values{"limit": {strconv.Itoa(pagination.MaxLimit)}}

	for {
		var body squadList

		status, err := c.get(ctx, api.SquadsPath+"?"+query.Encode(), userID, &body)
		if err != nil {
			return nil, err
		}

		if status != http.StatusOK {
			return nil, unavailable(fmt.Sprintf("unexpected status %d", status))
		}

		for _, s := range body.Squads {
			squads = append(squads, domain.Squad{ID: s.ID, Name: s.Name})
		}

		if body.NextCursor == nil {
			return squads, nil
		}

		query.Set("cursor", *body.NextCursor)
	}
}

// get sends a GET to path as userID and decodes a 200 body's data into out.
// It returns the status for 200 and 404 and ErrUserServiceUnavailable for
// anything else, a transport failure, or an undecodable body.
func (c *UserServiceClient) get(ctx context.Context, path, userID string, out any) (int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return 0, err
	}

	req.Header.Set(sharedmw.UserIDHeader, userID)
	// Pass the request ID on so user-service logs the same ID.
	if id := sharedmw.RequestIDFromContext(ctx); id != "" {
		req.Header.Set(sharedmw.RequestIDHeader, id)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return 0, unavailable(err.Error())
	}
	defer func() { _ = resp.Body.Close() }()

	switch resp.StatusCode {
	case http.StatusOK:
		envelope := response.Response[any]{Data: out}
		if err := json.NewDecoder(resp.Body).Decode(&envelope); err != nil {
			return 0, unavailable("decode response: " + err.Error())
		}

		return resp.StatusCode, nil
	case http.StatusNotFound:
		return resp.StatusCode, nil
	default:
		return 0, unavailable(fmt.Sprintf("unexpected status %d", resp.StatusCode))
	}
}

// unavailable wraps detail in ErrUserServiceUnavailable.
func unavailable(detail string) error {
	return domainerr.New(domain.ErrUserServiceUnavailable, detail)
}
