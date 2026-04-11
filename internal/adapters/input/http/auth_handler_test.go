package http

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/claudioed/deployment-tail/api"
	"github.com/claudioed/deployment-tail/internal/adapters/input/http/middleware"
	"github.com/claudioed/deployment-tail/internal/application/ports/input"
	"github.com/claudioed/deployment-tail/internal/domain/user"
)

// Mock UserService
type MockUserService struct {
	authenticateWithGoogleFunc func(ctx context.Context, code string) (*input.AuthenticationResult, error)
	refreshUserTokenFunc       func(ctx context.Context, userID user.UserID) (string, error)
	revokeUserTokenFunc        func(ctx context.Context, tokenHash string, userID user.UserID) error
}

func (m *MockUserService) AuthenticateWithGoogle(ctx context.Context, code string) (*input.AuthenticationResult, error) {
	if m.authenticateWithGoogleFunc != nil {
		return m.authenticateWithGoogleFunc(ctx, code)
	}
	return nil, fmt.Errorf("not implemented")
}

func (m *MockUserService) RegisterOrUpdateUser(ctx context.Context, googleID, email, name string) (*user.User, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *MockUserService) GetUserProfile(ctx context.Context, userID user.UserID) (*user.User, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *MockUserService) ListUsers(ctx context.Context, filters input.UserListFilters) ([]*user.User, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *MockUserService) AssignRole(ctx context.Context, adminUserID, targetUserID user.UserID, newRole user.Role) error {
	return fmt.Errorf("not implemented")
}

func (m *MockUserService) RefreshUserToken(ctx context.Context, userID user.UserID) (string, error) {
	if m.refreshUserTokenFunc != nil {
		return m.refreshUserTokenFunc(ctx, userID)
	}
	return "", fmt.Errorf("not implemented")
}

func (m *MockUserService) RevokeUserToken(ctx context.Context, tokenHash string, userID user.UserID) error {
	if m.revokeUserTokenFunc != nil {
		return m.revokeUserTokenFunc(ctx, tokenHash, userID)
	}
	return fmt.Errorf("not implemented")
}

// Mock GoogleClient
type MockGoogleClient struct {
	getAuthURLFunc func(state string) string
}

func (m *MockGoogleClient) GetAuthURL(state string) string {
	if m.getAuthURLFunc != nil {
		return m.getAuthURLFunc(state)
	}
	return "https://accounts.google.com/o/oauth2/auth?mock=true"
}

// Helper functions
func createTestUser(role string) *user.User {
	googleID, _ := user.NewGoogleID("test-google-id")
	email, _ := user.NewEmail("test@example.com")
	name, _ := user.NewUserName("Test User")
	r, _ := user.NewRole(role)
	u, _ := user.NewUser(googleID, email, name, r)
	return u
}

func TestGoogleLogin(t *testing.T) {
	mockGoogleClient := &MockGoogleClient{
		getAuthURLFunc: func(state string) string {
			return "https://accounts.google.com/o/oauth2/auth?state=" + state
		},
	}

	handler := NewAuthHandler(nil, mockGoogleClient)

	req := httptest.NewRequest(http.MethodGet, "/auth/google/login", nil)
	rr := httptest.NewRecorder()

	handler.GoogleLogin(rr, req)

	// Should redirect to Google
	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("expected status %d, got %d", http.StatusTemporaryRedirect, rr.Code)
	}

	location := rr.Header().Get("Location")
	if location == "" {
		t.Error("expected Location header to be set")
	}

	if len(location) < 10 || location[:10] != "https://ac" {
		t.Errorf("expected Google OAuth URL, got %s", location)
	}
}

