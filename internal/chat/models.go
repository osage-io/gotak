package chat

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// ChatRoomType represents the type of chat room
type ChatRoomType string

const (
	RoomTypeGroup      ChatRoomType = "group"
	RoomTypePrivate    ChatRoomType = "private"
	RoomTypeTactical   ChatRoomType = "tactical"
	RoomTypeEmergency  ChatRoomType = "emergency"
)

// Classification represents security classification levels
type Classification string

const (
	ClassificationUnclassified Classification = "UNCLASSIFIED"
	ClassificationRestricted   Classification = "RESTRICTED"
	ClassificationConfidential Classification = "CONFIDENTIAL"
	ClassificationSecret       Classification = "SECRET"
	ClassificationTopSecret    Classification = "TOP_SECRET"
)

// MessageType represents different types of chat messages
type MessageType string

const (
	MessageTypeText           MessageType = "text"
	MessageTypeSystem         MessageType = "system"
	MessageTypePosition       MessageType = "position"
	MessageTypeEmergency      MessageType = "emergency"
	MessageTypeTacticalReport MessageType = "tactical_report"
)

// MessagePriority represents message urgency levels
type MessagePriority string

const (
	PriorityLow       MessagePriority = "low"
	PriorityNormal    MessagePriority = "normal"
	PriorityHigh      MessagePriority = "high"
	PriorityUrgent    MessagePriority = "urgent"
	PriorityEmergency MessagePriority = "emergency"
)

// ParticipantRole represents user roles in chat rooms
type ParticipantRole string

const (
	RoleAdmin     ParticipantRole = "admin"
	RoleModerator ParticipantRole = "moderator"
	RoleMember    ParticipantRole = "member"
	RoleObserver  ParticipantRole = "observer"
)

// ReactionType represents different message reactions
type ReactionType string

const (
	ReactionRoger     ReactionType = "roger"     // Acknowledged/understood
	ReactionWilco     ReactionType = "wilco"     // Will comply
	ReactionNegative  ReactionType = "negative"  // Cannot comply/disagree
	ReactionLike      ReactionType = "like"      // General approval
	ReactionImportant ReactionType = "important" // Mark as important
	ReactionQuestion  ReactionType = "question"  // Needs clarification
)

// JSONB represents a PostgreSQL JSONB field
type JSONB map[string]interface{}

// Value implements the driver.Valuer interface for database storage
func (j JSONB) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// Scan implements the sql.Scanner interface for database retrieval
func (j *JSONB) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}

	return json.Unmarshal(bytes, j)
}

// ChatRoom represents a chat room/channel
type ChatRoom struct {
	ID               uuid.UUID        `json:"id" db:"id"`
	Name             string           `json:"name" db:"name"`
	Description      *string          `json:"description,omitempty" db:"description"`
	Type             ChatRoomType     `json:"type" db:"type"`
	Classification   Classification   `json:"classification" db:"classification"`
	CreatedBy        *uuid.UUID       `json:"createdBy,omitempty" db:"created_by"`
	CreatedAt        time.Time        `json:"createdAt" db:"created_at"`
	UpdatedAt        time.Time        `json:"updatedAt" db:"updated_at"`
	IsActive         bool             `json:"isActive" db:"is_active"`
	
	// Tactical fields
	MissionID        *uuid.UUID       `json:"missionId,omitempty" db:"mission_id"`
	GeographicBounds JSONB            `json:"geographicBounds,omitempty" db:"geographic_bounds"`
	
	// Room settings
	Settings         JSONB            `json:"settings" db:"settings"`
	
	// Related data (populated by joins)
	Participants     []ChatRoomParticipant `json:"participants,omitempty" db:"-"`
	ParticipantCount int                   `json:"participantCount" db:"participant_count"`
	LastMessage      *ChatMessage          `json:"lastMessage,omitempty" db:"-"`
	UnreadCount      int                   `json:"unreadCount" db:"unread_count"`
}

// ChatRoomParticipant represents a user's participation in a chat room
type ChatRoomParticipant struct {
	ID          uuid.UUID       `json:"id" db:"id"`
	RoomID      uuid.UUID       `json:"roomId" db:"room_id"`
	UserID      *uuid.UUID      `json:"userId,omitempty" db:"user_id"`
	Callsign    *string         `json:"callsign,omitempty" db:"callsign"`
	Role        ParticipantRole `json:"role" db:"role"`
	Permissions JSONB           `json:"permissions" db:"permissions"`
	JoinedAt    time.Time       `json:"joinedAt" db:"joined_at"`
	LastSeen    time.Time       `json:"lastSeen" db:"last_seen"`
	IsActive    bool            `json:"isActive" db:"is_active"`
	
	// User information (populated by joins)
	Username    *string         `json:"username,omitempty" db:"username"`
	DisplayName *string         `json:"displayName,omitempty" db:"display_name"`
}

