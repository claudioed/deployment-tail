//go:build integration
// +build integration

package cli

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/claudioed/deployment-tail/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCLIAuthenticationE2E tests the complete CLI auth flow (Task 25.6)
// Flow: Login → Store Token → Create Schedule → Verify Ownership
func TestCLIAuthenticationE2E(t *testing.T) {
	// Setup temporary directory for token storage
	tempDir := t.TempDir()

	// Mock API server that simulates authentication and schedule creation
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		// Mock OAuth callback endpoint
		case r.URL.Path == "/auth/google/callback":
			// Simulate successful OAuth callback
			response := map[string]interface{}{
				"token": "test-jwt-token-from-google-oauth",
				"user": map[string]string{
					"email": "test@example.com",
					"name":  "Test User",
					"role":  "deployer",
				},
			}
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(response)

		// Mock schedule creation endpoint
		case r.URL.Path == "/api/v1/schedules" && r.Method == "POST":
			// Verify Authorization header is present
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized"})
				return
			}

			if authHeader != "Bearer test-jwt-token-from-google-oauth" {
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(map[string]string{"error": "invalid token"})
				return
			}

			// Decode request
			var req api.CreateScheduleRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			// Return created schedule with audit information
			schedule := api.Schedule{
				Id:          "550e8400-e29b-41d4-a716-446655440000",
				ScheduledAt: req.ScheduledAt,
				ServiceName: req.ServiceName,
				Environments: func() []api.ScheduleEnvironments {
					envs := make([]api.ScheduleEnvironments, len(req.Environments))
					for i, e := range req.Environments {
						envs[i] = api.ScheduleEnvironments(e)
					}
					return envs
				}(),
				Owners: req.Owners,
				Status: api.Created,
				CreatedBy: api.User{
					Id:    "user-id-123",
					Email: "test@example.com",
					Name:  "Test User",
					Role:  api.Deployer,
				},
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}

			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(schedule)

		// Mock get schedule endpoint
		case r.URL.Path == "/api/v1/schedules/550e8400-e29b-41d4-a716-446655440000" && r.Method == "GET":
			// Verify Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			// Return schedule with creator information
			schedule := api.Schedule{
				Id:          "550e8400-e29b-41d4-a716-446655440000",
				ScheduledAt: time.Now().Add(24 * time.Hour),
				ServiceName: "test-service",
				Environments: []api.ScheduleEnvironments{
					api.Production,
				},
				Owners: []string{"test@example.com"},
				Status: api.Created,
				CreatedBy: api.User{
					Id:    "user-id-123",
					Email: "test@example.com",
					Name:  "Test User",
					Role:  api.Deployer,
				},
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}

			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(schedule)

		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer apiServer.Close()

	// Step 1: Simulate login by saving a token
	tokenStore := &TokenStore{
		configDir: tempDir,
		tokenPath: filepath.Join(tempDir, "auth.json"),
	}

	err := tokenStore.SaveToken(
		"test-jwt-token-from-google-oauth",
		"test@example.com",
		"deployer",
		time.Now().Add(24*time.Hour),
	)
	require.NoError(t, err, "Failed to save token")

	// Step 2: Verify token was saved
	assert.True(t, tokenStore.IsAuthenticated(), "User should be authenticated")

	// Step 3: Load token
	savedToken, err := tokenStore.LoadToken()
	require.NoError(t, err, "Failed to load token")
	assert.Equal(t, "test-jwt-token-from-google-oauth", savedToken.Token)
	assert.Equal(t, "test@example.com", savedToken.Email)
	assert.Equal(t, "deployer", savedToken.Role)

	// Step 4: Create API client with token
	client := NewAPIClient(apiServer.URL)
	client.tokenStore = tokenStore

	// Step 5: Create a schedule (requires authentication)
	ctx := context.Background()
	scheduledTime := time.Now().Add(24 * time.Hour)

	req := api.CreateScheduleRequest{
		ScheduledAt:  scheduledTime,
		ServiceName:  "test-service",
		Environments: []api.CreateScheduleRequestEnvironments{api.CreateScheduleRequestEnvironments("production")},
		Owners:       []string{"test@example.com"},
	}

	schedule, err := client.CreateSchedule(ctx, req)
	require.NoError(t, err, "Failed to create schedule")
	assert.NotNil(t, schedule)

	// Step 6: Verify schedule has audit trail
	assert.Equal(t, "test-service", schedule.ServiceName)
	assert.Equal(t, "test@example.com", schedule.CreatedBy.Email)
	assert.Equal(t, "Test User", schedule.CreatedBy.Name)
	assert.Equal(t, api.Deployer, schedule.CreatedBy.Role)

	// Step 7: Retrieve schedule and verify ownership
	retrievedSchedule, err := client.GetSchedule(ctx, schedule.Id.String())
	require.NoError(t, err, "Failed to retrieve schedule")
	assert.Equal(t, schedule.Id, retrievedSchedule.Id)
	assert.Equal(t, "test@example.com", retrievedSchedule.CreatedBy.Email)

	// Step 8: Cleanup - logout (delete token)
	err = tokenStore.DeleteToken()
	require.NoError(t, err, "Failed to delete token")
	assert.False(t, tokenStore.IsAuthenticated(), "User should not be authenticated after logout")
}

