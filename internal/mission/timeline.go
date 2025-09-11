package mission

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/google/uuid"

	"github.com/dfedick/gotak/internal/events"
)

// GetMissionTimeline retrieves the complete timeline for a mission including critical path
func (s *Service) GetMissionTimeline(ctx context.Context, missionID uuid.UUID) (*Timeline, error) {
	userID := getUserIDFromContext(ctx)
	groupID := getGroupIDFromContext(ctx)
	
	if userID == "" {
		return nil, errors.New("user ID not found in context")
	}
	
	// Get mission for permission checking
	mission, err := s.getMissionFromDB(ctx, missionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get mission: %w", err)
	}
	
	// Check group access
	if mission.GroupID != groupID {
		return nil, errors.New("insufficient permissions to read mission timeline")
	}
	
	// Get tasks with dependencies for timeline calculation
	tasks, err := s.getTasksForTimeline(ctx, missionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tasks for timeline: %w", err)
	}
	
	// Get milestones
	milestones, err := s.getMilestonesByMission(ctx, missionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get mission milestones: %w", err)
	}
	
	// Convert tasks to timeline tasks and calculate scheduling
	timelineTasks, err := s.calculateTaskScheduling(tasks)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate task scheduling: %w", err)
	}
	
	// Calculate critical path
	criticalPath := s.calculateCriticalPath(timelineTasks)
	
	// Mark critical path tasks
	criticalPathMap := make(map[uuid.UUID]bool)
	for _, taskID := range criticalPath {
		criticalPathMap[taskID] = true
	}
	
	for i := range timelineTasks {
		timelineTasks[i].CriticalPath = criticalPathMap[timelineTasks[i].ID]
	}
	
	// Calculate overall timeline dates
	startDate := mission.StartDate
	endDate := mission.EndDate
	
	if startDate == nil && len(timelineTasks) > 0 {
		earliestStart := timelineTasks[0].StartDate
		for _, task := range timelineTasks {
			if task.StartDate.Before(earliestStart) {
				earliestStart = task.StartDate
			}
		}
		startDate = &earliestStart
	}
	
	if endDate == nil && len(timelineTasks) > 0 {
		latestEnd := timelineTasks[0].EndDate
		for _, task := range timelineTasks {
			if task.EndDate.After(latestEnd) {
				latestEnd = task.EndDate
			}
		}
		endDate = &latestEnd
	}
	
	timeline := &Timeline{
		MissionID:    missionID,
		StartDate:    *startDate,
		EndDate:      *endDate,
		Milestones:   milestones,
		Tasks:        timelineTasks,
		CriticalPath: criticalPath,
	}
	
	return timeline, nil
}

// calculateTaskScheduling performs forward and backward pass scheduling
func (s *Service) calculateTaskScheduling(tasks []Task) ([]TimelineTask, error) {
	if len(tasks) == 0 {
		return []TimelineTask{}, nil
	}
	
	// Build dependency graph
	taskMap := make(map[uuid.UUID]*Task)
	for i := range tasks {
		taskMap[tasks[i].ID] = &tasks[i]
	}
	
	// Convert to timeline tasks with initial scheduling
	timelineTasks := make([]TimelineTask, 0, len(tasks))
	for _, task := range tasks {
		timelineTask := TimelineTask{
			ID:           task.ID,
			Name:         task.Name,
			Dependencies: task.DependsOn,
			Status:       task.Status,
			Progress:     s.calculateTaskProgress(task),
		}
		
		// Calculate duration from estimated hours (assume 8-hour work days)
		if task.EstimatedHours > 0 {
			timelineTask.Duration = time.Duration(task.EstimatedHours) * time.Hour
		} else {
			timelineTask.Duration = 8 * time.Hour // Default 1 day
		}
		
		// Set assigned to username (would normally query user service)
		if task.AssignedTo != nil {
			username := task.AssignedTo.String() // Placeholder - would normally resolve username
			timelineTask.AssignedTo = &username
		}
		
		timelineTasks = append(timelineTasks, timelineTask)
	}
	
	// Perform forward pass to calculate early start/finish dates
	if err := s.forwardPass(timelineTasks, taskMap); err != nil {
		return nil, fmt.Errorf("forward pass failed: %w", err)
	}
	
	// Perform backward pass to calculate late start/finish dates and slack
	if err := s.backwardPass(timelineTasks, taskMap); err != nil {
		return nil, fmt.Errorf("backward pass failed: %w", err)
	}
	
	return timelineTasks, nil
}

