package domain

// RSVPStatus is a member's answer to an event. A member who hasn't answered
// has no RSVP, shown as pending.
type RSVPStatus string

// RSVP answers.
const (
	// RSVPGoing means the member is coming.
	RSVPGoing RSVPStatus = "GOING"

	// RSVPMaybe means the member might come.
	RSVPMaybe RSVPStatus = "MAYBE"

	// RSVPNotGoing means the member isn't coming.
	RSVPNotGoing RSVPStatus = "NOT_GOING"
)

// RSVP is one member's answer to one event.
type RSVP struct {
	// EventID is the event answered.
	EventID string

	// UserID is the member who answered. They may since have left.
	UserID string

	// Status is their answer.
	Status RSVPStatus
}

// RSVPCounts tallies a squad's answers to one event. Pending counts members
// who haven't answered.
type RSVPCounts struct {
	// Going counts members who answered GOING.
	Going int

	// Maybe counts members who answered MAYBE.
	Maybe int

	// NotGoing counts members who answered NOT_GOING.
	NotGoing int

	// Pending counts members who haven't answered.
	Pending int
}

// IsValid reports whether s is a supported answer.
func (s RSVPStatus) IsValid() bool {
	return s == RSVPGoing || s == RSVPMaybe || s == RSVPNotGoing
}
