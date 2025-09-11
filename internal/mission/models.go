package mission

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Common errors
var (
	ErrMissionNotFound   = errors.New("mission not found")
	ErrTaskNotFound      = errors.New("task not found")
	ErrMilestoneNotFound = errors.New("milestone not found")
	ErrPermissionDenied  = errors.New("permission denied")
	ErrInvalidInput      = errors.New("invalid input")
)

// Mission represents a military mission with all related data
type Mission struct {
	ID             uuid.UUID              `json:"id" db:"id"`
	Name           string                 `json:"name" db:"name"`
	Description    string                 `json:"description" db:"description"`
	Status         MissionStatus          `json:"status" db:"status"`
	Priority       int                    `json:"priority" db:"priority"`
	Classification Classification         `json:"classification" db:"classification"`
	StartDate      *time.Time             `json:"start_date" db:"start_date"`
	EndDate        *time.Time             `json:"end_date" db:"end_date"`
	CommanderID    *uuid.UUID             `json:"commander_id" db:"commander_id"`
	CreatedBy      uuid.UUID              `json:"created_by" db:"created_by"`
	GroupID        string                 `json:"group_id" db:"group_id"`
	Location       *Location              `json:"location,omitempty"`
	Objectives     []Objective            `json:"objectives,omitempty"`
	Tasks          []Task                 `json:"tasks,omitempty"`
	Resources      []ResourceRequest      `json:"resources,omitempty"`
	Milestones     []Milestone            `json:"milestones,omitempty"`
	Metadata       map[string]interface{} `json:"metadata" db:"metadata"`
	CreatedAt      time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at" db:"updated_at"`
}

// MissionStatus represents the current status of a mission
type MissionStatus string

const (
	StatusPlanning  MissionStatus = "planning"
	StatusApproved  MissionStatus = "approved"
	StatusActive    MissionStatus = "active"
	StatusOnHold    MissionStatus = "on_hold"
	StatusCompleted MissionStatus = "completed"
	StatusCancelled MissionStatus = "cancelled"
)

// Valid returns true if the mission status is valid
func (ms MissionStatus) Valid() bool {
	switch ms {
	case StatusPlanning, StatusApproved, StatusActive, StatusOnHold, StatusCompleted, StatusCancelled:
		return true
	}
	return false
}

// String returns the string representation of MissionStatus
func (ms MissionStatus) String() string {
	return string(ms)
}

// Scan implements the sql.Scanner interface for database scanning
func (ms *MissionStatus) Scan(value interface{}) error {
	if value == nil {
		*ms = StatusPlanning
		return nil
	}
	if str, ok := value.(string); ok {
		*ms = MissionStatus(str)
		return nil
	}
	return fmt.Errorf("cannot scan %T into MissionStatus", value)
}

// Value implements the driver.Valuer interface for database storage
func (ms MissionStatus) Value() (driver.Value, error) {
	return string(ms), nil
}

// Classification represents security classification levels
type Classification string

const (
	ClassificationUnclassified Classification = "UNCLASSIFIED"
	ClassificationRestricted   Classification = "RESTRICTED"
	ClassificationConfidential Classification = "CONFIDENTIAL"
	ClassificationSecret       Classification = "SECRET"
	ClassificationTopSecret    Classification = "TOP_SECRET"
)

// Valid returns true if the classification is valid
func (c Classification) Valid() bool {
	switch c {
	case ClassificationUnclassified, ClassificationRestricted, ClassificationConfidential, ClassificationSecret, ClassificationTopSecret:
		return true
	}
	return false
}

// String returns the string representation of Classification
func (c Classification) String() string {
	return string(c)
}

// Scan implements the sql.Scanner interface
func (c *Classification) Scan(value interface{}) error {
	if value == nil {
		*c = ClassificationRestricted
		return nil
	}
	if str, ok := value.(string); ok {
		*c = Classification(str)
		return nil
	}
	return fmt.Errorf("cannot scan %T into Classification", value)
}

// Value implements the driver.Valuer interface
func (c Classification) Value() (driver.Value, error) {
	return string(c), nil
}

// Location represents a geographical location
type Location struct {
	Latitude    float64 `json:"latitude" db:"latitude"`
	Longitude   float64 `json:"longitude" db:"longitude"`
	Name        string  `json:"name" db:"location_name"`
	Description string  `json:"description" db:"location_description"`
}

