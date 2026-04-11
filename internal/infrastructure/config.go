package infrastructure

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds application configuration
type Config struct {
	Database DatabaseConfig
	Server   ServerConfig
	JWT      JWTConfig
	OAuth    OAuthConfig
}

// DatabaseConfig holds database connection settings
type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Database string
}

// ServerConfig holds HTTP server settings
type ServerConfig struct {
	Host string
	Port int
}

// JWTConfig holds JWT token settings
type JWTConfig struct {
	Secret string
	Expiry time.Duration
	Issuer string
}

// OAuthConfig holds OAuth configuration
type OAuthConfig struct {
	Google GoogleOAuthConfig
}

// GoogleOAuthConfig holds Google OAuth settings
type GoogleOAuthConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

// LoadConfig loads configuration from environment variables
func LoadConfig() (*Config, error) {
	dbPort, err := strconv.Atoi(getEnv("DB_PORT", "3306"))
	if err != nil {
		return nil, fmt.Errorf("invalid DB_PORT: %w", err)
	}

	serverPort, err := strconv.Atoi(getEnv("SERVER_PORT", "8080"))
	if err != nil {
		return nil, fmt.Errorf("invalid SERVER_PORT: %w", err)
	}

	// Parse JWT expiry duration (default 24 hours)
	jwtExpiryStr := getEnv("JWT_EXPIRY", "24h")
	jwtExpiry, err := time.ParseDuration(jwtExpiryStr)
	if err != nil {
		return nil, fmt.Errorf("invalid JWT_EXPIRY: %w", err)
	}

	return &Config{
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     dbPort,
			User:     getEnv("DB_USER", "root"),
			Password: getEnv("DB_PASSWORD", ""),
			Database: getEnv("DB_NAME", "deployment_schedules"),
		},
		Server: ServerConfig{
			Host: getEnv("SERVER_HOST", "0.0.0.0"),
			Port: serverPort,
		},
		JWT: JWTConfig{
			Secret: getEnv("JWT_SECRET", ""),
			Expiry: jwtExpiry,
			Issuer: getEnv("JWT_ISSUER", "deployment-tail"),
		},
		OAuth: OAuthConfig{
			Google: GoogleOAuthConfig{
				ClientID:     getEnv("GOOGLE_CLIENT_ID", ""),
				ClientSecret: getEnv("GOOGLE_CLIENT_SECRET", ""),
				RedirectURL:  getEnv("GOOGLE_REDIRECT_URL", ""),
			},
		},
	}, nil
}

// DSN returns the MySQL connection string
func (c *DatabaseConfig) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true&multiStatements=true",
		c.User, c.Password, c.Host, c.Port, c.Database)
}

// Address returns the server address
func (c *ServerConfig) Address() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// Validate validates the configuration
func (c *Config) Validate() error {
	// Validate JWT configuration
	if len(c.JWT.Secret) < 32 {
		return fmt.Errorf("JWT_SECRET must be at least 32 characters long (current: %d)", len(c.JWT.Secret))
	}

	if c.JWT.Expiry < time.Minute {
		return fmt.Errorf("JWT_EXPIRY must be at least 1 minute (current: %s)", c.JWT.Expiry)
	}

	if c.JWT.Issuer == "" {
		return fmt.Errorf("JWT_ISSUER cannot be empty")
	}

	// Validate Google OAuth configuration
	if c.OAuth.Google.ClientID == "" {
		return fmt.Errorf("GOOGLE_CLIENT_ID cannot be empty")
	}

	if c.OAuth.Google.ClientSecret == "" {
		return fmt.Errorf("GOOGLE_CLIENT_SECRET cannot be empty")
	}

	if c.OAuth.Google.RedirectURL == "" {
		return fmt.Errorf("GOOGLE_REDIRECT_URL cannot be empty")
	}

	// Validate database configuration
	if c.Database.Host == "" {
		return fmt.Errorf("DB_HOST cannot be empty")
	}

	if c.Database.User == "" {
		return fmt.Errorf("DB_USER cannot be empty")
	}

	if c.Database.Database == "" {
		return fmt.Errorf("DB_NAME cannot be empty")
	}

	return nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
