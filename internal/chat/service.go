package chat

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"

	"github.com/dfedick/gotak/pkg/cot"
	"github.com/dfedick/gotak/pkg/logger"
)

// Service handles chat operations
type Service struct {
	db     *sqlx.DB
	logger *logger.Logger
}

// NewService creates a new chat service
func NewService(db *sqlx.DB, logger *logger.Logger) *Service {
	return &Service{
		db:     db,
		logger: logger,
	}
}

// CreateRoom creates a new chat room
func (s *Service) CreateRoom(ctx context.Context, req CreateChatRoomRequest, createdBy uuid.UUID) (*ChatRoom, error) {
	room := &ChatRoom{
		ID:             uuid.New(),
		Name:           req.Name,
		Description:    req.Description,
		Type:           req.Type,
		Classification: req.Classification,
		CreatedBy:      &createdBy,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		IsActive:       true,
		MissionID:      req.MissionID,
		GeographicBounds: req.GeographicBounds,
		Settings:       req.Settings,
	}

	// Set default settings if none provided
	if room.Settings == nil {
		room.Settings = JSONB{
			"maxParticipants":     100,
			"allowFileUploads":    false,
			"requireAcknowledgment": false,
		}
	}

	query := `
		INSERT INTO chat_rooms (id, name, description, type, classification, created_by, 
			created_at, updated_at, is_active, mission_id, geographic_bounds, settings)
		VALUES (:id, :name, :description, :type, :classification, :created_by,
			:created_at, :updated_at, :is_active, :mission_id, :geographic_bounds, :settings)
		RETURNING *`

	var result ChatRoom
	rows, err := s.db.NamedQueryContext(ctx, query, room)
	if err != nil {
		s.logger.Error().Err(err).Str("room_name", req.Name).Msg("Failed to create chat room")
		return nil, fmt.Errorf("failed to create chat room: %w", err)
	}
	defer rows.Close()

	if rows.Next() {
		if err := rows.StructScan(&result); err != nil {
			return nil, fmt.Errorf("failed to scan chat room: %w", err)
		}
	}

	// Add creator as admin participant
	_, err = s.AddParticipant(ctx, result.ID, createdBy, "", RoleAdmin)
	if err != nil {
		s.logger.Warn().Err(err).Str("room_id", result.ID.String()).Msg("Failed to add creator as participant")
	}

	s.logger.Info().
		Str("room_id", result.ID.String()).
		Str("room_name", result.Name).
		Str("room_type", string(result.Type)).
		Str("created_by", createdBy.String()).
		Msg("Chat room created successfully")

	return &result, nil
}

// GetRoom retrieves a chat room by ID
func (s *Service) GetRoom(ctx context.Context, roomID uuid.UUID, userID *uuid.UUID) (*ChatRoom, error) {
	query := `
		SELECT r.*, 
			COALESCE(COUNT(DISTINCT p.id), 0) as participant_count,
			COALESCE(COUNT(DISTINCT CASE WHEN m.requires_ack = true AND ma.id IS NULL THEN m.id END), 0) as unread_count
		FROM chat_rooms r
		LEFT JOIN chat_room_participants p ON r.id = p.room_id AND p.is_active = true
		LEFT JOIN chat_messages m ON r.id = m.room_id AND m.is_deleted = false AND m.requires_ack = true
		LEFT JOIN message_acknowledgments ma ON m.id = ma.message_id AND ma.user_id = $2
		WHERE r.id = $1 AND r.is_active = true
		GROUP BY r.id`

	var room ChatRoom
	err := s.db.GetContext(ctx, &room, query, roomID, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("chat room not found")
		}
		return nil, fmt.Errorf("failed to get chat room: %w", err)
	}

	// Get participants
	participants, err := s.GetParticipants(ctx, roomID)
	if err == nil {
		room.Participants = participants
	}

	// Get last message
	lastMessage, err := s.getLastMessage(ctx, roomID)
	if err == nil && lastMessage != nil {
		room.LastMessage = lastMessage
	}

	return &room, nil
}

