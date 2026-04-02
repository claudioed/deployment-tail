package http

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/claudioed/deployment-tail/api"
	"github.com/claudioed/deployment-tail/internal/application/ports/input"
	"github.com/claudioed/deployment-tail/internal/domain/schedule"
	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

// ScheduleHandler implements the API server interface
type ScheduleHandler struct {
	service input.ScheduleService
}

// NewScheduleHandler creates a new schedule handler
func NewScheduleHandler(service input.ScheduleService) *ScheduleHandler {
	return &ScheduleHandler{service: service}
}

// ListSchedules handles GET /schedules
func (h *ScheduleHandler) ListSchedules(w http.ResponseWriter, r *http.Request, params api.ListSchedulesParams) {
	var envStr *string
	if params.Environment != nil {
		env := string(*params.Environment)
		envStr = &env
	}

	var ownerStr *string
	if params.Owner != nil {
		ownerStr = params.Owner
	}

	var statusStr *string
	if params.Status != nil {
		status := string(*params.Status)
		statusStr = &status
	}

	query := input.ListSchedulesQuery{
		From:        params.From,
		To:          params.To,
		Environment: envStr,
		Owner:       ownerStr,
		Status:      statusStr,
	}

	schedules, err := h.service.ListSchedules(r.Context(), query)
	if err != nil {
		h.writeError(w, err, http.StatusInternalServerError)
		return
	}

	response := make([]api.Schedule, len(schedules))
	for i, sch := range schedules {
		response[i] = h.toAPISchedule(sch)
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

	cmd := input.CreateScheduleCommand{
		ScheduledAt: req.ScheduledAt,
		ServiceName: req.ServiceName,
		Environment: string(req.Environment),
		Owner:       req.Owner,
	}

	if req.Description != nil {
		cmd.Description = *req.Description
	}

	if req.RollbackPlan != nil {
		cmd.RollbackPlan = *req.RollbackPlan
	}

	sch, err := h.service.CreateSchedule(r.Context(), cmd)
	if err != nil {
		status := http.StatusBadRequest
		if errors.Is(err, schedule.ErrScheduleAlreadyExists) {
			status = http.StatusConflict
		}
		h.writeError(w, err, status)
		return
	}

	h.writeJSON(w, h.toAPISchedule(sch), http.StatusCreated)
}

// GetSchedule handles GET /schedules/{id}
func (h *ScheduleHandler) GetSchedule(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	sch, err := h.service.GetSchedule(r.Context(), id.String())
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, schedule.ErrScheduleNotFound) {
			status = http.StatusNotFound
		}
		h.writeError(w, err, status)
		return
	}

	h.writeJSON(w, h.toAPISchedule(sch), http.StatusOK)
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

	if req.Environment != nil {
		env := string(*req.Environment)
		cmd.Environment = &env
	}

	sch, err := h.service.UpdateSchedule(r.Context(), cmd)
	if err != nil {
		status := http.StatusBadRequest
		if errors.Is(err, schedule.ErrScheduleNotFound) {
			status = http.StatusNotFound
		}
		h.writeError(w, err, status)
		return
	}

	h.writeJSON(w, h.toAPISchedule(sch), http.StatusOK)
}

// DeleteSchedule handles DELETE /schedules/{id}
func (h *ScheduleHandler) DeleteSchedule(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	err := h.service.DeleteSchedule(r.Context(), id.String())
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

	sch, err := h.service.ApproveSchedule(r.Context(), cmd)
	if err != nil {
		status := http.StatusBadRequest
		if errors.Is(err, schedule.ErrScheduleNotFound) {
			status = http.StatusNotFound
		}
		h.writeError(w, err, status)
		return
	}

	h.writeJSON(w, h.toAPISchedule(sch), http.StatusOK)
}

// DenySchedule handles POST /schedules/{id}/deny
func (h *ScheduleHandler) DenySchedule(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	cmd := input.DenyScheduleCommand{
		ID: id.String(),
	}

	sch, err := h.service.DenySchedule(r.Context(), cmd)
	if err != nil {
		status := http.StatusBadRequest
		if errors.Is(err, schedule.ErrScheduleNotFound) {
			status = http.StatusNotFound
		}
		h.writeError(w, err, status)
		return
	}

	h.writeJSON(w, h.toAPISchedule(sch), http.StatusOK)
}

// toAPISchedule converts domain schedule to API schedule
func (h *ScheduleHandler) toAPISchedule(sch *schedule.Schedule) api.Schedule {
	id := uuid.MustParse(sch.ID().String())
	apiSch := api.Schedule{
		Id:          id,
		ScheduledAt: sch.ScheduledAt().Value(),
		ServiceName: sch.Service().Value(),
		Environment: api.ScheduleEnvironment(sch.Environment().String()),
		Owner:       sch.Owner().String(),
		Status:      api.ScheduleStatus(sch.Status().String()),
		CreatedAt:   sch.CreatedAt(),
		UpdatedAt:   sch.UpdatedAt(),
	}

	if !sch.Description().IsEmpty() {
		desc := sch.Description().Value()
		apiSch.Description = &desc
	}

	if !sch.RollbackPlan().IsEmpty() {
		plan := sch.RollbackPlan().String()
		apiSch.RollbackPlan = &plan
	}

	return apiSch
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
