package bdd

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/cucumber/godog"

	"github.com/claudioed/deployment-tail/internal/application/ports/input"
	"github.com/claudioed/deployment-tail/internal/domain/schedule"
	"github.com/claudioed/deployment-tail/internal/domain/user"
)

func RegisterScheduleSteps(ctx *godog.ScenarioContext) {
	s := &scheduleSteps{}

	// Phase A steps
	ctx.Step(`^I create a schedule with:$`, s.iCreateAScheduleWith)
	ctx.Step(`^the last schedule has service name "([^"]+)"$`, s.theLastScheduleHasServiceName)
	ctx.Step(`^the last schedule status is "([^"]+)"$`, s.theLastScheduleHasStatus)
	ctx.Step(`^I retrieve the schedule by ID$`, s.iRetrieveTheScheduleByID)
	ctx.Step(`^the schedule has service name "([^"]+)"$`, s.theScheduleHasServiceName)

	// Phase B: Bulk and delete operations
	ctx.Step(`^I create (\d+) schedules$`, s.iCreateNSchedules)
	ctx.Step(`^I delete the last schedule$`, s.iDeleteTheLastSchedule)

	// Phase C: Additional operations
	ctx.Step(`^I list all schedules$`, s.iListAllSchedules)
	ctx.Step(`^I create a schedule with service name "([^"]+)"$`, s.iCreateAScheduleWithServiceName)
	ctx.Step(`^I update the last schedule description to "([^"]+)"$`, s.iUpdateTheLastScheduleDescriptionTo)
	ctx.Step(`^I get the last schedule by ID$`, s.iGetTheLastScheduleByID)
}

type scheduleSteps struct{}

func (s *scheduleSteps) iCreateAScheduleWith(ctx context.Context, table *godog.Table) error {
	w := getWorld(ctx)

	if w.CurrentUser == nil {
		return fmt.Errorf("no authenticated user set; use 'Given I am authenticated as' first")
	}

	// Parse vertical table (key-value pairs) into a map
	// Example:
	//   | service_name | api-service |
	//   | scheduled_at | 2026-06-15T14:00:00Z |
	params := make(map[string]string)
	for _, row := range table.Rows {
		if len(row.Cells) != 2 {
			return fmt.Errorf("each row must have exactly 2 cells (key, value), got %d", len(row.Cells))
		}
		key := row.Cells[0].Value
		value := row.Cells[1].Value
		params[key] = value
	}

	// Build CreateScheduleCommand (with string fields as per API contract)
	cmd := input.CreateScheduleCommand{}

	// Parse required fields
	if serviceName, ok := params["service_name"]; ok {
		// Validate via domain object
		_, err := schedule.NewServiceName(serviceName)
		if err != nil {
			return fmt.Errorf("invalid service_name: %w", err)
		}
		cmd.ServiceName = serviceName
	} else {
		return fmt.Errorf("missing required field: service_name")
	}

	if scheduledAt, ok := params["scheduled_at"]; ok {
		t, err := time.Parse(time.RFC3339, scheduledAt)
		if err != nil {
			return fmt.Errorf("invalid scheduled_at format (expected RFC3339): %w", err)
		}
		// Validate via domain object
		_, err = schedule.NewScheduledTime(t)
		if err != nil {
			return fmt.Errorf("invalid scheduled_at: %w", err)
		}
		cmd.ScheduledAt = t
	} else {
		return fmt.Errorf("missing required field: scheduled_at")
	}

	// Parse environments
	if envStr, ok := params["environments"]; ok {
		envs := strings.Split(envStr, ",")
		for _, e := range envs {
			// Validate via domain object
			env := strings.TrimSpace(e)
			_, err := schedule.NewEnvironment(env)
			if err != nil {
				return fmt.Errorf("invalid environment %q: %w", e, err)
			}
			cmd.Environments = append(cmd.Environments, env)
		}
	}

	// Parse owners
	if ownerStr, ok := params["owners"]; ok {
		ownerEmails := strings.Split(ownerStr, ",")
		for _, emailStr := range ownerEmails {
			// Validate via domain object
			email := strings.TrimSpace(emailStr)
			_, err := user.NewEmail(email)
			if err != nil {
				return fmt.Errorf("invalid owner email %q: %w", emailStr, err)
			}
			cmd.Owners = append(cmd.Owners, email)
		}
	}

	// Parse optional description
	if desc, ok := params["description"]; ok {
		// Description doesn't need validation
		cmd.Description = desc
	}

	// Call service with authenticated user ID
	sch, err := w.ScheduleService.CreateSchedule(ctx, cmd, w.CurrentUser.ID())
	w.LastError = err
	if err == nil {
		w.LastSchedule = sch
	}

	return nil
}

func (s *scheduleSteps) theLastScheduleHasServiceName(ctx context.Context, expectedName string) error {
	w := getWorld(ctx)
	if w.LastSchedule == nil {
		return fmt.Errorf("no schedule was created")
	}
	if w.LastSchedule.Service().Value() != expectedName {
		return fmt.Errorf("expected service name %q but got %q", expectedName, w.LastSchedule.Service().Value())
	}
	return nil
}

