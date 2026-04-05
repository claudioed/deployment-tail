package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/claudioed/deployment-tail/api"
)

// APIClient wraps HTTP client for API calls
type APIClient struct {
	baseURL    string
	httpClient *http.Client
	tokenStore *TokenStore
}

// NewAPIClient creates a new API client
func NewAPIClient(baseURL string) *APIClient {
	return &APIClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		tokenStore: NewTokenStore(),
	}
}

// CreateSchedule creates a new schedule
func (c *APIClient) CreateSchedule(ctx context.Context, req api.CreateScheduleRequest) (*api.Schedule, error) {
	var result api.Schedule
	err := c.doRequest(ctx, "POST", "/api/v1/schedules", req, &result)
	return &result, err
}

// GetSchedule retrieves a schedule by ID
func (c *APIClient) GetSchedule(ctx context.Context, id string) (*api.Schedule, error) {
	var result api.Schedule
	err := c.doRequest(ctx, "GET", fmt.Sprintf("/api/v1/schedules/%s", id), nil, &result)
	return &result, err
}

// ListSchedules retrieves all schedules
func (c *APIClient) ListSchedules(ctx context.Context, from, to *time.Time, environments, owners []string, status *string) ([]api.Schedule, error) {
	url := "/api/v1/schedules?"
	if from != nil {
		url += fmt.Sprintf("from=%s&", from.Format(time.RFC3339))
	}
	if to != nil {
		url += fmt.Sprintf("to=%s&", to.Format(time.RFC3339))
	}
	for _, env := range environments {
		url += fmt.Sprintf("environment=%s&", env)
	}
	for _, owner := range owners {
		url += fmt.Sprintf("owner=%s&", owner)
	}
	if status != nil {
		url += fmt.Sprintf("status=%s&", *status)
	}

	var result []api.Schedule
	err := c.doRequest(ctx, "GET", url, nil, &result)
	return result, err
}

// UpdateSchedule updates a schedule
func (c *APIClient) UpdateSchedule(ctx context.Context, id string, req api.UpdateScheduleRequest) (*api.Schedule, error) {
	var result api.Schedule
	err := c.doRequest(ctx, "PUT", fmt.Sprintf("/api/v1/schedules/%s", id), req, &result)
	return &result, err
}

// DeleteSchedule deletes a schedule
func (c *APIClient) DeleteSchedule(ctx context.Context, id string) error {
	return c.doRequest(ctx, "DELETE", fmt.Sprintf("/api/v1/schedules/%s", id), nil, nil)
}

// ApproveSchedule approves a schedule
func (c *APIClient) ApproveSchedule(ctx context.Context, id string) (*api.Schedule, error) {
	var result api.Schedule
	err := c.doRequest(ctx, "POST", fmt.Sprintf("/api/v1/schedules/%s/approve", id), nil, &result)
	return &result, err
}

// DenySchedule denies a schedule
func (c *APIClient) DenySchedule(ctx context.Context, id string) (*api.Schedule, error) {
	var result api.Schedule
	err := c.doRequest(ctx, "POST", fmt.Sprintf("/api/v1/schedules/%s/deny", id), nil, &result)
	return &result, err
}

// ListGroups retrieves all groups for an owner
func (c *APIClient) ListGroups(ctx context.Context, owner string) ([]interface{}, error) {
	url := fmt.Sprintf("/api/v1/groups?owner=%s", owner)
	var result []interface{}
	err := c.doRequest(ctx, "GET", url, nil, &result)
	return result, err
}

// FavoriteGroup marks a group as favorite
func (c *APIClient) FavoriteGroup(ctx context.Context, groupID string) error {
	return c.doRequest(ctx, "POST", fmt.Sprintf("/api/v1/groups/%s/favorite", groupID), nil, nil)
}

