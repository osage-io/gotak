package events

import (
	"context"
	"encoding/json"
	"time"

	nats "github.com/nats-io/nats.go"
	"github.com/google/uuid"

	"github.com/dfedick/gotak/pkg/logger"
)

// Publisher defines the interface for event publishing
type Publisher interface {
	PublishMissionEvent(ctx context.Context, event *MissionEvent) error
	PublishTaskEvent(ctx context.Context, event *TaskEvent) error
	PublishMilestoneEvent(ctx context.Context, event *MilestoneEvent) error
	Close() error
}

// NATSPublisher implements the Publisher interface using NATS
type NATSPublisher struct {
	conn   *nats.Conn
	logger *logger.Logger
}

// NewNATSPublisher creates a new NATS event publisher
func NewNATSPublisher(natsURL string, logger *logger.Logger) (*NATSPublisher, error) {
	conn, err := nats.Connect(natsURL)
	if err != nil {
		return nil, err
	}

	return &NATSPublisher{
		conn:   conn,
		logger: logger,
	}, nil
}

// Close closes the NATS connection
func (p *NATSPublisher) Close() error {
	if p.conn != nil {
		p.conn.Close()
	}
	return nil
}

// Event types
const (
	// Mission events
	MissionCreatedEvent     = "mission.created"
	MissionUpdatedEvent     = "mission.updated"
	MissionDeletedEvent     = "mission.deleted"
	MissionStatusChanged    = "mission.status_changed"
	
	// Task events
	TaskCreatedEvent        = "task.created"
	TaskUpdatedEvent        = "task.updated"
	TaskDeletedEvent        = "task.deleted"
	TaskAssignedEvent       = "task.assigned"
	TaskStatusChangedEvent  = "task.status_changed"
	TaskCompletedEvent      = "task.completed"
	
	// Milestone events
	MilestoneCreatedEvent   = "milestone.created"
	MilestoneUpdatedEvent   = "milestone.updated"
	MilestoneCompletedEvent = "milestone.completed"
)

// NATS subjects
const (
	MissionSubject   = "mission.events"
	TaskSubject      = "task.events"
	MilestoneSubject = "milestone.events"
)

// BaseEvent contains common fields for all events
type BaseEvent struct {
	ID        uuid.UUID `json:"id"`
	Type      string    `json:"type"`
	UserID    string    `json:"user_id"`
	GroupID   string    `json:"group_id"`
	Timestamp time.Time `json:"timestamp"`
	Source    string    `json:"source"`
	Version   string    `json:"version"`
}

// MissionEvent represents mission-related events
type MissionEvent struct {
	BaseEvent
	MissionID uuid.UUID   `json:"mission_id"`
	Data      interface{} `json:"data,omitempty"`
	Changes   interface{} `json:"changes,omitempty"`
	Metadata  interface{} `json:"metadata,omitempty"`
}

// TaskEvent represents task-related events
type TaskEvent struct {
	BaseEvent
	TaskID    uuid.UUID   `json:"task_id"`
	MissionID uuid.UUID   `json:"mission_id"`
	Data      interface{} `json:"data,omitempty"`
	Changes   interface{} `json:"changes,omitempty"`
	Metadata  interface{} `json:"metadata,omitempty"`
}

// MilestoneEvent represents milestone-related events
type MilestoneEvent struct {
	BaseEvent
	MilestoneID uuid.UUID   `json:"milestone_id"`
	MissionID   uuid.UUID   `json:"mission_id"`
	Data        interface{} `json:"data,omitempty"`
	Changes     interface{} `json:"changes,omitempty"`
	Metadata    interface{} `json:"metadata,omitempty"`
}

// Specific event payloads

// MissionStatusChangePayload represents mission status change details
type MissionStatusChangePayload struct {
	OldStatus string `json:"old_status"`
	NewStatus string `json:"new_status"`
	Reason    string `json:"reason"`
}

// TaskAssignmentPayload represents task assignment details
type TaskAssignmentPayload struct {
	AssignedTo   string `json:"assigned_to"`
	AssignedBy   string `json:"assigned_by"`
	PreviousUser string `json:"previous_user,omitempty"`
}

// TaskStatusChangePayload represents task status change details
type TaskStatusChangePayload struct {
	OldStatus string `json:"old_status"`
	NewStatus string `json:"new_status"`
}

// MilestoneCompletionPayload represents milestone completion details
type MilestoneCompletionPayload struct {
	Completed   bool      `json:"completed"`
	CompletedAt time.Time `json:"completed_at,omitempty"`
}