func TestGoogleCallback_Success(t *testing.T) {
	testUser := createTestUser(user.RoleDeployer)

	mockUserService := &MockUserService{
		authenticateWithGoogleFunc: func(ctx context.Context, code string) (*input.AuthenticationResult, error) {
			if code != "valid-code" {
				return nil, fmt.Errorf("invalid code")
			}
			return &input.AuthenticationResult{
				User:  testUser,
				Token: "jwt-token-123",
			}, nil
		},
	}

	handler := NewAuthHandler(mockUserService, nil)

	req := httptest.NewRequest(http.MethodGet, "/auth/google/callback?code=valid-code&state=test-state", nil)
	req.AddCookie(&http.Cookie{Name: oauthStateCookieName, Value: "test-state"})
	rr := httptest.NewRecorder()

	params := api.GoogleCallbackParams{
		Code:  "valid-code",
		State: "test-state",
	}
	handler.GoogleCallback(rr, req, params)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	// GoogleCallback returns an HTML bootstrap page that stores the token in
	// localStorage and redirects the browser. Verify the token and user
	// details are embedded in that HTML rather than trying to decode JSON.
	body := rr.Body.String()
	if !strings.Contains(body, "jwt-token-123") {
		t.Errorf("expected token 'jwt-token-123' in response body")
	}
	if !strings.Contains(body, testUser.Email().String()) {
		t.Errorf("expected email %s in response body", testUser.Email().String())
	}

	contentType := rr.Header().Get("Content-Type")
	if !strings.HasPrefix(contentType, "text/html") {
		t.Errorf("expected Content-Type text/html, got %s", contentType)
	}
}

func TestGoogleCallback_MissingCode(t *testing.T) {
	handler := NewAuthHandler(nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/auth/google/callback?state=test-state", nil)
	rr := httptest.NewRecorder()

	params := api.GoogleCallbackParams{
		Code:  "",
		State: "test-state",
	}
	handler.GoogleCallback(rr, req, params)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response["error"] == nil {
		t.Error("expected error in response")
	}
}

func TestGoogleCallback_MissingState(t *testing.T) {
	handler := NewAuthHandler(nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/auth/google/callback?code=valid-code", nil)
	rr := httptest.NewRecorder()

	params := api.GoogleCallbackParams{
		Code:  "valid-code",
		State: "",
	}
	handler.GoogleCallback(rr, req, params)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
}

func TestGoogleCallback_ErrorFromGoogle(t *testing.T) {
	handler := NewAuthHandler(nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/auth/google/callback?error=access_denied&state=test-state", nil)
	rr := httptest.NewRecorder()

	params := api.GoogleCallbackParams{
		Code:  "",
		State: "test-state",
	}
	handler.GoogleCallback(rr, req, params)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	errorMsg, ok := response["error"].(string)
	if !ok || errorMsg == "" {
		t.Error("expected error message in response")
	}
}

func TestGoogleCallback_AuthenticationFailed(t *testing.T) {
	mockUserService := &MockUserService{
		authenticateWithGoogleFunc: func(ctx context.Context, code string) (*input.AuthenticationResult, error) {
			return nil, fmt.Errorf("authentication failed")
		},
	}

	handler := NewAuthHandler(mockUserService, nil)

	req := httptest.NewRequest(http.MethodGet, "/auth/google/callback?code=invalid-code&state=test-state", nil)
	req.AddCookie(&http.Cookie{Name: oauthStateCookieName, Value: "test-state"})
	rr := httptest.NewRecorder()

	params := api.GoogleCallbackParams{
		Code:  "invalid-code",
		State: "test-state",
	}
	handler.GoogleCallback(rr, req, params)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, rr.Code)
	}
}

func TestRefreshToken_Success(t *testing.T) {
	testUser := createTestUser(user.RoleDeployer)

	mockUserService := &MockUserService{
		refreshUserTokenFunc: func(ctx context.Context, userID user.UserID) (string, error) {
			if userID.Equals(testUser.ID()) {
				return "new-jwt-token-456", nil
			}
			return "", fmt.Errorf("user not found")
		},
	}

	handler := NewAuthHandler(mockUserService, nil)

	// Create context with authenticated user using the middleware helper
	ctx := middleware.UserToContext(context.Background(), testUser)
	req := httptest.NewRequest(http.MethodPost, "/auth/refresh", nil).WithContext(ctx)
	rr := httptest.NewRecorder()

	handler.RefreshToken(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response["token"] != "new-jwt-token-456" {
		t.Errorf("expected token 'new-jwt-token-456', got %v", response["token"])
	}
}

func TestRefreshToken_NoUserInContext(t *testing.T) {
	handler := NewAuthHandler(nil, nil)

	req := httptest.NewRequest(http.MethodPost, "/auth/refresh", nil)
	// No user in context

	rr := httptest.NewRecorder()

	handler.RefreshToken(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, rr.Code)
	}
}

func TestRefreshToken_ServiceError(t *testing.T) {
	testUser := createTestUser(user.RoleDeployer)

	mockUserService := &MockUserService{
		refreshUserTokenFunc: func(ctx context.Context, userID user.UserID) (string, error) {
			return "", fmt.Errorf("token generation failed")
		},
	}

	handler := NewAuthHandler(mockUserService, nil)

	ctx := middleware.UserToContext(context.Background(), testUser)
	req := httptest.NewRequest(http.MethodPost, "/auth/refresh", nil).WithContext(ctx)
	rr := httptest.NewRecorder()

	handler.RefreshToken(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, rr.Code)
	}
}

