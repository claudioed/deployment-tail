package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/claudioed/deployment-tail/internal/application/ports/input"
	"github.com/claudioed/deployment-tail/internal/domain/schedule"
)

// MockRepository is a mock implementation of schedule.Repository
type MockRepository struct {
	schedules map[string]*schedule.Schedule
	createErr error
	findErr   error
}

func NewMockRepository() *MockRepository {
	return &MockRepository{
		schedules: make(map[string]*schedule.Schedule),
	}
}

func (m *MockRepository) Create(ctx context.Context, sch *schedule.Schedule) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.schedules[sch.ID().String()] = sch
	return nil
}

func (m *MockRepository) FindByID(ctx context.Context, id schedule.ScheduleID) (*schedule.Schedule, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	sch, ok := m.schedules[id.String()]
	if !ok {
		return nil, schedule.ErrScheduleNotFound
	}
	return sch, nil
}

func (m *MockRepository) FindAll(ctx context.Context, filters schedule.Filters) ([]*schedule.Schedule, error) {
	var result []*schedule.Schedule
	for _, sch := range m.schedules {
		result = append(result, sch)
	}
	return result, nil
}

func (m *MockRepository) Update(ctx context.Context, sch *schedule.Schedule) error {
	if _, ok := m.schedules[sch.ID().String()]; !ok {
		return schedule.ErrScheduleNotFound
	}
	m.schedules[sch.ID().String()] = sch
	return nil
}

func (m *MockRepository) Delete(ctx context.Context, id schedule.ScheduleID) error {
	if _, ok := m.schedules[id.String()]; !ok {
		return schedule.ErrScheduleNotFound
	}
	delete(m.schedules, id.String())
	return nil
}

func TestCreateSchedule(t *testing.T) {
	repo := NewMockRepository()
	service := NewScheduleService(repo)

	cmd := input.CreateScheduleCommand{
		ScheduledAt: time.Now().Add(24 * time.Hour),
		ServiceName: "test-service",
		Environment: "production",
		Description: "Test deployment",
	}

	sch, err := service.CreateSchedule(context.Background(), cmd)

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
	service := NewScheduleService(repo)

	cmd := input.CreateScheduleCommand{
		ScheduledAt: time.Now().Add(24 * time.Hour),
		ServiceName: "test-service",
		Environment: "invalid",
		Description: "Test deployment",
	}

	_, err := service.CreateSchedule(context.Background(), cmd)

	if err == nil {
		t.Fatal("expected error for invalid environment")
	}
}

func TestGetSchedule(t *testing.T) {
	repo := NewMockRepository()
	service := NewScheduleService(repo)

	// Create a schedule first
	cmd := input.CreateScheduleCommand{
		ScheduledAt: time.Now().Add(24 * time.Hour),
		ServiceName: "test-service",
		Environment: "production",
		Description: "Test deployment",
	}

	created, _ := service.CreateSchedule(context.Background(), cmd)

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
	service := NewScheduleService(repo)

	id := schedule.NewScheduleID()

	_, err := service.GetSchedule(context.Background(), id.String())

	if !errors.Is(err, schedule.ErrScheduleNotFound) {
		t.Errorf("expected ErrScheduleNotFound, got %v", err)
	}
}

func TestDeleteSchedule(t *testing.T) {
	repo := NewMockRepository()
	service := NewScheduleService(repo)

	// Create a schedule
	cmd := input.CreateScheduleCommand{
		ScheduledAt: time.Now().Add(24 * time.Hour),
		ServiceName: "test-service",
		Environment: "production",
	}

	created, _ := service.CreateSchedule(context.Background(), cmd)

	// Delete it
	err := service.DeleteSchedule(context.Background(), created.ID().String())

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify it's gone
	_, err = service.GetSchedule(context.Background(), created.ID().String())
	if !errors.Is(err, schedule.ErrScheduleNotFound) {
		t.Error("expected schedule to be deleted")
	}
}
