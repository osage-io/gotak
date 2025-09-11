package mission

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"

	"github.com/dfedick/gotak/internal/events"
	"github.com/dfedick/gotak/pkg/database"
)

// CreateTask creates a new task for a mission
func (s *Service) CreateTask(ctx context.Context, req *CreateTaskRequest) (*Task, error) {
	userID := getUserIDFromContext(ctx)
	groupID := getGroupIDFromContext(ctx)
	
	if userID == "" {
		return nil, errors.New("user ID not found in context")
	}
	
	// Verify mission exists and user has access
	mission, err := s.getMissionFromDB(ctx, req.MissionID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("mission not found")
		}
		return nil, fmt.Errorf("failed to get mission: %w", err)
	}
	
	// Check group access
	if mission.GroupID != groupID {
		return nil, errors.New("insufficient permissions to create task")
	}
	
	// Validate dependencies exist and belong to the same mission
	if len(req.DependsOn) > 0 {
		if err := s.validateTaskDependencies(ctx, req.MissionID, req.DependsOn); err != nil {
			return nil, fmt.Errorf("invalid task dependencies: %w", err)
		}
	}
	
	task := &Task{
		ID:             uuid.New(),
		MissionID:      req.MissionID,
		Name:           req.Name,
		Description:    req.Description,
		Status:         TaskStatusTodo,
		Priority:       req.Priority,
		AssignedTo:     req.AssignedTo,
		Dependencies:   req.Dependencies,
		DependsOn:      req.DependsOn,
		Duration:       req.Duration,
		EstimatedHours: req.EstimatedHours,
		ActualHours:    0,
		DueDate:        req.DueDate,
		CreatedBy:      uuid.MustParse(userID),
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	
	// Auto-assign status if assignee is provided
	if task.AssignedTo != nil {
		task.Status = TaskStatusTodo // Keep as todo until they start working
	}
	
	// Begin transaction
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()
	
	// Insert task
	if err := s.insertTask(ctx, tx, task); err != nil {
		return nil, fmt.Errorf("failed to insert task: %w", err)
	}
	
	// Insert task dependencies
	if len(task.DependsOn) > 0 {
		if err := s.insertTaskDependencies(ctx, tx, task.ID, task.DependsOn); err != nil {
			return nil, fmt.Errorf("failed to insert task dependencies: %w", err)
		}
	}
	
	// Commit transaction
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}
	
	// Publish task created event
	if s.publisher != nil {
		event := events.NewTaskCreatedEvent(userID, groupID, task.ID, task.MissionID, task)
		if err := s.publisher.PublishTaskEvent(ctx, event); err != nil {
			s.logger.Error().Err(err).Msg("Failed to publish task created event")
		}
	}
	
	s.logger.Info().
		Str("task_id", task.ID.String()).
		Str("mission_id", task.MissionID.String()).
		Str("user_id", userID).
		Str("task_name", task.Name).
		Msg("Task created successfully")
	
	return task, nil
}

// GetTask retrieves a task by ID
func (s *Service) GetTask(ctx context.Context, taskID uuid.UUID) (*Task, error) {
	userID := getUserIDFromContext(ctx)
	groupID := getGroupIDFromContext(ctx)
	
	if userID == "" {
		return nil, errors.New("user ID not found in context")
	}
	
	// Get task with mission info for permission checking
	task, err := s.getTaskWithMissionInfo(ctx, taskID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("task not found")
		}
		return nil, fmt.Errorf("failed to get task: %w", err)
	}
	
	// Check group access (task includes mission GroupID)
	if task.GroupID != groupID {
		return nil, errors.New("insufficient permissions to read task")
	}
	
	// Load dependencies
	dependencies, err := s.getTaskDependencies(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to load task dependencies: %w", err)
	}
	task.DependsOn = dependencies
	
	// Convert TaskWithGroupID back to Task for API consistency
	result := &Task{
		ID:             task.ID,
		MissionID:      task.MissionID,
		Name:           task.Name,
		Description:    task.Description,
		Status:         task.Status,
		Priority:       task.Priority,
		AssignedTo:     task.AssignedTo,
		Dependencies:   task.Dependencies,
		DependsOn:      task.DependsOn,
		Duration:       task.Duration,
		EstimatedHours: task.EstimatedHours,
		ActualHours:    task.ActualHours,
		Resources:      task.Resources,
		DueDate:        task.DueDate,
		CompletedAt:    task.CompletedAt,
		CreatedBy:      task.CreatedBy,
		CreatedAt:      task.CreatedAt,
		UpdatedAt:      task.UpdatedAt,
	}
	
	return result, nil
}

