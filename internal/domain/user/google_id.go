package user

import (
	"fmt"
	"strings"
)

// GoogleID represents a Google account identifier
type GoogleID struct {
	value string
}

// NewGoogleID creates a new GoogleID with validation
func NewGoogleID(id string) (GoogleID, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return GoogleID{}, fmt.Errorf("google ID cannot be empty")
	}
	if len(id) > 255 {
		return GoogleID{}, fmt.Errorf("google ID cannot exceed 255 characters")
	}
	return GoogleID{value: id}, nil
}

// String returns the string representation of the GoogleID
func (g GoogleID) String() string {
	return g.value
}

// Equals checks if two GoogleIDs are equal
func (g GoogleID) Equals(other GoogleID) bool {
	return g.value == other.value
}
