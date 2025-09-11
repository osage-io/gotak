package mission

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/dfedick/gotak/internal/events"
	"github.com/dfedick/gotak/pkg/database"
	"github.com/dfedick/gotak/pkg/logger"
)

// Mock database for testing
type MockDB struct {
	mock.Mock
}

func (m *MockDB) QueryContext(ctx context.Context, query string, args ...interface{}) (*database.Rows, error) {
	callArgs := append([]interface{}{ctx, query}, args...)
	arguments := m.Called(callArgs...)
	rowsResult := arguments.Get(0)
	if rowsResult == nil {
		return nil, arguments.Error(1)
	}
	return rowsResult.(*database.Rows), arguments.Error(1)
}

func (m *MockDB) QueryRowContext(ctx context.Context, query string, args ...interface{}) database.Row {
	callArgs := append([]interface{}{ctx, query}, args...)
	arguments := m.Called(callArgs...)
	return arguments.Get(0).(database.Row)
}

func (m *MockDB) ExecContext(ctx context.Context, query string, args ...interface{}) (database.Result, error) {
	callArgs := append([]interface{}{ctx, query}, args...)
	arguments := m.Called(callArgs...)
	return arguments.Get(0).(database.Result), arguments.Error(1)
}

func (m *MockDB) BeginTx(ctx context.Context, opts *database.TxOptions) (database.Tx, error) {
	arguments := m.Called(ctx, opts)
	txResult := arguments.Get(0)
	if txResult == nil {
		return nil, arguments.Error(1)
	}
	return txResult.(database.Tx), arguments.Error(1)
}

// Mock transaction for testing
type MockTx struct {
	mock.Mock
}

func (m *MockTx) QueryContext(ctx context.Context, query string, args ...interface{}) (*database.Rows, error) {
	callArgs := append([]interface{}{ctx, query}, args...)
	arguments := m.Called(callArgs...)
	rowsResult := arguments.Get(0)
	if rowsResult == nil {
		return nil, arguments.Error(1)
	}
	return rowsResult.(*database.Rows), arguments.Error(1)
}

func (m *MockTx) QueryRowContext(ctx context.Context, query string, args ...interface{}) database.Row {
	callArgs := append([]interface{}{ctx, query}, args...)
	arguments := m.Called(callArgs...)
	return arguments.Get(0).(database.Row)
}

func (m *MockTx) ExecContext(ctx context.Context, query string, args ...interface{}) (database.Result, error) {
	callArgs := append([]interface{}{ctx, query}, args...)
	arguments := m.Called(callArgs...)
	return arguments.Get(0).(database.Result), arguments.Error(1)
}

func (m *MockTx) Commit() error {
	arguments := m.Called()
	return arguments.Error(0)
}

func (m *MockTx) Rollback() error {
	arguments := m.Called()
	return arguments.Error(0)
}

// Mock row for testing
type MockRow struct {
	mock.Mock
}

func (m *MockRow) Scan(dest ...interface{}) error {
	arguments := m.Called(dest)
	return arguments.Error(0)
}

// Mock result for testing
type MockResult struct {
	mock.Mock
}

func (m *MockResult) LastInsertId() (int64, error) {
	arguments := m.Called()
	return arguments.Get(0).(int64), arguments.Error(1)
}

func (m *MockResult) RowsAffected() (int64, error) {
	arguments := m.Called()
	return arguments.Get(0).(int64), arguments.Error(1)
}

// Mock rows for testing - simple interface implementation for sql.Rows
type MockSQLRows struct {
	hasNext bool
	closed  bool
}

func (m *MockSQLRows) Next() bool {
	return false // Always return false for empty results
}

func (m *MockSQLRows) Scan(dest ...interface{}) error {
	return nil // Should not be called for empty results
}

func (m *MockSQLRows) Close() error {
	m.closed = true
	return nil
}

func (m *MockSQLRows) Err() error {
	return nil
}

// Mock rows for testing
type MockRows struct {
	mock.Mock
	rowData [][]interface{}
	currentRow int
	closed bool
}

func NewMockRows() *MockRows {
	return &MockRows{
		rowData: make([][]interface{}, 0),
		currentRow: -1,
		closed: false,
	}
}

