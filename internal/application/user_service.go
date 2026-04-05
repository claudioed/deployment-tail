package application

import (
	"context"
	"fmt"

	"github.com/claudioed/deployment-tail/internal/application/ports/input"
	"github.com/claudioed/deployment-tail/internal/domain/user"
	"github.com/claudioed/deployment-tail/internal/infrastructure/jwt"
	"github.com/claudioed/deployment-tail/internal/infrastructure/oauth"
)

// UserService implements user management use cases
type UserService struct {
	userRepo        user.Repository
	googleClient    *oauth.GoogleClient
	jwtService      *jwt.JWTService
	revocationStore *jwt.RevocationStore
}

// NewUserService creates a new user service
func NewUserService(
	userRepo user.Repository,
	googleClient *oauth.GoogleClient,
	jwtService *jwt.JWTService,
	revocationStore *jwt.RevocationStore,
) *UserService {
	return &UserService{
		userRepo:        userRepo,
		googleClient:    googleClient,
		jwtService:      jwtService,
		revocationStore: revocationStore,
	}
}

// AuthenticateWithGoogle authenticates a user via Google OAuth code
func (s *UserService) AuthenticateWithGoogle(ctx context.Context, code string) (*input.AuthenticationResult, error) {
	// Exchange code for access token
	token, err := s.googleClient.ExchangeCode(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code: %w", err)
	}

	// Get user info from Google (OIDC-compliant: extracts claims from ID token)
	googleUserInfo, err := s.googleClient.GetUserInfo(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}

	// Register or update user
	u, err := s.RegisterOrUpdateUser(ctx, googleUserInfo.ID, googleUserInfo.Email, googleUserInfo.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to register/update user: %w", err)
	}

	// Update last login
	if err := s.userRepo.UpdateLastLogin(ctx, u.ID()); err != nil {
		// Log error but don't fail authentication
		fmt.Printf("Warning: failed to update last login: %v\n", err)
	}

	// Generate JWT token
	jwtToken, err := s.jwtService.GenerateToken(u)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &input.AuthenticationResult{
		User:  u,
		Token: jwtToken,
	}, nil
}

// RegisterOrUpdateUser creates or updates a user from Google profile
func (s *UserService) RegisterOrUpdateUser(ctx context.Context, googleID, email, name string) (*user.User, error) {
	// Parse value objects
	gid, err := user.NewGoogleID(googleID)
	if err != nil {
		return nil, fmt.Errorf("invalid Google ID: %w", err)
	}

	e, err := user.NewEmail(email)
	if err != nil {
		return nil, fmt.Errorf("invalid email: %w", err)
	}

	n, err := user.NewUserName(name)
	if err != nil {
		return nil, fmt.Errorf("invalid user name: %w", err)
	}

	// Check if user already exists
	existingUser, err := s.userRepo.FindByGoogleID(ctx, gid)
	if err == nil {
		// User exists, update profile if changed
		if !existingUser.Email().Equals(e) || !existingUser.Name().Equals(n) {
			if err := existingUser.UpdateProfile(n, e); err != nil {
				return nil, fmt.Errorf("failed to update profile: %w", err)
			}
			if err := s.userRepo.Update(ctx, existingUser); err != nil {
				return nil, fmt.Errorf("failed to save updated user: %w", err)
			}
		}
		return existingUser, nil
	}

	// Check if it's a "not found" error
	if _, ok := err.(user.ErrUserNotFound); !ok {
		return nil, fmt.Errorf("failed to check existing user: %w", err)
	}

	// Create new user with viewer role by default
	role, _ := user.NewRole(user.RoleViewer)
	newUser, err := user.NewUser(gid, e, n, role)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	if err := s.userRepo.Create(ctx, newUser); err != nil {
		return nil, fmt.Errorf("failed to save new user: %w", err)
	}

	return newUser, nil
}

// GetUserProfile retrieves a user's profile by ID
func (s *UserService) GetUserProfile(ctx context.Context, userID user.UserID) (*user.User, error) {
	u, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}
	return u, nil
}

// ListUsers retrieves users with optional filtering and pagination
func (s *UserService) ListUsers(ctx context.Context, filters input.UserListFilters) ([]*user.User, error) {
	repoFilters := user.ListFilters{
		Role:   filters.Role,
		Limit:  filters.Limit,
		Offset: filters.Offset,
	}

	users, err := s.userRepo.List(ctx, repoFilters)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	return users, nil
}

// AssignRole assigns a role to a user (admin only)
func (s *UserService) AssignRole(ctx context.Context, adminUserID, targetUserID user.UserID, newRole user.Role) error {
	// Verify admin has permission
	admin, err := s.userRepo.FindByID(ctx, adminUserID)
	if err != nil {
		return fmt.Errorf("failed to find admin user: %w", err)
	}

	if !admin.CanManageUsers() {
		return user.ErrUnauthorized{
			UserID:    adminUserID.String(),
			Operation: "assign role",
			Reason:    "requires admin role",
		}
	}

	// Get target user
	targetUser, err := s.userRepo.FindByID(ctx, targetUserID)
	if err != nil {
		return fmt.Errorf("failed to find target user: %w", err)
	}

	// Update role
	if err := targetUser.UpdateRole(newRole); err != nil {
		return fmt.Errorf("failed to update role: %w", err)
	}

	// Save changes
	if err := s.userRepo.Update(ctx, targetUser); err != nil {
		return fmt.Errorf("failed to save role change: %w", err)
	}

	return nil
}

// RefreshUserToken generates a new JWT token for a user
func (s *UserService) RefreshUserToken(ctx context.Context, userID user.UserID) (string, error) {
	// Get current user data
	u, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return "", fmt.Errorf("failed to find user: %w", err)
	}

	// Generate new token
	token, err := s.jwtService.GenerateToken(u)
	if err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}

	return token, nil
}

// RevokeUserToken revokes a specific token
func (s *UserService) RevokeUserToken(ctx context.Context, tokenHash string, userID user.UserID) error {
	// Parse claims to get expiry
	claims, err := s.jwtService.ParseClaims(tokenHash)
	if err != nil {
		// If we can't parse it, we can't revoke it properly
		// But we can still try to add it to blacklist with a default expiry
		expiresAt := claims.ExpiresAt.Time
		if claims.ExpiresAt == nil {
			expiresAt = claims.ExpiresAt.Time
		}
		return s.revocationStore.AddToBlacklist(ctx, s.jwtService.HashToken(tokenHash), userID.String(), expiresAt)
	}

	// Add to revocation blacklist
	expiresAt := claims.ExpiresAt.Time
	tokenHashHex := s.jwtService.HashToken(tokenHash)

	if err := s.revocationStore.AddToBlacklist(ctx, tokenHashHex, userID.String(), expiresAt); err != nil {
		return fmt.Errorf("failed to revoke token: %w", err)
	}

	return nil
}