// forwardPass calculates early start and finish dates
func (s *Service) forwardPass(timelineTasks []TimelineTask, taskMap map[uuid.UUID]*Task) error {
	taskTimelineMap := make(map[uuid.UUID]*TimelineTask)
	for i := range timelineTasks {
		taskTimelineMap[timelineTasks[i].ID] = &timelineTasks[i]
	}
	
	processed := make(map[uuid.UUID]bool)
	
	// Process tasks in topological order
	var processTask func(taskID uuid.UUID) error
	processTask = func(taskID uuid.UUID) error {
		if processed[taskID] {
			return nil
		}
		
		task, exists := taskTimelineMap[taskID]
		if !exists {
			return fmt.Errorf("task %s not found", taskID)
		}
		
		// Process dependencies first
		for _, depID := range task.Dependencies {
			if err := processTask(depID); err != nil {
				return err
			}
		}
		
		// Calculate early start date
		var earlyStart time.Time
		if len(task.Dependencies) == 0 {
			// No dependencies - start immediately
			earlyStart = time.Now().Truncate(24 * time.Hour) // Start of today
			if taskFromDB := taskMap[taskID]; taskFromDB.DueDate != nil {
				// If there's a due date, work backwards
				earlyFinish := taskFromDB.DueDate.Add(-task.Duration)
				if earlyFinish.Before(earlyStart) {
					earlyStart = earlyFinish
				}
			}
		} else {
			// Start after latest dependency finish
			for _, depID := range task.Dependencies {
				depTask := taskTimelineMap[depID]
				if depTask.EndDate.After(earlyStart) {
					earlyStart = depTask.EndDate
				}
			}
			// Align to business hours (8 AM start)
			if earlyStart.Hour() != 8 {
				earlyStart = time.Date(earlyStart.Year(), earlyStart.Month(), earlyStart.Day()+1, 8, 0, 0, 0, earlyStart.Location())
			}
		}
		
		// Calculate early finish date
		earlyFinish := s.addBusinessHours(earlyStart, task.Duration)
		
		task.StartDate = earlyStart
		task.EndDate = earlyFinish
		
		processed[taskID] = true
		return nil
	}
	
	// Process all tasks
	for _, task := range timelineTasks {
		if err := processTask(task.ID); err != nil {
			return err
		}
	}
	
	return nil
}

// backwardPass calculates late start and finish dates
func (s *Service) backwardPass(timelineTasks []TimelineTask, taskMap map[uuid.UUID]*Task) error {
	taskTimelineMap := make(map[uuid.UUID]*TimelineTask)
	successors := make(map[uuid.UUID][]uuid.UUID)
	
	for i := range timelineTasks {
		task := &timelineTasks[i]
		taskTimelineMap[task.ID] = task
		
		// Build successor relationships
		for _, depID := range task.Dependencies {
			successors[depID] = append(successors[depID], task.ID)
		}
	}
	
	// Find project end date (latest early finish)
	projectEnd := timelineTasks[0].EndDate
	for _, task := range timelineTasks {
		if task.EndDate.After(projectEnd) {
			projectEnd = task.EndDate
		}
	}
	
	processed := make(map[uuid.UUID]bool)
	
	var processTask func(taskID uuid.UUID) error
	processTask = func(taskID uuid.UUID) error {
		if processed[taskID] {
			return nil
		}
		
		task := taskTimelineMap[taskID]
		
		// Process successors first
		for _, succID := range successors[taskID] {
			if err := processTask(succID); err != nil {
				return err
			}
		}
		
		// Calculate late finish date
		var lateFinish time.Time
		if len(successors[taskID]) == 0 {
			// No successors - can finish at project end
			lateFinish = projectEnd
		} else {
			// Must finish before earliest successor late start
			lateFinish = time.Date(9999, 12, 31, 0, 0, 0, 0, time.UTC) // Far future
			for _, succID := range successors[taskID] {
				succTask := taskTimelineMap[succID]
				succLateStart := s.subtractBusinessHours(succTask.EndDate, succTask.Duration)
				if succLateStart.Before(lateFinish) {
					lateFinish = succLateStart
				}
			}
		}
		
		// Calculate late start date
		lateStart := s.subtractBusinessHours(lateFinish, task.Duration)
		
		// Calculate slack (float)
		slack := lateStart.Sub(task.StartDate)
		task.Slack = slack
		
		processed[taskID] = true
		return nil
	}
	
	// Process all tasks
	for _, task := range timelineTasks {
		if err := processTask(task.ID); err != nil {
			return err
		}
	}
	
	return nil
}

