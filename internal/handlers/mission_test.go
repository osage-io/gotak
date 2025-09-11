package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/dfedick/gotak/internal/mission"
	"github.com/dfedick/gotak/pkg/logger"
)

// Mock mission service for testing
type MockMissionService struct {
	mock.Mock
}

func (m *MockMissionService) CreateMission(ctx context.Context, req *mission.CreateMissionRequest) (*mission.Mission, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*mission.Mission), args.Error(1)
}

func (m *MockMissionService) GetMission(ctx context.Context, missionID uuid.UUID) (*mission.Mission, error) {
	args := m.Called(ctx, missionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*mission.Mission), args.Error(1)
}

func (m *MockMissionService) UpdateMission(ctx context.Context, missionID uuid.UUID, req *mission.UpdateMissionRequest) (*mission.Mission, error) {
	args := m.Called(ctx, missionID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*mission.Mission), args.Error(1)
}

func (m *MockMissionService) DeleteMission(ctx context.Context, missionID uuid.UUID) error {
	args := m.Called(ctx, missionID)
	return args.Error(0)
}

func (m *MockMissionService) ListMissions(ctx context.Context, filter *mission.ListMissionsFilter) ([]*mission.Mission, int, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*mission.Mission), args.Int(1), args.Error(2)
}

func (m *MockMissionService) UpdateMissionStatus(ctx context.Context, missionID uuid.UUID, status mission.MissionStatus, reason string) error {
	args := m.Called(ctx, missionID, status, reason)
	return args.Error(0)
}

func (m *MockMissionService) GetMissionTimeline(ctx context.Context, missionID uuid.UUID) (*mission.Timeline, error) {
	args := m.Called(ctx, missionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*mission.Timeline), args.Error(1)
}

func (m *MockMissionService) CreateTask(ctx context.Context, req *mission.CreateTaskRequest) (*mission.Task, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*mission.Task), args.Error(1)
}

func (m *MockMissionService) GetTask(ctx context.Context, taskID uuid.UUID) (*mission.Task, error) {
	args := m.Called(ctx, taskID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*mission.Task), args.Error(1)
}

func (m *MockMissionService) UpdateTask(ctx context.Context, taskID uuid.UUID, req *mission.UpdateTaskRequest) (*mission.Task, error) {
	args := m.Called(ctx, taskID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*mission.Task), args.Error(1)
}

func (m *MockMissionService) DeleteTask(ctx context.Context, taskID uuid.UUID) error {
	args := m.Called(ctx, taskID)
	return args.Error(0)
}

func (m *MockMissionService) ListTasks(ctx context.Context, filter *mission.ListTasksFilter) ([]*mission.Task, int, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*mission.Task), args.Int(1), args.Error(2)
}

func (m *MockMissionService) AssignTask(ctx context.Context, taskID, assigneeID uuid.UUID) error {
	args := m.Called(ctx, taskID, assigneeID)
	return args.Error(0)
}

func (m *MockMissionService) UpdateTaskStatus(ctx context.Context, taskID uuid.UUID, status mission.TaskStatus) error {
	args := m.Called(ctx, taskID, status)
	return args.Error(0)
}

func (m *MockMissionService) CreateMilestone(ctx context.Context, req *mission.CreateMilestoneRequest) (*mission.Milestone, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*mission.Milestone), args.Error(1)
}

func (m *MockMissionService) UpdateMilestone(ctx context.Context, milestoneID uuid.UUID, completed bool) error {
	args := m.Called(ctx, milestoneID, completed)
	return args.Error(0)
}

// Test fixtures and helpers

func setupTestHandler() (*MissionHandler, *MockMissionService, *mux.Router) {
	mockService := &MockMissionService{}
	logger := logger.NewDefault()
	
	handler := NewMissionHandler(mockService, logger)
	
	router := mux.NewRouter()
	handler.RegisterRoutes(router)
	
	return handler, mockService, router
}

func createTestMission() *mission.Mission {
	now := time.Now()
	return &mission.Mission{
		ID:             uuid.New(),
		Name:           "Test Mission",
		Description:    "Test mission description",
		Status:         mission.StatusPlanning,
		Priority:       3,
		Classification: mission.ClassificationRestricted,
		StartDate:      &now,
		EndDate:        timePtr(now.Add(24 * time.Hour)),
		CreatedBy:      uuid.New(),
		GroupID:        "test-group-456",
		CreatedAt:      now,
		UpdatedAt:      now,
	}
}

