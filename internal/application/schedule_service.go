package application

import (
	"context"
	"fmt"

	"github.com/claudioed/deployment-tail/internal/application/ports/input"
	"github.com/claudioed/deployment-tail/internal/domain/schedule"
)

// ScheduleService implements the schedule use cases
type ScheduleService struct {
	repo schedule.Repository
}

// NewScheduleService creates a new schedule service
func NewScheduleService(repo schedule.Repository) *ScheduleService {
	return &ScheduleService{repo: repo}
}

// CreateSchedule creates a new schedule
func (s *ScheduleService) CreateSchedule(ctx context.Context, cmd input.CreateScheduleCommand) (*schedule.Schedule, error) {
	// Create value objects with validation
	scheduledAt, err := schedule.NewScheduledTime(cmd.ScheduledAt)
	if err != nil {
		return nil, fmt.Errorf("invalid scheduled time: %w", err)
	}

	serviceName, err := schedule.NewServiceName(cmd.ServiceName)
	if err != nil {
		return nil, fmt.Errorf("invalid service name: %w", err)
	}

	env, err := schedule.NewEnvironment(cmd.Environment)
	if err != nil {
		return nil, fmt.Errorf("invalid environment: %w", err)
	}

	desc := schedule.NewDescription(cmd.Description)

	owner, err := schedule.NewOwner(cmd.Owner)
	if err != nil {
		return nil, fmt.Errorf("invalid owner: %w", err)
	}

	rollbackPlan, err := schedule.NewRollbackPlan(cmd.RollbackPlan)
	if err != nil {
		return nil, fmt.Errorf("invalid rollback plan: %w", err)
	}

	// Create the schedule aggregate
	sch, err := schedule.NewSchedule(scheduledAt, serviceName, env, desc, owner, rollbackPlan)
	if err != nil {
		return nil, fmt.Errorf("failed to create schedule: %w", err)
	}

	// Persist
	if err := s.repo.Create(ctx, sch); err != nil {
		return nil, fmt.Errorf("failed to save schedule: %w", err)
	}

	return sch, nil
}

// GetSchedule retrieves a schedule by ID
func (s *ScheduleService) GetSchedule(ctx context.Context, id string) (*schedule.Schedule, error) {
	scheduleID, err := schedule.ParseScheduleID(id)
	if err != nil {
		return nil, fmt.Errorf("invalid schedule ID: %w", err)
	}

	sch, err := s.repo.FindByID(ctx, scheduleID)
	if err != nil {
		return nil, err
	}

	return sch, nil
}

// ListSchedules retrieves schedules with optional filters
func (s *ScheduleService) ListSchedules(ctx context.Context, query input.ListSchedulesQuery) ([]*schedule.Schedule, error) {
	filters := schedule.Filters{
		From: query.From,
		To:   query.To,
	}

	// Parse environment if provided
	if query.Environment != nil && *query.Environment != "" {
		env, err := schedule.NewEnvironment(*query.Environment)
		if err != nil {
			return nil, fmt.Errorf("invalid environment filter: %w", err)
		}
		filters.Environment = &env
	}

	// Parse owner if provided
	if query.Owner != nil && *query.Owner != "" {
		owner, err := schedule.NewOwner(*query.Owner)
		if err != nil {
			return nil, fmt.Errorf("invalid owner filter: %w", err)
		}
		filters.Owner = &owner
	}

	// Parse status if provided
	if query.Status != nil && *query.Status != "" {
		status, err := schedule.ParseStatus(*query.Status)
		if err != nil {
			return nil, fmt.Errorf("invalid status filter: %w", err)
		}
		filters.Status = &status
	}

	schedules, err := s.repo.FindAll(ctx, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to list schedules: %w", err)
	}

	return schedules, nil
}

