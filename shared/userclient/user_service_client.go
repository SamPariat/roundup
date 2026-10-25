// Package userclient is the one client for user-service's squad routes,
// shared by every service that needs to know who is in which squad. A service
// maps its types and sentinel errors onto its own domain.
package userclient

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/SamPariat/roundup/shared/api"
	"github.com/SamPariat/roundup/shared/domainerr"
	"github.com/SamPariat/roundup/shared/pagination"
	"github.com/SamPariat/roundup/shared/serviceclient"
)

var (
	// ErrSquadNotFound is returned by Members when the squad does not exist or
	// the caller is not in it. user-service answers both with 404.
	ErrSquadNotFound = errors.New("squad not found")

	// ErrUserServiceUnavailable is returned when user-service cannot answer: it
	// is unreachable, timed out, or replied with an unexpected status or body.
	ErrUserServiceUnavailable = errors.New("user service unavailable")
)

// SquadMember is a squad member as user-service describes them. Role is the
// squad role, "ADMIN" or "MEMBER".
type SquadMember struct {
	// UserID is the member's user ID.
	UserID string

	// Name is the member's display name.
	Name string

	// AvatarURL is the member's avatar; "" when they have none.
	AvatarURL string

	// Role is the squad role, "ADMIN" or "MEMBER".
	Role string
}

// Squad is a squad as it appears in a user's list.
type Squad struct {
	// ID is the squad's UUID.
	ID string

	// Name is the squad's display name.
	Name string
}

// requestTimeout bounds each call so a slow user-service fails the request
// fast instead of hanging it.
const requestTimeout = 3 * time.Second

// UserServiceClient calls user-service directly (not through the gateway) as
// the caller, so user-service's own membership rules apply.
type UserServiceClient struct {
	client *serviceclient.Client
}

// NewUserServiceClient constructs a client for the user-service at baseURL,
// e.g. "http://localhost:3002".
func NewUserServiceClient(baseURL string) *UserServiceClient {
	return &UserServiceClient{client: serviceclient.New(baseURL, requestTimeout)}
}

// squadDetail is the part of user-service's GET /squads/{id} body we read.
type squadDetail struct {
	// Members are the squad's members.
	Members []struct {
		// UserID is the member's user ID.
		UserID string `json:"userID"`

		// Name is the member's display name.
		Name string `json:"name"`

		// AvatarURL is the member's avatar; "" when they have none.
		AvatarURL string `json:"avatarURL"`

		// Role is "ADMIN" or "MEMBER".
		Role string `json:"role"`
	} `json:"members"`
}

// squadList is the part of user-service's GET /squads body we read.
type squadList struct {
	// Squads are this page of the caller's squads.
	Squads []struct {
		// ID is the squad's UUID.
		ID string `json:"id"`

		// Name is the squad's display name.
		Name string `json:"name"`
	} `json:"squads"`

	// NextCursor fetches the next page; null on the last page.
	NextCursor *string `json:"nextCursor"`
}

// Members calls GET /api/v1/squads/{squadID}. That route answers 404 to
// non-members, so membership is its status; anything else but 200,
// including a timeout, is ErrUserServiceUnavailable.
func (c *UserServiceClient) Members(ctx context.Context, squadID, userID string) ([]SquadMember, error) {
	var body squadDetail

	status, err := c.get(ctx, api.SquadsPath+"/"+url.PathEscape(squadID), userID, &body)
	if err != nil {
		return nil, err
	}

	if status == http.StatusNotFound {
		return nil, domainerr.New(ErrSquadNotFound, squadID)
	}

	members := make([]SquadMember, 0, len(body.Members))
	for _, m := range body.Members {
		members = append(members, SquadMember{UserID: m.UserID, Name: m.Name, AvatarURL: m.AvatarURL, Role: m.Role})
	}

	return members, nil
}

// Squads calls GET /api/v1/squads, which lists the caller's squads, and
// follows nextCursor with the largest page size until the last page, so the
// result holds every squad.
func (c *UserServiceClient) Squads(ctx context.Context, userID string) ([]Squad, error) {
	squads := []Squad{}
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
			squads = append(squads, Squad{ID: s.ID, Name: s.Name})
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
	resp, err := c.client.Do(ctx, http.MethodGet, path, userID, nil)
	if err != nil {
		return 0, unavailable(err.Error())
	}
	defer func() { _ = resp.Body.Close() }()

	switch resp.StatusCode {
	case http.StatusOK:
		if err := serviceclient.DecodeData(resp.Body, out); err != nil {
			return 0, unavailable(err.Error())
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
	return domainerr.New(ErrUserServiceUnavailable, detail)
}