// UpdateTask updates an existing task
func (s *Service) UpdateTask(ctx context.Context, taskID uuid.UUID, req *UpdateTaskRequest) (*Task, error) {
	userID := getUserIDFromContext(ctx)
	groupID := getGroupIDFromContext(ctx)
	
	if userID == "" {
		return nil, errors.New("user ID not found in context")
	}
	
	// Get existing task for validation
	task, err := s.getTaskWithMissionInfo(ctx, taskID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("task not found")
		}
		return nil, fmt.Errorf("failed to get task: %w", err)
	}
	
	// Check group access and permissions
	if task.GroupID != groupID {
		return nil, errors.New("insufficient permissions to update task")
	}
	
	// Allow task assignee to update their own tasks
	canUpdate := false
	if task.AssignedTo != nil && task.AssignedTo.String() == userID {
		canUpdate = true
	} else {
		// Check if user has general mission update permissions
		// This would normally involve RBAC check, for now assume group membership is sufficient
		canUpdate = true
	}
	
	if !canUpdate {
		return nil, errors.New("insufficient permissions to update task")
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
	
	if req.AssignedTo != nil {
		setParts = append(setParts, fmt.Sprintf("assigned_to = $%d", argIndex))
		args = append(args, *req.AssignedTo)
		argIndex++
		
		// Update status to in progress if currently just todo
		if task.Status == TaskStatusTodo {
			setParts = append(setParts, fmt.Sprintf("status = $%d", argIndex))
			args = append(args, TaskStatusTodo) // Keep as todo until they explicitly start
			argIndex++
		}
	}
	
	if req.EstimatedHours != nil {
		setParts = append(setParts, fmt.Sprintf("estimated_hours = $%d", argIndex))
		args = append(args, *req.EstimatedHours)
		argIndex++
	}
	
	if req.ActualHours != nil {
		setParts = append(setParts, fmt.Sprintf("actual_hours = $%d", argIndex))
		args = append(args, *req.ActualHours)
		argIndex++
	}
	
	if req.DueDate != nil {
		setParts = append(setParts, fmt.Sprintf("due_date = $%d", argIndex))
		args = append(args, *req.DueDate)
		argIndex++
	}
	
	if len(setParts) == 0 {
		return s.GetTask(ctx, taskID) // No updates, return current state
	}
	
	// Add updated_at and task ID
	setParts = append(setParts, "updated_at = NOW()")
	args = append(args, taskID)
	
	query := fmt.Sprintf("UPDATE tasks SET %s WHERE id = $%d", strings.Join(setParts, ", "), argIndex)
	
	_, err = s.db.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to update task: %w", err)
	}
	
	s.logger.Info().
		Str("task_id", taskID.String()).
		Str("user_id", userID).
		Msg("Task updated successfully")
	
	// Return updated task
	return s.GetTask(ctx, taskID)
}

// DeleteTask deletes a task by ID
func (s *Service) DeleteTask(ctx context.Context, taskID uuid.UUID) error {
	userID := getUserIDFromContext(ctx)
	groupID := getGroupIDFromContext(ctx)
	
	if userID == "" {
		return errors.New("user ID not found in context")
	}
	
	// Get task for permission checking
	task, err := s.getTaskWithMissionInfo(ctx, taskID)
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("task not found")
		}
		return fmt.Errorf("failed to get task: %w", err)
	}
	
	// Check group access
	if task.GroupID != groupID {
		return errors.New("insufficient permissions to delete task")
	}
	
	// Check if task has dependencies that would be broken
	dependentTasks, err := s.getTasksDependentOn(ctx, taskID)
	if err != nil {
		return fmt.Errorf("failed to check dependent tasks: %w", err)
	}
	
	if len(dependentTasks) > 0 {
		return fmt.Errorf("cannot delete task: %d other tasks depend on this task", len(dependentTasks))
	}
	
	// Delete task (cascade will handle dependencies)
	query := "DELETE FROM tasks WHERE id = $1"
	result, err := s.db.ExecContext(ctx, query, taskID)
	if err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return errors.New("task not found")
	}
	
	s.logger.Info().
		Str("task_id", taskID.String()).
		Str("user_id", userID).
		Msg("Task deleted successfully")
	
	return nil
}

