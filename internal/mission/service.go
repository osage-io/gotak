package mission

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"

	"github.com/dfedick/gotak/internal/events"
	"github.com/dfedick/gotak/pkg/database"
	"github.com/dfedick/gotak/pkg/logger"
)

// Service provides mission management functionality
type Service struct {
	db        database.DB
	logger    logger.Logger
	publisher events.Publisher
}

// NewService creates a new mission service instance
func NewService(db database.DB, logger logger.Logger, publisher events.Publisher) *Service {
	return &Service{
		db:        db,
		logger:    logger,
		publisher: publisher,
	}
}

// CreateMission creates a new mission with the provided details
func (s *Service) CreateMission(ctx context.Context, req *CreateMissionRequest) (*Mission, error) {
	userID := getUserIDFromContext(ctx)
	groupID := getGroupIDFromContext(ctx)
	
	if userID == "" {
		return nil, errors.New("user ID not found in context")
	}
	
	if groupID == "" {
		return nil, errors.New("group ID not found in context")
	}
	
	// Validate end date is after start date
	if req.StartDate != nil && req.EndDate != nil && req.EndDate.Before(*req.StartDate) {
		return nil, errors.New("end date must be after start date")
	}
	
	mission := &Mission{
		ID:             uuid.New(),
		Name:           req.Name,
		Description:    req.Description,
		Status:         StatusPlanning,
		Priority:       req.Priority,
		Classification: req.Classification,
		StartDate:      req.StartDate,
		EndDate:        req.EndDate,
		CommanderID:    req.CommanderID,
		CreatedBy:      uuid.MustParse(userID),
		GroupID:        groupID,
		Metadata:       req.Metadata,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	
	// Set location if provided
	if req.Location != nil {
		mission.Location = req.Location
	}
	
	// Begin transaction
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()
	
	// Insert mission
	if err := s.insertMission(ctx, tx, mission); err != nil {
		return nil, fmt.Errorf("failed to insert mission: %w", err)
	}
	
	// Insert objectives if provided
	if len(req.Objectives) > 0 {
		objectives := make([]Objective, 0, len(req.Objectives))
		for _, objReq := range req.Objectives {
			objective := Objective{
				ID:          uuid.New(),
				MissionID:   mission.ID,
				Description: objReq.Description,
				Priority:    objReq.Priority,
				Completed:   false,
				CreatedAt:   time.Now(),
			}
			objectives = append(objectives, objective)
		}
		
		if err := s.insertObjectives(ctx, tx, objectives); err != nil {
			return nil, fmt.Errorf("failed to insert objectives: %w", err)
		}
		
		mission.Objectives = objectives
	}
	
	// Commit transaction
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}
	
	// Publish mission created event
	if s.publisher != nil {
		event := events.NewMissionCreatedEvent(userID, groupID, mission.ID, mission)
		if err := s.publisher.PublishMissionEvent(ctx, event); err != nil {
			s.logger.Error().Err(err).Msg("Failed to publish mission created event")
		}
	}
	
	s.logger.Info().
		Str("mission_id", mission.ID.String()).
		Str("user_id", userID).
		Str("mission_name", mission.Name).
		Msg("Mission created successfully")
	
	return mission, nil
}

// GetMission retrieves a mission by ID with all related data
func (s *Service) GetMission(ctx context.Context, missionID uuid.UUID) (*Mission, error) {
	userID := getUserIDFromContext(ctx)
	groupID := getGroupIDFromContext(ctx)
	
	if userID == "" {
		return nil, errors.New("user ID not found in context")
	}
	
	// Get mission from database
	mission, err := s.getMissionFromDB(ctx, missionID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("mission not found")
		}
		return nil, fmt.Errorf("failed to get mission: %w", err)
	}
	
	// Check group access
	if mission.GroupID != groupID {
		return nil, errors.New("insufficient permissions to read mission")
	}
	
	// Load objectives
	objectives, err := s.getObjectivesByMission(ctx, missionID)
	if err != nil {
		return nil, fmt.Errorf("failed to load mission objectives: %w", err)
	}
	mission.Objectives = objectives
	
	// Load tasks with dependencies
	tasks, err := s.getTasksByMissionWithDependencies(ctx, missionID)
	if err != nil {
		return nil, fmt.Errorf("failed to load mission tasks: %w", err)
	}
	mission.Tasks = tasks
	
	// Load milestones
	milestones, err := s.getMilestonesByMission(ctx, missionID)
	if err != nil {
		return nil, fmt.Errorf("failed to load mission milestones: %w", err)
	}
	mission.Milestones = milestones
	
	// Load resource requests
	resources, err := s.getResourceRequestsByMission(ctx, missionID)
	if err != nil {
		return nil, fmt.Errorf("failed to load mission resources: %w", err)
	}
	mission.Resources = resources
	
	return mission, nil
}

