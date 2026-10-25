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
	// status is the HTTP status code.
	status int
	// code is Google's status name, e.g. "INVALID_ARGUMENT".
	code string
	// message is Google's human-readable explanation.
	message string
	// reason is the ErrorInfo detail's reason, e.g. "API_KEY_INVALID"; "" when absent.
	reason string
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
	// TextQuery is what to look for, e.g. "rooftop bar".
	TextQuery string `json:"textQuery"`
	// IncludedType limits results to one Google place type, e.g. "bar"; omitted when empty.
	IncludedType string `json:"includedType,omitempty"`
	// StrictTypeFiltering drops places that don't have IncludedType.
	StrictTypeFiltering bool `json:"strictTypeFiltering,omitempty"`
	// PageSize is how many places one page holds, at most 20.
	PageSize int `json:"pageSize"`
	// PageToken is the previous page's NextPageToken; omitted for the first page.
	PageToken string `json:"pageToken,omitempty"`
	// LocationRestriction excludes places outside the search area.
	LocationRestriction locationRestriction `json:"locationRestriction"`
}

// locationRestriction is a hard search area; Text Search only accepts a rectangle.
type locationRestriction struct {
	// Rectangle is the area results must fall inside.
	Rectangle rectangle `json:"rectangle"`
}

// rectangle is a viewport given by two corners.
type rectangle struct {
	// Low is the south-west corner.
	Low latLng `json:"low"`
	// High is the north-east corner. A High longitude below Low's crosses 180°.
	High latLng `json:"high"`
}

// latLng is a WGS-84 point.
type latLng struct {
	// Latitude in degrees, -90 to 90.
	Latitude float64 `json:"latitude"`
	// Longitude in degrees, -180 to 180.
	Longitude float64 `json:"longitude"`
}

// searchTextResponse is one page of places:searchText results.
type searchTextResponse struct {
	// Places are this page's matches; absent when there are none.
	Places []place `json:"places"`
	// NextPageToken fetches the next page; "" on the last page.
	NextPageToken string `json:"nextPageToken"`
}

// place is a Place resource; only the fields in the masks are filled.
type place struct {
	// ID is the Google Place ID.
	ID string `json:"id"`
	// DisplayName is the place's name in the local language.
	DisplayName localizedText `json:"displayName"`
	// FormattedAddress is the full street address.
	FormattedAddress string `json:"formattedAddress"`
	// Location is where the place is.
	Location latLng `json:"location"`
	// Rating is the average rating, 1–5; 0 when unrated.
	Rating float32 `json:"rating"`
	// UserRatingCount is how many ratings Rating is based on.
	UserRatingCount int `json:"userRatingCount"`
	// PriceLevel is Google's price enum, e.g. "PRICE_LEVEL_MODERATE"; "" when unknown.
	PriceLevel string `json:"priceLevel"`
	// Types are Google's category tags, e.g. "bar", "restaurant".
	Types []string `json:"types"`
	// CurrentOpeningHours carries OpenNow; nil when Google doesn't know the hours.
	CurrentOpeningHours *openingHours `json:"currentOpeningHours"`
	// RegularOpeningHours carries the weekly schedule; nil when unknown.
	RegularOpeningHours *openingHours `json:"regularOpeningHours"`
	// Photos are up to 10 photos, best first.
	Photos []photo `json:"photos"`
	// InternationalPhoneNumber is the phone number with its country code.
	InternationalPhoneNumber string `json:"internationalPhoneNumber"`
	// WebsiteURI is the place's own website.
	WebsiteURI string `json:"websiteUri"`
	// Reviews are up to 5 recent reviews.
	Reviews []review `json:"reviews"`
	// EditorialSummary is Google's one-line description; nil when there is none.
	EditorialSummary *localizedText `json:"editorialSummary"`
}

// localizedText is text in one language.
type localizedText struct {
	// Text is the text itself.
	Text string `json:"text"`
}

// openingHours is the part of Google's opening hours the adapter reads.
type openingHours struct {
	// OpenNow is whether the place is open right now; nil when unknown.
	OpenNow *bool `json:"openNow"`
	// WeekdayDescriptions are one line per day, e.g. "Monday: 12:00 PM – 1:30 AM",
	// with narrow no-break spaces around the times.
	WeekdayDescriptions []string `json:"weekdayDescriptions"`
}

// photo is a reference to one place photo.
type photo struct {
	// Name is the photo's resource name, places/{placeId}/photos/{ref}.
	Name string `json:"name"`
	// WidthPx is the original width in pixels.
	WidthPx int `json:"widthPx"`
	// HeightPx is the original height in pixels.
	HeightPx int `json:"heightPx"`
}

// review is one user review.
type review struct {
	// Rating is the reviewer's score, 1–5.
	Rating float32 `json:"rating"`
	// Text is the review body; nil for a rating without text.
	Text *localizedText `json:"text"`
	// AuthorAttribution names the reviewer.
	AuthorAttribution struct {
		// DisplayName is the reviewer's public name.
		DisplayName string `json:"displayName"`
	} `json:"authorAttribution"`
	// PublishTime is when the review was posted.
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
