package http

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/claudioed/deployment-tail/api"
	"github.com/claudioed/deployment-tail/internal/adapters/input/http/middleware"
	"github.com/claudioed/deployment-tail/internal/application/ports/input"
	"github.com/claudioed/deployment-tail/internal/domain/group"
	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

// GroupHandler implements the group API endpoints
type GroupHandler struct {
	groupService    input.GroupService
	scheduleService input.ScheduleService
}

// NewGroupHandler creates a new group handler
func NewGroupHandler(groupService input.GroupService, scheduleService input.ScheduleService) *GroupHandler {
	return &GroupHandler{
		groupService:    groupService,
		scheduleService: scheduleService,
	}
}

// ListGroups handles GET /groups
func (h *GroupHandler) ListGroups(w http.ResponseWriter, r *http.Request, params api.ListGroupsParams) {
	if params.Owner == "" {
		h.writeError(w, errors.New("owner parameter is required"), http.StatusBadRequest)
		return
	}

	query := input.ListGroupsQuery{
		Owner: params.Owner,
	}

	// Try to get authenticated user from context
	userID, err := middleware.UserIDFromContext(r.Context())

	// If user is authenticated, use favorites-aware listing
	if err == nil && userID != "" {
		groups, favorites, err := h.groupService.ListGroupsWithFavorites(r.Context(), query, userID)
		if err != nil {
			h.writeError(w, err, http.StatusInternalServerError)
			return
		}

		response := make([]api.Group, len(groups))
		for i, grp := range groups {
			isFavorite := favorites[grp.ID().String()]
			response[i] = h.toAPIGroupWithFavorite(grp, isFavorite)
		}

		h.writeJSON(w, response, http.StatusOK)
		return
	}

	// Fallback to non-favorites listing for unauthenticated users
	groups, err := h.groupService.ListGroups(r.Context(), query)
	if err != nil {
		h.writeError(w, err, http.StatusInternalServerError)
		return
	}

	response := make([]api.Group, len(groups))
	for i, grp := range groups {
		response[i] = h.toAPIGroup(grp)
	}

	h.writeJSON(w, response, http.StatusOK)
}

// CreateGroup handles POST /groups
func (h *GroupHandler) CreateGroup(w http.ResponseWriter, r *http.Request) {
	var req api.CreateGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, errors.New("invalid request body"), http.StatusBadRequest)
		return
	}

	cmd := input.CreateGroupCommand{
		Name:  req.Name,
		Owner: req.Owner,
	}

	if req.Description != nil {
		cmd.Description = *req.Description
	}

	grp, err := h.groupService.CreateGroup(r.Context(), cmd)
	if err != nil {
		status := http.StatusBadRequest
		if errors.Is(err, group.ErrDuplicateGroupName) {
			status = http.StatusConflict
		}
		h.writeError(w, err, status)
		return
	}

	h.writeJSON(w, h.toAPIGroup(grp), http.StatusCreated)
}

// GetGroup handles GET /groups/{id}
func (h *GroupHandler) GetGroup(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	grp, err := h.groupService.GetGroup(r.Context(), id.String())
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, group.ErrGroupNotFound) {
			status = http.StatusNotFound
		}
		h.writeError(w, err, status)
		return
	}

	h.writeJSON(w, h.toAPIGroup(grp), http.StatusOK)
}

// UpdateGroup handles PUT /groups/{id}
func (h *GroupHandler) UpdateGroup(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	var req api.UpdateGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, errors.New("invalid request body"), http.StatusBadRequest)
		return
	}

	cmd := input.UpdateGroupCommand{
		ID:   id.String(),
		Name: req.Name,
	}

	if req.Description != nil {
		cmd.Description = *req.Description
	}

	grp, err := h.groupService.UpdateGroup(r.Context(), cmd)
	if err != nil {
		status := http.StatusBadRequest
		if errors.Is(err, group.ErrGroupNotFound) {
			status = http.StatusNotFound
		} else if errors.Is(err, group.ErrDuplicateGroupName) {
			status = http.StatusConflict
		}
		h.writeError(w, err, status)
		return
	}

	h.writeJSON(w, h.toAPIGroup(grp), http.StatusOK)
}

// DeleteGroup handles DELETE /groups/{id}
func (h *GroupHandler) DeleteGroup(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	cmd := input.DeleteGroupCommand{
		ID: id.String(),
	}

	err := h.groupService.DeleteGroup(r.Context(), cmd)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, group.ErrGroupNotFound) {
			status = http.StatusNotFound
		}
		h.writeError(w, err, status)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// AssignScheduleToGroups handles POST /schedules/{id}/groups