// UpdateMission updates an existing mission
func (s *Service) UpdateMission(ctx context.Context, missionID uuid.UUID, req *UpdateMissionRequest) (*Mission, error) {
	userID := getUserIDFromContext(ctx)
	groupID := getGroupIDFromContext(ctx)
	
	if userID == "" {
		return nil, errors.New("user ID not found in context")
	}
	
	// Get existing mission for validation
	mission, err := s.getMissionFromDB(ctx, missionID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("mission not found")
		}
		return nil, fmt.Errorf("failed to get mission: %w", err)
	}
	
	// Check group access and permissions
	if mission.GroupID != groupID {
		return nil, errors.New("insufficient permissions to update mission")
	}
	
	// Build update query dynamically
	setParts := []string{}
	args := []interface{}{}
	argIndex := 1
	
	if req.Name != nil {
		setParts = append(setParts, fmt.Sprintf("name = $%d", argIndex))
		args = append(args, *req.Name)
		argIndex++
	}
	
	if req.Description != nil {
		setParts = append(setParts, fmt.Sprintf("description = $%d", argIndex))
		args = append(args, *req.Description)
		argIndex++
	}
	
	if req.Priority != nil {
		setParts = append(setParts, fmt.Sprintf("priority = $%d", argIndex))
		args = append(args, *req.Priority)
		argIndex++
	}
	
	if req.Classification != nil {
		setParts = append(setParts, fmt.Sprintf("classification = $%d", argIndex))
		args = append(args, *req.Classification)
		argIndex++
	}
	
	if req.StartDate != nil {
		setParts = append(setParts, fmt.Sprintf("start_date = $%d", argIndex))
		args = append(args, *req.StartDate)
		argIndex++
	}
	
	if req.EndDate != nil {
		setParts = append(setParts, fmt.Sprintf("end_date = $%d", argIndex))
		args = append(args, *req.EndDate)
		argIndex++
	}
	
	if req.CommanderID != nil {
		setParts = append(setParts, fmt.Sprintf("commander_id = $%d", argIndex))
		args = append(args, *req.CommanderID)
		argIndex++
	}
	
	if req.Location != nil {
		setParts = append(setParts, fmt.Sprintf("latitude = $%d, longitude = $%d, location_name = $%d, location_description = $%d", 
			argIndex, argIndex+1, argIndex+2, argIndex+3))
		args = append(args, req.Location.Latitude, req.Location.Longitude, req.Location.Name, req.Location.Description)
		argIndex += 4
	}
	
	if req.Metadata != nil {
		metadataJSON, err := json.Marshal(req.Metadata)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal metadata: %w", err)
		}
		setParts = append(setParts, fmt.Sprintf("metadata = $%d", argIndex))
		args = append(args, metadataJSON)
		argIndex++
	}
	
	if len(setParts) == 0 {
		return s.GetMission(ctx, missionID) // No updates, return current state
	}
	
	// Add updated_at and mission ID
	setParts = append(setParts, fmt.Sprintf("updated_at = NOW()"))
	args = append(args, missionID)
	
	query := fmt.Sprintf("UPDATE missions SET %s WHERE id = $%d", strings.Join(setParts, ", "), argIndex)
	
	_, err = s.db.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to update mission: %w", err)
	}
	
	s.logger.Info().
		Str("mission_id", missionID.String()).
		Str("user_id", userID).
		Msg("Mission updated successfully")
	
	// Return updated mission
	return s.GetMission(ctx, missionID)
}

