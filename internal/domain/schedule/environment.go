package schedule

import "fmt"

// Environment represents a deployment environment
type Environment string

const (
	EnvironmentProduction  Environment = "production"
	EnvironmentStaging     Environment = "staging"
	EnvironmentDevelopment Environment = "development"
)

// NewEnvironment creates a new environment
func NewEnvironment(env string) (Environment, error) {
	switch Environment(env) {
	case EnvironmentProduction, EnvironmentStaging, EnvironmentDevelopment:
		return Environment(env), nil
	default:
		return "", fmt.Errorf("invalid environment: %s (must be production, staging, or development)", env)
	}
}

// String returns the string representation
func (e Environment) String() string {
	return string(e)
}

// IsValid checks if the environment is valid
func (e Environment) IsValid() bool {
	switch e {
	case EnvironmentProduction, EnvironmentStaging, EnvironmentDevelopment:
		return true
	default:
		return false
	}
}
