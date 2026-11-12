package dtos

import (
	"github.com/SamPariat/roundup/shared/api"
	"github.com/SamPariat/roundup/shared/pagination"

	"github.com/SamPariat/roundup/services/venue-service/internal/domain"
)

// VenueToResponse maps a domain.Venue to a VenueResponse. urls holds each
// photo's direct URL by reference; photos without one are left out.
func VenueToResponse(v domain.Venue, urls map[string]string) VenueResponse {
	photos := make([]api.PhotoResponse, 0, len(v.PhotoRefs))
	for _, p := range v.PhotoRefs {
		if photo := optionalPhoto(p, urls); photo != nil {
			photos = append(photos, *photo)
		}
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
// a VenueDetailResponse, with the photos urls holds.
func VenueDetailToResponse(d domain.VenueDetail, savedIn []domain.SquadSaved, urls map[string]string) VenueDetailResponse {
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
		VenueResponse:    VenueToResponse(d.Venue, urls),
		PhoneNumber:      d.PhoneNumber,
		Website:          d.Website,
		OpeningHours:     d.OpeningHours,
		Reviews:          reviews,
		EditorialSummary: d.EditorialSummary,
		SavedIn:          savedInToResponse(savedIn),
	}
}

// FavoritesToResponse maps one page of domain.SquadFavorite to a
// FavoritesResponse; next is the next page's cursor, "" on the last page, and
// urls holds the photos' direct URLs.
func FavoritesToResponse(favorites []domain.SquadFavorite, next string, urls map[string]string) FavoritesResponse {
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
			Photo:         optionalPhoto(f.Photo, urls),
			SavedBy:       savedBy,
			SavedByMe:     f.SavedByMe,
			LastSavedAt:   f.LastSavedAt,
		})
	}

	return FavoritesResponse{Favorites: responses, NextCursor: pagination.NextCursor(next)}
}

// VisitHistoryToResponse maps one page of domain.VisitSummary to a
// VisitHistoryResponse; next is the next page's cursor, "" on the last page,
// and urls holds the photos' direct URLs.
func VisitHistoryToResponse(summaries []domain.VisitSummary, next string, urls map[string]string) VisitHistoryResponse {
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
			Photo:           optionalPhoto(s.Photo, urls),
			VisitCount:      s.VisitCount,
			LastVisitedAt:   s.LastVisitedAt,
			AvgSpendInPaise: avgSpend,
		})
	}

	return VisitHistoryResponse{History: responses, NextCursor: pagination.NextCursor(next)}
}

// SearchResultToResponse maps one page of search results to a SearchResponse,
// with the photos urls holds.
func SearchResultToResponse(result *domain.SearchResult, urls map[string]string) SearchResponse {
	responses := make([]VenueResponse, 0, len(result.Venues))
	for _, v := range result.Venues {
		responses = append(responses, VenueToResponse(v, urls))
	}

	return SearchResponse{Venues: responses, NextCursor: pagination.NextCursor(result.NextPageToken)}
}

// optionalPhoto maps p to a PhotoResponse with its URL from urls, or nil when
// urls has none for it (no photo, or its lookup failed).
func optionalPhoto(p domain.PhotoRef, urls map[string]string) *api.PhotoResponse {
	url, ok := urls[p.Reference]
	if !ok {
		return nil
	}

	attributions := make([]api.PhotoAttributionResponse, 0, len(p.Attributions))
	for _, a := range p.Attributions {
		attributions = append(attributions, api.PhotoAttributionResponse{Name: a.Name, URI: a.URI})
	}

	return &api.PhotoResponse{URL: url, Ref: p.Reference, Width: p.Width, Height: p.Height, Attributions: attributions}
}

// savedInToResponse maps the caller's squads to SavedInResponses, never nil.
func savedInToResponse(savedIn []domain.SquadSaved) []SavedInResponse {
	responses := make([]SavedInResponse, 0, len(savedIn))
	for _, s := range savedIn {
		responses = append(responses, SavedInResponse{SquadID: s.SquadID, SquadName: s.SquadName, Saved: s.Saved})
	}

	return responses
}
