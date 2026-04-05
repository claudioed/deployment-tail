// +build integration

package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/claudioed/deployment-tail/internal/domain/user"
)

func setupUserTestDB(t *testing.T) *sql.DB {
	dsn := os.Getenv("TEST_DB_DSN")
	if dsn == "" {
		dsn = "root:rootpass@tcp(localhost:3306)/deployment_schedules_test?parseTime=true"
	}

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}

	// Create test table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id CHAR(36) PRIMARY KEY,
			google_id VARCHAR(255) UNIQUE NOT NULL,
			email VARCHAR(255) UNIQUE NOT NULL,
			name VARCHAR(255) NOT NULL,
			role ENUM('viewer', 'deployer', 'admin') NOT NULL DEFAULT 'viewer',
			last_login_at TIMESTAMP NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			INDEX idx_google_id (google_id),
			INDEX idx_email (email),
			INDEX idx_role (role)
		)
	`)
	if err != nil {
		t.Fatalf("failed to create test table: %v", err)
	}

	return db
}

func cleanupUserTestDB(t *testing.T, db *sql.DB) {
	_, err := db.Exec("DELETE FROM users")
	if err != nil {
		t.Logf("warning: failed to cleanup test data: %v", err)
	}
	db.Close()
}

func TestUserRepository_Create(t *testing.T) {
	db := setupUserTestDB(t)
	defer cleanupUserTestDB(t, db)

	repo := NewUserRepository(db)

	googleID, _ := user.NewGoogleID("108123456789012345678")
	email, _ := user.NewEmail("test@example.com")
	name, _ := user.NewUserName("Test User")
	role, _ := user.NewRole(user.RoleDeployer)

	u, _ := user.NewUser(googleID, email, name, role)

	err := repo.Create(context.Background(), u)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify it was saved
	retrieved, err := repo.FindByID(context.Background(), u.ID())
	if err != nil {
		t.Fatalf("expected to find user, got error: %v", err)
	}

	if !retrieved.ID().Equals(u.ID()) {
		t.Errorf("expected ID %v, got %v", u.ID(), retrieved.ID())
	}
	if !retrieved.Email().Equals(email) {
		t.Errorf("expected email %v, got %v", email, retrieved.Email())
	}
}

func TestUserRepository_Create_DuplicateGoogleID(t *testing.T) {
	db := setupUserTestDB(t)
	defer cleanupUserTestDB(t, db)

	repo := NewUserRepository(db)

	googleID, _ := user.NewGoogleID("108123456789012345678")
	email1, _ := user.NewEmail("test1@example.com")
	email2, _ := user.NewEmail("test2@example.com")
	name, _ := user.NewUserName("Test User")
	role, _ := user.NewRole(user.RoleDeployer)

	u1, _ := user.NewUser(googleID, email1, name, role)
	u2, _ := user.NewUser(googleID, email2, name, role)

	// First creation should succeed
	err := repo.Create(context.Background(), u1)
	if err != nil {
		t.Fatalf("expected no error on first create, got %v", err)
	}

	// Second creation with same Google ID should fail
	err = repo.Create(context.Background(), u2)
	if err == nil {
		t.Fatal("expected error on duplicate Google ID, got nil")
	}

	if _, ok := err.(user.ErrUserAlreadyExists); !ok {
		t.Errorf("expected ErrUserAlreadyExists, got %T", err)
	}
}

func TestUserRepository_FindByGoogleID(t *testing.T) {
	db := setupUserTestDB(t)
	defer cleanupUserTestDB(t, db)

	repo := NewUserRepository(db)

	googleID, _ := user.NewGoogleID("108123456789012345678")
	email, _ := user.NewEmail("test@example.com")
	name, _ := user.NewUserName("Test User")
	role, _ := user.NewRole(user.RoleDeployer)

	u, _ := user.NewUser(googleID, email, name, role)
	repo.Create(context.Background(), u)

	// Find by Google ID
	retrieved, err := repo.FindByGoogleID(context.Background(), googleID)
	if err != nil {
		t.Fatalf("expected to find user, got error: %v", err)
	}

	if !retrieved.GoogleID().Equals(googleID) {
		t.Errorf("expected Google ID %v, got %v", googleID, retrieved.GoogleID())
	}
}

func TestUserRepository_FindByEmail(t *testing.T) {
	db := setupUserTestDB(t)
	defer cleanupUserTestDB(t, db)

	repo := NewUserRepository(db)

	googleID, _ := user.NewGoogleID("108123456789012345678")
	email, _ := user.NewEmail("test@example.com")
	name, _ := user.NewUserName("Test User")
	role, _ := user.NewRole(user.RoleDeployer)

	u, _ := user.NewUser(googleID, email, name, role)
	repo.Create(context.Background(), u)

	// Find by email
	retrieved, err := repo.FindByEmail(context.Background(), email)
	if err != nil {
		t.Fatalf("expected to find user, got error: %v", err)
	}

	if !retrieved.Email().Equals(email) {
		t.Errorf("expected email %v, got %v", email, retrieved.Email())
	}
}

func TestUserRepository_Update(t *testing.T) {
	db := setupUserTestDB(t)
	defer cleanupUserTestDB(t, db)

	repo := NewUserRepository(db)

	googleID, _ := user.NewGoogleID("108123456789012345678")
	email, _ := user.NewEmail("test@example.com")
	name, _ := user.NewUserName("Test User")
	role, _ := user.NewRole(user.RoleDeployer)

	u, _ := user.NewUser(googleID, email, name, role)
	repo.Create(context.Background(), u)

	// Update name
	newName, _ := user.NewUserName("Updated Name")
	newEmail, _ := user.NewEmail("updated@example.com")
	u.UpdateProfile(newName, newEmail)

	err := repo.Update(context.Background(), u)
	if err != nil {
		t.Fatalf("expected no error on update, got %v", err)
	}

	// Verify update
	retrieved, _ := repo.FindByID(context.Background(), u.ID())
	if !retrieved.Name().Equals(newName) {
		t.Errorf("expected name %v, got %v", newName, retrieved.Name())
	}
	if !retrieved.Email().Equals(newEmail) {
		t.Errorf("expected email %v, got %v", newEmail, retrieved.Email())
	}
}

func TestUserRepository_UpdateRole(t *testing.T) {
	db := setupUserTestDB(t)
	defer cleanupUserTestDB(t, db)

	repo := NewUserRepository(db)

	googleID, _ := user.NewGoogleID("108123456789012345678")
	email, _ := user.NewEmail("test@example.com")
	name, _ := user.NewUserName("Test User")
	role, _ := user.NewRole(user.RoleDeployer)

	u, _ := user.NewUser(googleID, email, name, role)
	repo.Create(context.Background(), u)

	// Update role
	newRole, _ := user.NewRole(user.RoleAdmin)
	err := repo.UpdateRole(context.Background(), u.ID(), newRole)
	if err != nil {
		t.Fatalf("expected no error on role update, got %v", err)
	}

	// Verify role updated
	retrieved, _ := repo.FindByID(context.Background(), u.ID())
	if !retrieved.Role().Equals(newRole) {
		t.Errorf("expected role %v, got %v", newRole, retrieved.Role())
	}
}

func TestUserRepository_UpdateLastLogin(t *testing.T) {
	db := setupUserTestDB(t)
	defer cleanupUserTestDB(t, db)

	repo := NewUserRepository(db)

	googleID, _ := user.NewGoogleID("108123456789012345678")
	email, _ := user.NewEmail("test@example.com")
	name, _ := user.NewUserName("Test User")
	role, _ := user.NewRole(user.RoleDeployer)

	u, _ := user.NewUser(googleID, email, name, role)
	repo.Create(context.Background(), u)

	// Initially no last login
	if u.LastLoginAt() != nil {
		t.Error("expected no last login initially")
	}

	// Update last login
	err := repo.UpdateLastLogin(context.Background(), u.ID())
	if err != nil {
		t.Fatalf("expected no error on last login update, got %v", err)
	}

	// Verify last login updated
	retrieved, _ := repo.FindByID(context.Background(), u.ID())
	if retrieved.LastLoginAt() == nil {
		t.Error("expected last login to be set")
	}
	if time.Since(*retrieved.LastLoginAt()) > 5*time.Second {
		t.Error("expected last login to be recent")
	}
}

func TestUserRepository_List(t *testing.T) {
	db := setupUserTestDB(t)
	defer cleanupUserTestDB(t, db)

	repo := NewUserRepository(db)

	// Create test users with different roles
	roles := []string{user.RoleViewer, user.RoleDeployer, user.RoleAdmin}
	for i, roleStr := range roles {
		googleID, _ := user.NewGoogleID(fmt.Sprintf("10812345678901234567%d", i))
		email, _ := user.NewEmail(fmt.Sprintf("user%d@example.com", i))
		name, _ := user.NewUserName(fmt.Sprintf("User %d", i))
		role, _ := user.NewRole(roleStr)

		u, _ := user.NewUser(googleID, email, name, role)
		repo.Create(context.Background(), u)
	}

	// List all users
	users, err := repo.List(context.Background(), user.ListFilters{})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(users) != 3 {
		t.Errorf("expected 3 users, got %d", len(users))
	}

	// List with role filter
	adminRole, _ := user.NewRole(user.RoleAdmin)
	users, err = repo.List(context.Background(), user.ListFilters{Role: &adminRole})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(users) != 1 {
		t.Errorf("expected 1 admin user, got %d", len(users))
	}

	// List with pagination
	users, err = repo.List(context.Background(), user.ListFilters{Limit: 2, Offset: 0})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(users) != 2 {
		t.Errorf("expected 2 users with limit, got %d", len(users))
	}
}

func TestUserRepository_FindByID_NotFound(t *testing.T) {
	db := setupUserTestDB(t)
	defer cleanupUserTestDB(t, db)

	repo := NewUserRepository(db)

	nonExistentID := user.NewUserID()
	_, err := repo.FindByID(context.Background(), nonExistentID)

	if err == nil {
		t.Fatal("expected error for non-existent user, got nil")
	}

	if _, ok := err.(user.ErrUserNotFound); !ok {
		t.Errorf("expected ErrUserNotFound, got %T", err)
	}
}
