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

// FavoritesToResponse maps a slice of domain.SavedVenue to a FavoritesResponse.
func FavoritesToResponse(favorites []domain.SavedVenue) FavoritesResponse {
	responses := make([]SavedVenueResponse, 0, len(favorites))
	for _, f := range favorites {
		responses = append(responses, SavedVenueResponse{
			ID:      f.ID,
			PlaceID: f.PlaceID,
			Name:    f.Name,
			UserID:  f.UserID,
			SavedAt: f.SavedAt,
		})
	}

	return FavoritesResponse{Favorites: responses}
}

// VisitHistoryToResponse maps a slice of domain.VisitSummary to a VisitHistoryResponse.
func VisitHistoryToResponse(summaries []domain.VisitSummary) VisitHistoryResponse {
	responses := make([]VisitSummaryResponse, 0, len(summaries))
	for _, s := range summaries {
		responses = append(responses, VisitSummaryResponse{
			PlaceID:       s.PlaceID,
			Name:          s.Name,
			VisitCount:    s.VisitCount,
			LastVisitedAt: s.LastVisitedAt,
		})
	}
	return VisitHistoryResponse{History: responses}
}

// VenuesToSearchResponse maps a slice of domain.Venue to a SearchResponse.
func VenuesToSearchResponse(venues []domain.Venue) SearchResponse {
	responses := make([]VenueResponse, 0, len(venues))
	for _, v := range venues {
		responses = append(responses, VenueToResponse(v))
	}

	return SearchResponse{Venues: responses}
}