func TestLogout_Success(t *testing.T) {
	testUser := createTestUser(user.RoleDeployer)

	mockUserService := &MockUserService{
		revokeUserTokenFunc: func(ctx context.Context, tokenHash string, userID user.UserID) error {
			if tokenHash == "valid-token" && userID.Equals(testUser.ID()) {
				return nil
			}
			return fmt.Errorf("invalid token or user")
		},
	}

	handler := NewAuthHandler(mockUserService, nil)

	ctx := middleware.UserToContext(context.Background(), testUser)
	req := httptest.NewRequest(http.MethodPost, "/auth/logout", nil).WithContext(ctx)
	req.Header.Set("Authorization", "Bearer valid-token")

	rr := httptest.NewRecorder()

	handler.Logout(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response["message"] == nil {
		t.Error("expected success message in response")
	}
}

func TestLogout_NoUserInContext(t *testing.T) {
	handler := NewAuthHandler(nil, nil)

	req := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
	req.Header.Set("Authorization", "Bearer valid-token")

	rr := httptest.NewRecorder()

	handler.Logout(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, rr.Code)
	}
}

func TestLogout_MissingAuthorizationHeader(t *testing.T) {
	testUser := createTestUser(user.RoleDeployer)

	handler := NewAuthHandler(nil, nil)

	ctx := middleware.UserToContext(context.Background(), testUser)
	req := httptest.NewRequest(http.MethodPost, "/auth/logout", nil).WithContext(ctx)
	// No Authorization header

	rr := httptest.NewRecorder()

	handler.Logout(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
}

func TestLogout_ServiceError(t *testing.T) {
	testUser := createTestUser(user.RoleDeployer)

	mockUserService := &MockUserService{
		revokeUserTokenFunc: func(ctx context.Context, tokenHash string, userID user.UserID) error {
			return fmt.Errorf("revocation failed")
		},
	}

	handler := NewAuthHandler(mockUserService, nil)

	ctx := middleware.UserToContext(context.Background(), testUser)
	req := httptest.NewRequest(http.MethodPost, "/auth/logout", nil).WithContext(ctx)
	req.Header.Set("Authorization", "Bearer valid-token")

	rr := httptest.NewRecorder()

	handler.Logout(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, rr.Code)
	}
}

func TestWriteJSONResponse(t *testing.T) {
	handler := NewAuthHandler(nil, nil)

	data := map[string]interface{}{
		"message": "test",
		"value":   123,
	}

	rr := httptest.NewRecorder()

	handler.writeJSONResponse(rr, data, http.StatusOK)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	contentType := rr.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("expected Content-Type 'application/json', got %s", contentType)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response["message"] != "test" {
		t.Error("response data does not match")
	}
}

func TestWriteErrorResponse(t *testing.T) {
	handler := NewAuthHandler(nil, nil)

	rr := httptest.NewRecorder()

	handler.writeErrorResponse(rr, "test error", http.StatusBadRequest)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response["error"] != "test error" {
		t.Errorf("expected error 'test error', got %v", response["error"])
	}
}
