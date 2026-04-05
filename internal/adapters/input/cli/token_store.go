package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// TokenData represents the stored authentication token
type TokenData struct {
	Token     string    `json:"token"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	ExpiresAt time.Time `json:"expires_at"`
}

// TokenStore manages local token storage
type TokenStore struct {
	configDir string
	tokenPath string
}

// NewTokenStore creates a new token store
func NewTokenStore() *TokenStore {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}

	configDir := filepath.Join(homeDir, ".deployment-tail")
	tokenPath := filepath.Join(configDir, "auth.json")

	return &TokenStore{
		configDir: configDir,
		tokenPath: tokenPath,
	}
}

// SaveToken saves the authentication token to disk
func (ts *TokenStore) SaveToken(token, email, role string, expiresAt time.Time) error {
	// Create config directory if it doesn't exist
	if err := os.MkdirAll(ts.configDir, 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Create token data
	tokenData := TokenData{
		Token:     token,
		Email:     email,
		Role:      role,
		ExpiresAt: expiresAt,
	}

	// Marshal to JSON
	data, err := json.MarshalIndent(tokenData, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal token data: %w", err)
	}

	// Write to file with secure permissions (0600 = read/write for owner only)
	if err := os.WriteFile(ts.tokenPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write token file: %w", err)
	}

	// Verify file permissions
	if err := ts.enforcePermissions(); err != nil {
		return fmt.Errorf("failed to enforce permissions: %w", err)
	}

	return nil
}

// LoadToken loads the authentication token from disk
func (ts *TokenStore) LoadToken() (*TokenData, error) {
	// Read token file
	data, err := os.ReadFile(ts.tokenPath)
	if err != nil {
		return nil, err
	}

	// Unmarshal JSON
	var tokenData TokenData
	if err := json.Unmarshal(data, &tokenData); err != nil {
		return nil, fmt.Errorf("failed to parse token file: %w", err)
	}

	return &tokenData, nil
}

// DeleteToken removes the authentication token from disk
func (ts *TokenStore) DeleteToken() error {
	if err := os.Remove(ts.tokenPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete token file: %w", err)
	}
	return nil
}

// enforcePermissions ensures the token file has secure permissions
func (ts *TokenStore) enforcePermissions() error {
	// Ensure auth.json is 0600 (read/write for owner only)
	if err := os.Chmod(ts.tokenPath, 0600); err != nil {
		return fmt.Errorf("failed to set file permissions: %w", err)
	}

	// Ensure directory is 0700 (read/write/execute for owner only)
	if err := os.Chmod(ts.configDir, 0700); err != nil {
		return fmt.Errorf("failed to set directory permissions: %w", err)
	}

	return nil
}

// IsAuthenticated checks if a valid token exists
func (ts *TokenStore) IsAuthenticated() bool {
	token, err := ts.LoadToken()
	if err != nil {
		return false
	}

	// Check if token is expired
	if time.Now().After(token.ExpiresAt) {
		return false
	}

	return true
}
