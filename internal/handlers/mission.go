package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/google/uuid"
	"github.com/go-playground/validator/v10"

	"github.com/dfedick/gotak/internal/mission"
	"github.com/dfedick/gotak/pkg/logger"
)

// MissionService interface defines the methods required by the mission handler
type MissionService interface {
	CreateMission(ctx context.Context, req *mission.CreateMissionRequest) (*mission.Mission, error)
	GetMission(ctx context.Context, missionID uuid.UUID) (*mission.Mission, error)
	UpdateMission(ctx context.Context, missionID uuid.UUID, req *mission.UpdateMissionRequest) (*mission.Mission, error)
	DeleteMission(ctx context.Context, missionID uuid.UUID) error
	ListMissions(ctx context.Context, filter *mission.ListMissionsFilter) ([]*mission.Mission, int, error)
	UpdateMissionStatus(ctx context.Context, missionID uuid.UUID, status mission.MissionStatus, reason string) error
	GetMissionTimeline(ctx context.Context, missionID uuid.UUID) (*mission.Timeline, error)
	CreateTask(ctx context.Context, req *mission.CreateTaskRequest) (*mission.Task, error)
	GetTask(ctx context.Context, taskID uuid.UUID) (*mission.Task, error)
	UpdateTask(ctx context.Context, taskID uuid.UUID, req *mission.UpdateTaskRequest) (*mission.Task, error)
	DeleteTask(ctx context.Context, taskID uuid.UUID) error
	ListTasks(ctx context.Context, filter *mission.ListTasksFilter) ([]*mission.Task, int, error)
	AssignTask(ctx context.Context, taskID, assigneeID uuid.UUID) error
	UpdateTaskStatus(ctx context.Context, taskID uuid.UUID, status mission.TaskStatus) error
	CreateMilestone(ctx context.Context, req *mission.CreateMilestoneRequest) (*mission.Milestone, error)
	UpdateMilestone(ctx context.Context, milestoneID uuid.UUID, completed bool) error
}

// MissionHandler handles HTTP requests for mission management
type MissionHandler struct {
	missionService MissionService
	logger         *logger.Logger
	validator      *validator.Validate
}

// NewMissionHandler creates a new mission handler instance
func NewMissionHandler(missionService MissionService, logger *logger.Logger) *MissionHandler {
	return &MissionHandler{
		missionService: missionService,
		logger:         logger,
		validator:      validator.New(),
	}
}

// RegisterRoutes registers mission-related routes
func (h *MissionHandler) RegisterRoutes(r *mux.Router) {
	// Mission routes
	r.HandleFunc("/missions", h.ListMissions).Methods("GET")
	r.HandleFunc("/missions", h.CreateMission).Methods("POST")
	r.HandleFunc("/missions/{id}", h.GetMission).Methods("GET")
	r.HandleFunc("/missions/{id}", h.UpdateMission).Methods("PUT")
	r.HandleFunc("/missions/{id}", h.DeleteMission).Methods("DELETE")
	r.HandleFunc("/missions/{id}/status", h.UpdateMissionStatus).Methods("POST")
	r.HandleFunc("/missions/{id}/timeline", h.GetMissionTimeline).Methods("GET")
	
	// Task routes
	r.HandleFunc("/missions/{id}/tasks", h.ListMissionTasks).Methods("GET")
	r.HandleFunc("/missions/{id}/tasks", h.CreateTask).Methods("POST")
	r.HandleFunc("/tasks/{id}", h.GetTask).Methods("GET")
	r.HandleFunc("/tasks/{id}", h.UpdateTask).Methods("PUT")
	r.HandleFunc("/tasks/{id}", h.DeleteTask).Methods("DELETE")
	r.HandleFunc("/tasks/{id}/assign", h.AssignTask).Methods("POST")
	r.HandleFunc("/tasks/{id}/status", h.UpdateTaskStatus).Methods("POST")
	r.HandleFunc("/tasks", h.ListTasks).Methods("GET")
	
	// Milestone routes
	r.HandleFunc("/missions/{id}/milestones", h.CreateMilestone).Methods("POST")
	r.HandleFunc("/milestones/{id}", h.UpdateMilestone).Methods("PUT")
}

// Mission handlers

