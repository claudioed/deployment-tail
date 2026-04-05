package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/claudioed/deployment-tail/internal/application/ports/input"
	"github.com/claudioed/deployment-tail/internal/domain/schedule"
	"github.com/claudioed/deployment-tail/internal/domain/user"
)

// Test helper functions
func createTestDeployer(t *testing.T, userRepo *MockUserRepository) *user.User {
	t.Helper()
	googleID, _ := user.NewGoogleID("deployer123")
	email, _ := user.NewEmail("deployer@example.com")
	name, _ := user.NewUserName("Test Deployer")
	role, _ := user.NewRole(user.RoleDeployer)
	u, _ := user.NewUser(googleID, email, name, role)
	userRepo.users[u.ID().String()] = u
	return u
}

func createTestAdmin(t *testing.T, userRepo *MockUserRepository) *user.User {
	t.Helper()
	googleID, _ := user.NewGoogleID("admin123")
	email, _ := user.NewEmail("admin@example.com")
	name, _ := user.NewUserName("Test Admin")
	role, _ := user.NewRole(user.RoleAdmin)
	u, _ := user.NewUser(googleID, email, name, role)
	userRepo.users[u.ID().String()] = u
	return u
}

func createTestViewer(t *testing.T, userRepo *MockUserRepository) *user.User {
	t.Helper()
	googleID, _ := user.NewGoogleID("viewer123")
	email, _ := user.NewEmail("viewer@example.com")
	name, _ := user.NewUserName("Test Viewer")
	role, _ := user.NewRole(user.RoleViewer)
	u, _ := user.NewUser(googleID, email, name, role)
	userRepo.users[u.ID().String()] = u
	return u
}

func TestCreateSchedule(t *testing.T) {
	repo := NewMockRepository()
	userRepo := NewMockUserRepository()
	deployer := createTestDeployer(t, userRepo)
	service := NewScheduleService(repo, userRepo)

	cmd := input.CreateScheduleCommand{
		ScheduledAt:  time.Now().Add(24 * time.Hour),
		ServiceName:  "test-service",
		Environments: []string{"production"},
		Description:  "Test deployment",
		Owners:       []string{"test-user"},
	}

	sch, err := service.CreateSchedule(context.Background(), cmd, deployer.ID())

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if sch == nil {
		t.Fatal("expected schedule to be created")
	}

	if sch.Service().Value() != "test-service" {
		t.Errorf("expected service 'test-service', got %v", sch.Service().Value())
	}
}

func TestCreateScheduleWithInvalidEnvironment(t *testing.T) {
	repo := NewMockRepository()
	userRepo := NewMockUserRepository()
	deployer := createTestDeployer(t, userRepo)
	service := NewScheduleService(repo, userRepo)

	cmd := input.CreateScheduleCommand{
		ScheduledAt:  time.Now().Add(24 * time.Hour),
		Owners:       []string{"test-user"},
		ServiceName:  "test-service",
		Environments: []string{"invalid"},
		Description:  "Test deployment",
	}

	_, err := service.CreateSchedule(context.Background(), cmd, deployer.ID())

	if err == nil {
		t.Fatal("expected error for invalid environment")
	}
}

func TestGetSchedule(t *testing.T) {
	repo := NewMockRepository()
	userRepo := NewMockUserRepository()
	deployer := createTestDeployer(t, userRepo)
	service := NewScheduleService(repo, userRepo)

	// Create a schedule first
	cmd := input.CreateScheduleCommand{
		ScheduledAt:  time.Now().Add(24 * time.Hour),
		ServiceName:  "test-service",
		Environments: []string{"production"},
		Description:  "Test deployment",
		Owners:       []string{"test-user"},
	}

	created, _ := service.CreateSchedule(context.Background(), cmd, deployer.ID())

	// Get the schedule
	retrieved, err := service.GetSchedule(context.Background(), created.ID().String())

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if retrieved.ID() != created.ID() {
		t.Errorf("expected ID %v, got %v", created.ID(), retrieved.ID())
	}
}