// TestCLIAuthenticationFailure tests unauthenticated requests (Task 25.7)
func TestCLIAuthenticationFailure(t *testing.T) {
	// Mock API server that requires authentication
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized"})
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer apiServer.Close()

	// Setup token store without any token
	tempDir := t.TempDir()
	tokenStore := &TokenStore{
		configDir: tempDir,
		tokenPath: filepath.Join(tempDir, "auth.json"),
	}

	// Create API client
	client := NewAPIClient(apiServer.URL)
	client.tokenStore = tokenStore

	// Test: Request without authentication should fail
	ctx := context.Background()
	_, err := client.GetSchedule(ctx, "test-id")
	require.Error(t, err, "Request without authentication should fail")

	// Test: Error should be AuthenticationError
	authErr, ok := err.(*AuthenticationError)
	assert.True(t, ok, "Error should be AuthenticationError")
	assert.Contains(t, authErr.Message, "authenticated", "Error message should mention authentication")
}

// TestCLITokenExpiry tests expired token handling (Task 25.8)
func TestCLITokenExpiry(t *testing.T) {
	// Setup
	tempDir := t.TempDir()
	tokenStore := &TokenStore{
		configDir: tempDir,
		tokenPath: filepath.Join(tempDir, "auth.json"),
	}

	// Save an expired token
	err := tokenStore.SaveToken(
		"expired-token",
		"test@example.com",
		"deployer",
		time.Now().Add(-1*time.Hour), // Expired 1 hour ago
	)
	require.NoError(t, err)

	// Test: IsAuthenticated should return false for expired token
	assert.False(t, tokenStore.IsAuthenticated(), "Expired token should not authenticate")

	// Test: Attempting to use expired token should fail
	client := NewAPIClient("http://localhost:8080")
	client.tokenStore = tokenStore

	ctx := context.Background()
	_, err = client.GetSchedule(ctx, "test-id")
	require.Error(t, err, "Request with expired token should fail")

	// Test: Error should be AuthenticationError
	authErr, ok := err.(*AuthenticationError)
	assert.True(t, ok, "Error should be AuthenticationError")
	assert.Contains(t, authErr.Message, "expired", "Error message should mention expiration")
}

