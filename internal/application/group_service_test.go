package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/claudioed/deployment-tail/internal/application/applicationtest"
	"github.com/claudioed/deployment-tail/internal/application/ports/input"
	"github.com/claudioed/deployment-tail/internal/domain/group"
	"github.com/claudioed/deployment-tail/internal/domain/user"
)

// Test Group Service

func TestCreateGroup(t *testing.T) {
	groupRepo := applicationtest.NewMockGroupRepository()
	scheduleRepo := applicationtest.NewMockRepository()
	service := NewGroupService(groupRepo, scheduleRepo)

	cmd := input.CreateGroupCommand{
		Name:        "Project Alpha",
		Description: "All schedules for Project Alpha",
		Owner:       "test-user",
	}

	grp, err := service.CreateGroup(context.Background(), cmd)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if grp == nil {
		t.Fatal("expected group to be created")
	}

	if grp.Name().String() != "Project Alpha" {
		t.Errorf("expected name 'Project Alpha', got %v", grp.Name().String())
	}

	if grp.Description().String() != "All schedules for Project Alpha" {
		t.Errorf("unexpected description: %v", grp.Description().String())
	}
}

func TestCreateGroupWithDuplicateName(t *testing.T) {
	groupRepo := applicationtest.NewMockGroupRepository()
	scheduleRepo := applicationtest.NewMockRepository()
	service := NewGroupService(groupRepo, scheduleRepo)

	cmd := input.CreateGroupCommand{
		Name:  "Project Alpha",
		Owner: "test-user",
	}

	// First creation should succeed
	_, err := service.CreateGroup(context.Background(), cmd)
	if err != nil {
		t.Fatalf("expected no error on first create, got %v", err)
	}

	// Simulate duplicate name error
	groupRepo.DuplicateNameErr = true

	// Second creation should fail
	_, err = service.CreateGroup(context.Background(), cmd)
	if !errors.Is(err, group.ErrDuplicateGroupName) {
		t.Errorf("expected ErrDuplicateGroupName, got %v", err)
	}
}

func TestCreateGroupWithInvalidName(t *testing.T) {
	groupRepo := applicationtest.NewMockGroupRepository()
	scheduleRepo := applicationtest.NewMockRepository()
	service := NewGroupService(groupRepo, scheduleRepo)

	cmd := input.CreateGroupCommand{
		Name:  "", // Empty name should fail
		Owner: "test-user",
	}

	_, err := service.CreateGroup(context.Background(), cmd)
	if err == nil {
		t.Fatal("expected error for empty group name")
	}
}

func TestGetGroup(t *testing.T) {
	groupRepo := applicationtest.NewMockGroupRepository()
	scheduleRepo := applicationtest.NewMockRepository()
	service := NewGroupService(groupRepo, scheduleRepo)

	// Create a group first
	cmd := input.CreateGroupCommand{
		Name:  "Project Alpha",
		Owner: "test-user",
	}
	created, _ := service.CreateGroup(context.Background(), cmd)

	// Get the group
	retrieved, err := service.GetGroup(context.Background(), created.ID().String())

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !retrieved.ID().Equals(created.ID()) {
		t.Errorf("expected ID %v, got %v", created.ID(), retrieved.ID())
	}
}

func TestGetGroupNotFound(t *testing.T) {
	groupRepo := applicationtest.NewMockGroupRepository()
	scheduleRepo := applicationtest.NewMockRepository()
	service := NewGroupService(groupRepo, scheduleRepo)

	id := group.NewGroupID()

	_, err := service.GetGroup(context.Background(), id.String())

	if !errors.Is(err, group.ErrGroupNotFound) {
		t.Errorf("expected ErrGroupNotFound, got %v", err)
	}
}

func TestListGroups(t *testing.T) {
	groupRepo := applicationtest.NewMockGroupRepository()
	scheduleRepo := applicationtest.NewMockRepository()
	service := NewGroupService(groupRepo, scheduleRepo)

	// Create multiple groups
	for i := 0; i < 3; i++ {
		cmd := input.CreateGroupCommand{
			Name:  "Group " + string(rune('A'+i)),
			Owner: "test-user",
		}
		_, _ = service.CreateGroup(context.Background(), cmd)
	}

	query := input.ListGroupsQuery{
		Owner: "test-user",
	}

	groups, err := service.ListGroups(context.Background(), query)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(groups) != 3 {
		t.Errorf("expected 3 groups, got %d", len(groups))
	}
}

