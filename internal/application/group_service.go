package application

import (
	"context"
	"fmt"

	"github.com/claudioed/deployment-tail/internal/application/ports/input"
	"github.com/claudioed/deployment-tail/internal/domain/group"
	"github.com/claudioed/deployment-tail/internal/domain/schedule"
	"github.com/claudioed/deployment-tail/internal/domain/user"
)

// GroupService implements the group use cases
type GroupService struct {
	groupRepo    group.Repository
	scheduleRepo schedule.Repository
}

// NewGroupService creates a new group service
func NewGroupService(groupRepo group.Repository, scheduleRepo schedule.Repository) *GroupService {
	return &GroupService{
		groupRepo:    groupRepo,
		scheduleRepo: scheduleRepo,
	}
}

// CreateGroup creates a new group
func (s *GroupService) CreateGroup(ctx context.Context, cmd input.CreateGroupCommand) (*group.Group, error) {
	// Create value objects with validation
	name, err := group.NewGroupName(cmd.Name)
	if err != nil {
		return nil, fmt.Errorf("invalid group name: %w", err)
	}

	description, err := group.NewDescription(cmd.Description)
	if err != nil {
		return nil, fmt.Errorf("invalid description: %w", err)
	}

	// Default visibility to private if not specified
	visibilityStr := cmd.Visibility
	if visibilityStr == "" {
		visibilityStr = "private"
	}

	visibility, err := group.NewVisibility(visibilityStr)
	if err != nil {
		return nil, fmt.Errorf("invalid visibility: %w", err)
	}

	owner, err := schedule.NewOwner(cmd.Owner)
	if err != nil {
		return nil, fmt.Errorf("invalid owner: %w", err)
	}

	// Create the group aggregate
	grp, err := group.NewGroup(name, description, visibility, owner)
	if err != nil {
		return nil, fmt.Errorf("failed to create group: %w", err)
	}

	// Persist
	if err := s.groupRepo.Create(ctx, grp); err != nil {
		return nil, fmt.Errorf("failed to save group: %w", err)
	}

	return grp, nil
}

// GetGroup retrieves a group by ID
func (s *GroupService) GetGroup(ctx context.Context, id string) (*group.Group, error) {
	groupID, err := group.ParseGroupID(id)
	if err != nil {
		return nil, fmt.Errorf("invalid group ID: %w", err)
	}

	grp, err := s.groupRepo.FindByID(ctx, groupID)
	if err != nil {
		return nil, err
	}

	return grp, nil
}

// ListGroups retrieves all groups for an owner
func (s *GroupService) ListGroups(ctx context.Context, query input.ListGroupsQuery) ([]*group.Group, error) {
	owner, err := schedule.NewOwner(query.Owner)
	if err != nil {
		return nil, fmt.Errorf("invalid owner: %w", err)
	}

	groups, err := s.groupRepo.FindAll(ctx, owner)
	if err != nil {
		return nil, fmt.Errorf("failed to list groups: %w", err)
	}

	return groups, nil
}

// ListGroupsWithFavorites retrieves all groups for an owner with favorite status for a user
func (s *GroupService) ListGroupsWithFavorites(ctx context.Context, query input.ListGroupsQuery, userIDStr string) ([]*group.Group, map[string]bool, error) {
	owner, err := schedule.NewOwner(query.Owner)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid owner: %w", err)
	}

	userID, err := user.ParseUserID(userIDStr)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid user ID: %w", err)
	}

	groups, favorites, err := s.groupRepo.FindAllWithFavorites(ctx, userID, owner)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list groups with favorites: %w", err)
	}

	// Convert map keys from GroupID to string
	favoritesStr := make(map[string]bool)
	for groupID, isFav := range favorites {
		favoritesStr[groupID.String()] = isFav
	}

	return groups, favoritesStr, nil
}

