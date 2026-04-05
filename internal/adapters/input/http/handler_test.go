package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/claudioed/deployment-tail/api"
	"github.com/claudioed/deployment-tail/internal/adapters/input/http/middleware"
	"github.com/claudioed/deployment-tail/internal/application"
	"github.com/claudioed/deployment-tail/internal/application/ports/input"
	"github.com/claudioed/deployment-tail/internal/domain/group"
	"github.com/claudioed/deployment-tail/internal/domain/schedule"
	"github.com/claudioed/deployment-tail/internal/domain/user"
	"github.com/google/uuid"

	openapi_types "github.com/oapi-codegen/runtime/types"
)

// Test helpers

func setupTestServices(t *testing.T) (input.ScheduleService, input.GroupService, input.UserService, *application.MockUserRepository) {
	scheduleRepo := application.NewMockRepository()
	groupRepo := application.NewMockGroupRepository()
	userRepo := application.NewMockUserRepository()

	scheduleService := application.NewScheduleService(scheduleRepo, userRepo)
	groupService := application.NewGroupService(groupRepo, scheduleRepo)
	userService := application.NewUserService(userRepo, nil, nil, nil) // OAuth, JWT, and RevocationStore not needed for tests

	return scheduleService, groupService, userService, userRepo
}

func createTestDeployer(t *testing.T, userRepo *application.MockUserRepository) *user.User {
	t.Helper()
	googleID, _ := user.NewGoogleID("deployer123")
	email, _ := user.NewEmail("deployer@example.com")
	name, _ := user.NewUserName("Test Deployer")
	role, _ := user.NewRole(user.RoleDeployer)
	u, _ := user.NewUser(googleID, email, name, role)
	userRepo.Create(context.Background(), u)
	return u
}

func createTestSchedule(t *testing.T, scheduleService input.ScheduleService, userRepo *application.MockUserRepository) *schedule.Schedule {
	t.Helper()
	return createTestScheduleWithUser(t, scheduleService, userRepo).schedule
}

// createTestScheduleWithUser creates a test schedule and returns both schedule and user
func createTestScheduleWithUser(t *testing.T, scheduleService input.ScheduleService, userRepo *application.MockUserRepository) struct {
	schedule *schedule.Schedule
	user     *user.User
} {
	t.Helper()

	// Create a test user for the schedule
	googleID, _ := user.NewGoogleID("test-user")
	email, _ := user.NewEmail("test@example.com")
	name, _ := user.NewUserName("Test User")
	role, _ := user.NewRole(user.RoleDeployer)
	testUser, _ := user.NewUser(googleID, email, name, role)

	// Add user to repository if provided
	if userRepo != nil {
		userRepo.Create(context.Background(), testUser)
	}

	sch, err := scheduleService.CreateSchedule(context.Background(), input.CreateScheduleCommand{
		ScheduledAt:  time.Now().Add(24 * time.Hour),
		ServiceName:  "test-service",
		Environments: []string{"production"},
		Description:  "Test deployment",
		Owners:       []string{"test-user"},
		RollbackPlan: "Rollback instructions",
	}, testUser.ID())
	if err != nil {
		t.Fatalf("failed to create test schedule: %v", err)
	}

	return struct {
		schedule *schedule.Schedule
		user     *user.User
	}{sch, testUser}
}

func createTestGroup(t *testing.T, groupService input.GroupService, name, owner string) *group.Group {
	t.Helper()

	grp, err := groupService.CreateGroup(context.Background(), input.CreateGroupCommand{
		Name:        name,
		Description: "Test group",
		Owner:       owner,
	})
	if err != nil {
		t.Fatalf("failed to create test group: %v", err)
	}

	return grp
}

func stringPtr(s string) *string {
	return &s
}

// Group Handler Tests