func TestUpdateGroup(t *testing.T) {
	groupRepo := applicationtest.NewMockGroupRepository()
	scheduleRepo := applicationtest.NewMockRepository()
	service := NewGroupService(groupRepo, scheduleRepo)

	// Create a group
	cmd := input.CreateGroupCommand{
		Name:  "Project Alpha",
		Owner: "test-user",
	}
	created, _ := service.CreateGroup(context.Background(), cmd)

	// Update the group
	updateCmd := input.UpdateGroupCommand{
		ID:          created.ID().String(),
		Name:        "Project Beta",
		Description: "Updated description",
	}

	updated, err := service.UpdateGroup(context.Background(), updateCmd)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if updated.Name().String() != "Project Beta" {
		t.Errorf("expected name 'Project Beta', got %v", updated.Name().String())
	}

	if updated.Description().String() != "Updated description" {
		t.Errorf("unexpected description: %v", updated.Description().String())
	}
}

func TestDeleteGroup(t *testing.T) {
	groupRepo := applicationtest.NewMockGroupRepository()
	scheduleRepo := applicationtest.NewMockRepository()
	service := NewGroupService(groupRepo, scheduleRepo)

	// Create a group
	cmd := input.CreateGroupCommand{
		Name:  "Project Alpha",
		Owner: "test-user",
	}
	created, _ := service.CreateGroup(context.Background(), cmd)

	// Delete the group
	deleteCmd := input.DeleteGroupCommand{
		ID: created.ID().String(),
	}

	err := service.DeleteGroup(context.Background(), deleteCmd)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify group is gone
	_, err = service.GetGroup(context.Background(), created.ID().String())
	if !errors.Is(err, group.ErrGroupNotFound) {
		t.Errorf("expected group to be deleted")
	}
}

func TestAssignScheduleToGroups(t *testing.T) {
	groupRepo := applicationtest.NewMockGroupRepository()
	scheduleRepo := applicationtest.NewMockRepository()
	service := NewGroupService(groupRepo, scheduleRepo)

	// Create a group
	groupCmd := input.CreateGroupCommand{
		Name:  "Project Alpha",
		Owner: "test-user",
	}
	grp, _ := service.CreateGroup(context.Background(), groupCmd)

	// Create a schedule
	userRepo := applicationtest.NewMockUserRepository()
	googleID, _ := user.NewGoogleID("deployer123")
	email, _ := user.NewEmail("deployer@example.com")
	name, _ := user.NewUserName("Test Deployer")
	role, _ := user.NewRole(user.RoleDeployer)
	deployer, _ := user.NewUser(googleID, email, name, role)
	userRepo.Users[deployer.ID().String()] = deployer

	schedCmd := input.CreateScheduleCommand{
		ScheduledAt:  time.Now().Add(24 * time.Hour),
		ServiceName:  "test-service",
		Environments: []string{"production"},
		Owners:       []string{"test-user"},
	}
	schedService := NewScheduleService(scheduleRepo, userRepo)
	sch, _ := schedService.CreateSchedule(context.Background(), schedCmd, deployer.ID())

	// Assign schedule to group
	assignCmd := input.AssignScheduleCommand{
		ScheduleID: sch.ID().String(),
		GroupIDs:   []string{grp.ID().String()},
		AssignedBy: "test-user",
	}

	err := service.AssignScheduleToGroups(context.Background(), assignCmd)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify assignment
	groups, err := service.GetGroupsForSchedule(context.Background(), sch.ID().String())
	if err != nil {
		t.Fatalf("expected no error getting groups, got %v", err)
	}

	if len(groups) != 1 {
		t.Errorf("expected 1 group, got %d", len(groups))
	}
}

