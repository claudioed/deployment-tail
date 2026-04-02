package schedule

import (
	"fmt"
	"time"
)

// Schedule is the aggregate root for deployment schedules
type Schedule struct {
	id          ScheduleID
	scheduledAt ScheduledTime
	service     ServiceName
	environment Environment
	description Description
	createdAt   time.Time
	updatedAt   time.Time
}

// NewSchedule creates a new schedule
func NewSchedule(
	scheduledAt ScheduledTime,
	service ServiceName,
	environment Environment,
	description Description,
) (*Schedule, error) {
	now := time.Now().UTC()

	return &Schedule{
		id:          NewScheduleID(),
		scheduledAt: scheduledAt,
		service:     service,
		environment: environment,
		description: description,
		createdAt:   now,
		updatedAt:   now,
	}, nil
}

// Reconstitute recreates a schedule from storage
func Reconstitute(
	id ScheduleID,
	scheduledAt ScheduledTime,
	service ServiceName,
	environment Environment,
	description Description,
	createdAt time.Time,
	updatedAt time.Time,
) *Schedule {
	return &Schedule{
		id:          id,
		scheduledAt: scheduledAt,
		service:     service,
		environment: environment,
		description: description,
		createdAt:   createdAt,
		updatedAt:   updatedAt,
	}
}

// Update updates the schedule fields
func (s *Schedule) Update(
	scheduledAt *ScheduledTime,
	service *ServiceName,
	environment *Environment,
	description *Description,
) error {
	if scheduledAt != nil {
		s.scheduledAt = *scheduledAt
	}
	if service != nil {
		s.service = *service
	}
	if environment != nil {
		if !environment.IsValid() {
			return fmt.Errorf("invalid environment")
		}
		s.environment = *environment
	}
	if description != nil {
		s.description = *description
	}
	s.updatedAt = time.Now().UTC()
	return nil
}

// Getters

func (s *Schedule) ID() ScheduleID {
	return s.id
}

func (s *Schedule) ScheduledAt() ScheduledTime {
	return s.scheduledAt
}

func (s *Schedule) Service() ServiceName {
	return s.service
}

func (s *Schedule) Environment() Environment {
	return s.environment
}

func (s *Schedule) Description() Description {
	return s.description
}

func (s *Schedule) CreatedAt() time.Time {
	return s.createdAt
}

func (s *Schedule) UpdatedAt() time.Time {
	return s.updatedAt
}
