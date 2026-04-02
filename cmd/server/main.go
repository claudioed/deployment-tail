package main

import (
	"log"
	"net/http"

	httphandler "github.com/claudioed/deployment-tail/internal/adapters/input/http"
	"github.com/claudioed/deployment-tail/internal/adapters/output/mysql"
	"github.com/claudioed/deployment-tail/internal/application"
	"github.com/claudioed/deployment-tail/internal/infrastructure"
)

func main() {
	// Load configuration
	cfg, err := infrastructure.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
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

	// Initialize repository
	repo := mysql.NewScheduleRepository(db)

	// Initialize application service
	service := application.NewScheduleService(repo)

	// Create HTTP server
	server := httphandler.NewServer(service)

	// Start server
	addr := cfg.Server.Address()
	logger.Infof("Server listening on %s", addr)
	if err := http.ListenAndServe(addr, server); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