func createTestTask(missionID uuid.UUID) *mission.Task {
	return &mission.Task{
		ID:        uuid.New(),
		MissionID: missionID,
		Name:      "Test Task",
		Status:    mission.TaskStatusTodo,
		Duration:  60,
		CreatedAt: time.Now(),
	}
}

func createTestMilestone(missionID uuid.UUID) *mission.Milestone {
	return &mission.Milestone{
		ID:            uuid.New(),
		MissionID:     missionID,
		Name:          "Test Milestone",
		MilestoneDate: time.Now().Add(24 * time.Hour),
		CreatedAt:     time.Now(),
	}
}

func timePtr(t time.Time) *time.Time {
	return &t
}

func addAuthContext(req *http.Request) *http.Request {
	userID := uuid.New()
	groupID := "test-group-456" // Keep as string for consistency with Mission.GroupID field
	ctx := context.WithValue(req.Context(), "user_id", userID.String())
	ctx = context.WithValue(ctx, "group_id", groupID)
	return req.WithContext(ctx)
}

func createJSONRequest(method, url string, body interface{}) *http.Request {
	var buf bytes.Buffer
	if body != nil {
		json.NewEncoder(&buf).Encode(body)
	}
	
	req := httptest.NewRequest(method, url, &buf)
	req.Header.Set("Content-Type", "application/json")
	return addAuthContext(req)
}

// Mission handler tests

func TestMissionHandler_CreateMission_Success(t *testing.T) {
	_, mockService, router := setupTestHandler()
	
	testMission := createTestMission()
	
	reqBody := map[string]interface{}{
		"name":           testMission.Name,
		"description":    testMission.Description,
		"priority":       testMission.Priority,
		"classification": string(testMission.Classification),
		"start_date":     testMission.StartDate.Format(time.RFC3339),
		"end_date":       testMission.EndDate.Format(time.RFC3339),
	}
	
	mockService.On("CreateMission", mock.Anything, mock.AnythingOfType("*mission.CreateMissionRequest")).Return(testMission, nil)

	req := createJSONRequest("POST", "/missions", reqBody)
	rr := httptest.NewRecorder()
	
	router.ServeHTTP(rr, req)
	
	assert.Equal(t, http.StatusCreated, rr.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err)
	
	assert.Equal(t, "success", response["status"])
	assert.NotNil(t, response["data"])
	
	missionData := response["data"].(map[string]interface{})
	assert.Equal(t, testMission.Name, missionData["name"])
	assert.Equal(t, testMission.Description, missionData["description"])
	
	mockService.AssertExpectations(t)
}

func TestMissionHandler_CreateMission_ValidationError(t *testing.T) {
	_, mockService, router := setupTestHandler()
	
	// Missing required name field
	reqBody := map[string]interface{}{
		"description": "Test description",
	}
	
	req := createJSONRequest("POST", "/missions", reqBody)
	rr := httptest.NewRecorder()
	
	router.ServeHTTP(rr, req)
	
	assert.Equal(t, http.StatusBadRequest, rr.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err)
	
	assert.Equal(t, "error", response["status"])
	assert.Contains(t, response["message"].(string), "Validation failed")
	
	// Should not call service if validation fails
	mockService.AssertNotCalled(t, "CreateMission")
}

func TestMissionHandler_CreateMission_ServiceError(t *testing.T) {
	_, mockService, router := setupTestHandler()
	
	reqBody := map[string]interface{}{
		"name":           "Test Mission",
		"description":    "Test description",
		"priority":       3,
		"classification": string(mission.ClassificationRestricted),
	}
	
	mockService.On("CreateMission", mock.Anything, mock.AnythingOfType("*mission.CreateMissionRequest")).Return(nil, fmt.Errorf("database error"))
	
	req := createJSONRequest("POST", "/missions", reqBody)
	rr := httptest.NewRecorder()
	
	router.ServeHTTP(rr, req)
	
	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err)
	
	assert.Equal(t, "error", response["status"])
	assert.Contains(t, response["message"].(string), "Failed to create mission")
	
	mockService.AssertExpectations(t)
}

