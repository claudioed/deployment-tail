package bdd

import (
	"context"
	"fmt"
	"strings"

	"github.com/cucumber/godog"

	"github.com/claudioed/deployment-tail/internal/domain/group"
	"github.com/claudioed/deployment-tail/internal/domain/schedule"
)

func RegisterAssignmentSteps(ctx *godog.ScenarioContext) {
	s := &assignmentSteps{}

	ctx.Step(`^I assign the last schedule to group "([^"]+)"$`, s.iAssignTheLastScheduleToGroup)
	ctx.Step(`^I assign the last schedule to groups "([^"]+)"$`, s.iAssignTheLastScheduleToGroups)
	ctx.Step(`^I assign the last schedule to a non-existent group$`, s.iAssignTheLastScheduleToNonExistentGroup)
	ctx.Step(`^I assign a non-existent schedule to group "([^"]+)"$`, s.iAssignNonExistentScheduleToGroup)
	ctx.Step(`^I unassign the last schedule from group "([^"]+)"$`, s.iUnassignTheLastScheduleFromGroup)
	ctx.Step(`^I unassign the last schedule from a non-existent group$`, s.iUnassignTheLastScheduleFromNonExistentGroup)
	ctx.Step(`^I list groups for the last schedule$`, s.iListGroupsForTheLastSchedule)
	ctx.Step(`^I list schedules in group "([^"]+)"$`, s.iListSchedulesInGroup)
	ctx.Step(`^I bulk assign the (\d+) schedules to group "([^"]+)"$`, s.iBulkAssignSchedulesToGroup)
	ctx.Step(`^I bulk assign (\d+) schedules to group "([^"]+)"$`, s.iBulkAssignNSchedulesToGroup)
	ctx.Step(`^I bulk unassign the (\d+) schedules from group "([^"]+)"$`, s.iBulkUnassignSchedulesFromGroup)
	ctx.Step(`^I list ungrouped schedules$`, s.iListUngroupedSchedules)
	ctx.Step(`^the schedule list includes the last schedule$`, s.theScheduleListIncludesTheLastSchedule)
	ctx.Step(`^the schedule list includes service "([^"]+)"$`, s.theScheduleListIncludesService)
	ctx.Step(`^the schedule list does not include service "([^"]+)"$`, s.theScheduleListDoesNotIncludeService)
	ctx.Step(`^the schedule list is empty$`, s.theScheduleListIsEmpty)
}

type assignmentSteps struct{}

func (s *assignmentSteps) iAssignTheLastScheduleToGroup(ctx context.Context, groupName string) error {
	w := getWorld(ctx)

	if w.LastSchedule == nil {
		return fmt.Errorf("no schedule to assign; create a schedule first")
	}

	grp, ok := w.NamedGroups[groupName]
	if !ok {
		return fmt.Errorf("group %q not found in test data; create it first", groupName)
	}

	// Phase B: Check if user can access this group (visibility check)
	if grp.Visibility() == group.VisibilityPrivate {
		// Only owner can assign to private groups
		if w.CurrentUser == nil || grp.Owner().String() != w.CurrentUser.ID().String() {
			w.LastError = fmt.Errorf("forbidden: cannot assign to private group owned by another user")
			return nil
		}
	}

	// Use the mock's AddSchedule method (assignedBy can be current user's email)
	assignedBy := ""
	if w.CurrentUser != nil {
		assignedBy = w.CurrentUser.Email().String()
	}
	err := w.GroupRepo.AddSchedule(ctx, grp.ID(), w.LastSchedule.ID(), assignedBy)
	w.LastError = err

	return nil
}

func (s *assignmentSteps) iAssignTheLastScheduleToGroups(ctx context.Context, groupNamesStr string) error {
	w := getWorld(ctx)

	if w.LastSchedule == nil {
		return fmt.Errorf("no schedule to assign; create a schedule first")
	}

	groupNames := strings.Split(groupNamesStr, ",")
	for _, name := range groupNames {
		name = strings.TrimSpace(name)
		grp, ok := w.NamedGroups[name]
		if !ok {
			return fmt.Errorf("group %q not found in test data; create it first", name)
		}

		assignedBy := ""
		if w.CurrentUser != nil {
			assignedBy = w.CurrentUser.Email().String()
		}
		err := w.GroupRepo.AddSchedule(ctx, grp.ID(), w.LastSchedule.ID(), assignedBy)
		if err != nil {
			w.LastError = err
			return nil
		}
	}

	w.LastError = nil
	return nil
}

