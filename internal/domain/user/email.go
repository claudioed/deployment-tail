package user

import (
	"fmt"
	"regexp"
	"strings"
)

// Email represents a validated email address
type Email struct {
	value string
}

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

// NewEmail creates a new Email with format validation
func NewEmail(email string) (Email, error) {
	email = strings.TrimSpace(email)
	if email == "" {
		return Email{}, fmt.Errorf("email cannot be empty")
	}
	if len(email) > 255 {
		return Email{}, fmt.Errorf("email cannot exceed 255 characters")
	}
	if !emailRegex.MatchString(email) {
		return Email{}, fmt.Errorf("invalid email format")
	}
	return Email{value: strings.ToLower(email)}, nil
}

// String returns the string representation of the Email
func (e Email) String() string {
	return e.value
}

// Equals checks if two Emails are equal
func (e Email) Equals(other Email) bool {
	return e.value == other.value
}
