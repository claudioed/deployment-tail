package schedule

import (
	"fmt"
	"time"
)

// Schedule is the aggregate root for deployment schedules
type Schedule struct {
	id           ScheduleID
	scheduledAt  ScheduledTime
	service      ServiceName
	environment  Environment
	description  Description
	owner        Owner
	status       Status
	rollbackPlan RollbackPlan
	createdAt    time.Time
	updatedAt    time.Time
}

// NewSchedule creates a new schedule
func NewSchedule(
	scheduledAt ScheduledTime,
	service ServiceName,
	environment Environment,
	description Description,
	owner Owner,
	rollbackPlan RollbackPlan,
) (*Schedule, error) {
	now := time.Now().UTC()

	return &Schedule{
		id:           NewScheduleID(),
		scheduledAt:  scheduledAt,
		service:      service,
		environment:  environment,
		description:  description,
		owner:        owner,
		status:       StatusCreated, // New schedules start with created status
		rollbackPlan: rollbackPlan,
		createdAt:    now,
		updatedAt:    now,
	}, nil
}

// Reconstitute recreates a schedule from storage
func Reconstitute(
	id ScheduleID,
	scheduledAt ScheduledTime,
	service ServiceName,
	environment Environment,
	description Description,
	owner Owner,
	status Status,
	rollbackPlan RollbackPlan,
	createdAt time.Time,
	updatedAt time.Time,
) *Schedule {
	return &Schedule{
		id:           id,
		scheduledAt:  scheduledAt,
		service:      service,
		environment:  environment,
		description:  description,
		owner:        owner,
		status:       status,
		rollbackPlan: rollbackPlan,
		createdAt:    createdAt,
		updatedAt:    updatedAt,
	}
}

// Update updates the schedule fields (owner cannot be changed)
func (s *Schedule) Update(
	scheduledAt *ScheduledTime,
	service *ServiceName,
	environment *Environment,
	description *Description,
	rollbackPlan *RollbackPlan,
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
	if rollbackPlan != nil {
		s.rollbackPlan = *rollbackPlan
	}
	s.updatedAt = time.Now().UTC()
	return nil
}

// Approve changes the schedule status to approved
func (s *Schedule) Approve() error {
	if err := s.status.CanTransitionTo(StatusApproved); err != nil {
		return err
	}
	s.status = StatusApproved
	s.updatedAt = time.Now().UTC()
	return nil
}

// Deny changes the schedule status to denied
func (s *Schedule) Deny() error {
	if err := s.status.CanTransitionTo(StatusDenied); err != nil {
		return err
	}
	s.status = StatusDenied
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

func (s *Schedule) Owner() Owner {
	return s.owner
}

func (s *Schedule) Status() Status {
	return s.status
}

func (s *Schedule) RollbackPlan() RollbackPlan {
	return s.rollbackPlan
}
