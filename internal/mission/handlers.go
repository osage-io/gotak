package mission

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"

	"github.com/dfedick/gotak/pkg/logger"
)

// Handlers manages HTTP handlers for mission-related endpoints
type Handlers struct {
	service *Service
	logger  *logger.Logger
}

// NewHandlers creates a new instance of mission handlers
func NewHandlers(service *Service, logger *logger.Logger) *Handlers {
	return &Handlers{
		service: service,
		logger:  logger,
	}
}

// RegisterRoutes registers all mission-related routes
func (h *Handlers) RegisterRoutes(router *mux.Router) {
	// Mission management routes
	router.HandleFunc("/missions", WithAuthContext(h.CreateMission)).Methods(http.MethodPost)
	router.HandleFunc("/missions", WithAuthContext(h.ListMissions)).Methods(http.MethodGet)
	router.HandleFunc("/missions/{missionId}", WithAuthContext(h.GetMission)).Methods(http.MethodGet)
	router.HandleFunc("/missions/{missionId}", WithAuthContext(h.UpdateMission)).Methods(http.MethodPut)
	router.HandleFunc("/missions/{missionId}", WithAuthContext(h.DeleteMission)).Methods(http.MethodDelete)
	router.HandleFunc("/missions/{missionId}/status", WithAuthContext(h.UpdateMissionStatus)).Methods(http.MethodPatch)

	// Task management routes
	router.HandleFunc("/missions/{missionId}/tasks", WithAuthContext(h.CreateTask)).Methods(http.MethodPost)
	router.HandleFunc("/tasks", WithAuthContext(h.ListTasks)).Methods(http.MethodGet)
	router.HandleFunc("/tasks/{taskId}", WithAuthContext(h.GetTask)).Methods(http.MethodGet)
	router.HandleFunc("/tasks/{taskId}", WithAuthContext(h.UpdateTask)).Methods(http.MethodPut)
	router.HandleFunc("/tasks/{taskId}", WithAuthContext(h.DeleteTask)).Methods(http.MethodDelete)
	router.HandleFunc("/tasks/{taskId}/status", WithAuthContext(h.UpdateTaskStatus)).Methods(http.MethodPatch)
	router.HandleFunc("/tasks/{taskId}/assign", WithAuthContext(h.AssignTask)).Methods(http.MethodPatch)

	// Objective management routes
	router.HandleFunc("/missions/{missionId}/objectives", WithAuthContext(h.CreateObjective)).Methods(http.MethodPost)
	router.HandleFunc("/objectives/{objectiveId}", WithAuthContext(h.UpdateObjective)).Methods(http.MethodPut)
	router.HandleFunc("/objectives/{objectiveId}", WithAuthContext(h.DeleteObjective)).Methods(http.MethodDelete)
	router.HandleFunc("/objectives/{objectiveId}/complete", WithAuthContext(h.CompleteObjective)).Methods(http.MethodPatch)

	// Timeline and planning routes
	router.HandleFunc("/missions/{missionId}/timeline", WithAuthContext(h.GetMissionTimeline)).Methods(http.MethodGet)
	router.HandleFunc("/missions/{missionId}/participants", WithAuthContext(h.GetMissionParticipants)).Methods(http.MethodGet)
	router.HandleFunc("/missions/{missionId}/events", WithAuthContext(h.GetMissionEvents)).Methods(http.MethodGet)
}

// CreateMission handles mission creation requests
func (h *Handlers) CreateMission(w http.ResponseWriter, r *http.Request) {
	var req CreateMissionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Validate request
	if err := h.validateCreateMissionRequest(&req); err != nil {
		h.respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Create mission
	mission, err := h.service.CreateMission(r.Context(), &req)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to create mission")
		h.respondWithError(w, http.StatusInternalServerError, "Failed to create mission")
		return
	}

	h.respondWithJSON(w, http.StatusCreated, mission)
}

// GetMission handles mission retrieval requests
func (h *Handlers) GetMission(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	missionIDStr := vars["missionId"]

	missionID, err := uuid.Parse(missionIDStr)
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid mission ID")
		return
	}

	mission, err := h.service.GetMission(r.Context(), missionID)
	if err != nil {
		if err.Error() == "mission not found" {
			h.respondWithError(w, http.StatusNotFound, "Mission not found")
			return
		}
		h.logger.Error().Err(err).Msg("Failed to get mission")
		h.respondWithError(w, http.StatusInternalServerError, "Failed to get mission")
		return
	}

	h.respondWithJSON(w, http.StatusOK, mission)
}

