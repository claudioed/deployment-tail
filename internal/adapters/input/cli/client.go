package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/claudioed/deployment-tail/api"
)

// APIClient wraps HTTP client for API calls
type APIClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewAPIClient creates a new API client
func NewAPIClient(baseURL string) *APIClient {
	return &APIClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
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
func (c *APIClient) ListSchedules(ctx context.Context, from, to *time.Time, env *string) ([]api.Schedule, error) {
	url := "/api/v1/schedules?"
	if from != nil {
		url += fmt.Sprintf("from=%s&", from.Format(time.RFC3339))
	}
	if to != nil {
		url += fmt.Sprintf("to=%s&", to.Format(time.RFC3339))
	}
	if env != nil {
		url += fmt.Sprintf("environment=%s&", *env)
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

// doRequest performs an HTTP request
func (c *APIClient) doRequest(ctx context.Context, method, path string, body, result interface{}) error {
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

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("API request failed: %w (is the API server running at %s?)", err, c.baseURL)
	}
	defer resp.Body.Close()

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
