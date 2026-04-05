package authorization

import (
	"testing"
	"time"

	"github.com/claudioed/deployment-tail/internal/domain/schedule"
	"github.com/claudioed/deployment-tail/internal/domain/user"
)

// Test helper functions
func createTestUser(role string) *user.User {
	googleID, _ := user.NewGoogleID("test-google-id-" + role)
	email, _ := user.NewEmail(role + "@example.com")
	name, _ := user.NewUserName("Test " + role)
	r, _ := user.NewRole(role)
	u, _ := user.NewUser(googleID, email, name, r)
	return u
}

func createTestSchedule(createdBy user.UserID) *schedule.Schedule {
	scheduledAt, _ := schedule.NewScheduledTime(time.Now().Add(24 * time.Hour))
	serviceName, _ := schedule.NewServiceName("test-service")
	env, _ := schedule.NewEnvironment("production")
	desc := schedule.NewDescription("Test schedule")
	owner, _ := schedule.NewOwner("test-owner")
	rollback, _ := schedule.NewRollbackPlan("rollback")
	sch, _ := schedule.NewSchedule(scheduledAt, serviceName, []schedule.Environment{env}, desc, []schedule.Owner{owner}, rollback, createdBy)
	return sch
}

func TestCanCreateSchedule(t *testing.T) {
	tests := []struct {
		name     string
		role     string
		expected bool
	}{
		{"viewer cannot create", user.RoleViewer, false},
		{"deployer can create", user.RoleDeployer, true},
		{"admin can create", user.RoleAdmin, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := createTestUser(tt.role)
			result := CanCreateSchedule(u)

			if result != tt.expected {
				t.Errorf("CanCreateSchedule() = %v, expected %v for role %s", result, tt.expected, tt.role)
			}
		})
	}
}

