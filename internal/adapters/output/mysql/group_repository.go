package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/claudioed/deployment-tail/internal/domain/group"
	"github.com/claudioed/deployment-tail/internal/domain/schedule"
	"github.com/claudioed/deployment-tail/internal/domain/user"
	"github.com/go-sql-driver/mysql"
)

// GroupRepository implements the group.Repository interface for MySQL
type GroupRepository struct {
	db *sql.DB
}

// NewGroupRepository creates a new MySQL group repository
func NewGroupRepository(db *sql.DB) *GroupRepository {
	return &GroupRepository{db: db}
}

// Create saves a new group
func (r *GroupRepository) Create(ctx context.Context, grp *group.Group) error {
	query := `
		INSERT INTO `+"`groups`"+` (id, name, description, owner, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.ExecContext(ctx, query,
		grp.ID().String(),
		grp.Name().String(),
		grp.Description().String(),
		grp.Owner().String(),
		grp.CreatedAt(),
		grp.UpdatedAt(),
	)

	if err != nil {
		// Check for duplicate name constraint violation
		if mysqlErr, ok := err.(*mysql.MySQLError); ok && mysqlErr.Number == 1062 {
			return group.ErrDuplicateGroupName
		}
		return fmt.Errorf("failed to create group: %w", err)
	}

	return nil
}

// FindAll retrieves all groups for a given owner
func (r *GroupRepository) FindAll(ctx context.Context, owner schedule.Owner) ([]*group.Group, error) {
	query := `
		SELECT id, name, description, owner, created_at, updated_at
		FROM `+"`groups`"+`
		WHERE owner = ?
		ORDER BY name ASC
	`

	rows, err := r.db.QueryContext(ctx, query, owner.String())
	if err != nil {
		return nil, fmt.Errorf("failed to query groups: %w", err)
	}
	defer rows.Close()

	var groups []*group.Group

	for rows.Next() {
		var (
			idStr       string
			name        string
			description string
			ownerStr    string
			createdAt   time.Time
			updatedAt   time.Time
		)

		if err := rows.Scan(&idStr, &name, &description, &ownerStr, &createdAt, &updatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan group: %w", err)
		}

		grp, err := r.mapToGroup(idStr, name, description, ownerStr, createdAt, updatedAt)
		if err != nil {
			return nil, err
		}

		groups = append(groups, grp)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating groups: %w", err)
	}

	return groups, nil
}

// FindByID retrieves a group by its ID
func (r *GroupRepository) FindByID(ctx context.Context, id group.GroupID) (*group.Group, error) {
	query := `
		SELECT id, name, description, owner, created_at, updated_at
		FROM `+"`groups`"+`
		WHERE id = ?
	`

	var (
		idStr       string
		name        string
		description string
		ownerStr    string
		createdAt   time.Time
		updatedAt   time.Time
	)

	err := r.db.QueryRowContext(ctx, query, id.String()).Scan(
		&idStr,
		&name,
		&description,
		&ownerStr,
		&createdAt,
		&updatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, group.ErrGroupNotFound
		}
		return nil, fmt.Errorf("failed to find group: %w", err)
	}

	return r.mapToGroup(idStr, name, description, ownerStr, createdAt, updatedAt)
}

// Update updates an existing group
func (r *GroupRepository) Update(ctx context.Context, grp *group.Group) error {
	query := `
		UPDATE `+"`groups`"+`
		SET name = ?, description = ?, updated_at = ?
		WHERE id = ?
	`

	result, err := r.db.ExecContext(ctx, query,
		grp.Name().String(),
		grp.Description().String(),
		grp.UpdatedAt(),
		grp.ID().String(),
	)

	if err != nil {
		// Check for duplicate name constraint violation
		if mysqlErr, ok := err.(*mysql.MySQLError); ok && mysqlErr.Number == 1062 {
			return group.ErrDuplicateGroupName
		}
		return fmt.Errorf("failed to update group: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return group.ErrGroupNotFound
	}

	return nil
}

// Delete removes a group
func (r *GroupRepository) Delete(ctx context.Context, id group.GroupID) error {
	query := "DELETE FROM `groups` WHERE id = ?"

	result, err := r.db.ExecContext(ctx, query, id.String())
	if err != nil {
		return fmt.Errorf("failed to delete group: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return group.ErrGroupNotFound
	}

	return nil
}

// AddSchedule assigns a schedule to a group
func (r *GroupRepository) AddSchedule(ctx context.Context, groupID group.GroupID, scheduleID schedule.ScheduleID, assignedBy string) error {
	query := `
		INSERT INTO schedule_groups (schedule_id, group_id, assigned_at, assigned_by)
		VALUES (?, ?, ?, ?)
	`

	_, err := r.db.ExecContext(ctx, query,
		scheduleID.String(),
		groupID.String(),
		time.Now().UTC(),
		assignedBy,
	)

	if err != nil {
		// Check for duplicate entry (already assigned)
		if mysqlErr, ok := err.(*mysql.MySQLError); ok && mysqlErr.Number == 1062 {
			// Already assigned, not an error
			return nil
		}
		return fmt.Errorf("failed to add schedule to group: %w", err)
	}

	return nil
}

// RemoveSchedule unassigns a schedule from a group
func (r *GroupRepository) RemoveSchedule(ctx context.Context, groupID group.GroupID, scheduleID schedule.ScheduleID) error {
	query := "DELETE FROM schedule_groups WHERE group_id = ? AND schedule_id = ?"

	result, err := r.db.ExecContext(ctx, query, groupID.String(), scheduleID.String())
	if err != nil {
		return fmt.Errorf("failed to remove schedule from group: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		// Not assigned, not an error
		return nil
	}

	return nil
}

// GetSchedulesInGroup retrieves all schedule IDs in a group
func (r *GroupRepository) GetSchedulesInGroup(ctx context.Context, groupID group.GroupID) ([]schedule.ScheduleID, error) {
	query := `
		SELECT schedule_id
		FROM schedule_groups
		WHERE group_id = ?
		ORDER BY assigned_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, groupID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to query schedules in group: %w", err)
	}
	defer rows.Close()

	var scheduleIDs []schedule.ScheduleID

	for rows.Next() {
		var scheduleIDStr string
		if err := rows.Scan(&scheduleIDStr); err != nil {
			return nil, fmt.Errorf("failed to scan schedule ID: %w", err)
		}

		scheduleID, err := schedule.ParseScheduleID(scheduleIDStr)
		if err != nil {
			return nil, fmt.Errorf("invalid schedule ID in database: %w", err)
		}

		scheduleIDs = append(scheduleIDs, scheduleID)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating schedule IDs: %w", err)
	}

	return scheduleIDs, nil
}

