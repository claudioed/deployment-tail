package http

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/claudioed/deployment-tail/api"
	"github.com/claudioed/deployment-tail/internal/adapters/input/http/middleware"
	"github.com/claudioed/deployment-tail/internal/application/ports/input"
	"github.com/claudioed/deployment-tail/internal/domain/group"
	"github.com/claudioed/deployment-tail/internal/domain/schedule"
	"github.com/claudioed/deployment-tail/internal/domain/user"
	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

// ScheduleHandler implements the API server interface
type ScheduleHandler struct {
	scheduleService input.ScheduleService
	groupService    input.GroupService
	userService     input.UserService
}

// NewScheduleHandler creates a new schedule handler
func NewScheduleHandler(scheduleService input.ScheduleService, groupService input.GroupService, userService input.UserService) *ScheduleHandler {
	return &ScheduleHandler{
		scheduleService: scheduleService,
		groupService:    groupService,
		userService:     userService,
	}
}

// ListSchedules handles GET /schedules
func (h *ScheduleHandler) ListSchedules(w http.ResponseWriter, r *http.Request, params api.ListSchedulesParams) {
	var environments []string
	if params.Environment != nil {
		for _, env := range *params.Environment {
			environments = append(environments, string(env))
		}
	}

	var owners []string
	if params.Owner != nil {
		owners = *params.Owner
	}

	var statusStr *string
	if params.Status != nil {
		status := string(*params.Status)
		statusStr = &status
	}

	query := input.ListSchedulesQuery{
		From:         params.From,
		To:           params.To,
		Environments: environments,
		Owners:       owners,
		Status:       statusStr,
	}

	schedules, err := h.scheduleService.ListSchedules(r.Context(), query)
	if err != nil {
		h.writeError(w, err, http.StatusInternalServerError)
		return
	}

	response := make([]api.Schedule, len(schedules))
	for i, sch := range schedules {
		apiSch, err := h.toAPIScheduleWithGroups(r.Context(), sch)
		if err != nil {
			h.writeError(w, err, http.StatusInternalServerError)
			return
		}
		response[i] = apiSch
	}

	h.writeJSON(w, response, http.StatusOK)
}

// CreateSchedule handles POST /schedules
func (h *ScheduleHandler) CreateSchedule(w http.ResponseWriter, r *http.Request) {
	var req api.CreateScheduleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, errors.New("invalid request body"), http.StatusBadRequest)
		return
	}

	// Convert environments
	environments := make([]string, len(req.Environments))
	for i, env := range req.Environments {
		environments[i] = string(env)
	}

	cmd := input.CreateScheduleCommand{
		ScheduledAt:  req.ScheduledAt,
		ServiceName:  req.ServiceName,
		Environments: environments,
		Owners:       req.Owners,
	}

	if req.Description != nil {
		cmd.Description = *req.Description
	}

	if req.RollbackPlan != nil {
		cmd.RollbackPlan = *req.RollbackPlan
	}

	// Get authenticated user from context
	authenticatedUser, err := middleware.UserFromContext(r.Context())
	if err != nil {
		h.writeError(w, fmt.Errorf("authentication required"), http.StatusUnauthorized)
		return
	}

	sch, err := h.scheduleService.CreateSchedule(r.Context(), cmd, authenticatedUser.ID())
	if err != nil {
		status := http.StatusBadRequest
		if errors.Is(err, schedule.ErrScheduleAlreadyExists) {
			status = http.StatusConflict
		}
		h.writeError(w, err, status)
		return
	}

	apiSch, err := h.toAPIScheduleWithGroups(r.Context(), sch)
	if err != nil {
		h.writeError(w, err, http.StatusInternalServerError)
		return
	}
	h.writeJSON(w, apiSch, http.StatusCreated)
}