func TestCanUpdateSchedule(t *testing.T) {
	deployer := createTestUser(user.RoleDeployer)
	otherDeployer := createTestUser(user.RoleDeployer)
	admin := createTestUser(user.RoleAdmin)
	viewer := createTestUser(user.RoleViewer)

	deployerSchedule := createTestSchedule(deployer.ID())

	tests := []struct {
		name     string
		user     *user.User
		schedule *schedule.Schedule
		expected bool
	}{
		{"deployer can update own schedule", deployer, deployerSchedule, true},
		{"deployer cannot update other's schedule", otherDeployer, deployerSchedule, false},
		{"admin can update any schedule", admin, deployerSchedule, true},
		{"viewer cannot update schedule", viewer, deployerSchedule, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CanUpdateSchedule(tt.user, tt.schedule)

			if result != tt.expected {
				t.Errorf("CanUpdateSchedule() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestCanDeleteSchedule(t *testing.T) {
	deployer := createTestUser(user.RoleDeployer)
	otherDeployer := createTestUser(user.RoleDeployer)
	admin := createTestUser(user.RoleAdmin)
	viewer := createTestUser(user.RoleViewer)

	deployerSchedule := createTestSchedule(deployer.ID())

	tests := []struct {
		name     string
		user     *user.User
		schedule *schedule.Schedule
		expected bool
	}{
		{"deployer can delete own schedule", deployer, deployerSchedule, true},
		{"deployer cannot delete other's schedule", otherDeployer, deployerSchedule, false},
		{"admin can delete any schedule", admin, deployerSchedule, true},
		{"viewer cannot delete schedule", viewer, deployerSchedule, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CanDeleteSchedule(tt.user, tt.schedule)

			if result != tt.expected {
				t.Errorf("CanDeleteSchedule() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestCanListUsers(t *testing.T) {
	tests := []struct {
		name     string
		role     string
		expected bool
	}{
		{"viewer cannot list users", user.RoleViewer, false},
		{"deployer cannot list users", user.RoleDeployer, false},
		{"admin can list users", user.RoleAdmin, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := createTestUser(tt.role)
			result := CanListUsers(u)

			if result != tt.expected {
				t.Errorf("CanListUsers() = %v, expected %v for role %s", result, tt.expected, tt.role)
			}
		})
	}
}

func TestCanAssignRole(t *testing.T) {
	tests := []struct {
		name     string
		role     string
		expected bool
	}{
		{"viewer cannot assign roles", user.RoleViewer, false},
		{"deployer cannot assign roles", user.RoleDeployer, false},
		{"admin can assign roles", user.RoleAdmin, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := createTestUser(tt.role)
			result := CanAssignRole(u)

			if result != tt.expected {
				t.Errorf("CanAssignRole() = %v, expected %v for role %s", result, tt.expected, tt.role)
			}
		})
	}
}

func TestCanViewUserProfile(t *testing.T) {
	viewer := createTestUser(user.RoleViewer)
	deployer := createTestUser(user.RoleDeployer)
	admin := createTestUser(user.RoleAdmin)
	otherUser := createTestUser(user.RoleDeployer)

	tests := []struct {
		name           string
		requestingUser *user.User
		targetUserID   user.UserID
		expected       bool
	}{
		{"user can view own profile (viewer)", viewer, viewer.ID(), true},
		{"user can view own profile (deployer)", deployer, deployer.ID(), true},
		{"user can view own profile (admin)", admin, admin.ID(), true},
		{"viewer cannot view other's profile", viewer, deployer.ID(), false},
		{"deployer cannot view other's profile", deployer, otherUser.ID(), false},
		{"admin can view any profile", admin, viewer.ID(), true},
		{"admin can view any profile (2)", admin, deployer.ID(), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CanViewUserProfile(tt.requestingUser, tt.targetUserID)

			if result != tt.expected {
				t.Errorf("CanViewUserProfile() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestCanApproveSchedule(t *testing.T) {
	tests := []struct {
		name     string
		role     string
		expected bool
	}{
		{"viewer cannot approve", user.RoleViewer, false},
		{"deployer cannot approve", user.RoleDeployer, false},
		{"admin can approve", user.RoleAdmin, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := createTestUser(tt.role)
			result := CanApproveSchedule(u)

			if result != tt.expected {
				t.Errorf("CanApproveSchedule() = %v, expected %v for role %s", result, tt.expected, tt.role)
			}
		})
	}
}

func TestCanDenySchedule(t *testing.T) {
	tests := []struct {
		name     string
		role     string
		expected bool
	}{
		{"viewer cannot deny", user.RoleViewer, false},
		{"deployer cannot deny", user.RoleDeployer, false},
		{"admin can deny", user.RoleAdmin, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := createTestUser(tt.role)
			result := CanDenySchedule(u)

			if result != tt.expected {
				t.Errorf("CanDenySchedule() = %v, expected %v for role %s", result, tt.expected, tt.role)
			}
		})
	}
}

func TestCanManageGroups(t *testing.T) {
	tests := []struct {
		name     string
		role     string
		expected bool
	}{
		{"viewer cannot manage groups", user.RoleViewer, false},
		{"deployer can manage groups", user.RoleDeployer, true},
		{"admin can manage groups", user.RoleAdmin, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := createTestUser(tt.role)
			result := CanManageGroups(u)

			if result != tt.expected {
				t.Errorf("CanManageGroups() = %v, expected %v for role %s", result, tt.expected, tt.role)
			}
		})
	}
}

// Edge case tests
func TestCanViewUserProfile_SameUserID(t *testing.T) {
	user1 := createTestUser(user.RoleViewer)

	// User should always be able to view their own profile
	if !CanViewUserProfile(user1, user1.ID()) {
		t.Error("User should be able to view their own profile")
	}
}

func TestCanUpdateSchedule_AdminOverridesOwnership(t *testing.T) {
	deployer := createTestUser(user.RoleDeployer)
	admin := createTestUser(user.RoleAdmin)

	deployerSchedule := createTestSchedule(deployer.ID())

	// Admin should be able to update even though they didn't create it
	if !CanUpdateSchedule(admin, deployerSchedule) {
		t.Error("Admin should be able to update any schedule")
	}

	// Deployer should only be able to update their own
	if !CanUpdateSchedule(deployer, deployerSchedule) {
		t.Error("Deployer should be able to update their own schedule")
	}
}

func TestCanDeleteSchedule_AdminOverridesOwnership(t *testing.T) {
	deployer := createTestUser(user.RoleDeployer)
	admin := createTestUser(user.RoleAdmin)

	deployerSchedule := createTestSchedule(deployer.ID())

	// Admin should be able to delete even though they didn't create it
	if !CanDeleteSchedule(admin, deployerSchedule) {
		t.Error("Admin should be able to delete any schedule")
	}

	// Deployer should only be able to delete their own
	if !CanDeleteSchedule(deployer, deployerSchedule) {
		t.Error("Deployer should be able to delete their own schedule")
	}
}

func TestAuthorizationConsistency(t *testing.T) {
	// Test that authorization policies are consistent across similar operations
	deployer := createTestUser(user.RoleDeployer)

	// If a user can create schedules, they should also be able to manage groups
	if CanCreateSchedule(deployer) != CanManageGroups(deployer) {
		t.Error("Schedule creation and group management permissions should be consistent")
	}

	// If a user can assign roles, they should also be able to list users
	admin := createTestUser(user.RoleAdmin)
	if CanAssignRole(admin) != CanListUsers(admin) {
		t.Error("Role assignment and user listing permissions should be consistent")
	}

	// If a user can approve schedules, they should also be able to deny them
	if CanApproveSchedule(admin) != CanDenySchedule(admin) {
		t.Error("Approve and deny permissions should be consistent")
	}
}