// UpdateMission handles mission update requests
func (h *Handlers) UpdateMission(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	missionIDStr := vars["missionId"]

	missionID, err := uuid.Parse(missionIDStr)
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid mission ID")
		return
	}

	var req UpdateMissionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Validate request
	if err := h.validateUpdateMissionRequest(&req); err != nil {
		h.respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	mission, err := h.service.UpdateMission(r.Context(), missionID, &req)
	if err != nil {
		if err.Error() == "mission not found" {
			h.respondWithError(w, http.StatusNotFound, "Mission not found")
			return
		}
		h.logger.Error().Err(err).Msg("Failed to update mission")
		h.respondWithError(w, http.StatusInternalServerError, "Failed to update mission")
		return
	}

	h.respondWithJSON(w, http.StatusOK, mission)
}

// DeleteMission handles mission deletion requests
func (h *Handlers) DeleteMission(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	missionIDStr := vars["missionId"]

	missionID, err := uuid.Parse(missionIDStr)
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid mission ID")
		return
	}

	err = h.service.DeleteMission(r.Context(), missionID)
	if err != nil {
		if err.Error() == "mission not found" {
			h.respondWithError(w, http.StatusNotFound, "Mission not found")
			return
		}
		h.logger.Error().Err(err).Msg("Failed to delete mission")
		h.respondWithError(w, http.StatusInternalServerError, "Failed to delete mission")
		return
	}

	h.respondWithJSON(w, http.StatusNoContent, nil)
}

// ListMissions handles mission listing requests
func (h *Handlers) ListMissions(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	filter := h.parseMissionListFilter(r)

	missions, totalCount, err := h.service.ListMissions(r.Context(), filter)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to list missions")
		h.respondWithError(w, http.StatusInternalServerError, "Failed to list missions")
		return
	}

	response := map[string]interface{}{
		"missions":    missions,
		"total_count": totalCount,
		"limit":       filter.Limit,
		"offset":      filter.Offset,
	}

	h.respondWithJSON(w, http.StatusOK, response)
}

