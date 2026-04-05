package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	httphandler "github.com/claudioed/deployment-tail/internal/adapters/input/http"
	"github.com/claudioed/deployment-tail/internal/adapters/input/http/middleware"
	"github.com/claudioed/deployment-tail/internal/adapters/output/mysql"
	"github.com/claudioed/deployment-tail/internal/application"
	"github.com/claudioed/deployment-tail/internal/infrastructure"
	"github.com/claudioed/deployment-tail/internal/infrastructure/jwt"
	"github.com/claudioed/deployment-tail/internal/infrastructure/oauth"
)

func main() {
	// Load configuration
	cfg, err := infrastructure.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		log.Fatalf("Configuration validation failed: %v", err)
	}

	// Initialize logger
	logger := infrastructure.NewLogger()
	logger.Info("Starting deployment-tail API server")

	// Connect to database
	db, err := infrastructure.NewDatabase(cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	logger.Info("Connected to database")

	// Run migrations
	if err := infrastructure.RunMigrations(db, "migrations"); err != nil {
		logger.Errorf("Failed to run migrations: %v", err)
		log.Fatalf("Migration error: %v", err)
	}

	logger.Info("Migrations completed successfully")

	// Initialize repositories
	scheduleRepo := mysql.NewScheduleRepository(db)
	groupRepo := mysql.NewGroupRepository(db)
	userRepo := mysql.NewUserRepository(db)

	// Initialize JWT service
	jwtService, err := jwt.NewJWTService(jwt.Config{
		Secret: cfg.JWT.Secret,
		Expiry: cfg.JWT.Expiry,
		Issuer: cfg.JWT.Issuer,
	})
	if err != nil {
		log.Fatalf("Failed to initialize JWT service: %v", err)
	}

	// Initialize token revocation store
	revocationStore := jwt.NewRevocationStore(db)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := revocationStore.LoadFromDatabase(ctx); err != nil {
		logger.Errorf("Failed to load revocation blacklist: %v", err)
	}

	// Start background sync and cleanup
	go func() {
		if err := revocationStore.Start(ctx); err != nil {
			logger.Errorf("Revocation store error: %v", err)
		}
	}()

	// Initialize Google OAuth client
	googleClient, err := oauth.NewGoogleClient(oauth.Config{
		ClientID:     cfg.OAuth.Google.ClientID,
		ClientSecret: cfg.OAuth.Google.ClientSecret,
		RedirectURL:  cfg.OAuth.Google.RedirectURL,
	})
	if err != nil {
		log.Fatalf("Failed to initialize Google OAuth client: %v", err)
	}

	// Initialize authentication middleware
	authMiddleware := middleware.NewAuthenticationMiddleware(jwtService, revocationStore, userRepo)

	// Initialize application services
	scheduleService := application.NewScheduleService(scheduleRepo, userRepo)
	groupService := application.NewGroupService(groupRepo, scheduleRepo)
	userService := application.NewUserService(userRepo, googleClient, jwtService, revocationStore)

	// Initialize auth handler
	authHandler := httphandler.NewAuthHandler(userService, googleClient)

	// Create HTTP server
	server := httphandler.NewServer(scheduleService, groupService, userService, authHandler, authMiddleware)

	// Create HTTP server with graceful shutdown
	addr := cfg.Server.Address()
	httpServer := &http.Server{
		Addr:    addr,
		Handler: server,
	}

	// Setup graceful shutdown
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	// Start server in goroutine
	go func() {
		logger.Infof("Server listening on %s", addr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Wait for shutdown signal
	<-shutdown
	logger.Info("Shutdown signal received, gracefully stopping server...")

	// Cancel background context to stop revocation store goroutine
	cancel()

	// Give server 30 seconds to shutdown gracefully
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		logger.Errorf("Server shutdown error: %v", err)
	}

	logger.Info("Server stopped")
}
