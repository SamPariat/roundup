package dtos

import "time"

// CreateEventRequest carries the JSON body for creating an event.
type CreateEventRequest struct {
	// Name is the display name. Trimmed; must not be blank.
	Name string `json:"name"`

	// Type is DRINKS, DINING or BOTH.
	Type string `json:"type"`

	// StartsAt is when the night starts, as an RFC 3339 timestamp.
	StartsAt time.Time `json:"startsAt"`

	// VotingDeadline is when voting closes. Optional; see domain.NewSchedule.
	VotingDeadline *time.Time `json:"votingDeadline"`
}

// UpdateEventRequest carries the JSON body for editing an event. Omitted
// fields are left unchanged.
type UpdateEventRequest struct {
	// Name is the new display name. Trimmed; must not be blank.
	Name *string `json:"name"`

	// Type is the new type: DRINKS, DINING or BOTH.
	Type *string `json:"type"`

	// StartsAt is the new start. Without VotingDeadline, the deadline moves to it.
	StartsAt *time.Time `json:"startsAt"`

	// VotingDeadline is the new deadline; in the future and no later than the start.
	VotingDeadline *time.Time `json:"votingDeadline"`
}

// RSVPRequest carries the caller's answer.
type RSVPRequest struct {
	// Status is GOING, MAYBE or NOT_GOING.
	Status string `json:"status"`
}

// PlaceRequest names a place, for suggesting it or voting for it.
type PlaceRequest struct {
	// PlaceID is the Google Place ID.
	PlaceID string `json:"placeID"`
}

// PhotoResponse is a venue photo the client can load directly.
type PhotoResponse struct {
	// URL is the photo's path on this API, e.g. "/api/v1/venues/photos/…".
	URL string `json:"url"`

	// Width is the original width in pixels.
	Width int `json:"width"`

	// Height is the original height in pixels.
	Height int `json:"height"`
}

// PlaceResponse is a suggested or winning place's display fields.
type PlaceResponse struct {
	// PlaceID is the Google Place ID.
	PlaceID string `json:"placeID"`

	// Name is the place's display name when it was suggested.
	Name string `json:"name"`

	// Address is the formatted street address.
	Address string `json:"address"`

	// Latitude is the WGS-84 latitude.
	Latitude float64 `json:"latitude"`

	// Longitude is the WGS-84 longitude.
	Longitude float64 `json:"longitude"`

	// Rating is the average rating, 1–5; 0 when unrated.
	Rating float32 `json:"rating"`

	// PriceLevel is 0–4 (0 = free or unknown, 4 = very expensive).
	PriceLevel int `json:"priceLevel"`

	// Photo is the place's first photo; null if it had none.
	Photo *PhotoResponse `json:"photo"`
}

// WinnerResponse is the confirmed place.
type WinnerResponse struct {
	PlaceResponse

	// Votes is how many votes it won with. Only on the event detail.
	Votes *int `json:"votes,omitempty"`
}

// MemberResponse is a squad member as event screens show them.
type MemberResponse struct {
	// UserID is the member's user ID.
	UserID string `json:"userID"`

	// Name is the member's display name.
	Name string `json:"name"`

	// AvatarURL is the member's avatar; "" when they have none.
	AvatarURL string `json:"avatarURL"`
}

// EventResponse holds the fields every event representation shares.
type EventResponse struct {
	// ID is the event's UUID.
	ID string `json:"id"`

	// SquadID is the squad the event belongs to.
	SquadID string `json:"squadID"`

	// Name is the display name.
	Name string `json:"name"`

	// Type is DRINKS, DINING or BOTH.
	Type string `json:"type"`

	// Status is DRAFT, VOTING, CONFIRMED, ACTIVE or CLOSED.
	Status string `json:"status"`

	// CreatedBy is the creator's user ID.
	CreatedBy string `json:"createdBy"`

	// StartsAt is when the night starts.
	StartsAt time.Time `json:"startsAt"`

	// VotingDeadline is when voting closes on its own.
	VotingDeadline time.Time `json:"votingDeadline"`

	// Winner is the confirmed place; null until voting closes.
	Winner *WinnerResponse `json:"winner"`

	// ClosedAt is when the night closed; null until then.
	ClosedAt *time.Time `json:"closedAt"`

	// CreatedAt is when the event was created.
	CreatedAt time.Time `json:"createdAt"`
}

// RSVPCountsResponse tallies a squad's answers.
type RSVPCountsResponse struct {
	// Going counts current members who answered GOING.
	Going int `json:"going"`

	// Maybe counts current members who answered MAYBE.
	Maybe int `json:"maybe"`

	// NotGoing counts current members who answered NOT_GOING.
	NotGoing int `json:"notGoing"`

	// Pending counts current members who haven't answered.
	Pending int `json:"pending"`
}