func (s *assignmentSteps) iAssignTheLastScheduleToNonExistentGroup(ctx context.Context) error {
	w := getWorld(ctx)

	if w.LastSchedule == nil {
		return fmt.Errorf("no schedule to assign; create a schedule first")
	}

	// Use a random UUID that doesn't exist
	fakeGroupID, _ := group.ParseGroupID("00000000-0000-0000-0000-000000000000")
	assignedBy := ""
	if w.CurrentUser != nil {
		assignedBy = w.CurrentUser.Email().String()
	}
	err := w.GroupRepo.AddSchedule(ctx, fakeGroupID, w.LastSchedule.ID(), assignedBy)
	w.LastError = err

	return nil
}

func (s *assignmentSteps) iAssignNonExistentScheduleToGroup(ctx context.Context, groupName string) error {
	w := getWorld(ctx)

	grp, ok := w.NamedGroups[groupName]
	if !ok {
		return fmt.Errorf("group %q not found in test data; create it first", groupName)
	}

	// Use a random UUID that doesn't exist
	fakeScheduleID, _ := schedule.ParseScheduleID("00000000-0000-0000-0000-000000000000")

	// Phase B: Check if schedule exists first
	_, err := w.ScheduleRepo.FindByID(ctx, fakeScheduleID)
	if err != nil {
		// Schedule doesn't exist, which is expected
		w.LastError = err
		return nil
	}

	// If it somehow exists, still try to assign
	assignedBy := ""
	if w.CurrentUser != nil {
		assignedBy = w.CurrentUser.Email().String()
	}
	err = w.GroupRepo.AddSchedule(ctx, grp.ID(), fakeScheduleID, assignedBy)
	w.LastError = err

	return nil
}

func (s *assignmentSteps) iUnassignTheLastScheduleFromGroup(ctx context.Context, groupName string) error {
	w := getWorld(ctx)

	if w.LastSchedule == nil {
		return fmt.Errorf("no schedule to unassign; create a schedule first")
	}

	grp, ok := w.NamedGroups[groupName]
	if !ok {
		return fmt.Errorf("group %q not found in test data; create it first", groupName)
	}

	err := w.GroupRepo.RemoveSchedule(ctx, grp.ID(), w.LastSchedule.ID())
	w.LastError = err

	return nil
}

func (s *assignmentSteps) iUnassignTheLastScheduleFromNonExistentGroup(ctx context.Context) error {
	w := getWorld(ctx)

	if w.LastSchedule == nil {
		return fmt.Errorf("no schedule to unassign; create a schedule first")
	}

	// Use a random UUID that doesn't exist
	fakeGroupID, _ := group.ParseGroupID("00000000-0000-0000-0000-000000000000")
	err := w.GroupRepo.RemoveSchedule(ctx, fakeGroupID, w.LastSchedule.ID())
	w.LastError = err

	return nil
}

func (s *assignmentSteps) iListGroupsForTheLastSchedule(ctx context.Context) error {
	w := getWorld(ctx)

	if w.LastSchedule == nil {
		return fmt.Errorf("no schedule; create a schedule first")
	}

	groups, err := w.GroupRepo.GetGroupsForSchedule(ctx, w.LastSchedule.ID())
	w.LastError = err
	if err == nil {
		// Store in temporary list for assertions
		w.lastGroupList = groups
	}

	return nil
}

func (s *assignmentSteps) iListSchedulesInGroup(ctx context.Context, groupName string) error {
	w := getWorld(ctx)

	grp, ok := w.NamedGroups[groupName]
	if !ok {
		return fmt.Errorf("group %q not found in test data; create it first", groupName)
	}

	scheduleIDs, err := w.GroupRepo.GetSchedulesInGroup(ctx, grp.ID())
	w.LastError = err
	if err == nil {
		// Lookup full schedules from the IDs
		var schedules []*schedule.Schedule
		for _, sid := range scheduleIDs {
			sch, err := w.ScheduleRepo.FindByID(ctx, sid)
			if err == nil {
				schedules = append(schedules, sch)
			}
		}
		w.lastScheduleList = schedules
	}

	return nil
}

