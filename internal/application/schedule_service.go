package application

import (
	"context"
	"fmt"

	"github.com/claudioed/deployment-tail/internal/application/ports/input"
	"github.com/claudioed/deployment-tail/internal/domain/schedule"
	"github.com/claudioed/deployment-tail/internal/domain/user"
)

// ScheduleService implements the schedule use cases
type ScheduleService struct {
	repo     schedule.Repository
	userRepo user.Repository
}

// NewScheduleService creates a new schedule service
func NewScheduleService(repo schedule.Repository, userRepo user.Repository) *ScheduleService {
	return &ScheduleService{
		repo:     repo,
		userRepo: userRepo,
	}
}

// CreateSchedule creates a new schedule
func (s *ScheduleService) CreateSchedule(ctx context.Context, cmd input.CreateScheduleCommand, authenticatedUserID user.UserID) (*schedule.Schedule, error) {
	// Verify user has permission to create schedules
	authenticatedUser, err := s.userRepo.FindByID(ctx, authenticatedUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to find authenticated user: %w", err)
	}

	if !authenticatedUser.CanCreateSchedule() {
		return nil, user.ErrUnauthorized{
			UserID:    authenticatedUserID.String(),
			Operation: "create schedule",
			Reason:    "requires deployer or admin role",
		}
	}

	// Create value objects with validation
	scheduledAt, err := schedule.NewScheduledTime(cmd.ScheduledAt)
	if err != nil {
		return nil, fmt.Errorf("invalid scheduled time: %w", err)
	}

	serviceName, err := schedule.NewServiceName(cmd.ServiceName)
	if err != nil {
		return nil, fmt.Errorf("invalid service name: %w", err)
	}

	// Parse environments
	environments := []schedule.Environment{}
	for _, envStr := range cmd.Environments {
		env, err := schedule.NewEnvironment(envStr)
		if err != nil {
			return nil, fmt.Errorf("invalid environment %q: %w", envStr, err)
		}
		environments = append(environments, env)
	}

	desc := schedule.NewDescription(cmd.Description)

	// Parse owners
	owners := []schedule.Owner{}
	for _, ownerStr := range cmd.Owners {
		owner, err := schedule.NewOwner(ownerStr)
		if err != nil {
			return nil, fmt.Errorf("invalid owner %q: %w", ownerStr, err)
		}
		owners = append(owners, owner)
	}

	rollbackPlan, err := schedule.NewRollbackPlan(cmd.RollbackPlan)
	if err != nil {
		return nil, fmt.Errorf("invalid rollback plan: %w", err)
	}

	// Create the schedule aggregate with createdBy audit field
	sch, err := schedule.NewSchedule(scheduledAt, serviceName, environments, desc, owners, rollbackPlan, authenticatedUserID)
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

	// Parse environments if provided
	if len(query.Environments) > 0 {
		environments := []schedule.Environment{}
		for _, envStr := range query.Environments {
			if envStr != "" {
				env, err := schedule.NewEnvironment(envStr)
				if err != nil {
					return nil, fmt.Errorf("invalid environment filter %q: %w", envStr, err)
				}
				environments = append(environments, env)
			}
		}
		filters.Environments = environments
	}

	// Parse owners if provided
	if len(query.Owners) > 0 {
		owners := []schedule.Owner{}
		for _, ownerStr := range query.Owners {
			if ownerStr != "" {
				owner, err := schedule.NewOwner(ownerStr)
				if err != nil {
					return nil, fmt.Errorf("invalid owner filter %q: %w", ownerStr, err)
				}
				owners = append(owners, owner)
			}
		}
		filters.Owners = owners
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
func (s *ScheduleService) UpdateSchedule(ctx context.Context, cmd input.UpdateScheduleCommand, authenticatedUserID user.UserID) (*schedule.Schedule, error) {
	// Get authenticated user
	authenticatedUser, err := s.userRepo.FindByID(ctx, authenticatedUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to find authenticated user: %w", err)
	}

	// Get existing schedule
	scheduleID, err := schedule.ParseScheduleID(cmd.ID)
	if err != nil {
		return nil, fmt.Errorf("invalid schedule ID: %w", err)
	}

	sch, err := s.repo.FindByID(ctx, scheduleID)
	if err != nil {
		return nil, err
	}

	// Check ownership: deployer can only update own schedules, admin can update any
	if !authenticatedUser.CanModifySchedule(sch.CreatedBy()) {
		return nil, user.ErrUnauthorized{
			UserID:    authenticatedUserID.String(),
			Operation: "update schedule",
			Reason:    "deployers can only update their own schedules",
		}
	}

	// Build update values
	var scheduledAt *schedule.ScheduledTime
	var serviceName *schedule.ServiceName
	var environments *[]schedule.Environment
	var owners *[]schedule.Owner
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

	if cmd.Environments != nil {
		envs := []schedule.Environment{}
		for _, envStr := range *cmd.Environments {
			env, err := schedule.NewEnvironment(envStr)
			if err != nil {
				return nil, fmt.Errorf("invalid environment %q: %w", envStr, err)
			}
			envs = append(envs, env)
		}
		environments = &envs
	}

	if cmd.Owners != nil {
		ownrs := []schedule.Owner{}
		for _, ownerStr := range *cmd.Owners {
			owner, err := schedule.NewOwner(ownerStr)
			if err != nil {
				return nil, fmt.Errorf("invalid owner %q: %w", ownerStr, err)
			}
			ownrs = append(ownrs, owner)
		}
		owners = &ownrs
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

	// Update the schedule with updatedBy audit field
	if err := sch.Update(scheduledAt, serviceName, environments, description, owners, rollbackPlan, authenticatedUserID); err != nil {
		return nil, fmt.Errorf("failed to update schedule: %w", err)
	}

	// Persist
	if err := s.repo.Update(ctx, sch); err != nil {
		return nil, fmt.Errorf("failed to save updated schedule: %w", err)
	}

	return sch, nil
}

// DeleteSchedule deletes a schedule
func (s *ScheduleService) DeleteSchedule(ctx context.Context, id string, authenticatedUserID user.UserID) error {
	// Get authenticated user
	authenticatedUser, err := s.userRepo.FindByID(ctx, authenticatedUserID)
	if err != nil {
		return fmt.Errorf("failed to find authenticated user: %w", err)
	}

	scheduleID, err := schedule.ParseScheduleID(id)
	if err != nil {
		return fmt.Errorf("invalid schedule ID: %w", err)
	}

	// Get schedule to check ownership
	sch, err := s.repo.FindByID(ctx, scheduleID)
	if err != nil {
		return err
	}

	// Check ownership: deployer can only delete own schedules, admin can delete any
	if !authenticatedUser.CanModifySchedule(sch.CreatedBy()) {
		return user.ErrUnauthorized{
			UserID:    authenticatedUserID.String(),
			Operation: "delete schedule",
			Reason:    "deployers can only delete their own schedules",
		}
	}

	// Soft delete with deletedBy audit field
	if err := s.repo.Delete(ctx, scheduleID, authenticatedUserID); err != nil {
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
