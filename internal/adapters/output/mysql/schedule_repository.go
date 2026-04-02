package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/claudioed/deployment-tail/internal/domain/schedule"
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
	query := `
		INSERT INTO schedules (id, scheduled_at, service_name, environment, description, owner, status, rollback_plan, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	var rollbackPlan *string
	if !sch.RollbackPlan().IsEmpty() {
		val := sch.RollbackPlan().String()
		rollbackPlan = &val
	}

	_, err := r.db.ExecContext(ctx, query,
		sch.ID().String(),
		sch.ScheduledAt().Value(),
		sch.Service().Value(),
		sch.Environment().String(),
		sch.Description().Value(),
		sch.Owner().String(),
		sch.Status().String(),
		rollbackPlan,
		sch.CreatedAt(),
		sch.UpdatedAt(),
	)

	if err != nil {
		return fmt.Errorf("failed to create schedule: %w", err)
	}

	return nil
}

// FindByID retrieves a schedule by its ID
func (r *ScheduleRepository) FindByID(ctx context.Context, id schedule.ScheduleID) (*schedule.Schedule, error) {
	query := `
		SELECT id, scheduled_at, service_name, environment, description, owner, status, rollback_plan, created_at, updated_at
		FROM schedules
		WHERE id = ?
	`

	var (
		idStr        string
		scheduledAt  time.Time
		serviceName  string
		environment  string
		description  string
		owner        string
		status       string
		rollbackPlan sql.NullString
		createdAt    time.Time
		updatedAt    time.Time
	)

	err := r.db.QueryRowContext(ctx, query, id.String()).Scan(
		&idStr,
		&scheduledAt,
		&serviceName,
		&environment,
		&description,
		&owner,
		&status,
		&rollbackPlan,
		&createdAt,
		&updatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, schedule.ErrScheduleNotFound
		}
		return nil, fmt.Errorf("failed to find schedule: %w", err)
	}

	rollbackPlanStr := ""
	if rollbackPlan.Valid {
		rollbackPlanStr = rollbackPlan.String
	}

	return r.mapToSchedule(idStr, scheduledAt, serviceName, environment, description, owner, status, rollbackPlanStr, createdAt, updatedAt)
}

// FindAll retrieves schedules with optional filters
func (r *ScheduleRepository) FindAll(ctx context.Context, filters schedule.Filters) ([]*schedule.Schedule, error) {
	query := "SELECT id, scheduled_at, service_name, environment, description, owner, status, rollback_plan, created_at, updated_at FROM schedules WHERE 1=1"
	args := []interface{}{}

	// Apply filters
	if filters.From != nil {
		query += " AND scheduled_at >= ?"
		args = append(args, *filters.From)
	}

	if filters.To != nil {
		query += " AND scheduled_at <= ?"
		args = append(args, *filters.To)
	}

	if filters.Environment != nil {
		query += " AND environment = ?"
		args = append(args, filters.Environment.String())
	}

	if filters.Owner != nil {
		query += " AND owner = ?"
		args = append(args, filters.Owner.String())
	}

	if filters.Status != nil {
		query += " AND status = ?"
		args = append(args, filters.Status.String())
	}

	query += " ORDER BY scheduled_at ASC"

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
			environment  string
			description  string
			owner        string
			status       string
			rollbackPlan sql.NullString
			createdAt    time.Time
			updatedAt    time.Time
		)

		if err := rows.Scan(&idStr, &scheduledAt, &serviceName, &environment, &description, &owner, &status, &rollbackPlan, &createdAt, &updatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan schedule: %w", err)
		}

		rollbackPlanStr := ""
		if rollbackPlan.Valid {
			rollbackPlanStr = rollbackPlan.String
		}

		sch, err := r.mapToSchedule(idStr, scheduledAt, serviceName, environment, description, owner, status, rollbackPlanStr, createdAt, updatedAt)
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
	query := `
		UPDATE schedules
		SET scheduled_at = ?, service_name = ?, environment = ?, description = ?, status = ?, rollback_plan = ?, updated_at = ?
		WHERE id = ?
	`

	var rollbackPlan *string
	if !sch.RollbackPlan().IsEmpty() {
		val := sch.RollbackPlan().String()
		rollbackPlan = &val
	}

	result, err := r.db.ExecContext(ctx, query,
		sch.ScheduledAt().Value(),
		sch.Service().Value(),
		sch.Environment().String(),
		sch.Description().Value(),
		sch.Status().String(),
		rollbackPlan,
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

	return nil
}

// Delete removes a schedule
func (r *ScheduleRepository) Delete(ctx context.Context, id schedule.ScheduleID) error {
	query := "DELETE FROM schedules WHERE id = ?"

	result, err := r.db.ExecContext(ctx, query, id.String())
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

// mapToSchedule converts database row to domain schedule
func (r *ScheduleRepository) mapToSchedule(
	idStr string,
	scheduledAt time.Time,
	serviceName string,
	environment string,
	description string,
	ownerStr string,
	statusStr string,
	rollbackPlanStr string,
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

	env, err := schedule.NewEnvironment(strings.TrimSpace(environment))
	if err != nil {
		return nil, fmt.Errorf("invalid environment in database: %w", err)
	}

	desc := schedule.NewDescription(description)

	owner, err := schedule.NewOwner(ownerStr)
	if err != nil {
		return nil, fmt.Errorf("invalid owner in database: %w", err)
	}

	status, err := schedule.ParseStatus(statusStr)
	if err != nil {
		return nil, fmt.Errorf("invalid status in database: %w", err)
	}

	rollbackPlan, err := schedule.NewRollbackPlan(rollbackPlanStr)
	if err != nil {
		return nil, fmt.Errorf("invalid rollback plan in database: %w", err)
	}

	return schedule.Reconstitute(id, st, sn, env, desc, owner, status, rollbackPlan, createdAt, updatedAt), nil
}
