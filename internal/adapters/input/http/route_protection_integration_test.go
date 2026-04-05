// +build integration

package http

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/claudioed/deployment-tail/internal/adapters/input/http/middleware"
	"github.com/claudioed/deployment-tail/internal/domain/user"
	"github.com/claudioed/deployment-tail/internal/infrastructure/jwt"
	"github.com/go-chi/chi/v5"
)

// TestRouteProtection_UnauthenticatedRequests verifies that routes requiring authentication
// reject unauthenticated requests with 401 Unauthorized
func TestRouteProtection_UnauthenticatedRequests(t *testing.T) {
	// Create JWT service for token generation
	jwtService, err := jwt.NewJWTService(jwt.Config{
		Secret: "test-secret-key-for-integration-tests-12345",
		Expiry: 1 * time.Hour,
		Issuer: "test-issuer",
	})
	if err != nil {
		t.Fatalf("failed to create JWT service: %v", err)
	}

	// Create revocation store (in-memory only for tests)
	revocationStore := jwt.NewRevocationStore(nil)

	// Create auth middleware
	authMiddleware := middleware.NewAuthenticationMiddleware(jwtService, revocationStore)

	// Create test router with protected routes
	router := chi.NewRouter()

	// Protected routes (require authentication)
	router.Group(func(r chi.Router) {
		r.Use(authMiddleware.Authenticate)
		r.Get("/api/v1/schedules", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"schedules":[]}`))
		})
		r.Post("/api/v1/schedules", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusCreated)
		})
	})

	tests := []struct {
		name           string
		method         string
		path           string
		expectedStatus int
	}{
		{
			name:           "GET /api/v1/schedules without auth",
			method:         "GET",
			path:           "/api/v1/schedules",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "POST /api/v1/schedules without auth",
			method:         "POST",
			path:           "/api/v1/schedules",
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

// TestRouteProtection_AuthenticatedRequests verifies that valid authenticated requests
// are allowed through the middleware
func TestRouteProtection_AuthenticatedRequests(t *testing.T) {
	// Create JWT service
	jwtService, err := jwt.NewJWTService(jwt.Config{
		Secret: "test-secret-key-for-integration-tests-12345",
		Expiry: 1 * time.Hour,
		Issuer: "test-issuer",
	})
	if err != nil {
		t.Fatalf("failed to create JWT service: %v", err)
	}

	// Create revocation store
	revocationStore := jwt.NewRevocationStore(nil)

	// Create test user
	googleID, _ := user.NewGoogleID("google-123")
	email, _ := user.NewEmail("test@example.com")
	userName, _ := user.NewUserName("Test User")
	userRole, _ := user.NewRole(user.RoleDeployer)
	testUser, _ := user.NewUser(googleID, email, userName, userRole)

	// Generate valid token
	token, err := jwtService.GenerateToken(testUser)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	// Create auth middleware
	authMiddleware := middleware.NewAuthenticationMiddleware(jwtService, revocationStore)

	// Create test router with protected routes
	router := chi.NewRouter()

	router.Group(func(r chi.Router) {
		r.Use(authMiddleware.Authenticate)
		r.Get("/api/v1/schedules", func(w http.ResponseWriter, r *http.Request) {
			// Verify user is in context
			user, err := middleware.UserFromContext(r.Context())
			if err != nil {
				t.Errorf("expected user in context, got error: %v", err)
			}
			if user.Email().String() != "test@example.com" {
				t.Errorf("expected email test@example.com, got %s", user.Email().String())
			}
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"schedules":[]}`))
		})
	})

	req := httptest.NewRequest("GET", "/api/v1/schedules", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

// TestRouteProtection_RevokedTokenRejected verifies that revoked tokens are rejected
func TestRouteProtection_RevokedTokenRejected(t *testing.T) {
	// Create JWT service
	jwtService, err := jwt.NewJWTService(jwt.Config{
		Secret: "test-secret-key-for-integration-tests-12345",
		Expiry: 1 * time.Hour,
		Issuer: "test-issuer",
	})
	if err != nil {
		t.Fatalf("failed to create JWT service: %v", err)
	}

	// Create revocation store
	revocationStore := jwt.NewRevocationStore(nil)

	// Create test user
	googleID, _ := user.NewGoogleID("google-123")
	email, _ := user.NewEmail("test@example.com")
	userName, _ := user.NewUserName("Test User")
	userRole, _ := user.NewRole(user.RoleDeployer)
	testUser, _ := user.NewUser(googleID, email, userName, userRole)

	// Generate token
	token, err := jwtService.GenerateToken(testUser)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	// Revoke the token
	tokenHash := jwtService.HashToken(token)
	ctx := context.Background()
	expiresAt := time.Now().Add(1 * time.Hour)
	err = revocationStore.AddToBlacklist(ctx, tokenHash, testUser.ID().String(), expiresAt)
	if err != nil {
		t.Fatalf("failed to revoke token: %v", err)
	}

	// Create auth middleware
	authMiddleware := middleware.NewAuthenticationMiddleware(jwtService, revocationStore)

	// Create test router
	router := chi.NewRouter()
	router.Group(func(r chi.Router) {
		r.Use(authMiddleware.Authenticate)
		r.Get("/api/v1/schedules", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
	})

	req := httptest.NewRequest("GET", "/api/v1/schedules", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401 for revoked token, got %d", w.Code)
	}
}

// TestRouteProtection_RoleBasedAccess verifies that role-based access control works correctly
func TestRouteProtection_RoleBasedAccess(t *testing.T) {
	// Create JWT service
	jwtService, err := jwt.NewJWTService(jwt.Config{
		Secret: "test-secret-key-for-integration-tests-12345",
		Expiry: 1 * time.Hour,
		Issuer: "test-issuer",
	})
	if err != nil {
		t.Fatalf("failed to create JWT service: %v", err)
	}

	// Create revocation store
	revocationStore := jwt.NewRevocationStore(nil)

	// Create auth middleware
	authMiddleware := middleware.NewAuthenticationMiddleware(jwtService, revocationStore)

	tests := []struct {
		name              string
		userRole          string
		requiredRole      string
		path              string
		expectedStatus    int
		expectedBodyError string
	}{
		{
			name:           "viewer cannot create schedules",
			userRole:       user.RoleViewer,
			requiredRole:   user.RoleDeployer,
			path:           "/api/v1/schedules",
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "deployer can create schedules",
			userRole:       user.RoleDeployer,
			requiredRole:   user.RoleDeployer,
			path:           "/api/v1/schedules",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "admin can create schedules",
			userRole:       user.RoleAdmin,
			requiredRole:   user.RoleDeployer,
			path:           "/api/v1/schedules",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "viewer cannot access admin routes",
			userRole:       user.RoleViewer,
			requiredRole:   user.RoleAdmin,
			path:           "/api/v1/users",
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "deployer cannot access admin routes",
			userRole:       user.RoleDeployer,
			requiredRole:   user.RoleAdmin,
			path:           "/api/v1/users",
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "admin can access admin routes",
			userRole:       user.RoleAdmin,
			requiredRole:   user.RoleAdmin,
			path:           "/api/v1/users",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test user with specific role
			googleID, _ := user.NewGoogleID("google-" + tt.userRole)
			email, _ := user.NewEmail(tt.userRole + "@example.com")
			userName, _ := user.NewUserName("Test " + tt.userRole)
			userRole, _ := user.NewRole(tt.userRole)
			testUser, _ := user.NewUser(googleID, email, userName, userRole)

			// Generate token
			token, err := jwtService.GenerateToken(testUser)
			if err != nil {
				t.Fatalf("failed to generate token: %v", err)
			}

			// Create test router with role-protected routes
			router := chi.NewRouter()
			router.Group(func(r chi.Router) {
				r.Use(authMiddleware.Authenticate)
				r.Use(middleware.RequireRole(tt.requiredRole))
				r.Post(tt.path, func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(`{"success":true}`))
				})
			})

			req := httptest.NewRequest("POST", tt.path, strings.NewReader(`{}`))
			req.Header.Set("Authorization", "Bearer "+token)
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			// For forbidden responses, verify error message
			if w.Code == http.StatusForbidden {
				var response map[string]interface{}
				if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
					t.Errorf("failed to decode response: %v", err)
				}
				if _, ok := response["error"]; !ok {
					t.Error("expected error field in forbidden response")
				}
			}
		})
	}
}

