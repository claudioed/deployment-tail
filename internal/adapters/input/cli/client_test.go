package cli

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestAPIClient_AuthenticationHeader(t *testing.T) {
	// Create a test server that checks for Authorization header
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth == "" {
			t.Error("Expected Authorization header to be set")
		}

		if auth != "Bearer test-token-123" {
			t.Errorf("Expected 'Bearer test-token-123', got '%s'", auth)
		}

		w.WriteHeader(http.StatusOK)
		// Return a valid UUID format
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":          "550e8400-e29b-41d4-a716-446655440000",
			"serviceName": "test-service",
		})
	}))
	defer server.Close()

	// Create client with test server
	client := NewAPIClient(server.URL)

	// Setup test token - expires in 2 hours (won't trigger refresh)
	tempDir := t.TempDir()
	client.tokenStore = &TokenStore{
		configDir: tempDir,
		tokenPath: tempDir + "/auth.json",
	}

	// Save a valid token
	err := client.tokenStore.SaveToken("test-token-123", "test@example.com", "deployer", time.Now().Add(2*time.Hour))
	if err != nil {
		t.Fatalf("Failed to save token: %v", err)
	}

	// Make a request
	_, err = client.GetSchedule(context.Background(), "550e8400-e29b-41d4-a716-446655440000")
	if err != nil {
		t.Errorf("GetSchedule failed: %v", err)
	}
}

func TestAPIClient_401Unauthorized(t *testing.T) {
	// Create a test server that returns 401
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "unauthorized",
		})
	}))
	defer server.Close()

	// Create client
	client := NewAPIClient(server.URL)

	// Setup test token
	tempDir := t.TempDir()
	client.tokenStore = &TokenStore{
		configDir: tempDir,
		tokenPath: tempDir + "/auth.json",
	}

	// Save a valid token
	err := client.tokenStore.SaveToken("test-token", "test@example.com", "deployer", time.Now().Add(time.Hour))
	if err != nil {
		t.Fatalf("Failed to save token: %v", err)
	}

	// Make a request
	_, err = client.GetSchedule(context.Background(), "test-id")
	if err == nil {
		t.Error("Expected error for 401 response")
	}

	// Check if it's an AuthenticationError
	if _, ok := err.(*AuthenticationError); !ok {
		t.Errorf("Expected AuthenticationError, got %T: %v", err, err)
	}
}

func TestAPIClient_403Forbidden(t *testing.T) {
	// Create a test server that returns 403
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "forbidden",
		})
	}))
	defer server.Close()

	// Create client
	client := NewAPIClient(server.URL)

	// Setup test token
	tempDir := t.TempDir()
	client.tokenStore = &TokenStore{
		configDir: tempDir,
		tokenPath: tempDir + "/auth.json",
	}

	// Save a valid token
	err := client.tokenStore.SaveToken("test-token", "test@example.com", "deployer", time.Now().Add(time.Hour))
	if err != nil {
		t.Fatalf("Failed to save token: %v", err)
	}

	// Make a request
	_, err = client.GetSchedule(context.Background(), "test-id")
	if err == nil {
		t.Error("Expected error for 403 response")
	}

	// Check if it's a PermissionError
	if _, ok := err.(*PermissionError); !ok {
		t.Errorf("Expected PermissionError, got %T: %v", err, err)
	}
}

func TestAPIClient_TokenExpired(t *testing.T) {
	// Create client
	client := NewAPIClient("http://localhost:8080")

	// Setup expired token
	tempDir := t.TempDir()
	client.tokenStore = &TokenStore{
		configDir: tempDir,
		tokenPath: tempDir + "/auth.json",
	}

	// Save an expired token
	err := client.tokenStore.SaveToken("expired-token", "test@example.com", "deployer", time.Now().Add(-time.Hour))
	if err != nil {
		t.Fatalf("Failed to save token: %v", err)
	}

	// Try to make a request
	_, err = client.GetSchedule(context.Background(), "test-id")
	if err == nil {
		t.Error("Expected error for expired token")
	}

	// Check if it's an AuthenticationError
	if _, ok := err.(*AuthenticationError); !ok {
		t.Errorf("Expected AuthenticationError for expired token, got %T: %v", err, err)
	}
}

func TestAPIClient_NoToken(t *testing.T) {
	// Create client
	client := NewAPIClient("http://localhost:8080")

	// Setup token store with no token
	tempDir := t.TempDir()
	client.tokenStore = &TokenStore{
		configDir: tempDir,
		tokenPath: tempDir + "/auth.json",
	}

	// Try to make a request without a token
	_, err := client.GetSchedule(context.Background(), "test-id")
	if err == nil {
		t.Error("Expected error when no token exists")
	}

	// Check if it's an AuthenticationError
	if _, ok := err.(*AuthenticationError); !ok {
		t.Errorf("Expected AuthenticationError when no token, got %T: %v", err, err)
	}
}

func TestAPIClient_TokenRefresh(t *testing.T) {
	refreshCalled := false

	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/auth/refresh" {
			refreshCalled = true
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"token": "new-token-456",
				"user": map[string]string{
					"email": "test@example.com",
					"role":  "deployer",
				},
			})
			return
		}

		// Regular request - return valid UUID
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":          "550e8400-e29b-41d4-a716-446655440000",
			"serviceName": "test-service",
		})
	}))
	defer server.Close()

	// Create client
	client := NewAPIClient(server.URL)

	// Setup token store
	tempDir := t.TempDir()
	client.tokenStore = &TokenStore{
		configDir: tempDir,
		tokenPath: tempDir + "/auth.json",
	}

	// Save a token that expires in 30 minutes (should trigger refresh)
	err := client.tokenStore.SaveToken("old-token", "test@example.com", "deployer", time.Now().Add(30*time.Minute))
	if err != nil {
		t.Fatalf("Failed to save token: %v", err)
	}

	// Make a request
	_, err = client.GetSchedule(context.Background(), "550e8400-e29b-41d4-a716-446655440000")
	if err != nil {
		t.Errorf("GetSchedule failed: %v", err)
	}

	// Verify refresh was called
	if !refreshCalled {
		t.Error("Expected token refresh to be called")
	}

	// Verify new token was saved
	newToken, err := client.tokenStore.LoadToken()
	if err != nil {
		t.Fatalf("Failed to load new token: %v", err)
	}

	if newToken.Token != "new-token-456" {
		t.Errorf("Expected new token 'new-token-456', got '%s'", newToken.Token)
	}
}

func TestAuthenticationError(t *testing.T) {
	err := &AuthenticationError{Message: "test error"}
	if err.Error() != "test error" {
		t.Errorf("Expected 'test error', got '%s'", err.Error())
	}
}

func TestPermissionError(t *testing.T) {
	err := &PermissionError{Message: "permission denied"}
	if err.Error() != "permission denied" {
		t.Errorf("Expected 'permission denied', got '%s'", err.Error())
	}
}