// Objective represents a mission objective
type Objective struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	MissionID   uuid.UUID  `json:"mission_id" db:"mission_id"`
	Description string     `json:"description" db:"description"`
	Priority    int        `json:"priority" db:"priority"`
	Completed   bool       `json:"completed" db:"completed"`
	CompletedAt *time.Time `json:"completed_at" db:"completed_at"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
}

// Task represents an individual task within a mission
type Task struct {
	ID             uuid.UUID       `json:"id" db:"id"`
	MissionID      uuid.UUID       `json:"mission_id" db:"mission_id"`
	Name           string          `json:"name" db:"name"`
	Description    string          `json:"description" db:"description"`
	Status         TaskStatus      `json:"status" db:"status"`
	Priority       int             `json:"priority" db:"priority"`
	AssignedTo     *uuid.UUID      `json:"assigned_to" db:"assigned_to"`
	Dependencies   []uuid.UUID     `json:"dependencies,omitempty"`  // Alias for DependsOn
	DependsOn      []uuid.UUID     `json:"depends_on,omitempty"`
	Duration       int             `json:"duration" db:"estimated_hours"`  // Duration in minutes
	EstimatedHours int             `json:"estimated_hours" db:"estimated_hours"`
	ActualHours    int             `json:"actual_hours" db:"actual_hours"`
	Resources      []TaskResource  `json:"resources,omitempty"`
	DueDate        *time.Time      `json:"due_date" db:"due_date"`
	CompletedAt    *time.Time      `json:"completed_at" db:"completed_at"`
	CreatedBy      uuid.UUID       `json:"created_by" db:"created_by"`
	CreatedAt      time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at" db:"updated_at"`
}

// TaskStatus represents the current status of a task
type TaskStatus string

const (
	TaskStatusTodo       TaskStatus = "todo"
	TaskStatusInProgress TaskStatus = "in_progress"
	TaskStatusReview     TaskStatus = "review"
	TaskStatusBlocked    TaskStatus = "blocked"
	TaskStatusCompleted  TaskStatus = "completed"
	TaskStatusCancelled  TaskStatus = "cancelled"
)

// Valid returns true if the task status is valid
func (ts TaskStatus) Valid() bool {
	switch ts {
	case TaskStatusTodo, TaskStatusInProgress, TaskStatusReview, TaskStatusBlocked, TaskStatusCompleted, TaskStatusCancelled:
		return true
	}
	return false
}

// String returns the string representation of TaskStatus
func (ts TaskStatus) String() string {
	return string(ts)
}

// Scan implements the sql.Scanner interface
func (ts *TaskStatus) Scan(value interface{}) error {
	if value == nil {
		*ts = TaskStatusTodo
		return nil
	}
	if str, ok := value.(string); ok {
		*ts = TaskStatus(str)
		return nil
	}
	return fmt.Errorf("cannot scan %T into TaskStatus", value)
}

// Value implements the driver.Valuer interface
func (ts TaskStatus) Value() (driver.Value, error) {
	return string(ts), nil
}

// Milestone represents a mission milestone
type Milestone struct {
	ID            uuid.UUID  `json:"id" db:"id"`
	MissionID     uuid.UUID  `json:"mission_id" db:"mission_id"`
	Name          string     `json:"name" db:"name"`
	Description   string     `json:"description" db:"description"`
	MilestoneDate time.Time  `json:"milestone_date" db:"milestone_date"`
	Completed     bool       `json:"completed" db:"completed"`
	CompletedAt   *time.Time `json:"completed_at" db:"completed_at"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
}

// ResourceRequest represents a resource allocation request
type ResourceRequest struct {
	ID           uuid.UUID     `json:"id" db:"id"`
	MissionID    uuid.UUID     `json:"mission_id" db:"mission_id"`
	ResourceType ResourceType  `json:"resource_type" db:"resource_type"`
	ResourceID   string        `json:"resource_id" db:"resource_id"`
	ResourceName string        `json:"resource_name" db:"resource_name"`
	Quantity     int           `json:"quantity" db:"quantity"`
	RequiredDate *time.Time    `json:"required_date" db:"required_date"`
	Status       ResourceStatus `json:"status" db:"status"`
	RequestedBy  uuid.UUID     `json:"requested_by" db:"requested_by"`
	ApprovedBy   *uuid.UUID    `json:"approved_by" db:"approved_by"`
	Notes        string        `json:"notes" db:"notes"`
	CreatedAt    time.Time     `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time     `json:"updated_at" db:"updated_at"`
}

// ResourceType represents the type of resource being requested
type ResourceType string

const (
	ResourceTypePersonnel ResourceType = "personnel"
	ResourceTypeEquipment ResourceType = "equipment"
	ResourceTypeVehicle   ResourceType = "vehicle"
	ResourceTypeMaterial  ResourceType = "material"
)

// Valid returns true if the resource type is valid
func (rt ResourceType) Valid() bool {
	switch rt {
	case ResourceTypePersonnel, ResourceTypeEquipment, ResourceTypeVehicle, ResourceTypeMaterial:
		return true
	}
	return false
}

// String returns the string representation of ResourceType
func (rt ResourceType) String() string {
	return string(rt)
}