func TestListGroups(t *testing.T) {
	scheduleService, groupService, _, _ := setupTestServices(t)
	handler := NewGroupHandler(groupService, scheduleService)

	// Create test groups
	createTestGroup(t, groupService, "Project Alpha", "test-user")
	createTestGroup(t, groupService, "Project Beta", "test-user")
	createTestGroup(t, groupService, "Other Project", "other-user")

	req := httptest.NewRequest(http.MethodGet, "/groups?owner=test-user", nil)
	w := httptest.NewRecorder()

	handler.ListGroups(w, req, api.ListGroupsParams{Owner: "test-user"})

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var groups []api.Group
	json.NewDecoder(w.Body).Decode(&groups)

	if len(groups) != 2 {
		t.Errorf("expected 2 groups for test-user, got %d", len(groups))
	}
}

func TestListGroupsRequiresOwner(t *testing.T) {
	scheduleService, groupService, _, _ := setupTestServices(t)
	handler := NewGroupHandler(groupService, scheduleService)

	req := httptest.NewRequest(http.MethodGet, "/groups", nil)
	w := httptest.NewRecorder()

	handler.ListGroups(w, req, api.ListGroupsParams{Owner: ""})

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestCreateGroup(t *testing.T) {
	scheduleService, groupService, _, _ := setupTestServices(t)
	handler := NewGroupHandler(groupService, scheduleService)

	body := api.CreateGroupRequest{
		Name:        "Project Alpha",
		Description: stringPtr("All schedules for Project Alpha"),
		Owner:       "test-user",
	}

	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/groups", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.CreateGroup(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d: %s", w.Code, w.Body.String())
	}

	var group api.Group
	json.NewDecoder(w.Body).Decode(&group)

	if group.Name != "Project Alpha" {
		t.Errorf("expected name 'Project Alpha', got %s", group.Name)
	}

	if group.Owner != "test-user" {
		t.Errorf("expected owner 'test-user', got %s", group.Owner)
	}
}

func TestCreateGroupWithInvalidName(t *testing.T) {
	scheduleService, groupService, _, _ := setupTestServices(t)
	handler := NewGroupHandler(groupService, scheduleService)

	body := api.CreateGroupRequest{
		Name:  "", // Empty name
		Owner: "test-user",
	}

	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/groups", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.CreateGroup(w, req)

	if w.Code == http.StatusCreated {
		t.Fatal("expected error for empty name")
	}
}

func TestGetGroup(t *testing.T) {
	scheduleService, groupService, _, _ := setupTestServices(t)
	handler := NewGroupHandler(groupService, scheduleService)

	created := createTestGroup(t, groupService, "Project Alpha", "test-user")

	req := httptest.NewRequest(http.MethodGet, "/groups/"+created.ID().String(), nil)
	w := httptest.NewRecorder()

	uuid, _ := uuid.Parse(created.ID().String())
	handler.GetGroup(w, req, uuid)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var group api.Group
	json.NewDecoder(w.Body).Decode(&group)

	if group.Id.String() != created.ID().String() {
		t.Errorf("expected ID %s, got %s", created.ID().String(), group.Id.String())
	}
}

func TestGetGroupNotFound(t *testing.T) {
	scheduleService, groupService, _, _ := setupTestServices(t)
	handler := NewGroupHandler(groupService, scheduleService)

	fakeID := group.NewGroupID()

	req := httptest.NewRequest(http.MethodGet, "/groups/"+fakeID.String(), nil)
	w := httptest.NewRecorder()

	uuid, _ := uuid.Parse(fakeID.String())
	handler.GetGroup(w, req, uuid)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w.Code)
	}
}

func TestUpdateGroup(t *testing.T) {
	scheduleService, groupService, _, _ := setupTestServices(t)
	handler := NewGroupHandler(groupService, scheduleService)

	created := createTestGroup(t, groupService, "Project Alpha", "test-user")

	body := api.UpdateGroupRequest{
		Name:        "Project Beta",
		Description: stringPtr("Updated description"),
	}

	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPut, "/groups/"+created.ID().String(), bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	uuid, _ := uuid.Parse(created.ID().String())
	handler.UpdateGroup(w, req, uuid)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var group api.Group
	json.NewDecoder(w.Body).Decode(&group)

	if group.Name != "Project Beta" {
		t.Errorf("expected name 'Project Beta', got %s", group.Name)
	}
}