// EventSummaryResponse is one entry in a squad's event list.
type EventSummaryResponse struct {
	EventResponse

	// RSVPCounts tallies the current members' answers.
	RSVPCounts RSVPCountsResponse `json:"rsvpCounts"`

	// GoingPreview is up to 5 members who are going, for an avatar stack.
	GoingPreview []MemberResponse `json:"goingPreview"`

	// NominationCount is how many places have been suggested.
	NominationCount int `json:"nominationCount"`

	// MyRSVP is the caller's answer; null if they haven't answered.
	MyRSVP *string `json:"myRsvp"`

	// HasVoted is whether the caller has voted.
	HasVoted bool `json:"hasVoted"`
}

// SquadEventsResponse is the body of the squad events list.
type SquadEventsResponse struct {
	// Events holds open events soonest first, then closed ones most recent
	// first. Empty, never null.
	Events []EventSummaryResponse `json:"events"`

	// NextCursor fetches the next page as ?cursor=; null on the last page.
	NextCursor *string `json:"nextCursor"`

	// OpenCount counts all the squad's events that aren't closed.
	OpenCount int `json:"openCount"`
}

// MemberRSVPResponse is a squad member with their answer.
type MemberRSVPResponse struct {
	MemberResponse

	// RSVP is their answer; null if they haven't answered.
	RSVP *string `json:"rsvp"`

	// IsMe is whether this member is the caller.
	IsMe bool `json:"isMe"`
}

// NominationResponse is a suggested place with its votes.
type NominationResponse struct {
	PlaceResponse

	// NominatedBy is the suggester's user ID.
	NominatedBy string `json:"nominatedBy"`

	// NominatedByName is null if the suggester left the squad.
	NominatedByName *string `json:"nominatedByName"`

	// NominatedAt is when the place was suggested; ties go to the earliest.
	NominatedAt time.Time `json:"nominatedAt"`

	// VoteCount counts every vote cast for it, including by members who left.
	VoteCount int `json:"voteCount"`

	// Voters are the current members who voted for it.
	Voters []MemberResponse `json:"voters"`

	// VotedByMe is whether the caller voted for it.
	VotedByMe bool `json:"votedByMe"`

	// CanRemove is whether the caller may withdraw it now.
	CanRemove bool `json:"canRemove"`
}

// PermissionsResponse lists the actions the caller may take right now.
type PermissionsResponse struct {
	// CanEdit allows changing the name, type or times.
	CanEdit bool `json:"canEdit"`

	// CanRSVP allows answering going, maybe or not going.
	CanRSVP bool `json:"canRsvp"`

	// CanNominate allows suggesting a place.
	CanNominate bool `json:"canNominate"`

	// CanVote allows voting for a suggested place.
	CanVote bool `json:"canVote"`

	// CanStartVoting allows moving DRAFT to VOTING.
	CanStartVoting bool `json:"canStartVoting"`

	// CanCloseVoting allows moving VOTING to CONFIRMED before the deadline.
	CanCloseVoting bool `json:"canCloseVoting"`

	// CanStart allows moving CONFIRMED to ACTIVE ("Start the night").
	CanStart bool `json:"canStart"`

	// CanClose allows moving ACTIVE to CLOSED.
	CanClose bool `json:"canClose"`
}

// EventDetailResponse is everything the event screen shows.
type EventDetailResponse struct {
	EventResponse

	// CreatedByName is null if the creator left the squad.
	CreatedByName *string `json:"createdByName"`

	// MemberCount is how many members the squad has now.
	MemberCount int `json:"memberCount"`

	// VotedCount is how many current members have voted.
	VotedCount int `json:"votedCount"`

	// TotalVotes counts every vote cast, including by members who left.
	TotalVotes int `json:"totalVotes"`

	// MyRSVP is the caller's answer; null if they haven't answered.
	MyRSVP *string `json:"myRsvp"`

	// MyVote is the place ID the caller voted for; null if they haven't voted.
	MyVote *string `json:"myVote"`

	// Members are the squad's current members with their answers.
	Members []MemberRSVPResponse `json:"members"`

	// Nominations are the suggested places, earliest first.
	Nominations []NominationResponse `json:"nominations"`

	// Permissions are the actions the caller may take right now.
	Permissions PermissionsResponse `json:"permissions"`
}

// DueResponse is the body of the internal confirm-due and close-due routes.
type DueResponse struct {
	// Count is how many events this call moved on.
	Count int `json:"count"`
}
