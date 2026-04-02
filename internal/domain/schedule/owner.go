package schedule

import (
	"fmt"
	"regexp"
)

// Owner represents the owner of a schedule (immutable after creation)
type Owner struct {
	value string
}

// ownerPattern allows alphanumeric characters, dots, hyphens, underscores, and @ symbol
var ownerPattern = regexp.MustCompile(`^[a-zA-Z0-9._@-]+$`)

// NewOwner creates a new Owner value object with validation
func NewOwner(owner string) (Owner, error) {
	if owner == "" {
		return Owner{}, fmt.Errorf("owner cannot be empty")
	}

	if len(owner) > 255 {
		return Owner{}, fmt.Errorf("owner too long (max 255 characters)")
	}

	if !ownerPattern.MatchString(owner) {
		return Owner{}, fmt.Errorf("owner must contain only alphanumeric characters, dots, hyphens, underscores, or @ symbol")
	}

	return Owner{value: owner}, nil
}

// String returns the string representation of the owner
func (o Owner) String() string {
	return o.value
}

// Equals checks if two owners are equal
func (o Owner) Equals(other Owner) bool {
	return o.value == other.value
}