// GetSchedule handles GET /schedules/{id}
func (h *ScheduleHandler) GetSchedule(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	sch, err := h.scheduleService.GetSchedule(r.Context(), id.String())
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, schedule.ErrScheduleNotFound) {
			status = http.StatusNotFound
		}
		h.writeError(w, err, status)
		return
	}

	apiSch, err := h.toAPIScheduleWithGroups(r.Context(), sch)
	if err != nil {
		h.writeError(w, err, http.StatusInternalServerError)
		return
	}
	h.writeJSON(w, apiSch, http.StatusOK)
}

// UpdateSchedule handles PUT /schedules/{id}
func (h *ScheduleHandler) UpdateSchedule(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	var req api.UpdateScheduleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, errors.New("invalid request body"), http.StatusBadRequest)
		return
	}

	cmd := input.UpdateScheduleCommand{
		ID:           id.String(),
		ScheduledAt:  req.ScheduledAt,
		ServiceName:  req.ServiceName,
		Description:  req.Description,
		RollbackPlan: req.RollbackPlan,
	}

	if req.Environments != nil {
		environments := make([]string, len(*req.Environments))
		for i, env := range *req.Environments {
			environments[i] = string(env)
		}
		cmd.Environments = &environments
	}

	if req.Owners != nil {
		cmd.Owners = req.Owners
	}

	// Get authenticated user from context
	authenticatedUser, err := middleware.UserFromContext(r.Context())
	if err != nil {
		h.writeError(w, fmt.Errorf("authentication required"), http.StatusUnauthorized)
		return
	}

	sch, err := h.scheduleService.UpdateSchedule(r.Context(), cmd, authenticatedUser.ID())
	if err != nil {
		status := http.StatusBadRequest
		if errors.Is(err, schedule.ErrScheduleNotFound) {
			status = http.StatusNotFound
		}
		h.writeError(w, err, status)
		return
	}

	apiSch, err := h.toAPIScheduleWithGroups(r.Context(), sch)
	if err != nil {
		h.writeError(w, err, http.StatusInternalServerError)
		return
	}
	h.writeJSON(w, apiSch, http.StatusOK)
}

// DeleteSchedule handles DELETE /schedules/{id}
func (h *ScheduleHandler) DeleteSchedule(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	// Get authenticated user from context
	authenticatedUser, err := middleware.UserFromContext(r.Context())
	if err != nil {
		h.writeError(w, fmt.Errorf("authentication required"), http.StatusUnauthorized)
		return
	}

	err = h.scheduleService.DeleteSchedule(r.Context(), id.String(), authenticatedUser.ID())
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, schedule.ErrScheduleNotFound) {
			status = http.StatusNotFound
		}
		h.writeError(w, err, status)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ApproveSchedule handles POST /schedules/{id}/approve
func (h *ScheduleHandler) ApproveSchedule(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	cmd := input.ApproveScheduleCommand{
		ID: id.String(),
	}

	sch, err := h.scheduleService.ApproveSchedule(r.Context(), cmd)
	if err != nil {
		status := http.StatusBadRequest
		if errors.Is(err, schedule.ErrScheduleNotFound) {
			status = http.StatusNotFound
		}
		h.writeError(w, err, status)
		return
	}

	apiSch, err := h.toAPIScheduleWithGroups(r.Context(), sch)
	if err != nil {
		h.writeError(w, err, http.StatusInternalServerError)
		return
	}
	h.writeJSON(w, apiSch, http.StatusOK)
}

// DenySchedule handles POST /schedules/{id}/deny
func (h *ScheduleHandler) DenySchedule(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	cmd := input.DenyScheduleCommand{
		ID: id.String(),
	}

	sch, err := h.scheduleService.DenySchedule(r.Context(), cmd)
	if err != nil {
		status := http.StatusBadRequest
		if errors.Is(err, schedule.ErrScheduleNotFound) {
			status = http.StatusNotFound
		}
		h.writeError(w, err, status)
		return
	}

	apiSch, err := h.toAPIScheduleWithGroups(r.Context(), sch)
	if err != nil {
		h.writeError(w, err, http.StatusInternalServerError)
		return
	}
	h.writeJSON(w, apiSch, http.StatusOK)
}