// ListMissions handles GET /missions
func (h *MissionHandler) ListMissions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	// Parse query parameters
	filter := &mission.ListMissionsFilter{
		Limit:  parseIntParam(r.URL.Query().Get("limit"), 50),
		Offset: parseIntParam(r.URL.Query().Get("offset"), 0),
	}
	
	if status := r.URL.Query().Get("status"); status != "" {
		missionStatus := mission.MissionStatus(status)
		if missionStatus.Valid() {
			filter.Status = &missionStatus
		}
	}
	
	if priority := r.URL.Query().Get("priority"); priority != "" {
		if p := parseIntParam(priority, 0); p > 0 {
			filter.Priority = &p
		}
	}
	
	if commanderID := r.URL.Query().Get("commander_id"); commanderID != "" {
		if id, err := uuid.Parse(commanderID); err == nil {
			filter.CommanderID = &id
		}
	}
	
	if startDate := r.URL.Query().Get("start_date_after"); startDate != "" {
		if date, err := time.Parse(time.RFC3339, startDate); err == nil {
			filter.StartDate = &date
		}
	}
	
	if endDate := r.URL.Query().Get("end_date_before"); endDate != "" {
		if date, err := time.Parse(time.RFC3339, endDate); err == nil {
			filter.EndDate = &date
		}
	}
	
	missions, total, err := h.missionService.ListMissions(ctx, filter)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to list missions")
		h.writeError(w, "Failed to list missions", http.StatusInternalServerError)
		return
	}
	
	response := map[string]interface{}{
		"missions": missions,
		"total":    total,
		"limit":    filter.Limit,
		"offset":   filter.Offset,
	}
	
	h.writeJSON(w, response, http.StatusOK)
}

// CreateMission handles POST /missions
func (h *MissionHandler) CreateMission(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	var req mission.CreateMissionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	
	if err := h.validator.Struct(req); err != nil {
		h.writeValidationError(w, err)
		return
	}
	
	// Validate classification
	if !req.Classification.Valid() {
		h.writeError(w, "Invalid classification level", http.StatusBadRequest)
		return
	}
	
	// Validate date ranges
	if req.StartDate != nil && req.EndDate != nil && req.EndDate.Before(*req.StartDate) {
		h.writeError(w, "End date must be after start date", http.StatusBadRequest)
		return
	}
	
	createdMission, err := h.missionService.CreateMission(ctx, &req)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to create mission")
		
		if strings.Contains(err.Error(), "insufficient permissions") {
			h.writeError(w, "Insufficient permissions", http.StatusForbidden)
			return
		}
		
		if strings.Contains(err.Error(), "not found") {
			h.writeError(w, "Referenced resource not found", http.StatusBadRequest)
			return
		}
		
		h.writeError(w, "Failed to create mission", http.StatusInternalServerError)
		return
	}
	
	h.writeJSON(w, createdMission, http.StatusCreated)
}

// GetMission handles GET /missions/{id}
func (h *MissionHandler) GetMission(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	
	missionID, err := uuid.Parse(vars["id"])
	if err != nil {
		h.writeError(w, "Invalid mission ID", http.StatusBadRequest)
		return
	}
	
	mission, err := h.missionService.GetMission(ctx, missionID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			h.writeError(w, "Mission not found", http.StatusNotFound)
			return
		}
		
		if strings.Contains(err.Error(), "insufficient permissions") {
			h.writeError(w, "Insufficient permissions", http.StatusForbidden)
			return
		}
		
		h.logger.Error().Err(err).Str("mission_id", missionID.String()).Msg("Failed to get mission")
		h.writeError(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	
	h.writeJSON(w, mission, http.StatusOK)
}

// UpdateMission handles PUT /missions/{id}
func (h *MissionHandler) UpdateMission(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	
	missionID, err := uuid.Parse(vars["id"])
	if err != nil {
		h.writeError(w, "Invalid mission ID", http.StatusBadRequest)
		return
	}
	
	var req mission.UpdateMissionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	
	if err := h.validator.Struct(req); err != nil {
		h.writeValidationError(w, err)
		return
	}
	
	// Validate classification if provided
	if req.Classification != nil && !req.Classification.Valid() {
		h.writeError(w, "Invalid classification level", http.StatusBadRequest)
		return
	}
	
	// Validate date ranges if both provided
	if req.StartDate != nil && req.EndDate != nil && req.EndDate.Before(*req.StartDate) {
		h.writeError(w, "End date must be after start date", http.StatusBadRequest)
		return
	}
	
	updatedMission, err := h.missionService.UpdateMission(ctx, missionID, &req)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			h.writeError(w, "Mission not found", http.StatusNotFound)
			return
		}
		
		if strings.Contains(err.Error(), "insufficient permissions") {
			h.writeError(w, "Insufficient permissions", http.StatusForbidden)
			return
		}
		
		h.logger.Error().Err(err).Str("mission_id", missionID.String()).Msg("Failed to update mission")
		h.writeError(w, "Failed to update mission", http.StatusInternalServerError)
		return
	}
	
	h.writeJSON(w, updatedMission, http.StatusOK)
}

