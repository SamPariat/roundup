package dtos

import "github.com/SamPariat/roundup/services/user-service/internal/domain"

// UserToResponse maps a domain.User to a UserResponse.
func UserToResponse(u *domain.User) UserResponse {
	return UserResponse{
		ID:        u.ID,
		Phone:     u.Phone,
		Name:      u.Name,
		AvatarURL: u.AvatarURL,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}

// DeviceTokenToResponse maps a domain.DeviceToken to a DeviceTokenResponse.
func DeviceTokenToResponse(t *domain.DeviceToken) DeviceTokenResponse {
	return DeviceTokenResponse{
		ID:        t.ID,
		Token:     t.Token,
		Platform:  string(t.Platform),
		CreatedAt: t.CreatedAt,
	}
}

// UsersToLookupResponse maps lookup matches to a LookupUsersResponse.
func UsersToLookupResponse(users []*domain.User) LookupUsersResponse {
	out := make([]UserSummaryResponse, 0, len(users))
	for _, u := range users {
		out = append(out, UserSummaryResponse{ID: u.ID, Phone: u.Phone, Name: u.Name, AvatarURL: u.AvatarURL})
	}

	return LookupUsersResponse{Users: out}
}

// SquadToResponse maps a domain.Squad to a SquadResponse.
func SquadToResponse(s *domain.Squad) SquadResponse {
	return SquadResponse{
		ID:        s.ID,
		Name:      s.Name,
		AvatarURL: s.AvatarURL,
		CreatedBy: s.CreatedBy,
		CreatedAt: s.CreatedAt,
	}
}

// SquadsToResponse maps the caller's squads to a SquadsResponse.
func SquadsToResponse(squads []domain.SquadSummary) SquadsResponse {
	out := make([]SquadSummaryResponse, 0, len(squads))
	for _, s := range squads {
		out = append(out, SquadSummaryResponse{
			SquadResponse: SquadToResponse(&s.Squad),
			Role:          string(s.Role),
			MemberCount:   s.MemberCount,
		})
	}

	return SquadsResponse{Squads: out}
}

// SquadMemberToResponse maps a domain.SquadMember to a SquadMemberResponse.
func SquadMemberToResponse(m *domain.SquadMember) SquadMemberResponse {
	return SquadMemberResponse{
		UserID:    m.UserID,
		Name:      m.Name,
		AvatarURL: m.AvatarURL,
		Role:      string(m.Role),
		JoinedAt:  m.JoinedAt,
	}
}

// SquadDetailToResponse maps a domain.SquadDetail to a SquadDetailResponse.
func SquadDetailToResponse(d *domain.SquadDetail) SquadDetailResponse {
	members := make([]SquadMemberResponse, 0, len(d.Members))
	for i := range d.Members {
		members = append(members, SquadMemberToResponse(&d.Members[i]))
	}

	return SquadDetailResponse{SquadResponse: SquadToResponse(&d.Squad), Members: members}
}
