package domain

// RSVPStatus is a member's answer to an event. A member who hasn't answered
// has no RSVP, shown as pending.
type RSVPStatus string

// RSVP answers.
const (
	RSVPGoing    RSVPStatus = "GOING"
	RSVPMaybe    RSVPStatus = "MAYBE"
	RSVPNotGoing RSVPStatus = "NOT_GOING"
)

// RSVP is one member's answer to one event.
type RSVP struct {
	EventID string
	UserID  string
	Status  RSVPStatus
}

// RSVPCounts tallies a squad's answers to one event. Pending counts members
// who haven't answered.
type RSVPCounts struct {
	Going    int
	Maybe    int
	NotGoing int
	Pending  int
}

// IsValid reports whether s is a supported answer.
func (s RSVPStatus) IsValid() bool {
	return s == RSVPGoing || s == RSVPMaybe || s == RSVPNotGoing
}
