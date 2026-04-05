package user

import (
	"fmt"
	"strings"
)

// UserName represents a user's display name
type UserName struct {
	value string
}

// NewUserName creates a new UserName with validation
func NewUserName(name string) (UserName, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return UserName{}, fmt.Errorf("user name cannot be empty")
	}
	if len(name) > 255 {
		return UserName{}, fmt.Errorf("user name cannot exceed 255 characters")
	}
	return UserName{value: name}, nil
}

// String returns the string representation of the UserName
func (u UserName) String() string {
	return u.value
}

// Equals checks if two UserNames are equal
func (u UserName) Equals(other UserName) bool {
	return u.value == other.value
}