func TestDeleteGroup(t *testing.T) {
	scheduleService, groupService, _, _ := setupTestServices(t)
	handler := NewGroupHandler(groupService, scheduleService)

	created := createTestGroup(t, groupService, "Project Alpha", "test-user")

	req := httptest.NewRequest(http.MethodDelete, "/groups/"+created.ID().String(), nil)
	w := httptest.NewRecorder()

	uuid, _ := uuid.Parse(created.ID().String())
	handler.DeleteGroup(w, req, uuid)

	if w.Code != http.StatusNoContent {
		t.Errorf("expected status 204, got %d", w.Code)
	}

	// Verify group is deleted
	req2 := httptest.NewRequest(http.MethodGet, "/groups/"+created.ID().String(), nil)
	w2 := httptest.NewRecorder()

	handler.GetGroup(w2, req2, uuid)

	if w2.Code != http.StatusNotFound {
		t.Error("expected 404 when getting deleted group")
	}
}

func TestAssignScheduleToGroups(t *testing.T) {
	scheduleService, groupService, _, userRepo := setupTestServices(t)
	handler := NewGroupHandler(groupService, scheduleService)

	group1 := createTestGroup(t, groupService, "Project Alpha", "test-user")
	group2 := createTestGroup(t, groupService, "Project Beta", "test-user")
	sch := createTestSchedule(t, scheduleService, userRepo)

	body := api.AssignScheduleRequest{
		GroupIds:   []openapi_types.UUID{parseUUID(t, group1.ID().String()), parseUUID(t, group2.ID().String())},
		AssignedBy: stringPtr("test-user"),
	}

	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/schedules/"+sch.ID().String()+"/groups", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	schedUUID := parseUUID(t, sch.ID().String())
	handler.AssignScheduleToGroups(w, req, schedUUID)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	// Verify assignment
	req2 := httptest.NewRequest(http.MethodGet, "/schedules/"+sch.ID().String()+"/groups", nil)
	w2 := httptest.NewRecorder()

	handler.GetGroupsForSchedule(w2, req2, schedUUID)

	var groups []api.Group
	json.NewDecoder(w2.Body).Decode(&groups)

	if len(groups) != 2 {
		t.Errorf("expected 2 groups assigned, got %d", len(groups))
	}
}

func TestUnassignScheduleFromGroup(t *testing.T) {
	scheduleService, groupService, _, userRepo := setupTestServices(t)
	handler := NewGroupHandler(groupService, scheduleService)

	grp := createTestGroup(t, groupService, "Project Alpha", "test-user")
	sch := createTestSchedule(t, scheduleService, userRepo)

	// First assign
	assignBody := api.AssignScheduleRequest{
		GroupIds:   []openapi_types.UUID{parseUUID(t, grp.ID().String())},
		AssignedBy: stringPtr("test-user"),
	}
	assignBytes, _ := json.Marshal(assignBody)
	req1 := httptest.NewRequest(http.MethodPost, "/schedules/"+sch.ID().String()+"/groups", bytes.NewReader(assignBytes))
	req1.Header.Set("Content-Type", "application/json")
	w1 := httptest.NewRecorder()

	schedUUID := parseUUID(t, sch.ID().String())
	grpUUID := parseUUID(t, grp.ID().String())

	handler.AssignScheduleToGroups(w1, req1, schedUUID)

	// Then unassign
	req2 := httptest.NewRequest(http.MethodDelete, "/schedules/"+sch.ID().String()+"/groups/"+grp.ID().String(), nil)
	w2 := httptest.NewRecorder()

	handler.UnassignScheduleFromGroup(w2, req2, schedUUID, grpUUID)

	if w2.Code != http.StatusNoContent {
		t.Errorf("expected status 204, got %d", w2.Code)
	}

	// Verify unassignment
	req3 := httptest.NewRequest(http.MethodGet, "/schedules/"+sch.ID().String()+"/groups", nil)
	w3 := httptest.NewRecorder()
	handler.GetGroupsForSchedule(w3, req3, schedUUID)

	var groups []api.Group
	json.NewDecoder(w3.Body).Decode(&groups)

	if len(groups) != 0 {
		t.Errorf("expected 0 groups after unassignment, got %d", len(groups))
	}
}

