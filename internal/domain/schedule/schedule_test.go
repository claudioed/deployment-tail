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

	schedule, err := NewSchedule(scheduledAt, serviceName, env, desc)

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
}

func TestScheduleUpdate(t *testing.T) {
	// Create initial schedule
	scheduledAt, _ := NewScheduledTime(time.Now().Add(24 * time.Hour))
	serviceName, _ := NewServiceName("test-service")
	env, _ := NewEnvironment("production")
	desc := NewDescription("Test deployment")

	schedule, _ := NewSchedule(scheduledAt, serviceName, env, desc)

	// Update
	newTime, _ := NewScheduledTime(time.Now().Add(48 * time.Hour))
	err := schedule.Update(&newTime, nil, nil, nil)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if schedule.ScheduledAt() != newTime {
		t.Errorf("expected updated time %v, got %v", newTime, schedule.ScheduledAt())
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
