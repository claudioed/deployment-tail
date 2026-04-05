package schedule

import (
	"context"
	"time"

	"github.com/claudioed/deployment-tail/internal/domain/user"
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

	// Delete soft-deletes a schedule
	Delete(ctx context.Context, id ScheduleID, deletedBy user.UserID) error

	// FindUngrouped retrieves schedules that are not assigned to any group
	FindUngrouped(ctx context.Context, filters Filters) ([]*Schedule, error)
}

// Filters represents query filters for schedules
type Filters struct {
	From         *time.Time
	To           *time.Time
	Environments []Environment // OR logic - match ANY
	Owners       []Owner       // OR logic - match ANY
	Status       *Status
}
