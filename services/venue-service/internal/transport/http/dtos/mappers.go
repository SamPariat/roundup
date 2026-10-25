package dtos

import (
	"net/url"

	"github.com/SamPariat/roundup/shared/pagination"

	"github.com/SamPariat/roundup/services/venue-service/internal/domain"
)

// photoPath is where clients load venue photos; the gateway serves it without
// a token so it works as an image source.
const photoPath = "/api/v1/venues/photos/"

// VenueToResponse maps a domain.Venue to a VenueResponse.
func VenueToResponse(v domain.Venue) VenueResponse {
	photos := make([]PhotoResponse, 0, len(v.PhotoRefs))
	for _, p := range v.PhotoRefs {
		photos = append(photos, photoToResponse(p))
	}

	return VenueResponse{
		PlaceID:       v.PlaceID,
		Name:          v.Name,
		Address:       v.Address,
		Latitude:      v.Latitude,
		Longitude:     v.Longitude,
		AverageRating: v.AverageRating,
		RatingCount:   v.RatingCount,
		PriceLevel:    v.PriceLevel,
		Types:         v.Types,
		IsOpen:        v.IsOpen,
		Photos:        photos,
	}
}

// VenueDetailToResponse maps a domain.VenueDetail and the caller's squads to
// a VenueDetailResponse.
func VenueDetailToResponse(d domain.VenueDetail, savedIn []domain.SquadSaved) VenueDetailResponse {
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
		SavedIn:          savedInToResponse(savedIn),
	}
}

// FavoritesToResponse maps one page of domain.SquadFavorite to a
// FavoritesResponse; next is the next page's cursor, "" on the last page.
func FavoritesToResponse(favorites []domain.SquadFavorite, next string) FavoritesResponse {
	responses := make([]FavoriteResponse, 0, len(favorites))
	for _, f := range favorites {
		savedBy := make([]MemberResponse, 0, len(f.SavedBy))
		for _, m := range f.SavedBy {
			savedBy = append(savedBy, MemberResponse{UserID: m.UserID, Name: m.Name, AvatarURL: m.AvatarURL})
		}

		responses = append(responses, FavoriteResponse{
			PlaceID:       f.PlaceID,
			Name:          f.Name,
			Address:       f.Address,
			AverageRating: f.AverageRating,
			RatingCount:   f.RatingCount,
			PriceLevel:    f.PriceLevel,
			Photo:         optionalPhoto(f.Photo),
			SavedBy:       savedBy,
			SavedByMe:     f.SavedByMe,
			LastSavedAt:   f.LastSavedAt,
		})
	}

	return FavoritesResponse{Favorites: responses, NextCursor: pagination.NextCursor(next)}
}

// VisitHistoryToResponse maps one page of domain.VisitSummary to a
// VisitHistoryResponse; next is the next page's cursor, "" on the last page.
func VisitHistoryToResponse(summaries []domain.VisitSummary, next string) VisitHistoryResponse {
	responses := make([]VisitSummaryResponse, 0, len(summaries))
	for _, s := range summaries {
		var avgSpend *int64
		if s.AvgSpendInPaise > 0 {
			avgSpend = &s.AvgSpendInPaise
		}

		responses = append(responses, VisitSummaryResponse{
			PlaceID:         s.PlaceID,
			Name:            s.Name,
			Address:         s.Address,
			Photo:           optionalPhoto(s.Photo),
			VisitCount:      s.VisitCount,
			LastVisitedAt:   s.LastVisitedAt,
			AvgSpendInPaise: avgSpend,
		})
	}

	return VisitHistoryResponse{History: responses, NextCursor: pagination.NextCursor(next)}
}

// SearchResultToResponse maps one page of search results to a SearchResponse.
func SearchResultToResponse(result *domain.SearchResult) SearchResponse {
	responses := make([]VenueResponse, 0, len(result.Venues))
	for _, v := range result.Venues {
		responses = append(responses, VenueToResponse(v))
	}

	return SearchResponse{Venues: responses, NextCursor: pagination.NextCursor(result.NextPageToken)}
}

// photoToResponse maps a domain.PhotoRef to a PhotoResponse.
func photoToResponse(p domain.PhotoRef) PhotoResponse {
	return PhotoResponse{
		URL:    photoPath + url.PathEscape(p.Reference),
		Width:  p.Width,
		Height: p.Height,
	}
}

// optionalPhoto maps p to a PhotoResponse, or nil when it has no reference.
func optionalPhoto(p domain.PhotoRef) *PhotoResponse {
	if p.Reference == "" {
		return nil
	}

	photo := photoToResponse(p)

	return &photo
}

// savedInToResponse maps the caller's squads to SavedInResponses, never nil.
func savedInToResponse(savedIn []domain.SquadSaved) []SavedInResponse {
	responses := make([]SavedInResponse, 0, len(savedIn))
	for _, s := range savedIn {
		responses = append(responses, SavedInResponse{SquadID: s.SquadID, SquadName: s.SquadName, Saved: s.Saved})
	}

	return responses
}
