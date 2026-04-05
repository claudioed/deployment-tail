package application

import (
	"context"
	"testing"

	"github.com/claudioed/deployment-tail/internal/application/ports/input"
	"github.com/claudioed/deployment-tail/internal/domain/user"
)

func TestRegisterOrUpdateUser_NewUser(t *testing.T) {
	repo := NewMockUserRepository()
	service := NewUserService(repo, nil, nil, nil)

	ctx := context.Background()
	u, err := service.RegisterOrUpdateUser(ctx, "108123456789", "new@example.com", "New User")

	if err != nil {
		t.Fatalf("RegisterOrUpdateUser() error = %v", err)
	}

	if u.Email().String() != "new@example.com" {
		t.Errorf("Expected email 'new@example.com', got %s", u.Email().String())
	}

	// Verify user was saved
	if len(repo.users) != 1 {
		t.Errorf("Expected 1 user in repository, got %d", len(repo.users))
	}

	// New users should have viewer role by default
	if !u.Role().IsViewer() {
		t.Errorf("New user should have viewer role, got %s", u.Role().String())
	}
}

func TestRegisterOrUpdateUser_ExistingUser(t *testing.T) {
	repo := NewMockUserRepository()
	service := NewUserService(repo, nil, nil, nil)

	ctx := context.Background()

	// Create initial user
	googleID := "108123456789"
	u1, _ := service.RegisterOrUpdateUser(ctx, googleID, "old@example.com", "Old Name")
	initialID := u1.ID()

	// Update user with same Google ID but different email/name
	u2, err := service.RegisterOrUpdateUser(ctx, googleID, "new@example.com", "New Name")
	if err != nil {
		t.Fatalf("RegisterOrUpdateUser() error = %v", err)
	}

	// Should be the same user (same ID)
	if !u2.ID().Equals(initialID) {
		t.Error("Updated user should have same ID")
	}

	// Should have updated email and name
	if u2.Email().String() != "new@example.com" {
		t.Errorf("Email should be updated to 'new@example.com', got %s", u2.Email().String())
	}
	if u2.Name().String() != "New Name" {
		t.Errorf("Name should be updated to 'New Name', got %s", u2.Name().String())
	}

	// Should still only have one user
	if len(repo.users) != 1 {
		t.Errorf("Expected 1 user in repository, got %d", len(repo.users))
	}
}

func TestGetUserProfile(t *testing.T) {
	repo := NewMockUserRepository()
	service := NewUserService(repo, nil, nil, nil)

	ctx := context.Background()

	// Create a user
	googleID, _ := user.NewGoogleID("108123456789")
	email, _ := user.NewEmail("test@example.com")
	name, _ := user.NewUserName("Test User")
	role, _ := user.NewRole(user.RoleDeployer)
	u, _ := user.NewUser(googleID, email, name, role)
	repo.users[u.ID().String()] = u

	// Get profile
	retrieved, err := service.GetUserProfile(ctx, u.ID())
	if err != nil {
		t.Fatalf("GetUserProfile() error = %v", err)
	}

	if !retrieved.ID().Equals(u.ID()) {
		t.Error("Retrieved user should match created user")
	}
}

func TestGetUserProfile_NotFound(t *testing.T) {
	repo := NewMockUserRepository()
	service := NewUserService(repo, nil, nil, nil)

	ctx := context.Background()
	nonExistentID := user.NewUserID()

	_, err := service.GetUserProfile(ctx, nonExistentID)
	if err == nil {
		t.Error("GetUserProfile() should return error for non-existent user")
	}
}

