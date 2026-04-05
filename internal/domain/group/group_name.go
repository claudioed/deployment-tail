package group

import "fmt"

// GroupName represents the name of a group
type GroupName struct {
	value string
}

// NewGroupName creates a new GroupName value object with validation
func NewGroupName(name string) (GroupName, error) {
	if name == "" {
		return GroupName{}, fmt.Errorf("group name cannot be empty")
	}

	if len(name) > 100 {
		return GroupName{}, fmt.Errorf("group name too long (max 100 characters)")
	}

	return GroupName{value: name}, nil
}

// String returns the string representation of the group name
func (g GroupName) String() string {
	return g.value
}

// Equals checks if two group names are equal
func (g GroupName) Equals(other GroupName) bool {
	return g.value == other.value
}
