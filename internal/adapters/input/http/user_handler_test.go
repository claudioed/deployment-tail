package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/claudioed/deployment-tail/api"
	"github.com/claudioed/deployment-tail/internal/adapters/input/http/middleware"
	"github.com/claudioed/deployment-tail/internal/application/ports/input"
	"github.com/claudioed/deployment-tail/internal/domain/user"
	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

// MockUserServiceForHandler is a mock for testing user handlers
type MockUserServiceForHandler struct {
	getUserProfileFunc func(ctx context.Context, userID user.UserID) (*user.User, error)
	listUsersFunc      func(ctx context.Context, filters input.UserListFilters) ([]*user.User, error)
	assignRoleFunc     func(ctx context.Context, adminUserID, targetUserID user.UserID, newRole user.Role) error
}

func (m *MockUserServiceForHandler) GetUserProfile(ctx context.Context, userID user.UserID) (*user.User, error) {
	if m.getUserProfileFunc != nil {
		return m.getUserProfileFunc(ctx, userID)
	}
	return nil, fmt.Errorf("not implemented")
}

func (m *MockUserServiceForHandler) ListUsers(ctx context.Context, filters input.UserListFilters) ([]*user.User, error) {
	if m.listUsersFunc != nil {
		return m.listUsersFunc(ctx, filters)
	}
	return nil, fmt.Errorf("not implemented")
}

func (m *MockUserServiceForHandler) AssignRole(ctx context.Context, adminUserID, targetUserID user.UserID, newRole user.Role) error {
	if m.assignRoleFunc != nil {
		return m.assignRoleFunc(ctx, adminUserID, targetUserID, newRole)
	}
	return fmt.Errorf("not implemented")
}

func (m *MockUserServiceForHandler) AuthenticateWithGoogle(ctx context.Context, code string) (*input.AuthenticationResult, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *MockUserServiceForHandler) RegisterOrUpdateUser(ctx context.Context, googleID, email, name string) (*user.User, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *MockUserServiceForHandler) RefreshUserToken(ctx context.Context, userID user.UserID) (string, error) {
	return "", fmt.Errorf("not implemented")
}

func (m *MockUserServiceForHandler) RevokeUserToken(ctx context.Context, tokenHash string, userID user.UserID) error {
	return fmt.Errorf("not implemented")
}

// Helper function to create test users
func createUserForTest(role string) *user.User {
	googleID, _ := user.NewGoogleID("test-google-id-" + role)
	email, _ := user.NewEmail(role + "@example.com")
	name, _ := user.NewUserName("Test " + role)
	r, _ := user.NewRole(role)
	u, _ := user.NewUser(googleID, email, name, r)
	return u
}

func TestGetMyProfile_Success(t *testing.T) {
	testUser := createUserForTest(user.RoleDeployer)

	handler := NewUserHandler(nil)

	ctx := middleware.UserToContext(context.Background(), testUser)
	req := httptest.NewRequest(http.MethodGet, "/users/me", nil).WithContext(ctx)
	rr := httptest.NewRecorder()

	handler.GetMyProfile(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response["email"] != testUser.Email().String() {
		t.Errorf("expected email %s, got %v", testUser.Email().String(), response["email"])
	}

	if response["role"] != testUser.Role().String() {
		t.Errorf("expected role %s, got %v", testUser.Role().String(), response["role"])
	}
}

func TestGetMyProfile_NoUserInContext(t *testing.T) {
	handler := NewUserHandler(nil)

	req := httptest.NewRequest(http.MethodGet, "/users/me", nil)
	rr := httptest.NewRecorder()

	handler.GetMyProfile(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, rr.Code)
	}
}

func TestGetUserByID_Success(t *testing.T) {
	targetUser := createUserForTest(user.RoleDeployer)

	mockService := &MockUserServiceForHandler{
		getUserProfileFunc: func(ctx context.Context, userID user.UserID) (*user.User, error) {
			if userID.Equals(targetUser.ID()) {
				return targetUser, nil
			}
			return nil, user.ErrUserNotFound{ID: userID.String(), SearchType: "id"}
		},
	}

	handler := NewUserHandler(mockService)

	req := httptest.NewRequest(http.MethodGet, "/users/"+targetUser.ID().String(), nil)
	rr := httptest.NewRecorder()

	targetUUID, _ := uuid.Parse(targetUser.ID().String())
	handler.GetUserByID(rr, req, openapi_types.UUID(targetUUID))

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response["id"] != targetUser.ID().String() {
		t.Errorf("expected ID %s, got %v", targetUser.ID().String(), response["id"])
	}
}

func TestGetUserByID_InvalidID(t *testing.T) {
	handler := NewUserHandler(nil)

	req := httptest.NewRequest(http.MethodGet, "/users/invalid-id", nil)
	rr := httptest.NewRecorder()

	// Pass an invalid UUID (all zeros to simulate parse failure)
	handler.GetUserByID(rr, req, openapi_types.UUID{})

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
}