func TestGetScheduleNotFound(t *testing.T) {
	repo := NewMockRepository()
	userRepo := NewMockUserRepository()
	service := NewScheduleService(repo, userRepo)

	id := schedule.NewScheduleID()

	_, err := service.GetSchedule(context.Background(), id.String())

	if !errors.Is(err, schedule.ErrScheduleNotFound) {
		t.Errorf("expected ErrScheduleNotFound, got %v", err)
	}
}

func TestDeleteSchedule(t *testing.T) {
	repo := NewMockRepository()
	userRepo := NewMockUserRepository()
	deployer := createTestDeployer(t, userRepo)
	service := NewScheduleService(repo, userRepo)

	// Create a schedule
	cmd := input.CreateScheduleCommand{
		ScheduledAt:  time.Now().Add(24 * time.Hour),
		ServiceName:  "test-service",
		Environments: []string{"production"},
		Owners:       []string{"test-user"},
	}

	created, _ := service.CreateSchedule(context.Background(), cmd, deployer.ID())

	// Delete it
	err := service.DeleteSchedule(context.Background(), created.ID().String(), deployer.ID())

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify it's gone
	_, err = service.GetSchedule(context.Background(), created.ID().String())
	if !errors.Is(err, schedule.ErrScheduleNotFound) {
		t.Error("expected schedule to be deleted")
	}
}

func TestCreateSchedule_ViewerUnauthorized(t *testing.T) {
	repo := NewMockRepository()
	userRepo := NewMockUserRepository()
	viewer := createTestViewer(t, userRepo)
	service := NewScheduleService(repo, userRepo)

	cmd := input.CreateScheduleCommand{
		ScheduledAt:  time.Now().Add(24 * time.Hour),
		ServiceName:  "test-service",
		Environments: []string{"production"},
		Description:  "Test deployment",
		Owners:       []string{"test-user"},
	}

	_, err := service.CreateSchedule(context.Background(), cmd, viewer.ID())

	if err == nil {
		t.Fatal("expected error for viewer creating schedule")
	}

	if _, ok := err.(user.ErrUnauthorized); !ok {
		t.Errorf("expected ErrUnauthorized, got %T", err)
	}
}

func TestUpdateSchedule_DeployerCanUpdateOwn(t *testing.T) {
	repo := NewMockRepository()
	userRepo := NewMockUserRepository()
	deployer := createTestDeployer(t, userRepo)
	service := NewScheduleService(repo, userRepo)

	// Create a schedule
	cmd := input.CreateScheduleCommand{
		ScheduledAt:  time.Now().Add(24 * time.Hour),
		ServiceName:  "test-service",
		Environments: []string{"production"},
		Owners:       []string{"test-user"},
	}

	created, _ := service.CreateSchedule(context.Background(), cmd, deployer.ID())

	// Update it
	newService := "updated-service"
	updateCmd := input.UpdateScheduleCommand{
		ID:          created.ID().String(),
		ServiceName: &newService,
	}

	updated, err := service.UpdateSchedule(context.Background(), updateCmd, deployer.ID())

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if updated.Service().Value() != "updated-service" {
		t.Errorf("expected service to be updated")
	}
}

func TestUpdateSchedule_DeployerCannotUpdateOthers(t *testing.T) {
	repo := NewMockRepository()
	userRepo := NewMockUserRepository()
	deployer1 := createTestDeployer(t, userRepo)

	// Create second deployer
	googleID, _ := user.NewGoogleID("deployer456")
	email, _ := user.NewEmail("deployer2@example.com")
	name, _ := user.NewUserName("Test Deployer 2")
	role, _ := user.NewRole(user.RoleDeployer)
	deployer2, _ := user.NewUser(googleID, email, name, role)
	userRepo.users[deployer2.ID().String()] = deployer2

	service := NewScheduleService(repo, userRepo)

	// Deployer1 creates a schedule
	cmd := input.CreateScheduleCommand{
		ScheduledAt:  time.Now().Add(24 * time.Hour),
		ServiceName:  "test-service",
		Environments: []string{"production"},
		Owners:       []string{"test-user"},
	}

	created, _ := service.CreateSchedule(context.Background(), cmd, deployer1.ID())

	// Deployer2 tries to update it
	newService := "hacked-service"
	updateCmd := input.UpdateScheduleCommand{
		ID:          created.ID().String(),
		ServiceName: &newService,
	}

	_, err := service.UpdateSchedule(context.Background(), updateCmd, deployer2.ID())

	if err == nil {
		t.Fatal("expected error for deployer updating another's schedule")
	}

	if _, ok := err.(user.ErrUnauthorized); !ok {
		t.Errorf("expected ErrUnauthorized, got %T", err)
	}
}

