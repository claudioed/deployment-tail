package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/claudioed/deployment-tail/internal/domain/user"
	"github.com/claudioed/deployment-tail/internal/infrastructure/jwt"
)

// contextKey is a custom type for context keys to avoid collisions
type contextKey string

const (
	// userContextKey is the key used to store the authenticated user in the request context
	userContextKey contextKey = "authenticated_user"
)

// JWTService defines the interface for JWT token operations
type JWTService interface {
	ValidateToken(tokenString string) (*jwt.Claims, error)
	HashToken(tokenString string) string
}

// RevocationStore defines the interface for token revocation checks
type RevocationStore interface {
	IsRevoked(tokenHash string) bool
}

// AuthenticationMiddleware validates JWT tokens and loads the authenticated user
type AuthenticationMiddleware struct {
	jwtService      JWTService
	revocationStore RevocationStore
	userRepo        user.Repository
}

// NewAuthenticationMiddleware creates a new authentication middleware
func NewAuthenticationMiddleware(
	jwtService JWTService,
	revocationStore RevocationStore,
	userRepo user.Repository,
) *AuthenticationMiddleware {
	return &AuthenticationMiddleware{
		jwtService:      jwtService,
		revocationStore: revocationStore,
		userRepo:        userRepo,
	}
}

// Authenticate is the middleware handler that validates JWT and loads user
func (m *AuthenticationMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract token from Authorization header
		tokenString, err := extractBearerToken(r)
		if err != nil {
			http.Error(w, `{"error": "missing or invalid authorization header"}`, http.StatusUnauthorized)
			return
		}

		// Validate token signature and claims
		claims, err := m.jwtService.ValidateToken(tokenString)
		if err != nil {
			http.Error(w, fmt.Sprintf(`{"error": "invalid token: %s"}`, err.Error()), http.StatusUnauthorized)
			return
		}

		// Check if token is revoked
		tokenHash := m.jwtService.HashToken(tokenString)
		if m.revocationStore.IsRevoked(tokenHash) {
			http.Error(w, `{"error": "token has been revoked"}`, http.StatusUnauthorized)
			return
		}

		// Parse user ID from claims
		userID, err := user.ParseUserID(claims.UserID)
		if err != nil {
			http.Error(w, `{"error": "invalid user ID in token"}`, http.StatusUnauthorized)
			return
		}

		// Load user from repository
		authenticatedUser, err := m.userRepo.FindByID(r.Context(), userID)
		if err != nil {
			// User might have been deleted
			http.Error(w, `{"error": "user not found"}`, http.StatusUnauthorized)
			return
		}

		// Add user to request context
		ctx := userToContext(r.Context(), authenticatedUser)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireRole creates a middleware that ensures the authenticated user has one of the specified roles
func RequireRole(roles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get authenticated user from context
			authenticatedUser, err := UserFromContext(r.Context())
			if err != nil {
				// Should not happen if Authenticate middleware ran first
				http.Error(w, `{"error": "authentication required"}`, http.StatusUnauthorized)
				return
			}

			// Check if user has one of the required roles
			userRole := authenticatedUser.Role().String()
			hasRole := false
			for _, requiredRole := range roles {
				if userRole == requiredRole {
					hasRole = true
					break
				}
			}

			if !hasRole {
				http.Error(w, fmt.Sprintf(`{"error": "requires one of roles: %v"}`, roles), http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// extractBearerToken extracts the JWT token from the Authorization header
// Expects format: "Authorization: Bearer <token>"
func extractBearerToken(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", fmt.Errorf("authorization header is missing")
	}

	// Split "Bearer <token>"
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return "", fmt.Errorf("authorization header format must be 'Bearer <token>'")
	}

	token := strings.TrimSpace(parts[1])
	if token == "" {
		return "", fmt.Errorf("token is empty")
	}

	return token, nil
}

// userToContext adds an authenticated user to the request context
func userToContext(ctx context.Context, u *user.User) context.Context {
	return context.WithValue(ctx, userContextKey, u)
}

// UserFromContext extracts the authenticated user from the request context
// This is exported so handlers can access the authenticated user
func UserFromContext(ctx context.Context) (*user.User, error) {
	u, ok := ctx.Value(userContextKey).(*user.User)
	if !ok || u == nil {
		return nil, fmt.Errorf("no authenticated user in context")
	}
	return u, nil
}

// UserToContext adds a user to the context - exported for testing purposes
// Tests should use this function instead of directly manipulating context
func UserToContext(ctx context.Context, u *user.User) context.Context {
	return context.WithValue(ctx, userContextKey, u)
}

// UserIDFromContext extracts the authenticated user ID from the request context
// Returns empty string and error if no user is authenticated
func UserIDFromContext(ctx context.Context) (string, error) {
	u, err := UserFromContext(ctx)
	if err != nil {
		return "", err
	}
	return u.ID().String(), nil
}

// GetUserIDFromContext extracts the authenticated user.UserID from the request context
// Returns the UserID and a boolean indicating success
func GetUserIDFromContext(ctx context.Context) (user.UserID, bool) {
	u, err := UserFromContext(ctx)
	if err != nil {
		return user.UserID{}, false
	}
	return u.ID(), true
}
