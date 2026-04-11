package input

import (
	"context"

	"github.com/claudioed/deployment-tail/internal/domain/group"
)

// CreateGroupCommand represents the command to create a group
type CreateGroupCommand struct {
	Name        string
	Description string
	Visibility  string
	Owner       string
}

// UpdateGroupCommand represents the command to update a group
type UpdateGroupCommand struct {
	ID          string
	Name        string
	Description string
	Visibility  string
}

// DeleteGroupCommand represents the command to delete a group
type DeleteGroupCommand struct {
	ID string
}

// AssignScheduleCommand represents the command to assign a schedule to groups
type AssignScheduleCommand struct {
	ScheduleID string
	GroupIDs   []string
	AssignedBy string
}

// UnassignScheduleCommand represents the command to unassign a schedule from a group
type UnassignScheduleCommand struct {
	ScheduleID string
	GroupID    string
}

// BulkAssignCommand represents the command to bulk assign schedules to a group
type BulkAssignCommand struct {
	GroupID     string
	ScheduleIDs []string
	AssignedBy  string
}

// BulkUnassignCommand represents the command to bulk unassign schedules from a group
type BulkUnassignCommand struct {
	GroupID     string
	ScheduleIDs []string
}

// ListGroupsQuery represents the query to list groups
type ListGroupsQuery struct {
	Owner string
}

// GroupService defines the inbound port for group operations
type GroupService interface {
	// CreateGroup creates a new group
	CreateGroup(ctx context.Context, cmd CreateGroupCommand) (*group.Group, error)

	// GetGroup retrieves a group by ID
	GetGroup(ctx context.Context, id string) (*group.Group, error)

	// ListGroups retrieves all groups for an owner
	ListGroups(ctx context.Context, query ListGroupsQuery) ([]*group.Group, error)

	// UpdateGroup updates an existing group
	UpdateGroup(ctx context.Context, cmd UpdateGroupCommand) (*group.Group, error)

	// DeleteGroup deletes a group
	DeleteGroup(ctx context.Context, cmd DeleteGroupCommand) error

	// AssignScheduleToGroups assigns a schedule to multiple groups
	AssignScheduleToGroups(ctx context.Context, cmd AssignScheduleCommand) error

	// UnassignScheduleFromGroup unassigns a schedule from a group
	UnassignScheduleFromGroup(ctx context.Context, cmd UnassignScheduleCommand) error

	// BulkAssignSchedules bulk assigns schedules to a group
	BulkAssignSchedules(ctx context.Context, cmd BulkAssignCommand) error

	// BulkUnassignSchedules bulk unassigns schedules from a group
	BulkUnassignSchedules(ctx context.Context, cmd BulkUnassignCommand) error

	// GetGroupsForSchedule retrieves all groups that a schedule belongs to
	GetGroupsForSchedule(ctx context.Context, scheduleID string) ([]*group.Group, error)

	// GetSchedulesInGroup retrieves all schedule IDs in a group
	GetSchedulesInGroup(ctx context.Context, groupID string) ([]string, error)

	// FavoriteGroup marks a group as favorite for a user
	FavoriteGroup(ctx context.Context, userID string, groupID string) error

	// UnfavoriteGroup removes favorite status for a group
	UnfavoriteGroup(ctx context.Context, userID string, groupID string) error

	// ListGroupsWithFavorites retrieves all groups for an owner with favorite status for a user
	ListGroupsWithFavorites(ctx context.Context, query ListGroupsQuery, userID string) ([]*group.Group, map[string]bool, error)
}
