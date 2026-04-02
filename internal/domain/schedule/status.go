package schedule

import "fmt"

// Status represents the approval status of a schedule
type Status int

const (
	StatusCreated Status = iota
	StatusApproved
	StatusDenied
)

// String returns the string representation of the status
func (s Status) String() string {
	switch s {
	case StatusCreated:
		return "created"
	case StatusApproved:
		return "approved"
	case StatusDenied:
		return "denied"
	default:
		return "unknown"
	}
}

// ParseStatus creates a Status from a string
func ParseStatus(status string) (Status, error) {
	switch status {
	case "created":
		return StatusCreated, nil
	case "approved":
		return StatusApproved, nil
	case "denied":
		return StatusDenied, nil
	default:
		return 0, fmt.Errorf("invalid status: must be created, approved, or denied")
	}
}

// CanTransitionTo checks if the status can transition to another status
func (s Status) CanTransitionTo(target Status) error {
	// Valid transitions: created → approved, created → denied
	// Invalid: approved → denied, denied → approved, approved → created, denied → created

	if s == target {
		return fmt.Errorf("schedule is already %s", s.String())
	}

	switch s {
	case StatusCreated:
		// Can transition to approved or denied
		if target == StatusApproved || target == StatusDenied {
			return nil
		}
		return fmt.Errorf("invalid status transition from %s to %s", s.String(), target.String())
	case StatusApproved:
		return fmt.Errorf("cannot change status of approved schedule")
	case StatusDenied:
		return fmt.Errorf("cannot change status of denied schedule")
	default:
		return fmt.Errorf("unknown status: %s", s.String())
	}
}
