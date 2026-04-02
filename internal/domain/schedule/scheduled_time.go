package schedule

import (
	"fmt"
	"time"
)

// ScheduledTime represents when a deployment is scheduled
type ScheduledTime struct {
	value time.Time
}

// NewScheduledTime creates a new scheduled time
func NewScheduledTime(t time.Time) (ScheduledTime, error) {
	if t.IsZero() {
		return ScheduledTime{}, fmt.Errorf("scheduled time cannot be zero")
	}
	// Ensure UTC
	return ScheduledTime{value: t.UTC()}, nil
}

// Value returns the underlying time value
func (s ScheduledTime) Value() time.Time {
	return s.value
}

// String returns ISO 8601 formatted string
func (s ScheduledTime) String() string {
	return s.value.Format(time.RFC3339)
}

// Before checks if this time is before another
func (s ScheduledTime) Before(other ScheduledTime) bool {
	return s.value.Before(other.value)
}

// After checks if this time is after another
func (s ScheduledTime) After(other ScheduledTime) bool {
	return s.value.After(other.value)
}