// UpdateMissionStatus handles mission status update requests
func (h *Handlers) UpdateMissionStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	missionIDStr := vars["missionId"]

	missionID, err := uuid.Parse(missionIDStr)
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid mission ID")
		return
	}

	var req struct {
		Status MissionStatus `json:"status"`
		Reason string        `json:"reason"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Validate status
	validStatuses := []MissionStatus{StatusPlanning, StatusApproved, StatusActive, StatusOnHold, StatusCompleted, StatusCancelled}
	statusValid := false
	for _, validStatus := range validStatuses {
		if req.Status == validStatus {
			statusValid = true
			break
		}
	}

	if !statusValid {
		h.respondWithError(w, http.StatusBadRequest, "Invalid mission status")
		return
	}
	if err != nil {
		if err.Error() == "mission not found" {
			h.respondWithError(w, http.StatusNotFound, "Mission not found")
			return
		}
		h.logger.Error().Err(err).Msg("Failed to update mission status")
		h.respondWithError(w, http.StatusInternalServerError, "Failed to update mission status")
		return
	}

	// Get the updated mission to return
	mission, err := h.service.GetMission(r.Context(), missionID)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to get updated mission")
		h.respondWithError(w, http.StatusInternalServerError, "Failed to get updated mission")
		return
	}

	h.respondWithJSON(w, http.StatusOK, mission)
}

// CreateTask handles task creation requests
func (h *Handlers) CreateTask(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	missionIDStr := vars["missionId"]

	missionID, err := uuid.Parse(missionIDStr)
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid mission ID")
		return
	}

	var req CreateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Set mission ID from URL
	req.MissionID = missionID

	task, err := h.service.CreateTask(r.Context(), &req)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to create task")
		h.respondWithError(w, http.StatusInternalServerError, "Failed to create task")
		return
	}

	h.respondWithJSON(w, http.StatusCreated, task)
}

// GetTask handles task retrieval requests
func (h *Handlers) GetTask(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	taskIDStr := vars["taskId"]

	taskID, err := uuid.Parse(taskIDStr)
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid task ID")
		return
	}

	task, err := h.service.GetTask(r.Context(), taskID)
	if err != nil {
		if err.Error() == "task not found" {
			h.respondWithError(w, http.StatusNotFound, "Task not found")
			return
		}
		h.logger.Error().Err(err).Msg("Failed to get task")
		h.respondWithError(w, http.StatusInternalServerError, "Failed to get task")
		return
	}

	h.respondWithJSON(w, http.StatusOK, task)
}

// UpdateTask handles task update requests
func (h *Handlers) UpdateTask(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	taskIDStr := vars["taskId"]

	taskID, err := uuid.Parse(taskIDStr)
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid task ID")
		return
	}

	var req UpdateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	task, err := h.service.UpdateTask(r.Context(), taskID, &req)
	if err != nil {
		if err.Error() == "task not found" {
			h.respondWithError(w, http.StatusNotFound, "Task not found")
			return
		}
		h.logger.Error().Err(err).Msg("Failed to update task")
		h.respondWithError(w, http.StatusInternalServerError, "Failed to update task")
		return
	}

	h.respondWithJSON(w, http.StatusOK, task)
}

// DeleteTask handles task deletion requests
func (h *Handlers) DeleteTask(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	taskIDStr := vars["taskId"]

	taskID, err := uuid.Parse(taskIDStr)
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid task ID")
		return
	}

	err = h.service.DeleteTask(r.Context(), taskID)
	if err != nil {
		if err.Error() == "task not found" {
			h.respondWithError(w, http.StatusNotFound, "Task not found")
			return
		}
		h.logger.Error().Err(err).Msg("Failed to delete task")
		h.respondWithError(w, http.StatusInternalServerError, "Failed to delete task")
		return
	}

	h.respondWithJSON(w, http.StatusNoContent, nil)
}

// ListTasks handles task listing requests
func (h *Handlers) ListTasks(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	filter := h.parseTaskListFilter(r)

	tasks, totalCount, err := h.service.ListTasks(r.Context(), filter)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to list tasks")
		h.respondWithError(w, http.StatusInternalServerError, "Failed to list tasks")
		return
	}

	response := map[string]interface{}{
		"tasks":       tasks,
		"total_count": totalCount,
		"limit":       filter.Limit,
		"offset":      filter.Offset,
	}

	h.respondWithJSON(w, http.StatusOK, response)
}

// UpdateTaskStatus handles task status update requests
func (h *Handlers) UpdateTaskStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	taskIDStr := vars["taskId"]

	taskID, err := uuid.Parse(taskIDStr)
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid task ID")
		return
	}

	var req struct {
		Status TaskStatus `json:"status"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	err = h.service.UpdateTaskStatus(r.Context(), taskID, req.Status)
	if err != nil {
		if err.Error() == "task not found" {
			h.respondWithError(w, http.StatusNotFound, "Task not found")
			return
		}
		h.logger.Error().Err(err).Msg("Failed to update task status")
		h.respondWithError(w, http.StatusInternalServerError, "Failed to update task status")
		return
	}

	// Get the updated task to return
	task, err := h.service.GetTask(r.Context(), taskID)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to get updated task")
		h.respondWithError(w, http.StatusInternalServerError, "Failed to get updated task")
		return
	}

	h.respondWithJSON(w, http.StatusOK, task)
}

// AssignTask handles task assignment requests
func (h *Handlers) AssignTask(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	taskIDStr := vars["taskId"]

	taskID, err := uuid.Parse(taskIDStr)
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid task ID")
		return
	}

	var req struct {
		AssignedTo uuid.UUID `json:"assigned_to"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	err = h.service.AssignTask(r.Context(), taskID, req.AssignedTo)
	if err != nil {
		if err.Error() == "task not found" {
			h.respondWithError(w, http.StatusNotFound, "Task not found")
			return
		}
		h.logger.Error().Err(err).Msg("Failed to assign task")
		h.respondWithError(w, http.StatusInternalServerError, "Failed to assign task")
		return
	}

	// Get the updated task to return
	task, err := h.service.GetTask(r.Context(), taskID)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to get updated task")
		h.respondWithError(w, http.StatusInternalServerError, "Failed to get updated task")
		return
	}

	h.respondWithJSON(w, http.StatusOK, task)
}

// CreateObjective handles objective creation requests
func (h *Handlers) CreateObjective(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	missionIDStr := vars["missionId"]

	missionID, err := uuid.Parse(missionIDStr)
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid mission ID")
		return
	}

	var req CreateObjectiveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	objective, err := h.service.CreateObjective(r.Context(), missionID, &req)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to create objective")
		h.respondWithError(w, http.StatusInternalServerError, "Failed to create objective")
		return
	}

	h.respondWithJSON(w, http.StatusCreated, objective)
}

// UpdateObjective handles objective update requests
func (h *Handlers) UpdateObjective(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	objectiveIDStr := vars["objectiveId"]

	objectiveID, err := uuid.Parse(objectiveIDStr)
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid objective ID")
		return
	}

	var req struct {
		Description string `json:"description"`
		Priority    int    `json:"priority"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	objective, err := h.service.UpdateObjective(r.Context(), objectiveID, req.Description, req.Priority)
	if err != nil {
		if err.Error() == "objective not found" {
			h.respondWithError(w, http.StatusNotFound, "Objective not found")
			return
		}
		h.logger.Error().Err(err).Msg("Failed to update objective")
		h.respondWithError(w, http.StatusInternalServerError, "Failed to update objective")
		return
	}

	h.respondWithJSON(w, http.StatusOK, objective)
}