// ListTasks lists tasks with filtering and pagination
func (s *Service) ListTasks(ctx context.Context, filter *ListTasksFilter) ([]Task, int, error) {
	userID := getUserIDFromContext(ctx)
	groupID := getGroupIDFromContext(ctx)
	
	if userID == "" {
		return nil, 0, errors.New("user ID not found in context")
	}
	
	// Build WHERE clause - always filter by group via mission
	whereClause := []string{"m.group_id = $1"}
	args := []interface{}{groupID}
	argIndex := 2
	
	if filter.MissionID != nil {
		whereClause = append(whereClause, fmt.Sprintf("t.mission_id = $%d", argIndex))
		args = append(args, *filter.MissionID)
		argIndex++
	}
	
	if filter.Status != nil {
		whereClause = append(whereClause, fmt.Sprintf("t.status = $%d", argIndex))
		args = append(args, *filter.Status)
		argIndex++
	}
	
	if filter.Priority != nil {
		whereClause = append(whereClause, fmt.Sprintf("t.priority = $%d", argIndex))
		args = append(args, *filter.Priority)
		argIndex++
	}
	
	if filter.AssignedTo != nil {
		whereClause = append(whereClause, fmt.Sprintf("t.assigned_to = $%d", argIndex))
		args = append(args, *filter.AssignedTo)
		argIndex++
	}
	
	if filter.DueDate != nil {
		whereClause = append(whereClause, fmt.Sprintf("t.due_date <= $%d", argIndex))
		args = append(args, *filter.DueDate)
		argIndex++
	}
	
	whereSQL := strings.Join(whereClause, " AND ")
	
	// Count total tasks
	countQuery := fmt.Sprintf(`
		SELECT COUNT(*)
		FROM tasks t
		JOIN missions m ON t.mission_id = m.id
		WHERE %s`, whereSQL)
	
	var total int
	err := s.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count tasks: %w", err)
	}
	
	// Query tasks with dependencies
	query := fmt.Sprintf(`
		SELECT t.id, t.mission_id, t.name, t.description, t.status, t.priority, t.assigned_to,
		       t.estimated_hours, t.actual_hours, t.due_date, t.completed_at, t.created_at, t.updated_at,
		       COALESCE(array_agg(td.depends_on_task_id) FILTER (WHERE td.depends_on_task_id IS NOT NULL), '{}') as dependencies
		FROM tasks t
		JOIN missions m ON t.mission_id = m.id
		LEFT JOIN task_dependencies td ON t.id = td.task_id
		WHERE %s
		GROUP BY t.id, t.mission_id, t.name, t.description, t.status, t.priority, t.assigned_to,
		         t.estimated_hours, t.actual_hours, t.due_date, t.completed_at, t.created_at, t.updated_at
		ORDER BY t.priority ASC, t.created_at ASC
		LIMIT $%d OFFSET $%d`,
		whereSQL, argIndex, argIndex+1)
	
	args = append(args, filter.Limit, filter.Offset)
	
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query tasks: %w", err)
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
			return nil, 0, fmt.Errorf("failed to scan task: %w", err)
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
	
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating task rows: %w", err)
	}
	
	return tasks, total, nil
}

// AssignTask assigns a task to a user
func (s *Service) AssignTask(ctx context.Context, taskID uuid.UUID, assigneeID uuid.UUID) error {
	userID := getUserIDFromContext(ctx)
	groupID := getGroupIDFromContext(ctx)
	
	if userID == "" {
		return errors.New("user ID not found in context")
	}
	
	// Get task for permission checking
	task, err := s.getTaskWithMissionInfo(ctx, taskID)
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("task not found")
		}
		return fmt.Errorf("failed to get task: %w", err)
	}
	
	// Check group access
	if task.GroupID != groupID {
		return errors.New("insufficient permissions to assign task")
	}
	
	// Check if task can be assigned (not completed or cancelled)
	if task.Status == TaskStatusCompleted || task.Status == TaskStatusCancelled {
		return fmt.Errorf("cannot assign task with status: %s", task.Status)
	}
	
	// Update task assignment and status
	query := `
		UPDATE tasks 
		SET assigned_to = $1, status = $2, updated_at = NOW()
		WHERE id = $3`
	
	newStatus := TaskStatusTodo
	if task.Status == TaskStatusInProgress {
		newStatus = TaskStatusInProgress // Keep in progress if already started
	}
	
	_, err = s.db.ExecContext(ctx, query, assigneeID, newStatus, taskID)
	if err != nil {
		return fmt.Errorf("failed to assign task: %w", err)
	}
	
	// Publish task assigned event
	if s.publisher != nil {
		event := events.NewTaskAssignedEvent(userID, groupID, taskID, task.MissionID, assigneeID.String(), userID)
		if err := s.publisher.PublishTaskEvent(ctx, event); err != nil {
			s.logger.Error().Err(err).Msg("Failed to publish task assigned event")
		}
	}
	
	s.logger.Info().
		Str("task_id", taskID.String()).
		Str("mission_id", task.MissionID.String()).
		Str("assigned_to", assigneeID.String()).
		Str("assigned_by", userID).
		Msg("Task assigned successfully")
	
	return nil
}