// GetRooms retrieves chat rooms for a user
func (s *Service) GetRooms(ctx context.Context, userID uuid.UUID, roomType *ChatRoomType) ([]ChatRoom, error) {
	query := `
		SELECT r.*, 
			COALESCE(COUNT(DISTINCT p.id), 0) as participant_count,
			COALESCE(COUNT(DISTINCT CASE WHEN m.requires_ack = true AND ma.id IS NULL AND m.sender_id != $1 THEN m.id END), 0) as unread_count
		FROM chat_rooms r
		LEFT JOIN chat_room_participants rp ON r.id = rp.room_id AND rp.user_id = $1 AND rp.is_active = true
		LEFT JOIN chat_room_participants p ON r.id = p.room_id AND p.is_active = true
		LEFT JOIN chat_messages m ON r.id = m.room_id AND m.is_deleted = false AND m.created_at > rp.last_seen
		LEFT JOIN message_acknowledgments ma ON m.id = ma.message_id AND ma.user_id = $1
		WHERE r.is_active = true 
			AND (rp.user_id IS NOT NULL OR r.type = 'group')
			AND ($2::text IS NULL OR r.type = $2)
		GROUP BY r.id, rp.last_seen
		ORDER BY r.updated_at DESC`

	var rooms []ChatRoom
	var roomTypeStr *string
	if roomType != nil {
		str := string(*roomType)
		roomTypeStr = &str
	}

	err := s.db.SelectContext(ctx, &rooms, query, userID, roomTypeStr)
	if err != nil {
		return nil, fmt.Errorf("failed to get chat rooms: %w", err)
	}

	// Get last message for each room
	for i, room := range rooms {
		lastMessage, err := s.getLastMessage(ctx, room.ID)
		if err == nil && lastMessage != nil {
			rooms[i].LastMessage = lastMessage
		}
	}

	return rooms, nil
}

// AddParticipant adds a user to a chat room
func (s *Service) AddParticipant(ctx context.Context, roomID, userID uuid.UUID, callsign string, role ParticipantRole) (*ChatRoomParticipant, error) {
	participant := &ChatRoomParticipant{
		ID:       uuid.New(),
		RoomID:   roomID,
		UserID:   &userID,
		Role:     role,
		JoinedAt: time.Now(),
		LastSeen: time.Now(),
		IsActive: true,
		Permissions: JSONB{},
	}

	if callsign != "" {
		participant.Callsign = &callsign
	}

	query := `
		INSERT INTO chat_room_participants (id, room_id, user_id, callsign, role, permissions, joined_at, last_seen, is_active)
		VALUES (:id, :room_id, :user_id, :callsign, :role, :permissions, :joined_at, :last_seen, :is_active)
		ON CONFLICT (room_id, user_id) DO UPDATE SET
			is_active = true,
			last_seen = NOW()
		RETURNING *`

	var result ChatRoomParticipant
	rows, err := s.db.NamedQueryContext(ctx, query, participant)
	if err != nil {
		return nil, fmt.Errorf("failed to add participant: %w", err)
	}
	defer rows.Close()

	if rows.Next() {
		if err := rows.StructScan(&result); err != nil {
			return nil, fmt.Errorf("failed to scan participant: %w", err)
		}
	}

	s.logger.Info().
		Str("room_id", roomID.String()).
		Str("user_id", userID.String()).
		Str("role", string(role)).
		Msg("Participant added to chat room")

	return &result, nil
}

// GetParticipants retrieves participants of a chat room
func (s *Service) GetParticipants(ctx context.Context, roomID uuid.UUID) ([]ChatRoomParticipant, error) {
	query := `
		SELECT p.*, u.username, u.display_name
		FROM chat_room_participants p
		LEFT JOIN auth_users u ON p.user_id = u.id
		WHERE p.room_id = $1 AND p.is_active = true
		ORDER BY p.role ASC, p.joined_at ASC`

	var participants []ChatRoomParticipant
	err := s.db.SelectContext(ctx, &participants, query, roomID)
	if err != nil {
		return nil, fmt.Errorf("failed to get participants: %w", err)
	}

	return participants, nil
}

// RemoveParticipant removes a user from a chat room
func (s *Service) RemoveParticipant(ctx context.Context, roomID, userID uuid.UUID) error {
	query := `
		UPDATE chat_room_participants 
		SET is_active = false, left_at = NOW()
		WHERE room_id = $1 AND user_id = $2 AND is_active = true`

	_, err := s.db.ExecContext(ctx, query, roomID, userID)
	if err != nil {
		return fmt.Errorf("failed to remove participant: %w", err)
	}

	s.logger.Info().
		Str("room_id", roomID.String()).
		Str("user_id", userID.String()).
		Msg("Participant removed from chat room")

	return nil
}