func TestGetGroupsForSchedule(t *testing.T) {
	scheduleService, groupService, _, userRepo := setupTestServices(t)
	handler := NewGroupHandler(groupService, scheduleService)

	group1 := createTestGroup(t, groupService, "Project Alpha", "test-user")
	group2 := createTestGroup(t, groupService, "Project Beta", "test-user")
	sch := createTestSchedule(t, scheduleService, userRepo)

	// Assign schedule to groups
	body := api.AssignScheduleRequest{
		GroupIds:   []openapi_types.UUID{parseUUID(t, group1.ID().String()), parseUUID(t, group2.ID().String())},
		AssignedBy: stringPtr("test-user"),
	}
	bodyBytes, _ := json.Marshal(body)
	req1 := httptest.NewRequest(http.MethodPost, "/schedules/"+sch.ID().String()+"/groups", bytes.NewReader(bodyBytes))
	req1.Header.Set("Content-Type", "application/json")
	w1 := httptest.NewRecorder()

	schedUUID := parseUUID(t, sch.ID().String())
	handler.AssignScheduleToGroups(w1, req1, schedUUID)

	// Get groups for schedule
	req2 := httptest.NewRequest(http.MethodGet, "/schedules/"+sch.ID().String()+"/groups", nil)
	w2 := httptest.NewRecorder()

	handler.GetGroupsForSchedule(w2, req2, schedUUID)

	if w2.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w2.Code)
	}

	var groups []api.Group
	json.NewDecoder(w2.Body).Decode(&groups)

	if len(groups) != 2 {
		t.Errorf("expected 2 groups, got %d", len(groups))
	}
}

func TestGetSchedulesInGroup(t *testing.T) {
	scheduleService, groupService, _, userRepo := setupTestServices(t)
	handler := NewGroupHandler(groupService, scheduleService)

	grp := createTestGroup(t, groupService, "Project Alpha", "test-user")
	sch1 := createTestSchedule(t, scheduleService, userRepo)
	sch2 := createTestSchedule(t, scheduleService, userRepo)

	grpUUID := parseUUID(t, grp.ID().String())

	// Assign schedules to group
	body := api.AssignScheduleRequest{
		GroupIds:   []openapi_types.UUID{grpUUID},
		AssignedBy: stringPtr("test-user"),
	}

	// Assign first schedule
	bodyBytes, _ := json.Marshal(body)
	req1 := httptest.NewRequest(http.MethodPost, "/schedules/"+sch1.ID().String()+"/groups", bytes.NewReader(bodyBytes))
	req1.Header.Set("Content-Type", "application/json")
	w1 := httptest.NewRecorder()
	handler.AssignScheduleToGroups(w1, req1, parseUUID(t, sch1.ID().String()))

	// Assign second schedule
	req2 := httptest.NewRequest(http.MethodPost, "/schedules/"+sch2.ID().String()+"/groups", bytes.NewReader(bodyBytes))
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	handler.AssignScheduleToGroups(w2, req2, parseUUID(t, sch2.ID().String()))

	// Get schedules in group
	req3 := httptest.NewRequest(http.MethodGet, "/groups/"+grp.ID().String()+"/schedules", nil)
	w3 := httptest.NewRecorder()

	handler.GetSchedulesInGroup(w3, req3, grpUUID)

	if w3.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w3.Code)
	}

	var schedules []api.Schedule
	json.NewDecoder(w3.Body).Decode(&schedules)

	if len(schedules) != 2 {
		t.Errorf("expected 2 schedules, got %d", len(schedules))
	}
}