// DeleteMission deletes a mission by ID
func (s *Service) DeleteMission(ctx context.Context, missionID uuid.UUID) error {
	userID := getUserIDFromContext(ctx)
	groupID := getGroupIDFromContext(ctx)
	
	if userID == "" {
		return errors.New("user ID not found in context")
	}
	
	// Get mission for permission checking
	mission, err := s.getMissionFromDB(ctx, missionID)
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("mission not found")
		}
		return fmt.Errorf("failed to get mission: %w", err)
	}
	
	// Check group access
	if mission.GroupID != groupID {
		return errors.New("insufficient permissions to delete mission")
	}
	
	// Delete mission (cascade will handle related tables)
	query := "DELETE FROM missions WHERE id = $1"
	result, err := s.db.ExecContext(ctx, query, missionID)
	if err != nil {
		return fmt.Errorf("failed to delete mission: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return errors.New("mission not found")
	}
	
	s.logger.Info().
		Str("mission_id", missionID.String()).
		Str("user_id", userID).
		Msg("Mission deleted successfully")
	
	return nil
}

// ListMissions lists missions with filtering and pagination
func (s *Service) ListMissions(ctx context.Context, filter *ListMissionsFilter) ([]MissionSummary, int, error) {
	userID := getUserIDFromContext(ctx)
	groupID := getGroupIDFromContext(ctx)
	
	if userID == "" {
		return nil, 0, errors.New("user ID not found in context")
	}
	
	// Build WHERE clause
	whereClause := []string{"group_id = $1"}
	args := []interface{}{groupID}
	argIndex := 2
	
	if filter.Status != nil {
		whereClause = append(whereClause, fmt.Sprintf("status = $%d", argIndex))
		args = append(args, *filter.Status)
		argIndex++
	}
	
	if filter.Priority != nil {
		whereClause = append(whereClause, fmt.Sprintf("priority = $%d", argIndex))
		args = append(args, *filter.Priority)
		argIndex++
	}
	
	if filter.CommanderID != nil {
		whereClause = append(whereClause, fmt.Sprintf("commander_id = $%d", argIndex))
		args = append(args, *filter.CommanderID)
		argIndex++
	}
	
	if filter.StartDate != nil {
		whereClause = append(whereClause, fmt.Sprintf("start_date >= $%d", argIndex))
		args = append(args, *filter.StartDate)
		argIndex++
	}
	
	if filter.EndDate != nil {
		whereClause = append(whereClause, fmt.Sprintf("end_date <= $%d", argIndex))
		args = append(args, *filter.EndDate)
		argIndex++
	}
	
	whereSQL := strings.Join(whereClause, " AND ")
	
	// Count total missions
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM missions WHERE %s", whereSQL)
	var total int
	err := s.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count missions: %w", err)
	}
	
	// Query missions with task counts and progress
	query := fmt.Sprintf(`
		SELECT m.id, m.name, m.status, m.priority, m.classification, 
		       m.start_date, m.end_date, m.commander_id,
		       m.created_at, m.updated_at,
		       COALESCE(task_counts.total_tasks, 0) as task_count,
		       COALESCE(task_counts.completed_tasks, 0) as completed_tasks,
		       CASE 
		           WHEN COALESCE(task_counts.total_tasks, 0) = 0 THEN 0.0
		           ELSE COALESCE(task_counts.completed_tasks, 0)::float / task_counts.total_tasks
		       END as progress
		FROM missions m
		LEFT JOIN (
		    SELECT mission_id,
		           COUNT(*) as total_tasks,
		           COUNT(CASE WHEN status = 'completed' THEN 1 END) as completed_tasks
		    FROM tasks
		    GROUP BY mission_id
		) task_counts ON m.id = task_counts.mission_id
		WHERE %s
		ORDER BY m.created_at DESC
		LIMIT $%d OFFSET $%d`,
		whereSQL, argIndex, argIndex+1)
	
	args = append(args, filter.Limit, filter.Offset)
	
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query missions: %w", err)
	}
	defer rows.Close()
	
	missions := make([]MissionSummary, 0)
	for rows.Next() {
		var mission MissionSummary
		err := rows.Scan(
			&mission.ID, &mission.Name, &mission.Status, &mission.Priority, &mission.Classification,
			&mission.StartDate, &mission.EndDate, &mission.CommanderID,
			&mission.CreatedAt, &mission.UpdatedAt,
			&mission.TaskCount, &mission.CompletedTasks, &mission.Progress,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan mission: %w", err)
		}
		missions = append(missions, mission)
	}
	
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating mission rows: %w", err)
	}
	
	return missions, total, nil
}

