package middleware

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/claudioed/deployment-tail/internal/domain/user"
	"github.com/claudioed/deployment-tail/internal/infrastructure/jwt"
	jwtgo "github.com/golang-jwt/jwt/v5"
)

// Mock JWT Service
type MockJWTService struct {
	validateTokenFunc func(string) (*jwt.Claims, error)
	hashTokenFunc     func(string) string
}

func (m *MockJWTService) ValidateToken(tokenString string) (*jwt.Claims, error) {
	if m.validateTokenFunc != nil {
		return m.validateTokenFunc(tokenString)
	}
	return nil, fmt.Errorf("not implemented")
}

func (m *MockJWTService) HashToken(tokenString string) string {
	if m.hashTokenFunc != nil {
		return m.hashTokenFunc(tokenString)
	}
	return "mocked-hash"
}

func (m *MockJWTService) GenerateToken(u *user.User) (string, error) {
	return "mocked-token", nil
}

func (m *MockJWTService) ParseClaims(tokenString string) (*jwt.Claims, error) {
	return m.ValidateToken(tokenString)
}

func (m *MockJWTService) RefreshToken(tokenString string) (string, error) {
	return "refreshed-token", nil
}

// Mock Revocation Store
type MockRevocationStore struct {
	isRevokedFunc func(string) bool
}

func (m *MockRevocationStore) IsRevoked(tokenHash string) bool {
	if m.isRevokedFunc != nil {
		return m.isRevokedFunc(tokenHash)
	}
	return false
}

func (m *MockRevocationStore) AddToBlacklist(ctx context.Context, tokenHash, userID string, expiresAt time.Time) error {
	return nil
}

func (m *MockRevocationStore) Start(ctx context.Context) error {
	return nil
}

func (m *MockRevocationStore) Stop() {}

func (m *MockRevocationStore) GetBlacklistSize() int {
	return 0
}

// Mock User Repository
type MockUserRepository struct {
	users       map[string]*user.User
	findByIDErr error
}

func (m *MockUserRepository) FindByID(ctx context.Context, id user.UserID) (*user.User, error) {
	if m.findByIDErr != nil {
		return nil, m.findByIDErr
	}
	u, ok := m.users[id.String()]
	if !ok {
		return nil, user.ErrUserNotFound{ID: id.String(), SearchType: "id"}
	}
	return u, nil
}

func (m *MockUserRepository) FindByGoogleID(ctx context.Context, googleID user.GoogleID) (*user.User, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *MockUserRepository) FindByEmail(ctx context.Context, email user.Email) (*user.User, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *MockUserRepository) Create(ctx context.Context, u *user.User) error {
	return nil
}

func (m *MockUserRepository) Update(ctx context.Context, u *user.User) error {
	return nil
}

func (m *MockUserRepository) List(ctx context.Context, filters user.ListFilters) ([]*user.User, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *MockUserRepository) UpdateRole(ctx context.Context, userID user.UserID, role user.Role) error {
	return nil
}

func (m *MockUserRepository) UpdateLastLogin(ctx context.Context, userID user.UserID) error {
	return nil
}

// Test helper functions
func createTestUser(role string) *user.User {
	googleID, _ := user.NewGoogleID("test-google-id")
	email, _ := user.NewEmail("test@example.com")
	name, _ := user.NewUserName("Test User")
	r, _ := user.NewRole(role)
	u, _ := user.NewUser(googleID, email, name, r)
	return u
}

func TestExtractBearerToken(t *testing.T) {
	tests := []struct {
		name        string
		authHeader  string
		expectToken string
		expectError bool
	}{
		{
			name:        "valid bearer token",
			authHeader:  "Bearer valid-token-123",
			expectToken: "valid-token-123",
			expectError: false,
		},
		{
			name:        "valid bearer token with extra spaces",
			authHeader:  "Bearer   valid-token-123  ",
			expectToken: "valid-token-123",
			expectError: false,
		},
		{
			name:        "lowercase bearer",
			authHeader:  "bearer valid-token-123",
			expectToken: "valid-token-123",
			expectError: false,
		},
		{
			name:        "missing header",
			authHeader:  "",
			expectToken: "",
			expectError: true,
		},
		{
			name:        "missing token",
			authHeader:  "Bearer",
			expectToken: "",
			expectError: true,
		},
		{
			name:        "empty token",
			authHeader:  "Bearer   ",
			expectToken: "",
			expectError: true,
		},
		{
			name:        "invalid format (no Bearer prefix)",
			authHeader:  "valid-token-123",
			expectToken: "",
			expectError: true,
		},
		{
			name:        "invalid format (wrong prefix)",
			authHeader:  "Basic valid-token-123",
			expectToken: "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			token, err := extractBearerToken(req)

			if tt.expectError {
				if err == nil {
					t.Error("expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if token != tt.expectToken {
					t.Errorf("expected token %q, got %q", tt.expectToken, token)
				}
			}
		})
	}
}