func (h *GroupHandler) AssignScheduleToGroups(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	var req api.AssignScheduleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, errors.New("invalid request body"), http.StatusBadRequest)
		return
	}

	groupIDs := make([]string, len(req.GroupIds))
	for i, gid := range req.GroupIds {
		groupIDs[i] = gid.String()
	}

	assignedBy := ""
	if req.AssignedBy != nil {
		assignedBy = *req.AssignedBy
	}

	cmd := input.AssignScheduleCommand{
		ScheduleID: id.String(),
		GroupIDs:   groupIDs,
		AssignedBy: assignedBy,
	}

	err := h.groupService.AssignScheduleToGroups(r.Context(), cmd)
	if err != nil {
		h.writeError(w, err, http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// UnassignScheduleFromGroup handles DELETE /schedules/{id}/groups/{groupId}
func (h *GroupHandler) UnassignScheduleFromGroup(w http.ResponseWriter, r *http.Request, id openapi_types.UUID, groupId openapi_types.UUID) {
	cmd := input.UnassignScheduleCommand{
		ScheduleID: id.String(),
		GroupID:    groupId.String(),
	}

	err := h.groupService.UnassignScheduleFromGroup(r.Context(), cmd)
	if err != nil {
		h.writeError(w, err, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetGroupsForSchedule handles GET /schedules/{id}/groups
func (h *GroupHandler) GetGroupsForSchedule(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	groups, err := h.groupService.GetGroupsForSchedule(r.Context(), id.String())
	if err != nil {
		h.writeError(w, err, http.StatusInternalServerError)
		return
	}

	response := make([]api.Group, len(groups))
	for i, grp := range groups {
		response[i] = h.toAPIGroup(grp)
	}

	h.writeJSON(w, response, http.StatusOK)
}

// GetSchedulesInGroup handles GET /groups/{id}/schedules
func (h *GroupHandler) GetSchedulesInGroup(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	scheduleIDs, err := h.groupService.GetSchedulesInGroup(r.Context(), id.String())
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, group.ErrGroupNotFound) {
			status = http.StatusNotFound
		}
		h.writeError(w, err, status)
		return
	}

	// Convert string IDs to UUIDs
	response := make([]openapi_types.UUID, len(scheduleIDs))
	for i, sid := range scheduleIDs {
		uid, err := uuid.Parse(sid)
		if err != nil {
			h.writeError(w, err, http.StatusInternalServerError)
			return
		}
		response[i] = uid
	}

	h.writeJSON(w, response, http.StatusOK)
}

// BulkAssignSchedules handles POST /groups/{id}/schedules
func (h *GroupHandler) BulkAssignSchedules(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	var req api.BulkAssignRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, errors.New("invalid request body"), http.StatusBadRequest)
		return
	}

	scheduleIDs := make([]string, len(req.ScheduleIds))
	for i, sid := range req.ScheduleIds {
		scheduleIDs[i] = sid.String()
	}

	assignedBy := ""
	if req.AssignedBy != nil {
		assignedBy = *req.AssignedBy
	}

	cmd := input.BulkAssignCommand{
		GroupID:     id.String(),
		ScheduleIDs: scheduleIDs,
		AssignedBy:  assignedBy,
	}

	err := h.groupService.BulkAssignSchedules(r.Context(), cmd)
	if err != nil {
		h.writeError(w, err, http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// FavoriteGroup handles POST /groups/{id}/favorite
func (h *GroupHandler) FavoriteGroup(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	// Extract authenticated user from context
	userID, err := middleware.UserIDFromContext(r.Context())
	if err != nil {
		h.writeError(w, errors.New("authentication required"), http.StatusUnauthorized)
		return
	}

	// Call favorite use case
	err = h.groupService.FavoriteGroup(r.Context(), userID, id.String())
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, group.ErrGroupNotFound) {
			status = http.StatusNotFound
		}
		h.writeError(w, err, status)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// UnfavoriteGroup handles DELETE /groups/{id}/favorite
func (h *GroupHandler) UnfavoriteGroup(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	// Extract authenticated user from context
	userID, err := middleware.UserIDFromContext(r.Context())
	if err != nil {
		h.writeError(w, errors.New("authentication required"), http.StatusUnauthorized)
		return
	}

	// Call unfavorite use case
	err = h.groupService.UnfavoriteGroup(r.Context(), userID, id.String())
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, group.ErrGroupNotFound) {
			status = http.StatusNotFound
		}
		h.writeError(w, err, status)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// toAPIGroup converts domain group to API group
func (h *GroupHandler) toAPIGroup(grp *group.Group) api.Group {
	id := uuid.MustParse(grp.ID().String())
	apiGroup := api.Group{
		Id:        id,
		Name:      grp.Name().String(),
		Owner:     grp.Owner().String(),
		CreatedAt: grp.CreatedAt(),
		UpdatedAt: grp.UpdatedAt(),
	}

	if !grp.Description().IsEmpty() {
		desc := grp.Description().String()
		apiGroup.Description = &desc
	}

	return apiGroup
}

// toAPIGroupWithFavorite converts domain group to API group with favorite status
func (h *GroupHandler) toAPIGroupWithFavorite(grp *group.Group, isFavorite bool) api.Group {
	apiGroup := h.toAPIGroup(grp)
	apiGroup.IsFavorite = &isFavorite
	return apiGroup
}

// writeJSON writes a JSON response
func (h *GroupHandler) writeJSON(w http.ResponseWriter, data interface{}, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// writeError writes an error response
func (h *GroupHandler) writeError(w http.ResponseWriter, err error, status int) {
	errorResponse := api.Error{
		Message: err.Error(),
	}
	h.writeJSON(w, errorResponse, status)
}
