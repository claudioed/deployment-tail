package input

import (
	"context"

	"github.com/claudioed/deployment-tail/internal/domain/user"
)

// UserService defines the application use cases for user management
type UserService interface {
	// AuthenticateWithGoogle authenticates a user via Google OAuth code
	AuthenticateWithGoogle(ctx context.Context, code string) (*AuthenticationResult, error)

	// RegisterOrUpdateUser creates or updates a user from Google profile
	RegisterOrUpdateUser(ctx context.Context, googleID, email, name string) (*user.User, error)

	// GetUserProfile retrieves a user's profile by ID
	GetUserProfile(ctx context.Context, userID user.UserID) (*user.User, error)

	// ListUsers retrieves users with optional filtering and pagination
	ListUsers(ctx context.Context, filters UserListFilters) ([]*user.User, error)

	// AssignRole assigns a role to a user (admin only)
	AssignRole(ctx context.Context, adminUserID, targetUserID user.UserID, newRole user.Role) error

	// RefreshUserToken generates a new JWT token for a user
	RefreshUserToken(ctx context.Context, userID user.UserID) (string, error)

	// RevokeUserToken revokes a specific token
	RevokeUserToken(ctx context.Context, tokenHash string, userID user.UserID) error
}

// AuthenticationResult contains the result of a successful authentication
type AuthenticationResult struct {
	User  *user.User
	Token string
}

// UserListFilters defines filtering options for listing users
type UserListFilters struct {
	Role   *user.Role
	Limit  int
	Offset int
}