// UpdateGroup updates an existing group
func (s *GroupService) UpdateGroup(ctx context.Context, cmd input.UpdateGroupCommand) (*group.Group, error) {
	// Get existing group
	groupID, err := group.ParseGroupID(cmd.ID)
	if err != nil {
		return nil, fmt.Errorf("invalid group ID: %w", err)
	}

	grp, err := s.groupRepo.FindByID(ctx, groupID)
	if err != nil {
		return nil, err
	}

	// Update name
	name, err := group.NewGroupName(cmd.Name)
	if err != nil {
		return nil, fmt.Errorf("invalid group name: %w", err)
	}

	if err := grp.Rename(name); err != nil {
		return nil, fmt.Errorf("failed to rename group: %w", err)
	}

	// Update description
	description, err := group.NewDescription(cmd.Description)
	if err != nil {
		return nil, fmt.Errorf("invalid description: %w", err)
	}

	if err := grp.UpdateDescription(description); err != nil {
		return nil, fmt.Errorf("failed to update description: %w", err)
	}

	// Update visibility if provided
	if cmd.Visibility != "" {
		visibility, err := group.NewVisibility(cmd.Visibility)
		if err != nil {
			return nil, fmt.Errorf("invalid visibility: %w", err)
		}

		if err := grp.SetVisibility(visibility); err != nil {
			return nil, fmt.Errorf("failed to update visibility: %w", err)
		}
	}

	// Persist
	if err := s.groupRepo.Update(ctx, grp); err != nil {
		return nil, fmt.Errorf("failed to save updated group: %w", err)
	}

	return grp, nil
}

// DeleteGroup deletes a group
func (s *GroupService) DeleteGroup(ctx context.Context, cmd input.DeleteGroupCommand) error {
	groupID, err := group.ParseGroupID(cmd.ID)
	if err != nil {
		return fmt.Errorf("invalid group ID: %w", err)
	}

	if err := s.groupRepo.Delete(ctx, groupID); err != nil {
		return fmt.Errorf("failed to delete group: %w", err)
	}

	return nil
}

// AssignScheduleToGroups assigns a schedule to multiple groups
func (s *GroupService) AssignScheduleToGroups(ctx context.Context, cmd input.AssignScheduleCommand) error {
	// Validate schedule ID
	scheduleID, err := schedule.ParseScheduleID(cmd.ScheduleID)
	if err != nil {
		return fmt.Errorf("invalid schedule ID: %w", err)
	}

	// Verify schedule exists
	if _, err := s.scheduleRepo.FindByID(ctx, scheduleID); err != nil {
		return fmt.Errorf("schedule not found: %w", err)
	}

	// Assign to each group
	for _, groupIDStr := range cmd.GroupIDs {
		groupID, err := group.ParseGroupID(groupIDStr)
		if err != nil {
			return fmt.Errorf("invalid group ID %s: %w", groupIDStr, err)
		}

		// Verify group exists
		if _, err := s.groupRepo.FindByID(ctx, groupID); err != nil {
			return fmt.Errorf("group %s not found: %w", groupIDStr, err)
		}

		// Assign schedule to group
		if err := s.groupRepo.AddSchedule(ctx, groupID, scheduleID, cmd.AssignedBy); err != nil {
			return fmt.Errorf("failed to assign schedule to group %s: %w", groupIDStr, err)
		}
	}

	return nil
}

// UnassignScheduleFromGroup unassigns a schedule from a group
func (s *GroupService) UnassignScheduleFromGroup(ctx context.Context, cmd input.UnassignScheduleCommand) error {
	scheduleID, err := schedule.ParseScheduleID(cmd.ScheduleID)
	if err != nil {
		return fmt.Errorf("invalid schedule ID: %w", err)
	}

	groupID, err := group.ParseGroupID(cmd.GroupID)
	if err != nil {
		return fmt.Errorf("invalid group ID: %w", err)
	}

	if err := s.groupRepo.RemoveSchedule(ctx, groupID, scheduleID); err != nil {
		return fmt.Errorf("failed to unassign schedule from group: %w", err)
	}

	return nil
}

// BulkAssignSchedules bulk assigns schedules to a group
func (s *GroupService) BulkAssignSchedules(ctx context.Context, cmd input.BulkAssignCommand) error {
	// Validate group ID
	groupID, err := group.ParseGroupID(cmd.GroupID)
	if err != nil {
		return fmt.Errorf("invalid group ID: %w", err)
	}

	// Verify group exists
	if _, err := s.groupRepo.FindByID(ctx, groupID); err != nil {
		return fmt.Errorf("group not found: %w", err)
	}

	// Assign each schedule to the group
	for _, scheduleIDStr := range cmd.ScheduleIDs {
		scheduleID, err := schedule.ParseScheduleID(scheduleIDStr)
		if err != nil {
			return fmt.Errorf("invalid schedule ID %s: %w", scheduleIDStr, err)
		}

		// Verify schedule exists
		if _, err := s.scheduleRepo.FindByID(ctx, scheduleID); err != nil {
			return fmt.Errorf("schedule %s not found: %w", scheduleIDStr, err)
		}

		// Assign schedule to group
		if err := s.groupRepo.AddSchedule(ctx, groupID, scheduleID, cmd.AssignedBy); err != nil {
			return fmt.Errorf("failed to assign schedule %s to group: %w", scheduleIDStr, err)
		}
	}

	return nil
}