func TestBulkAssignSchedules(t *testing.T) {
	scheduleService, groupService, _, userRepo := setupTestServices(t)
	handler := NewGroupHandler(groupService, scheduleService)

	grp := createTestGroup(t, groupService, "Project Alpha", "test-user")
	sch1 := createTestSchedule(t, scheduleService, userRepo)
	sch2 := createTestSchedule(t, scheduleService, userRepo)
	sch3 := createTestSchedule(t, scheduleService, userRepo)

	body := api.BulkAssignRequest{
		ScheduleIds: []openapi_types.UUID{
			parseUUID(t, sch1.ID().String()),
			parseUUID(t, sch2.ID().String()),
			parseUUID(t, sch3.ID().String()),
		},
		AssignedBy: stringPtr("test-user"),
	}

	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/groups/"+grp.ID().String()+"/schedules/bulk-assign", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	grpUUID := parseUUID(t, grp.ID().String())
	handler.BulkAssignSchedules(w, req, grpUUID)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	// Verify all schedules were assigned
	req2 := httptest.NewRequest(http.MethodGet, "/groups/"+grp.ID().String()+"/schedules", nil)
	w2 := httptest.NewRecorder()
	handler.GetSchedulesInGroup(w2, req2, grpUUID)

	var schedules []api.Schedule
	json.NewDecoder(w2.Body).Decode(&schedules)

	if len(schedules) != 3 {
		t.Errorf("expected 3 schedules assigned, got %d", len(schedules))
	}
}

// Schedule Handler Tests

func TestCreateSchedule(t *testing.T) {
	scheduleService, groupService, userService, userRepo := setupTestServices(t)
	handler := NewScheduleHandler(scheduleService, groupService, userService)

	// Create authenticated user and add to repository
	googleID, _ := user.NewGoogleID("deployer123")
	email, _ := user.NewEmail("deployer@example.com")
	name, _ := user.NewUserName("Test Deployer")
	role, _ := user.NewRole(user.RoleDeployer)
	testUser, _ := user.NewUser(googleID, email, name, role)
	userRepo.Create(context.Background(), testUser)

	scheduledAt := time.Now().Add(24 * time.Hour)
	body := api.CreateScheduleRequest{
		ScheduledAt:  scheduledAt,
		ServiceName:  "test-service",
		Environments: []api.CreateScheduleRequestEnvironments{"production"},
		Description:  stringPtr("Test deployment"),
		Owners:       []string{"test-user"},
		RollbackPlan: stringPtr("Rollback instructions"),
	}

	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/schedules", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	// Add authenticated user to context using middleware helper
	ctx := middleware.UserToContext(req.Context(), testUser)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()

	handler.CreateSchedule(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d: %s", w.Code, w.Body.String())
	}

	var schedule api.Schedule
	json.NewDecoder(w.Body).Decode(&schedule)

	if schedule.ServiceName != "test-service" {
		t.Errorf("expected service 'test-service', got %s", schedule.ServiceName)
	}
}

func TestGetSchedule(t *testing.T) {
	scheduleService, groupService, userService, userRepo := setupTestServices(t)
	handler := NewScheduleHandler(scheduleService, groupService, userService)

	sch := createTestSchedule(t, scheduleService, userRepo)

	req := httptest.NewRequest(http.MethodGet, "/schedules/"+sch.ID().String(), nil)
	w := httptest.NewRecorder()

	uuid := parseUUID(t, sch.ID().String())
	handler.GetSchedule(w, req, uuid)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var schedule api.Schedule
	json.NewDecoder(w.Body).Decode(&schedule)

	if schedule.Id.String() != sch.ID().String() {
		t.Errorf("expected ID %s, got %s", sch.ID().String(), schedule.Id.String())
	}

	// Should include empty groups array
	if schedule.Groups == nil {
		t.Error("expected groups array to be present")
	}
}

