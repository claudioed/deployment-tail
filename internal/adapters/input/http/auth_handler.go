package http

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/claudioed/deployment-tail/api"
	"github.com/claudioed/deployment-tail/internal/adapters/input/http/middleware"
	"github.com/claudioed/deployment-tail/internal/application/ports/input"
)

// GoogleOAuthClient defines the interface for Google OAuth operations
type GoogleOAuthClient interface {
	GetAuthURL(state string) string
}

// AuthHandler handles authentication-related HTTP requests
type AuthHandler struct {
	userService  input.UserService
	googleClient GoogleOAuthClient
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(userService input.UserService, googleClient GoogleOAuthClient) *AuthHandler {
	return &AuthHandler{
		userService:  userService,
		googleClient: googleClient,
	}
}

// GoogleLogin redirects the user to Google OAuth login page
// GET /auth/google/login
func (h *AuthHandler) GoogleLogin(w http.ResponseWriter, r *http.Request) {
	// Generate a random state parameter for CSRF protection
	// In production, this should be stored in session/cookie and validated in callback
	state := "random-state-" + fmt.Sprintf("%d", r.Context().Value("request-id"))

	// Get Google OAuth URL
	authURL := h.googleClient.GetAuthURL(state)

	// Redirect to Google
	http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
}

// GoogleCallback handles the OAuth callback from Google
// GET /auth/google/callback
func (h *AuthHandler) GoogleCallback(w http.ResponseWriter, r *http.Request, params api.GoogleCallbackParams) {
	// Get authorization code from params
	code := params.Code
	if code == "" {
		log.Printf("Error: missing authorization code in callback")
		h.writeErrorResponse(w, "missing authorization code", http.StatusBadRequest)
		return
	}

	// Get state parameter (should validate this matches the one we sent)
	state := params.State
	if state == "" {
		log.Printf("Error: missing state parameter in callback")
		h.writeErrorResponse(w, "missing state parameter", http.StatusBadRequest)
		return
	}

	// Check for error from Google
	if errParam := r.URL.Query().Get("error"); errParam != "" {
		log.Printf("Error from Google OAuth: %s", errParam)
		h.writeErrorResponse(w, fmt.Sprintf("authentication failed: %s", errParam), http.StatusBadRequest)
		return
	}

	// Authenticate with Google (exchanges code, registers/updates user, generates JWT)
	result, err := h.userService.AuthenticateWithGoogle(r.Context(), code)
	if err != nil {
		log.Printf("Error authenticating with Google: %v", err)
		h.writeErrorResponse(w, "authentication failed", http.StatusInternalServerError)
		return
	}

	// Return JWT token and user info
	response := map[string]interface{}{
		"token": result.Token,
		"user": map[string]interface{}{
			"id":    result.User.ID().String(),
			"email": result.User.Email().String(),
			"name":  result.User.Name().String(),
			"role":  result.User.Role().String(),
		},
	}

	h.writeJSONResponse(w, response, http.StatusOK)
}

// RefreshToken generates a new JWT token for the authenticated user
// POST /auth/refresh
func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	// Get authenticated user from context (set by AuthenticationMiddleware)
	user, err := middleware.UserFromContext(r.Context())
	if err != nil {
		log.Printf("Error: no authenticated user in context")
		h.writeErrorResponse(w, "authentication required", http.StatusUnauthorized)
		return
	}

	// Generate new token
	newToken, err := h.userService.RefreshUserToken(r.Context(), user.ID())
	if err != nil {
		log.Printf("Error refreshing token for user %s: %v", user.ID().String(), err)
		h.writeErrorResponse(w, "failed to refresh token", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"token": newToken,
		"user": map[string]interface{}{
			"id":    user.ID().String(),
			"email": user.Email().String(),
			"name":  user.Name().String(),
			"role":  user.Role().String(),
		},
	}

	h.writeJSONResponse(w, response, http.StatusOK)
}

// Logout revokes the current JWT token
// POST /auth/logout
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	// Get authenticated user from context
	user, err := middleware.UserFromContext(r.Context())
	if err != nil {
		log.Printf("Error: no authenticated user in context")
		h.writeErrorResponse(w, "authentication required", http.StatusUnauthorized)
		return
	}

	// Extract token from Authorization header
	tokenString := r.Header.Get("Authorization")
	if tokenString == "" {
		log.Printf("Error: missing authorization header in logout")
		h.writeErrorResponse(w, "missing authorization header", http.StatusBadRequest)
		return
	}

	// Remove "Bearer " prefix
	if len(tokenString) > 7 && tokenString[:7] == "Bearer " {
		tokenString = tokenString[7:]
	}

	// Revoke token
	err = h.userService.RevokeUserToken(r.Context(), tokenString, user.ID())
	if err != nil {
		log.Printf("Error revoking token for user %s: %v", user.ID().String(), err)
		h.writeErrorResponse(w, "failed to revoke token", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"message": "logged out successfully",
	}

	h.writeJSONResponse(w, response, http.StatusOK)
}

// Helper methods for consistent response formatting

func (h *AuthHandler) writeJSONResponse(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Error encoding JSON response: %v", err)
	}
}

func (h *AuthHandler) writeErrorResponse(w http.ResponseWriter, message string, statusCode int) {
	response := map[string]interface{}{
		"error": message,
	}
	h.writeJSONResponse(w, response, statusCode)
}