// BulkUnassignSchedules bulk unassigns schedules from a group
func (s *GroupService) BulkUnassignSchedules(ctx context.Context, cmd input.BulkUnassignCommand) error {
	groupID, err := group.ParseGroupID(cmd.GroupID)
	if err != nil {
		return fmt.Errorf("invalid group ID: %w", err)
	}

	// Unassign each schedule from the group
	for _, scheduleIDStr := range cmd.ScheduleIDs {
		scheduleID, err := schedule.ParseScheduleID(scheduleIDStr)
		if err != nil {
			return fmt.Errorf("invalid schedule ID %s: %w", scheduleIDStr, err)
		}

		if err := s.groupRepo.RemoveSchedule(ctx, groupID, scheduleID); err != nil {
			return fmt.Errorf("failed to unassign schedule %s from group: %w", scheduleIDStr, err)
		}
	}

	return nil
}

// GetGroupsForSchedule retrieves all groups that a schedule belongs to
func (s *GroupService) GetGroupsForSchedule(ctx context.Context, scheduleID string) ([]*group.Group, error) {
	sID, err := schedule.ParseScheduleID(scheduleID)
	if err != nil {
		return nil, fmt.Errorf("invalid schedule ID: %w", err)
	}

	groups, err := s.groupRepo.GetGroupsForSchedule(ctx, sID)
	if err != nil {
		return nil, fmt.Errorf("failed to get groups for schedule: %w", err)
	}

	return groups, nil
}

// GetSchedulesInGroup retrieves all schedule IDs in a group
func (s *GroupService) GetSchedulesInGroup(ctx context.Context, groupID string) ([]string, error) {
	gID, err := group.ParseGroupID(groupID)
	if err != nil {
		return nil, fmt.Errorf("invalid group ID: %w", err)
	}

	scheduleIDs, err := s.groupRepo.GetSchedulesInGroup(ctx, gID)
	if err != nil {
		return nil, fmt.Errorf("failed to get schedules in group: %w", err)
	}

	// Convert to strings
	result := make([]string, len(scheduleIDs))
	for i, id := range scheduleIDs {
		result[i] = id.String()
	}

	return result, nil
}

// FavoriteGroup marks a group as favorite for a user
func (s *GroupService) FavoriteGroup(ctx context.Context, userIDStr string, groupIDStr string) error {
	// Validate user ID
	userID, err := user.ParseUserID(userIDStr)
	if err != nil {
		return fmt.Errorf("invalid user ID: %w", err)
	}

	// Validate and verify group exists
	groupID, err := group.ParseGroupID(groupIDStr)
	if err != nil {
		return fmt.Errorf("invalid group ID: %w", err)
	}

	if _, err := s.groupRepo.FindByID(ctx, groupID); err != nil {
		return fmt.Errorf("group not found: %w", err)
	}

	// Favorite the group
	if err := s.groupRepo.FavoriteGroup(ctx, userID, groupID); err != nil {
		return fmt.Errorf("failed to favorite group: %w", err)
	}

	return nil
}

// UnfavoriteGroup removes favorite status for a group
func (s *GroupService) UnfavoriteGroup(ctx context.Context, userIDStr string, groupIDStr string) error {
	// Validate user ID
	userID, err := user.ParseUserID(userIDStr)
	if err != nil {
		return fmt.Errorf("invalid user ID: %w", err)
	}

	// Validate and verify group exists
	groupID, err := group.ParseGroupID(groupIDStr)
	if err != nil {
		return fmt.Errorf("invalid group ID: %w", err)
	}

	if _, err := s.groupRepo.FindByID(ctx, groupID); err != nil {
		return fmt.Errorf("group not found: %w", err)
	}

	// Unfavorite the group
	if err := s.groupRepo.UnfavoriteGroup(ctx, userID, groupID); err != nil {
		return fmt.Errorf("failed to unfavorite group: %w", err)
	}

	return nil
}
