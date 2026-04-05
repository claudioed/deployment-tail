package user

import "fmt"

// Role represents a user's role with authorization level
type Role struct {
	value string
}

const (
	RoleViewer   = "viewer"
	RoleDeployer = "deployer"
	RoleAdmin    = "admin"
)

// NewRole creates a new Role with enum validation
func NewRole(role string) (Role, error) {
	switch role {
	case RoleViewer, RoleDeployer, RoleAdmin:
		return Role{value: role}, nil
	default:
		return Role{}, fmt.Errorf("invalid role: %s (must be viewer, deployer, or admin)", role)
	}
}

// String returns the string representation of the Role
func (r Role) String() string {
	return r.value
}

// Equals checks if two Roles are equal
func (r Role) Equals(other Role) bool {
	return r.value == other.value
}

// IsViewer checks if the role is viewer
func (r Role) IsViewer() bool {
	return r.value == RoleViewer
}

// IsDeployer checks if the role is deployer
func (r Role) IsDeployer() bool {
	return r.value == RoleDeployer
}

// IsAdmin checks if the role is admin
func (r Role) IsAdmin() bool {
	return r.value == RoleAdmin
}

// CanCreateSchedule checks if the role can create schedules
func (r Role) CanCreateSchedule() bool {
	return r.IsDeployer() || r.IsAdmin()
}

// CanModifyAnySchedule checks if the role can modify any schedule
func (r Role) CanModifyAnySchedule() bool {
	return r.IsAdmin()
}

// CanManageUsers checks if the role can manage users
func (r Role) CanManageUsers() bool {
	return r.IsAdmin()
}
