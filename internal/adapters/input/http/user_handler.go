package http

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/claudioed/deployment-tail/api"
	"github.com/claudioed/deployment-tail/internal/adapters/input/http/middleware"
	"github.com/claudioed/deployment-tail/internal/application/ports/input"
	"github.com/claudioed/deployment-tail/internal/domain/user"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

// UserHandler handles user management HTTP requests
type UserHandler struct {
	userService input.UserService
}

// NewUserHandler creates a new user handler
func NewUserHandler(userService input.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// GetMyProfile retrieves the authenticated user's own profile
// GET /users/me
func (h *UserHandler) GetMyProfile(w http.ResponseWriter, r *http.Request) {
	// Get authenticated user from context
	authenticatedUser, err := middleware.UserFromContext(r.Context())
	if err != nil {
		log.Printf("Error: no authenticated user in context")
		h.writeErrorResponse(w, "authentication required", http.StatusUnauthorized)
		return
	}

	// Return user profile
	response := h.userToResponse(authenticatedUser)
	h.writeJSONResponse(w, response, http.StatusOK)
}

// GetUserByID retrieves a specific user's profile (admin only)
// GET /users/{id}
func (h *UserHandler) GetUserByID(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	// Defense-in-depth: enforce admin-only access inside the handler, in
	// addition to the route-level RequireRole("admin") middleware.
	authenticatedUser, err := middleware.UserFromContext(r.Context())
	if err != nil {
		log.Printf("Error: no authenticated user in context")
		h.writeErrorResponse(w, "authentication required", http.StatusUnauthorized)
		return
	}
	if !authenticatedUser.CanManageUsers() {
		log.Printf("Error: user %s attempted to access GetUserByID without admin role", authenticatedUser.ID().String())
		h.writeErrorResponse(w, "insufficient permissions", http.StatusForbidden)
		return
	}

	// Parse user ID
	targetUserID, err := user.ParseUserID(id.String())
	if err != nil {
		log.Printf("Error parsing user ID: %v", err)
		h.writeErrorResponse(w, "invalid user ID", http.StatusBadRequest)
		return
	}

	// Get user profile
	targetUser, err := h.userService.GetUserProfile(r.Context(), targetUserID)
	if err != nil {
		if _, ok := err.(user.ErrUserNotFound); ok {
			h.writeErrorResponse(w, "user not found", http.StatusNotFound)
			return
		}
		log.Printf("Error getting user profile: %v", err)
		h.writeErrorResponse(w, "failed to get user profile", http.StatusInternalServerError)
		return
	}

	response := h.userToResponse(targetUser)
	h.writeJSONResponse(w, response, http.StatusOK)
}

// ListUsers retrieves all users with optional filtering (admin only)
// GET /users
func (h *UserHandler) ListUsers(w http.ResponseWriter, r *http.Request, params api.ListUsersParams) {
	// Defense-in-depth: enforce admin-only access inside the handler, in
	// addition to the route-level RequireRole("admin") middleware. If the
	// middleware is ever misconfigured, this prevents user enumeration.
	authenticatedUser, err := middleware.UserFromContext(r.Context())
	if err != nil {
		log.Printf("Error: no authenticated user in context")
		h.writeErrorResponse(w, "authentication required", http.StatusUnauthorized)
		return
	}
	if !authenticatedUser.CanManageUsers() {
		log.Printf("Error: user %s attempted to list users without admin role", authenticatedUser.ID().String())
		h.writeErrorResponse(w, "insufficient permissions", http.StatusForbidden)
		return
	}

	filters := input.UserListFilters{}

	// Parse role filter from params
	if params.Role != nil {
		roleStr := string(*params.Role)
		role, err := user.NewRole(roleStr)
		if err != nil {
			log.Printf("Error parsing role filter: %v", err)
			h.writeErrorResponse(w, fmt.Sprintf("invalid role: %s", roleStr), http.StatusBadRequest)
			return
		}
		filters.Role = &role
	}

	// Get users
	users, err := h.userService.ListUsers(r.Context(), filters)
	if err != nil {
		log.Printf("Error listing users: %v", err)
		h.writeErrorResponse(w, "failed to list users", http.StatusInternalServerError)
		return
	}

	// Convert to response format
	response := make([]map[string]interface{}, len(users))
	for i, u := range users {
		response[i] = h.userToResponse(u)
	}

	h.writeJSONResponse(w, response, http.StatusOK)
}

// AssignRole assigns a role to a user (admin only)
// PUT /users/{id}/role
func (h *UserHandler) AssignRole(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	// Get authenticated user (admin) from context
	authenticatedUser, err := middleware.UserFromContext(r.Context())
	if err != nil {
		log.Printf("Error: no authenticated user in context")
		h.writeErrorResponse(w, "authentication required", http.StatusUnauthorized)
		return
	}

	// Parse target user ID
	targetUserID, err := user.ParseUserID(id.String())
	if err != nil {
		log.Printf("Error parsing user ID: %v", err)
		h.writeErrorResponse(w, "invalid user ID", http.StatusBadRequest)
		return
	}

	// Parse request body
	var req struct {
		Role string `json:"role"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Error decoding request body: %v", err)
		h.writeErrorResponse(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Validate and parse role
	newRole, err := user.NewRole(req.Role)
	if err != nil {
		log.Printf("Error parsing role: %v", err)
		h.writeErrorResponse(w, fmt.Sprintf("invalid role: %s", req.Role), http.StatusBadRequest)
		return
	}

	// Assign role
	err = h.userService.AssignRole(r.Context(), authenticatedUser.ID(), targetUserID, newRole)
	if err != nil {
		// Check for authorization error
		if _, ok := err.(user.ErrUnauthorized); ok {
			h.writeErrorResponse(w, "insufficient permissions", http.StatusForbidden)
			return
		}

		// Check for user not found
		if _, ok := err.(user.ErrUserNotFound); ok {
			h.writeErrorResponse(w, "user not found", http.StatusNotFound)
			return
		}

		log.Printf("Error assigning role: %v", err)
		h.writeErrorResponse(w, "failed to assign role", http.StatusInternalServerError)
		return
	}

	// Get updated user
	updatedUser, err := h.userService.GetUserProfile(r.Context(), targetUserID)
	if err != nil {
		log.Printf("Error getting updated user: %v", err)
		h.writeErrorResponse(w, "role assigned but failed to retrieve updated user", http.StatusInternalServerError)
		return
	}

	response := h.userToResponse(updatedUser)
	h.writeJSONResponse(w, response, http.StatusOK)
}

// Helper methods

func (h *UserHandler) userToResponse(u *user.User) map[string]interface{} {
	response := map[string]interface{}{
		"id":         u.ID().String(),
		"email":      u.Email().String(),
		"name":       u.Name().String(),
		"role":       u.Role().String(),
		"created_at": u.CreatedAt().Format("2006-01-02T15:04:05Z07:00"),
		"updated_at": u.UpdatedAt().Format("2006-01-02T15:04:05Z07:00"),
	}

	if u.LastLoginAt() != nil {
		response["last_login_at"] = u.LastLoginAt().Format("2006-01-02T15:04:05Z07:00")
	}

	return response
}

func (h *UserHandler) writeJSONResponse(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Error encoding JSON response: %v", err)
	}
}

func (h *UserHandler) writeErrorResponse(w http.ResponseWriter, message string, statusCode int) {
	response := map[string]interface{}{
		"error": message,
	}
	h.writeJSONResponse(w, response, statusCode)
}