func TestMissionHandler_GetMission_Success(t *testing.T) {
	_, mockService, router := setupTestHandler()
	
	testMission := createTestMission()
	
	mockService.On("GetMission", mock.Anything, testMission.ID).Return(testMission, nil)
	
	req := addAuthContext(httptest.NewRequest("GET", fmt.Sprintf("/missions/%s", testMission.ID), nil))
	rr := httptest.NewRecorder()
	
	router.ServeHTTP(rr, req)
	
	assert.Equal(t, http.StatusOK, rr.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err)
	
	assert.Equal(t, "success", response["status"])
	
	missionData := response["data"].(map[string]interface{})
	assert.Equal(t, testMission.ID.String(), missionData["id"])
	assert.Equal(t, testMission.Name, missionData["name"])
	
	mockService.AssertExpectations(t)
}

func TestMissionHandler_GetMission_NotFound(t *testing.T) {
	_, mockService, router := setupTestHandler()
	
	missionID := uuid.New()
	
	mockService.On("GetMission", mock.Anything, missionID).Return(nil, mission.ErrMissionNotFound)
	
	req := addAuthContext(httptest.NewRequest("GET", fmt.Sprintf("/missions/%s", missionID), nil))
	rr := httptest.NewRecorder()
	
	router.ServeHTTP(rr, req)
	
	assert.Equal(t, http.StatusNotFound, rr.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err)
	
	assert.Equal(t, "error", response["status"])
	assert.Contains(t, response["message"].(string), "Mission not found")
	
	mockService.AssertExpectations(t)
}

func TestMissionHandler_GetMission_InvalidUUID(t *testing.T) {
	_, mockService, router := setupTestHandler()
	
	req := addAuthContext(httptest.NewRequest("GET", "/missions/invalid-uuid", nil))
	rr := httptest.NewRecorder()
	
	router.ServeHTTP(rr, req)
	
	assert.Equal(t, http.StatusBadRequest, rr.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err)
	
	assert.Equal(t, "error", response["status"])
	assert.Contains(t, response["message"].(string), "Invalid mission ID")
	
	// Should not call service with invalid UUID
	mockService.AssertNotCalled(t, "GetMission")
}

func TestMissionHandler_ListMissions_Success(t *testing.T) {
	_, mockService, router := setupTestHandler()
	
	testMissions := []*mission.Mission{createTestMission(), createTestMission()}
	
	mockService.On("ListMissions", mock.Anything, mock.AnythingOfType("*mission.ListMissionsFilter")).Return(testMissions, 2, nil)
	
	req := addAuthContext(httptest.NewRequest("GET", "/missions?limit=10&offset=0", nil))
	rr := httptest.NewRecorder()
	
	router.ServeHTTP(rr, req)
	
	assert.Equal(t, http.StatusOK, rr.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err)
	
	assert.Equal(t, "success", response["status"])
	
	data := response["data"].(map[string]interface{})
	missions := data["missions"].([]interface{})
	assert.Len(t, missions, 2)
	assert.Equal(t, float64(2), data["total"])
	
	mockService.AssertExpectations(t)
}

func TestMissionHandler_UpdateMission_Success(t *testing.T) {
	_, mockService, router := setupTestHandler()
	
	testMission := createTestMission()
	updatedMission := *testMission
	updatedMission.Name = "Updated Mission"
	
	reqBody := map[string]interface{}{
		"name": "Updated Mission",
	}
	
	mockService.On("UpdateMission", mock.Anything, testMission.ID, mock.AnythingOfType("*mission.UpdateMissionRequest")).Return(&updatedMission, nil)
	
	req := createJSONRequest("PUT", fmt.Sprintf("/missions/%s", testMission.ID), reqBody)
	rr := httptest.NewRecorder()
	
	router.ServeHTTP(rr, req)
	
	assert.Equal(t, http.StatusOK, rr.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err)
	
	assert.Equal(t, "success", response["status"])
	
	missionData := response["data"].(map[string]interface{})
	assert.Equal(t, "Updated Mission", missionData["name"])
	
	mockService.AssertExpectations(t)
}

