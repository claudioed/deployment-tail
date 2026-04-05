package cli

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestTokenStore_SaveAndLoad(t *testing.T) {
	// Create temp directory for testing
	tempDir := t.TempDir()

	ts := &TokenStore{
		configDir: tempDir,
		tokenPath: filepath.Join(tempDir, "auth.json"),
	}

	// Test data
	token := "test-jwt-token-123"
	email := "test@example.com"
	role := "deployer"
	expiresAt := time.Now().Add(24 * time.Hour)

	// Save token
	err := ts.SaveToken(token, email, role, expiresAt)
	if err != nil {
		t.Fatalf("SaveToken failed: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(ts.tokenPath); os.IsNotExist(err) {
		t.Fatal("Token file was not created")
	}

	// Load token
	loaded, err := ts.LoadToken()
	if err != nil {
		t.Fatalf("LoadToken failed: %v", err)
	}

	// Verify data
	if loaded.Token != token {
		t.Errorf("Expected token %s, got %s", token, loaded.Token)
	}
	if loaded.Email != email {
		t.Errorf("Expected email %s, got %s", email, loaded.Email)
	}
	if loaded.Role != role {
		t.Errorf("Expected role %s, got %s", role, loaded.Role)
	}
	// Allow 1 second difference for time comparison
	if loaded.ExpiresAt.Sub(expiresAt).Abs() > time.Second {
		t.Errorf("Expected expiresAt %v, got %v", expiresAt, loaded.ExpiresAt)
	}
}

func TestTokenStore_FilePermissions(t *testing.T) {
	tempDir := t.TempDir()

	ts := &TokenStore{
		configDir: tempDir,
		tokenPath: filepath.Join(tempDir, "auth.json"),
	}

	// Save token
	err := ts.SaveToken("token", "test@example.com", "deployer", time.Now().Add(time.Hour))
	if err != nil {
		t.Fatalf("SaveToken failed: %v", err)
	}

	// Check file permissions (0600 = -rw-------)
	fileInfo, err := os.Stat(ts.tokenPath)
	if err != nil {
		t.Fatalf("Failed to stat token file: %v", err)
	}

	mode := fileInfo.Mode().Perm()
	expected := os.FileMode(0600)
	if mode != expected {
		t.Errorf("Expected file permissions %o, got %o", expected, mode)
	}

	// Check directory permissions (0700 = drwx------)
	dirInfo, err := os.Stat(ts.configDir)
	if err != nil {
		t.Fatalf("Failed to stat config directory: %v", err)
	}

	dirMode := dirInfo.Mode().Perm()
	expectedDir := os.FileMode(0700)
	if dirMode != expectedDir {
		t.Errorf("Expected directory permissions %o, got %o", expectedDir, dirMode)
	}
}

func TestTokenStore_DeleteToken(t *testing.T) {
	tempDir := t.TempDir()

	ts := &TokenStore{
		configDir: tempDir,
		tokenPath: filepath.Join(tempDir, "auth.json"),
	}

	// Save token
	err := ts.SaveToken("token", "test@example.com", "deployer", time.Now().Add(time.Hour))
	if err != nil {
		t.Fatalf("SaveToken failed: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(ts.tokenPath); os.IsNotExist(err) {
		t.Fatal("Token file was not created")
	}

	// Delete token
	err = ts.DeleteToken()
	if err != nil {
		t.Fatalf("DeleteToken failed: %v", err)
	}

	// Verify file is deleted
	if _, err := os.Stat(ts.tokenPath); !os.IsNotExist(err) {
		t.Fatal("Token file was not deleted")
	}
}

func TestTokenStore_LoadToken_NotFound(t *testing.T) {
	tempDir := t.TempDir()

	ts := &TokenStore{
		configDir: tempDir,
		tokenPath: filepath.Join(tempDir, "auth.json"),
	}

	// Try to load non-existent token
	_, err := ts.LoadToken()
	if err == nil {
		t.Fatal("Expected error when loading non-existent token")
	}

	if !os.IsNotExist(err) {
		t.Errorf("Expected os.IsNotExist error, got: %v", err)
	}
}

func TestTokenStore_IsAuthenticated(t *testing.T) {
	tempDir := t.TempDir()

	ts := &TokenStore{
		configDir: tempDir,
		tokenPath: filepath.Join(tempDir, "auth.json"),
	}

	// Test: no token
	if ts.IsAuthenticated() {
		t.Error("Expected IsAuthenticated to return false when no token exists")
	}

	// Test: valid token
	err := ts.SaveToken("token", "test@example.com", "deployer", time.Now().Add(time.Hour))
	if err != nil {
		t.Fatalf("SaveToken failed: %v", err)
	}

	if !ts.IsAuthenticated() {
		t.Error("Expected IsAuthenticated to return true with valid token")
	}

	// Test: expired token
	err = ts.SaveToken("token", "test@example.com", "deployer", time.Now().Add(-time.Hour))
	if err != nil {
		t.Fatalf("SaveToken failed: %v", err)
	}

	if ts.IsAuthenticated() {
		t.Error("Expected IsAuthenticated to return false with expired token")
	}
}

func TestTokenStore_DeleteToken_NotExists(t *testing.T) {
	tempDir := t.TempDir()

	ts := &TokenStore{
		configDir: tempDir,
		tokenPath: filepath.Join(tempDir, "auth.json"),
	}

	// Delete non-existent token should not error
	err := ts.DeleteToken()
	if err != nil {
		t.Errorf("DeleteToken should not error on non-existent file: %v", err)
	}
}
