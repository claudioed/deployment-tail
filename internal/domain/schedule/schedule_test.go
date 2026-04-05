package schedule

import (
	"testing"
	"time"

	"github.com/claudioed/deployment-tail/internal/domain/user"
)

func TestNewSchedule(t *testing.T) {
	scheduledAt, _ := NewScheduledTime(time.Now().Add(24 * time.Hour))
	serviceName, _ := NewServiceName("test-service")
	env1, _ := NewEnvironment("production")
	env2, _ := NewEnvironment("staging")
	environments := []Environment{env1, env2}
	desc := NewDescription("Test deployment")
	owner1, _ := NewOwner("john.doe")
	owner2, _ := NewOwner("jane.smith")
	owners := []Owner{owner1, owner2}
	rollbackPlan, _ := NewRollbackPlan("Revert to v1.2.2")
	createdBy := user.NewUserID()

	schedule, err := NewSchedule(scheduledAt, serviceName, environments, desc, owners, rollbackPlan, createdBy)

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

	if len(schedule.Environments()) != 2 {
		t.Errorf("expected 2 environments, got %d", len(schedule.Environments()))
	}

	if len(schedule.Owners()) != 2 {
		t.Errorf("expected 2 owners, got %d", len(schedule.Owners()))
	}

	if schedule.Status() != StatusCreated {
		t.Errorf("expected new schedule to have status 'created', got %v", schedule.Status())
	}

	if !schedule.RollbackPlan().Equals(rollbackPlan) {
		t.Errorf("expected rollback plan %v, got %v", rollbackPlan, schedule.RollbackPlan())
	}
}

func TestNewSchedule_MinimumRequirements(t *testing.T) {
	scheduledAt, _ := NewScheduledTime(time.Now().Add(24 * time.Hour))
	serviceName, _ := NewServiceName("test-service")
	env, _ := NewEnvironment("production")
	environments := []Environment{env}
	desc := NewDescription("Test deployment")
	owner, _ := NewOwner("john.doe")
	owners := []Owner{owner}
	rollbackPlan, _ := NewRollbackPlan("Revert to v1.2.2")

	schedule, err := NewSchedule(scheduledAt, serviceName, environments, desc, owners, rollbackPlan, user.NewUserID())

	if err != nil {
		t.Fatalf("expected no error with minimum requirements, got %v", err)
	}

	if len(schedule.Environments()) != 1 {
		t.Errorf("expected 1 environment, got %d", len(schedule.Environments()))
	}

	if len(schedule.Owners()) != 1 {
		t.Errorf("expected 1 owner, got %d", len(schedule.Owners()))
	}
}

func TestNewSchedule_ValidationErrors(t *testing.T) {
	scheduledAt, _ := NewScheduledTime(time.Now().Add(24 * time.Hour))
	serviceName, _ := NewServiceName("test-service")
	env, _ := NewEnvironment("production")
	desc := NewDescription("Test deployment")
	owner, _ := NewOwner("john.doe")
	rollbackPlan, _ := NewRollbackPlan("Revert to v1.2.2")

	// Test: no owners
	_, err := NewSchedule(scheduledAt, serviceName, []Environment{env}, desc, []Owner{}, rollbackPlan, user.NewUserID())
	if err == nil {
		t.Error("expected error when creating schedule with no owners")
	}

	// Test: no environments
	_, err = NewSchedule(scheduledAt, serviceName, []Environment{}, desc, []Owner{owner}, rollbackPlan, user.NewUserID())
	if err == nil {
		t.Error("expected error when creating schedule with no environments")
	}
}

func TestSchedule_Deduplication(t *testing.T) {
	scheduledAt, _ := NewScheduledTime(time.Now().Add(24 * time.Hour))
	serviceName, _ := NewServiceName("test-service")
	env, _ := NewEnvironment("production")
	desc := NewDescription("Test deployment")
	owner, _ := NewOwner("john.doe")
	rollbackPlan, _ := NewRollbackPlan("Revert to v1.2.2")

	// Create schedule with duplicate owners and environments
	environments := []Environment{env, env, env}
	owners := []Owner{owner, owner}

	schedule, err := NewSchedule(scheduledAt, serviceName, environments, desc, owners, rollbackPlan, user.NewUserID())

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Should deduplicate to 1 environment and 1 owner
	if len(schedule.Environments()) != 1 {
		t.Errorf("expected 1 deduplicated environment, got %d", len(schedule.Environments()))
	}

	if len(schedule.Owners()) != 1 {
		t.Errorf("expected 1 deduplicated owner, got %d", len(schedule.Owners()))
	}
}