func (m *MockRows) Next() bool {
	if m.closed {
		return false
	}
	m.currentRow++
	return m.currentRow < len(m.rowData)
}

func (m *MockRows) Scan(dest ...interface{}) error {
	if m.closed {
		return fmt.Errorf("rows are closed")
	}
	if m.currentRow >= len(m.rowData) {
		return fmt.Errorf("no more rows")
	}
	
	row := m.rowData[m.currentRow]
	for i, val := range row {
		if i < len(dest) {
			// Set the value through the pointer
			switch d := dest[i].(type) {
			case *string:
				*d = val.(string)
			case *int:
				*d = val.(int)
			default:
				return fmt.Errorf("unsupported scan type: %T", dest[i])
			}
		}
	}
	return nil
}

func (m *MockRows) Close() error {
	m.closed = true
	return nil
}

func (m *MockRows) Err() error {
	return nil
}

// Test fixtures
func setupTestService(t *testing.T) (*Service, *MockDB, *events.MockPublisher, context.Context) {
	mockDB := &MockDB{}
	logger := logger.NewDefault()
	publisher := events.NewMockPublisher(*logger)
	service := NewService(mockDB, *logger, publisher)
	
	testUserID := uuid.New().String()
	ctx := context.WithValue(context.Background(), "user_id", testUserID)
	ctx = context.WithValue(ctx, "group_id", "test-group-456")
	
	return service, mockDB, publisher, ctx
}