func TestUnassignScheduleFromGroup(t *testing.T) {
	groupRepo := applicationtest.NewMockGroupRepository()
	scheduleRepo := applicationtest.NewMockRepository()
	service := NewGroupService(groupRepo, scheduleRepo)

	// Create a group
	groupCmd := input.CreateGroupCommand{
		Name:  "Project Alpha",
		Owner: "test-user",
	}
	grp, _ := service.CreateGroup(context.Background(), groupCmd)

	// Create a schedule
	userRepo := applicationtest.NewMockUserRepository()
	googleID, _ := user.NewGoogleID("deployer123")
	email, _ := user.NewEmail("deployer@example.com")
	name, _ := user.NewUserName("Test Deployer")
	role, _ := user.NewRole(user.RoleDeployer)
	deployer, _ := user.NewUser(googleID, email, name, role)
	userRepo.Users[deployer.ID().String()] = deployer

	schedCmd := input.CreateScheduleCommand{
		ScheduledAt:  time.Now().Add(24 * time.Hour),
		ServiceName:  "test-service",
		Environments: []string{"production"},
		Owners:       []string{"test-user"},
	}
	schedService := NewScheduleService(scheduleRepo, userRepo)
	sch, _ := schedService.CreateSchedule(context.Background(), schedCmd, deployer.ID())

	// Assign and then unassign
	assignCmd := input.AssignScheduleCommand{
		ScheduleID: sch.ID().String(),
		GroupIDs:   []string{grp.ID().String()},
		AssignedBy: "test-user",
	}
	_ = service.AssignScheduleToGroups(context.Background(), assignCmd)

	unassignCmd := input.UnassignScheduleCommand{
		ScheduleID: sch.ID().String(),
		GroupID:    grp.ID().String(),
	}

	err := service.UnassignScheduleFromGroup(context.Background(), unassignCmd)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify unassignment
	groups, _ := service.GetGroupsForSchedule(context.Background(), sch.ID().String())
	if len(groups) != 0 {
		t.Errorf("expected 0 groups after unassignment, got %d", len(groups))
	}
}

// Favorite operations tests

func TestFavoriteGroup(t *testing.T) {
	groupRepo := applicationtest.NewMockGroupRepository()
	scheduleRepo := applicationtest.NewMockRepository()
	service := NewGroupService(groupRepo, scheduleRepo)

	// Create a group
	cmd := input.CreateGroupCommand{
		Name:  "Project Alpha",
		Owner: "test-user",
	}
	grp, _ := service.CreateGroup(context.Background(), cmd)

	// Create a user
	userID := user.NewUserID()

	// Favorite the group
	err := service.FavoriteGroup(context.Background(), userID.String(), grp.ID().String())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify it's favorited
	isFav, err := groupRepo.IsFavorite(context.Background(), userID, grp.ID())
	if err != nil {
		t.Fatalf("expected no error checking favorite, got %v", err)
	}
	if !isFav {
		t.Error("expected group to be favorited")
	}
}

func TestFavoriteGroupInvalidUserID(t *testing.T) {
	groupRepo := applicationtest.NewMockGroupRepository()
	scheduleRepo := applicationtest.NewMockRepository()
	service := NewGroupService(groupRepo, scheduleRepo)

	// Create a group
	cmd := input.CreateGroupCommand{
		Name:  "Project Alpha",
		Owner: "test-user",
	}
	grp, _ := service.CreateGroup(context.Background(), cmd)

	// Try to favorite with invalid user ID
	err := service.FavoriteGroup(context.Background(), "invalid-uuid", grp.ID().String())
	if err == nil {
		t.Fatal("expected error for invalid user ID")
	}
}

func TestFavoriteGroupInvalidGroupID(t *testing.T) {
	groupRepo := applicationtest.NewMockGroupRepository()
	scheduleRepo := applicationtest.NewMockRepository()
	service := NewGroupService(groupRepo, scheduleRepo)

	userID := user.NewUserID()

	// Try to favorite with invalid group ID
	err := service.FavoriteGroup(context.Background(), userID.String(), "invalid-uuid")
	if err == nil {
		t.Fatal("expected error for invalid group ID")
	}
}

func TestFavoriteGroupNotFound(t *testing.T) {
	groupRepo := applicationtest.NewMockGroupRepository()
	scheduleRepo := applicationtest.NewMockRepository()
	service := NewGroupService(groupRepo, scheduleRepo)

	userID := user.NewUserID()
	groupID := group.NewGroupID()

	// Try to favorite non-existent group
	err := service.FavoriteGroup(context.Background(), userID.String(), groupID.String())
	if !errors.Is(err, group.ErrGroupNotFound) {
		t.Errorf("expected ErrGroupNotFound, got %v", err)
	}
}

