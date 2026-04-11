package jwt

import (
	"testing"
	"time"

	"github.com/claudioed/deployment-tail/internal/domain/user"
)

func TestNewJWTService(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: Config{
				Secret: "this-is-a-very-secure-secret-key-with-32-chars",
				Expiry: 24 * time.Hour,
				Issuer: "test-issuer",
			},
			wantErr: false,
		},
		{
			name: "secret too short",
			config: Config{
				Secret: "short",
				Expiry: 24 * time.Hour,
				Issuer: "test-issuer",
			},
			wantErr: true,
		},
		{
			name: "missing secret",
			config: Config{
				Expiry: 24 * time.Hour,
				Issuer: "test-issuer",
			},
			wantErr: true,
		},
		{
			name: "zero expiry",
			config: Config{
				Secret: "this-is-a-very-secure-secret-key-with-32-chars",
				Expiry: 0,
				Issuer: "test-issuer",
			},
			wantErr: true,
		},
		{
			name: "negative expiry",
			config: Config{
				Secret: "this-is-a-very-secure-secret-key-with-32-chars",
				Expiry: -1 * time.Hour,
				Issuer: "test-issuer",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewJWTService(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewJWTService() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGenerateToken(t *testing.T) {
	service, err := NewJWTService(Config{
		Secret: "this-is-a-very-secure-secret-key-with-32-chars",
		Expiry: 24 * time.Hour,
		Issuer: "test-issuer",
	})
	if err != nil {
		t.Fatalf("Failed to create JWT service: %v", err)
	}

	googleID, _ := user.NewGoogleID("108123456789012345678")
	email, _ := user.NewEmail("test@example.com")
	name, _ := user.NewUserName("Test User")
	role, _ := user.NewRole(user.RoleDeployer)
	u, _ := user.NewUser(googleID, email, name, role)

	token, err := service.GenerateToken(u)
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}

	if token == "" {
		t.Error("GenerateToken() returned empty token")
	}

	// Token should have three parts separated by dots (header.payload.signature)
	parts := 0
	for _, c := range token {
		if c == '.' {
			parts++
		}
	}
	if parts != 2 {
		t.Errorf("Token should have 3 parts, found %d dots", parts)
	}
}

func TestValidateToken(t *testing.T) {
	service, _ := NewJWTService(Config{
		Secret: "this-is-a-very-secure-secret-key-with-32-chars",
		Expiry: 24 * time.Hour,
		Issuer: "test-issuer",
	})

	googleID, _ := user.NewGoogleID("108123456789012345678")
	email, _ := user.NewEmail("test@example.com")
	name, _ := user.NewUserName("Test User")
	role, _ := user.NewRole(user.RoleDeployer)
	u, _ := user.NewUser(googleID, email, name, role)

	token, _ := service.GenerateToken(u)

	// Validate the token
	claims, err := service.ValidateToken(token)
	if err != nil {
		t.Fatalf("ValidateToken() error = %v", err)
	}

	if claims.UserID != u.ID().String() {
		t.Errorf("Expected UserID %s, got %s", u.ID().String(), claims.UserID)
	}
	if claims.Email != email.String() {
		t.Errorf("Expected Email %s, got %s", email.String(), claims.Email)
	}
	if claims.Role != role.String() {
		t.Errorf("Expected Role %s, got %s", role.String(), claims.Role)
	}
}

func TestValidateToken_Invalid(t *testing.T) {
	service, _ := NewJWTService(Config{
		Secret: "this-is-a-very-secure-secret-key-with-32-chars",
		Expiry: 24 * time.Hour,
		Issuer: "test-issuer",
	})

	tests := []struct {
		name  string
		token string
	}{
		{"empty token", ""},
		{"malformed token", "not.a.valid.jwt"},
		{"invalid signature", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.ValidateToken(tt.token)
			if err == nil {
				t.Error("ValidateToken() expected error, got nil")
			}
		})
	}
}

func TestValidateToken_Expired(t *testing.T) {
	// Create service with very short expiry
	service, _ := NewJWTService(Config{
		Secret: "this-is-a-very-secure-secret-key-with-32-chars",
		Expiry: 1 * time.Millisecond,
		Issuer: "test-issuer",
	})

	googleID, _ := user.NewGoogleID("108123456789012345678")
	email, _ := user.NewEmail("test@example.com")
	name, _ := user.NewUserName("Test User")
	role, _ := user.NewRole(user.RoleDeployer)
	u, _ := user.NewUser(googleID, email, name, role)

	token, _ := service.GenerateToken(u)

	// Wait for token to expire
	time.Sleep(10 * time.Millisecond)

	_, err := service.ValidateToken(token)
	if err == nil {
		t.Error("ValidateToken() expected error for expired token, got nil")
	}
}

func TestRefreshToken(t *testing.T) {
	service, _ := NewJWTService(Config{
		Secret: "this-is-a-very-secure-secret-key-with-32-chars",
		Expiry: 24 * time.Hour,
		Issuer: "test-issuer",
	})

	googleID, _ := user.NewGoogleID("108123456789012345678")
	email, _ := user.NewEmail("test@example.com")
	name, _ := user.NewUserName("Test User")
	role, _ := user.NewRole(user.RoleDeployer)
	u, _ := user.NewUser(googleID, email, name, role)

	originalToken, _ := service.GenerateToken(u)
	time.Sleep(1 * time.Second) // Ensure different issued time (JWT timestamps are second-precision)

	refreshedToken, err := service.RefreshToken(originalToken)
	if err != nil {
		t.Fatalf("RefreshToken() error = %v", err)
	}

	if refreshedToken == originalToken {
		t.Error("RefreshToken() should return a different token")
	}

	// Validate refreshed token
	claims, err := service.ValidateToken(refreshedToken)
	if err != nil {
		t.Fatalf("Refreshed token validation error = %v", err)
	}

	if claims.UserID != u.ID().String() {
		t.Errorf("Refreshed token has wrong UserID")
	}
}

func TestParseClaims(t *testing.T) {
	service, _ := NewJWTService(Config{
		Secret: "this-is-a-very-secure-secret-key-with-32-chars",
		Expiry: 24 * time.Hour,
		Issuer: "test-issuer",
	})

	googleID, _ := user.NewGoogleID("108123456789012345678")
	email, _ := user.NewEmail("test@example.com")
	name, _ := user.NewUserName("Test User")
	role, _ := user.NewRole(user.RoleAdmin)
	u, _ := user.NewUser(googleID, email, name, role)

	token, _ := service.GenerateToken(u)

	claims, err := service.ParseClaims(token)
	if err != nil {
		t.Fatalf("ParseClaims() error = %v", err)
	}

	if claims.Email != email.String() {
		t.Errorf("Expected email %s, got %s", email.String(), claims.Email)
	}
	if claims.Role != role.String() {
		t.Errorf("Expected role %s, got %s", role.String(), claims.Role)
	}
}

func TestHashToken(t *testing.T) {
	service, _ := NewJWTService(Config{
		Secret: "this-is-a-very-secure-secret-key-with-32-chars",
		Expiry: 24 * time.Hour,
		Issuer: "test-issuer",
	})

	token1 := "test-token-1"
	token2 := "test-token-2"

	hash1 := service.HashToken(token1)
	hash2 := service.HashToken(token2)

	// Hashes should be different
	if hash1 == hash2 {
		t.Error("Different tokens should produce different hashes")
	}

	// Hash should be consistent
	hash1Again := service.HashToken(token1)
	if hash1 != hash1Again {
		t.Error("Same token should produce same hash")
	}

	// Hash should be 64 characters (SHA256 hex)
	if len(hash1) != 64 {
		t.Errorf("Hash should be 64 characters, got %d", len(hash1))
	}
}

func TestValidateToken_WrongSecret(t *testing.T) {
	service1, _ := NewJWTService(Config{
		Secret: "this-is-a-very-secure-secret-key-with-32-chars",
		Expiry: 24 * time.Hour,
		Issuer: "test-issuer",
	})

	service2, _ := NewJWTService(Config{
		Secret: "different-secret-key-that-is-32-chars-long",
		Expiry: 24 * time.Hour,
		Issuer: "test-issuer",
	})

	googleID, _ := user.NewGoogleID("108123456789012345678")
	email, _ := user.NewEmail("test@example.com")
	name, _ := user.NewUserName("Test User")
	role, _ := user.NewRole(user.RoleDeployer)
	u, _ := user.NewUser(googleID, email, name, role)

	token, _ := service1.GenerateToken(u)

	// Try to validate with different secret
	_, err := service2.ValidateToken(token)
	if err == nil {
		t.Error("ValidateToken() should fail with wrong secret")
	}
}
