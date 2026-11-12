package dtos

import (
	"github.com/SamPariat/roundup/services/event-service/internal/application"
	"github.com/SamPariat/roundup/services/event-service/internal/domain"
	"github.com/SamPariat/roundup/shared/api"
	"github.com/SamPariat/roundup/shared/pagination"
)

// EventToResponse maps a domain.Event to an EventResponse; urls holds the
// photos' fresh URLs by reference.
func EventToResponse(e domain.Event, urls map[string]string) EventResponse {
	resp := EventResponse{
		ID:             e.ID,
		SquadID:        e.SquadID,
		Name:           e.Name,
		Type:           string(e.Type),
		Status:         string(e.Status),
		CreatedBy:      e.CreatedBy,
		StartsAt:       e.StartsAt,
		VotingDeadline: e.VotingDeadline,
		ClosedAt:       e.ClosedAt,
		CreatedAt:      e.CreatedAt,
	}

	if e.Winner != nil {
		resp.Winner = &WinnerResponse{PlaceResponse: placeToResponse(e.Winner.Place, urls)}
	}

	return resp
}

// SquadEventsToResponse maps a page of a squad's events, with the photos urls holds.
func SquadEventsToResponse(page *application.SquadEvents, urls map[string]string) SquadEventsResponse {
	events := make([]EventSummaryResponse, 0, len(page.Events))
	for _, s := range page.Events {
		events = append(events, EventSummaryResponse{
			EventResponse: EventToResponse(s.Event, urls),
			RSVPCounts: RSVPCountsResponse{
				Going:    s.RSVPCounts.Going,
				Maybe:    s.RSVPCounts.Maybe,
				NotGoing: s.RSVPCounts.NotGoing,
				Pending:  s.RSVPCounts.Pending,
			},
			GoingPreview:    membersToResponse(s.GoingPreview),
			NominationCount: s.NominationCount,
			MyRSVP:          optional(string(s.MyRSVP)),
			HasVoted:        s.HasVoted,
		})
	}

	return SquadEventsResponse{
		Events:     events,
		NextCursor: pagination.NextCursor(page.NextCursor),
		OpenCount:  page.OpenCount,
	}
}

// EventDetailToResponse maps a domain.EventDetail, with the photos urls holds.
func EventDetailToResponse(d *domain.EventDetail, urls map[string]string) EventDetailResponse {
	resp := EventDetailResponse{
		EventResponse: EventToResponse(d.Event, urls),
		CreatedByName: optional(d.CreatedByName),
		MemberCount:   len(d.Members),
		VotedCount:    d.VotedCount,
		TotalVotes:    d.TotalVotes,
		MyRSVP:        optional(string(d.MyRSVP)),
		MyVote:        optional(d.MyVote),
		Members:       make([]MemberRSVPResponse, 0, len(d.Members)),
		Nominations:   make([]NominationResponse, 0, len(d.Nominations)),
		Permissions: PermissionsResponse{
			CanEdit:        d.Permissions.CanEdit,
			CanRSVP:        d.Permissions.CanRSVP,
			CanNominate:    d.Permissions.CanNominate,
			CanVote:        d.Permissions.CanVote,
			CanStartVoting: d.Permissions.CanStartVoting,
			CanCloseVoting: d.Permissions.CanCloseVoting,
			CanStart:       d.Permissions.CanStart,
			CanClose:       d.Permissions.CanClose,
		},
	}

	for _, m := range d.Members {
		resp.Members = append(resp.Members, MemberRSVPResponse{
			MemberResponse: memberToResponse(m.SquadMember),
			RSVP:           optional(string(m.RSVP)),
			IsMe:           m.IsMe,
		})
	}

	for _, n := range d.Nominations {
		resp.Nominations = append(resp.Nominations, NominationResponse{
			PlaceResponse:   placeToResponse(n.Place, urls),
			NominatedBy:     n.NominatedBy,
			NominatedByName: optional(n.NominatedByName),
			NominatedAt:     n.NominatedAt,
			VoteCount:       n.VoteCount,
			Voters:          membersToResponse(n.Voters),
			VotedByMe:       n.VotedByMe,
			CanRemove:       n.CanRemove,
		})

		if resp.Winner != nil && resp.Winner.PlaceID == n.PlaceID {
			votes := n.VoteCount
			resp.Winner.Votes = &votes
		}
	}

	return resp
}

// placeToResponse maps p; its photo is included only when urls has a URL for
// it (none stored, or the lookup failed, leaves photo null).
func placeToResponse(p domain.Place, urls map[string]string) PlaceResponse {
	resp := PlaceResponse{
		PlaceID:    p.PlaceID,
		Name:       p.Name,
		Address:    p.Address,
		Latitude:   p.Lat,
		Longitude:  p.Lng,
		Rating:     p.Rating,
		PriceLevel: p.PriceLevel,
	}

	if url, ok := urls[p.PhotoRef]; ok && p.PhotoRef != "" {
		attributions := []api.PhotoAttributionResponse{}
		if p.PhotoAuthorName != "" {
			attributions = append(attributions, api.PhotoAttributionResponse{Name: p.PhotoAuthorName, URI: p.PhotoAuthorURI})
		}

		resp.Photo = &api.PhotoResponse{
			URL:          url,
			Ref:          p.PhotoRef,
			Width:        p.PhotoWidth,
			Height:       p.PhotoHeight,
			Attributions: attributions,
		}
	}

	return resp
}

// memberToResponse maps a squad member to a MemberResponse.
func memberToResponse(m domain.SquadMember) MemberResponse {
	return MemberResponse{UserID: m.UserID, Name: m.Name, AvatarURL: m.AvatarURL}
}

// membersToResponse maps squad members to MemberResponses, never nil.
func membersToResponse(members []domain.SquadMember) []MemberResponse {
	resp := make([]MemberResponse, 0, len(members))
	for _, m := range members {
		resp = append(resp, memberToResponse(m))
	}

	return resp
}

// optional returns nil for "", so an unset value serialises as JSON null.
func optional(s string) *string {
	if s == "" {
		return nil
	}

	return &s
}