// PublishMissionEvent publishes a mission-related event
func (p *NATSPublisher) PublishMissionEvent(ctx context.Context, event *MissionEvent) error {
	// Set common fields if not already set
	if event.ID == uuid.Nil {
		event.ID = uuid.New()
	}
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}
	if event.Source == "" {
		event.Source = "gotak-mission-service"
	}
	if event.Version == "" {
		event.Version = "1.0"
	}

	// Serialize event to JSON
	data, err := json.Marshal(event)
	if err != nil {
		p.logger.Error().Err(err).Str("event_type", event.Type).Msg("Failed to marshal mission event")
		return err
	}

	// Publish to NATS
	err = p.conn.Publish(MissionSubject, data)
	if err != nil {
		p.logger.Error().Err(err).Str("event_type", event.Type).Msg("Failed to publish mission event")
		return err
	}

	p.logger.Debug().
		Str("event_id", event.ID.String()).
		Str("event_type", event.Type).
		Str("mission_id", event.MissionID.String()).
		Str("user_id", event.UserID).
		Msg("Published mission event")

	return nil
}

// PublishTaskEvent publishes a task-related event
func (p *NATSPublisher) PublishTaskEvent(ctx context.Context, event *TaskEvent) error {
	// Set common fields if not already set
	if event.ID == uuid.Nil {
		event.ID = uuid.New()
	}
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}
	if event.Source == "" {
		event.Source = "gotak-mission-service"
	}
	if event.Version == "" {
		event.Version = "1.0"
	}

	// Serialize event to JSON
	data, err := json.Marshal(event)
	if err != nil {
		p.logger.Error().Err(err).Str("event_type", event.Type).Msg("Failed to marshal task event")
		return err
	}

	// Publish to NATS
	err = p.conn.Publish(TaskSubject, data)
	if err != nil {
		p.logger.Error().Err(err).Str("event_type", event.Type).Msg("Failed to publish task event")
		return err
	}

	p.logger.Debug().
		Str("event_id", event.ID.String()).
		Str("event_type", event.Type).
		Str("task_id", event.TaskID.String()).
		Str("mission_id", event.MissionID.String()).
		Str("user_id", event.UserID).
		Msg("Published task event")

	return nil
}

// PublishMilestoneEvent publishes a milestone-related event
func (p *NATSPublisher) PublishMilestoneEvent(ctx context.Context, event *MilestoneEvent) error {
	// Set common fields if not already set
	if event.ID == uuid.Nil {
		event.ID = uuid.New()
	}
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}
	if event.Source == "" {
		event.Source = "gotak-mission-service"
	}
	if event.Version == "" {
		event.Version = "1.0"
	}

	// Serialize event to JSON
	data, err := json.Marshal(event)
	if err != nil {
		p.logger.Error().Err(err).Str("event_type", event.Type).Msg("Failed to marshal milestone event")
		return err
	}

	// Publish to NATS
	err = p.conn.Publish(MilestoneSubject, data)
	if err != nil {
		p.logger.Error().Err(err).Str("event_type", event.Type).Msg("Failed to publish milestone event")
		return err
	}

	p.logger.Debug().
		Str("event_id", event.ID.String()).
		Str("event_type", event.Type).
		Str("milestone_id", event.MilestoneID.String()).
		Str("mission_id", event.MissionID.String()).
		Str("user_id", event.UserID).
		Msg("Published milestone event")

	return nil
}

// Event builder functions for common use cases

// NewMissionCreatedEvent creates a mission created event
func NewMissionCreatedEvent(userID, groupID string, missionID uuid.UUID, missionData interface{}) *MissionEvent {
	return &MissionEvent{
		BaseEvent: BaseEvent{
			Type:    MissionCreatedEvent,
			UserID:  userID,
			GroupID: groupID,
		},
		MissionID: missionID,
		Data:      missionData,
	}
}

// NewMissionUpdatedEvent creates a mission updated event
func NewMissionUpdatedEvent(userID, groupID string, missionID uuid.UUID, changes interface{}) *MissionEvent {
	return &MissionEvent{
		BaseEvent: BaseEvent{
			Type:    MissionUpdatedEvent,
			UserID:  userID,
			GroupID: groupID,
		},
		MissionID: missionID,
		Changes:   changes,
	}
}

// NewMissionStatusChangedEvent creates a mission status changed event
func NewMissionStatusChangedEvent(userID, groupID string, missionID uuid.UUID, oldStatus, newStatus, reason string) *MissionEvent {
	payload := MissionStatusChangePayload{
		OldStatus: oldStatus,
		NewStatus: newStatus,
		Reason:    reason,
	}
	
	return &MissionEvent{
		BaseEvent: BaseEvent{
			Type:    MissionStatusChanged,
			UserID:  userID,
			GroupID: groupID,
		},
		MissionID: missionID,
		Data:      payload,
	}
}

// NewTaskCreatedEvent creates a task created event
func NewTaskCreatedEvent(userID, groupID string, taskID, missionID uuid.UUID, taskData interface{}) *TaskEvent {
	return &TaskEvent{
		BaseEvent: BaseEvent{
			Type:    TaskCreatedEvent,
			UserID:  userID,
			GroupID: groupID,
		},
		TaskID:    taskID,
		MissionID: missionID,
		Data:      taskData,
	}
}