func TestUserToContext_UserFromContext(t *testing.T) {
	testUser := createTestUser(user.RoleDeployer)

	// Test adding user to context
	ctx := context.Background()
	ctxWithUser := userToContext(ctx, testUser)

	// Test retrieving user from context
	retrievedUser, err := UserFromContext(ctxWithUser)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !retrievedUser.ID().Equals(testUser.ID()) {
		t.Error("retrieved user does not match original user")
	}
}

func TestUserFromContext_NoUser(t *testing.T) {
	ctx := context.Background()

	_, err := UserFromContext(ctx)
	if err == nil {
		t.Error("expected error when no user in context")
	}
}

func TestAuthenticationMiddleware_Success(t *testing.T) {
	testUser := createTestUser(user.RoleDeployer)

	jwtService := &MockJWTService{
		validateTokenFunc: func(token string) (*jwt.Claims, error) {
			return &jwt.Claims{
				UserID: testUser.ID().String(),
				Email:  testUser.Email().String(),
				Role:   testUser.Role().String(),
				RegisteredClaims: jwtgo.RegisteredClaims{
					ExpiresAt: jwtgo.NewNumericDate(time.Now().Add(1 * time.Hour)),
				},
			}, nil
		},
		hashTokenFunc: func(token string) string {
			return "hashed-token"
		},
	}

	revocationStore := &MockRevocationStore{
		isRevokedFunc: func(hash string) bool {
			return false
		},
	}

	userRepo := &MockUserRepository{
		users: map[string]*user.User{
			testUser.ID().String(): testUser,
		},
	}

	middleware := NewAuthenticationMiddleware(jwtService, revocationStore, userRepo)

	// Create a test handler that checks if user is in context
	handlerCalled := false
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		u, err := UserFromContext(r.Context())
		if err != nil {
			t.Errorf("expected user in context, got error: %v", err)
		}
		if !u.ID().Equals(testUser.ID()) {
			t.Error("user in context does not match expected user")
		}
		w.WriteHeader(http.StatusOK)
	})

	// Create request with valid token
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer valid-token")

	rr := httptest.NewRecorder()

	// Execute middleware
	middleware.Authenticate(testHandler).ServeHTTP(rr, req)

	if !handlerCalled {
		t.Error("handler was not called")
	}

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}
}

func TestAuthenticationMiddleware_MissingToken(t *testing.T) {
	middleware := NewAuthenticationMiddleware(
		&MockJWTService{},
		&MockRevocationStore{},
		&MockUserRepository{},
	)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	// No Authorization header

	rr := httptest.NewRecorder()

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called")
	})

	middleware.Authenticate(testHandler).ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", rr.Code)
	}
}

func TestAuthenticationMiddleware_InvalidToken(t *testing.T) {
	jwtService := &MockJWTService{
		validateTokenFunc: func(token string) (*jwt.Claims, error) {
			return nil, fmt.Errorf("invalid signature")
		},
	}

	middleware := NewAuthenticationMiddleware(
		jwtService,
		&MockRevocationStore{},
		&MockUserRepository{},
	)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")

	rr := httptest.NewRecorder()

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called")
	})

	middleware.Authenticate(testHandler).ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", rr.Code)
	}
}

func TestAuthenticationMiddleware_RevokedToken(t *testing.T) {
	testUser := createTestUser(user.RoleDeployer)

	jwtService := &MockJWTService{
		validateTokenFunc: func(token string) (*jwt.Claims, error) {
			return &jwt.Claims{
				UserID: testUser.ID().String(),
				Email:  testUser.Email().String(),
				Role:   testUser.Role().String(),
			}, nil
		},
		hashTokenFunc: func(token string) string {
			return "revoked-hash"
		},
	}

	revocationStore := &MockRevocationStore{
		isRevokedFunc: func(hash string) bool {
			return hash == "revoked-hash"
		},
	}

	middleware := NewAuthenticationMiddleware(
		jwtService,
		revocationStore,
		&MockUserRepository{},
	)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer revoked-token")

	rr := httptest.NewRecorder()

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called")
	})

	middleware.Authenticate(testHandler).ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", rr.Code)
	}
}