// UpdateTaskStatus updates the status of a task
func (s *Service) UpdateTaskStatus(ctx context.Context, taskID uuid.UUID, status TaskStatus) error {
	userID := getUserIDFromContext(ctx)
	groupID := getGroupIDFromContext(ctx)
	
	if userID == "" {
		return errors.New("user ID not found in context")
	}
	
	// Get task for permission and validation checking
	task, err := s.getTaskWithMissionInfo(ctx, taskID)
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("task not found")
		}
		return fmt.Errorf("failed to get task: %w", err)
	}
	
	// Check group access
	if task.GroupID != groupID {
		return errors.New("insufficient permissions to update task status")
	}
	
	// Check if user can update this task (assignee or has mission permissions)
	canUpdate := false
	if task.AssignedTo != nil && task.AssignedTo.String() == userID {
		canUpdate = true
	} else {
		// General mission permissions (would normally use RBAC)
		canUpdate = true
	}
	
	if !canUpdate {
		return errors.New("insufficient permissions to update task status")
	}
	
	// Validate status transition
	if !isValidTaskStatusTransition(task.Status, status) {
		return fmt.Errorf("invalid task status transition from %s to %s", task.Status, status)
	}
	
	// Check dependencies are met if trying to start task
	if status == TaskStatusInProgress && task.Status != TaskStatusInProgress {
		if err := s.checkTaskDependenciesComplete(ctx, taskID); err != nil {
			return fmt.Errorf("cannot start task: %w", err)
		}
	}
	
	// Build update query
	query := "UPDATE tasks SET status = $1, updated_at = NOW()"
	args := []interface{}{status}
	argIndex := 2
	
	// Set completed timestamp if status is completed
	if status == TaskStatusCompleted {
		query += fmt.Sprintf(", completed_at = NOW()")
	}
	
	query += fmt.Sprintf(" WHERE id = $%d", argIndex)
	args = append(args, taskID)
	
	_, err = s.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update task status: %w", err)
	}
	
	// Publish task status changed event
	if s.publisher != nil {
		event := events.NewTaskStatusChangedEvent(userID, groupID, taskID, task.MissionID, string(task.Status), string(status))
		if err := s.publisher.PublishTaskEvent(ctx, event); err != nil {
			s.logger.Error().Err(err).Msg("Failed to publish task status changed event")
		}
	}
	
	s.logger.Info().
		Str("task_id", taskID.String()).
		Str("mission_id", task.MissionID.String()).
		Str("user_id", userID).
		Str("old_status", string(task.Status)).
		Str("new_status", string(status)).
		Msg("Task status updated")
	
	return nil
}

// Database helper methods for tasks

func (s *Service) insertTask(ctx context.Context, tx database.Tx, task *Task) error {
	query := `
		INSERT INTO tasks (
			id, mission_id, name, description, status, priority, assigned_to,
			estimated_hours, actual_hours, due_date, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`
	
	_, err := tx.ExecContext(ctx, query,
		task.ID, task.MissionID, task.Name, task.Description, task.Status, task.Priority, task.AssignedTo,
		task.EstimatedHours, task.ActualHours, task.DueDate, task.CreatedAt, task.UpdatedAt)
	
	return err
}

func (s *Service) insertTaskDependencies(ctx context.Context, tx database.Tx, taskID uuid.UUID, dependencies []uuid.UUID) error {
	if len(dependencies) == 0 {
		return nil
	}
	
	query := "INSERT INTO task_dependencies (task_id, depends_on_task_id) VALUES ($1, $2)"
	
	for _, depID := range dependencies {
		_, err := tx.ExecContext(ctx, query, taskID, depID)
		if err != nil {
			return fmt.Errorf("failed to insert dependency %s -> %s: %w", taskID, depID, err)
		}
	}
	
	return nil
}

