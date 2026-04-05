package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/claudioed/deployment-tail/internal/domain/schedule"
	"github.com/claudioed/deployment-tail/internal/domain/user"
)

// ScheduleRepository implements the schedule.Repository interface for MySQL
type ScheduleRepository struct {
	db *sql.DB
}

// NewScheduleRepository creates a new MySQL schedule repository
func NewScheduleRepository(db *sql.DB) *ScheduleRepository {
	return &ScheduleRepository{db: db}
}

// Create saves a new schedule
func (r *ScheduleRepository) Create(ctx context.Context, sch *schedule.Schedule) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Insert schedule record
	query := `
		INSERT INTO schedules (id, scheduled_at, service_name, description, status, rollback_plan, created_by, updated_by, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	var rollbackPlan *string
	if !sch.RollbackPlan().IsEmpty() {
		val := sch.RollbackPlan().String()
		rollbackPlan = &val
	}

	_, err = tx.ExecContext(ctx, query,
		sch.ID().String(),
		sch.ScheduledAt().Value(),
		sch.Service().Value(),
		sch.Description().Value(),
		sch.Status().String(),
		rollbackPlan,
		sch.CreatedBy().String(),
		sch.UpdatedBy().String(),
		sch.CreatedAt(),
		sch.UpdatedAt(),
	)

	if err != nil {
		return fmt.Errorf("failed to create schedule: %w", err)
	}

	// Insert owners
	ownerQuery := "INSERT INTO schedule_owners (schedule_id, owner) VALUES (?, ?)"
	for _, owner := range sch.Owners() {
		_, err = tx.ExecContext(ctx, ownerQuery, sch.ID().String(), owner.String())
		if err != nil {
			return fmt.Errorf("failed to insert owner: %w", err)
		}
	}

	// Insert environments
	envQuery := "INSERT INTO schedule_environments (schedule_id, environment) VALUES (?, ?)"
	for _, env := range sch.Environments() {
		_, err = tx.ExecContext(ctx, envQuery, sch.ID().String(), env.String())
		if err != nil {
			return fmt.Errorf("failed to insert environment: %w", err)
		}
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// FindByID retrieves a schedule by its ID
func (r *ScheduleRepository) FindByID(ctx context.Context, id schedule.ScheduleID) (*schedule.Schedule, error) {
	query := `
		SELECT id, scheduled_at, service_name, description, status, rollback_plan, created_by, updated_by, created_at, updated_at
		FROM schedules
		WHERE id = ? AND deleted_at IS NULL
	`

	var (
		idStr        string
		scheduledAt  time.Time
		serviceName  string
		description  string
		status       string
		rollbackPlan sql.NullString
		createdBy    sql.NullString
		updatedBy    sql.NullString
		createdAt    time.Time
		updatedAt    time.Time
	)

	err := r.db.QueryRowContext(ctx, query, id.String()).Scan(
		&idStr,
		&scheduledAt,
		&serviceName,
		&description,
		&status,
		&rollbackPlan,
		&createdBy,
		&updatedBy,
		&createdAt,
		&updatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, schedule.ErrScheduleNotFound
		}
		return nil, fmt.Errorf("failed to find schedule: %w", err)
	}

	// Load owners
	owners, err := r.loadOwners(ctx, id.String())
	if err != nil {
		return nil, fmt.Errorf("failed to load owners: %w", err)
	}

	// Load environments
	environments, err := r.loadEnvironments(ctx, id.String())
	if err != nil {
		return nil, fmt.Errorf("failed to load environments: %w", err)
	}

	rollbackPlanStr := ""
	if rollbackPlan.Valid {
		rollbackPlanStr = rollbackPlan.String
	}

	createdByStr := ""
	if createdBy.Valid {
		createdByStr = createdBy.String
	}

	updatedByStr := ""
	if updatedBy.Valid {
		updatedByStr = updatedBy.String
	}

	return r.mapToSchedule(idStr, scheduledAt, serviceName, description, owners, environments, status, rollbackPlanStr, createdByStr, updatedByStr, createdAt, updatedAt)
}

// FindAll retrieves schedules with optional filters
func (r *ScheduleRepository) FindAll(ctx context.Context, filters schedule.Filters) ([]*schedule.Schedule, error) {
	query := "SELECT DISTINCT s.id, s.scheduled_at, s.service_name, s.description, s.status, s.rollback_plan, s.created_by, s.updated_by, s.created_at, s.updated_at FROM schedules s"
	args := []interface{}{}
	joins := []string{}
	conditions := []string{"1=1", "s.deleted_at IS NULL"}

	// Apply filters
	if filters.From != nil {
		conditions = append(conditions, "s.scheduled_at >= ?")
		args = append(args, *filters.From)
	}

	if filters.To != nil {
		conditions = append(conditions, "s.scheduled_at <= ?")
		args = append(args, *filters.To)
	}

	if len(filters.Environments) > 0 {
		joins = append(joins, "INNER JOIN schedule_environments se ON s.id = se.schedule_id")
		placeholders := make([]string, len(filters.Environments))
		for i, env := range filters.Environments {
			placeholders[i] = "?"
			args = append(args, env.String())
		}
		conditions = append(conditions, "se.environment IN ("+strings.Join(placeholders, ",")+")")
	}

	if len(filters.Owners) > 0 {
		joins = append(joins, "INNER JOIN schedule_owners so ON s.id = so.schedule_id")
		placeholders := make([]string, len(filters.Owners))
		for i, owner := range filters.Owners {
			placeholders[i] = "?"
			args = append(args, owner.String())
		}
		conditions = append(conditions, "so.owner IN ("+strings.Join(placeholders, ",")+")")
	}

	if filters.Status != nil {
		conditions = append(conditions, "s.status = ?")
		args = append(args, filters.Status.String())
	}

	// Build final query
	if len(joins) > 0 {
		query += " " + strings.Join(joins, " ")
	}
	query += " WHERE " + strings.Join(conditions, " AND ")
	query += " ORDER BY s.scheduled_at ASC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query schedules: %w", err)
	}
	defer rows.Close()

	var schedules []*schedule.Schedule

	for rows.Next() {
		var (
			idStr        string
			scheduledAt  time.Time
			serviceName  string
			description  string
			status       string
			rollbackPlan sql.NullString
			createdBy    sql.NullString
			updatedBy    sql.NullString
			createdAt    time.Time
			updatedAt    time.Time
		)

		if err := rows.Scan(&idStr, &scheduledAt, &serviceName, &description, &status, &rollbackPlan, &createdBy, &updatedBy, &createdAt, &updatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan schedule: %w", err)
		}

		// Load owners
		owners, err := r.loadOwners(ctx, idStr)
		if err != nil {
			return nil, fmt.Errorf("failed to load owners: %w", err)
		}

		// Load environments
		environments, err := r.loadEnvironments(ctx, idStr)
		if err != nil {
			return nil, fmt.Errorf("failed to load environments: %w", err)
		}

		rollbackPlanStr := ""
		if rollbackPlan.Valid {
			rollbackPlanStr = rollbackPlan.String
		}

		createdByStr := ""
		if createdBy.Valid {
			createdByStr = createdBy.String
		}

		updatedByStr := ""
		if updatedBy.Valid {
			updatedByStr = updatedBy.String
		}

		sch, err := r.mapToSchedule(idStr, scheduledAt, serviceName, description, owners, environments, status, rollbackPlanStr, createdByStr, updatedByStr, createdAt, updatedAt)
		if err != nil {
			return nil, err
		}

		schedules = append(schedules, sch)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating schedules: %w", err)
	}

	return schedules, nil
}

// Update updates an existing schedule
func (r *ScheduleRepository) Update(ctx context.Context, sch *schedule.Schedule) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Update schedule record
	query := `
		UPDATE schedules
		SET scheduled_at = ?, service_name = ?, description = ?, status = ?, rollback_plan = ?, updated_by = ?, updated_at = ?
		WHERE id = ?
	`

	var rollbackPlan *string
	if !sch.RollbackPlan().IsEmpty() {
		val := sch.RollbackPlan().String()
		rollbackPlan = &val
	}

	result, err := tx.ExecContext(ctx, query,
		sch.ScheduledAt().Value(),
		sch.Service().Value(),
		sch.Description().Value(),
		sch.Status().String(),
		rollbackPlan,
		sch.UpdatedBy().String(),
		sch.UpdatedAt(),
		sch.ID().String(),
	)

	if err != nil {
		return fmt.Errorf("failed to update schedule: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return schedule.ErrScheduleNotFound
	}

	// Delta update for owners: delete all and re-insert
	_, err = tx.ExecContext(ctx, "DELETE FROM schedule_owners WHERE schedule_id = ?", sch.ID().String())
	if err != nil {
		return fmt.Errorf("failed to delete existing owners: %w", err)
	}

	ownerQuery := "INSERT INTO schedule_owners (schedule_id, owner) VALUES (?, ?)"
	for _, owner := range sch.Owners() {
		_, err = tx.ExecContext(ctx, ownerQuery, sch.ID().String(), owner.String())
		if err != nil {
			return fmt.Errorf("failed to insert owner: %w", err)
		}
	}

	// Delta update for environments: delete all and re-insert
	_, err = tx.ExecContext(ctx, "DELETE FROM schedule_environments WHERE schedule_id = ?", sch.ID().String())
	if err != nil {
		return fmt.Errorf("failed to delete existing environments: %w", err)
	}

	envQuery := "INSERT INTO schedule_environments (schedule_id, environment) VALUES (?, ?)"
	for _, env := range sch.Environments() {
		_, err = tx.ExecContext(ctx, envQuery, sch.ID().String(), env.String())
		if err != nil {
			return fmt.Errorf("failed to insert environment: %w", err)
		}
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// Delete removes a schedule
func (r *ScheduleRepository) Delete(ctx context.Context, id schedule.ScheduleID, deletedBy user.UserID) error {
	// Soft delete: set deleted_at and deleted_by
	query := "UPDATE schedules SET deleted_at = ?, deleted_by = ? WHERE id = ? AND deleted_at IS NULL"

	now := time.Now().UTC()
	result, err := r.db.ExecContext(ctx, query, now, deletedBy.String(), id.String())
	if err != nil {
		return fmt.Errorf("failed to delete schedule: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return schedule.ErrScheduleNotFound
	}

	return nil
}

// FindUngrouped retrieves schedules that are not assigned to any group
func (r *ScheduleRepository) FindUngrouped(ctx context.Context, filters schedule.Filters) ([]*schedule.Schedule, error) {
	query := `
		SELECT DISTINCT s.id, s.scheduled_at, s.service_name, s.description, s.status, s.rollback_plan, s.created_by, s.updated_by, s.created_at, s.updated_at
		FROM schedules s
		LEFT JOIN schedule_groups sg ON s.id = sg.schedule_id
		WHERE sg.schedule_id IS NULL AND s.deleted_at IS NULL
	`
	args := []interface{}{}
	additionalJoins := []string{}

	// Apply filters
	if filters.From != nil {
		query += " AND s.scheduled_at >= ?"
		args = append(args, *filters.From)
	}

	if filters.To != nil {
		query += " AND s.scheduled_at <= ?"
		args = append(args, *filters.To)
	}

	if len(filters.Environments) > 0 {
		additionalJoins = append(additionalJoins, "INNER JOIN schedule_environments se ON s.id = se.schedule_id")
		placeholders := make([]string, len(filters.Environments))
		for i, env := range filters.Environments {
			placeholders[i] = "?"
			args = append(args, env.String())
		}
		query += " AND se.environment IN (" + strings.Join(placeholders, ",") + ")"
	}

	if len(filters.Owners) > 0 {
		additionalJoins = append(additionalJoins, "INNER JOIN schedule_owners so ON s.id = so.schedule_id")
		placeholders := make([]string, len(filters.Owners))
		for i, owner := range filters.Owners {
			placeholders[i] = "?"
			args = append(args, owner.String())
		}
		query += " AND so.owner IN (" + strings.Join(placeholders, ",") + ")"
	}

	if filters.Status != nil {
		query += " AND s.status = ?"
		args = append(args, filters.Status.String())
	}

	// Build final query with additional joins if needed
	if len(additionalJoins) > 0 {
		// Rebuild query to include additional joins
		query = `
			SELECT DISTINCT s.id, s.scheduled_at, s.service_name, s.description, s.status, s.rollback_plan, s.created_by, s.updated_by, s.created_at, s.updated_at
			FROM schedules s
			LEFT JOIN schedule_groups sg ON s.id = sg.schedule_id
		` + strings.Join(additionalJoins, " ") + `
			WHERE sg.schedule_id IS NULL AND s.deleted_at IS NULL
		`
		// Re-add filter conditions
		if filters.From != nil {
			query += " AND s.scheduled_at >= ?"
		}
		if filters.To != nil {
			query += " AND s.scheduled_at <= ?"
		}
		if len(filters.Environments) > 0 {
			placeholders := make([]string, len(filters.Environments))
			for i := range filters.Environments {
				placeholders[i] = "?"
			}
			query += " AND se.environment IN (" + strings.Join(placeholders, ",") + ")"
		}
		if len(filters.Owners) > 0 {
			placeholders := make([]string, len(filters.Owners))
			for i := range filters.Owners {
				placeholders[i] = "?"
			}
			query += " AND so.owner IN (" + strings.Join(placeholders, ",") + ")"
		}
		if filters.Status != nil {
			query += " AND s.status = ?"
		}
	}

	query += " ORDER BY s.scheduled_at ASC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query ungrouped schedules: %w", err)
	}
	defer rows.Close()

	var schedules []*schedule.Schedule

	for rows.Next() {
		var (
			idStr        string
			scheduledAt  time.Time
			serviceName  string
			description  string
			status       string
			rollbackPlan sql.NullString
			createdBy    sql.NullString
			updatedBy    sql.NullString
			createdAt    time.Time
			updatedAt    time.Time
		)

		if err := rows.Scan(&idStr, &scheduledAt, &serviceName, &description, &status, &rollbackPlan, &createdBy, &updatedBy, &createdAt, &updatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan schedule: %w", err)
		}

		// Load owners
		owners, err := r.loadOwners(ctx, idStr)
		if err != nil {
			return nil, fmt.Errorf("failed to load owners: %w", err)
		}

		// Load environments
		environments, err := r.loadEnvironments(ctx, idStr)
		if err != nil {
			return nil, fmt.Errorf("failed to load environments: %w", err)
		}

		rollbackPlanStr := ""
		if rollbackPlan.Valid {
			rollbackPlanStr = rollbackPlan.String
		}

		createdByStr := ""
		if createdBy.Valid {
			createdByStr = createdBy.String
		}

		updatedByStr := ""
		if updatedBy.Valid {
			updatedByStr = updatedBy.String
		}

		sch, err := r.mapToSchedule(idStr, scheduledAt, serviceName, description, owners, environments, status, rollbackPlanStr, createdByStr, updatedByStr, createdAt, updatedAt)
		if err != nil {
			return nil, err
		}

		schedules = append(schedules, sch)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating ungrouped schedules: %w", err)
	}

	return schedules, nil
}

// loadOwners retrieves all owners for a schedule
func (r *ScheduleRepository) loadOwners(ctx context.Context, scheduleID string) ([]string, error) {
	query := "SELECT owner FROM schedule_owners WHERE schedule_id = ? ORDER BY owner ASC"
	rows, err := r.db.QueryContext(ctx, query, scheduleID)
	if err != nil {
		return nil, fmt.Errorf("failed to query owners: %w", err)
	}
	defer rows.Close()

	var owners []string
	for rows.Next() {
		var owner string
		if err := rows.Scan(&owner); err != nil {
			return nil, fmt.Errorf("failed to scan owner: %w", err)
		}
		owners = append(owners, owner)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating owners: %w", err)
	}

	return owners, nil
}

// loadEnvironments retrieves all environments for a schedule
func (r *ScheduleRepository) loadEnvironments(ctx context.Context, scheduleID string) ([]string, error) {
	query := "SELECT environment FROM schedule_environments WHERE schedule_id = ? ORDER BY environment ASC"
	rows, err := r.db.QueryContext(ctx, query, scheduleID)
	if err != nil {
		return nil, fmt.Errorf("failed to query environments: %w", err)
	}
	defer rows.Close()

	var environments []string
	for rows.Next() {
		var env string
		if err := rows.Scan(&env); err != nil {
			return nil, fmt.Errorf("failed to scan environment: %w", err)
		}
		environments = append(environments, env)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating environments: %w", err)
	}

	return environments, nil
}

// mapToSchedule converts database row to domain schedule
func (r *ScheduleRepository) mapToSchedule(
	idStr string,
	scheduledAt time.Time,
	serviceName string,
	description string,
	ownerStrs []string,
	environmentStrs []string,
	statusStr string,
	rollbackPlanStr string,
	createdByStr string,
	updatedByStr string,
	createdAt time.Time,
	updatedAt time.Time,
) (*schedule.Schedule, error) {
	id, err := schedule.ParseScheduleID(idStr)
	if err != nil {
		return nil, fmt.Errorf("invalid schedule ID in database: %w", err)
	}

	st, err := schedule.NewScheduledTime(scheduledAt)
	if err != nil {
		return nil, fmt.Errorf("invalid scheduled time in database: %w", err)
	}

	sn, err := schedule.NewServiceName(serviceName)
	if err != nil {
		return nil, fmt.Errorf("invalid service name in database: %w", err)
	}

	// Parse owners
	owners := []schedule.Owner{}
	for _, ownerStr := range ownerStrs {
		owner, err := schedule.NewOwner(ownerStr)
		if err != nil {
			return nil, fmt.Errorf("invalid owner in database: %w", err)
		}
		owners = append(owners, owner)
	}

	// Parse environments
	environments := []schedule.Environment{}
	for _, envStr := range environmentStrs {
		env, err := schedule.NewEnvironment(strings.TrimSpace(envStr))
		if err != nil {
			return nil, fmt.Errorf("invalid environment in database: %w", err)
		}
		environments = append(environments, env)
	}

	desc := schedule.NewDescription(description)

	status, err := schedule.ParseStatus(statusStr)
	if err != nil {
		return nil, fmt.Errorf("invalid status in database: %w", err)
	}

	rollbackPlan, err := schedule.NewRollbackPlan(rollbackPlanStr)
	if err != nil {
		return nil, fmt.Errorf("invalid rollback plan in database: %w", err)
	}

	// Parse audit user IDs (may be empty for legacy records)
	var createdBy, updatedBy user.UserID
	if createdByStr != "" {
		createdBy, err = user.ParseUserID(createdByStr)
		if err != nil {
			return nil, fmt.Errorf("invalid created_by user ID in database: %w", err)
		}
	}
	if updatedByStr != "" {
		updatedBy, err = user.ParseUserID(updatedByStr)
		if err != nil {
			return nil, fmt.Errorf("invalid updated_by user ID in database: %w", err)
		}
	}

	return schedule.Reconstitute(id, st, sn, environments, desc, owners, status, rollbackPlan, createdBy, updatedBy, createdAt, updatedAt), nil
}
