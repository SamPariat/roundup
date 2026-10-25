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
	Name           *string    `json:"name"`
	Type           *string    `json:"type"`
	StartsAt       *time.Time `json:"startsAt"`
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
	URL    string `json:"url"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

// PlaceResponse is a suggested or winning place's display fields.
type PlaceResponse struct {
	PlaceID    string         `json:"placeID"`
	Name       string         `json:"name"`
	Address    string         `json:"address"`
	Latitude   float64        `json:"latitude"`
	Longitude  float64        `json:"longitude"`
	Rating     float32        `json:"rating"`
	PriceLevel int            `json:"priceLevel"`
	Photo      *PhotoResponse `json:"photo"`
}

// WinnerResponse is the confirmed place.
type WinnerResponse struct {
	PlaceResponse

	// Votes is how many votes it won with. Only on the event detail.
	Votes *int `json:"votes,omitempty"`
}

// MemberResponse is a squad member as event screens show them.
type MemberResponse struct {
	UserID    string `json:"userID"`
	Name      string `json:"name"`
	AvatarURL string `json:"avatarURL"`
}

// EventResponse holds the fields every event representation shares.
type EventResponse struct {
	ID             string          `json:"id"`
	SquadID        string          `json:"squadID"`
	Name           string          `json:"name"`
	Type           string          `json:"type"`
	Status         string          `json:"status"`
	CreatedBy      string          `json:"createdBy"`
	StartsAt       time.Time       `json:"startsAt"`
	VotingDeadline time.Time       `json:"votingDeadline"`
	Winner         *WinnerResponse `json:"winner"`
	ClosedAt       *time.Time      `json:"closedAt"`
	CreatedAt      time.Time       `json:"createdAt"`
}

// RSVPCountsResponse tallies a squad's answers.
type RSVPCountsResponse struct {
	Going    int `json:"going"`
	Maybe    int `json:"maybe"`
	NotGoing int `json:"notGoing"`
	Pending  int `json:"pending"`
}

// EventSummaryResponse is one entry in a squad's event list.
type EventSummaryResponse struct {
	EventResponse
	RSVPCounts      RSVPCountsResponse `json:"rsvpCounts"`
	GoingPreview    []MemberResponse   `json:"goingPreview"`
	NominationCount int                `json:"nominationCount"`

	// MyRSVP is the caller's answer; null if they haven't answered.
	MyRSVP   *string `json:"myRsvp"`
	HasVoted bool    `json:"hasVoted"`
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
	IsMe bool    `json:"isMe"`
}

// NominationResponse is a suggested place with its votes.
type NominationResponse struct {
	PlaceResponse
	NominatedBy string `json:"nominatedBy"`

	// NominatedByName is null if the suggester left the squad.
	NominatedByName *string          `json:"nominatedByName"`
	NominatedAt     time.Time        `json:"nominatedAt"`
	VoteCount       int              `json:"voteCount"`
	Voters          []MemberResponse `json:"voters"`
	VotedByMe       bool             `json:"votedByMe"`
	CanRemove       bool             `json:"canRemove"`
}

// PermissionsResponse lists the actions the caller may take right now.
type PermissionsResponse struct {
	CanEdit        bool `json:"canEdit"`
	CanRSVP        bool `json:"canRsvp"`
	CanNominate    bool `json:"canNominate"`
	CanVote        bool `json:"canVote"`
	CanStartVoting bool `json:"canStartVoting"`
	CanCloseVoting bool `json:"canCloseVoting"`
	CanStart       bool `json:"canStart"`
	CanClose       bool `json:"canClose"`
}

// EventDetailResponse is everything the event screen shows.
type EventDetailResponse struct {
	EventResponse

	// CreatedByName is null if the creator left the squad.
	CreatedByName *string              `json:"createdByName"`
	MemberCount   int                  `json:"memberCount"`
	VotedCount    int                  `json:"votedCount"`
	TotalVotes    int                  `json:"totalVotes"`
	MyRSVP        *string              `json:"myRsvp"`
	MyVote        *string              `json:"myVote"`
	Members       []MemberRSVPResponse `json:"members"`
	Nominations   []NominationResponse `json:"nominations"`
	Permissions   PermissionsResponse  `json:"permissions"`
}

// DueResponse is the body of the internal confirm-due and close-due routes.
type DueResponse struct {
	// Count is how many events this call moved on.
	Count int `json:"count"`
}