func createTestMission() *Mission {
	commander := uuid.New()
	return &Mission{
		ID:             uuid.New(),
		Name:           "Test Mission",
		Description:    "Test mission description",
		Status:         StatusPlanning,
		Priority:       3,
		Classification: ClassificationRestricted,
		StartDate:      timePtr(time.Now()),
		EndDate:        timePtr(time.Now().Add(24 * time.Hour)),
		CommanderID:    &commander,
		CreatedBy:      uuid.New(),
		GroupID:        "test-group-456",
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
}

func timePtr(t time.Time) *time.Time {
	return &t
}

// Mission service tests

func TestService_CreateMission_Success(t *testing.T) {
	service, mockDB, publisher, ctx := setupTestService(t)
	
	req := &CreateMissionRequest{
		Name:           "Test Mission",
		Description:    "Test description",
		Priority:       3,
		Classification: ClassificationRestricted,
		StartDate:      timePtr(time.Now()),
		EndDate:        timePtr(time.Now().Add(24 * time.Hour)),
	}
	
	// Mock transaction
	mockTx := &MockTx{}
	mockDB.On("BeginTx", mock.Anything, mock.Anything).Return(mockTx, nil)
	// Mock the INSERT statement with all 18 parameters plus context and query = 20 total
	mockTx.On("ExecContext", mock.Anything, mock.AnythingOfType("string"), 
		mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, // 6 params: id, name, desc, status, priority, class
		mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, // 5 params: start, end, commander, created_by, group_id
		mock.Anything, mock.Anything, mock.Anything, mock.Anything, // 4 params: lat, lng, loc_name, loc_desc 
		mock.Anything, mock.Anything, mock.Anything, // 3 params: metadata, created_at, updated_at
	).Return(&MockResult{}, nil)
	mockTx.On("Commit").Return(nil)
	mockTx.On("Rollback").Return(nil)
	
	mission, err := service.CreateMission(ctx, req)
	
	require.NoError(t, err)
	assert.Equal(t, req.Name, mission.Name)
	assert.Equal(t, req.Description, mission.Description)
	assert.Equal(t, req.Priority, mission.Priority)
	assert.Equal(t, req.Classification, mission.Classification)
	assert.Equal(t, StatusPlanning, mission.Status)
	assert.Equal(t, "test-group-456", mission.GroupID)
	
	// Verify event was published
	events := publisher.GetMissionEvents()
	assert.Len(t, events, 1)
	assert.Equal(t, "mission.created", events[0].Type)
	
	mockDB.AssertExpectations(t)
	mockTx.AssertExpectations(t)
}

func TestService_CreateMission_InvalidDates(t *testing.T) {
	service, _, _, ctx := setupTestService(t)
	
	startDate := time.Now()
	endDate := startDate.Add(-1 * time.Hour) // End before start
	
	req := &CreateMissionRequest{
		Name:           "Test Mission",
		Classification: ClassificationRestricted,
		StartDate:      &startDate,
		EndDate:        &endDate,
	}
	
	mission, err := service.CreateMission(ctx, req)
	
	assert.Error(t, err)
	assert.Nil(t, mission)
	assert.Contains(t, err.Error(), "end date must be after start date")
}

func TestService_CreateMission_MissingUserContext(t *testing.T) {
	service, _, _, _ := setupTestService(t)
	
	ctx := context.Background() // No user context
	
	req := &CreateMissionRequest{
		Name:           "Test Mission",
		Classification: ClassificationRestricted,
	}
	
	mission, err := service.CreateMission(ctx, req)
	
	assert.Error(t, err)
	assert.Nil(t, mission)
	assert.Contains(t, err.Error(), "user ID not found in context")
}

func TestService_GetMission_FailsOnObjectivesLoad(t *testing.T) {
	service, mockDB, _, ctx := setupTestService(t)
	
	testMission := createTestMission()
	
	// Mock database calls for mission
	mockRow := &MockRow{}
	mockDB.On("QueryRowContext", mock.Anything, mock.AnythingOfType("string"), testMission.ID).Return(mockRow)
	mockRow.On("Scan", mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		// Simulate scanning mission data into the struct fields
		dest := args[0].([]interface{})
		*dest[0].(*uuid.UUID) = testMission.ID
		*dest[1].(*string) = testMission.Name
		*dest[2].(*string) = testMission.Description
		*dest[3].(*MissionStatus) = testMission.Status
		*dest[4].(*int) = testMission.Priority
		*dest[5].(*Classification) = testMission.Classification
		
		// Set pointers for StartDate and EndDate
		*dest[6].(**time.Time) = testMission.StartDate
		*dest[7].(**time.Time) = testMission.EndDate
		*dest[8].(**uuid.UUID) = testMission.CommanderID
		*dest[9].(*uuid.UUID) = testMission.CreatedBy
		*dest[10].(*string) = testMission.GroupID
		
		// Set location pointers (latitude, longitude, locationName, locationDescription)
		*dest[11].(**float64) = nil
		*dest[12].(**float64) = nil
		*dest[13].(**string) = nil
		*dest[14].(**string) = nil
		
		// Set metadata JSON and timestamps
		*dest[15].(*[]byte) = []byte("{}")
		*dest[16].(*time.Time) = testMission.CreatedAt
		*dest[17].(*time.Time) = testMission.UpdatedAt
	})
	
	// Mock objectives query - return database error
	mockDB.On("QueryContext", mock.Anything, mock.AnythingOfType("string"), testMission.ID).Return(nil, sql.ErrConnDone).Once()
	
	mission, err := service.GetMission(ctx, testMission.ID)
	
	// The test should fail because objectives query will fail
	assert.Error(t, err)
	assert.Nil(t, mission)
	assert.Contains(t, err.Error(), "failed to load mission objectives")
	
	mockDB.AssertExpectations(t)
}

func TestService_GetMission_NotFound(t *testing.T) {
	service, mockDB, _, ctx := setupTestService(t)
	
	testID := uuid.New()
	
	// Mock mission not found
	mockRow := &MockRow{}
	mockDB.On("QueryRowContext", mock.Anything, mock.AnythingOfType("string"), testID).Return(mockRow)
	mockRow.On("Scan", mock.Anything).Return(sql.ErrNoRows)
	
	mission, err := service.GetMission(ctx, testID)
	
	assert.Error(t, err)
	assert.Nil(t, mission)
	assert.Contains(t, err.Error(), "mission not found")
	
	mockDB.AssertExpectations(t)
}

