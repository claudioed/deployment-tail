package http

import (
	"net/http"

	"github.com/claudioed/deployment-tail/api"
	authmiddleware "github.com/claudioed/deployment-tail/internal/adapters/input/http/middleware"
	"github.com/claudioed/deployment-tail/internal/application"
	"github.com/claudioed/deployment-tail/internal/application/ports/input"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

// Server represents the HTTP server
type Server struct {
	scheduleService input.ScheduleService
	groupService    input.GroupService
	userService     input.UserService
	serviceService  *application.ServiceService
	authHandler     *AuthHandler
	userHandler     *UserHandler
	authMiddleware  *authmiddleware.AuthenticationMiddleware
	router          *chi.Mux
}

// NewServer creates a new HTTP server
func NewServer(
	scheduleService input.ScheduleService,
	groupService input.GroupService,
	userService input.UserService,
	serviceService *application.ServiceService,
	authHandler *AuthHandler,
	authMiddleware *authmiddleware.AuthenticationMiddleware,
) *Server {
	s := &Server{
		scheduleService: scheduleService,
		groupService:    groupService,
		userService:     userService,
		serviceService:  serviceService,
		authHandler:     authHandler,
		userHandler:     NewUserHandler(userService),
		authMiddleware:  authMiddleware,
		router:          chi.NewRouter(),
	}

	s.setupMiddleware()
	s.setupRoutes()

	return s
}

// setupMiddleware configures middleware
func (s *Server) setupMiddleware() {
	s.router.Use(chimiddleware.Logger)
	s.router.Use(chimiddleware.Recoverer)
	s.router.Use(chimiddleware.RequestID)
	s.router.Use(chimiddleware.RealIP)
	s.router.Use(corsMiddleware)
	s.router.Use(ValidationMiddleware)

	// Apply authentication to protected routes
	conditionalAuth := authmiddleware.NewConditionalAuthMiddleware(s.authMiddleware)
	s.router.Use(conditionalAuth.Handle)
}

// corsMiddleware adds CORS headers
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// setupRoutes configures the API routes
func (s *Server) setupRoutes() {
	// Health check (public)
	s.router.Get("/health", s.healthCheck)

	// Combined handler implements all API endpoints
	combinedHandler := &CombinedHandler{
		ScheduleHandler: NewScheduleHandler(s.scheduleService, s.groupService, s.userService),
		GroupHandler:    NewGroupHandler(s.groupService, s.scheduleService),
		UserHandler:     s.userHandler,
		AuthHandler:     s.authHandler,
		ServiceHandler:  NewServiceHandler(s.serviceService),
	}

	// Mount auth routes at root level (public paths like /auth/google/login)
	// Mount API routes at root level (will be prefixed in OpenAPI if needed)
	// The OpenAPI spec defines which routes require authentication
	api.HandlerFromMux(combinedHandler, s.router)

	// Serve static files from web directory (public)
	fileServer := http.FileServer(http.Dir("./web"))
	s.router.Get("/*", func(w http.ResponseWriter, r *http.Request) {
		fileServer.ServeHTTP(w, r)
	})
}

// ServeHTTP implements http.Handler
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

// healthCheck endpoint
func (s *Server) healthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"healthy"}`))
}
