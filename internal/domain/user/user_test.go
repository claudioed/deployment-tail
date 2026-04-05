package user

import (
	"testing"
	"time"
)

func TestNewUserID(t *testing.T) {
	id1 := NewUserID()
	id2 := NewUserID()

	if id1.String() == "" {
		t.Error("UserID should not be empty")
	}

	if id1.Equals(id2) {
		t.Error("Two new UserIDs should not be equal")
	}
}

func TestParseUserID(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid UUID", "123e4567-e89b-12d3-a456-426614174000", false},
		{"invalid format", "not-a-uuid", true},
		{"empty string", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseUserID(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseUserID() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewGoogleID(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid ID", "108123456789012345678", false},
		{"empty string", "", true},
		{"whitespace only", "   ", true},
		{"too long", string(make([]byte, 256)), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewGoogleID(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewGoogleID() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewEmail(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid email", "user@example.com", false},
		{"valid with plus", "user+tag@example.com", false},
		{"empty string", "", true},
		{"no @", "userexample.com", true},
		{"no domain", "user@", true},
		{"no tld", "user@example", true},
		{"too long", "user@" + string(make([]byte, 250)) + ".com", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			email, err := NewEmail(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewEmail() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err == nil && email.String() != tt.input {
				// Email should be normalized to lowercase
				if email.String() != tt.input && email.String() != tt.input {
					t.Errorf("NewEmail() got = %v, want %v", email.String(), tt.input)
				}
			}
		})
	}
}

func TestNewUserName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid name", "John Doe", false},
		{"with trim", "  John Doe  ", false},
		{"empty string", "", true},
		{"whitespace only", "   ", true},
		{"too long", string(make([]byte, 256)), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewUserName(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewUserName() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewRole(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"viewer", RoleViewer, false},
		{"deployer", RoleDeployer, false},
		{"admin", RoleAdmin, false},
		{"invalid", "superadmin", true},
		{"empty", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewRole(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewRole() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRolePermissions(t *testing.T) {
	viewer, _ := NewRole(RoleViewer)
	deployer, _ := NewRole(RoleDeployer)
	admin, _ := NewRole(RoleAdmin)

	tests := []struct {
		role                string
		canCreateSchedule   bool
		canModifyAnySchedule bool
		canManageUsers      bool
	}{
		{RoleViewer, false, false, false},
		{RoleDeployer, true, false, false},
		{RoleAdmin, true, true, true},
	}

	for _, tt := range tests {
		t.Run(tt.role, func(t *testing.T) {
			var role Role
			switch tt.role {
			case RoleViewer:
				role = viewer
			case RoleDeployer:
				role = deployer
			case RoleAdmin:
				role = admin
			}

			if role.CanCreateSchedule() != tt.canCreateSchedule {
				t.Errorf("CanCreateSchedule() = %v, want %v", role.CanCreateSchedule(), tt.canCreateSchedule)
			}
			if role.CanModifyAnySchedule() != tt.canModifyAnySchedule {
				t.Errorf("CanModifyAnySchedule() = %v, want %v", role.CanModifyAnySchedule(), tt.canModifyAnySchedule)
			}
			if role.CanManageUsers() != tt.canManageUsers {
				t.Errorf("CanManageUsers() = %v, want %v", role.CanManageUsers(), tt.canManageUsers)
			}
		})
	}
}

func TestNewUser(t *testing.T) {
	googleID, _ := NewGoogleID("108123456789012345678")
	email, _ := NewEmail("user@example.com")
	name, _ := NewUserName("John Doe")
	role, _ := NewRole(RoleDeployer)

	user, err := NewUser(googleID, email, name, role)
	if err != nil {
		t.Fatalf("NewUser() error = %v", err)
	}

	if user.ID().String() == "" {
		t.Error("User ID should not be empty")
	}
	if !user.GoogleID().Equals(googleID) {
		t.Error("GoogleID mismatch")
	}
	if !user.Email().Equals(email) {
		t.Error("Email mismatch")
	}
	if !user.Name().Equals(name) {
		t.Error("Name mismatch")
	}
	if !user.Role().Equals(role) {
		t.Error("Role mismatch")
	}
	if user.LastLoginAt() != nil {
		t.Error("LastLoginAt should be nil for new user")
	}
}

func TestUserUpdateProfile(t *testing.T) {
	googleID, _ := NewGoogleID("108123456789012345678")
	email, _ := NewEmail("user@example.com")
	name, _ := NewUserName("John Doe")
	role, _ := NewRole(RoleDeployer)

	user, _ := NewUser(googleID, email, name, role)
	oldUpdatedAt := user.UpdatedAt()

	time.Sleep(1 * time.Millisecond) // Ensure time difference

	newName, _ := NewUserName("Jane Doe")
	newEmail, _ := NewEmail("jane@example.com")
	err := user.UpdateProfile(newName, newEmail)
	if err != nil {
		t.Fatalf("UpdateProfile() error = %v", err)
	}

	if !user.Name().Equals(newName) {
		t.Error("Name should be updated")
	}
	if !user.Email().Equals(newEmail) {
		t.Error("Email should be updated")
	}
	if !user.UpdatedAt().After(oldUpdatedAt) {
		t.Error("UpdatedAt should be updated")
	}
}

func TestUserUpdateRole(t *testing.T) {
	googleID, _ := NewGoogleID("108123456789012345678")
	email, _ := NewEmail("user@example.com")
	name, _ := NewUserName("John Doe")
	role, _ := NewRole(RoleDeployer)

	user, _ := NewUser(googleID, email, name, role)
	oldUpdatedAt := user.UpdatedAt()

	time.Sleep(1 * time.Millisecond) // Ensure time difference

	newRole, _ := NewRole(RoleAdmin)
	err := user.UpdateRole(newRole)
	if err != nil {
		t.Fatalf("UpdateRole() error = %v", err)
	}

	if !user.Role().Equals(newRole) {
		t.Error("Role should be updated")
	}
	if !user.UpdatedAt().After(oldUpdatedAt) {
		t.Error("UpdatedAt should be updated")
	}
}

func TestUserRecordLogin(t *testing.T) {
	googleID, _ := NewGoogleID("108123456789012345678")
	email, _ := NewEmail("user@example.com")
	name, _ := NewUserName("John Doe")
	role, _ := NewRole(RoleDeployer)

	user, _ := NewUser(googleID, email, name, role)

	if user.LastLoginAt() != nil {
		t.Error("LastLoginAt should be nil initially")
	}

	user.RecordLogin()

	if user.LastLoginAt() == nil {
		t.Error("LastLoginAt should not be nil after RecordLogin")
	}
}

func TestUserPermissions(t *testing.T) {
	googleID, _ := NewGoogleID("108123456789012345678")
	email, _ := NewEmail("user@example.com")
	name, _ := NewUserName("John Doe")

	ownerID := NewUserID()
	otherID := NewUserID()

	tests := []struct {
		name               string
		userRole           string
		canCreateSchedule  bool
		canModifyOwn       bool
		canModifyOther     bool
		canManageUsers     bool
	}{
		{"viewer", RoleViewer, false, false, false, false},
		{"deployer", RoleDeployer, true, true, false, false},
		{"admin", RoleAdmin, true, true, true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			role, _ := NewRole(tt.userRole)
			user, _ := NewUser(googleID, email, name, role)

			// Override ID for testing ownership
			user.id = ownerID

			if user.CanCreateSchedule() != tt.canCreateSchedule {
				t.Errorf("CanCreateSchedule() = %v, want %v", user.CanCreateSchedule(), tt.canCreateSchedule)
			}
			if user.CanModifySchedule(ownerID) != tt.canModifyOwn {
				t.Errorf("CanModifySchedule(own) = %v, want %v", user.CanModifySchedule(ownerID), tt.canModifyOwn)
			}
			if user.CanModifySchedule(otherID) != tt.canModifyOther {
				t.Errorf("CanModifySchedule(other) = %v, want %v", user.CanModifySchedule(otherID), tt.canModifyOther)
			}
			if user.CanManageUsers() != tt.canManageUsers {
				t.Errorf("CanManageUsers() = %v, want %v", user.CanManageUsers(), tt.canManageUsers)
			}
		})
	}
}
