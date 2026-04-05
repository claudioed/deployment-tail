package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
)

// This script creates a bootstrap admin user for initial system setup
// Usage: go run scripts/seed_admin_user.go -email admin@example.com -name "Admin User" -google-id 123456789

func main() {
	// Parse command-line flags
	email := flag.String("email", "", "Admin user email (required)")
	name := flag.String("name", "", "Admin user name (required)")
	googleID := flag.String("google-id", "", "Admin user Google ID (required)")
	dbDSN := flag.String("db-dsn", os.Getenv("DB_DSN"), "Database DSN (default from DB_DSN env var)")
	flag.Parse()

	// Validate required flags
	if *email == "" || *name == "" || *googleID == "" {
		fmt.Println("Error: All flags are required")
		flag.Usage()
		os.Exit(1)
	}

	if *dbDSN == "" {
		// Build DSN from environment variables if not provided
		dbHost := getEnv("DB_HOST", "localhost")
		dbPort := getEnv("DB_PORT", "3306")
		dbUser := getEnv("DB_USER", "root")
		dbPassword := getEnv("DB_PASSWORD", "")
		dbName := getEnv("DB_NAME", "deployment_tail")
		*dbDSN = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true", dbUser, dbPassword, dbHost, dbPort, dbName)
	}

	// Connect to database
	db, err := sql.Open("mysql", *dbDSN)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Test connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	// Check if user already exists
	var existingID string
	err = db.QueryRow("SELECT id FROM users WHERE email = ? OR google_id = ?", *email, *googleID).Scan(&existingID)
	if err == nil {
		log.Printf("User already exists with ID: %s", existingID)
		log.Printf("To update role: UPDATE users SET role = 'admin' WHERE id = '%s';", existingID)
		return
	}
	if err != sql.ErrNoRows {
		log.Fatalf("Error checking for existing user: %v", err)
	}

	// Create admin user
	userID := uuid.New().String()
	ctx := context.Background()

	query := `
		INSERT INTO users (id, google_id, email, name, role, created_at, updated_at)
		VALUES (?, ?, ?, ?, 'admin', NOW(), NOW())
	`

	_, err = db.ExecContext(ctx, query, userID, *googleID, *email, *name)
	if err != nil {
		log.Fatalf("Failed to create admin user: %v", err)
	}

	log.Printf("✓ Successfully created admin user:")
	log.Printf("  ID: %s", userID)
	log.Printf("  Email: %s", *email)
	log.Printf("  Name: %s", *name)
	log.Printf("  Google ID: %s", *googleID)
	log.Printf("  Role: admin")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