func (s *Service) getTaskWithMissionInfo(ctx context.Context, taskID uuid.UUID) (*TaskWithGroupID, error) {
	query := `
		SELECT t.id, t.mission_id, t.name, t.description, t.status, t.priority, t.assigned_to,
		       t.estimated_hours, t.actual_hours, t.due_date, t.completed_at, t.created_at, t.updated_at,
		       m.group_id
		FROM tasks t
		JOIN missions m ON t.mission_id = m.id
		WHERE t.id = $1`
	
	task := &TaskWithGroupID{}
	
	err := s.db.QueryRowContext(ctx, query, taskID).Scan(
		&task.ID, &task.MissionID, &task.Name, &task.Description, &task.Status, &task.Priority, &task.AssignedTo,
		&task.EstimatedHours, &task.ActualHours, &task.DueDate, &task.CompletedAt, &task.CreatedAt, &task.UpdatedAt,
		&task.GroupID,
	)
	
	return task, err
}

func (s *Service) getTaskDependencies(ctx context.Context, taskID uuid.UUID) ([]uuid.UUID, error) {
	query := "SELECT depends_on_task_id FROM task_dependencies WHERE task_id = $1"
	
	rows, err := s.db.QueryContext(ctx, query, taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	dependencies := make([]uuid.UUID, 0)
	for rows.Next() {
		var depID uuid.UUID
		if err := rows.Scan(&depID); err != nil {
			return nil, err
		}
		dependencies = append(dependencies, depID)
	}
	
	return dependencies, rows.Err()
}

func (s *Service) validateTaskDependencies(ctx context.Context, missionID uuid.UUID, dependencies []uuid.UUID) error {
	if len(dependencies) == 0 {
		return nil
	}
	
	// Check all dependencies exist and belong to the same mission
	placeholders := make([]string, len(dependencies))
	args := make([]interface{}, len(dependencies)+1)
	args[0] = missionID
	
	for i, dep := range dependencies {
		placeholders[i] = fmt.Sprintf("$%d", i+2)
		args[i+1] = dep
	}
	
	query := fmt.Sprintf(`
		SELECT COUNT(*) FROM tasks 
		WHERE mission_id = $1 AND id IN (%s)`,
		strings.Join(placeholders, ","))
	
	var count int
	err := s.db.QueryRowContext(ctx, query, args...).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to validate dependencies: %w", err)
	}
	
	if count != len(dependencies) {
		return errors.New("one or more dependencies do not exist or belong to different mission")
	}
	
	return nil
}

func (s *Service) getTasksDependentOn(ctx context.Context, taskID uuid.UUID) ([]uuid.UUID, error) {
	query := "SELECT task_id FROM task_dependencies WHERE depends_on_task_id = $1"
	
	rows, err := s.db.QueryContext(ctx, query, taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	dependentTasks := make([]uuid.UUID, 0)
	for rows.Next() {
		var taskID uuid.UUID
		if err := rows.Scan(&taskID); err != nil {
			return nil, err
		}
		dependentTasks = append(dependentTasks, taskID)
	}
	
	return dependentTasks, rows.Err()
}

func (s *Service) checkTaskDependenciesComplete(ctx context.Context, taskID uuid.UUID) error {
	query := `
		SELECT COUNT(*) 
		FROM task_dependencies td
		JOIN tasks t ON td.depends_on_task_id = t.id
		WHERE td.task_id = $1 AND t.status != 'completed'`
	
	var incompleteCount int
	err := s.db.QueryRowContext(ctx, query, taskID).Scan(&incompleteCount)
	if err != nil {
		return fmt.Errorf("failed to check task dependencies: %w", err)
	}
	
	if incompleteCount > 0 {
		return fmt.Errorf("task has %d incomplete dependencies", incompleteCount)
	}
	
	return nil
}

// Business logic helper functions

func isValidTaskStatusTransition(from, to TaskStatus) bool {
	validTransitions := map[TaskStatus][]TaskStatus{
		TaskStatusTodo:       {TaskStatusInProgress, TaskStatusBlocked, TaskStatusCancelled},
		TaskStatusInProgress: {TaskStatusCompleted, TaskStatusReview, TaskStatusBlocked, TaskStatusCancelled},
		TaskStatusReview:     {TaskStatusCompleted, TaskStatusInProgress, TaskStatusCancelled},
		TaskStatusBlocked:    {TaskStatusTodo, TaskStatusInProgress, TaskStatusCancelled},
		TaskStatusCompleted:  {}, // Terminal state
		TaskStatusCancelled:  {}, // Terminal state
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

// TaskWithGroupID is defined in models.go
