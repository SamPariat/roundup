package dtos

import "github.com/SamPariatIL/roundup/services/venue-service/internal/domain"

// mappers.go contains functions that translate domain types into DTO response types.
// Nothing outside this file should construct response DTOs manually.

// VenueToResponse maps a domain.Venue to a VenueResponse.
func VenueToResponse(v domain.Venue) VenueResponse {
	photos := make([]PhotoRefResponse, 0, len(v.PhotoRefs))
	for _, p := range v.PhotoRefs {
		photos = append(photos, PhotoRefResponse{
			Reference: p.Reference,
			Width:     p.Width,
			Height:    p.Height,
		})
	}

	return VenueResponse{
		PlaceID:       v.PlaceID,
		Name:          v.Name,
		Address:       v.Address,
		Latitude:      v.Latitude,
		Longitude:     v.Longitude,
		AverageRating: v.AverageRating,
		PriceLevel:    v.PriceLevel,
		Types:         v.Types,
		IsOpen:        v.IsOpen,
		PhotoRefs:     photos,
	}
}

// VenueDetailToResponse maps a domain.VenueDetail to a VenueDetailResponse.
func VenueDetailToResponse(d domain.VenueDetail) VenueDetailResponse {
	reviews := make([]ReviewResponse, 0, len(d.Reviews))
	for _, r := range d.Reviews {
		reviews = append(reviews, ReviewResponse{
			AuthorName:  r.AuthorName,
			Rating:      r.Rating,
			Text:        r.Text,
			PublishedAt: r.PublishedAt,
		})
	}

	return VenueDetailResponse{
		VenueResponse:    VenueToResponse(d.Venue),
		PhoneNumber:      d.PhoneNumber,
		Website:          d.Website,
		OpeningHours:     d.OpeningHours,
		Reviews:          reviews,
		EditorialSummary: d.EditorialSummary,
	}
}

// VenuesToSearchResponse maps a slice of domain.Venue to a SearchResponse.
func VenuesToSearchResponse(venues []domain.Venue) SearchResponse {
	responses := make([]VenueResponse, 0, len(venues))
	for _, v := range venues {
		responses = append(responses, VenueToResponse(v))
	}

	return SearchResponse{Venues: responses}
}