func TestScheduleUpdate(t *testing.T) {
	// Create initial schedule
	scheduledAt, _ := NewScheduledTime(time.Now().Add(24 * time.Hour))
	serviceName, _ := NewServiceName("test-service")
	env, _ := NewEnvironment("production")
	environments := []Environment{env}
	desc := NewDescription("Test deployment")
	owner, _ := NewOwner("john.doe")
	owners := []Owner{owner}
	rollbackPlan, _ := NewRollbackPlan("Revert to v1.2.2")

	schedule, _ := NewSchedule(scheduledAt, serviceName, environments, desc, owners, rollbackPlan, user.NewUserID())

	// Update
	newTime, _ := NewScheduledTime(time.Now().Add(48 * time.Hour))
	newPlan, _ := NewRollbackPlan("New rollback plan")
	stagingEnv, _ := NewEnvironment("staging")
	newEnvironments := []Environment{env, stagingEnv}
	newOwner, _ := NewOwner("jane.smith")
	newOwners := []Owner{owner, newOwner}

	updatedBy := user.NewUserID()
	err := schedule.Update(&newTime, nil, &newEnvironments, nil, &newOwners, &newPlan, updatedBy)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if schedule.ScheduledAt() != newTime {
		t.Errorf("expected updated time %v, got %v", newTime, schedule.ScheduledAt())
	}

	if !schedule.RollbackPlan().Equals(newPlan) {
		t.Errorf("expected rollback plan %v, got %v", newPlan, schedule.RollbackPlan())
	}

	// Check updated environments
	if len(schedule.Environments()) != 2 {
		t.Errorf("expected 2 environments after update, got %d", len(schedule.Environments()))
	}

	// Check updated owners
	if len(schedule.Owners()) != 2 {
		t.Errorf("expected 2 owners after update, got %d", len(schedule.Owners()))
	}
}

func TestScheduleUpdate_ValidationErrors(t *testing.T) {
	scheduledAt, _ := NewScheduledTime(time.Now().Add(24 * time.Hour))
	serviceName, _ := NewServiceName("test-service")
	env, _ := NewEnvironment("production")
	desc := NewDescription("Test deployment")
	owner, _ := NewOwner("john.doe")
	rollbackPlan, _ := NewRollbackPlan("Revert to v1.2.2")

	schedule, _ := NewSchedule(scheduledAt, serviceName, []Environment{env}, desc, []Owner{owner}, rollbackPlan, user.NewUserID())

	// Test: cannot update to empty owners
	emptyOwners := []Owner{}
	err := schedule.Update(nil, nil, nil, nil, &emptyOwners, nil, user.NewUserID())
	if err == nil {
		t.Error("expected error when updating to empty owners list")
	}

	// Test: cannot update to empty environments
	emptyEnvironments := []Environment{}
	err = schedule.Update(nil, nil, &emptyEnvironments, nil, nil, nil, user.NewUserID())
	if err == nil {
		t.Error("expected error when updating to empty environments list")
	}
}

func TestSchedule_AddRemoveOwner(t *testing.T) {
	scheduledAt, _ := NewScheduledTime(time.Now().Add(24 * time.Hour))
	serviceName, _ := NewServiceName("test-service")
	env, _ := NewEnvironment("production")
	desc := NewDescription("Test deployment")
	owner1, _ := NewOwner("john.doe")
	rollbackPlan, _ := NewRollbackPlan("Revert to v1.2.2")

	schedule, _ := NewSchedule(scheduledAt, serviceName, []Environment{env}, desc, []Owner{owner1}, rollbackPlan, user.NewUserID())

	// Add owner
	owner2, _ := NewOwner("jane.smith")
	schedule.AddOwner(owner2)

	if len(schedule.Owners()) != 2 {
		t.Errorf("expected 2 owners after add, got %d", len(schedule.Owners()))
	}

	// Add duplicate owner (should not increase count)
	schedule.AddOwner(owner1)
	if len(schedule.Owners()) != 2 {
		t.Errorf("expected 2 owners after adding duplicate, got %d", len(schedule.Owners()))
	}

	// Remove owner
	schedule.RemoveOwner(owner2)
	if len(schedule.Owners()) != 1 {
		t.Errorf("expected 1 owner after remove, got %d", len(schedule.Owners()))
	}
}