// Scan implements the sql.Scanner interface
func (rt *ResourceType) Scan(value interface{}) error {
	if value == nil {
		*rt = ResourceTypePersonnel
		return nil
	}
	if str, ok := value.(string); ok {
		*rt = ResourceType(str)
		return nil
	}
	return fmt.Errorf("cannot scan %T into ResourceType", value)
}

// Value implements the driver.Valuer interface
func (rt ResourceType) Value() (driver.Value, error) {
	return string(rt), nil
}

// ResourceStatus represents the status of a resource request
type ResourceStatus string

const (
	ResourceStatusRequested ResourceStatus = "requested"
	ResourceStatusApproved  ResourceStatus = "approved"
	ResourceStatusAllocated ResourceStatus = "allocated"
	ResourceStatusDenied    ResourceStatus = "denied"
	ResourceStatusCancelled ResourceStatus = "cancelled"
)

// Valid returns true if the resource status is valid
func (rs ResourceStatus) Valid() bool {
	switch rs {
	case ResourceStatusRequested, ResourceStatusApproved, ResourceStatusAllocated, ResourceStatusDenied, ResourceStatusCancelled:
		return true
	}
	return false
}

// String returns the string representation of ResourceStatus
func (rs ResourceStatus) String() string {
	return string(rs)
}

// Scan implements the sql.Scanner interface
func (rs *ResourceStatus) Scan(value interface{}) error {
	if value == nil {
		*rs = ResourceStatusRequested
		return nil
	}
	if str, ok := value.(string); ok {
		*rs = ResourceStatus(str)
		return nil
	}
	return fmt.Errorf("cannot scan %T into ResourceStatus", value)
}

// Value implements the driver.Valuer interface
func (rs ResourceStatus) Value() (driver.Value, error) {
	return string(rs), nil
}

// MissionStatusHistory represents the history of mission status changes
type MissionStatusHistory struct {
	ID        uuid.UUID     `json:"id" db:"id"`
	MissionID uuid.UUID     `json:"mission_id" db:"mission_id"`
	OldStatus *MissionStatus `json:"old_status" db:"old_status"`
	NewStatus MissionStatus `json:"new_status" db:"new_status"`
	ChangedBy uuid.UUID     `json:"changed_by" db:"changed_by"`
	Reason    string        `json:"reason" db:"reason"`
	CreatedAt time.Time     `json:"created_at" db:"created_at"`
}