// SendMessage sends a message to a chat room
func (s *Service) SendMessage(ctx context.Context, req SendMessageRequest, senderID *uuid.UUID, senderCallsign string) (*ChatMessage, error) {
	message := &ChatMessage{
		ID:             uuid.New(),
		RoomID:         req.RoomID,
		SenderID:       senderID,
		MessageText:    req.MessageText,
		MessageType:    req.MessageType,
		Priority:       req.Priority,
		Classification: req.Classification,
		Keywords:       req.Keywords,
		LocationLat:    req.LocationLat,
		LocationLng:    req.LocationLng,
		LocationAlt:    req.LocationAlt,
		ReplyToID:      req.ReplyToID,
		RequiresAck:    req.RequiresAck,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		IsDeleted:      false,
	}

	if senderCallsign != "" {
		message.SenderCallsign = &senderCallsign
	}

	// Set default values
	if message.MessageType == "" {
		message.MessageType = MessageTypeText
	}
	if message.Priority == "" {
		message.Priority = PriorityNormal
	}
	if message.Classification == "" {
		message.Classification = ClassificationUnclassified
	}

	// Generate thread ID for replies
	if req.ReplyToID != nil {
		// Get the thread ID from the parent message or use the parent's ID as thread ID
		var threadID *uuid.UUID
		err := s.db.GetContext(ctx, &threadID, 
			"SELECT COALESCE(thread_id, id) FROM chat_messages WHERE id = $1", req.ReplyToID)
		if err == nil && threadID != nil {
			message.ThreadID = threadID
		}
	}

	query := `
		INSERT INTO chat_messages (id, room_id, sender_id, sender_callsign, message_text, 
			message_type, priority, classification, keywords, location_lat, location_lng, 
			location_alt, reply_to_id, thread_id, requires_ack, created_at, updated_at, is_deleted)
		VALUES (:id, :room_id, :sender_id, :sender_callsign, :message_text, :message_type, 
			:priority, :classification, :keywords, :location_lat, :location_lng, :location_alt, 
			:reply_to_id, :thread_id, :requires_ack, :created_at, :updated_at, :is_deleted)
		RETURNING *`

	var result ChatMessage
	rows, err := s.db.NamedQueryContext(ctx, query, message)
	if err != nil {
		s.logger.Error().Err(err).Str("room_id", req.RoomID.String()).Msg("Failed to send message")
		return nil, fmt.Errorf("failed to send message: %w", err)
	}
	defer rows.Close()

	if rows.Next() {
		if err := rows.StructScan(&result); err != nil {
			return nil, fmt.Errorf("failed to scan message: %w", err)
		}
	}

	// Update room updated_at timestamp
	_, err = s.db.ExecContext(ctx, "UPDATE chat_rooms SET updated_at = NOW() WHERE id = $1", req.RoomID)
	if err != nil {
		s.logger.Warn().Err(err).Str("room_id", req.RoomID.String()).Msg("Failed to update room timestamp")
	}

	s.logger.Info().
		Str("message_id", result.ID.String()).
		Str("room_id", req.RoomID.String()).
		Str("message_type", string(result.MessageType)).
		Str("priority", string(result.Priority)).
		Msg("Message sent successfully")

	return &result, nil
}

// GetMessages retrieves messages from a chat room
func (s *Service) GetMessages(ctx context.Context, req GetMessagesRequest, userID *uuid.UUID) ([]ChatMessage, error) {
	var conditions []string
	var args []interface{}
	argIndex := 1

	// Base conditions
	conditions = append(conditions, "m.room_id = $"+fmt.Sprint(argIndex))
	args = append(args, req.RoomID)
	argIndex++

	conditions = append(conditions, "m.is_deleted = false")

	// Optional filters
	if req.BeforeID != nil {
		conditions = append(conditions, "m.created_at < (SELECT created_at FROM chat_messages WHERE id = $"+fmt.Sprint(argIndex)+")")
		args = append(args, *req.BeforeID)
		argIndex++
	}

	if req.AfterID != nil {
		conditions = append(conditions, "m.created_at > (SELECT created_at FROM chat_messages WHERE id = $"+fmt.Sprint(argIndex)+")")
		args = append(args, *req.AfterID)
		argIndex++
	}

	if req.MessageType != nil {
		conditions = append(conditions, "m.message_type = $"+fmt.Sprint(argIndex))
		args = append(args, string(*req.MessageType))
		argIndex++
	}

	if req.Priority != nil {
		conditions = append(conditions, "m.priority = $"+fmt.Sprint(argIndex))
		args = append(args, string(*req.Priority))
		argIndex++
	}

	if req.Classification != nil {
		conditions = append(conditions, "m.classification = $"+fmt.Sprint(argIndex))
		args = append(args, string(*req.Classification))
		argIndex++
	}

	if len(req.Keywords) > 0 {
		conditions = append(conditions, "m.keywords && $"+fmt.Sprint(argIndex))
		args = append(args, pq.Array(req.Keywords))
		argIndex++
	}

	if req.StartTime != nil {
		conditions = append(conditions, "m.created_at >= $"+fmt.Sprint(argIndex))
		args = append(args, *req.StartTime)
		argIndex++
	}

	if req.EndTime != nil {
		conditions = append(conditions, "m.created_at <= $"+fmt.Sprint(argIndex))
		args = append(args, *req.EndTime)
		argIndex++
	}

	// Build query
	whereClause := strings.Join(conditions, " AND ")
	
	query := fmt.Sprintf(`
		SELECT m.*, 
			u.username as sender_username,
			u.display_name as sender_display_name,
			CASE WHEN ma.id IS NOT NULL THEN true ELSE false END as is_acknowledged
		FROM chat_messages m
		LEFT JOIN auth_users u ON m.sender_id = u.id
		LEFT JOIN message_acknowledgments ma ON m.id = ma.message_id AND ma.user_id = $%d
		WHERE %s
		ORDER BY m.created_at DESC
		LIMIT $%d OFFSET $%d`,
		argIndex, whereClause, argIndex+1, argIndex+2)

	args = append(args, userID, req.Limit, req.Offset)

	var messages []ChatMessage
	err := s.db.SelectContext(ctx, &messages, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get messages: %w", err)
	}

	// Load reactions and reply-to messages for each message
	for i := range messages {
		// Load reactions
		reactions, err := s.getMessageReactions(ctx, messages[i].ID)
		if err == nil {
			messages[i].Reactions = reactions
		}

		// Load reply-to message
		if messages[i].ReplyToID != nil {
			replyTo, err := s.getMessageByID(ctx, *messages[i].ReplyToID)
			if err == nil {
				messages[i].ReplyTo = replyTo
			}
		}
	}

	return messages, nil
}