func TestUpdateSchedule_AdminCanUpdateAny(t *testing.T) {
	repo := NewMockRepository()
	userRepo := NewMockUserRepository()
	deployer := createTestDeployer(t, userRepo)
	admin := createTestAdmin(t, userRepo)
	service := NewScheduleService(repo, userRepo)

	// Deployer creates a schedule
	cmd := input.CreateScheduleCommand{
		ScheduledAt:  time.Now().Add(24 * time.Hour),
		ServiceName:  "test-service",
		Environments: []string{"production"},
		Owners:       []string{"test-user"},
	}

	created, _ := service.CreateSchedule(context.Background(), cmd, deployer.ID())

	// Admin updates it
	newService := "admin-updated-service"
	updateCmd := input.UpdateScheduleCommand{
		ID:          created.ID().String(),
		ServiceName: &newService,
	}

	updated, err := service.UpdateSchedule(context.Background(), updateCmd, admin.ID())

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if updated.Service().Value() != "admin-updated-service" {
		t.Errorf("expected service to be updated by admin")
	}
}

func TestDeleteSchedule_DeployerCanDeleteOwn(t *testing.T) {
	repo := NewMockRepository()
	userRepo := NewMockUserRepository()
	deployer := createTestDeployer(t, userRepo)
	service := NewScheduleService(repo, userRepo)

	// Create a schedule
	cmd := input.CreateScheduleCommand{
		ScheduledAt:  time.Now().Add(24 * time.Hour),
		ServiceName:  "test-service",
		Environments: []string{"production"},
		Owners:       []string{"test-user"},
	}

	created, _ := service.CreateSchedule(context.Background(), cmd, deployer.ID())

	// Delete it
	err := service.DeleteSchedule(context.Background(), created.ID().String(), deployer.ID())

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestDeleteSchedule_DeployerCannotDeleteOthers(t *testing.T) {
	repo := NewMockRepository()
	userRepo := NewMockUserRepository()
	deployer1 := createTestDeployer(t, userRepo)

	// Create second deployer
	googleID, _ := user.NewGoogleID("deployer456")
	email, _ := user.NewEmail("deployer2@example.com")
	name, _ := user.NewUserName("Test Deployer 2")
	role, _ := user.NewRole(user.RoleDeployer)
	deployer2, _ := user.NewUser(googleID, email, name, role)
	userRepo.users[deployer2.ID().String()] = deployer2

	service := NewScheduleService(repo, userRepo)

	// Deployer1 creates a schedule
	cmd := input.CreateScheduleCommand{
		ScheduledAt:  time.Now().Add(24 * time.Hour),
		ServiceName:  "test-service",
		Environments: []string{"production"},
		Owners:       []string{"test-user"},
	}

	created, _ := service.CreateSchedule(context.Background(), cmd, deployer1.ID())

	// Deployer2 tries to delete it
	err := service.DeleteSchedule(context.Background(), created.ID().String(), deployer2.ID())

	if err == nil {
		t.Fatal("expected error for deployer deleting another's schedule")
	}

	if _, ok := err.(user.ErrUnauthorized); !ok {
		t.Errorf("expected ErrUnauthorized, got %T", err)
	}
}

func TestDeleteSchedule_AdminCanDeleteAny(t *testing.T) {
	repo := NewMockRepository()
	userRepo := NewMockUserRepository()
	deployer := createTestDeployer(t, userRepo)
	admin := createTestAdmin(t, userRepo)
	service := NewScheduleService(repo, userRepo)

	// Deployer creates a schedule
	cmd := input.CreateScheduleCommand{
		ScheduledAt:  time.Now().Add(24 * time.Hour),
		ServiceName:  "test-service",
		Environments: []string{"production"},
		Owners:       []string{"test-user"},
	}

	created, _ := service.CreateSchedule(context.Background(), cmd, deployer.ID())

	// Admin deletes it
	err := service.DeleteSchedule(context.Background(), created.ID().String(), admin.ID())

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}