func TestUnfavoriteGroup(t *testing.T) {
	groupRepo := applicationtest.NewMockGroupRepository()
	scheduleRepo := applicationtest.NewMockRepository()
	service := NewGroupService(groupRepo, scheduleRepo)

	// Create a group
	cmd := input.CreateGroupCommand{
		Name:  "Project Alpha",
		Owner: "test-user",
	}
	grp, _ := service.CreateGroup(context.Background(), cmd)

	userID := user.NewUserID()

	// Favorite first
	service.FavoriteGroup(context.Background(), userID.String(), grp.ID().String())

	// Now unfavorite
	err := service.UnfavoriteGroup(context.Background(), userID.String(), grp.ID().String())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify it's not favorited
	isFav, err := groupRepo.IsFavorite(context.Background(), userID, grp.ID())
	if err != nil {
		t.Fatalf("expected no error checking favorite, got %v", err)
	}
	if isFav {
		t.Error("expected group to not be favorited")
	}
}

func TestUnfavoriteGroupIdempotent(t *testing.T) {
	groupRepo := applicationtest.NewMockGroupRepository()
	scheduleRepo := applicationtest.NewMockRepository()
	service := NewGroupService(groupRepo, scheduleRepo)

	// Create a group
	cmd := input.CreateGroupCommand{
		Name:  "Project Alpha",
		Owner: "test-user",
	}
	grp, _ := service.CreateGroup(context.Background(), cmd)

	userID := user.NewUserID()

	// Unfavorite without favoriting first - should not error
	err := service.UnfavoriteGroup(context.Background(), userID.String(), grp.ID().String())
	if err != nil {
		t.Errorf("expected no error on unfavoriting non-favorited group, got %v", err)
	}
}

func TestListGroupsWithFavorites(t *testing.T) {
	groupRepo := applicationtest.NewMockGroupRepository()
	scheduleRepo := applicationtest.NewMockRepository()
	service := NewGroupService(groupRepo, scheduleRepo)

	// Create multiple groups
	grp1Cmd := input.CreateGroupCommand{
		Name:  "Alpha",
		Owner: "test-user",
	}
	grp1, _ := service.CreateGroup(context.Background(), grp1Cmd)

	grp2Cmd := input.CreateGroupCommand{
		Name:  "Beta",
		Owner: "test-user",
	}
	grp2, _ := service.CreateGroup(context.Background(), grp2Cmd)

	grp3Cmd := input.CreateGroupCommand{
		Name:  "Gamma",
		Owner: "test-user",
	}
	grp3, _ := service.CreateGroup(context.Background(), grp3Cmd)

	userID := user.NewUserID()

	// Favorite Beta and Gamma
	service.FavoriteGroup(context.Background(), userID.String(), grp2.ID().String())
	service.FavoriteGroup(context.Background(), userID.String(), grp3.ID().String())

	// Get groups with favorites
	query := input.ListGroupsQuery{
		Owner: "test-user",
	}

	groups, favorites, err := service.ListGroupsWithFavorites(context.Background(), query, userID.String())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(groups) != 3 {
		t.Fatalf("expected 3 groups, got %d", len(groups))
	}

	// Check favorites map (keys are strings in the service response)
	if !favorites[grp2.ID().String()] {
		t.Error("expected Beta to be favorited")
	}
	if !favorites[grp3.ID().String()] {
		t.Error("expected Gamma to be favorited")
	}
	if favorites[grp1.ID().String()] {
		t.Error("expected Alpha to not be favorited")
	}

	// Verify sorting: favorites first (Beta, Gamma), then non-favorites (Alpha)
	if groups[0].Name().String() != "Beta" {
		t.Errorf("expected first group to be Beta (favorited), got %v", groups[0].Name().String())
	}
	if groups[1].Name().String() != "Gamma" {
		t.Errorf("expected second group to be Gamma (favorited), got %v", groups[1].Name().String())
	}
	if groups[2].Name().String() != "Alpha" {
		t.Errorf("expected third group to be Alpha (not favorited), got %v", groups[2].Name().String())
	}
}

func TestListGroupsWithFavoritesInvalidUserID(t *testing.T) {
	groupRepo := applicationtest.NewMockGroupRepository()
	scheduleRepo := applicationtest.NewMockRepository()
	service := NewGroupService(groupRepo, scheduleRepo)

	query := input.ListGroupsQuery{
		Owner: "test-user",
	}

	// Try with invalid user ID
	_, _, err := service.ListGroupsWithFavorites(context.Background(), query, "invalid-uuid")
	if err == nil {
		t.Fatal("expected error for invalid user ID")
	}
}