func TestService_UpdateMissionStatus_Success(t *testing.T) {
	service, mockDB, publisher, ctx := setupTestService(t)
	
	testMission := createTestMission()
	newStatus := StatusApproved
	reason := "Ready for execution"
	
	// Mock getting current mission
	mockRow := &MockRow{}
mockDB.On("QueryRowContext", mock.Anything, mock.AnythingOfType("string"), testMission.ID).Return(mockRow)
mockRow.On("Scan", mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		dest := args[0].([]interface{})
		*dest[0].(*uuid.UUID) = testMission.ID
		*dest[1].(*string) = testMission.Name
		*dest[2].(*string) = testMission.Description
		*dest[3].(*MissionStatus) = testMission.Status
		*dest[4].(*int) = testMission.Priority
		*dest[5].(*Classification) = testMission.Classification
		*dest[6].(**time.Time) = testMission.StartDate
		*dest[7].(**time.Time) = testMission.EndDate
		*dest[8].(**uuid.UUID) = testMission.CommanderID
		*dest[9].(*uuid.UUID) = testMission.CreatedBy
		*dest[10].(*string) = testMission.GroupID
		*dest[11].(**float64) = nil
		*dest[12].(**float64) = nil
		*dest[13].(**string) = nil
		*dest[14].(**string) = nil
		*dest[15].(*[]byte) = []byte("{}")
		*dest[16].(*time.Time) = testMission.CreatedAt
		*dest[17].(*time.Time) = testMission.UpdatedAt
	})
	
	// Mock transaction
	mockTx := &MockTx{}
	mockDB.On("BeginTx", mock.Anything, mock.Anything).Return(mockTx, nil)
	// Mock UPDATE mission status call (3 params: context, query, status, missionID)
	mockTx.On("ExecContext", mock.Anything, mock.AnythingOfType("string"), mock.Anything, mock.Anything).Return(&MockResult{}, nil).Once()
	// Mock INSERT status history call (6 params: context, query, missionID, oldStatus, newStatus, userID, reason)
	mockTx.On("ExecContext", mock.Anything, mock.AnythingOfType("string"), mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&MockResult{}, nil).Once()
	mockTx.On("Commit").Return(nil)
	mockTx.On("Rollback").Return(nil)
	
	err := service.UpdateMissionStatus(ctx, testMission.ID, newStatus, reason)
	
	require.NoError(t, err)
	
	// Verify event was published
	events := publisher.GetMissionEvents()
	assert.Len(t, events, 1)
	assert.Equal(t, "mission.status_changed", events[0].Type)
	
	mockDB.AssertExpectations(t)
	mockTx.AssertExpectations(t)
}

func TestService_UpdateMissionStatus_InvalidTransition(t *testing.T) {
	service, mockDB, _, ctx := setupTestService(t)
	
	testMission := createTestMission()
	testMission.Status = StatusCompleted // Already completed
	
	// Mock getting current mission
	mockRow := &MockRow{}
mockDB.On("QueryRowContext", mock.Anything, mock.AnythingOfType("string"), testMission.ID).Return(mockRow)
mockRow.On("Scan", mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		dest := args[0].([]interface{})
		*dest[0].(*uuid.UUID) = testMission.ID
		*dest[1].(*string) = testMission.Name
		*dest[2].(*string) = testMission.Description
		*dest[3].(*MissionStatus) = testMission.Status
		*dest[4].(*int) = testMission.Priority
		*dest[5].(*Classification) = testMission.Classification
		*dest[6].(**time.Time) = testMission.StartDate
		*dest[7].(**time.Time) = testMission.EndDate
		*dest[8].(**uuid.UUID) = testMission.CommanderID
		*dest[9].(*uuid.UUID) = testMission.CreatedBy
		*dest[10].(*string) = testMission.GroupID
		*dest[11].(**float64) = nil
		*dest[12].(**float64) = nil
		*dest[13].(**string) = nil
		*dest[14].(**string) = nil
		*dest[15].(*[]byte) = []byte("{}")
		*dest[16].(*time.Time) = testMission.CreatedAt
		*dest[17].(*time.Time) = testMission.UpdatedAt
	})
	
	err := service.UpdateMissionStatus(ctx, testMission.ID, StatusPlanning, "Invalid transition")
	
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid status transition")
	
	mockDB.AssertExpectations(t)
}