// calculateCriticalPath finds the longest path through the network
func (s *Service) calculateCriticalPath(tasks []TimelineTask) []uuid.UUID {
	criticalTasks := make([]uuid.UUID, 0)
	
	// Tasks with zero or minimal slack are on critical path
	const slackThreshold = time.Minute * 30 // Allow 30 minutes tolerance
	
	for _, task := range tasks {
		if task.Slack <= slackThreshold {
			criticalTasks = append(criticalTasks, task.ID)
		}
	}
	
	// Sort critical tasks by start date for logical ordering
	sort.Slice(criticalTasks, func(i, j int) bool {
		var taskI, taskJ *TimelineTask
		for k := range tasks {
			if tasks[k].ID == criticalTasks[i] {
				taskI = &tasks[k]
			}
			if tasks[k].ID == criticalTasks[j] {
				taskJ = &tasks[k]
			}
		}
		return taskI.StartDate.Before(taskJ.StartDate)
	})
	
	return criticalTasks
}

// calculateTaskProgress determines task completion percentage
func (s *Service) calculateTaskProgress(task Task) float64 {
	switch task.Status {
	case TaskStatusTodo:
		return 0.0
	case TaskStatusReview:
		return 0.8 // 80% for being in review
	case TaskStatusInProgress:
		// Use actual vs estimated hours if available
		if task.EstimatedHours > 0 && task.ActualHours > 0 {
			progress := float64(task.ActualHours) / float64(task.EstimatedHours)
			// Cap at 90% until completed
			if progress > 0.9 {
				progress = 0.9
			}
			return progress
		}
		return 0.5 // Default 50% for in progress
	case TaskStatusCompleted:
		return 1.0
	case TaskStatusBlocked:
		return 0.3 // Some progress made but blocked
	case TaskStatusCancelled:
		return 0.0
	default:
		return 0.0
	}
}

// addBusinessHours adds duration considering business hours (8 AM - 6 PM, Mon-Fri)
func (s *Service) addBusinessHours(start time.Time, duration time.Duration) time.Time {
	current := start
	remaining := duration
	
	for remaining > 0 {
		// Skip weekends
		if current.Weekday() == time.Saturday || current.Weekday() == time.Sunday {
			current = current.AddDate(0, 0, 1)
			current = time.Date(current.Year(), current.Month(), current.Day(), 8, 0, 0, 0, current.Location())
			continue
		}
		
		// Ensure we're in business hours
		if current.Hour() < 8 {
			current = time.Date(current.Year(), current.Month(), current.Day(), 8, 0, 0, 0, current.Location())
		} else if current.Hour() >= 18 {
			current = current.AddDate(0, 0, 1)
			current = time.Date(current.Year(), current.Month(), current.Day(), 8, 0, 0, 0, current.Location())
			continue
		}
		
		// Calculate how much time left in current business day
		endOfDay := time.Date(current.Year(), current.Month(), current.Day(), 18, 0, 0, 0, current.Location())
		timeLeftToday := endOfDay.Sub(current)
		
		if remaining <= timeLeftToday {
			// Task fits in current day
			current = current.Add(remaining)
			remaining = 0
		} else {
			// Move to next business day
			remaining -= timeLeftToday
			current = current.AddDate(0, 0, 1)
			current = time.Date(current.Year(), current.Month(), current.Day(), 8, 0, 0, 0, current.Location())
		}
	}
	
	return current
}