// AcknowledgeMessage marks a message as acknowledged by a user
func (s *Service) AcknowledgeMessage(ctx context.Context, messageID, userID uuid.UUID) error {
	query := `
		INSERT INTO message_acknowledgments (id, message_id, user_id, acknowledged_at)
		VALUES ($1, $2, $3, NOW())
		ON CONFLICT (message_id, user_id) DO NOTHING`

	_, err := s.db.ExecContext(ctx, query, uuid.New(), messageID, userID)
	if err != nil {
		return fmt.Errorf("failed to acknowledge message: %w", err)
	}

	s.logger.Debug().
		Str("message_id", messageID.String()).
		Str("user_id", userID.String()).
		Msg("Message acknowledged")

	return nil
}

// AddReaction adds a reaction to a message
func (s *Service) AddReaction(ctx context.Context, messageID, userID uuid.UUID, reactionType ReactionType) error {
	query := `
		INSERT INTO message_reactions (id, message_id, user_id, reaction_type, created_at)
		VALUES ($1, $2, $3, $4, NOW())
		ON CONFLICT (message_id, user_id, reaction_type) DO NOTHING`

	_, err := s.db.ExecContext(ctx, query, uuid.New(), messageID, userID, string(reactionType))
	if err != nil {
		return fmt.Errorf("failed to add reaction: %w", err)
	}

	return nil
}

// CreateChatMessageFromCoT creates a chat message from a CoT event
func (s *Service) CreateChatMessageFromCoT(ctx context.Context, event *cot.Event, roomID uuid.UUID, senderCallsign string) (*ChatMessage, error) {
	messageText := ""
	if event.Detail != nil && event.Detail.Remarks != nil {
		messageText = event.Detail.Remarks.Text
	}

	if messageText == "" {
		messageText = fmt.Sprintf("CoT %s event from %s", event.Type, senderCallsign)
	}

	// Determine message type and priority from CoT type
	var messageType MessageType = MessageTypeText
	var priority MessagePriority = PriorityNormal

	if strings.Contains(event.Type, "emergency") {
		messageType = MessageTypeEmergency
		priority = PriorityEmergency
	} else if cot.IsTypeSystem(event.Type) {
		messageType = MessageTypeSystem
	}

	// Get location if available
	var lat, lng *float64
	if latVal, lonVal, err := event.GetPosition(); err == nil {
		lat = &latVal
		lng = &lonVal
	}

	// Convert CoT to raw XML
	xmlData, err := event.ToXML()
	var rawXML *string
	if err == nil {
		str := string(xmlData)
		rawXML = &str
	}

	req := SendMessageRequest{
		RoomID:         roomID,
		MessageText:    messageText,
		MessageType:    messageType,
		Priority:       priority,
		Classification: ClassificationUnclassified,
		LocationLat:    lat,
		LocationLng:    lng,
		RequiresAck:    priority >= PriorityUrgent,
	}

	message, err := s.SendMessage(ctx, req, nil, senderCallsign)
	if err != nil {
		return nil, err
	}

	// Update with CoT-specific fields
	updateQuery := `
		UPDATE chat_messages 
		SET cot_event_uid = $1, cot_event_type = $2, cot_raw_xml = $3
		WHERE id = $4`

	_, err = s.db.ExecContext(ctx, updateQuery, event.UID, event.Type, rawXML, message.ID)
	if err != nil {
		s.logger.Warn().Err(err).Str("message_id", message.ID.String()).Msg("Failed to update CoT metadata")
	}

	return message, nil
}