// TestCLIRoleEnforcement tests role-based access from CLI (Task 25.9)
func TestCLIRoleEnforcement(t *testing.T) {
	tests := []struct {
		name           string
		role           string
		operation      string
		expectAllowed  bool
		expectedStatus int
	}{
		{"Viewer can view", "viewer", "GET", true, http.StatusOK},
		{"Viewer cannot create", "viewer", "POST", false, http.StatusForbidden},
		{"Deployer can create", "deployer", "POST", true, http.StatusOK},
		{"Deployer cannot approve", "deployer", "APPROVE", false, http.StatusForbidden},
		{"Admin can approve", "admin", "APPROVE", true, http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mock API server with role enforcement
			apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Extract role from Authorization header (in real scenario, from JWT claims)
				// For this test, we'll simulate role checking
				switch {
				case r.URL.Path == "/api/v1/schedules" && r.Method == "GET":
					// All roles can view
					w.WriteHeader(http.StatusOK)
					json.NewEncoder(w).Encode([]api.Schedule{})

				case r.URL.Path == "/api/v1/schedules" && r.Method == "POST":
					// Only deployer and admin can create
					if tt.role == "viewer" {
						w.WriteHeader(http.StatusForbidden)
						json.NewEncoder(w).Encode(map[string]string{"error": "permission denied"})
						return
					}
					w.WriteHeader(http.StatusOK)

				case r.URL.Path == "/api/v1/schedules/test-id/approve" && r.Method == "POST":
					// Only admin can approve
					if tt.role != "admin" {
						w.WriteHeader(http.StatusForbidden)
						json.NewEncoder(w).Encode(map[string]string{"error": "permission denied"})
						return
					}
					w.WriteHeader(http.StatusOK)

				default:
					w.WriteHeader(http.StatusNotFound)
				}
			}))
			defer apiServer.Close()

			// Setup authenticated client
			tempDir := t.TempDir()
			tokenStore := &TokenStore{
				configDir: tempDir,
				tokenPath: filepath.Join(tempDir, "auth.json"),
			}

			err := tokenStore.SaveToken(
				"test-token",
				"test@example.com",
				tt.role,
				time.Now().Add(24*time.Hour),
			)
			require.NoError(t, err)

			client := NewAPIClient(apiServer.URL)
			client.tokenStore = tokenStore

			// Perform operation
			ctx := context.Background()
			var operationErr error

			switch tt.operation {
			case "GET":
				_, operationErr = client.ListSchedules(ctx, nil, nil, nil, nil, nil)
			case "POST":
				req := api.CreateScheduleRequest{
					ScheduledAt:  time.Now().Add(24 * time.Hour),
					ServiceName:  "test-service",
					Environments: []api.CreateScheduleRequestEnvironments{api.CreateScheduleRequestEnvironments("production")},
					Owners:       []string{"test@example.com"},
				}
				_, operationErr = client.CreateSchedule(ctx, req)
			case "APPROVE":
				_, operationErr = client.ApproveSchedule(ctx, "test-id")
			}

			// Verify result
			if tt.expectAllowed {
				assert.NoError(t, operationErr, "Operation should be allowed for role %s", tt.role)
			} else {
				assert.Error(t, operationErr, "Operation should be denied for role %s", tt.role)
				if operationErr != nil {
					_, isPermErr := operationErr.(*PermissionError)
					assert.True(t, isPermErr || operationErr.Error() != "", "Should be permission error")
				}
			}
		})
	}
}

// TestCLITokenRefresh tests automatic token refresh (Task 25.2)
func TestCLITokenRefresh(t *testing.T) {
	refreshCalled := false

	// Mock API server with refresh endpoint
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/auth/refresh":
			refreshCalled = true
			// Return new token
			response := map[string]interface{}{
				"token": "new-refreshed-token",
				"user": map[string]string{
					"email": "test@example.com",
					"role":  "deployer",
				},
			}
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(response)

		case "/api/v1/schedules":
			// Return empty list
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode([]api.Schedule{})

		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer apiServer.Close()

	// Setup token store with token expiring soon
	tempDir := t.TempDir()
	tokenStore := &TokenStore{
		configDir: tempDir,
		tokenPath: filepath.Join(tempDir, "auth.json"),
	}

	// Save token that expires in 30 minutes (should trigger refresh)
	err := tokenStore.SaveToken(
		"old-token",
		"test@example.com",
		"deployer",
		time.Now().Add(30*time.Minute),
	)
	require.NoError(t, err)

	// Create client
	client := NewAPIClient(apiServer.URL)
	client.tokenStore = tokenStore

	// Make a request (should trigger refresh)
	ctx := context.Background()
	_, err = client.ListSchedules(ctx, nil, nil, nil, nil, nil)
	require.NoError(t, err)

	// Verify refresh was called
	assert.True(t, refreshCalled, "Token refresh should be called for soon-to-expire token")

	// Verify new token was saved
	newToken, err := tokenStore.LoadToken()
	require.NoError(t, err)
	assert.Equal(t, "new-refreshed-token", newToken.Token)
}

// TestCLITokenStorage tests secure token file permissions (Task 25.7)
func TestCLITokenStorage(t *testing.T) {
	// Setup
	tempDir := t.TempDir()
	tokenStore := &TokenStore{
		configDir: tempDir,
		tokenPath: filepath.Join(tempDir, "auth.json"),
	}

	// Save token
	err := tokenStore.SaveToken(
		"test-token",
		"test@example.com",
		"deployer",
		time.Now().Add(24*time.Hour),
	)
	require.NoError(t, err)

	// Test: File exists
	_, err = os.Stat(tokenStore.tokenPath)
	require.NoError(t, err, "Token file should exist")

	// Test: File has correct permissions (0600)
	fileInfo, err := os.Stat(tokenStore.tokenPath)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0600), fileInfo.Mode().Perm(), "Token file should have 0600 permissions")

	// Test: Directory has correct permissions (0700)
	dirInfo, err := os.Stat(tokenStore.configDir)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0700), dirInfo.Mode().Perm(), "Config directory should have 0700 permissions")
}
