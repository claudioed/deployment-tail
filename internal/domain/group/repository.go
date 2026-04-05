package group

import (
	"context"

	"github.com/claudioed/deployment-tail/internal/domain/schedule"
	"github.com/claudioed/deployment-tail/internal/domain/user"
)

// Repository defines the interface for group persistence
type Repository interface {
	// Create saves a new group
	Create(ctx context.Context, group *Group) error

	// FindAll retrieves all groups for a given owner
	FindAll(ctx context.Context, owner schedule.Owner) ([]*Group, error)

	// FindByID retrieves a group by its ID
	FindByID(ctx context.Context, id GroupID) (*Group, error)

	// Update updates an existing group
	Update(ctx context.Context, group *Group) error

	// Delete removes a group
	Delete(ctx context.Context, id GroupID) error

	// AddSchedule assigns a schedule to a group
	AddSchedule(ctx context.Context, groupID GroupID, scheduleID schedule.ScheduleID, assignedBy string) error

	// RemoveSchedule unassigns a schedule from a group
	RemoveSchedule(ctx context.Context, groupID GroupID, scheduleID schedule.ScheduleID) error

	// GetSchedulesInGroup retrieves all schedule IDs in a group
	GetSchedulesInGroup(ctx context.Context, groupID GroupID) ([]schedule.ScheduleID, error)

	// GetGroupsForSchedule retrieves all groups that a schedule belongs to
	GetGroupsForSchedule(ctx context.Context, scheduleID schedule.ScheduleID) ([]*Group, error)

	// FavoriteGroup marks a group as favorite for a user
	FavoriteGroup(ctx context.Context, userID user.UserID, groupID GroupID) error

	// UnfavoriteGroup removes favorite status for a group
	UnfavoriteGroup(ctx context.Context, userID user.UserID, groupID GroupID) error

	// IsFavorite checks if a group is favorited by a user
	IsFavorite(ctx context.Context, userID user.UserID, groupID GroupID) (bool, error)

	// FindAllWithFavorites retrieves all groups for a given owner with favorite status for a user
	FindAllWithFavorites(ctx context.Context, userID user.UserID, owner schedule.Owner) ([]*Group, map[GroupID]bool, error)
}