// DeleteObjective handles objective deletion requests
func (h *Handlers) DeleteObjective(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	objectiveIDStr := vars["objectiveId"]

	objectiveID, err := uuid.Parse(objectiveIDStr)
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid objective ID")
		return
	}

	err = h.service.DeleteObjective(r.Context(), objectiveID)
	if err != nil {
		if err.Error() == "objective not found" {
			h.respondWithError(w, http.StatusNotFound, "Objective not found")
			return
		}
		h.logger.Error().Err(err).Msg("Failed to delete objective")
		h.respondWithError(w, http.StatusInternalServerError, "Failed to delete objective")
		return
	}

	h.respondWithJSON(w, http.StatusNoContent, nil)
}

// CompleteObjective handles objective completion requests
func (h *Handlers) CompleteObjective(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	objectiveIDStr := vars["objectiveId"]

	objectiveID, err := uuid.Parse(objectiveIDStr)
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid objective ID")
		return
	}

	objective, err := h.service.CompleteObjective(r.Context(), objectiveID)
	if err != nil {
		if err.Error() == "objective not found" {
			h.respondWithError(w, http.StatusNotFound, "Objective not found")
			return
		}
		h.logger.Error().Err(err).Msg("Failed to complete objective")
		h.respondWithError(w, http.StatusInternalServerError, "Failed to complete objective")
		return
	}

	h.respondWithJSON(w, http.StatusOK, objective)
}

// GetMissionTimeline handles mission timeline requests
func (h *Handlers) GetMissionTimeline(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	missionIDStr := vars["missionId"]

	missionID, err := uuid.Parse(missionIDStr)
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid mission ID")
		return
	}

	timeline, err := h.service.GetMissionTimeline(r.Context(), missionID)
	if err != nil {
		if err.Error() == "mission not found" {
			h.respondWithError(w, http.StatusNotFound, "Mission not found")
			return
		}
		h.logger.Error().Err(err).Msg("Failed to get mission timeline")
		h.respondWithError(w, http.StatusInternalServerError, "Failed to get mission timeline")
		return
	}

	h.respondWithJSON(w, http.StatusOK, timeline)
}

// GetMissionParticipants handles mission participants requests
func (h *Handlers) GetMissionParticipants(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	missionIDStr := vars["missionId"]

	missionID, err := uuid.Parse(missionIDStr)
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid mission ID")
		return
	}

	participants, err := h.service.GetMissionParticipants(r.Context(), missionID)
	if err != nil {
		if err.Error() == "mission not found" {
			h.respondWithError(w, http.StatusNotFound, "Mission not found")
			return
		}
		h.logger.Error().Err(err).Msg("Failed to get mission participants")
		h.respondWithError(w, http.StatusInternalServerError, "Failed to get mission participants")
		return
	}

	h.respondWithJSON(w, http.StatusOK, participants)
}

// GetMissionEvents handles mission events requests
func (h *Handlers) GetMissionEvents(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	missionIDStr := vars["missionId"]

	missionID, err := uuid.Parse(missionIDStr)
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid mission ID")
		return
	}

	events, err := h.service.GetMissionEvents(r.Context(), missionID)
	if err != nil {
		if err.Error() == "mission not found" {
			h.respondWithError(w, http.StatusNotFound, "Mission not found")
			return
		}
		h.logger.Error().Err(err).Msg("Failed to get mission events")
		h.respondWithError(w, http.StatusInternalServerError, "Failed to get mission events")
		return
	}

	h.respondWithJSON(w, http.StatusOK, events)
}

// Helper functions