// ChatMessage represents a chat message
type ChatMessage struct {
	ID               uuid.UUID        `json:"id" db:"id"`
	RoomID           uuid.UUID        `json:"roomId" db:"room_id"`
	SenderID         *uuid.UUID       `json:"senderId,omitempty" db:"sender_id"`
	SenderCallsign   *string          `json:"senderCallsign,omitempty" db:"sender_callsign"`
	
	// Message content
	MessageText      string           `json:"messageText" db:"message_text"`
	MessageType      MessageType      `json:"messageType" db:"message_type"`
	Priority         MessagePriority  `json:"priority" db:"priority"`
	
	// CoT metadata
	CotEventUID      *string          `json:"cotEventUid,omitempty" db:"cot_event_uid"`
	CotEventType     *string          `json:"cotEventType,omitempty" db:"cot_event_type"`
	CotRawXML        *string          `json:"cotRawXml,omitempty" db:"cot_raw_xml"`
	
	// Location data
	LocationLat      *float64         `json:"locationLat,omitempty" db:"location_lat"`
	LocationLng      *float64         `json:"locationLng,omitempty" db:"location_lng"`
	LocationAlt      *float64         `json:"locationAlt,omitempty" db:"location_alt"`
	
	// Security and metadata
	Classification   Classification   `json:"classification" db:"classification"`
	Keywords         []string         `json:"keywords,omitempty" db:"keywords"`
	
	// Threading
	ReplyToID        *uuid.UUID       `json:"replyToId,omitempty" db:"reply_to_id"`
	ThreadID         *uuid.UUID       `json:"threadId,omitempty" db:"thread_id"`
	
	// Status
	IsDeleted        bool             `json:"isDeleted" db:"is_deleted"`
	DeletedAt        *time.Time       `json:"deletedAt,omitempty" db:"deleted_at"`
	DeletedBy        *uuid.UUID       `json:"deletedBy,omitempty" db:"deleted_by"`
	
	// Timestamps
	CreatedAt        time.Time        `json:"createdAt" db:"created_at"`
	UpdatedAt        time.Time        `json:"updatedAt" db:"updated_at"`
	
	// Acknowledgment
	RequiresAck      bool             `json:"requiresAck" db:"requires_ack"`
	
	// Related data (populated by joins/queries)
	SenderUsername   *string          `json:"senderUsername,omitempty" db:"sender_username"`
	SenderDisplayName *string         `json:"senderDisplayName,omitempty" db:"sender_display_name"`
	ReplyTo          *ChatMessage     `json:"replyTo,omitempty" db:"-"`
	Reactions        []MessageReaction `json:"reactions,omitempty" db:"-"`
	Acknowledgments  []MessageAcknowledgment `json:"acknowledgments,omitempty" db:"-"`
	IsAcknowledged   bool             `json:"isAcknowledged" db:"is_acknowledged"`
}

// MessageAcknowledgment represents a user's acknowledgment of a message
type MessageAcknowledgment struct {
	ID             uuid.UUID `json:"id" db:"id"`
	MessageID      uuid.UUID `json:"messageId" db:"message_id"`
	UserID         uuid.UUID `json:"userId" db:"user_id"`
	AcknowledgedAt time.Time `json:"acknowledgedAt" db:"acknowledged_at"`
	
	// User info (populated by joins)
	Username    *string `json:"username,omitempty" db:"username"`
	Callsign    *string `json:"callsign,omitempty" db:"callsign"`
}

// MessageReaction represents a reaction to a message
type MessageReaction struct {
	ID           uuid.UUID    `json:"id" db:"id"`
	MessageID    uuid.UUID    `json:"messageId" db:"message_id"`
	UserID       uuid.UUID    `json:"userId" db:"user_id"`
	ReactionType ReactionType `json:"reactionType" db:"reaction_type"`
	CreatedAt    time.Time    `json:"createdAt" db:"created_at"`
	
	// User info (populated by joins)
	Username *string `json:"username,omitempty" db:"username"`
	Callsign *string `json:"callsign,omitempty" db:"callsign"`
}

