package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/claudioed/deployment-tail/internal/domain/user"
	"github.com/go-sql-driver/mysql"
)

// UserRepository implements the user.Repository interface for MySQL
type UserRepository struct {
	db *sql.DB
}

// NewUserRepository creates a new MySQL user repository
func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create saves a new user
func (r *UserRepository) Create(ctx context.Context, u *user.User) error {
	query := `
		INSERT INTO users (id, google_id, email, name, role, last_login_at, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.ExecContext(ctx, query,
		u.ID().String(),
		u.GoogleID().String(),
		u.Email().String(),
		u.Name().String(),
		u.Role().String(),
		u.LastLoginAt(),
		u.CreatedAt(),
		u.UpdatedAt(),
	)

	if err != nil {
		// Check for duplicate key errors
		if isDuplicateKeyError(err) {
			return user.ErrUserAlreadyExists{
				GoogleID: u.GoogleID().String(),
				Email:    u.Email().String(),
			}
		}
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// FindByID retrieves a user by their ID
func (r *UserRepository) FindByID(ctx context.Context, id user.UserID) (*user.User, error) {
	query := `
		SELECT id, google_id, email, name, role, last_login_at, created_at, updated_at
		FROM users
		WHERE id = ?
	`

	var (
		idStr       string
		googleID    string
		email       string
		name        string
		role        string
		lastLoginAt sql.NullTime
		createdAt   time.Time
		updatedAt   time.Time
	)

	err := r.db.QueryRowContext(ctx, query, id.String()).Scan(
		&idStr,
		&googleID,
		&email,
		&name,
		&role,
		&lastLoginAt,
		&createdAt,
		&updatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, user.ErrUserNotFound{ID: id.String(), SearchType: "id"}
		}
		return nil, fmt.Errorf("failed to find user by ID: %w", err)
	}

	return r.mapToUser(idStr, googleID, email, name, role, lastLoginAt, createdAt, updatedAt)
}

// FindByGoogleID retrieves a user by their Google ID
func (r *UserRepository) FindByGoogleID(ctx context.Context, googleID user.GoogleID) (*user.User, error) {
	query := `
		SELECT id, google_id, email, name, role, last_login_at, created_at, updated_at
		FROM users
		WHERE google_id = ?
	`

	var (
		idStr       string
		gid         string
		email       string
		name        string
		role        string
		lastLoginAt sql.NullTime
		createdAt   time.Time
		updatedAt   time.Time
	)

	err := r.db.QueryRowContext(ctx, query, googleID.String()).Scan(
		&idStr,
		&gid,
		&email,
		&name,
		&role,
		&lastLoginAt,
		&createdAt,
		&updatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, user.ErrUserNotFound{GoogleID: googleID.String(), SearchType: "google_id"}
		}
		return nil, fmt.Errorf("failed to find user by Google ID: %w", err)
	}

	return r.mapToUser(idStr, gid, email, name, role, lastLoginAt, createdAt, updatedAt)
}

// FindByEmail retrieves a user by their email
func (r *UserRepository) FindByEmail(ctx context.Context, email user.Email) (*user.User, error) {
	query := `
		SELECT id, google_id, email, name, role, last_login_at, created_at, updated_at
		FROM users
		WHERE email = ?
	`

	var (
		idStr       string
		googleID    string
		emailStr    string
		name        string
		role        string
		lastLoginAt sql.NullTime
		createdAt   time.Time
		updatedAt   time.Time
	)

	err := r.db.QueryRowContext(ctx, query, email.String()).Scan(
		&idStr,
		&googleID,
		&emailStr,
		&name,
		&role,
		&lastLoginAt,
		&createdAt,
		&updatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, user.ErrUserNotFound{Email: email.String(), SearchType: "email"}
		}
		return nil, fmt.Errorf("failed to find user by email: %w", err)
	}

	return r.mapToUser(idStr, googleID, emailStr, name, role, lastLoginAt, createdAt, updatedAt)
}

// Update persists changes to an existing user
func (r *UserRepository) Update(ctx context.Context, u *user.User) error {
	query := `
		UPDATE users
		SET google_id = ?, email = ?, name = ?, role = ?, last_login_at = ?, updated_at = ?
		WHERE id = ?
	`

	result, err := r.db.ExecContext(ctx, query,
		u.GoogleID().String(),
		u.Email().String(),
		u.Name().String(),
		u.Role().String(),
		u.LastLoginAt(),
		u.UpdatedAt(),
		u.ID().String(),
	)

	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return user.ErrUserNotFound{ID: u.ID().String(), SearchType: "id"}
	}

	return nil
}

// List retrieves users with optional filtering and pagination
func (r *UserRepository) List(ctx context.Context, filters user.ListFilters) ([]*user.User, error) {
	query := `
		SELECT id, google_id, email, name, role, last_login_at, created_at, updated_at
		FROM users
		WHERE 1=1
	`
	args := []interface{}{}

	// Apply role filter
	if filters.Role != nil {
		query += " AND role = ?"
		args = append(args, filters.Role.String())
	}

	// Apply sorting
	query += " ORDER BY created_at DESC"

	// Apply pagination
	if filters.Limit > 0 {
		query += " LIMIT ?"
		args = append(args, filters.Limit)
	}
	if filters.Offset > 0 {
		query += " OFFSET ?"
		args = append(args, filters.Offset)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	defer rows.Close()

	var users []*user.User
	for rows.Next() {
		var (
			idStr       string
			googleID    string
			email       string
			name        string
			role        string
			lastLoginAt sql.NullTime
			createdAt   time.Time
			updatedAt   time.Time
		)

		err := rows.Scan(&idStr, &googleID, &email, &name, &role, &lastLoginAt, &createdAt, &updatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user row: %w", err)
		}

		u, err := r.mapToUser(idStr, googleID, email, name, role, lastLoginAt, createdAt, updatedAt)
		if err != nil {
			return nil, err
		}

		users = append(users, u)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating user rows: %w", err)
	}

	return users, nil
}

// UpdateRole updates a user's role
func (r *UserRepository) UpdateRole(ctx context.Context, userID user.UserID, role user.Role) error {
	query := `
		UPDATE users
		SET role = ?, updated_at = ?
		WHERE id = ?
	`

	result, err := r.db.ExecContext(ctx, query, role.String(), time.Now().UTC(), userID.String())
	if err != nil {
		return fmt.Errorf("failed to update user role: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return user.ErrUserNotFound{ID: userID.String(), SearchType: "id"}
	}

	return nil
}

// UpdateLastLogin updates the user's last login timestamp
func (r *UserRepository) UpdateLastLogin(ctx context.Context, userID user.UserID) error {
	query := `
		UPDATE users
		SET last_login_at = ?, updated_at = ?
		WHERE id = ?
	`

	now := time.Now().UTC()
	result, err := r.db.ExecContext(ctx, query, now, now, userID.String())
	if err != nil {
		return fmt.Errorf("failed to update last login: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return user.ErrUserNotFound{ID: userID.String(), SearchType: "id"}
	}

	return nil
}

// mapToUser converts database row data to a User domain object
func (r *UserRepository) mapToUser(
	idStr string,
	googleID string,
	email string,
	name string,
	role string,
	lastLoginAt sql.NullTime,
	createdAt time.Time,
	updatedAt time.Time,
) (*user.User, error) {
	id, err := user.ParseUserID(idStr)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID in database: %w", err)
	}

	gid, err := user.NewGoogleID(googleID)
	if err != nil {
		return nil, fmt.Errorf("invalid Google ID in database: %w", err)
	}

	e, err := user.NewEmail(email)
	if err != nil {
		return nil, fmt.Errorf("invalid email in database: %w", err)
	}

	n, err := user.NewUserName(name)
	if err != nil {
		return nil, fmt.Errorf("invalid user name in database: %w", err)
	}

	roleObj, err := user.NewRole(role)
	if err != nil {
		return nil, fmt.Errorf("invalid role in database: %w", err)
	}

	var lastLogin *time.Time
	if lastLoginAt.Valid {
		lastLogin = &lastLoginAt.Time
	}

	return user.Reconstitute(id, gid, e, n, roleObj, lastLogin, createdAt, updatedAt), nil
}

// isDuplicateKeyError checks if the error is a MySQL duplicate key error (1062).
// It uses errors.As to safely unwrap driver errors rather than inspecting the
// error message, which is brittle and can panic on short strings.
func isDuplicateKeyError(err error) bool {
	if err == nil {
		return false
	}
	var mysqlErr *mysql.MySQLError
	return errors.As(err, &mysqlErr) && mysqlErr.Number == 1062
}