func (h *Handlers) parseMissionListFilter(r *http.Request) *ListMissionsFilter {
	filter := &ListMissionsFilter{
		Limit:  50,
		Offset: 0,
	}

	query := r.URL.Query()

	// Parse status
	if status := query.Get("status"); status != "" {
		missionStatus := MissionStatus(status)
		filter.Status = &missionStatus
	}

	// Parse priority
	if priorityStr := query.Get("priority"); priorityStr != "" {
		if priority, err := strconv.Atoi(priorityStr); err == nil {
			filter.Priority = &priority
		}
	}

	// Parse commander ID
	if commanderIDStr := query.Get("commander_id"); commanderIDStr != "" {
		if commanderID, err := uuid.Parse(commanderIDStr); err == nil {
			filter.CommanderID = &commanderID
		}
	}

	// Parse group ID
	if groupID := query.Get("group_id"); groupID != "" {
		filter.GroupID = &groupID
	}

	// Parse date filters
	if startDateStr := query.Get("start_date_after"); startDateStr != "" {
		if startDate, err := time.Parse(time.RFC3339, startDateStr); err == nil {
			filter.StartDate = &startDate
		}
	}

	if endDateStr := query.Get("end_date_before"); endDateStr != "" {
		if endDate, err := time.Parse(time.RFC3339, endDateStr); err == nil {
			filter.EndDate = &endDate
		}
	}

	// Parse pagination
	if limitStr := query.Get("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 && limit <= 1000 {
			filter.Limit = limit
		}
	}

	if offsetStr := query.Get("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil && offset >= 0 {
			filter.Offset = offset
		}
	}

	return filter
}

func (h *Handlers) parseTaskListFilter(r *http.Request) *ListTasksFilter {
	filter := &ListTasksFilter{
		Limit:  50,
		Offset: 0,
	}

	query := r.URL.Query()

	// Parse mission ID
	if missionIDStr := query.Get("mission_id"); missionIDStr != "" {
		if missionID, err := uuid.Parse(missionIDStr); err == nil {
			filter.MissionID = &missionID
		}
	}

	// Parse status
	if status := query.Get("status"); status != "" {
		taskStatus := TaskStatus(status)
		filter.Status = &taskStatus
	}

	// Parse priority
	if priorityStr := query.Get("priority"); priorityStr != "" {
		if priority, err := strconv.Atoi(priorityStr); err == nil {
			filter.Priority = &priority
		}
	}

	// Parse assigned to
	if assignedToStr := query.Get("assigned_to"); assignedToStr != "" {
		if assignedTo, err := uuid.Parse(assignedToStr); err == nil {
			filter.AssignedTo = &assignedTo
		}
	}

	// Parse due date
	if dueDateStr := query.Get("due_date_before"); dueDateStr != "" {
		if dueDate, err := time.Parse(time.RFC3339, dueDateStr); err == nil {
			filter.DueDate = &dueDate
		}
	}

	// Parse pagination
	if limitStr := query.Get("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 && limit <= 1000 {
			filter.Limit = limit
		}
	}

	if offsetStr := query.Get("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil && offset >= 0 {
			filter.Offset = offset
		}
	}

	return filter
}

func (h *Handlers) validateCreateMissionRequest(req *CreateMissionRequest) error {
	if req.Name == "" {
		return errors.New("mission name is required")
	}
	if len(req.Name) > 255 {
		return errors.New("mission name must not exceed 255 characters")
	}
	if req.Priority < 1 || req.Priority > 5 {
		return errors.New("mission priority must be between 1 and 5")
	}
	if req.StartDate != nil && req.EndDate != nil {
		if req.EndDate.Before(*req.StartDate) {
			return errors.New("end date must be after start date")
		}
	}
	return nil
}

func (h *Handlers) validateUpdateMissionRequest(req *UpdateMissionRequest) error {
	if req.Name != nil && *req.Name == "" {
		return errors.New("mission name cannot be empty")
	}
	if req.Name != nil && len(*req.Name) > 255 {
		return errors.New("mission name must not exceed 255 characters")
	}
	if req.Priority != nil && (*req.Priority < 1 || *req.Priority > 5) {
		return errors.New("mission priority must be between 1 and 5")
	}
	if req.StartDate != nil && req.EndDate != nil {
		if req.EndDate.Before(*req.StartDate) {
			return errors.New("end date must be after start date")
		}
	}
	return nil
}

func (h *Handlers) respondWithError(w http.ResponseWriter, code int, message string) {
	h.respondWithJSON(w, code, map[string]string{"error": message})
}

func (h *Handlers) respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	if payload == nil && code == http.StatusNoContent {
		w.WriteHeader(code)
		return
	}

	response, err := json.Marshal(payload)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to marshal JSON response")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Internal server error"}`))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}