func TestGetUserByID_UserNotFound(t *testing.T) {
	testUserID := user.NewUserID()

	mockService := &MockUserServiceForHandler{
		getUserProfileFunc: func(ctx context.Context, userID user.UserID) (*user.User, error) {
			return nil, user.ErrUserNotFound{ID: userID.String(), SearchType: "id"}
		},
	}

	handler := NewUserHandler(mockService)

	req := httptest.NewRequest(http.MethodGet, "/users/"+testUserID.String(), nil)
	rr := httptest.NewRecorder()

	testUUID, _ := uuid.Parse(testUserID.String())
	handler.GetUserByID(rr, req, openapi_types.UUID(testUUID))

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, rr.Code)
	}
}

func TestListUsers_Success(t *testing.T) {
	user1 := createUserForTest(user.RoleViewer)
	user2 := createUserForTest(user.RoleDeployer)
	user3 := createUserForTest(user.RoleAdmin)

	mockService := &MockUserServiceForHandler{
		listUsersFunc: func(ctx context.Context, filters input.UserListFilters) ([]*user.User, error) {
			return []*user.User{user1, user2, user3}, nil
		},
	}

	handler := NewUserHandler(mockService)

	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	rr := httptest.NewRecorder()

	handler.ListUsers(rr, req, api.ListUsersParams{})

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	var response []map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(response) != 3 {
		t.Errorf("expected 3 users, got %d", len(response))
	}
}

func TestListUsers_WithRoleFilter(t *testing.T) {
	admin := createUserForTest(user.RoleAdmin)

	mockService := &MockUserServiceForHandler{
		listUsersFunc: func(ctx context.Context, filters input.UserListFilters) ([]*user.User, error) {
			// Verify role filter was applied
			if filters.Role != nil && filters.Role.String() == user.RoleAdmin {
				return []*user.User{admin}, nil
			}
			return []*user.User{}, nil
		},
	}

	handler := NewUserHandler(mockService)

	req := httptest.NewRequest(http.MethodGet, "/users?role=admin", nil)
	rr := httptest.NewRecorder()

	roleAdmin := api.ListUsersParamsRoleAdmin
	handler.ListUsers(rr, req, api.ListUsersParams{Role: &roleAdmin})

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	var response []map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(response) != 1 {
		t.Errorf("expected 1 admin user, got %d", len(response))
	}

	if response[0]["role"] != user.RoleAdmin {
		t.Errorf("expected role admin, got %v", response[0]["role"])
	}
}

func TestListUsers_InvalidRoleFilter(t *testing.T) {
	handler := NewUserHandler(nil)

	req := httptest.NewRequest(http.MethodGet, "/users?role=invalid-role", nil)
	rr := httptest.NewRecorder()

	invalidRole := api.ListUsersParamsRole("invalid-role")
	handler.ListUsers(rr, req, api.ListUsersParams{Role: &invalidRole})

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
}

func TestAssignRole_Success(t *testing.T) {
	admin := createUserForTest(user.RoleAdmin)
	targetUser := createUserForTest(user.RoleViewer)

	mockService := &MockUserServiceForHandler{
		assignRoleFunc: func(ctx context.Context, adminUserID, targetUserID user.UserID, newRole user.Role) error {
			if adminUserID.Equals(admin.ID()) && targetUserID.Equals(targetUser.ID()) {
				return nil
			}
			return fmt.Errorf("invalid users")
		},
		getUserProfileFunc: func(ctx context.Context, userID user.UserID) (*user.User, error) {
			if userID.Equals(targetUser.ID()) {
				// Return user with updated role
				googleID, _ := user.NewGoogleID("test-google-id")
				email, _ := user.NewEmail("viewer@example.com")
				name, _ := user.NewUserName("Test viewer")
				newRole, _ := user.NewRole(user.RoleDeployer)
				u, _ := user.NewUser(googleID, email, name, newRole)
				return u, nil
			}
			return nil, user.ErrUserNotFound{ID: userID.String(), SearchType: "id"}
		},
	}

	handler := NewUserHandler(mockService)

	requestBody := map[string]string{
		"role": user.RoleDeployer,
	}
	bodyBytes, _ := json.Marshal(requestBody)

	ctx := middleware.UserToContext(context.Background(), admin)
	req := httptest.NewRequest(http.MethodPut, "/users/"+targetUser.ID().String()+"/role", bytes.NewReader(bodyBytes)).WithContext(ctx)
	rr := httptest.NewRecorder()

	targetUUID, _ := uuid.Parse(targetUser.ID().String())
	handler.AssignRole(rr, req, openapi_types.UUID(targetUUID))

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response["role"] != user.RoleDeployer {
		t.Errorf("expected role deployer, got %v", response["role"])
	}
}

func TestAssignRole_NoUserInContext(t *testing.T) {
	handler := NewUserHandler(nil)

	targetUserID := user.NewUserID()
	requestBody := map[string]string{"role": user.RoleDeployer}
	bodyBytes, _ := json.Marshal(requestBody)

	req := httptest.NewRequest(http.MethodPut, "/users/"+targetUserID.String()+"/role", bytes.NewReader(bodyBytes))
	rr := httptest.NewRecorder()

	targetUUID, _ := uuid.Parse(targetUserID.String())
	handler.AssignRole(rr, req, openapi_types.UUID(targetUUID))

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, rr.Code)
	}
}

