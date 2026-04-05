package authorization

import (
	"github.com/claudioed/deployment-tail/internal/domain/schedule"
	"github.com/claudioed/deployment-tail/internal/domain/user"
)

// Policy provides centralized authorization logic for the application layer.
// All permission checks should go through these functions to ensure consistent
// authorization behavior across use cases.

// CanCreateSchedule checks if a user has permission to create schedules.
// Only deployers and admins can create schedules.
func CanCreateSchedule(u *user.User) bool {
	return u.CanCreateSchedule()
}

// CanUpdateSchedule checks if a user has permission to update a specific schedule.
// Deployers can only update their own schedules, admins can update any schedule.
func CanUpdateSchedule(u *user.User, sch *schedule.Schedule) bool {
	return u.CanModifySchedule(sch.CreatedBy())
}

// CanDeleteSchedule checks if a user has permission to delete a specific schedule.
// Deployers can only delete their own schedules, admins can delete any schedule.
func CanDeleteSchedule(u *user.User, sch *schedule.Schedule) bool {
	return u.CanModifySchedule(sch.CreatedBy())
}

// CanListUsers checks if a user has permission to list all users.
// Only admins can list users.
func CanListUsers(u *user.User) bool {
	return u.CanManageUsers()
}

// CanAssignRole checks if a user has permission to assign roles to other users.
// Only admins can assign roles.
func CanAssignRole(u *user.User) bool {
	return u.CanManageUsers()
}

// CanViewUserProfile checks if a user has permission to view another user's profile.
// Users can view their own profile, admins can view any profile.
func CanViewUserProfile(requestingUser *user.User, targetUserID user.UserID) bool {
	// Users can always view their own profile
	if requestingUser.ID().Equals(targetUserID) {
		return true
	}

	// Admins can view any profile
	return requestingUser.CanManageUsers()
}

// CanApproveSchedule checks if a user has permission to approve schedules.
// Only admins can approve schedules.
func CanApproveSchedule(u *user.User) bool {
	return u.Role().IsAdmin()
}

// CanDenySchedule checks if a user has permission to deny schedules.
// Only admins can deny schedules.
func CanDenySchedule(u *user.User) bool {
	return u.Role().IsAdmin()
}

// CanManageGroups checks if a user has permission to manage schedule groups.
// Deployers and admins can manage groups.
func CanManageGroups(u *user.User) bool {
	return u.CanCreateSchedule()
}
