// +build integration

package mysql

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/claudioed/deployment-tail/internal/domain/schedule"
	"github.com/claudioed/deployment-tail/internal/domain/user"
)

func setupTestDB(t *testing.T) *sql.DB {
	dsn := os.Getenv("TEST_DB_DSN")
	if dsn == "" {
		dsn = "root:rootpass@tcp(localhost:3306)/deployment_schedules_test?parseTime=true"
	}

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}

	// Create test tables
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id VARCHAR(36) PRIMARY KEY,
			google_id VARCHAR(255) UNIQUE NOT NULL,
			email VARCHAR(255) UNIQUE NOT NULL,
			name VARCHAR(255) NOT NULL,
			role ENUM('viewer', 'deployer', 'admin') NOT NULL DEFAULT 'viewer',
			last_login_at DATETIME,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		t.Fatalf("failed to create users table: %v", err)
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS schedules (
			id VARCHAR(36) PRIMARY KEY,
			scheduled_at DATETIME NOT NULL,
			service_name VARCHAR(255) NOT NULL,
			environment ENUM('production', 'staging', 'development') NOT NULL,
			description TEXT,
			created_by VARCHAR(36) NOT NULL,
			updated_by VARCHAR(36) NOT NULL,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			FOREIGN KEY (created_by) REFERENCES users(id),
			FOREIGN KEY (updated_by) REFERENCES users(id)
		)
	`)
	if err != nil {
		t.Fatalf("failed to create test table: %v", err)
	}

	return db
}

func cleanupTestDB(t *testing.T, db *sql.DB) {
	_, err := db.Exec("DELETE FROM schedules")
	if err != nil {
		t.Logf("warning: failed to cleanup schedules: %v", err)
	}
	_, err = db.Exec("DELETE FROM users")
	if err != nil {
		t.Logf("warning: failed to cleanup users: %v", err)
	}
	db.Close()
}

// createTestUser creates a test user for schedule integration tests
func createTestUser(t *testing.T, db *sql.DB) user.UserID {
	googleID, _ := user.NewGoogleID("test-google-id-123")
	email, _ := user.NewEmail("test@example.com")
	userName, _ := user.NewUserName("Test User")
	role, _ := user.NewRole(user.RoleDeployer)

	testUser, err := user.NewUser(googleID, email, userName, role)
	if err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}

	_, err = db.Exec(`
		INSERT INTO users (id, google_id, email, name, role, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, testUser.ID().String(), testUser.GoogleID().String(), testUser.Email().String(),
		testUser.Name().String(), testUser.Role().String(), time.Now(), time.Now())

	if err != nil {
		t.Fatalf("failed to insert test user: %v", err)
	}

	return testUser.ID()
}

func TestRepositoryCreate(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	repo := NewScheduleRepository(db)
	testUserID := createTestUser(t, db)

	scheduledAt, _ := schedule.NewScheduledTime(time.Now().Add(24 * time.Hour))
	serviceName, _ := schedule.NewServiceName("test-service")
	env, _ := schedule.NewEnvironment("production")
	desc := schedule.NewDescription("Test")
	owner, _ := schedule.NewOwner("test-user")
	rollbackPlan, _ := schedule.NewRollbackPlan("")

	sch, _ := schedule.NewSchedule(scheduledAt, serviceName, []schedule.Environment{env}, desc, []schedule.Owner{owner}, rollbackPlan, testUserID)

	err := repo.Create(context.Background(), sch)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify it was saved
	retrieved, err := repo.FindByID(context.Background(), sch.ID())
	if err != nil {
		t.Fatalf("expected to find schedule, got error: %v", err)
	}

	if retrieved.ID() != sch.ID() {
		t.Errorf("expected ID %v, got %v", sch.ID(), retrieved.ID())
	}

	// Verify audit fields
	if retrieved.CreatedBy() != testUserID {
		t.Errorf("expected createdBy %v, got %v", testUserID, retrieved.CreatedBy())
	}

	if retrieved.UpdatedBy() != testUserID {
		t.Errorf("expected updatedBy %v, got %v", testUserID, retrieved.UpdatedBy())
	}
}

func TestRepositoryFindAll(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	repo := NewScheduleRepository(db)
	testUserID := createTestUser(t, db)

	// Create test schedules
	for i := 0; i < 3; i++ {
		scheduledAt, _ := schedule.NewScheduledTime(time.Now().Add(time.Duration(i*24) * time.Hour))
		serviceName, _ := schedule.NewServiceName("test-service")
		env, _ := schedule.NewEnvironment("production")
		desc := schedule.NewDescription("Test")
		owner, _ := schedule.NewOwner("test-user")
		rollbackPlan, _ := schedule.NewRollbackPlan("")

		sch, _ := schedule.NewSchedule(scheduledAt, serviceName, []schedule.Environment{env}, desc, []schedule.Owner{owner}, rollbackPlan, testUserID)
		repo.Create(context.Background(), sch)
	}

	schedules, err := repo.FindAll(context.Background(), schedule.Filters{})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(schedules) != 3 {
		t.Errorf("expected 3 schedules, got %d", len(schedules))
	}

	// Verify all schedules have audit fields populated
	for _, sch := range schedules {
		if sch.CreatedBy() != testUserID {
			t.Errorf("schedule %v: expected createdBy %v, got %v", sch.ID(), testUserID, sch.CreatedBy())
		}
		if sch.UpdatedBy() != testUserID {
			t.Errorf("schedule %v: expected updatedBy %v, got %v", sch.ID(), testUserID, sch.UpdatedBy())
		}
	}
}