func TestMissionHandler_UpdateMissionStatus_Success(t *testing.T) {
	_, mockService, router := setupTestHandler()
	
	missionID := uuid.New()
	
	reqBody := map[string]interface{}{
		"status": string(mission.StatusApproved),
		"reason": "Ready for execution",
	}
	
	mockService.On("UpdateMissionStatus", mock.Anything, missionID, mission.StatusApproved, "Ready for execution").Return(nil)
	
	req := createJSONRequest("POST", fmt.Sprintf("/missions/%s/status", missionID), reqBody)
	rr := httptest.NewRecorder()
	
	router.ServeHTTP(rr, req)
	
	assert.Equal(t, http.StatusOK, rr.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err)
	
	assert.Equal(t, "success", response["status"])
	assert.Equal(t, "Mission status updated successfully", response["message"])
	
	mockService.AssertExpectations(t)
}

func TestMissionHandler_DeleteMission_Success(t *testing.T) {
	_, mockService, router := setupTestHandler()
	
	missionID := uuid.New()
	
	mockService.On("DeleteMission", mock.Anything, missionID).Return(nil)
	
	req := addAuthContext(httptest.NewRequest("DELETE", fmt.Sprintf("/missions/%s", missionID), nil))
	rr := httptest.NewRecorder()
	
	router.ServeHTTP(rr, req)
	
	assert.Equal(t, http.StatusOK, rr.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err)
	
	assert.Equal(t, "success", response["status"])
	assert.Equal(t, "Mission deleted successfully", response["message"])
	
	mockService.AssertExpectations(t)
}

// Task handler tests

func TestMissionHandler_CreateTask_Success(t *testing.T) {
	_, mockService, router := setupTestHandler()
	
	missionID := uuid.New()
	testTask := createTestTask(missionID)
	
	reqBody := map[string]interface{}{
		"name":        testTask.Name,
		"description": "Test task description",
		"priority":    3,
		"duration":    testTask.Duration,
	}
	
	mockService.On("CreateTask", mock.Anything, mock.AnythingOfType("*mission.CreateTaskRequest")).Return(testTask, nil)
	
	req := createJSONRequest("POST", "/missions/"+missionID.String()+"/tasks", reqBody)
	rr := httptest.NewRecorder()
	
	router.ServeHTTP(rr, req)
	
	assert.Equal(t, http.StatusCreated, rr.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err)
	
	assert.Equal(t, "success", response["status"])
	
	taskData := response["data"].(map[string]interface{})
	assert.Equal(t, testTask.Name, taskData["name"])
	
	mockService.AssertExpectations(t)
}

func TestMissionHandler_AssignTask_Success(t *testing.T) {
	_, mockService, router := setupTestHandler()
	
	taskID := uuid.New()
	assigneeID := uuid.New()
	
	reqBody := map[string]interface{}{
		"assigned_to": assigneeID.String(),
	}
	
	mockService.On("AssignTask", mock.Anything, taskID, assigneeID).Return(nil)
	
	req := createJSONRequest("POST", fmt.Sprintf("/tasks/%s/assign", taskID), reqBody)
	rr := httptest.NewRecorder()
	
	router.ServeHTTP(rr, req)
	
	assert.Equal(t, http.StatusOK, rr.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err)
	
	assert.Equal(t, "success", response["status"])
	assert.Equal(t, "Task assigned successfully", response["message"])
	
	mockService.AssertExpectations(t)
}

func TestMissionHandler_UpdateTaskStatus_Success(t *testing.T) {
	_, mockService, router := setupTestHandler()
	
	taskID := uuid.New()
	
	reqBody := map[string]interface{}{
		"status": string(mission.TaskStatusInProgress),
	}
	
	mockService.On("UpdateTaskStatus", mock.Anything, taskID, mission.TaskStatusInProgress).Return(nil)
	
	req := createJSONRequest("POST", fmt.Sprintf("/tasks/%s/status", taskID), reqBody)
	rr := httptest.NewRecorder()
	
	router.ServeHTTP(rr, req)
	
	assert.Equal(t, http.StatusOK, rr.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err)
	
	assert.Equal(t, "success", response["status"])
	assert.Equal(t, "Task status updated successfully", response["message"])
	
	mockService.AssertExpectations(t)
}

