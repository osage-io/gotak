package mission

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

// CreateObjective creates a new objective for a mission
func (s *Service) CreateObjective(ctx context.Context, missionID uuid.UUID, req *CreateObjectiveRequest) (*Objective, error) {
	userID := getUserIDFromContext(ctx)
	groupID := getGroupIDFromContext(ctx)
	
	if userID == "" {
		return nil, errors.New("user ID not found in context")
	}
	
	// Verify mission exists and user has access
	mission, err := s.getMissionFromDB(ctx, missionID)
	if err != nil {
		return nil, errors.New("mission not found")
	}
	
	if mission.GroupID != groupID {
		return nil, errors.New("insufficient permissions")
	}
	
	// Create objective
	objective := &Objective{
		ID:          uuid.New(),
		MissionID:   missionID,
		Description: req.Description,
		Priority:    req.Priority,
		Completed:   false,
		CreatedAt:   time.Now(),
	}
	
	// Insert into database
	query := `
		INSERT INTO mission_objectives (id, mission_id, description, priority, completed, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)`
	
	_, err = s.db.ExecContext(ctx, query, 
		objective.ID, objective.MissionID, objective.Description,
		objective.Priority, objective.Completed, objective.CreatedAt)
	if err != nil {
		return nil, err
	}
	
	s.logger.Info().
		Str("objective_id", objective.ID.String()).
		Str("mission_id", missionID.String()).
		Msg("Objective created")
		
	return objective, nil
}

// UpdateObjective updates an existing objective
func (s *Service) UpdateObjective(ctx context.Context, objectiveID uuid.UUID, description string, priority int) (*Objective, error) {
	userID := getUserIDFromContext(ctx)
	groupID := getGroupIDFromContext(ctx)
	
	if userID == "" {
		return nil, errors.New("user ID not found in context")
	}
	
	// Get objective with mission info for permission check
	query := `
		SELECT o.id, o.mission_id, o.description, o.priority, o.completed, o.completed_at, o.created_at, m.group_id
		FROM mission_objectives o
		JOIN missions m ON o.mission_id = m.id
		WHERE o.id = $1`
		
	var objective Objective
	var missionGroupID string
	err := s.db.QueryRowContext(ctx, query, objectiveID).Scan(
		&objective.ID, &objective.MissionID, &objective.Description,
		&objective.Priority, &objective.Completed, &objective.CompletedAt,
		&objective.CreatedAt, &missionGroupID)
	if err != nil {
		return nil, errors.New("objective not found")
	}
	
	if missionGroupID != groupID {
		return nil, errors.New("insufficient permissions")
	}
	
	// Update objective
	updateQuery := `
		UPDATE mission_objectives
		SET description = $1, priority = $2, updated_at = NOW()
		WHERE id = $3`
		
	_, err = s.db.ExecContext(ctx, updateQuery, description, priority, objectiveID)
	if err != nil {
		return nil, err
	}
	
	objective.Description = description
	objective.Priority = priority
	
	return &objective, nil
}

// DeleteObjective deletes an objective
func (s *Service) DeleteObjective(ctx context.Context, objectiveID uuid.UUID) error {
	userID := getUserIDFromContext(ctx)
	groupID := getGroupIDFromContext(ctx)
	
	if userID == "" {
		return errors.New("user ID not found in context")
	}
	
	// Check permissions
	query := `
		SELECT m.group_id
		FROM mission_objectives o
		JOIN missions m ON o.mission_id = m.id
		WHERE o.id = $1`
		
	var missionGroupID string
	err := s.db.QueryRowContext(ctx, query, objectiveID).Scan(&missionGroupID)
	if err != nil {
		return errors.New("objective not found")
	}
	
	if missionGroupID != groupID {
		return errors.New("insufficient permissions")
	}
	
	// Delete objective
	deleteQuery := "DELETE FROM mission_objectives WHERE id = $1"
	_, err = s.db.ExecContext(ctx, deleteQuery, objectiveID)
	if err != nil {
		return err
	}
	
	return nil
}

// CompleteObjective marks an objective as complete
func (s *Service) CompleteObjective(ctx context.Context, objectiveID uuid.UUID) (*Objective, error) {
	userID := getUserIDFromContext(ctx)
	groupID := getGroupIDFromContext(ctx)
	
	if userID == "" {
		return nil, errors.New("user ID not found in context")
	}
	
	// Get objective with mission info for permission check
	query := `
		SELECT o.id, o.mission_id, o.description, o.priority, o.completed, o.completed_at, o.created_at, m.group_id
		FROM mission_objectives o
		JOIN missions m ON o.mission_id = m.id
		WHERE o.id = $1`
		
	var objective Objective
	var missionGroupID string
	err := s.db.QueryRowContext(ctx, query, objectiveID).Scan(
		&objective.ID, &objective.MissionID, &objective.Description,
		&objective.Priority, &objective.Completed, &objective.CompletedAt,
		&objective.CreatedAt, &missionGroupID)
	if err != nil {
		return nil, errors.New("objective not found")
	}
	
	if missionGroupID != groupID {
		return nil, errors.New("insufficient permissions")
	}
	
	if objective.Completed {
		return &objective, nil // Already completed
	}
	
	// Mark as complete
	now := time.Now()
	updateQuery := `
		UPDATE mission_objectives
		SET completed = true, completed_at = $1, updated_at = $1
		WHERE id = $2`
		
	_, err = s.db.ExecContext(ctx, updateQuery, now, objectiveID)
	if err != nil {
		return nil, err
	}
	
	objective.Completed = true
	objective.CompletedAt = &now
	
	return &objective, nil
}

// GetMissionParticipants gets participants for a mission
func (s *Service) GetMissionParticipants(ctx context.Context, missionID uuid.UUID) ([]interface{}, error) {
	// Stub implementation - return empty slice for now
	return []interface{}{}, nil
}

// GetMissionEvents gets events for a mission  
func (s *Service) GetMissionEvents(ctx context.Context, missionID uuid.UUID) ([]interface{}, error) {
	// Stub implementation - return empty slice for now
	return []interface{}{}, nil
}
