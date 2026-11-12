package api

import "slices"

// PhotoWidths are the image widths, in pixels, a client may ask photos for.
// The client picks the smallest that covers its slot on the screen (points
// times pixel density), so the same few widths repeat across devices and a
// cached photo URL serves all of them.
var PhotoWidths = []int{200, 400, 600, 800, 1200, 1600}

// DefaultPhotoWidth is the width used when a request does not ask for one.
const DefaultPhotoWidth = 800

// MaxVenuePhotos is the most photos a venue's page may ask for.
const MaxVenuePhotos = 5

// IsPhotoWidth reports whether width is one of PhotoWidths.
func IsPhotoWidth(width int) bool {
	return slices.Contains(PhotoWidths, width)
}

// VenuePhotoURLsPath is venue-service's internal route that turns stored
// photo references into fresh URLs. event-service calls POST on it for the
// places it keeps; the gateway never routes it.
const VenuePhotoURLsPath = InternalV1 + "/venues/photo-urls"

// MaxPhotoURLRefs caps one PhotoURLsRequest's refs; each uncached one is a
// billed Google call.
const MaxPhotoURLRefs = 50

// PhotoResponse is a venue photo the client can load directly. venue-service
// and event-service both answer with it, as the spec's Photo schema.
type PhotoResponse struct {
	// URL is the image on Google's servers. It needs no key or token, so it
	// works as an <img> or <Image> source, but it expires: use it from a
	// fresh response, never store it.
	URL string `json:"url"`

	// Ref is the photo's opaque reference. A service that keeps a venue's
	// photo stores this and turns it into a fresh URL with VenuePhotoURLsPath.
	Ref string `json:"ref"`

	// Width is the original image width in pixels.
	Width int `json:"width"`

	// Height is the original image height in pixels.
	Height int `json:"height"`

	// Attributions credit the photo's authors and must be shown with it.
	// Empty, never null, when there are none.
	Attributions []PhotoAttributionResponse `json:"attributions"`
}

// PhotoAttributionResponse credits one author of a photo.
type PhotoAttributionResponse struct {
	// Name is the author's display name.
	Name string `json:"name"`

	// URI is the author's profile page; "" when Google gives none.
	URI string `json:"uri"`
}

// PhotoURLsRequest is the body of POST VenuePhotoURLsPath.
type PhotoURLsRequest struct {
	// Refs are photo references, as in a PhotoResponse's Ref; at most MaxPhotoURLRefs.
	Refs []string `json:"refs"`

	// MaxWidth is the widest the images may be, one of PhotoWidths.
	MaxWidth int `json:"maxWidth"`
}

// PhotoURLsResponse is the data of POST VenuePhotoURLsPath's answer.
type PhotoURLsResponse struct {
	// URLs maps each ref to its direct URL. A ref whose lookup failed is left out.
	URLs map[string]string `json:"urls"`
}
