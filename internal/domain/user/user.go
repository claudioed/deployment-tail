package user

import (
	"fmt"
	"time"
)

// User represents the user aggregate root
type User struct {
	id          UserID
	googleID    GoogleID
	email       Email
	name        UserName
	role        Role
	lastLoginAt *time.Time
	createdAt   time.Time
	updatedAt   time.Time
}

// NewUser creates a new User with validation
func NewUser(googleID GoogleID, email Email, name UserName, role Role) (*User, error) {
	now := time.Now().UTC()
	return &User{
		id:        NewUserID(),
		googleID:  googleID,
		email:     email,
		name:      name,
		role:      role,
		createdAt: now,
		updatedAt: now,
	}, nil
}

// Reconstitute creates a User from storage (bypasses validation)
func Reconstitute(
	id UserID,
	googleID GoogleID,
	email Email,
	name UserName,
	role Role,
	lastLoginAt *time.Time,
	createdAt time.Time,
	updatedAt time.Time,
) *User {
	return &User{
		id:          id,
		googleID:    googleID,
		email:       email,
		name:        name,
		role:        role,
		lastLoginAt: lastLoginAt,
		createdAt:   createdAt,
		updatedAt:   updatedAt,
	}
}

// UpdateProfile updates the user's profile information
func (u *User) UpdateProfile(name UserName, email Email) error {
	u.name = name
	u.email = email
	u.updatedAt = time.Now().UTC()
	return nil
}

// UpdateRole updates the user's role
func (u *User) UpdateRole(role Role) error {
	if u.role.Equals(role) {
		return nil // No change needed
	}
	u.role = role
	u.updatedAt = time.Now().UTC()
	return nil
}

// RecordLogin updates the last login timestamp
func (u *User) RecordLogin() {
	now := time.Now().UTC()
	u.lastLoginAt = &now
	u.updatedAt = now
}

// Getters
func (u *User) ID() UserID {
	return u.id
}

func (u *User) GoogleID() GoogleID {
	return u.googleID
}

func (u *User) Email() Email {
	return u.email
}

func (u *User) Name() UserName {
	return u.name
}

func (u *User) Role() Role {
	return u.role
}

func (u *User) LastLoginAt() *time.Time {
	return u.lastLoginAt
}

func (u *User) CreatedAt() time.Time {
	return u.createdAt
}

func (u *User) UpdatedAt() time.Time {
	return u.updatedAt
}

// HasRole checks if the user has a specific role
func (u *User) HasRole(role string) bool {
	return u.role.String() == role
}

// CanCreateSchedule checks if the user can create schedules
func (u *User) CanCreateSchedule() bool {
	return u.role.CanCreateSchedule()
}

// CanModifySchedule checks if the user can modify a specific schedule
func (u *User) CanModifySchedule(scheduleOwnerID UserID) bool {
	// Admin can modify any schedule
	if u.role.CanModifyAnySchedule() {
		return true
	}
	// Deployer can only modify their own schedules
	if u.role.IsDeployer() {
		return u.id.Equals(scheduleOwnerID)
	}
	return false
}

// CanDeleteSchedule checks if the user can delete a specific schedule
func (u *User) CanDeleteSchedule(scheduleOwnerID UserID) bool {
	return u.CanModifySchedule(scheduleOwnerID)
}

// CanManageUsers checks if the user can manage other users
func (u *User) CanManageUsers() bool {
	return u.role.CanManageUsers()
}

// Validate performs business rule validation
func (u *User) Validate() error {
	if u.id.String() == "" {
		return fmt.Errorf("user ID cannot be empty")
	}
	if u.googleID.String() == "" {
		return fmt.Errorf("google ID cannot be empty")
	}
	if u.email.String() == "" {
		return fmt.Errorf("email cannot be empty")
	}
	if u.name.String() == "" {
		return fmt.Errorf("name cannot be empty")
	}
	return nil
}