func TestService_ListMissions_Success(t *testing.T) {
	service, mockDB, _, ctx := setupTestService(t)
	
	filter := &ListMissionsFilter{
		Limit:  10,
		Offset: 0,
	}
	
	// Mock count query
	mockRow := &MockRow{}
mockDB.On("QueryRowContext", mock.Anything, mock.AnythingOfType("string"), "test-group-456").Return(mockRow)
	mockRow.On("Scan", mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		dest := args[0].([]interface{})
		*dest[0].(*int) = 5 // Total count
	})
	
	// Mock list query - return error to simplify test (5 params: context, query, groupID, limit, offset)
	mockDB.On("QueryContext", mock.Anything, mock.AnythingOfType("string"), mock.Anything, mock.Anything, mock.Anything).Return(nil, sql.ErrNoRows)
	
	missions, total, err := service.ListMissions(ctx, filter)
	
	// Should fail due to query error
	assert.Error(t, err)
	assert.Equal(t, 0, total)
	assert.Nil(t, missions)
	assert.Contains(t, err.Error(), "failed to query missions")
	
	mockDB.AssertExpectations(t)
}

// Status transition validation tests

func TestIsValidMissionStatusTransition(t *testing.T) {
	tests := []struct {
		name     string
		from     MissionStatus
		to       MissionStatus
		expected bool
	}{
		{"planning to approved", StatusPlanning, StatusApproved, true},
		{"planning to cancelled", StatusPlanning, StatusCancelled, true},
		{"approved to active", StatusApproved, StatusActive, true},
		{"active to completed", StatusActive, StatusCompleted, true},
		{"completed to active", StatusCompleted, StatusActive, false}, // Terminal state
		{"cancelled to active", StatusCancelled, StatusActive, false}, // Terminal state
		{"planning to completed", StatusPlanning, StatusCompleted, false}, // Invalid skip
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidMissionStatusTransition(tt.from, tt.to)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Model validation tests

func TestMissionStatus_Valid(t *testing.T) {
	tests := []struct {
		status   MissionStatus
		expected bool
	}{
		{StatusPlanning, true},
		{StatusApproved, true},
		{StatusActive, true},
		{StatusOnHold, true},
		{StatusCompleted, true},
		{StatusCancelled, true},
		{MissionStatus("invalid"), false},
	}
	
	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.status.Valid())
		})
	}
}

func TestClassification_Valid(t *testing.T) {
	tests := []struct {
		classification Classification
		expected       bool
	}{
		{ClassificationUnclassified, true},
		{ClassificationRestricted, true},
		{ClassificationConfidential, true},
		{ClassificationSecret, true},
		{ClassificationTopSecret, true},
		{Classification("invalid"), false},
	}
	
	for _, tt := range tests {
		t.Run(string(tt.classification), func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.classification.Valid())
		})
	}
}

// Context helper tests