func TestGetScheduleWithGroups(t *testing.T) {
	scheduleService, groupService, userService, userRepo := setupTestServices(t)
	scheduleHandler := NewScheduleHandler(scheduleService, groupService, userService)
	groupHandler := NewGroupHandler(groupService, scheduleService)

	sch := createTestSchedule(t, scheduleService, userRepo)
	grp := createTestGroup(t, groupService, "Project Alpha", "test-user")

	// Assign schedule to group
	assignBody := api.AssignScheduleRequest{
		GroupIds:   []openapi_types.UUID{parseUUID(t, grp.ID().String())},
		AssignedBy: stringPtr("test-user"),
	}
	assignBytes, _ := json.Marshal(assignBody)
	req1 := httptest.NewRequest(http.MethodPost, "/schedules/"+sch.ID().String()+"/groups", bytes.NewReader(assignBytes))
	req1.Header.Set("Content-Type", "application/json")
	w1 := httptest.NewRecorder()
	groupHandler.AssignScheduleToGroups(w1, req1, parseUUID(t, sch.ID().String()))

	// Get schedule with groups
	req2 := httptest.NewRequest(http.MethodGet, "/schedules/"+sch.ID().String(), nil)
	w2 := httptest.NewRecorder()

	scheduleHandler.GetSchedule(w2, req2, parseUUID(t, sch.ID().String()))

	var schedule api.Schedule
	json.NewDecoder(w2.Body).Decode(&schedule)

	if schedule.Groups == nil || len(*schedule.Groups) != 1 {
		t.Errorf("expected 1 group in schedule, got %v", schedule.Groups)
	}

	if len(*schedule.Groups) > 0 {
		group := (*schedule.Groups)[0]
		if group.Name != "Project Alpha" {
			t.Errorf("expected group name 'Project Alpha', got %s", group.Name)
		}
	}
}

func TestListSchedules(t *testing.T) {
	scheduleService, groupService, userService, userRepo := setupTestServices(t)
	handler := NewScheduleHandler(scheduleService, groupService, userService)

	// Create test schedules
	createTestSchedule(t, scheduleService, userRepo)
	createTestSchedule(t, scheduleService, userRepo)
	createTestSchedule(t, scheduleService, userRepo)

	req := httptest.NewRequest(http.MethodGet, "/schedules?owner=test-user", nil)
	w := httptest.NewRecorder()

	owners := []string{"test-user"}
	handler.ListSchedules(w, req, api.ListSchedulesParams{Owner: &owners})

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var schedules []api.Schedule
	json.NewDecoder(w.Body).Decode(&schedules)

	if len(schedules) != 3 {
		t.Errorf("expected 3 schedules, got %d", len(schedules))
	}

	// Each schedule should have groups array (even if empty)
	for _, sch := range schedules {
		if sch.Groups == nil {
			t.Error("expected groups array to be present in schedule")
		}
	}
}

func TestDeleteSchedule(t *testing.T) {
	scheduleService, groupService, userService, userRepo := setupTestServices(t)
	handler := NewScheduleHandler(scheduleService, groupService, userService)

	// Create schedule with user and add user to repository
	result := createTestScheduleWithUser(t, scheduleService, userRepo)
	sch := result.schedule
	testUser := result.user

	req := httptest.NewRequest(http.MethodDelete, "/schedules/"+sch.ID().String(), nil)

	// Add authenticated user to context
	ctx := middleware.UserToContext(req.Context(), testUser)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()

	uuid := parseUUID(t, sch.ID().String())
	handler.DeleteSchedule(w, req, uuid)

	if w.Code != http.StatusNoContent {
		t.Errorf("expected status 204, got %d: %s", w.Code, w.Body.String())
	}

	// Verify schedule is deleted
	req2 := httptest.NewRequest(http.MethodGet, "/schedules/"+sch.ID().String(), nil)
	w2 := httptest.NewRecorder()

	handler.GetSchedule(w2, req2, uuid)

	if w2.Code != http.StatusNotFound {
		t.Error("expected 404 when getting deleted schedule")
	}
}