// subtractBusinessHours subtracts duration considering business hours
func (s *Service) subtractBusinessHours(end time.Time, duration time.Duration) time.Time {
	current := end
	remaining := duration
	
	for remaining > 0 {
		// Skip weekends
		if current.Weekday() == time.Saturday || current.Weekday() == time.Sunday {
			current = current.AddDate(0, 0, -1)
			current = time.Date(current.Year(), current.Month(), current.Day(), 18, 0, 0, 0, current.Location())
			continue
		}
		
		// Ensure we're in business hours
		if current.Hour() >= 18 {
			current = time.Date(current.Year(), current.Month(), current.Day(), 18, 0, 0, 0, current.Location())
		} else if current.Hour() < 8 {
			current = current.AddDate(0, 0, -1)
			current = time.Date(current.Year(), current.Month(), current.Day(), 18, 0, 0, 0, current.Location())
			continue
		}
		
		// Calculate how much time from start of current business day
		startOfDay := time.Date(current.Year(), current.Month(), current.Day(), 8, 0, 0, 0, current.Location())
		timeFromStart := current.Sub(startOfDay)
		
		if remaining <= timeFromStart {
			// Subtract within current day
			current = current.Add(-remaining)
			remaining = 0
		} else {
			// Move to previous business day
			remaining -= timeFromStart
			current = current.AddDate(0, 0, -1)
			current = time.Date(current.Year(), current.Month(), current.Day(), 18, 0, 0, 0, current.Location())
		}
	}
	
	return current
}