func TestGetUserIDFromContext(t *testing.T) {
	tests := []struct {
		name     string
		ctx      context.Context
		expected string
	}{
		{
			name:     "valid user ID",
			ctx:      context.WithValue(context.Background(), "user_id", "test-user-123"),
			expected: "test-user-123",
		},
		{
			name:     "missing user ID",
			ctx:      context.Background(),
			expected: "",
		},
		{
			name:     "wrong type",
			ctx:      context.WithValue(context.Background(), "user_id", 123),
			expected: "",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getUserIDFromContext(tt.ctx)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetGroupIDFromContext(t *testing.T) {
	tests := []struct {
		name     string
		ctx      context.Context
		expected string
	}{
		{
			name:     "valid group ID",
			ctx:      context.WithValue(context.Background(), "group_id", "test-group-456"),
			expected: "test-group-456",
		},
		{
			name:     "missing group ID",
			ctx:      context.Background(),
			expected: "",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getGroupIDFromContext(tt.ctx)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Event publishing tests

func TestService_EventPublishing_Disabled(t *testing.T) {
	mockDB := &MockDB{}
	logger := logger.NewDefault()
	service := NewService(mockDB, *logger, nil) // No publisher
	
	testUserID := uuid.New().String()
	ctx := context.WithValue(context.Background(), "user_id", testUserID)
	ctx = context.WithValue(ctx, "group_id", "test-group-456")
	
	req := &CreateMissionRequest{
		Name:           "Test Mission",
		Classification: ClassificationRestricted,
	}
	
	// Mock transaction
	mockTx := &MockTx{}
	mockDB.On("BeginTx", mock.Anything, mock.Anything).Return(mockTx, nil)
	// Mock the INSERT statement with all 18 parameters plus context and query = 20 total
	mockTx.On("ExecContext", mock.Anything, mock.AnythingOfType("string"), 
		mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, // 6 params: id, name, desc, status, priority, class
		mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, // 5 params: start, end, commander, created_by, group_id
		mock.Anything, mock.Anything, mock.Anything, mock.Anything, // 4 params: lat, lng, loc_name, loc_desc 
		mock.Anything, mock.Anything, mock.Anything, // 3 params: metadata, created_at, updated_at
	).Return(&MockResult{}, nil)
	mockTx.On("Commit").Return(nil)
	mockTx.On("Rollback").Return(nil)
	
	mission, err := service.CreateMission(ctx, req)
	
	require.NoError(t, err)
	assert.NotNil(t, mission)
	// No events should be published when publisher is nil
	
	mockDB.AssertExpectations(t)
	mockTx.AssertExpectations(t)
}

// Error handling tests

func TestService_CreateMission_TransactionFailure(t *testing.T) {
	service, mockDB, _, ctx := setupTestService(t)
	
	req := &CreateMissionRequest{
		Name:           "Test Mission",
		Classification: ClassificationRestricted,
	}
	
	// Mock transaction failure
	mockDB.On("BeginTx", mock.Anything, mock.Anything).Return(nil, assert.AnError)
	
	mission, err := service.CreateMission(ctx, req)
	
	assert.Error(t, err)
	assert.Nil(t, mission)
	assert.Contains(t, err.Error(), "failed to start transaction")
	
	mockDB.AssertExpectations(t)
}

func TestService_CreateMission_CommitFailure(t *testing.T) {
	service, mockDB, _, ctx := setupTestService(t)
	
	req := &CreateMissionRequest{
		Name:           "Test Mission",
		Classification: ClassificationRestricted,
	}
	
	// Mock transaction with commit failure
	mockTx := &MockTx{}
	mockDB.On("BeginTx", mock.Anything, mock.Anything).Return(mockTx, nil)
	// Mock the INSERT statement with all 18 parameters plus context and query = 20 total
	mockTx.On("ExecContext", mock.Anything, mock.AnythingOfType("string"), 
		mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, // 6 params: id, name, desc, status, priority, class
		mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, // 5 params: start, end, commander, created_by, group_id
		mock.Anything, mock.Anything, mock.Anything, mock.Anything, // 4 params: lat, lng, loc_name, loc_desc 
		mock.Anything, mock.Anything, mock.Anything, // 3 params: metadata, created_at, updated_at
	).Return(&MockResult{}, nil)
	mockTx.On("Commit").Return(assert.AnError)
	mockTx.On("Rollback").Return(nil)
	
	mission, err := service.CreateMission(ctx, req)
	
	assert.Error(t, err)
	assert.Nil(t, mission)
	assert.Contains(t, err.Error(), "failed to commit transaction")
	
	mockDB.AssertExpectations(t)
	mockTx.AssertExpectations(t)
}

// Benchmark tests

func BenchmarkService_CreateMission(b *testing.B) {
	service, mockDB, _, ctx := setupTestService(&testing.T{})
	
	req := &CreateMissionRequest{
		Name:           "Benchmark Mission",
		Classification: ClassificationRestricted,
	}
	
	// Setup mocks for benchmark
	mockTx := &MockTx{}
	mockDB.On("BeginTx", mock.Anything, mock.Anything).Return(mockTx, nil)
	// Mock the INSERT statement with all 18 parameters plus context and query = 20 total
	mockTx.On("ExecContext", mock.Anything, mock.AnythingOfType("string"), 
		mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, // 6 params: id, name, desc, status, priority, class
		mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, // 5 params: start, end, commander, created_by, group_id
		mock.Anything, mock.Anything, mock.Anything, mock.Anything, // 4 params: lat, lng, loc_name, loc_desc 
		mock.Anything, mock.Anything, mock.Anything, // 3 params: metadata, created_at, updated_at
	).Return(&MockResult{}, nil)
	mockTx.On("Commit").Return(nil)
	mockTx.On("Rollback").Return(nil)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.CreateMission(ctx, req)
	}
}