// TaskDependency represents a dependency relationship between tasks
type TaskDependency struct {
	TaskID         uuid.UUID `json:"task_id" db:"task_id"`
	DependsOnTaskID uuid.UUID `json:"depends_on_task_id" db:"depends_on_task_id"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
}

// Timeline represents a mission timeline with tasks and milestones
type Timeline struct {
	MissionID    uuid.UUID      `json:"mission_id"`
	StartDate    time.Time      `json:"start_date"`
	EndDate      time.Time      `json:"end_date"`
	Milestones   []Milestone    `json:"milestones"`
	Tasks        []TimelineTask `json:"tasks"`
	CriticalPath []uuid.UUID    `json:"critical_path"`
}

// TimelineTask represents a task in the context of a timeline
type TimelineTask struct {
	ID           uuid.UUID     `json:"id"`
	Name         string        `json:"name"`
	StartDate    time.Time     `json:"start_date"`
	EndDate      time.Time     `json:"end_date"`
	Duration     time.Duration `json:"duration"`
	Dependencies []uuid.UUID   `json:"dependencies"`
	AssignedTo   *string       `json:"assigned_to"`
	Status       TaskStatus    `json:"status"`
	CriticalPath bool          `json:"critical_path"`
	Progress     float64       `json:"progress"` // 0.0 to 1.0
	Slack        time.Duration `json:"slack"`   // Float/slack time for critical path
}

// CreateMissionRequest represents a request to create a new mission
type CreateMissionRequest struct {
	Name           string                 `json:"name" validate:"required,min=1,max=255"`
	Description    string                 `json:"description"`
	Priority       int                    `json:"priority" validate:"min=1,max=5"`
	Classification Classification         `json:"classification" validate:"required"`
	StartDate      *time.Time             `json:"start_date"`
	EndDate        *time.Time             `json:"end_date"`
	CommanderID    *uuid.UUID             `json:"commander_id"`
	Location       *Location              `json:"location"`
	Objectives     []CreateObjectiveRequest `json:"objectives"`
	Metadata       map[string]interface{} `json:"metadata"`
}

// UpdateMissionRequest represents a request to update a mission
type UpdateMissionRequest struct {
	Name           *string                `json:"name,omitempty" validate:"omitempty,min=1,max=255"`
	Description    *string                `json:"description,omitempty"`
	Priority       *int                   `json:"priority,omitempty" validate:"omitempty,min=1,max=5"`
	Classification *Classification        `json:"classification,omitempty"`
	StartDate      *time.Time             `json:"start_date,omitempty"`
	EndDate        *time.Time             `json:"end_date,omitempty"`
	CommanderID    *uuid.UUID             `json:"commander_id,omitempty"`
	Location       *Location              `json:"location,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// CreateObjectiveRequest represents a request to create a mission objective
type CreateObjectiveRequest struct {
	Description string `json:"description" validate:"required,min=1"`
	Priority    int    `json:"priority" validate:"min=1,max=5"`
}

// CreateTaskRequest represents a request to create a new task
type CreateTaskRequest struct {
	MissionID      uuid.UUID             `json:"mission_id" validate:"required"`
	Name           string                `json:"name" validate:"required,min=1,max=255"`
	Description    string                `json:"description"`
	Priority       int                   `json:"priority" validate:"min=1,max=5"`
	AssignedTo     *uuid.UUID            `json:"assigned_to"`
	Dependencies   []uuid.UUID           `json:"dependencies"`
	DependsOn      []uuid.UUID           `json:"depends_on"`
	Duration       int                   `json:"duration" validate:"min=0"`          // Duration in minutes
	EstimatedHours int                   `json:"estimated_hours" validate:"min=0"`
	Resources      []TaskResourceRequest `json:"resources"`
	DueDate        *time.Time            `json:"due_date"`
}

// UpdateTaskRequest represents a request to update a task
type UpdateTaskRequest struct {
	Name           *string     `json:"name,omitempty" validate:"omitempty,min=1,max=255"`
	Description    *string     `json:"description,omitempty"`
	Priority       *int        `json:"priority,omitempty" validate:"omitempty,min=1,max=5"`
	AssignedTo     *uuid.UUID  `json:"assigned_to,omitempty"`
	Duration       *int        `json:"duration,omitempty" validate:"omitempty,min=0"`
	EstimatedHours *int        `json:"estimated_hours,omitempty" validate:"omitempty,min=0"`
	ActualHours    *int        `json:"actual_hours,omitempty" validate:"omitempty,min=0"`
	DueDate        *time.Time  `json:"due_date,omitempty"`
}

// ListMissionsFilter represents filters for listing missions
type ListMissionsFilter struct {
	Status      *MissionStatus  `json:"status"`
	Priority    *int            `json:"priority"`
	CommanderID *uuid.UUID      `json:"commander_id"`
	GroupID     *string         `json:"group_id"`
	StartDate   *time.Time      `json:"start_date_after"`
	EndDate     *time.Time      `json:"end_date_before"`
	Limit       int             `json:"limit" validate:"min=1,max=1000"`
	Offset      int             `json:"offset" validate:"min=0"`
}

// ListTasksFilter represents filters for listing tasks
type ListTasksFilter struct {
	MissionID  *uuid.UUID  `json:"mission_id"`
	Status     *TaskStatus `json:"status"`
	Priority   *int        `json:"priority"`
	AssignedTo *uuid.UUID  `json:"assigned_to"`
	DueDate    *time.Time  `json:"due_date_before"`
	Limit      int         `json:"limit" validate:"min=1,max=1000"`
	Offset     int         `json:"offset" validate:"min=0"`
}

// MissionSummary represents a summary view of a mission for list responses
type MissionSummary struct {
	ID             uuid.UUID     `json:"id"`
	Name           string        `json:"name"`
	Status         MissionStatus `json:"status"`
	Priority       int           `json:"priority"`
	Classification Classification `json:"classification"`
	StartDate      *time.Time    `json:"start_date"`
	EndDate        *time.Time    `json:"end_date"`
	CommanderID    *uuid.UUID    `json:"commander_id"`
	TaskCount      int           `json:"task_count"`
	CompletedTasks int           `json:"completed_tasks"`
	Progress       float64       `json:"progress"`
	CreatedAt      time.Time     `json:"created_at"`
	UpdatedAt      time.Time     `json:"updated_at"`
}

// TaskResource represents a resource required for a task
type TaskResource struct {
	ID       uuid.UUID    `json:"id" db:"id"`
	TaskID   uuid.UUID    `json:"task_id" db:"task_id"`
	Type     ResourceType `json:"type" db:"type"`
	Name     string       `json:"name" db:"name"`
	Quantity int          `json:"quantity" db:"quantity"`
}

// TaskResourceRequest represents a request to create a task resource
type TaskResourceRequest struct {
	Type     ResourceType `json:"type" validate:"required"`
	Name     string       `json:"name" validate:"required,min=1,max=255"`
	Quantity int          `json:"quantity" validate:"min=1"`
}


// TaskWithGroupID is a task with group ID for permission checking
type TaskWithGroupID struct {
	Task
	GroupID string `json:"group_id" db:"group_id"`
}