// UpdateMissionStatus updates the status of a mission with history tracking
func (s *Service) UpdateMissionStatus(ctx context.Context, missionID uuid.UUID, status MissionStatus, reason string) error {
	userID := getUserIDFromContext(ctx)
	groupID := getGroupIDFromContext(ctx)
	
	if userID == "" {
		return errors.New("user ID not found in context")
	}
	
	// Get current mission
	mission, err := s.getMissionFromDB(ctx, missionID)
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("mission not found")
		}
		return fmt.Errorf("failed to get mission: %w", err)
	}
	
	// Check group access
	if mission.GroupID != groupID {
		return errors.New("insufficient permissions to update mission status")
	}
	
	oldStatus := mission.Status
	
	// Validate status transition
	if !isValidMissionStatusTransition(oldStatus, status) {
		return fmt.Errorf("invalid status transition from %s to %s", oldStatus, status)
	}
	
	// Begin transaction
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()
	
	// Update mission status
	updateQuery := "UPDATE missions SET status = $1, updated_at = NOW() WHERE id = $2"
	_, err = tx.ExecContext(ctx, updateQuery, status, missionID)
	if err != nil {
		return fmt.Errorf("failed to update mission status: %w", err)
	}
	
	// Insert status history record
	historyQuery := `
		INSERT INTO mission_status_history (mission_id, old_status, new_status, changed_by, reason)
		VALUES ($1, $2, $3, $4, $5)`
	
	_, err = tx.ExecContext(ctx, historyQuery, missionID, oldStatus, status, userID, reason)
	if err != nil {
		return fmt.Errorf("failed to insert status history: %w", err)
	}
	
	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	
	// Publish mission status changed event
	if s.publisher != nil {
		event := events.NewMissionStatusChangedEvent(userID, groupID, missionID, string(oldStatus), string(status), reason)
		if err := s.publisher.PublishMissionEvent(ctx, event); err != nil {
			s.logger.Error().Err(err).Msg("Failed to publish mission status changed event")
		}
	}
	
	s.logger.Info().
		Str("mission_id", missionID.String()).
		Str("user_id", userID).
		Str("old_status", string(oldStatus)).
		Str("new_status", string(status)).
		Str("reason", reason).
		Msg("Mission status updated")
	
	return nil
}

// Database helper methods

func (s *Service) insertMission(ctx context.Context, tx database.Tx, mission *Mission) error {
	var metadataJSON []byte
	var err error
	
	if mission.Metadata != nil {
		metadataJSON, err = json.Marshal(mission.Metadata)
		if err != nil {
			return fmt.Errorf("failed to marshal metadata: %w", err)
		}
	}
	
	query := `
		INSERT INTO missions (
			id, name, description, status, priority, classification,
			start_date, end_date, commander_id, created_by, group_id,
			latitude, longitude, location_name, location_description,
			metadata, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)`
	
	var latitude, longitude *float64
	var locationName, locationDescription *string
	
	if mission.Location != nil {
		latitude = &mission.Location.Latitude
		longitude = &mission.Location.Longitude
		locationName = &mission.Location.Name
		locationDescription = &mission.Location.Description
	}
	
	_, err = tx.ExecContext(ctx, query,
		mission.ID, mission.Name, mission.Description, mission.Status, mission.Priority, mission.Classification,
		mission.StartDate, mission.EndDate, mission.CommanderID, mission.CreatedBy, mission.GroupID,
		latitude, longitude, locationName, locationDescription,
		metadataJSON, mission.CreatedAt, mission.UpdatedAt)
	
	return err
}