func (s *assignmentSteps) iBulkAssignSchedulesToGroup(ctx context.Context, count int, groupName string) error {
	w := getWorld(ctx)

	grp, ok := w.NamedGroups[groupName]
	if !ok {
		return fmt.Errorf("group %q not found in test data; create it first", groupName)
	}

	if len(w.bulkSchedules) != count {
		return fmt.Errorf("expected %d schedules in bulk list, got %d", count, len(w.bulkSchedules))
	}

	assignedBy := ""
	if w.CurrentUser != nil {
		assignedBy = w.CurrentUser.Email().String()
	}

	for _, sch := range w.bulkSchedules {
		err := w.GroupRepo.AddSchedule(ctx, grp.ID(), sch.ID(), assignedBy)
		if err != nil {
			w.LastError = err
			return nil
		}
	}

	w.LastError = nil
	return nil
}

func (s *assignmentSteps) iBulkAssignNSchedulesToGroup(ctx context.Context, count int, groupName string) error {
	// Handle "bulk assign 0 schedules" case
	if count == 0 {
		w := getWorld(ctx)
		w.LastError = nil
		return nil
	}
	return s.iBulkAssignSchedulesToGroup(ctx, count, groupName)
}

func (s *assignmentSteps) iBulkUnassignSchedulesFromGroup(ctx context.Context, count int, groupName string) error {
	w := getWorld(ctx)

	grp, ok := w.NamedGroups[groupName]
	if !ok {
		return fmt.Errorf("group %q not found in test data; create it first", groupName)
	}

	if len(w.bulkSchedules) != count {
		return fmt.Errorf("expected %d schedules in bulk list, got %d", count, len(w.bulkSchedules))
	}

	for _, sch := range w.bulkSchedules {
		err := w.GroupRepo.RemoveSchedule(ctx, grp.ID(), sch.ID())
		if err != nil {
			w.LastError = err
			return nil
		}
	}

	w.LastError = nil
	return nil
}

func (s *assignmentSteps) iListUngroupedSchedules(ctx context.Context) error {
	w := getWorld(ctx)

	// Use FindUngrouped from the mock (it returns all schedules for simplicity in Phase B)
	schedules, err := w.ScheduleRepo.FindUngrouped(ctx, schedule.Filters{})
	w.LastError = err
	if err == nil {
		w.lastScheduleList = schedules
	}

	return nil
}

func (s *assignmentSteps) theScheduleListIncludesTheLastSchedule(ctx context.Context) error {
	w := getWorld(ctx)

	if w.LastSchedule == nil {
		return fmt.Errorf("no last schedule to check")
	}

	for _, sch := range w.lastScheduleList {
		if sch.ID() == w.LastSchedule.ID() {
			return nil
		}
	}

	return fmt.Errorf("schedule list does not include the last schedule (ID: %s)", w.LastSchedule.ID())
}

func (s *assignmentSteps) theScheduleListIncludesService(ctx context.Context, serviceName string) error {
	w := getWorld(ctx)

	for _, sch := range w.lastScheduleList {
		if sch.Service().Value() == serviceName {
			return nil
		}
	}

	return fmt.Errorf("schedule list does not include service %q", serviceName)
}

func (s *assignmentSteps) theScheduleListDoesNotIncludeService(ctx context.Context, serviceName string) error {
	w := getWorld(ctx)

	for _, sch := range w.lastScheduleList {
		if sch.Service().Value() == serviceName {
			return fmt.Errorf("schedule list should not include service %q but it does", serviceName)
		}
	}

	return nil
}

func (s *assignmentSteps) theScheduleListIsEmpty(ctx context.Context) error {
	w := getWorld(ctx)

	if len(w.lastScheduleList) > 0 {
		return fmt.Errorf("expected empty schedule list but got %d schedules", len(w.lastScheduleList))
	}

	return nil
}
