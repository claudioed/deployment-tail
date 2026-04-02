package schedule

import (
	"testing"
	"time"
)

func TestNewSchedule(t *testing.T) {
	scheduledAt, _ := NewScheduledTime(time.Now().Add(24 * time.Hour))
	serviceName, _ := NewServiceName("test-service")
	env, _ := NewEnvironment("production")
	desc := NewDescription("Test deployment")
	owner, _ := NewOwner("john.doe")
	rollbackPlan, _ := NewRollbackPlan("Revert to v1.2.2")

	schedule, err := NewSchedule(scheduledAt, serviceName, env, desc, owner, rollbackPlan)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if schedule == nil {
		t.Fatal("expected schedule to be created")
	}

	if schedule.ScheduledAt() != scheduledAt {
		t.Errorf("expected scheduled time %v, got %v", scheduledAt, schedule.ScheduledAt())
	}

	if schedule.Service() != serviceName {
		t.Errorf("expected service name %v, got %v", serviceName, schedule.Service())
	}

	if schedule.Environment() != env {
		t.Errorf("expected environment %v, got %v", env, schedule.Environment())
	}

	if schedule.Status() != StatusCreated {
		t.Errorf("expected new schedule to have status 'created', got %v", schedule.Status())
	}

	if !schedule.Owner().Equals(owner) {
		t.Errorf("expected owner %v, got %v", owner, schedule.Owner())
	}

	if !schedule.RollbackPlan().Equals(rollbackPlan) {
		t.Errorf("expected rollback plan %v, got %v", rollbackPlan, schedule.RollbackPlan())
	}
}

func TestScheduleUpdate(t *testing.T) {
	// Create initial schedule
	scheduledAt, _ := NewScheduledTime(time.Now().Add(24 * time.Hour))
	serviceName, _ := NewServiceName("test-service")
	env, _ := NewEnvironment("production")
	desc := NewDescription("Test deployment")
	owner, _ := NewOwner("john.doe")
	rollbackPlan, _ := NewRollbackPlan("Revert to v1.2.2")

	schedule, _ := NewSchedule(scheduledAt, serviceName, env, desc, owner, rollbackPlan)

	// Update
	newTime, _ := NewScheduledTime(time.Now().Add(48 * time.Hour))
	newPlan, _ := NewRollbackPlan("New rollback plan")
	err := schedule.Update(&newTime, nil, nil, nil, &newPlan)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if schedule.ScheduledAt() != newTime {
		t.Errorf("expected updated time %v, got %v", newTime, schedule.ScheduledAt())
	}

	if !schedule.RollbackPlan().Equals(newPlan) {
		t.Errorf("expected rollback plan %v, got %v", newPlan, schedule.RollbackPlan())
	}

	// Owner should remain unchanged (immutable)
	if !schedule.Owner().Equals(owner) {
		t.Error("Owner should not change during update")
	}
}

func TestEnvironmentValidation(t *testing.T) {
	tests := []struct {
		name    string
		env     string
		wantErr bool
	}{
		{"valid production", "production", false},
		{"valid staging", "staging", false},
		{"valid development", "development", false},
		{"invalid environment", "invalid", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewEnvironment(tt.env)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewEnvironment() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestServiceNameValidation(t *testing.T) {
	tests := []struct {
		name    string
		service string
		wantErr bool
	}{
		{"valid name", "api-service", false},
		{"empty name", "", true},
		{"whitespace only", "   ", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewServiceName(tt.service)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewServiceName() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSchedule_Approve(t *testing.T) {
	scheduledAt, _ := NewScheduledTime(time.Now().Add(24 * time.Hour))
	service, _ := NewServiceName("api-service")
	environment, _ := NewEnvironment("production")
	description := NewDescription("Deploy v1.2.3")
	owner, _ := NewOwner("john.doe")
	rollbackPlan, _ := NewRollbackPlan("Revert to v1.2.2")

	schedule, _ := NewSchedule(scheduledAt, service, environment, description, owner, rollbackPlan)

	// Approve created schedule - should succeed
	err := schedule.Approve()
	if err != nil {
		t.Errorf("Approve() error = %v", err)
	}

	if schedule.Status() != StatusApproved {
		t.Errorf("Expected status 'approved', got %v", schedule.Status())
	}

	// Try to approve again - should fail
	err = schedule.Approve()
	if err == nil {
		t.Error("Expected error when approving already approved schedule")
	}
}

func TestSchedule_Deny(t *testing.T) {
	scheduledAt, _ := NewScheduledTime(time.Now().Add(24 * time.Hour))
	service, _ := NewServiceName("api-service")
	environment, _ := NewEnvironment("production")
	description := NewDescription("Deploy v1.2.3")
	owner, _ := NewOwner("john.doe")
	rollbackPlan, _ := NewRollbackPlan("Revert to v1.2.2")

	schedule, _ := NewSchedule(scheduledAt, service, environment, description, owner, rollbackPlan)

	// Deny created schedule - should succeed
	err := schedule.Deny()
	if err != nil {
		t.Errorf("Deny() error = %v", err)
	}

	if schedule.Status() != StatusDenied {
		t.Errorf("Expected status 'denied', got %v", schedule.Status())
	}

	// Try to deny again - should fail
	err = schedule.Deny()
	if err == nil {
		t.Error("Expected error when denying already denied schedule")
	}
}

func TestSchedule_InvalidStatusTransitions(t *testing.T) {
	scheduledAt, _ := NewScheduledTime(time.Now().Add(24 * time.Hour))
	service, _ := NewServiceName("api-service")
	environment, _ := NewEnvironment("production")
	description := NewDescription("Deploy v1.2.3")
	owner, _ := NewOwner("john.doe")
	rollbackPlan, _ := NewRollbackPlan("Revert to v1.2.2")

	// Test: cannot approve denied schedule
	schedule1, _ := NewSchedule(scheduledAt, service, environment, description, owner, rollbackPlan)
	schedule1.Deny()
	if err := schedule1.Approve(); err == nil {
		t.Error("Expected error when approving denied schedule")
	}

	// Test: cannot deny approved schedule
	schedule2, _ := NewSchedule(scheduledAt, service, environment, description, owner, rollbackPlan)
	schedule2.Approve()
	if err := schedule2.Deny(); err == nil {
		t.Error("Expected error when denying approved schedule")
	}
}