// UnfavoriteGroup removes favorite status from a group
func (c *APIClient) UnfavoriteGroup(ctx context.Context, groupID string) error {
	return c.doRequest(ctx, "DELETE", fmt.Sprintf("/api/v1/groups/%s/favorite", groupID), nil, nil)
}

// doRequest performs an authenticated HTTP request
func (c *APIClient) doRequest(ctx context.Context, method, path string, body, result interface{}) error {
	// Validate and refresh token before request
	if err := c.ensureValidToken(); err != nil {
		return err
	}

	url := c.baseURL + path

	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	// Add authentication header
	token, err := c.tokenStore.LoadToken()
	if err == nil {
		req.Header.Set("Authorization", "Bearer "+token.Token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("API request failed: %w (is the API server running at %s?)", err, c.baseURL)
	}
	defer resp.Body.Close()

	// Handle authentication errors
	if resp.StatusCode == http.StatusUnauthorized {
		return &AuthenticationError{
			Message: "Authentication required or token expired. Please run 'deployment-tail auth login'",
		}
	}

	if resp.StatusCode == http.StatusForbidden {
		return &PermissionError{
			Message: "Permission denied. You don't have access to perform this operation",
		}
	}

	if resp.StatusCode >= 400 {
		var errResp api.Error
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err == nil {
			return fmt.Errorf("API error (%d): %s", resp.StatusCode, errResp.Message)
		}
		return fmt.Errorf("API error: status %d", resp.StatusCode)
	}

	if result != nil && resp.StatusCode != http.StatusNoContent {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}

	return nil
}

// ensureValidToken validates the token and refreshes if near expiry
func (c *APIClient) ensureValidToken() error {
	token, err := c.tokenStore.LoadToken()
	if err != nil {
		if os.IsNotExist(err) {
			return &AuthenticationError{
				Message: "Not authenticated. Please run 'deployment-tail auth login'",
			}
		}
		return fmt.Errorf("failed to load token: %w", err)
	}

	// Check if token is expired
	if time.Now().After(token.ExpiresAt) {
		return &AuthenticationError{
			Message: "Token expired. Please run 'deployment-tail auth login'",
		}
	}

	// Refresh token if it expires within 1 hour
	timeUntilExpiry := time.Until(token.ExpiresAt)
	if timeUntilExpiry < time.Hour {
		if err := c.refreshToken(); err != nil {
			// If refresh fails, just log a warning and continue
			// The token might still be valid for the current request
			fmt.Fprintf(os.Stderr, "Warning: Failed to refresh token: %v\n", err)
		}
	}

	return nil
}

// refreshToken refreshes the authentication token
func (c *APIClient) refreshToken() error {
	token, err := c.tokenStore.LoadToken()
	if err != nil {
		return fmt.Errorf("failed to load current token: %w", err)
	}

	// Call refresh endpoint
	req, err := http.NewRequest("POST", c.baseURL+"/auth/refresh", nil)
	if err != nil {
		return fmt.Errorf("failed to create refresh request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token.Token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to refresh token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("token refresh failed with status %d", resp.StatusCode)
	}

	// Parse new token
	var refreshResp struct {
		Token string `json:"token"`
		User  struct {
			Email string `json:"email"`
			Role  string `json:"role"`
		} `json:"user"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&refreshResp); err != nil {
		return fmt.Errorf("failed to parse refresh response: %w", err)
	}

	// Save new token
	newExpiresAt := time.Now().Add(24 * time.Hour) // Assume 24h expiry
	if err := c.tokenStore.SaveToken(refreshResp.Token, refreshResp.User.Email, refreshResp.User.Role, newExpiresAt); err != nil {
		return fmt.Errorf("failed to save refreshed token: %w", err)
	}

	return nil
}

// AuthenticationError represents an authentication failure
type AuthenticationError struct {
	Message string
}

func (e *AuthenticationError) Error() string {
	return e.Message
}

// PermissionError represents a permission denial
type PermissionError struct {
	Message string
}

func (e *PermissionError) Error() string {
	return e.Message
}