// NewTaskAssignedEvent creates a task assigned event
func NewTaskAssignedEvent(userID, groupID string, taskID, missionID uuid.UUID, assignedTo, assignedBy string) *TaskEvent {
	payload := TaskAssignmentPayload{
		AssignedTo: assignedTo,
		AssignedBy: assignedBy,
	}
	
	return &TaskEvent{
		BaseEvent: BaseEvent{
			Type:    TaskAssignedEvent,
			UserID:  userID,
			GroupID: groupID,
		},
		TaskID:    taskID,
		MissionID: missionID,
		Data:      payload,
	}
}

// NewTaskStatusChangedEvent creates a task status changed event
func NewTaskStatusChangedEvent(userID, groupID string, taskID, missionID uuid.UUID, oldStatus, newStatus string) *TaskEvent {
	payload := TaskStatusChangePayload{
		OldStatus: oldStatus,
		NewStatus: newStatus,
	}
	
	return &TaskEvent{
		BaseEvent: BaseEvent{
			Type:    TaskStatusChangedEvent,
			UserID:  userID,
			GroupID: groupID,
		},
		TaskID:    taskID,
		MissionID: missionID,
		Data:      payload,
	}
}

// NewMilestoneCreatedEvent creates a milestone created event
func NewMilestoneCreatedEvent(userID, groupID string, milestoneID, missionID uuid.UUID, milestoneData interface{}) *MilestoneEvent {
	return &MilestoneEvent{
		BaseEvent: BaseEvent{
			Type:    MilestoneCreatedEvent,
			UserID:  userID,
			GroupID: groupID,
		},
		MilestoneID: milestoneID,
		MissionID:   missionID,
		Data:        milestoneData,
	}
}

// NewMilestoneCompletedEvent creates a milestone completion event
func NewMilestoneCompletedEvent(userID, groupID string, milestoneID, missionID uuid.UUID, completed bool, completedAt *time.Time) *MilestoneEvent {
	payload := MilestoneCompletionPayload{
		Completed: completed,
	}
	if completedAt != nil {
		payload.CompletedAt = *completedAt
	}
	
	return &MilestoneEvent{
		BaseEvent: BaseEvent{
			Type:    MilestoneCompletedEvent,
			UserID:  userID,
			GroupID: groupID,
		},
		MilestoneID: milestoneID,
		MissionID:   missionID,
		Data:        payload,
	}
}

// MockPublisher for testing
type MockPublisher struct {
	Events []interface{}
	logger logger.Logger
}

// NewMockPublisher creates a mock publisher for testing
func NewMockPublisher(logger logger.Logger) *MockPublisher {
	return &MockPublisher{
		Events: make([]interface{}, 0),
		logger: logger,
	}
}

func (m *MockPublisher) PublishMissionEvent(ctx context.Context, event *MissionEvent) error {
	m.Events = append(m.Events, event)
	m.logger.Debug().Str("event_type", event.Type).Msg("Mock published mission event")
	return nil
}

func (m *MockPublisher) PublishTaskEvent(ctx context.Context, event *TaskEvent) error {
	m.Events = append(m.Events, event)
	m.logger.Debug().Str("event_type", event.Type).Msg("Mock published task event")
	return nil
}

func (m *MockPublisher) PublishMilestoneEvent(ctx context.Context, event *MilestoneEvent) error {
	m.Events = append(m.Events, event)
	m.logger.Debug().Str("event_type", event.Type).Msg("Mock published milestone event")
	return nil
}

func (m *MockPublisher) Close() error {
	return nil
}

// GetMissionEvents returns only mission events from the mock
func (m *MockPublisher) GetMissionEvents() []*MissionEvent {
	events := make([]*MissionEvent, 0)
	for _, event := range m.Events {
		if missionEvent, ok := event.(*MissionEvent); ok {
			events = append(events, missionEvent)
		}
	}
	return events
}

// GetTaskEvents returns only task events from the mock
func (m *MockPublisher) GetTaskEvents() []*TaskEvent {
	events := make([]*TaskEvent, 0)
	for _, event := range m.Events {
		if taskEvent, ok := event.(*TaskEvent); ok {
			events = append(events, taskEvent)
		}
	}
	return events
}

// GetMilestoneEvents returns only milestone events from the mock
func (m *MockPublisher) GetMilestoneEvents() []*MilestoneEvent {
	events := make([]*MilestoneEvent, 0)
	for _, event := range m.Events {
		if milestoneEvent, ok := event.(*MilestoneEvent); ok {
			events = append(events, milestoneEvent)
		}
	}
	return events
}

// Clear clears all events from the mock
func (m *MockPublisher) Clear() {
	m.Events = make([]interface{}, 0)
}
