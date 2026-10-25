// Package venues calls venue-service directly (not through the gateway), as
// the member acting, so venue-service's own membership rules apply.
package venues

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/SamPariat/roundup/services/event-service/internal/domain"
	"github.com/SamPariat/roundup/shared/api"
	"github.com/SamPariat/roundup/shared/domainerr"
	"github.com/SamPariat/roundup/shared/serviceclient"
)

// requestTimeout bounds each call. A place lookup may reach Google, so it
// gets longer than a membership check.
const requestTimeout = 10 * time.Second

// VenueServiceClient implements domain.VenueDirectory over HTTP.
type VenueServiceClient struct {
	client *serviceclient.Client
}

// venueDetail is the part of GET /api/v1/venues/{placeID} a nomination keeps.
type venueDetail struct {
	// PlaceID is the Google Place ID.
	PlaceID string `json:"placeID"`

	// Name is the venue's display name.
	Name string `json:"name"`

	// Address is the formatted street address.
	Address string `json:"address"`

	// Latitude is the WGS-84 latitude.
	Latitude float64 `json:"latitude"`

	// Longitude is the WGS-84 longitude.
	Longitude float64 `json:"longitude"`

	// AverageRating is the average rating, 1–5; 0 when unrated.
	AverageRating float32 `json:"averageRating"`

	// PriceLevel is 0–4 (0 = free or unknown, 4 = very expensive).
	PriceLevel int `json:"priceLevel"`

	// Photos are the venue's photos, best first; only the first is kept.
	Photos []struct {
		// URL is the photo's path on venue-service.
		URL string `json:"url"`

		// Width is the original width in pixels.
		Width int `json:"width"`

		// Height is the original height in pixels.
		Height int `json:"height"`
	} `json:"photos"`
}

// recordVisitRequest is the body of POST /api/v1/squads/{squadID}/history.
type recordVisitRequest struct {
	// EventID is the closed event; venue-service records one visit per event.
	EventID string `json:"eventID"`

	// PlaceID is the winning place.
	PlaceID string `json:"placeID"`

	// VisitedAt is when the night started.
	VisitedAt time.Time `json:"visitedAt"`
}

// NewVenueServiceClient constructs a client for the venue-service at
// baseURL, e.g. "http://localhost:3006".
func NewVenueServiceClient(baseURL string) *VenueServiceClient {
	return &VenueServiceClient{client: serviceclient.New(baseURL, requestTimeout)}
}

// Place returns placeID's display fields, with its first photo.
func (c *VenueServiceClient) Place(ctx context.Context, placeID, userID string) (domain.Place, error) {
	resp, err := c.client.Do(ctx, http.MethodGet, api.VenuesPath+"/"+url.PathEscape(placeID), userID, nil)
	if err != nil {
		return domain.Place{}, unavailable(err.Error())
	}
	defer func() { _ = resp.Body.Close() }()

	switch resp.StatusCode {
	case http.StatusOK:
	case http.StatusNotFound:
		return domain.Place{}, domainerr.New(domain.ErrVenueNotFound, placeID)
	default:
		return domain.Place{}, unavailable(fmt.Sprintf("unexpected status %d", resp.StatusCode))
	}

	var detail venueDetail
	if err := serviceclient.DecodeData(resp.Body, &detail); err != nil {
		return domain.Place{}, unavailable(err.Error())
	}

	place := domain.Place{
		PlaceID:    detail.PlaceID,
		Name:       detail.Name,
		Address:    detail.Address,
		Lat:        detail.Latitude,
		Lng:        detail.Longitude,
		Rating:     detail.AverageRating,
		PriceLevel: detail.PriceLevel,
	}

	if len(detail.Photos) > 0 {
		place.PhotoURL = detail.Photos[0].URL
		place.PhotoWidth = detail.Photos[0].Width
		place.PhotoHeight = detail.Photos[0].Height
	}

	return place, nil
}

// RecordVisit adds the event's winning place to the squad's history, dated
// when the night started. A visit already recorded counts as success.
func (c *VenueServiceClient) RecordVisit(ctx context.Context, e domain.Event, userID string) error {
	path := api.SquadsPath + "/" + url.PathEscape(e.SquadID) + "/history"
	body := recordVisitRequest{EventID: e.ID, PlaceID: e.Winner.PlaceID, VisitedAt: e.StartsAt}

	resp, err := c.client.Do(ctx, http.MethodPost, path, userID, body)
	if err != nil {
		return unavailable(err.Error())
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusConflict {
		return unavailable(fmt.Sprintf("record visit: unexpected status %d", resp.StatusCode))
	}

	return nil
}

func unavailable(detail string) error {
	return domainerr.New(domain.ErrVenueServiceUnavailable, detail)
}