// Helper functions

func (s *Service) getLastMessage(ctx context.Context, roomID uuid.UUID) (*ChatMessage, error) {
	query := `
		SELECT m.*, u.username as sender_username, u.display_name as sender_display_name
		FROM chat_messages m
		LEFT JOIN auth_users u ON m.sender_id = u.id
		WHERE m.room_id = $1 AND m.is_deleted = false
		ORDER BY m.created_at DESC
		LIMIT 1`

	var message ChatMessage
	err := s.db.GetContext(ctx, &message, query, roomID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &message, nil
}

func (s *Service) getMessageByID(ctx context.Context, messageID uuid.UUID) (*ChatMessage, error) {
	query := `
		SELECT m.*, u.username as sender_username, u.display_name as sender_display_name
		FROM chat_messages m
		LEFT JOIN auth_users u ON m.sender_id = u.id
		WHERE m.id = $1 AND m.is_deleted = false`

	var message ChatMessage
	err := s.db.GetContext(ctx, &message, query, messageID)
	if err != nil {
		return nil, err
	}

	return &message, nil
}

func (s *Service) getMessageReactions(ctx context.Context, messageID uuid.UUID) ([]MessageReaction, error) {
	query := `
		SELECT mr.*, u.username, p.callsign
		FROM message_reactions mr
		LEFT JOIN auth_users u ON mr.user_id = u.id
		LEFT JOIN chat_room_participants p ON mr.user_id = p.user_id
		WHERE mr.message_id = $1
		ORDER BY mr.created_at ASC`

	var reactions []MessageReaction
	err := s.db.SelectContext(ctx, &reactions, query, messageID)
	if err != nil {
		return nil, err
	}

	return reactions, nil
}

// GetStatistics returns chat system statistics
func (s *Service) GetStatistics(ctx context.Context) (*ChatStatistics, error) {
	var stats ChatStatistics

	// Get room counts
	err := s.db.GetContext(ctx, &stats.TotalRooms, "SELECT COUNT(*) FROM chat_rooms")
	if err != nil {
		return nil, fmt.Errorf("failed to get room count: %w", err)
	}

	err = s.db.GetContext(ctx, &stats.ActiveRooms, "SELECT COUNT(*) FROM chat_rooms WHERE is_active = true")
	if err != nil {
		return nil, fmt.Errorf("failed to get active room count: %w", err)
	}

	// Get message counts
	err = s.db.GetContext(ctx, &stats.TotalMessages, "SELECT COUNT(*) FROM chat_messages WHERE is_deleted = false")
	if err != nil {
		return nil, fmt.Errorf("failed to get message count: %w", err)
	}

	err = s.db.GetContext(ctx, &stats.MessagesToday, 
		"SELECT COUNT(*) FROM chat_messages WHERE is_deleted = false AND created_at >= CURRENT_DATE")
	if err != nil {
		return nil, fmt.Errorf("failed to get today's message count: %w", err)
	}

	// Get emergency message count
	err = s.db.GetContext(ctx, &stats.EmergencyMessages,
		"SELECT COUNT(*) FROM chat_messages WHERE message_type = 'emergency' AND is_deleted = false")
	if err != nil {
		return nil, fmt.Errorf("failed to get emergency message count: %w", err)
	}

	// Get unacknowledged message count
	err = s.db.GetContext(ctx, &stats.UnacknowledgedCount,
		`SELECT COUNT(*) FROM chat_messages m 
		 WHERE m.requires_ack = true AND m.is_deleted = false 
		 AND NOT EXISTS (SELECT 1 FROM message_acknowledgments ma WHERE ma.message_id = m.id)`)
	if err != nil {
		return nil, fmt.Errorf("failed to get unacknowledged count: %w", err)
	}

	// Get active user count (users with activity in last 24 hours)
	err = s.db.GetContext(ctx, &stats.ActiveUsers,
		"SELECT COUNT(DISTINCT user_id) FROM chat_room_participants WHERE last_seen >= NOW() - INTERVAL '24 hours'")
	if err != nil {
		return nil, fmt.Errorf("failed to get active user count: %w", err)
	}

	return &stats, nil
}