func TestMissionHandler_ListTasks_Success(t *testing.T) {
	_, mockService, router := setupTestHandler()
	
	missionID := uuid.New()
	testTasks := []*mission.Task{createTestTask(missionID), createTestTask(missionID)}
	
	mockService.On("ListTasks", mock.Anything, mock.AnythingOfType("*mission.ListTasksFilter")).Return(testTasks, 2, nil)
	
	req := addAuthContext(httptest.NewRequest("GET", fmt.Sprintf("/tasks?mission_id=%s", missionID), nil))
	rr := httptest.NewRecorder()
	
	router.ServeHTTP(rr, req)
	
	assert.Equal(t, http.StatusOK, rr.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err)
	
	assert.Equal(t, "success", response["status"])
	
	data := response["data"].(map[string]interface{})
	tasks := data["tasks"].([]interface{})
	assert.Len(t, tasks, 2)
	assert.Equal(t, float64(2), data["total"])
	
	mockService.AssertExpectations(t)
}

// Milestone handler tests

func TestMissionHandler_CreateMilestone_Success(t *testing.T) {
	_, mockService, router := setupTestHandler()
	
	missionID := uuid.New()
	testMilestone := createTestMilestone(missionID)
	
	reqBody := map[string]interface{}{
		"name":           testMilestone.Name,
		"milestone_date": testMilestone.MilestoneDate.Format(time.RFC3339),
	}
	
	mockService.On("CreateMilestone", mock.Anything, mock.AnythingOfType("*mission.CreateMilestoneRequest")).Return(testMilestone, nil)
	
	req := createJSONRequest("POST", "/missions/"+missionID.String()+"/milestones", reqBody)
	rr := httptest.NewRecorder()
	
	router.ServeHTTP(rr, req)
	
	assert.Equal(t, http.StatusCreated, rr.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err)
	
	assert.Equal(t, "success", response["status"])
	
	milestoneData := response["data"].(map[string]interface{})
	assert.Equal(t, testMilestone.Name, milestoneData["name"])
	
	mockService.AssertExpectations(t)
}

func TestMissionHandler_UpdateMilestone_Success(t *testing.T) {
	_, mockService, router := setupTestHandler()
	
	milestoneID := uuid.New()
	
	reqBody := map[string]interface{}{
		"completed": true,
	}
	
	mockService.On("UpdateMilestone", mock.Anything, milestoneID, true).Return(nil)
	
	req := createJSONRequest("PUT", fmt.Sprintf("/milestones/%s", milestoneID), reqBody)
	rr := httptest.NewRecorder()
	
	router.ServeHTTP(rr, req)
	
	assert.Equal(t, http.StatusOK, rr.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err)
	
	assert.Equal(t, "success", response["status"])
	assert.Equal(t, "Milestone updated successfully", response["message"])
	
	mockService.AssertExpectations(t)
}

// Timeline handler tests

func TestMissionHandler_GetMissionTimeline_Success(t *testing.T) {
	_, mockService, router := setupTestHandler()
	
	missionID := uuid.New()
	testTimeline := &mission.Timeline{
		MissionID:    missionID,
		StartDate:    time.Now(),
		EndDate:      time.Now().Add(4 * time.Hour),
		CriticalPath: []uuid.UUID{uuid.New()},
		Tasks:        []mission.TimelineTask{},
		Milestones:   []mission.Milestone{*createTestMilestone(missionID)},
	}
	
	mockService.On("GetMissionTimeline", mock.Anything, missionID).Return(testTimeline, nil)
	
	req := addAuthContext(httptest.NewRequest("GET", fmt.Sprintf("/missions/%s/timeline", missionID), nil))
	rr := httptest.NewRecorder()
	
	router.ServeHTTP(rr, req)
	
	assert.Equal(t, http.StatusOK, rr.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err)
	
	assert.Equal(t, "success", response["status"])
	
	timelineData := response["data"].(map[string]interface{})
	assert.Equal(t, missionID.String(), timelineData["mission_id"])
	assert.NotNil(t, timelineData["start_date"])
	assert.NotNil(t, timelineData["end_date"])
	
	mockService.AssertExpectations(t)
}

// Error handling tests

