package bdd

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/cucumber/godog"
	"github.com/tidwall/gjson"
)

func RegisterHTTPSteps(ctx *godog.ScenarioContext) {
	s := &httpSteps{}

	ctx.Step(`^I (GET|POST|PUT|DELETE|PATCH) "([^"]+)"$`, s.iRequestWithoutBody)
	ctx.Step(`^I (POST|PUT|PATCH) "([^"]+)" with body:$`, s.iRequestWithBody)
	ctx.Step(`^the response JSON field "([^"]+)" (equals|contains) "([^"]+)"$`, s.theResponseJSONField)
}

type httpSteps struct{}

func (s *httpSteps) iRequestWithoutBody(ctx context.Context, method, path string) error {
	return s.doHTTPRequest(ctx, method, path, "")
}

func (s *httpSteps) iRequestWithBody(ctx context.Context, method, path string, bodyDoc *godog.DocString) error {
	return s.doHTTPRequest(ctx, method, path, bodyDoc.Content)
}

func (s *httpSteps) doHTTPRequest(ctx context.Context, method, path, body string) error {
	w := getWorld(ctx)

	// Lazy-start HTTP server if not already started
	if w.HTTPServer == nil {
		w.startHTTPServer()

		// Re-mint token for current user if authenticated
		if w.CurrentUser != nil && w.JWTService != nil {
			token, err := w.JWTService.GenerateToken(w.CurrentUser)
			if err != nil {
				return fmt.Errorf("failed to mint JWT after lazy server start: %w", err)
			}
			w.CurrentToken = token
		}
	}

	url := w.HTTPServer.URL + path

	var bodyReader io.Reader
	if body != "" {
		bodyReader = bytes.NewBufferString(body)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Add Authorization header if token exists
	if w.CurrentToken != "" {
		req.Header.Set("Authorization", "Bearer "+w.CurrentToken)
	}

	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	w.LastStatusCode = resp.StatusCode

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}
	w.LastResponseBody = respBody

	return nil
}

func (s *httpSteps) theResponseJSONField(ctx context.Context, jsonPath, operator, expected string) error {
	w := getWorld(ctx)

	if len(w.LastResponseBody) == 0 {
		return fmt.Errorf("no response body available")
	}

	// Parse JSON and extract field
	result := gjson.GetBytes(w.LastResponseBody, jsonPath)
	if !result.Exists() {
		return fmt.Errorf("JSON field %q not found in response: %s", jsonPath, string(w.LastResponseBody))
	}

	actual := result.String()

	switch operator {
	case "equals":
		if actual != expected {
			return fmt.Errorf("expected %q to equal %q but got %q", jsonPath, expected, actual)
		}
	case "contains":
		if !strings.Contains(actual, expected) {
			return fmt.Errorf("expected %q to contain %q but got %q", jsonPath, expected, actual)
		}
	default:
		return fmt.Errorf("unknown operator %q", operator)
	}

	return nil
}