// TestRouteProtection_ExpiredTokenRejected verifies that expired tokens are rejected
func TestRouteProtection_ExpiredTokenRejected(t *testing.T) {
	// Create JWT service with very short expiry
	jwtService, err := jwt.NewJWTService(jwt.Config{
		Secret: "test-secret-key-for-integration-tests-12345",
		Expiry: 1 * time.Millisecond, // Token expires almost immediately
		Issuer: "test-issuer",
	})
	if err != nil {
		t.Fatalf("failed to create JWT service: %v", err)
	}

	// Create revocation store
	revocationStore := jwt.NewRevocationStore(nil)

	// Create test user
	googleID, _ := user.NewGoogleID("google-123")
	email, _ := user.NewEmail("test@example.com")
	userName, _ := user.NewUserName("Test User")
	userRole, _ := user.NewRole(user.RoleDeployer)
	testUser, _ := user.NewUser(googleID, email, userName, userRole)

	// Generate token
	token, err := jwtService.GenerateToken(testUser)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	// Wait for token to expire
	time.Sleep(10 * time.Millisecond)

	// Create auth middleware
	authMiddleware := middleware.NewAuthenticationMiddleware(jwtService, revocationStore)

	// Create test router
	router := chi.NewRouter()
	router.Group(func(r chi.Router) {
		r.Use(authMiddleware.Authenticate)
		r.Get("/api/v1/schedules", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
	})

	req := httptest.NewRequest("GET", "/api/v1/schedules", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401 for expired token, got %d", w.Code)
	}
}

// TestRouteProtection_MalformedTokenRejected verifies that malformed tokens are rejected
func TestRouteProtection_MalformedTokenRejected(t *testing.T) {
	// Create JWT service
	jwtService, err := jwt.NewJWTService(jwt.Config{
		Secret: "test-secret-key-for-integration-tests-12345",
		Expiry: 1 * time.Hour,
		Issuer: "test-issuer",
	})
	if err != nil {
		t.Fatalf("failed to create JWT service: %v", err)
	}

	// Create revocation store
	revocationStore := jwt.NewRevocationStore(nil)

	// Create auth middleware
	authMiddleware := middleware.NewAuthenticationMiddleware(jwtService, revocationStore)

	// Create test router
	router := chi.NewRouter()
	router.Group(func(r chi.Router) {
		r.Use(authMiddleware.Authenticate)
		r.Get("/api/v1/schedules", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
	})

	tests := []struct {
		name          string
		authorization string
		description   string
	}{
		{
			name:          "missing Bearer prefix",
			authorization: "not-a-valid-token",
			description:   "Authorization header without Bearer prefix",
		},
		{
			name:          "malformed JWT",
			authorization: "Bearer invalid.jwt.token",
			description:   "Malformed JWT token",
		},
		{
			name:          "empty token",
			authorization: "Bearer ",
			description:   "Empty token after Bearer",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/api/v1/schedules", nil)
			req.Header.Set("Authorization", tt.authorization)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != http.StatusUnauthorized {
				t.Errorf("%s: expected status 401, got %d", tt.description, w.Code)
			}
		})
	}
}
