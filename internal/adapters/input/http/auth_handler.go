package http

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/claudioed/deployment-tail/api"
	"github.com/claudioed/deployment-tail/internal/adapters/input/http/middleware"
	"github.com/claudioed/deployment-tail/internal/application/ports/input"
)

// oauthStateCookieName is the name of the cookie used to store the OAuth state
// parameter for CSRF protection during the Google login flow.
const oauthStateCookieName = "oauth_state"

// oauthStateCookieMaxAge is the lifetime of the OAuth state cookie in seconds.
const oauthStateCookieMaxAge = 600 // 10 minutes

// generateOAuthState returns a cryptographically random, base64-URL-encoded
// state token suitable for use as the OAuth 2.0 state parameter.
func generateOAuthState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate oauth state: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// clearOAuthStateCookie instructs the browser to delete the OAuth state cookie.
func clearOAuthStateCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     oauthStateCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})
}

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
	// Generate a cryptographically random state parameter for CSRF protection
	state, err := generateOAuthState()
	if err != nil {
		log.Printf("Error generating oauth state: %v", err)
		h.writeErrorResponse(w, "failed to initiate login", http.StatusInternalServerError)
		return
	}

	// Store state in an HttpOnly cookie so it can be validated on callback.
	http.SetCookie(w, &http.Cookie{
		Name:     oauthStateCookieName,
		Value:    state,
		Path:     "/",
		MaxAge:   oauthStateCookieMaxAge,
		HttpOnly: true,
		Secure:   r.TLS != nil,
		SameSite: http.SameSiteLaxMode,
	})

	// Get Google OAuth URL
	authURL := h.googleClient.GetAuthURL(state)

	// Redirect to Google
	http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
}

// GoogleCallback handles the OAuth callback from Google
// GET /auth/google/callback
func (h *AuthHandler) GoogleCallback(w http.ResponseWriter, r *http.Request, params api.GoogleCallbackParams) {
	// Validate OAuth state parameter (CSRF protection). The state cookie is
	// always cleared after being read, regardless of the outcome, so that a
	// single state value cannot be reused.
	stateCookie, cookieErr := r.Cookie(oauthStateCookieName)
	clearOAuthStateCookie(w)

	// Get authorization code from params
	code := params.Code
	if code == "" {
		log.Printf("Error: missing authorization code in callback")
		h.writeErrorResponse(w, "missing authorization code", http.StatusBadRequest)
		return
	}

	// Get state parameter and compare with the value stored in the cookie.
	state := params.State
	if state == "" {
		log.Printf("Error: missing state parameter in callback")
		h.writeErrorResponse(w, "missing state parameter", http.StatusBadRequest)
		return
	}

	if cookieErr != nil || stateCookie == nil || stateCookie.Value == "" {
		log.Printf("Error: missing oauth state cookie: %v", cookieErr)
		h.writeErrorResponse(w, "invalid state parameter", http.StatusBadRequest)
		return
	}

	if subtle.ConstantTimeCompare([]byte(stateCookie.Value), []byte(state)) != 1 {
		log.Printf("Error: oauth state mismatch")
		h.writeErrorResponse(w, "invalid state parameter", http.StatusBadRequest)
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

	// Return HTML that stores token in localStorage and redirects
	// URL fragments (#hash) are lost in HTTP redirects, so we return HTML directly
	html := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <title>Login Successful</title>
    <style>
        body { font-family: sans-serif; display: flex; justify-content: center; align-items: center; height: 100vh; margin: 0; background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%); color: white; }
        .container { text-align: center; }
    </style>
</head>
<body>
    <div class="container">
        <h2>✓ Login Successful</h2>
        <p>Redirecting...</p>
    </div>
    <script>
        localStorage.setItem('auth_token', %q);
        localStorage.setItem('user_email', %q);
        localStorage.setItem('user_name', %q);
        localStorage.setItem('user_role', %q);
        window.location.href = '/';
    </script>
</body>
</html>`,
		result.Token,
		result.User.Email().String(),
		result.User.Name().String(),
		result.User.Role().String(),
	)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(html))
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