// getTasksForTimeline retrieves tasks optimized for timeline calculation
func (s *Service) getTasksForTimeline(ctx context.Context, missionID uuid.UUID) ([]Task, error) {
	query := `
		SELECT t.id, t.mission_id, t.name, t.description, t.status, t.priority, t.assigned_to,
		       t.estimated_hours, t.actual_hours, t.due_date, t.completed_at, t.created_at, t.updated_at,
		       array_agg(td.depends_on_task_id) FILTER (WHERE td.depends_on_task_id IS NOT NULL) as dependencies
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
		var dependencyUUIDs []string
		
		err := rows.Scan(
			&task.ID, &task.MissionID, &task.Name, &task.Description, &task.Status, &task.Priority, &task.AssignedTo,
			&task.EstimatedHours, &task.ActualHours, &task.DueDate, &task.CompletedAt, &task.CreatedAt, &task.UpdatedAt,
			&dependencyUUIDs,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan task: %w", err)
		}
		
		// Parse dependencies
		if dependencyUUIDs != nil {
			for _, depStr := range dependencyUUIDs {
				if depStr != "" {
					depUUID, err := uuid.Parse(depStr)
					if err == nil {
						task.DependsOn = append(task.DependsOn, depUUID)
					}
				}
			}
		}
		
		tasks = append(tasks, task)
	}
	
	return tasks, rows.Err()
}

// CreateMilestone creates a new milestone for a mission
func (s *Service) CreateMilestone(ctx context.Context, req *CreateMilestoneRequest) (*Milestone, error) {
	userID := getUserIDFromContext(ctx)
	groupID := getGroupIDFromContext(ctx)
	
	if userID == "" {
		return nil, errors.New("user ID not found in context")
	}
	
	// Verify mission exists and user has access
	mission, err := s.getMissionFromDB(ctx, req.MissionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get mission: %w", err)
	}
	
	// Check group access
	if mission.GroupID != groupID {
		return nil, errors.New("insufficient permissions to create milestone")
	}
	
	milestone := &Milestone{
		ID:            uuid.New(),
		MissionID:     req.MissionID,
		Name:          req.Name,
		Description:   req.Description,
		MilestoneDate: req.MilestoneDate,
		Completed:     false,
		CreatedAt:     time.Now(),
	}
	
	// Insert milestone
	query := `
		INSERT INTO mission_milestones (id, mission_id, name, description, milestone_date, completed, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`
	
	_, err = s.db.ExecContext(ctx, query,
		milestone.ID, milestone.MissionID, milestone.Name, milestone.Description,
		milestone.MilestoneDate, milestone.Completed, milestone.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to insert milestone: %w", err)
	}
	
	// Publish milestone created event
	if s.publisher != nil {
		event := events.NewMilestoneCreatedEvent(userID, groupID, milestone.ID, milestone.MissionID, milestone)
		if err := s.publisher.PublishMilestoneEvent(ctx, event); err != nil {
			s.logger.Error().Err(err).Msg("Failed to publish milestone created event")
		}
	}
	
	s.logger.Info().
		Str("milestone_id", milestone.ID.String()).
		Str("mission_id", milestone.MissionID.String()).
		Str("user_id", userID).
		Str("milestone_name", milestone.Name).
		Msg("Milestone created successfully")
	
	return milestone, nil
}

// UpdateMilestone updates a milestone status
func (s *Service) UpdateMilestone(ctx context.Context, milestoneID uuid.UUID, completed bool) error {
	userID := getUserIDFromContext(ctx)
	groupID := getGroupIDFromContext(ctx)
	
	if userID == "" {
		return errors.New("user ID not found in context")
	}
	
	// Get milestone with mission info for permission checking
	query := `
		SELECT m.mission_id, ms.group_id 
		FROM mission_milestones m
		JOIN missions ms ON m.mission_id = ms.id
		WHERE m.id = $1`
	
	var missionID uuid.UUID
	var missionGroupID string
	err := s.db.QueryRowContext(ctx, query, milestoneID).Scan(&missionID, &missionGroupID)
	if err != nil {
		return fmt.Errorf("failed to get milestone: %w", err)
	}
	
	// Check group access
	if missionGroupID != groupID {
		return errors.New("insufficient permissions to update milestone")
	}
	
	// Update milestone
	updateQuery := `
		UPDATE mission_milestones 
		SET completed = $1, completed_at = CASE WHEN $1 THEN NOW() ELSE NULL END
		WHERE id = $2`
	
	_, err = s.db.ExecContext(ctx, updateQuery, completed, milestoneID)
	if err != nil {
		return fmt.Errorf("failed to update milestone: %w", err)
	}
	
	// Publish milestone completion event
	if s.publisher != nil {
		var completedAt *time.Time
		if completed {
			now := time.Now()
			completedAt = &now
		}
		event := events.NewMilestoneCompletedEvent(userID, groupID, milestoneID, missionID, completed, completedAt)
		if err := s.publisher.PublishMilestoneEvent(ctx, event); err != nil {
			s.logger.Error().Err(err).Msg("Failed to publish milestone completion event")
		}
	}
	
	s.logger.Info().
		Str("milestone_id", milestoneID.String()).
		Str("mission_id", missionID.String()).
		Str("user_id", userID).
		Bool("completed", completed).
		Msg("Milestone updated")
	
	return nil
}

// Request types for milestone management

// CreateMilestoneRequest represents a request to create a milestone
type CreateMilestoneRequest struct {
	MissionID     uuid.UUID `json:"mission_id" validate:"required"`
	Name          string    `json:"name" validate:"required,min=1,max=255"`
	Description   string    `json:"description"`
	MilestoneDate time.Time `json:"milestone_date" validate:"required"`
}

// UpdateMilestoneRequest represents a request to update a milestone
type UpdateMilestoneRequest struct {
	Name          *string    `json:"name,omitempty" validate:"omitempty,min=1,max=255"`
	Description   *string    `json:"description,omitempty"`
	MilestoneDate *time.Time `json:"milestone_date,omitempty"`
	Completed     *bool      `json:"completed,omitempty"`
}

// Additional field for TimelineTask to hold slack/float time
type TimelineTaskWithSlack struct {
	TimelineTask
	Slack time.Duration `json:"slack"`
}