func (s *Service) getMissionFromDB(ctx context.Context, missionID uuid.UUID) (*Mission, error) {
	query := `
		SELECT id, name, description, status, priority, classification,
		       start_date, end_date, commander_id, created_by, group_id,
		       latitude, longitude, location_name, location_description,
		       metadata, created_at, updated_at
		FROM missions WHERE id = $1`
	
	mission := &Mission{}
	var metadataJSON []byte
	var latitude, longitude *float64
	var locationName, locationDescription *string
	
	err := s.db.QueryRowContext(ctx, query, missionID).Scan(
		&mission.ID, &mission.Name, &mission.Description, &mission.Status, &mission.Priority, &mission.Classification,
		&mission.StartDate, &mission.EndDate, &mission.CommanderID, &mission.CreatedBy, &mission.GroupID,
		&latitude, &longitude, &locationName, &locationDescription,
		&metadataJSON, &mission.CreatedAt, &mission.UpdatedAt,
	)
	
	if err != nil {
		return nil, err
	}
	
	// Parse location if present
	if latitude != nil && longitude != nil {
		mission.Location = &Location{
			Latitude:  *latitude,
			Longitude: *longitude,
		}
		if locationName != nil {
			mission.Location.Name = *locationName
		}
		if locationDescription != nil {
			mission.Location.Description = *locationDescription
		}
	}
	
	// Parse metadata
	if len(metadataJSON) > 0 {
		err = json.Unmarshal(metadataJSON, &mission.Metadata)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
	}
	
	return mission, nil
}

func (s *Service) insertObjectives(ctx context.Context, tx database.Tx, objectives []Objective) error {
	if len(objectives) == 0 {
		return nil
	}
	
	query := `
		INSERT INTO mission_objectives (id, mission_id, description, priority, completed, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)`
	
	for _, objective := range objectives {
		_, err := tx.ExecContext(ctx, query,
			objective.ID, objective.MissionID, objective.Description, objective.Priority,
			objective.Completed, objective.CreatedAt)
		if err != nil {
			return fmt.Errorf("failed to insert objective %s: %w", objective.ID, err)
		}
	}
	
	return nil
}