// Favorite groups handler tests

func TestFavoriteGroup(t *testing.T) {
	scheduleService, groupService, _, _ := setupTestServices(t)
	handler := NewGroupHandler(groupService, scheduleService)

	// Create test group
	grp := createTestGroup(t, groupService, "Project Alpha", "test-user")

	// Create test user for authentication context
	testUser := user.NewUserID()
	googleID, _ := user.NewGoogleID("test-user")
	email, _ := user.NewEmail("test@example.com")
	name, _ := user.NewUserName("Test User")
	role, _ := user.NewRole(user.RoleViewer)
	u := user.Reconstitute(testUser, googleID, email, name, role, nil, time.Now(), time.Now())

	// Create request with authentication context
	req := httptest.NewRequest(http.MethodPost, "/groups/"+grp.ID().String()+"/favorite", nil)
	ctx := middleware.UserToContext(req.Context(), u)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	groupUUID := parseUUID(t, grp.ID().String())
	handler.FavoriteGroup(w, req, groupUUID)

	if w.Code != http.StatusNoContent {
		t.Errorf("expected status 204, got %d: %s", w.Code, w.Body.String())
	}
}

func TestFavoriteGroupUnauthenticated(t *testing.T) {
	scheduleService, groupService, _, _ := setupTestServices(t)
	handler := NewGroupHandler(groupService, scheduleService)

	// Create test group
	grp := createTestGroup(t, groupService, "Project Alpha", "test-user")

	// Create request WITHOUT authentication context
	req := httptest.NewRequest(http.MethodPost, "/groups/"+grp.ID().String()+"/favorite", nil)
	w := httptest.NewRecorder()

	groupUUID := parseUUID(t, grp.ID().String())
	handler.FavoriteGroup(w, req, groupUUID)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", w.Code)
	}
}

func TestFavoriteGroupNotFound(t *testing.T) {
	scheduleService, groupService, _, _ := setupTestServices(t)
	handler := NewGroupHandler(groupService, scheduleService)

	// Create test user for authentication context
	testUser := user.NewUserID()
	googleID, _ := user.NewGoogleID("test-user")
	email, _ := user.NewEmail("test@example.com")
	name, _ := user.NewUserName("Test User")
	role, _ := user.NewRole(user.RoleViewer)
	u := user.Reconstitute(testUser, googleID, email, name, role, nil, time.Now(), time.Now())

	// Use non-existent group ID
	fakeID := group.NewGroupID()

	req := httptest.NewRequest(http.MethodPost, "/groups/"+fakeID.String()+"/favorite", nil)
	ctx := middleware.UserToContext(req.Context(), u)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	groupUUID := parseUUID(t, fakeID.String())
	handler.FavoriteGroup(w, req, groupUUID)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w.Code)
	}
}

func TestUnfavoriteGroup(t *testing.T) {
	scheduleService, groupService, _, _ := setupTestServices(t)
	handler := NewGroupHandler(groupService, scheduleService)

	// Create test group
	grp := createTestGroup(t, groupService, "Project Alpha", "test-user")

	// Create test user for authentication context
	testUser := user.NewUserID()
	googleID, _ := user.NewGoogleID("test-user")
	email, _ := user.NewEmail("test@example.com")
	name, _ := user.NewUserName("Test User")
	role, _ := user.NewRole(user.RoleViewer)
	u := user.Reconstitute(testUser, googleID, email, name, role, nil, time.Now(), time.Now())

	// First favorite the group
	groupService.FavoriteGroup(context.Background(), testUser.String(), grp.ID().String())

	// Now unfavorite it
	req := httptest.NewRequest(http.MethodDelete, "/groups/"+grp.ID().String()+"/favorite", nil)
	ctx := middleware.UserToContext(req.Context(), u)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	groupUUID := parseUUID(t, grp.ID().String())
	handler.UnfavoriteGroup(w, req, groupUUID)

	if w.Code != http.StatusNoContent {
		t.Errorf("expected status 204, got %d: %s", w.Code, w.Body.String())
	}
}

