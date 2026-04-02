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

	// Create test table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS schedules (
			id VARCHAR(36) PRIMARY KEY,
			scheduled_at DATETIME NOT NULL,
			service_name VARCHAR(255) NOT NULL,
			environment ENUM('production', 'staging', 'development') NOT NULL,
			description TEXT,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
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
		t.Logf("warning: failed to cleanup test data: %v", err)
	}
	db.Close()
}

func TestRepositoryCreate(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	repo := NewScheduleRepository(db)

	scheduledAt, _ := schedule.NewScheduledTime(time.Now().Add(24 * time.Hour))
	serviceName, _ := schedule.NewServiceName("test-service")
	env, _ := schedule.NewEnvironment("production")
	desc := schedule.NewDescription("Test")

	sch, _ := schedule.NewSchedule(scheduledAt, serviceName, env, desc)

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
}

func TestRepositoryFindAll(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	repo := NewScheduleRepository(db)

	// Create test schedules
	for i := 0; i < 3; i++ {
		scheduledAt, _ := schedule.NewScheduledTime(time.Now().Add(time.Duration(i*24) * time.Hour))
		serviceName, _ := schedule.NewServiceName("test-service")
		env, _ := schedule.NewEnvironment("production")
		desc := schedule.NewDescription("Test")

		sch, _ := schedule.NewSchedule(scheduledAt, serviceName, env, desc)
		repo.Create(context.Background(), sch)
	}

	schedules, err := repo.FindAll(context.Background(), schedule.Filters{})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(schedules) != 3 {
		t.Errorf("expected 3 schedules, got %d", len(schedules))
	}
}