// CreateChatRoomRequest represents a request to create a new chat room
type CreateChatRoomRequest struct {
	Name             string           `json:"name" validate:"required,max=255"`
	Description      *string          `json:"description,omitempty" validate:"max=1000"`
	Type             ChatRoomType     `json:"type" validate:"required,oneof=group private tactical emergency"`
	Classification   Classification   `json:"classification" validate:"required,oneof=UNCLASSIFIED RESTRICTED CONFIDENTIAL SECRET TOP_SECRET"`
	MissionID        *uuid.UUID       `json:"missionId,omitempty"`
	GeographicBounds JSONB            `json:"geographicBounds,omitempty"`
	Settings         JSONB            `json:"settings,omitempty"`
}

// SendMessageRequest represents a request to send a chat message
type SendMessageRequest struct {
	RoomID          uuid.UUID       `json:"roomId" validate:"required"`
	MessageText     string          `json:"messageText" validate:"required,max=4000"`
	MessageType     MessageType     `json:"messageType" validate:"oneof=text system position emergency tactical_report"`
	Priority        MessagePriority `json:"priority" validate:"oneof=low normal high urgent emergency"`
	Classification  Classification  `json:"classification" validate:"oneof=UNCLASSIFIED RESTRICTED CONFIDENTIAL SECRET TOP_SECRET"`
	Keywords        []string        `json:"keywords,omitempty"`
	LocationLat     *float64        `json:"locationLat,omitempty" validate:"min=-90,max=90"`
	LocationLng     *float64        `json:"locationLng,omitempty" validate:"min=-180,max=180"`
	LocationAlt     *float64        `json:"locationAlt,omitempty"`
	ReplyToID       *uuid.UUID      `json:"replyToId,omitempty"`
	RequiresAck     bool            `json:"requiresAck"`
}

// GetMessagesRequest represents a request to retrieve messages
type GetMessagesRequest struct {
	RoomID        uuid.UUID        `json:"roomId" validate:"required"`
	Limit         int              `json:"limit" validate:"min=1,max=100"`
	Offset        int              `json:"offset" validate:"min=0"`
	BeforeID      *uuid.UUID       `json:"beforeId,omitempty"`
	AfterID       *uuid.UUID       `json:"afterId,omitempty"`
	MessageType   *MessageType     `json:"messageType,omitempty"`
	Priority      *MessagePriority `json:"priority,omitempty"`
	Classification *Classification `json:"classification,omitempty"`
	Keywords      []string         `json:"keywords,omitempty"`
	StartTime     *time.Time       `json:"startTime,omitempty"`
	EndTime       *time.Time       `json:"endTime,omitempty"`
}

// ChatStatistics represents chat system statistics
type ChatStatistics struct {
	TotalRooms        int `json:"totalRooms"`
	ActiveRooms       int `json:"activeRooms"`
	TotalMessages     int `json:"totalMessages"`
	MessagesToday     int `json:"messagesToday"`
	ActiveUsers       int `json:"activeUsers"`
	EmergencyMessages int `json:"emergencyMessages"`
	UnacknowledgedCount int `json:"unacknowledgedCount"`
}

// Validation methods

// Validate validates ChatRoomType
func (c ChatRoomType) Validate() error {
	switch c {
	case RoomTypeGroup, RoomTypePrivate, RoomTypeTactical, RoomTypeEmergency:
		return nil
	default:
		return fmt.Errorf("invalid chat room type: %s", c)
	}
}

// Validate validates Classification
func (c Classification) Validate() error {
	switch c {
	case ClassificationUnclassified, ClassificationRestricted, ClassificationConfidential, ClassificationSecret, ClassificationTopSecret:
		return nil
	default:
		return fmt.Errorf("invalid classification: %s", c)
	}
}

// Validate validates MessageType
func (m MessageType) Validate() error {
	switch m {
	case MessageTypeText, MessageTypeSystem, MessageTypePosition, MessageTypeEmergency, MessageTypeTacticalReport:
		return nil
	default:
		return fmt.Errorf("invalid message type: %s", m)
	}
}

// Validate validates MessagePriority
func (m MessagePriority) Validate() error {
	switch m {
	case PriorityLow, PriorityNormal, PriorityHigh, PriorityUrgent, PriorityEmergency:
		return nil
	default:
		return fmt.Errorf("invalid message priority: %s", m)
	}
}

// Validate validates ParticipantRole
func (p ParticipantRole) Validate() error {
	switch p {
	case RoleAdmin, RoleModerator, RoleMember, RoleObserver:
		return nil
	default:
		return fmt.Errorf("invalid participant role: %s", p)
	}
}

// Validate validates ReactionType
func (r ReactionType) Validate() error {
	switch r {
	case ReactionRoger, ReactionWilco, ReactionNegative, ReactionLike, ReactionImportant, ReactionQuestion:
		return nil
	default:
		return fmt.Errorf("invalid reaction type: %s", r)
	}
}
