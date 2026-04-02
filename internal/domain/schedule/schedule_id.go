package schedule

import (
	"fmt"

	"github.com/google/uuid"
)

// ScheduleID represents a unique schedule identifier
type ScheduleID struct {
	value string
}

// NewScheduleID creates a new schedule ID
func NewScheduleID() ScheduleID {
	return ScheduleID{value: uuid.New().String()}
}

// ParseScheduleID creates a schedule ID from a string
func ParseScheduleID(id string) (ScheduleID, error) {
	if _, err := uuid.Parse(id); err != nil {
		return ScheduleID{}, fmt.Errorf("invalid schedule ID: %w", err)
	}
	return ScheduleID{value: id}, nil
}

// String returns the string representation
func (s ScheduleID) String() string {
	return s.value
}

// Equals checks if two schedule IDs are equal
func (s ScheduleID) Equals(other ScheduleID) bool {
	return s.value == other.value
}
