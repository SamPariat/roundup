package places

import (
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
	"time"

	"github.com/SamPariat/roundup/services/venue-service/internal/domain"
)

// apiError is a non-2xx answer from Places API (New).
type apiError struct {
	status  int
	code    string // e.g. "INVALID_ARGUMENT"
	message string
	reason  string // from the ErrorInfo detail, e.g. "API_KEY_INVALID"
}

func (e *apiError) Error() string {
	return fmt.Sprintf("places: %d %s: %s", e.status, e.code, e.message)
}

// keyInvalid reports a bad or expired key. Google sends it as a 400, but it
// is our configuration, not the caller's input.
func (e *apiError) keyInvalid() bool {
	return e.reason == "API_KEY_INVALID"
}

// searchTextRequest is the body of places:searchText.
type searchTextRequest struct {
	TextQuery           string              `json:"textQuery"`
	IncludedType        string              `json:"includedType,omitempty"`
	StrictTypeFiltering bool                `json:"strictTypeFiltering,omitempty"`
	PageSize            int                 `json:"pageSize"`
	PageToken           string              `json:"pageToken,omitempty"`
	LocationRestriction locationRestriction `json:"locationRestriction"`
}

type locationRestriction struct {
	Rectangle rectangle `json:"rectangle"`
}

type rectangle struct {
	Low  latLng `json:"low"`
	High latLng `json:"high"`
}

type latLng struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type searchTextResponse struct {
	Places        []place `json:"places"`
	NextPageToken string  `json:"nextPageToken"`
}

// place is a Place resource; only the fields in the masks are filled.
type place struct {
	ID                       string         `json:"id"`
	DisplayName              localizedText  `json:"displayName"`
	FormattedAddress         string         `json:"formattedAddress"`
	Location                 latLng         `json:"location"`
	Rating                   float32        `json:"rating"`
	UserRatingCount          int            `json:"userRatingCount"`
	PriceLevel               string         `json:"priceLevel"`
	Types                    []string       `json:"types"`
	CurrentOpeningHours      *openingHours  `json:"currentOpeningHours"`
	RegularOpeningHours      *openingHours  `json:"regularOpeningHours"`
	Photos                   []photo        `json:"photos"`
	InternationalPhoneNumber string         `json:"internationalPhoneNumber"`
	WebsiteURI               string         `json:"websiteUri"`
	Reviews                  []review       `json:"reviews"`
	EditorialSummary         *localizedText `json:"editorialSummary"`
}

type localizedText struct {
	Text string `json:"text"`
}

type openingHours struct {
	OpenNow             *bool    `json:"openNow"`
	WeekdayDescriptions []string `json:"weekdayDescriptions"`
}

type photo struct {
	Name     string `json:"name"`
	WidthPx  int    `json:"widthPx"`
	HeightPx int    `json:"heightPx"`
}

type review struct {
	Rating            float32        `json:"rating"`
	Text              *localizedText `json:"text"`
	AuthorAttribution struct {
		DisplayName string `json:"displayName"`
	} `json:"authorAttribution"`
	PublishTime time.Time `json:"publishTime"`
}

// priceLevels is Google's price enum in order, so the index is the domain's 0–4.
var priceLevels = []string{
	"PRICE_LEVEL_FREE",
	"PRICE_LEVEL_INEXPENSIVE",
	"PRICE_LEVEL_MODERATE",
	"PRICE_LEVEL_EXPENSIVE",
	"PRICE_LEVEL_VERY_EXPENSIVE",
}

// toVenue converts a place into a domain.Venue. IsOpen is set only when
// Google knows the opening hours.
func (p place) toVenue() domain.Venue {
	venue := domain.Venue{
		PlaceID:       p.ID,
		Name:          p.DisplayName.Text,
		Address:       p.FormattedAddress,
		Latitude:      p.Location.Latitude,
		Longitude:     p.Location.Longitude,
		AverageRating: p.Rating,
		RatingCount:   p.UserRatingCount,
		PriceLevel:    max(slices.Index(priceLevels, p.PriceLevel), 0),
		Types:         p.Types,
		PhotoRefs:     make([]domain.PhotoRef, 0, len(p.Photos)),
	}

	if p.CurrentOpeningHours != nil {
		venue.IsOpen = p.CurrentOpeningHours.OpenNow
	}

	for _, ph := range p.Photos {
		venue.PhotoRefs = append(venue.PhotoRefs, domain.PhotoRef{
			Reference: encodePhotoReference(ph.Name),
			Width:     ph.WidthPx,
			Height:    ph.HeightPx,
		})
	}

	return venue
}

// toVenueDetail converts a place into a domain.VenueDetail.
func (p place) toVenueDetail() *domain.VenueDetail {
	detail := &domain.VenueDetail{
		Venue:       p.toVenue(),
		PhoneNumber: p.InternationalPhoneNumber,
		Website:     p.WebsiteURI,
		Reviews:     make([]domain.Review, 0, len(p.Reviews)),
	}

	if p.EditorialSummary != nil {
		detail.EditorialSummary = p.EditorialSummary.Text
	}

	if p.RegularOpeningHours != nil {
		detail.OpeningHours = p.RegularOpeningHours.WeekdayDescriptions
	}

	for _, r := range p.Reviews {
		var text string
		if r.Text != nil {
			text = r.Text.Text
		}

		detail.Reviews = append(detail.Reviews, domain.Review{
			AuthorName:  r.AuthorAttribution.DisplayName,
			Rating:      r.Rating,
			Text:        text,
			PublishedAt: r.PublishTime,
		})
	}

	return detail
}

// readAPIError builds an *apiError from Google's error body; a body it can't
// parse keeps just the status.
func readAPIError(resp *http.Response) error {
	var body struct {
		Error struct {
			Status  string `json:"status"`
			Message string `json:"message"`
			Details []struct {
				Reason string `json:"reason"`
			} `json:"details"`
		} `json:"error"`
	}

	_ = json.NewDecoder(resp.Body).Decode(&body)

	apiErr := &apiError{status: resp.StatusCode, code: body.Error.Status, message: body.Error.Message}
	for _, d := range body.Error.Details {
		if d.Reason != "" {
			apiErr.reason = d.Reason

			break
		}
	}

	return apiErr
}