func (s *Service) getObjectivesByMission(ctx context.Context, missionID uuid.UUID) ([]Objective, error) {
	query := `
		SELECT id, mission_id, description, priority, completed, completed_at, created_at
		FROM mission_objectives
		WHERE mission_id = $1
		ORDER BY priority ASC, created_at ASC`
	
	rows, err := s.db.QueryContext(ctx, query, missionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	objectives := make([]Objective, 0)
	for rows.Next() {
		var objective Objective
		err := rows.Scan(
			&objective.ID, &objective.MissionID, &objective.Description, &objective.Priority,
			&objective.Completed, &objective.CompletedAt, &objective.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan objective: %w", err)
		}
		objectives = append(objectives, objective)
	}
	
	return objectives, rows.Err()
}

func (s *Service) getTasksByMissionWithDependencies(ctx context.Context, missionID uuid.UUID) ([]Task, error) {
	query := `
		SELECT t.id, t.mission_id, t.name, t.description, t.status, t.priority, t.assigned_to,
		       t.estimated_hours, t.actual_hours, t.due_date, t.completed_at, t.created_at, t.updated_at,
		       COALESCE(array_agg(td.depends_on_task_id) FILTER (WHERE td.depends_on_task_id IS NOT NULL), '{}') as dependencies
		FROM tasks t
		LEFT JOIN task_dependencies td ON t.id = td.task_id
		WHERE t.mission_id = $1
		GROUP BY t.id, t.mission_id, t.name, t.description, t.status, t.priority, t.assigned_to,
		         t.estimated_hours, t.actual_hours, t.due_date, t.completed_at, t.created_at, t.updated_at
		ORDER BY t.priority ASC, t.created_at ASC`
	
	rows, err := s.db.QueryContext(ctx, query, missionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	tasks := make([]Task, 0)
	for rows.Next() {
		var task Task
		var dependencyUUIDs pq.StringArray
		
		err := rows.Scan(
			&task.ID, &task.MissionID, &task.Name, &task.Description, &task.Status, &task.Priority, &task.AssignedTo,
			&task.EstimatedHours, &task.ActualHours, &task.DueDate, &task.CompletedAt, &task.CreatedAt, &task.UpdatedAt,
			&dependencyUUIDs,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan task: %w", err)
		}
		
		// Parse dependencies
		for _, depStr := range dependencyUUIDs {
			if depStr != "" {
				depUUID, err := uuid.Parse(depStr)
				if err == nil {
					task.DependsOn = append(task.DependsOn, depUUID)
				}
			}
		}
		
		tasks = append(tasks, task)
	}
	
	return tasks, rows.Err()
}

func (s *Service) getMilestonesByMission(ctx context.Context, missionID uuid.UUID) ([]Milestone, error) {
	query := `
		SELECT id, mission_id, name, description, milestone_date, completed, completed_at, created_at
		FROM mission_milestones
		WHERE mission_id = $1
		ORDER BY milestone_date ASC`
	
	rows, err := s.db.QueryContext(ctx, query, missionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	milestones := make([]Milestone, 0)
	for rows.Next() {
		var milestone Milestone
		err := rows.Scan(
			&milestone.ID, &milestone.MissionID, &milestone.Name, &milestone.Description,
			&milestone.MilestoneDate, &milestone.Completed, &milestone.CompletedAt, &milestone.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan milestone: %w", err)
		}
		milestones = append(milestones, milestone)
	}
	
	return milestones, rows.Err()
}

func (s *Service) getResourceRequestsByMission(ctx context.Context, missionID uuid.UUID) ([]ResourceRequest, error) {
	query := `
		SELECT id, mission_id, resource_type, resource_id, resource_name, quantity,
		       required_date, status, requested_by, approved_by, notes, created_at, updated_at
		FROM mission_resource_requests
		WHERE mission_id = $1
		ORDER BY created_at ASC`
	
	rows, err := s.db.QueryContext(ctx, query, missionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	resources := make([]ResourceRequest, 0)
	for rows.Next() {
		var resource ResourceRequest
		err := rows.Scan(
			&resource.ID, &resource.MissionID, &resource.ResourceType, &resource.ResourceID, &resource.ResourceName,
			&resource.Quantity, &resource.RequiredDate, &resource.Status, &resource.RequestedBy, &resource.ApprovedBy,
			&resource.Notes, &resource.CreatedAt, &resource.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan resource request: %w", err)
		}
		resources = append(resources, resource)
	}
	
	return resources, rows.Err()
}

// Business logic helper functions

func isValidMissionStatusTransition(from, to MissionStatus) bool {
	validTransitions := map[MissionStatus][]MissionStatus{
		StatusPlanning:  {StatusApproved, StatusCancelled},
		StatusApproved:  {StatusActive, StatusOnHold, StatusCancelled, StatusPlanning},
		StatusActive:    {StatusOnHold, StatusCompleted, StatusCancelled},
		StatusOnHold:    {StatusActive, StatusCancelled},
		StatusCompleted: {}, // Terminal state
		StatusCancelled: {}, // Terminal state
	}
	
	allowedStatuses, exists := validTransitions[from]
	if !exists {
		return false
	}
	
	for _, allowed := range allowedStatuses {
		if allowed == to {
			return true
		}
	}
	
	return false
}

// Context helper functions

func getUserIDFromContext(ctx context.Context) string {
	if userID, ok := ctx.Value("user_id").(string); ok {
		return userID
	}
	return ""
}

func getGroupIDFromContext(ctx context.Context) string {
	if groupID, ok := ctx.Value("group_id").(string); ok {
		return groupID
	}
	return ""
}
