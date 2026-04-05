package group

import (
	"fmt"

	"github.com/google/uuid"
)

// GroupID represents a unique group identifier
type GroupID struct {
	value string
}

// NewGroupID creates a new group ID
func NewGroupID() GroupID {
	return GroupID{value: uuid.New().String()}
}

// ParseGroupID creates a group ID from a string
func ParseGroupID(id string) (GroupID, error) {
	if _, err := uuid.Parse(id); err != nil {
		return GroupID{}, fmt.Errorf("invalid group ID: %w", err)
	}
	return GroupID{value: id}, nil
}

// String returns the string representation
func (g GroupID) String() string {
	return g.value
}

// Equals checks if two group IDs are equal
func (g GroupID) Equals(other GroupID) bool {
	return g.value == other.value
}
