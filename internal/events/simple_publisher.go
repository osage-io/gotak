package events

import (
	"context"
	"github.com/dfedick/gotak/pkg/logger"
)

// SimplePublisher is a basic in-memory event publisher for development
type SimplePublisher struct {
	logger *logger.Logger
}

// NewSimplePublisher creates a new simple event publisher
func NewSimplePublisher(logger *logger.Logger) *SimplePublisher {
	return &SimplePublisher{
		logger: logger,
	}
}

// PublishMissionEvent logs a mission event
func (p *SimplePublisher) PublishMissionEvent(ctx context.Context, event *MissionEvent) error {
	p.logger.Debug().
		Str("event_type", event.Type).
		Str("mission_id", event.MissionID.String()).
		Str("user_id", event.UserID).
		Msg("Mission event published")
	return nil
}

// PublishTaskEvent logs a task event
func (p *SimplePublisher) PublishTaskEvent(ctx context.Context, event *TaskEvent) error {
	p.logger.Debug().
		Str("event_type", event.Type).
		Str("task_id", event.TaskID.String()).
		Str("mission_id", event.MissionID.String()).
		Str("user_id", event.UserID).
		Msg("Task event published")
	return nil
}

// PublishMilestoneEvent logs a milestone event
func (p *SimplePublisher) PublishMilestoneEvent(ctx context.Context, event *MilestoneEvent) error {
	p.logger.Debug().
		Str("event_type", event.Type).
		Str("milestone_id", event.MilestoneID.String()).
		Str("mission_id", event.MissionID.String()).
		Str("user_id", event.UserID).
		Msg("Milestone event published")
	return nil
}

// Close is a no-op for the simple publisher
func (p *SimplePublisher) Close() error {
	return nil
}