func TestSchedule_AddRemoveEnvironment(t *testing.T) {
	scheduledAt, _ := NewScheduledTime(time.Now().Add(24 * time.Hour))
	serviceName, _ := NewServiceName("test-service")
	env1, _ := NewEnvironment("production")
	desc := NewDescription("Test deployment")
	owner, _ := NewOwner("john.doe")
	rollbackPlan, _ := NewRollbackPlan("Revert to v1.2.2")

	schedule, _ := NewSchedule(scheduledAt, serviceName, []Environment{env1}, desc, []Owner{owner}, rollbackPlan, user.NewUserID())

	// Add environment
	env2, _ := NewEnvironment("staging")
	schedule.AddEnvironment(env2)

	if len(schedule.Environments()) != 2 {
		t.Errorf("expected 2 environments after add, got %d", len(schedule.Environments()))
	}

	// Add duplicate environment (should not increase count)
	schedule.AddEnvironment(env1)
	if len(schedule.Environments()) != 2 {
		t.Errorf("expected 2 environments after adding duplicate, got %d", len(schedule.Environments()))
	}

	// Remove environment
	schedule.RemoveEnvironment(env2)
	if len(schedule.Environments()) != 1 {
		t.Errorf("expected 1 environment after remove, got %d", len(schedule.Environments()))
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

	schedule, _ := NewSchedule(scheduledAt, service, []Environment{environment}, description, []Owner{owner}, rollbackPlan, user.NewUserID())

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

	schedule, _ := NewSchedule(scheduledAt, service, []Environment{environment}, description, []Owner{owner}, rollbackPlan, user.NewUserID())

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
	schedule1, _ := NewSchedule(scheduledAt, service, []Environment{environment}, description, []Owner{owner}, rollbackPlan, user.NewUserID())
	schedule1.Deny()
	if err := schedule1.Approve(); err == nil {
		t.Error("Expected error when approving denied schedule")
	}

	// Test: cannot deny approved schedule
	schedule2, _ := NewSchedule(scheduledAt, service, []Environment{environment}, description, []Owner{owner}, rollbackPlan, user.NewUserID())
	schedule2.Approve()
	if err := schedule2.Deny(); err == nil {
		t.Error("Expected error when denying approved schedule")
	}
}

func TestSchedule_AuditFields(t *testing.T) {
	scheduledAt, _ := NewScheduledTime(time.Now().Add(24 * time.Hour))
	serviceName, _ := NewServiceName("test-service")
	env, _ := NewEnvironment("production")
	desc := NewDescription("Test deployment")
	owner, _ := NewOwner("john.doe")
	rollbackPlan, _ := NewRollbackPlan("Revert to v1.2.2")
	createdBy := user.NewUserID()

	schedule, err := NewSchedule(scheduledAt, serviceName, []Environment{env}, desc, []Owner{owner}, rollbackPlan, createdBy)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify CreatedBy is set
	if !schedule.CreatedBy().Equals(createdBy) {
		t.Error("CreatedBy should match the user who created the schedule")
	}

	// Verify UpdatedBy initially equals CreatedBy
	if !schedule.UpdatedBy().Equals(createdBy) {
		t.Error("UpdatedBy should initially equal CreatedBy")
	}

	// Update schedule with different user
	updatedBy := user.NewUserID()
	newTime, _ := NewScheduledTime(time.Now().Add(48 * time.Hour))
	err = schedule.Update(&newTime, nil, nil, nil, nil, nil, updatedBy)
	if err != nil {
		t.Fatalf("expected no error on update, got %v", err)
	}

	// Verify UpdatedBy changed but CreatedBy did not
	if !schedule.CreatedBy().Equals(createdBy) {
		t.Error("CreatedBy should never change")
	}
	if !schedule.UpdatedBy().Equals(updatedBy) {
		t.Error("UpdatedBy should reflect the user who last updated the schedule")
	}
	if schedule.UpdatedBy().Equals(createdBy) {
		t.Error("UpdatedBy should be different from CreatedBy after update")
	}
}
