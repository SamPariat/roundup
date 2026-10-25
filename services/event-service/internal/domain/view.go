package domain

// GoingPreviewSize is how many going members an event card shows.
const GoingPreviewSize = 5

// EventSummary is an event as it appears in a squad's event list.
type EventSummary struct {
	Event

	// RSVPCounts tallies the current members' answers.
	RSVPCounts RSVPCounts

	// GoingPreview is up to GoingPreviewSize members who are going.
	GoingPreview []SquadMember

	// NominationCount is how many places were suggested.
	NominationCount int

	// MyRSVP is the caller's answer, "" if they haven't answered.
	MyRSVP RSVPStatus

	// HasVoted reports whether the caller voted.
	HasVoted bool
}

// MemberRSVP is a squad member with their answer.
type MemberRSVP struct {
	SquadMember

	// RSVP is their answer, "" if they haven't answered.
	RSVP RSVPStatus

	// IsMe marks the caller.
	IsMe bool
}

// NominationView is a suggested place with its votes.
type NominationView struct {
	Nomination

	// NominatedByName is the suggester's name, "" if they left the squad.
	NominatedByName string

	// VoteCount counts every vote for the place.
	VoteCount int

	// Voters are the current members who voted for it.
	Voters []SquadMember

	// VotedByMe reports whether the caller voted for it.
	VotedByMe bool

	// CanRemove reports whether the caller may withdraw it now.
	CanRemove bool
}

// Permissions says which actions the caller may take on the event right now,
// from the policy and the event's status, so the app shows only those buttons.
type Permissions struct {
	// CanEdit allows changing the name, type or times.
	CanEdit bool

	// CanRSVP allows answering going, maybe or not going.
	CanRSVP bool

	// CanNominate allows suggesting a place.
	CanNominate bool

	// CanVote allows voting for a suggested place.
	CanVote bool

	// CanStartVoting allows moving DRAFT to VOTING.
	CanStartVoting bool

	// CanCloseVoting allows moving VOTING to CONFIRMED before the deadline.
	CanCloseVoting bool

	// CanStart allows moving CONFIRMED to ACTIVE ("Start the night").
	CanStart bool

	// CanClose allows moving ACTIVE to CLOSED.
	CanClose bool
}

// EventDetail is everything the event screen shows.
type EventDetail struct {
	Event

	// CreatedByName is the creator's name, "" if they left the squad.
	CreatedByName string

	// Members are the squad's members with their answers.
	Members []MemberRSVP

	// Nominations are the suggested places, earliest first.
	Nominations []NominationView

	// VotedCount is how many current members voted.
	VotedCount int

	// TotalVotes counts every vote cast, for "won with 6 of 11".
	TotalVotes int

	// MyRSVP is the caller's answer, "" if they haven't answered.
	MyRSVP RSVPStatus

	// MyVote is the place the caller voted for, "" if none.
	MyVote string

	// Permissions are the caller's available actions.
	Permissions Permissions
}