// toAPISchedule converts domain schedule to API schedule
func (h *ScheduleHandler) toAPISchedule(ctx context.Context, sch *schedule.Schedule) (api.Schedule, error) {
	id := uuid.MustParse(sch.ID().String())

	// Convert environments to API format and sort alphabetically
	environments := make([]api.ScheduleEnvironments, len(sch.Environments()))
	for i, env := range sch.Environments() {
		environments[i] = api.ScheduleEnvironments(env.String())
	}

	// Convert owners to strings and sort alphabetically (already sorted from domain)
	owners := make([]string, len(sch.Owners()))
	for i, owner := range sch.Owners() {
		owners[i] = owner.String()
	}

	// Fetch creator user
	creatorUser, err := h.userService.GetUserProfile(ctx, sch.CreatedBy())
	if err != nil {
		return api.Schedule{}, fmt.Errorf("failed to fetch creator user: %w", err)
	}

	apiSch := api.Schedule{
		Id:           id,
		ScheduledAt:  sch.ScheduledAt().Value(),
		ServiceName:  sch.Service().Value(),
		Environments: environments,
		Owners:       owners,
		Status:       api.ScheduleStatus(sch.Status().String()),
		CreatedAt:    sch.CreatedAt(),
		UpdatedAt:    sch.UpdatedAt(),
		CreatedBy:    h.toAPIUser(creatorUser),
	}

	// Fetch updater user if different from creator
	if !sch.UpdatedBy().Equals(sch.CreatedBy()) {
		updaterUser, err := h.userService.GetUserProfile(ctx, sch.UpdatedBy())
		if err != nil {
			return api.Schedule{}, fmt.Errorf("failed to fetch updater user: %w", err)
		}
		apiUser := h.toAPIUser(updaterUser)
		apiSch.UpdatedBy = &apiUser
	}

	if !sch.Description().IsEmpty() {
		desc := sch.Description().Value()
		apiSch.Description = &desc
	}

	if !sch.RollbackPlan().IsEmpty() {
		plan := sch.RollbackPlan().String()
		apiSch.RollbackPlan = &plan
	}

	return apiSch, nil
}

// toAPIScheduleWithGroups converts domain schedule to API schedule with groups
func (h *ScheduleHandler) toAPIScheduleWithGroups(ctx context.Context, sch *schedule.Schedule) (api.Schedule, error) {
	apiSch, err := h.toAPISchedule(ctx, sch)
	if err != nil {
		return api.Schedule{}, err
	}

	// Fetch groups for this schedule
	groups, err := h.groupService.GetGroupsForSchedule(ctx, sch.ID().String())
	if err != nil {
		return apiSch, err
	}

	// Convert groups to API format
	apiGroups := make([]api.Group, len(groups))
	for i, grp := range groups {
		apiGroups[i] = h.toAPIGroup(grp)
	}

	apiSch.Groups = &apiGroups
	return apiSch, nil
}

// toAPIGroup converts domain group to API group
func (h *ScheduleHandler) toAPIGroup(grp *group.Group) api.Group {
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

// toAPIUser converts domain user to API user
func (h *ScheduleHandler) toAPIUser(u *user.User) api.User {
	userID := uuid.MustParse(u.ID().String())
	apiUser := api.User{
		Id:        openapi_types.UUID(userID),
		Email:     openapi_types.Email(u.Email().String()),
		Name:      u.Name().String(),
		Role:      api.UserRole(u.Role().String()),
		CreatedAt: u.CreatedAt(),
		UpdatedAt: u.UpdatedAt(),
	}

	if u.LastLoginAt() != nil {
		apiUser.LastLoginAt = u.LastLoginAt()
	}

	return apiUser
}

// writeJSON writes a JSON response
func (h *ScheduleHandler) writeJSON(w http.ResponseWriter, data interface{}, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// writeError writes an error response
func (h *ScheduleHandler) writeError(w http.ResponseWriter, err error, status int) {
	errorResponse := api.Error{
		Message: err.Error(),
	}
	h.writeJSON(w, errorResponse, status)
}