func TestUnfavoriteGroupUnauthenticated(t *testing.T) {
	scheduleService, groupService, _, _ := setupTestServices(t)
	handler := NewGroupHandler(groupService, scheduleService)

	// Create test group
	grp := createTestGroup(t, groupService, "Project Alpha", "test-user")

	// Create request WITHOUT authentication context
	req := httptest.NewRequest(http.MethodDelete, "/groups/"+grp.ID().String()+"/favorite", nil)
	w := httptest.NewRecorder()

	groupUUID := parseUUID(t, grp.ID().String())
	handler.UnfavoriteGroup(w, req, groupUUID)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", w.Code)
	}
}

func TestGetGroupsWithFavoriteStatus(t *testing.T) {
	scheduleService, groupService, _, _ := setupTestServices(t)
	handler := NewGroupHandler(groupService, scheduleService)

	// Create multiple groups
	grp1 := createTestGroup(t, groupService, "Alpha", "test-user")
	grp2 := createTestGroup(t, groupService, "Beta", "test-user")
	grp3 := createTestGroup(t, groupService, "Gamma", "test-user")

	// Create test user for authentication context
	testUser := user.NewUserID()
	googleID, _ := user.NewGoogleID("test-user")
	email, _ := user.NewEmail("test@example.com")
	name, _ := user.NewUserName("Test User")
	role, _ := user.NewRole(user.RoleViewer)
	u := user.Reconstitute(testUser, googleID, email, name, role, nil, time.Now(), time.Now())

	// Favorite Beta and Gamma
	groupService.FavoriteGroup(context.Background(), testUser.String(), grp2.ID().String())
	groupService.FavoriteGroup(context.Background(), testUser.String(), grp3.ID().String())

	// Get groups with authentication
	req := httptest.NewRequest(http.MethodGet, "/groups?owner=test-user", nil)
	ctx := middleware.UserToContext(req.Context(), u)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.ListGroups(w, req, api.ListGroupsParams{Owner: "test-user"})

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var groups []api.Group
	json.NewDecoder(w.Body).Decode(&groups)

	if len(groups) != 3 {
		t.Fatalf("expected 3 groups, got %d", len(groups))
	}

	// Check isFavorite fields
	for _, grp := range groups {
		if grp.Id.String() == grp2.ID().String() {
			if !*grp.IsFavorite {
				t.Error("expected Beta to be favorited")
			}
		}
		if grp.Id.String() == grp3.ID().String() {
			if !*grp.IsFavorite {
				t.Error("expected Gamma to be favorited")
			}
		}
		if grp.Id.String() == grp1.ID().String() {
			if grp.IsFavorite != nil && *grp.IsFavorite {
				t.Error("expected Alpha to not be favorited")
			}
		}
	}

	// Verify sorting: favorites first (Beta, Gamma), then non-favorites (Alpha)
	if groups[0].Name != "Beta" {
		t.Errorf("expected first group to be Beta (favorited), got %s", groups[0].Name)
	}
	if groups[1].Name != "Gamma" {
		t.Errorf("expected second group to be Gamma (favorited), got %s", groups[1].Name)
	}
	if groups[2].Name != "Alpha" {
		t.Errorf("expected third group to be Alpha (not favorited), got %s", groups[2].Name)
	}
}

// Helper function
func parseUUID(t *testing.T, s string) openapi_types.UUID {
	t.Helper()
	uuid, err := uuid.Parse(s)
	if err != nil {
		t.Fatalf("failed to parse UUID: %v", err)
	}
	return uuid
}
