package user

import "fmt"

// ErrUserNotFound is returned when a user cannot be found
type ErrUserNotFound struct {
	ID         string
	GoogleID   string
	Email      string
	SearchType string
}

func (e ErrUserNotFound) Error() string {
	switch e.SearchType {
	case "id":
		return fmt.Sprintf("user not found with ID: %s", e.ID)
	case "google_id":
		return fmt.Sprintf("user not found with Google ID: %s", e.GoogleID)
	case "email":
		return fmt.Sprintf("user not found with email: %s", e.Email)
	default:
		return "user not found"
	}
}

// ErrUserAlreadyExists is returned when attempting to create a user that already exists
type ErrUserAlreadyExists struct {
	GoogleID string
	Email    string
}

func (e ErrUserAlreadyExists) Error() string {
	if e.GoogleID != "" {
		return fmt.Sprintf("user already exists with Google ID: %s", e.GoogleID)
	}
	if e.Email != "" {
		return fmt.Sprintf("user already exists with email: %s", e.Email)
	}
	return "user already exists"
}

// ErrInvalidUserData is returned when user data validation fails
type ErrInvalidUserData struct {
	Field   string
	Message string
}

func (e ErrInvalidUserData) Error() string {
	return fmt.Sprintf("invalid user data - %s: %s", e.Field, e.Message)
}

// ErrUnauthorized is returned when a user lacks permission for an operation
type ErrUnauthorized struct {
	UserID    string
	Operation string
	Reason    string
}

func (e ErrUnauthorized) Error() string {
	if e.Reason != "" {
		return fmt.Sprintf("user %s unauthorized for %s: %s", e.UserID, e.Operation, e.Reason)
	}
	return fmt.Sprintf("user %s unauthorized for %s", e.UserID, e.Operation)
}