func TestAssignRole_InvalidUserID(t *testing.T) {
	admin := createUserForTest(user.RoleAdmin)

	handler := NewUserHandler(nil)

	requestBody := map[string]string{"role": user.RoleDeployer}
	bodyBytes, _ := json.Marshal(requestBody)

	ctx := middleware.UserToContext(context.Background(), admin)
	req := httptest.NewRequest(http.MethodPut, "/users/invalid-id/role", bytes.NewReader(bodyBytes)).WithContext(ctx)
	rr := httptest.NewRecorder()

	// Pass an invalid UUID (all zeros to simulate parse failure)
	handler.AssignRole(rr, req, openapi_types.UUID{})

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
}

func TestAssignRole_InvalidRole(t *testing.T) {
	admin := createUserForTest(user.RoleAdmin)
	targetUser := createUserForTest(user.RoleViewer)

	handler := NewUserHandler(nil)

	requestBody := map[string]string{"role": "invalid-role"}
	bodyBytes, _ := json.Marshal(requestBody)

	ctx := middleware.UserToContext(context.Background(), admin)
	req := httptest.NewRequest(http.MethodPut, "/users/"+targetUser.ID().String()+"/role", bytes.NewReader(bodyBytes)).WithContext(ctx)
	rr := httptest.NewRecorder()

	targetUUID, _ := uuid.Parse(targetUser.ID().String())
	handler.AssignRole(rr, req, openapi_types.UUID(targetUUID))

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
}

func TestAssignRole_Unauthorized(t *testing.T) {
	deployer := createUserForTest(user.RoleDeployer)
	targetUser := createUserForTest(user.RoleViewer)

	mockService := &MockUserServiceForHandler{
		assignRoleFunc: func(ctx context.Context, adminUserID, targetUserID user.UserID, newRole user.Role) error {
			return user.ErrUnauthorized{
				UserID:    adminUserID.String(),
				Operation: "assign role",
				Reason:    "requires admin role",
			}
		},
	}

	handler := NewUserHandler(mockService)

	requestBody := map[string]string{"role": user.RoleDeployer}
	bodyBytes, _ := json.Marshal(requestBody)

	ctx := middleware.UserToContext(context.Background(), deployer)
	req := httptest.NewRequest(http.MethodPut, "/users/"+targetUser.ID().String()+"/role", bytes.NewReader(bodyBytes)).WithContext(ctx)
	rr := httptest.NewRecorder()

	targetUUID, _ := uuid.Parse(targetUser.ID().String())
	handler.AssignRole(rr, req, openapi_types.UUID(targetUUID))

	if rr.Code != http.StatusForbidden {
		t.Errorf("expected status %d, got %d", http.StatusForbidden, rr.Code)
	}
}

func TestAssignRole_TargetUserNotFound(t *testing.T) {
	admin := createUserForTest(user.RoleAdmin)
	nonExistentUserID := user.NewUserID()

	mockService := &MockUserServiceForHandler{
		assignRoleFunc: func(ctx context.Context, adminUserID, targetUserID user.UserID, newRole user.Role) error {
			return user.ErrUserNotFound{ID: targetUserID.String(), SearchType: "id"}
		},
	}

	handler := NewUserHandler(mockService)

	requestBody := map[string]string{"role": user.RoleDeployer}
	bodyBytes, _ := json.Marshal(requestBody)

	ctx := middleware.UserToContext(context.Background(), admin)
	req := httptest.NewRequest(http.MethodPut, "/users/"+nonExistentUserID.String()+"/role", bytes.NewReader(bodyBytes)).WithContext(ctx)
	rr := httptest.NewRecorder()

	nonExistentUUID, _ := uuid.Parse(nonExistentUserID.String())
	handler.AssignRole(rr, req, openapi_types.UUID(nonExistentUUID))

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, rr.Code)
	}
}

func TestUserToResponse_WithLastLogin(t *testing.T) {
	testUser := createUserForTest(user.RoleDeployer)
	testUser.RecordLogin()

	handler := NewUserHandler(nil)

	response := handler.userToResponse(testUser)

	if response["id"] != testUser.ID().String() {
		t.Error("ID mismatch in response")
	}

	if response["email"] != testUser.Email().String() {
		t.Error("Email mismatch in response")
	}

	if response["role"] != testUser.Role().String() {
		t.Error("Role mismatch in response")
	}

	if response["last_login_at"] == nil {
		t.Error("Expected last_login_at to be present")
	}

	if response["created_at"] == nil {
		t.Error("Expected created_at to be present")
	}

	if response["updated_at"] == nil {
		t.Error("Expected updated_at to be present")
	}
}

func TestUserToResponse_WithoutLastLogin(t *testing.T) {
	testUser := createUserForTest(user.RoleViewer)

	handler := NewUserHandler(nil)

	response := handler.userToResponse(testUser)

	if response["last_login_at"] != nil {
		t.Error("Expected last_login_at to be absent")
	}
}
