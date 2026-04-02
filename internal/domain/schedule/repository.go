package schedule

import (
	"context"
	"time"
)

// Repository defines the interface for schedule persistence
type Repository interface {
	// Create saves a new schedule
	Create(ctx context.Context, schedule *Schedule) error

	// FindByID retrieves a schedule by its ID
	FindByID(ctx context.Context, id ScheduleID) (*Schedule, error)

	// FindAll retrieves schedules with optional filters
	FindAll(ctx context.Context, filters Filters) ([]*Schedule, error)

	// Update updates an existing schedule
	Update(ctx context.Context, schedule *Schedule) error

	// Delete removes a schedule
	Delete(ctx context.Context, id ScheduleID) error
}

// Filters represents query filters for schedules
type Filters struct {
	From        *time.Time
	To          *time.Time
	Environment *Environment
}