func TestAuthenticationMiddleware_UserNotFound(t *testing.T) {
	testUser := createTestUser(user.RoleDeployer)

	jwtService := &MockJWTService{
		validateTokenFunc: func(token string) (*jwt.Claims, error) {
			return &jwt.Claims{
				UserID: testUser.ID().String(),
				Email:  testUser.Email().String(),
				Role:   testUser.Role().String(),
			}, nil
		},
		hashTokenFunc: func(token string) string {
			return "hashed"
		},
	}

	revocationStore := &MockRevocationStore{
		isRevokedFunc: func(hash string) bool {
			return false
		},
	}

	userRepo := &MockUserRepository{
		users:       map[string]*user.User{}, // Empty - user not found
		findByIDErr: user.ErrUserNotFound{ID: testUser.ID().String(), SearchType: "id"},
	}

	middleware := NewAuthenticationMiddleware(jwtService, revocationStore, userRepo)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer valid-token")

	rr := httptest.NewRecorder()

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called")
	})

	middleware.Authenticate(testHandler).ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", rr.Code)
	}
}

func TestRequireRole_Success(t *testing.T) {
	testUser := createTestUser(user.RoleDeployer)

	handlerCalled := false
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	})

	// Create context with user
	ctx := userToContext(context.Background(), testUser)
	req := httptest.NewRequest(http.MethodGet, "/test", nil).WithContext(ctx)

	rr := httptest.NewRecorder()

	// Require deployer role
	RequireRole(user.RoleDeployer)(testHandler).ServeHTTP(rr, req)

	if !handlerCalled {
		t.Error("handler was not called")
	}

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}
}

func TestRequireRole_MultipleRoles(t *testing.T) {
	testUser := createTestUser(user.RoleDeployer)

	handlerCalled := false
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	})

	ctx := userToContext(context.Background(), testUser)
	req := httptest.NewRequest(http.MethodGet, "/test", nil).WithContext(ctx)

	rr := httptest.NewRecorder()

	// Require deployer OR admin role
	RequireRole(user.RoleDeployer, user.RoleAdmin)(testHandler).ServeHTTP(rr, req)

	if !handlerCalled {
		t.Error("handler was not called")
	}

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}
}

func TestRequireRole_InsufficientPermissions(t *testing.T) {
	testUser := createTestUser(user.RoleViewer)

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called")
	})

	ctx := userToContext(context.Background(), testUser)
	req := httptest.NewRequest(http.MethodGet, "/test", nil).WithContext(ctx)

	rr := httptest.NewRecorder()

	// Require admin role (user is viewer)
	RequireRole(user.RoleAdmin)(testHandler).ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("expected status 403, got %d", rr.Code)
	}
}

func TestRequireRole_NoUserInContext(t *testing.T) {
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	// No user in context

	rr := httptest.NewRecorder()

	RequireRole(user.RoleAdmin)(testHandler).ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", rr.Code)
	}
}

func TestRequireRole_ViewerCannotAccessAdminRoute(t *testing.T) {
	viewer := createTestUser(user.RoleViewer)

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called for viewer")
	})

	ctx := userToContext(context.Background(), viewer)
	req := httptest.NewRequest(http.MethodGet, "/admin/users", nil).WithContext(ctx)

	rr := httptest.NewRecorder()

	RequireRole(user.RoleAdmin)(testHandler).ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("expected status 403, got %d", rr.Code)
	}
}

func TestRequireRole_DeployerCanAccessDeployerRoute(t *testing.T) {
	deployer := createTestUser(user.RoleDeployer)

	handlerCalled := false
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	})

	ctx := userToContext(context.Background(), deployer)
	req := httptest.NewRequest(http.MethodPost, "/schedules", nil).WithContext(ctx)

	rr := httptest.NewRecorder()

	RequireRole(user.RoleDeployer, user.RoleAdmin)(testHandler).ServeHTTP(rr, req)

	if !handlerCalled {
		t.Error("handler should be called for deployer")
	}

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}
}