func (s *scheduleSteps) theLastScheduleHasStatus(ctx context.Context, expectedStatus string) error {
	w := getWorld(ctx)
	if w.LastSchedule == nil {
		return fmt.Errorf("no schedule was created")
	}
	actualStatus := w.LastSchedule.Status().String()
	if actualStatus != expectedStatus {
		return fmt.Errorf("expected status %q but got %q", expectedStatus, actualStatus)
	}
	return nil
}

func (s *scheduleSteps) iRetrieveTheScheduleByID(ctx context.Context) error {
	w := getWorld(ctx)
	if w.LastSchedule == nil {
		return fmt.Errorf("no schedule ID to retrieve; create a schedule first")
	}

	sch, err := w.ScheduleService.GetSchedule(ctx, w.LastSchedule.ID().String())
	w.LastError = err
	if err == nil {
		w.LastSchedule = sch
	}

	return nil
}

func (s *scheduleSteps) theScheduleHasServiceName(ctx context.Context, expectedName string) error {
	return s.theLastScheduleHasServiceName(ctx, expectedName)
}

// Phase B: Bulk operations
func (s *scheduleSteps) iCreateNSchedules(ctx context.Context, count int) error {
	w := getWorld(ctx)

	if w.CurrentUser == nil {
		return fmt.Errorf("no authenticated user set; use 'Given I am authenticated as' first")
	}

	w.bulkSchedules = make([]*schedule.Schedule, 0, count)

	for i := 0; i < count; i++ {
		cmd := input.CreateScheduleCommand{
			ServiceName:  fmt.Sprintf("bulk-service-%d", i+1),
			ScheduledAt:  time.Now().Add(time.Duration(i+1) * time.Hour),
			Environments: []string{"production"},
			Owners:       []string{w.CurrentUser.Email().String()},
		}

		sch, err := w.ScheduleService.CreateSchedule(ctx, cmd, w.CurrentUser.ID())
		if err != nil {
			w.LastError = err
			return nil
		}

		w.bulkSchedules = append(w.bulkSchedules, sch)
		w.LastSchedule = sch // Keep last one as LastSchedule
	}

	w.LastError = nil
	return nil
}

// Phase B: Delete operations
func (s *scheduleSteps) iDeleteTheLastSchedule(ctx context.Context) error {
	w := getWorld(ctx)

	if w.LastSchedule == nil {
		return fmt.Errorf("no schedule to delete; create a schedule first")
	}

	if w.CurrentUser == nil {
		return fmt.Errorf("no authenticated user; cannot delete schedule")
	}

	err := w.ScheduleRepo.Delete(ctx, w.LastSchedule.ID(), w.CurrentUser.ID())
	w.LastError = err

	return nil
}

// Phase C: Additional operations

func (s *scheduleSteps) iListAllSchedules(ctx context.Context) error {
	w := getWorld(ctx)

	query := input.ListSchedulesQuery{}
	schedules, err := w.ScheduleService.ListSchedules(ctx, query)
	w.LastError = err
	if err == nil {
		w.lastScheduleList = schedules
	}

	return nil
}

func (s *scheduleSteps) iCreateAScheduleWithServiceName(ctx context.Context, serviceName string) error {
	w := getWorld(ctx)

	if w.CurrentUser == nil {
		return fmt.Errorf("no authenticated user set; use 'Given I am authenticated as' first")
	}

	cmd := input.CreateScheduleCommand{
		ServiceName:  serviceName,
		ScheduledAt:  time.Now().Add(24 * time.Hour),
		Environments: []string{"production"},
		Owners:       []string{w.CurrentUser.Email().String()},
		Description:  "Test schedule",
	}

	sch, err := w.ScheduleService.CreateSchedule(ctx, cmd, w.CurrentUser.ID())
	w.LastError = err
	if err == nil {
		w.LastSchedule = sch
	}

	return nil
}

func (s *scheduleSteps) iUpdateTheLastScheduleDescriptionTo(ctx context.Context, description string) error {
	w := getWorld(ctx)

	if w.LastSchedule == nil {
		return fmt.Errorf("no schedule to update; create a schedule first")
	}

	if w.CurrentUser == nil {
		return fmt.Errorf("no authenticated user; cannot update schedule")
	}

	cmd := input.UpdateScheduleCommand{
		ID:          w.LastSchedule.ID().String(),
		Description: &description,
	}

	updatedSchedule, err := w.ScheduleService.UpdateSchedule(ctx, cmd, w.CurrentUser.ID())
	w.LastError = err
	if err == nil {
		w.LastSchedule = updatedSchedule
	}

	return nil
}

func (s *scheduleSteps) iGetTheLastScheduleByID(ctx context.Context) error {
	w := getWorld(ctx)

	if w.LastSchedule == nil {
		return fmt.Errorf("no schedule ID to retrieve; create a schedule first")
	}

	sch, err := w.ScheduleService.GetSchedule(ctx, w.LastSchedule.ID().String())
	w.LastError = err
	if err == nil {
		w.LastSchedule = sch
	}

	return nil
}
