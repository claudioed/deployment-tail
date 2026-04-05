package cli

import (
	"runtime"
	"testing"
)

func TestNewAuthCmd(t *testing.T) {
	cmd := NewAuthCmd()

	if cmd.Use != "auth" {
		t.Errorf("Expected Use 'auth', got '%s'", cmd.Use)
	}

	// Verify subcommands are registered
	subcommands := []string{"login", "logout", "status"}
	for _, subcmd := range subcommands {
		found := false
		for _, c := range cmd.Commands() {
			if c.Name() == subcmd {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected subcommand '%s' to be registered", subcmd)
		}
	}
}

func TestNewLoginCmd(t *testing.T) {
	cmd := NewLoginCmd()

	if cmd.Use != "login" {
		t.Errorf("Expected Use 'login', got '%s'", cmd.Use)
	}

	// Verify --manual flag exists
	flag := cmd.Flags().Lookup("manual")
	if flag == nil {
		t.Error("Expected --manual flag to be defined")
	}
}

func TestNewLogoutCmd(t *testing.T) {
	cmd := NewLogoutCmd()

	if cmd.Use != "logout" {
		t.Errorf("Expected Use 'logout', got '%s'", cmd.Use)
	}
}

func TestNewAuthStatusCmd(t *testing.T) {
	cmd := NewAuthStatusCmd()

	if cmd.Use != "status" {
		t.Errorf("Expected Use 'status', got '%s'", cmd.Use)
	}
}

func TestOpenBrowser_Exists(t *testing.T) {
	// This test verifies the function exists and has correct signature
	err := openBrowser("https://example.com")

	// On supported platforms (darwin, linux, windows), this should not error
	// or should error due to browser not being available
	// We just verify it doesn't panic
	_ = err

	// Verify runtime.GOOS is one of supported platforms
	supportedPlatforms := []string{"darwin", "linux", "windows"}
	supported := false
	for _, platform := range supportedPlatforms {
		if runtime.GOOS == platform {
			supported = true
			break
		}
	}

	if !supported {
		t.Logf("Running on unsupported platform: %s", runtime.GOOS)
	}
}

func TestTokenStore_Integration(t *testing.T) {
	// Create a token store
	ts := NewTokenStore()

	// Verify it creates a proper path structure
	if ts.configDir == "" {
		t.Error("Expected configDir to be set")
	}

	if ts.tokenPath == "" {
		t.Error("Expected tokenPath to be set")
	}
}
