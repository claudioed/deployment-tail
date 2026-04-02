package input

import (
	"context"
	"time"

	"github.com/claudioed/deployment-tail/internal/domain/schedule"
)

// CreateScheduleCommand represents the command to create a schedule
type CreateScheduleCommand struct {
	ScheduledAt time.Time
	ServiceName string
	Environment string
	Description string
}

// UpdateScheduleCommand represents the command to update a schedule
type UpdateScheduleCommand struct {
	ID          string
	ScheduledAt *time.Time
	ServiceName *string
	Environment *string
	Description *string
}

// ListSchedulesQuery represents the query to list schedules
type ListSchedulesQuery struct {
	From        *time.Time
	To          *time.Time
	Environment *string
}

// ScheduleService defines the inbound port for schedule operations
type ScheduleService interface {
	// CreateSchedule creates a new schedule
	CreateSchedule(ctx context.Context, cmd CreateScheduleCommand) (*schedule.Schedule, error)

	// GetSchedule retrieves a schedule by ID
	GetSchedule(ctx context.Context, id string) (*schedule.Schedule, error)

	// ListSchedules retrieves schedules with optional filters
	ListSchedules(ctx context.Context, query ListSchedulesQuery) ([]*schedule.Schedule, error)

	// UpdateSchedule updates an existing schedule
	UpdateSchedule(ctx context.Context, cmd UpdateScheduleCommand) (*schedule.Schedule, error)

	// DeleteSchedule deletes a schedule
	DeleteSchedule(ctx context.Context, id string) error
}