// GetGroupsForSchedule retrieves all groups that a schedule belongs to
func (r *GroupRepository) GetGroupsForSchedule(ctx context.Context, scheduleID schedule.ScheduleID) ([]*group.Group, error) {
	query := `
		SELECT g.id, g.name, g.description, g.owner, g.created_at, g.updated_at
		FROM `+"`groups`"+` g
		INNER JOIN schedule_groups sg ON g.id = sg.group_id
		WHERE sg.schedule_id = ?
		ORDER BY g.name ASC
	`

	rows, err := r.db.QueryContext(ctx, query, scheduleID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to query groups for schedule: %w", err)
	}
	defer rows.Close()

	var groups []*group.Group

	for rows.Next() {
		var (
			idStr       string
			name        string
			description string
			ownerStr    string
			createdAt   time.Time
			updatedAt   time.Time
		)

		if err := rows.Scan(&idStr, &name, &description, &ownerStr, &createdAt, &updatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan group: %w", err)
		}

		grp, err := r.mapToGroup(idStr, name, description, ownerStr, createdAt, updatedAt)
		if err != nil {
			return nil, err
		}

		groups = append(groups, grp)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating groups: %w", err)
	}

	return groups, nil
}

// mapToGroup converts database row to domain group
func (r *GroupRepository) mapToGroup(
	idStr string,
	name string,
	description string,
	ownerStr string,
	createdAt time.Time,
	updatedAt time.Time,
) (*group.Group, error) {
	id, err := group.ParseGroupID(idStr)
	if err != nil {
		return nil, fmt.Errorf("invalid group ID in database: %w", err)
	}

	groupName, err := group.NewGroupName(name)
	if err != nil {
		return nil, fmt.Errorf("invalid group name in database: %w", err)
	}

	desc, err := group.NewDescription(description)
	if err != nil {
		return nil, fmt.Errorf("invalid description in database: %w", err)
	}

	owner, err := schedule.NewOwner(strings.TrimSpace(ownerStr))
	if err != nil {
		return nil, fmt.Errorf("invalid owner in database: %w", err)
	}

	return group.Reconstitute(id, groupName, desc, owner, createdAt, updatedAt), nil
}

