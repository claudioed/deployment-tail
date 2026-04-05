package user

import (
	"fmt"

	"github.com/google/uuid"
)

// UserID represents a unique identifier for a user
type UserID struct {
	value uuid.UUID
}

// NewUserID creates a new UserID with a generated UUID
func NewUserID() UserID {
	return UserID{value: uuid.New()}
}

// ParseUserID creates a UserID from a string UUID
func ParseUserID(id string) (UserID, error) {
	parsed, err := uuid.Parse(id)
	if err != nil {
		return UserID{}, fmt.Errorf("invalid user ID format: %w", err)
	}
	return UserID{value: parsed}, nil
}

// String returns the string representation of the UserID
func (u UserID) String() string {
	return u.value.String()
}

// UUID returns the underlying uuid.UUID
func (u UserID) UUID() uuid.UUID {
	return u.value
}

// Equals checks if two UserIDs are equal
func (u UserID) Equals(other UserID) bool {
	return u.value == other.value
}