// UpdateSchedule updates an existing schedule
func (s *ScheduleService) UpdateSchedule(ctx context.Context, cmd input.UpdateScheduleCommand) (*schedule.Schedule, error) {
	// Get existing schedule
	scheduleID, err := schedule.ParseScheduleID(cmd.ID)
	if err != nil {
		return nil, fmt.Errorf("invalid schedule ID: %w", err)
	}

	sch, err := s.repo.FindByID(ctx, scheduleID)
	if err != nil {
		return nil, err
	}

	// Build update values
	var scheduledAt *schedule.ScheduledTime
	var serviceName *schedule.ServiceName
	var environment *schedule.Environment
	var description *schedule.Description
	var rollbackPlan *schedule.RollbackPlan

	if cmd.ScheduledAt != nil {
		st, err := schedule.NewScheduledTime(*cmd.ScheduledAt)
		if err != nil {
			return nil, fmt.Errorf("invalid scheduled time: %w", err)
		}
		scheduledAt = &st
	}

	if cmd.ServiceName != nil {
		sn, err := schedule.NewServiceName(*cmd.ServiceName)
		if err != nil {
			return nil, fmt.Errorf("invalid service name: %w", err)
		}
		serviceName = &sn
	}

	if cmd.Environment != nil {
		env, err := schedule.NewEnvironment(*cmd.Environment)
		if err != nil {
			return nil, fmt.Errorf("invalid environment: %w", err)
		}
		environment = &env
	}

	if cmd.Description != nil {
		desc := schedule.NewDescription(*cmd.Description)
		description = &desc
	}

	if cmd.RollbackPlan != nil {
		rp, err := schedule.NewRollbackPlan(*cmd.RollbackPlan)
		if err != nil {
			return nil, fmt.Errorf("invalid rollback plan: %w", err)
		}
		rollbackPlan = &rp
	}

	// Update the schedule
	if err := sch.Update(scheduledAt, serviceName, environment, description, rollbackPlan); err != nil {
		return nil, fmt.Errorf("failed to update schedule: %w", err)
	}

	// Persist
	if err := s.repo.Update(ctx, sch); err != nil {
		return nil, fmt.Errorf("failed to save updated schedule: %w", err)
	}

	return sch, nil
}

// DeleteSchedule deletes a schedule
func (s *ScheduleService) DeleteSchedule(ctx context.Context, id string) error {
	scheduleID, err := schedule.ParseScheduleID(id)
	if err != nil {
		return fmt.Errorf("invalid schedule ID: %w", err)
	}

	if err := s.repo.Delete(ctx, scheduleID); err != nil {
		return fmt.Errorf("failed to delete schedule: %w", err)
	}

	return nil
}

// ApproveSchedule approves a schedule
func (s *ScheduleService) ApproveSchedule(ctx context.Context, cmd input.ApproveScheduleCommand) (*schedule.Schedule, error) {
	scheduleID, err := schedule.ParseScheduleID(cmd.ID)
	if err != nil {
		return nil, fmt.Errorf("invalid schedule ID: %w", err)
	}

	sch, err := s.repo.FindByID(ctx, scheduleID)
	if err != nil {
		return nil, err
	}

	if err := sch.Approve(); err != nil {
		return nil, fmt.Errorf("failed to approve schedule: %w", err)
	}

	if err := s.repo.Update(ctx, sch); err != nil {
		return nil, fmt.Errorf("failed to save approved schedule: %w", err)
	}

	return sch, nil
}

// DenySchedule denies a schedule
func (s *ScheduleService) DenySchedule(ctx context.Context, cmd input.DenyScheduleCommand) (*schedule.Schedule, error) {
	scheduleID, err := schedule.ParseScheduleID(cmd.ID)
	if err != nil {
		return nil, fmt.Errorf("invalid schedule ID: %w", err)
	}

	sch, err := s.repo.FindByID(ctx, scheduleID)
	if err != nil {
		return nil, err
	}

	if err := sch.Deny(); err != nil {
		return nil, fmt.Errorf("failed to deny schedule: %w", err)
	}

	if err := s.repo.Update(ctx, sch); err != nil {
		return nil, fmt.Errorf("failed to save denied schedule: %w", err)
	}

	return sch, nil
}