func TestListUsers(t *testing.T) {
	repo := NewMockUserRepository()
	service := NewUserService(repo, nil, nil, nil)

	ctx := context.Background()

	// Create users with different roles
	roles := []string{user.RoleViewer, user.RoleDeployer, user.RoleAdmin}
	for i, roleStr := range roles {
		googleID, _ := user.NewGoogleID("10812345678901234567" + string(rune('0'+i)))
		email, _ := user.NewEmail("user" + string(rune('0'+i)) + "@example.com")
		name, _ := user.NewUserName("User " + string(rune('0'+i)))
		role, _ := user.NewRole(roleStr)
		u, _ := user.NewUser(googleID, email, name, role)
		repo.users[u.ID().String()] = u
	}

	// List all users
	users, err := service.ListUsers(ctx, input.UserListFilters{})
	if err != nil {
		t.Fatalf("ListUsers() error = %v", err)
	}

	if len(users) != 3 {
		t.Errorf("Expected 3 users, got %d", len(users))
	}

	// List with role filter
	adminRole, _ := user.NewRole(user.RoleAdmin)
	users, err = service.ListUsers(ctx, input.UserListFilters{Role: &adminRole})
	if err != nil {
		t.Fatalf("ListUsers() error = %v", err)
	}

	if len(users) != 1 {
		t.Errorf("Expected 1 admin user, got %d", len(users))
	}
}

func TestAssignRole(t *testing.T) {
	repo := NewMockUserRepository()
	service := NewUserService(repo, nil, nil, nil)

	ctx := context.Background()

	// Create admin user
	adminGoogleID, _ := user.NewGoogleID("admin123")
	adminEmail, _ := user.NewEmail("admin@example.com")
	adminName, _ := user.NewUserName("Admin")
	adminRole, _ := user.NewRole(user.RoleAdmin)
	admin, _ := user.NewUser(adminGoogleID, adminEmail, adminName, adminRole)
	repo.users[admin.ID().String()] = admin

	// Create target user
	targetGoogleID, _ := user.NewGoogleID("target123")
	targetEmail, _ := user.NewEmail("target@example.com")
	targetName, _ := user.NewUserName("Target")
	targetRole, _ := user.NewRole(user.RoleViewer)
	target, _ := user.NewUser(targetGoogleID, targetEmail, targetName, targetRole)
	repo.users[target.ID().String()] = target

	// Assign deployer role
	newRole, _ := user.NewRole(user.RoleDeployer)
	err := service.AssignRole(ctx, admin.ID(), target.ID(), newRole)
	if err != nil {
		t.Fatalf("AssignRole() error = %v", err)
	}

	// Verify role was updated
	updated, _ := repo.FindByID(ctx, target.ID())
	if !updated.Role().Equals(newRole) {
		t.Errorf("Role should be updated to deployer")
	}
}

func TestAssignRole_Unauthorized(t *testing.T) {
	repo := NewMockUserRepository()
	service := NewUserService(repo, nil, nil, nil)

	ctx := context.Background()

	// Create non-admin user
	googleID, _ := user.NewGoogleID("viewer123")
	email, _ := user.NewEmail("viewer@example.com")
	name, _ := user.NewUserName("Viewer")
	role, _ := user.NewRole(user.RoleViewer)
	viewer, _ := user.NewUser(googleID, email, name, role)
	repo.users[viewer.ID().String()] = viewer

	// Create target user
	targetGoogleID, _ := user.NewGoogleID("target123")
	targetEmail, _ := user.NewEmail("target@example.com")
	targetName, _ := user.NewUserName("Target")
	targetRole, _ := user.NewRole(user.RoleViewer)
	target, _ := user.NewUser(targetGoogleID, targetEmail, targetName, targetRole)
	repo.users[target.ID().String()] = target

	// Try to assign role (should fail)
	newRole, _ := user.NewRole(user.RoleAdmin)
	err := service.AssignRole(ctx, viewer.ID(), target.ID(), newRole)

	if err == nil {
		t.Error("AssignRole() should fail for non-admin user")
	}

	if _, ok := err.(user.ErrUnauthorized); !ok {
		t.Errorf("Expected ErrUnauthorized, got %T", err)
	}
}
