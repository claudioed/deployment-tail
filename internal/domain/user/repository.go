package user

import "context"

// Repository defines the port for user persistence
type Repository interface {
	// Create persists a new user
	Create(ctx context.Context, user *User) error

	// FindByID retrieves a user by their ID
	FindByID(ctx context.Context, id UserID) (*User, error)

	// FindByGoogleID retrieves a user by their Google ID
	FindByGoogleID(ctx context.Context, googleID GoogleID) (*User, error)

	// FindByEmail retrieves a user by their email
	FindByEmail(ctx context.Context, email Email) (*User, error)

	// Update persists changes to an existing user
	Update(ctx context.Context, user *User) error

	// List retrieves users with optional filtering and pagination
	List(ctx context.Context, filters ListFilters) ([]*User, error)

	// UpdateRole updates a user's role
	UpdateRole(ctx context.Context, userID UserID, role Role) error

	// UpdateLastLogin updates the user's last login timestamp
	UpdateLastLogin(ctx context.Context, userID UserID) error
}

// ListFilters defines filtering options for listing users
type ListFilters struct {
	Role   *Role
	Limit  int
	Offset int
}