// FavoriteGroup marks a group as favorite for a user
func (r *GroupRepository) FavoriteGroup(ctx context.Context, userID user.UserID, groupID group.GroupID) error {
	query := `
		INSERT INTO group_favorites (user_id, group_id, created_at)
		VALUES (?, ?, ?)
	`

	_, err := r.db.ExecContext(ctx, query,
		userID.String(),
		groupID.String(),
		time.Now().UTC(),
	)

	if err != nil {
		// Check for duplicate entry (already favorited)
		if mysqlErr, ok := err.(*mysql.MySQLError); ok && mysqlErr.Number == 1062 {
			// Already favorited, not an error (idempotent)
			return nil
		}
		return fmt.Errorf("failed to favorite group: %w", err)
	}

	return nil
}

// UnfavoriteGroup removes favorite status for a group
func (r *GroupRepository) UnfavoriteGroup(ctx context.Context, userID user.UserID, groupID group.GroupID) error {
	query := "DELETE FROM group_favorites WHERE user_id = ? AND group_id = ?"

	result, err := r.db.ExecContext(ctx, query, userID.String(), groupID.String())
	if err != nil {
		return fmt.Errorf("failed to unfavorite group: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		// Not favorited, not an error (idempotent)
		return nil
	}

	return nil
}

// IsFavorite checks if a group is favorited by a user
func (r *GroupRepository) IsFavorite(ctx context.Context, userID user.UserID, groupID group.GroupID) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM group_favorites
			WHERE user_id = ? AND group_id = ?
		)
	`

	var exists bool
	err := r.db.QueryRowContext(ctx, query, userID.String(), groupID.String()).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check if group is favorite: %w", err)
	}

	return exists, nil
}

// FindAllWithFavorites retrieves all groups for a given owner with favorite status for a user
func (r *GroupRepository) FindAllWithFavorites(ctx context.Context, userID user.UserID, owner schedule.Owner) ([]*group.Group, map[group.GroupID]bool, error) {
	query := `
		SELECT g.id, g.name, g.description, g.owner, g.created_at, g.updated_at,
		       COALESCE(gf.user_id IS NOT NULL, FALSE) AS is_favorite
		FROM ` + "`groups`" + ` g
		LEFT JOIN group_favorites gf ON g.id = gf.group_id AND gf.user_id = ?
		WHERE g.owner = ?
		ORDER BY is_favorite DESC, g.name ASC
	`

	rows, err := r.db.QueryContext(ctx, query, userID.String(), owner.String())
	if err != nil {
		return nil, nil, fmt.Errorf("failed to query groups with favorites: %w", err)
	}
	defer rows.Close()

	var groups []*group.Group
	favorites := make(map[group.GroupID]bool)

	for rows.Next() {
		var (
			idStr       string
			name        string
			description string
			ownerStr    string
			createdAt   time.Time
			updatedAt   time.Time
			isFavorite  bool
		)

		if err := rows.Scan(&idStr, &name, &description, &ownerStr, &createdAt, &updatedAt, &isFavorite); err != nil {
			return nil, nil, fmt.Errorf("failed to scan group: %w", err)
		}

		grp, err := r.mapToGroup(idStr, name, description, ownerStr, createdAt, updatedAt)
		if err != nil {
			return nil, nil, err
		}

		groups = append(groups, grp)
		favorites[grp.ID()] = isFavorite
	}

	if err := rows.Err(); err != nil {
		return nil, nil, fmt.Errorf("error iterating groups: %w", err)
	}

	return groups, favorites, nil
}