// DeleteMission handles DELETE /missions/{id}
func (h *MissionHandler) DeleteMission(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	
	missionID, err := uuid.Parse(vars["id"])
	if err != nil {
		h.writeError(w, "Invalid mission ID", http.StatusBadRequest)
		return
	}
	
	err = h.missionService.DeleteMission(ctx, missionID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			h.writeError(w, "Mission not found", http.StatusNotFound)
			return
		}
		
		if strings.Contains(err.Error(), "insufficient permissions") {
			h.writeError(w, "Insufficient permissions", http.StatusForbidden)
			return
		}
		
		h.logger.Error().Err(err).Str("mission_id", missionID.String()).Msg("Failed to delete mission")
		h.writeError(w, "Failed to delete mission", http.StatusInternalServerError)
		return
	}
	
	response := map[string]interface{}{
		"status":  "success",
		"message": "Mission deleted successfully",
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// UpdateMissionStatus handles POST /missions/{id}/status
func (h *MissionHandler) UpdateMissionStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	
	missionID, err := uuid.Parse(vars["id"])
	if err != nil {
		h.writeError(w, "Invalid mission ID", http.StatusBadRequest)
		return
	}
	
	var req struct {
		Status mission.MissionStatus `json:"status" validate:"required"`
		Reason string               `json:"reason"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	
	if err := h.validator.Struct(req); err != nil {
		h.writeValidationError(w, err)
		return
	}
	
	if !req.Status.Valid() {
		h.writeError(w, "Invalid mission status", http.StatusBadRequest)
		return
	}
	
	err = h.missionService.UpdateMissionStatus(ctx, missionID, req.Status, req.Reason)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			h.writeError(w, "Mission not found", http.StatusNotFound)
			return
		}
		
		if strings.Contains(err.Error(), "insufficient permissions") {
			h.writeError(w, "Insufficient permissions", http.StatusForbidden)
			return
		}
		
		if strings.Contains(err.Error(), "invalid status transition") {
			h.writeError(w, err.Error(), http.StatusBadRequest)
			return
		}
		
		h.logger.Error().Err(err).Str("mission_id", missionID.String()).Msg("Failed to update mission status")
		h.writeError(w, "Failed to update mission status", http.StatusInternalServerError)
		return
	}
	
	response := map[string]interface{}{
		"status":  "success",
		"message": "Mission status updated successfully",
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// GetMissionTimeline handles GET /missions/{id}/timeline
func (h *MissionHandler) GetMissionTimeline(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	
	missionID, err := uuid.Parse(vars["id"])
	if err != nil {
		h.writeError(w, "Invalid mission ID", http.StatusBadRequest)
		return
	}
	
	timeline, err := h.missionService.GetMissionTimeline(ctx, missionID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			h.writeError(w, "Mission not found", http.StatusNotFound)
			return
		}
		
		if strings.Contains(err.Error(), "insufficient permissions") {
			h.writeError(w, "Insufficient permissions", http.StatusForbidden)
			return
		}
		
		h.logger.Error().Err(err).Str("mission_id", missionID.String()).Msg("Failed to get mission timeline")
		h.writeError(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	
	h.writeJSON(w, timeline, http.StatusOK)
}

// Task handlers

// ListMissionTasks handles GET /missions/{id}/tasks
func (h *MissionHandler) ListMissionTasks(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	
	missionID, err := uuid.Parse(vars["id"])
	if err != nil {
		h.writeError(w, "Invalid mission ID", http.StatusBadRequest)
		return
	}
	
	// Parse query parameters
	filter := &mission.ListTasksFilter{
		MissionID: &missionID,
		Limit:     parseIntParam(r.URL.Query().Get("limit"), 50),
		Offset:    parseIntParam(r.URL.Query().Get("offset"), 0),
	}
	
	if status := r.URL.Query().Get("status"); status != "" {
		taskStatus := mission.TaskStatus(status)
		if taskStatus.Valid() {
			filter.Status = &taskStatus
		}
	}
	
	if priority := r.URL.Query().Get("priority"); priority != "" {
		if p := parseIntParam(priority, 0); p > 0 {
			filter.Priority = &p
		}
	}
	
	if assignedTo := r.URL.Query().Get("assigned_to"); assignedTo != "" {
		if id, err := uuid.Parse(assignedTo); err == nil {
			filter.AssignedTo = &id
		}
	}
	
	if dueDate := r.URL.Query().Get("due_date_before"); dueDate != "" {
		if date, err := time.Parse(time.RFC3339, dueDate); err == nil {
			filter.DueDate = &date
		}
	}
	
	tasks, total, err := h.missionService.ListTasks(ctx, filter)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to list mission tasks")
		h.writeError(w, "Failed to list mission tasks", http.StatusInternalServerError)
		return
	}
	
	response := map[string]interface{}{
		"tasks":  tasks,
		"total":  total,
		"limit":  filter.Limit,
		"offset": filter.Offset,
	}
	
	h.writeJSON(w, response, http.StatusOK)
}

// ListTasks handles GET /tasks (all tasks for user's group)
func (h *MissionHandler) ListTasks(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	// Parse query parameters
	filter := &mission.ListTasksFilter{
		Limit:  parseIntParam(r.URL.Query().Get("limit"), 50),
		Offset: parseIntParam(r.URL.Query().Get("offset"), 0),
	}
	
	if missionID := r.URL.Query().Get("mission_id"); missionID != "" {
		if id, err := uuid.Parse(missionID); err == nil {
			filter.MissionID = &id
		}
	}
	
	if status := r.URL.Query().Get("status"); status != "" {
		taskStatus := mission.TaskStatus(status)
		if taskStatus.Valid() {
			filter.Status = &taskStatus
		}
	}
	
	if priority := r.URL.Query().Get("priority"); priority != "" {
		if p := parseIntParam(priority, 0); p > 0 {
			filter.Priority = &p
		}
	}
	
	if assignedTo := r.URL.Query().Get("assigned_to"); assignedTo != "" {
		if id, err := uuid.Parse(assignedTo); err == nil {
			filter.AssignedTo = &id
		}
	}
	
	if dueDate := r.URL.Query().Get("due_date_before"); dueDate != "" {
		if date, err := time.Parse(time.RFC3339, dueDate); err == nil {
			filter.DueDate = &date
		}
	}
	
	tasks, total, err := h.missionService.ListTasks(ctx, filter)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to list tasks")
		h.writeError(w, "Failed to list tasks", http.StatusInternalServerError)
		return
	}
	
	response := map[string]interface{}{
		"tasks":  tasks,
		"total":  total,
		"limit":  filter.Limit,
		"offset": filter.Offset,
	}
	
	h.writeJSON(w, response, http.StatusOK)
}

// CreateTask handles POST /missions/{id}/tasks
func (h *MissionHandler) CreateTask(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	
	missionID, err := uuid.Parse(vars["id"])
	if err != nil {
		h.writeError(w, "Invalid mission ID", http.StatusBadRequest)
		return
	}
	
	var req mission.CreateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	
	// Set mission ID from URL
	req.MissionID = missionID
	
	if err := h.validator.Struct(req); err != nil {
		h.writeValidationError(w, err)
		return
	}
	
	createdTask, err := h.missionService.CreateTask(ctx, &req)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to create task")
		
		if strings.Contains(err.Error(), "insufficient permissions") {
			h.writeError(w, "Insufficient permissions", http.StatusForbidden)
			return
		}
		
		if strings.Contains(err.Error(), "not found") {
			h.writeError(w, "Mission not found", http.StatusBadRequest)
			return
		}
		
		if strings.Contains(err.Error(), "invalid task dependencies") {
			h.writeError(w, err.Error(), http.StatusBadRequest)
			return
		}
		
		h.writeError(w, "Failed to create task", http.StatusInternalServerError)
		return
	}
	
	h.writeJSON(w, createdTask, http.StatusCreated)
}

// GetTask handles GET /tasks/{id}
func (h *MissionHandler) GetTask(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	
	taskID, err := uuid.Parse(vars["id"])
	if err != nil {
		h.writeError(w, "Invalid task ID", http.StatusBadRequest)
		return
	}
	
	task, err := h.missionService.GetTask(ctx, taskID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			h.writeError(w, "Task not found", http.StatusNotFound)
			return
		}
		
		if strings.Contains(err.Error(), "insufficient permissions") {
			h.writeError(w, "Insufficient permissions", http.StatusForbidden)
			return
		}
		
		h.logger.Error().Err(err).Str("task_id", taskID.String()).Msg("Failed to get task")
		h.writeError(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	
	h.writeJSON(w, task, http.StatusOK)
}

// UpdateTask handles PUT /tasks/{id}
func (h *MissionHandler) UpdateTask(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	
	taskID, err := uuid.Parse(vars["id"])
	if err != nil {
		h.writeError(w, "Invalid task ID", http.StatusBadRequest)
		return
	}
	
	var req mission.UpdateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	
	if err := h.validator.Struct(req); err != nil {
		h.writeValidationError(w, err)
		return
	}
	
	updatedTask, err := h.missionService.UpdateTask(ctx, taskID, &req)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			h.writeError(w, "Task not found", http.StatusNotFound)
			return
		}
		
		if strings.Contains(err.Error(), "insufficient permissions") {
			h.writeError(w, "Insufficient permissions", http.StatusForbidden)
			return
		}
		
		h.logger.Error().Err(err).Str("task_id", taskID.String()).Msg("Failed to update task")
		h.writeError(w, "Failed to update task", http.StatusInternalServerError)
		return
	}
	
	h.writeJSON(w, updatedTask, http.StatusOK)
}

// DeleteTask handles DELETE /tasks/{id}
func (h *MissionHandler) DeleteTask(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	
	taskID, err := uuid.Parse(vars["id"])
	if err != nil {
		h.writeError(w, "Invalid task ID", http.StatusBadRequest)
		return
	}
	
	err = h.missionService.DeleteTask(ctx, taskID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			h.writeError(w, "Task not found", http.StatusNotFound)
			return
		}
		
		if strings.Contains(err.Error(), "insufficient permissions") {
			h.writeError(w, "Insufficient permissions", http.StatusForbidden)
			return
		}
		
		if strings.Contains(err.Error(), "other tasks depend") {
			h.writeError(w, err.Error(), http.StatusConflict)
			return
		}
		
		h.logger.Error().Err(err).Str("task_id", taskID.String()).Msg("Failed to delete task")
		h.writeError(w, "Failed to delete task", http.StatusInternalServerError)
		return
	}
	
	response := map[string]interface{}{
		"status":  "success",
		"message": "Task deleted successfully",
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// AssignTask handles POST /tasks/{id}/assign
func (h *MissionHandler) AssignTask(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	
	taskID, err := uuid.Parse(vars["id"])
	if err != nil {
		h.writeError(w, "Invalid task ID", http.StatusBadRequest)
		return
	}
	
	var req struct {
		AssignedTo uuid.UUID `json:"assigned_to" validate:"required"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	
	if err := h.validator.Struct(req); err != nil {
		h.writeValidationError(w, err)
		return
	}
	
	err = h.missionService.AssignTask(ctx, taskID, req.AssignedTo)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			h.writeError(w, "Task not found", http.StatusNotFound)
			return
		}
		
		if strings.Contains(err.Error(), "insufficient permissions") {
			h.writeError(w, "Insufficient permissions", http.StatusForbidden)
			return
		}
		
		if strings.Contains(err.Error(), "cannot assign task") {
			h.writeError(w, err.Error(), http.StatusBadRequest)
			return
		}
		
		h.logger.Error().Err(err).Str("task_id", taskID.String()).Msg("Failed to assign task")
		h.writeError(w, "Failed to assign task", http.StatusInternalServerError)
		return
	}
	
	response := map[string]interface{}{
		"status":  "success",
		"message": "Task assigned successfully",
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// UpdateTaskStatus handles POST /tasks/{id}/status
func (h *MissionHandler) UpdateTaskStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	
	taskID, err := uuid.Parse(vars["id"])
	if err != nil {
		h.writeError(w, "Invalid task ID", http.StatusBadRequest)
		return
	}
	
	var req struct {
		Status mission.TaskStatus `json:"status" validate:"required"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	
	if err := h.validator.Struct(req); err != nil {
		h.writeValidationError(w, err)
		return
	}
	
	if !req.Status.Valid() {
		h.writeError(w, "Invalid task status", http.StatusBadRequest)
		return
	}
	
	err = h.missionService.UpdateTaskStatus(ctx, taskID, req.Status)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			h.writeError(w, "Task not found", http.StatusNotFound)
			return
		}
		
		if strings.Contains(err.Error(), "insufficient permissions") {
			h.writeError(w, "Insufficient permissions", http.StatusForbidden)
			return
		}
		
		if strings.Contains(err.Error(), "invalid status transition") || strings.Contains(err.Error(), "cannot start task") {
			h.writeError(w, err.Error(), http.StatusBadRequest)
			return
		}
		
		h.logger.Error().Err(err).Str("task_id", taskID.String()).Msg("Failed to update task status")
		h.writeError(w, "Failed to update task status", http.StatusInternalServerError)
		return
	}
	
	response := map[string]interface{}{
		"status":  "success",
		"message": "Task status updated successfully",
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// Milestone handlers

// CreateMilestone handles POST /missions/{id}/milestones
func (h *MissionHandler) CreateMilestone(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	
	missionID, err := uuid.Parse(vars["id"])
	if err != nil {
		h.writeError(w, "Invalid mission ID", http.StatusBadRequest)
		return
	}
	
	var req mission.CreateMilestoneRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	
	// Set mission ID from URL
	req.MissionID = missionID
	
	if err := h.validator.Struct(req); err != nil {
		h.writeValidationError(w, err)
		return
	}
	
	milestone, err := h.missionService.CreateMilestone(ctx, &req)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to create milestone")
		
		if strings.Contains(err.Error(), "insufficient permissions") {
			h.writeError(w, "Insufficient permissions", http.StatusForbidden)
			return
		}
		
		if strings.Contains(err.Error(), "not found") {
			h.writeError(w, "Mission not found", http.StatusBadRequest)
			return
		}
		
		h.writeError(w, "Failed to create milestone", http.StatusInternalServerError)
		return
	}
	
	h.writeJSON(w, milestone, http.StatusCreated)
}

// UpdateMilestone handles PUT /milestones/{id}
func (h *MissionHandler) UpdateMilestone(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	
	milestoneID, err := uuid.Parse(vars["id"])
	if err != nil {
		h.writeError(w, "Invalid milestone ID", http.StatusBadRequest)
		return
	}
	
	var req struct {
		Completed bool `json:"completed"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	
	err = h.missionService.UpdateMilestone(ctx, milestoneID, req.Completed)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			h.writeError(w, "Milestone not found", http.StatusNotFound)
			return
		}
		
		if strings.Contains(err.Error(), "insufficient permissions") {
			h.writeError(w, "Insufficient permissions", http.StatusForbidden)
			return
		}
		
		h.logger.Error().Err(err).Str("milestone_id", milestoneID.String()).Msg("Failed to update milestone")
		h.writeError(w, "Failed to update milestone", http.StatusInternalServerError)
		return
	}
	
	response := map[string]interface{}{
		"status":  "success",
		"message": "Milestone updated successfully",
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// Helper methods

func (h *MissionHandler) writeJSON(w http.ResponseWriter, data interface{}, status int) {
	response := map[string]interface{}{
		"status": "success",
		"data":   data,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(response)
}

func (h *MissionHandler) writeError(w http.ResponseWriter, message string, status int) {
	response := map[string]interface{}{
		"status":  "error",
		"message": message,
		"code":    status,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(response)
}

func (h *MissionHandler) writeValidationError(w http.ResponseWriter, err error) {
	var errors []string
	for _, err := range err.(validator.ValidationErrors) {
		errors = append(errors, fmt.Sprintf("%s: %s", err.Field(), err.Tag()))
	}

	response := map[string]interface{}{
		"status":           "error",
		"message":          "Validation failed",
		"code":             http.StatusBadRequest,
		"validation_errors": errors,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(response)
}

func parseIntParam(param string, defaultValue int) int {
	if param == "" {
		return defaultValue
	}
	
	if value, err := strconv.Atoi(param); err == nil && value > 0 {
		return value
	}
	
	return defaultValue
}
