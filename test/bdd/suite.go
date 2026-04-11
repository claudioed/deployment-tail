package bdd

import (
	"context"
	"net/http/httptest"
	"time"

	httphandler "github.com/claudioed/deployment-tail/internal/adapters/input/http"
	"github.com/claudioed/deployment-tail/internal/adapters/input/http/middleware"
	"github.com/claudioed/deployment-tail/internal/application"
	"github.com/claudioed/deployment-tail/internal/application/applicationtest"
	"github.com/claudioed/deployment-tail/internal/application/ports/input"
	"github.com/claudioed/deployment-tail/internal/domain/group"
	"github.com/claudioed/deployment-tail/internal/domain/schedule"
	"github.com/claudioed/deployment-tail/internal/domain/user"
	"github.com/claudioed/deployment-tail/internal/infrastructure/jwt"
)

type worldKey struct{}

// World holds all state for a single scenario
type World struct {
	// Repositories (mocks from applicationtest)
	ScheduleRepo *applicationtest.MockRepository
	GroupRepo    *applicationtest.MockGroupRepository
	UserRepo     *applicationtest.MockUserRepository

	// Services under test
	ScheduleService input.ScheduleService
	GroupService    input.GroupService
	UserService     input.UserService

	// HTTP wiring (populated only for @http / @ui)
	HTTPServer *httptest.Server
	JWTService *jwt.JWTService
	RevocationStore *jwt.RevocationStore

	// chromedp wiring (@ui)
	BrowserCtx    context.Context
	BrowserCancel context.CancelFunc

	// Scenario state
	CurrentUser      *user.User
	CurrentToken     string
	LastSchedule     *schedule.Schedule
	LastGroup        *group.Group
	LastError        error
	LastStatusCode   int
	LastResponseBody []byte

	// Named entities for Given steps
	NamedUsers     map[string]*user.User
	NamedSchedules map[string]*schedule.Schedule
	NamedGroups    map[string]*group.Group

	// Phase B: Lists for assertions
	lastGroupList    []*group.Group
	lastScheduleList []*schedule.Schedule
	bulkSchedules    []*schedule.Schedule

	// Phase C: Auth and JWT state
	jwtClaims        map[string]interface{}
	authHeader       string
	jwtExpired       bool
	jwtTampered      bool
	revokedTokens    map[string]bool
	oldJWTExpiry     time.Time
	newJWTToken      string
	newJWTExpiry     time.Time
	jwtExpiryHours   int
	cleanupRan       bool
	locationHeader   string
	oauthState       string
	googleUserEmail  string
	googleUserName   string
	googleAPIError   bool
	googleMissingEmail bool

	// Phase C: User management state
	userProfile      *user.User
	userList         []*user.User
	userOldName      string
	userCountBefore  int
}

// NewWorld creates a fresh World with wired services
func NewWorld() *World {
	// Create mocks with shared state for ungrouped filtering
	groupRepo := applicationtest.NewMockGroupRepository()
	scheduleRepo := applicationtest.NewMockRepository()
	// Link the repositories so scheduleRepo can filter ungrouped schedules
	scheduleRepo.SetGroupRepository(groupRepo)
	userRepo := applicationtest.NewMockUserRepository()

	// Wire services
	scheduleService := application.NewScheduleService(scheduleRepo, userRepo)
	groupService := application.NewGroupService(groupRepo, scheduleRepo)

	// UserService needs OAuth/JWT wiring, which we'll provide during HTTP server setup
	// For service-level tests, we don't need UserService
	var userService input.UserService

	return &World{
		ScheduleRepo:    scheduleRepo,
		GroupRepo:       groupRepo,
		UserRepo:        userRepo,
		ScheduleService: scheduleService,
		GroupService:    groupService,
		UserService:     userService,
		NamedUsers:      make(map[string]*user.User),
		NamedSchedules:  make(map[string]*schedule.Schedule),
		NamedGroups:     make(map[string]*group.Group),
		revokedTokens:   make(map[string]bool),
		jwtExpiryHours:  24, // Default 24 hours
	}
}

// Reset clears all state for the next scenario
func (w *World) Reset() {
	if w.HTTPServer != nil {
		w.HTTPServer.Close()
	}
	if w.BrowserCancel != nil {
		w.BrowserCancel()
	}

	// Rebuild everything
	newWorld := NewWorld()
	*w = *newWorld
}

// startHTTPServer wires the real HTTP stack (middleware + router)
// Copied from route_protection_integration_test.go:23-52
func (w *World) startHTTPServer() {
	// JWT setup
	jwtService, err := jwt.NewJWTService(jwt.Config{
		Secret: "test-secret-key-for-bdd-testing-minimum-32-chars",
		Expiry: 24 * time.Hour,
		Issuer: "deployment-tail-bdd",
	})
	if err != nil {
		panic("failed to create JWT service: " + err.Error())
	}

	w.JWTService = jwtService
	w.RevocationStore = jwt.NewRevocationStore(nil) // nil for in-memory only

	// Mock Google OAuth client for auth handler
	mockGoogleClient := &mockGoogleClient{}

	// Wire UserService now that we have OAuth/JWT dependencies
	// Pass nil for googleClient since we don't test OAuth flows in BDD pilot
	w.UserService = application.NewUserService(w.UserRepo, nil, w.JWTService, w.RevocationStore)

	// Auth middleware
	authMiddleware := middleware.NewAuthenticationMiddleware(w.JWTService, w.RevocationStore, w.UserRepo)

	// Auth handler (needs the mock for GetAuthURL)
	authHandler := httphandler.NewAuthHandler(w.UserService, mockGoogleClient)

	// Create HTTP server
	handler := httphandler.NewServer(
		w.ScheduleService,
		w.GroupService,
		w.UserService,
		nil, // ServiceService - not needed for pilot
		authHandler,
		authMiddleware,
	)

	w.HTTPServer = httptest.NewServer(handler)
}

// mockGoogleClient is a minimal mock for OAuth testing
// Implements httphandler.GoogleOAuthClient interface
type mockGoogleClient struct{}

func (m *mockGoogleClient) GetAuthURL(state string) string {
	return "https://accounts.google.com/o/oauth2/auth?mock=true"
}

// startBrowser initializes chromedp context for UI tests
func (w *World) startBrowser() {
	w.BrowserCtx, w.BrowserCancel = context.WithCancel(context.Background())
	// Full chromedp setup deferred to Phase E when we implement ui_steps.go
}

// withWorld stores World in context
func withWorld(ctx context.Context, w *World) context.Context {
	return context.WithValue(ctx, worldKey{}, w)
}

// getWorld retrieves World from context
func getWorld(ctx context.Context) *World {
	return ctx.Value(worldKey{}).(*World)
}