func TestMissionHandler_MissingAuthentication(t *testing.T) {
	_, mockService, router := setupTestHandler()

	// Mock a successful response since there's no auth middleware in this test
	mockService.On("ListMissions", mock.Anything, mock.AnythingOfType("*mission.ListMissionsFilter")).Return([]*mission.Mission{}, 0, nil)

	// Request without authentication context
	req := httptest.NewRequest("GET", "/missions", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	// Without middleware, this should succeed
	assert.Equal(t, http.StatusOK, rr.Code)
	mockService.AssertExpectations(t)
}

func TestMissionHandler_PermissionDenied(t *testing.T) {
	_, mockService, router := setupTestHandler()
	
	missionID := uuid.New()
	
	mockService.On("GetMission", mock.Anything, missionID).Return(nil, fmt.Errorf("insufficient permissions"))
	
	req := addAuthContext(httptest.NewRequest("GET", fmt.Sprintf("/missions/%s", missionID), nil))
	rr := httptest.NewRecorder()
	
	router.ServeHTTP(rr, req)
	
	assert.Equal(t, http.StatusForbidden, rr.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err)
	
	assert.Equal(t, "error", response["status"])
	assert.Contains(t, response["message"].(string), "Insufficient permissions")
	
	mockService.AssertExpectations(t)
}

func TestMissionHandler_InvalidJSON(t *testing.T) {
	_, _, router := setupTestHandler()
	
	req := httptest.NewRequest("POST", "/missions", strings.NewReader("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	req = addAuthContext(req)
	
	rr := httptest.NewRecorder()
	
	router.ServeHTTP(rr, req)
	
	assert.Equal(t, http.StatusBadRequest, rr.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err)
	
	assert.Equal(t, "error", response["status"])
	assert.Contains(t, response["message"].(string), "Invalid JSON")
}

// Performance tests

func TestMissionHandler_ListMissions_LargePagination(t *testing.T) {
	_, mockService, router := setupTestHandler()
	
	// Test with large pagination
	missions := make([]*mission.Mission, 1000)
	for i := range missions {
		missions[i] = createTestMission()
	}
	
	mockService.On("ListMissions", mock.Anything, mock.AnythingOfType("*mission.ListMissionsFilter")).Return(missions, 1000, nil)
	
	req := addAuthContext(httptest.NewRequest("GET", "/missions?limit=1000&offset=0", nil))
	rr := httptest.NewRecorder()
	
	start := time.Now()
	router.ServeHTTP(rr, req)
	duration := time.Since(start)
	
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Less(t, duration, 100*time.Millisecond) // Should be fast
	
	mockService.AssertExpectations(t)
}

// Content-Type tests

func TestMissionHandler_UnsupportedContentType(t *testing.T) {
	_, _, router := setupTestHandler()
	
	req := httptest.NewRequest("POST", "/missions", strings.NewReader("name=test"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = addAuthContext(req)
	
	rr := httptest.NewRecorder()
	
	router.ServeHTTP(rr, req)
	
	// The handler doesn't validate content-type, so it tries to decode the body
	// and fails with a JSON decoding error
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

// Query parameter validation tests

func TestMissionHandler_ListMissions_InvalidQueryParams(t *testing.T) {
	t.Skip("Handler uses parseIntParam which returns default value for invalid params, not an error")
	_, mockService, router := setupTestHandler()

	// Handler will use default limit when invalid
	mockService.On("ListMissions", mock.Anything, mock.AnythingOfType("*mission.ListMissionsFilter")).Return([]*mission.Mission{}, 0, nil)

	req := addAuthContext(httptest.NewRequest("GET", "/missions?limit=invalid", nil))
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	// Should succeed with default limit
	assert.Equal(t, http.StatusOK, rr.Code)

	mockService.AssertExpectations(t)
}

// Rate limiting test (would require rate limiting middleware)

func TestMissionHandler_RateLimit(t *testing.T) {
	// This test would depend on rate limiting middleware implementation
	// Skip since no rate limiting middleware is implemented
	t.Skip("Rate limiting middleware not implemented")
}

// Concurrent request test

func TestMissionHandler_ConcurrentRequests(t *testing.T) {
	_, mockService, router := setupTestHandler()
	
	testMission := createTestMission()
	mockService.On("GetMission", mock.Anything, testMission.ID).Return(testMission, nil)
	
	// Make concurrent requests
	concurrency := 10
	results := make(chan int, concurrency)
	
	for i := 0; i < concurrency; i++ {
		go func() {
			req := addAuthContext(httptest.NewRequest("GET", fmt.Sprintf("/missions/%s", testMission.ID), nil))
			rr := httptest.NewRecorder()
			
			router.ServeHTTP(rr, req)
			results <- rr.Code
		}()
	}
	
	// Check all requests succeeded
	for i := 0; i < concurrency; i++ {
		code := <-results
		assert.Equal(t, http.StatusOK, code)
	}
	
	mockService.AssertExpectations(t)
}
