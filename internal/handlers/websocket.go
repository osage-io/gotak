package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/dfedick/gotak/internal/chat"
	"github.com/dfedick/gotak/pkg/logger"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// WebSocket upgrader with CORS support.
// Allowed origins come from GOTAK_WS_ALLOWED_ORIGINS (comma-separated); localhost
// dev origins are always allowed, as is same-origin (Origin host == request Host)
// and a missing Origin (non-browser clients). This lets the deployed UI
// (e.g. https://gotak.demoland.io, fronted by the gateway) connect.
var wsAllowedOrigins = func() map[string]bool {
	m := map[string]bool{
		"http://localhost:5173": true,
		"http://localhost:3000": true,
	}
	for _, o := range strings.Split(os.Getenv("GOTAK_WS_ALLOWED_ORIGINS"), ",") {
		if o = strings.TrimSpace(o); o != "" {
			m[o] = true
		}
	}
	return m
}()

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		if origin == "" { // non-browser client, no Origin header
			return true
		}
		if wsAllowedOrigins[origin] {
			return true
		}
		// Same-origin: the Origin's host matches the request Host (covers the
		// gateway-fronted deployment without hardcoding the hostname).
		if u, err := url.Parse(origin); err == nil && u.Host == r.Host {
			return true
		}
		return false
	},
}

// TacticalWSMessage represents a message sent to tactical WebSocket clients
type TacticalWSMessage struct {
	Type      string      `json:"type"`
	Payload   interface{} `json:"payload"`
	Timestamp time.Time   `json:"timestamp"`
	RoomID    *uuid.UUID  `json:"roomId,omitempty"` // For chat messages
}

// WebSocket message types
const (
	// Position and entity messages
	MsgTypePositionUpdate = "position_update"
	MsgTypeEntityRemoved  = "entity_removed"

	// Chat messages
	MsgTypeChatMessage     = "chat_message"
	MsgTypeChatRoomUpdate  = "chat_room_update"
	MsgTypeChatRoomJoined  = "chat_room_joined"
	MsgTypeChatRoomLeft    = "chat_room_left"
	MsgTypeMessageReaction = "message_reaction"
	MsgTypeMessageAck      = "message_acknowledgment"
	MsgTypeUserTyping      = "user_typing"
	MsgTypeUserOnline      = "user_online"
	MsgTypeUserOffline     = "user_offline"

	// System messages
	MsgTypeSystemAlert = "system_alert"
	MsgTypeHeartbeat   = "heartbeat"
	MsgTypeError       = "error"
)

// PositionUpdate represents a position update payload
type PositionUpdate struct {
	EntityID string   `json:"entityId"`
	Position Position `json:"position"`
}

// Position represents a tactical position
type Position struct {
	Lat       float64   `json:"lat"`
	Lng       float64   `json:"lng"`
	Altitude  *float64  `json:"altitude,omitempty"`
	Speed     *float64  `json:"speed,omitempty"`
	Course    *float64  `json:"course,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

// Chat-specific payload structures
type ChatMessagePayload struct {
	Message *chat.ChatMessage `json:"message"`
	Action  string            `json:"action"` // "new", "update", "delete"
}

type ChatRoomPayload struct {
	Room   *chat.ChatRoom `json:"room"`
	Action string         `json:"action"` // "created", "updated", "joined", "left"
}

type MessageReactionPayload struct {
	MessageID uuid.UUID             `json:"messageId"`
	Reaction  *chat.MessageReaction `json:"reaction"`
	Action    string                `json:"action"` // "added", "removed"
}

type UserTypingPayload struct {
	RoomID   uuid.UUID `json:"roomId"`
	UserID   uuid.UUID `json:"userId"`
	Username string    `json:"username"`
	Callsign *string   `json:"callsign,omitempty"`
	Typing   bool      `json:"typing"`
}

type UserStatusPayload struct {
	UserID   uuid.UUID `json:"userId"`
	Username string    `json:"username"`
	Callsign *string   `json:"callsign,omitempty"`
	Online   bool      `json:"online"`
	LastSeen time.Time `json:"lastSeen"`
}

// Incoming message structures for client requests
type IncomingMessage struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
	RoomID  *uuid.UUID      `json:"roomId,omitempty"`
}

type SendMessagePayload struct {
	RoomID         uuid.UUID             `json:"roomId"`
	MessageText    string                `json:"messageText"`
	MessageType    *chat.MessageType     `json:"messageType,omitempty"`
	Priority       *chat.MessagePriority `json:"priority,omitempty"`
	Classification *chat.Classification  `json:"classification,omitempty"`
	LocationLat    *float64              `json:"locationLat,omitempty"`
	LocationLng    *float64              `json:"locationLng,omitempty"`
	ReplyToID      *uuid.UUID            `json:"replyToId,omitempty"`
	RequiresAck    bool                  `json:"requiresAck"`
}

type JoinRoomPayload struct {
	RoomID uuid.UUID `json:"roomId"`
}

type ReactToMessagePayload struct {
	MessageID    uuid.UUID         `json:"messageId"`
	ReactionType chat.ReactionType `json:"reactionType"`
}

type AcknowledgeMessagePayload struct {
	MessageID uuid.UUID `json:"messageId"`
}

// TacticalWSClient represents a connected WebSocket client
type TacticalWSClient struct {
	conn     *websocket.Conn
	send     chan TacticalWSMessage
	hub      *TacticalWSHub
	userID   *uuid.UUID
	username string
	callsign *string
	rooms    map[uuid.UUID]bool // Rooms the client has joined
	typingIn *uuid.UUID         // Room ID where user is currently typing
	lastSeen time.Time
	mu       sync.RWMutex
}

// TacticalWSHub manages WebSocket connections and broadcasts
type TacticalWSHub struct {
	clients     map[*TacticalWSClient]bool
	broadcast   chan TacticalWSMessage
	register    chan *TacticalWSClient
	unregister  chan *TacticalWSClient
	mu          sync.RWMutex
	logger      *logger.Logger
	chatService *chat.Service
}

// NewTacticalWSHub creates a new WebSocket hub
func NewTacticalWSHub(logger *logger.Logger, chatService *chat.Service) *TacticalWSHub {
	return &TacticalWSHub{
		clients:     make(map[*TacticalWSClient]bool),
		broadcast:   make(chan TacticalWSMessage, 256),
		register:    make(chan *TacticalWSClient),
		unregister:  make(chan *TacticalWSClient),
		logger:      logger,
		chatService: chatService,
	}
}

// Run starts the WebSocket hub
func (h *TacticalWSHub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
			h.logger.Info().Int("client_count", len(h.clients)).Msg("WebSocket client connected")

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			h.mu.Unlock()
			h.logger.Info().Int("client_count", len(h.clients)).Msg("WebSocket client disconnected")

		case message := <-h.broadcast:
			h.mu.RLock()
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					delete(h.clients, client)
					close(client.send)
				}
			}
			h.mu.RUnlock()
		}
	}
}

// BroadcastPositionUpdate broadcasts a position update to all connected clients
func (h *TacticalWSHub) BroadcastPositionUpdate(entityID string, lat, lng float64, altitude, speed, course *float64) {
	message := TacticalWSMessage{
		Type: "position_update",
		Payload: PositionUpdate{
			EntityID: entityID,
			Position: Position{
				Lat:       lat,
				Lng:       lng,
				Altitude:  altitude,
				Speed:     speed,
				Course:    course,
				Timestamp: time.Now(),
			},
		},
		Timestamp: time.Now(),
	}

	select {
	case h.broadcast <- message:
	default:
		h.logger.Warn().Str("entity_id", entityID).Msg("Failed to broadcast position update - channel full")
	}
}

// BroadcastEntityRemoved broadcasts an entity removal to all connected clients
func (h *TacticalWSHub) BroadcastEntityRemoved(entityID string) {
	message := TacticalWSMessage{
		Type: "entity_removed",
		Payload: map[string]string{
			"entityId": entityID,
		},
		Timestamp: time.Now(),
	}

	select {
	case h.broadcast <- message:
	default:
		h.logger.Warn().Str("entity_id", entityID).Msg("Failed to broadcast entity removal - channel full")
	}
}

// BroadcastChatMessage broadcasts a chat message to room participants
func (h *TacticalWSHub) BroadcastChatMessage(roomID uuid.UUID, message *chat.ChatMessage, action string) {
	payload := ChatMessagePayload{
		Message: message,
		Action:  action,
	}

	wsMessage := TacticalWSMessage{
		Type:      MsgTypeChatMessage,
		Payload:   payload,
		Timestamp: time.Now(),
		RoomID:    &roomID,
	}

	// Send to all clients in the room
	h.mu.RLock()
	for client := range h.clients {
		client.mu.RLock()
		isInRoom := client.rooms[roomID]
		client.mu.RUnlock()

		if isInRoom {
			select {
			case client.send <- wsMessage:
			default:
				h.logger.Warn().Str("room_id", roomID.String()).Msg("Client send channel full, dropping chat message")
			}
		}
	}
	h.mu.RUnlock()
}

// BroadcastRoomUpdate broadcasts a room update to room participants
func (h *TacticalWSHub) BroadcastRoomUpdate(roomID uuid.UUID, room *chat.ChatRoom, action string) {
	payload := ChatRoomPayload{
		Room:   room,
		Action: action,
	}

	wsMessage := TacticalWSMessage{
		Type:      MsgTypeChatRoomUpdate,
		Payload:   payload,
		Timestamp: time.Now(),
		RoomID:    &roomID,
	}

	// Send to all clients in the room
	h.mu.RLock()
	for client := range h.clients {
		client.mu.RLock()
		isInRoom := client.rooms[roomID]
		client.mu.RUnlock()

		if isInRoom {
			select {
			case client.send <- wsMessage:
			default:
				h.logger.Warn().Str("room_id", roomID.String()).Msg("Client send channel full, dropping room update")
			}
		}
	}
	h.mu.RUnlock()
}

// BroadcastMessageReaction broadcasts a message reaction to room participants
func (h *TacticalWSHub) BroadcastMessageReaction(messageID uuid.UUID, reaction *chat.MessageReaction, action string) {
	payload := MessageReactionPayload{
		MessageID: messageID,
		Reaction:  reaction,
		Action:    action,
	}

	wsMessage := TacticalWSMessage{
		Type:      MsgTypeMessageReaction,
		Payload:   payload,
		Timestamp: time.Now(),
	}

	// Broadcast to all clients (they'll filter based on room membership)
	select {
	case h.broadcast <- wsMessage:
	default:
		h.logger.Warn().Str("message_id", messageID.String()).Msg("Failed to broadcast message reaction - channel full")
	}
}

// BroadcastTypingIndicator broadcasts typing indicators to room participants
func (h *TacticalWSHub) BroadcastTypingIndicator(roomID, userID uuid.UUID, username string, callsign *string, typing bool) {
	payload := UserTypingPayload{
		RoomID:   roomID,
		UserID:   userID,
		Username: username,
		Callsign: callsign,
		Typing:   typing,
	}

	wsMessage := TacticalWSMessage{
		Type:      MsgTypeUserTyping,
		Payload:   payload,
		Timestamp: time.Now(),
		RoomID:    &roomID,
	}

	// Send to all clients in the room except the sender
	h.mu.RLock()
	for client := range h.clients {
		client.mu.RLock()
		isInRoom := client.rooms[roomID]
		isSender := client.userID != nil && *client.userID == userID
		client.mu.RUnlock()

		if isInRoom && !isSender {
			select {
			case client.send <- wsMessage:
			default:
				h.logger.Warn().Str("room_id", roomID.String()).Msg("Client send channel full, dropping typing indicator")
			}
		}
	}
	h.mu.RUnlock()
}

// BroadcastUserStatus broadcasts user online/offline status
func (h *TacticalWSHub) BroadcastUserStatus(userID uuid.UUID, username string, callsign *string, online bool) {
	payload := UserStatusPayload{
		UserID:   userID,
		Username: username,
		Callsign: callsign,
		Online:   online,
		LastSeen: time.Now(),
	}

	var msgType string
	if online {
		msgType = MsgTypeUserOnline
	} else {
		msgType = MsgTypeUserOffline
	}

	wsMessage := TacticalWSMessage{
		Type:      msgType,
		Payload:   payload,
		Timestamp: time.Now(),
	}

	// Broadcast to all clients
	select {
	case h.broadcast <- wsMessage:
	default:
		h.logger.Warn().Str("user_id", userID.String()).Msg("Failed to broadcast user status - channel full")
	}
}

// HandleTacticalWebSocket handles WebSocket connections for tactical data
func HandleTacticalWebSocket(hub *TacticalWSHub, logger *logger.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			logger.Error().Err(err).Msg("WebSocket upgrade failed")
			return
		}

		// TODO: Extract user ID from authentication context
		userID := uuid.New() // For now, generate a random UUID for anonymous connections

		client := &TacticalWSClient{
			conn:     conn,
			send:     make(chan TacticalWSMessage, 256),
			hub:      hub,
			userID:   &userID,
			username: "Anonymous",
			rooms:    make(map[uuid.UUID]bool),
			lastSeen: time.Now(),
		}

		client.hub.register <- client

		// Start goroutines for handling the connection
		go client.writePump()
		go client.readPump()
	}
}

// readPump handles reading from the WebSocket connection
func (c *TacticalWSClient) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(2048) // Increased for chat messages
	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.hub.logger.Error().Err(err).Msg("WebSocket error")
			}
			break
		}

		// Parse incoming message
		var incomingMsg IncomingMessage
		if err := json.Unmarshal(message, &incomingMsg); err != nil {
			c.hub.logger.Error().Err(err).Msg("Failed to parse WebSocket message")
			c.sendError("Invalid message format")
			continue
		}

		// Update last seen timestamp
		c.mu.Lock()
		c.lastSeen = time.Now()
		c.mu.Unlock()

		// Handle different message types
		if err := c.handleMessage(incomingMsg); err != nil {
			c.hub.logger.Error().Err(err).Str("message_type", incomingMsg.Type).Msg("Failed to handle WebSocket message")
			c.sendError(fmt.Sprintf("Failed to handle message: %v", err))
		}
	}
}

// writePump handles writing to the WebSocket connection
func (c *TacticalWSClient) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}

			if err := json.NewEncoder(w).Encode(message); err != nil {
				return
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleMessage processes incoming WebSocket messages based on type
func (c *TacticalWSClient) handleMessage(msg IncomingMessage) error {
	switch msg.Type {
	case "send_message":
		return c.handleSendMessage(msg.Payload)
	case "join_room":
		return c.handleJoinRoom(msg.Payload)
	case "leave_room":
		return c.handleLeaveRoom(msg.Payload)
	case "react_to_message":
		return c.handleReactToMessage(msg.Payload)
	case "acknowledge_message":
		return c.handleAcknowledgeMessage(msg.Payload)
	case "typing_start":
		return c.handleTypingIndicator(msg.Payload, true)
	case "typing_stop":
		return c.handleTypingIndicator(msg.Payload, false)
	case "heartbeat":
		return c.handleHeartbeat()
	default:
		return fmt.Errorf("unknown message type: %s", msg.Type)
	}
}

// handleSendMessage processes a send message request
func (c *TacticalWSClient) handleSendMessage(payload json.RawMessage) error {
	var req SendMessagePayload
	if err := json.Unmarshal(payload, &req); err != nil {
		return fmt.Errorf("invalid send message payload: %w", err)
	}

	// Check if user is in the room
	c.mu.RLock()
	userID := c.userID
	isInRoom := c.rooms[req.RoomID]
	c.mu.RUnlock()

	if !isInRoom {
		return fmt.Errorf("not a member of room %s", req.RoomID)
	}

	// Create chat message request
	chatReq := chat.SendMessageRequest{
		RoomID:         req.RoomID,
		MessageText:    req.MessageText,
		MessageType:    chat.MessageTypeText,
		Priority:       chat.PriorityNormal,
		Classification: chat.ClassificationUnclassified,
		LocationLat:    req.LocationLat,
		LocationLng:    req.LocationLng,
		ReplyToID:      req.ReplyToID,
		RequiresAck:    req.RequiresAck,
	}

	// Set message type and priority if provided
	if req.MessageType != nil {
		chatReq.MessageType = *req.MessageType
	}
	if req.Priority != nil {
		chatReq.Priority = *req.Priority
	}
	if req.Classification != nil {
		chatReq.Classification = *req.Classification
	}

	// Send message via chat service
	ctx := context.Background()
	message, err := c.hub.chatService.SendMessage(ctx, chatReq, userID, c.getCallsign())
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	// Broadcast message to room participants
	c.hub.BroadcastChatMessage(req.RoomID, message, "new")

	return nil
}

// handleJoinRoom processes a join room request
func (c *TacticalWSClient) handleJoinRoom(payload json.RawMessage) error {
	var req JoinRoomPayload
	if err := json.Unmarshal(payload, &req); err != nil {
		return fmt.Errorf("invalid join room payload: %w", err)
	}

	c.mu.Lock()
	userID := c.userID
	c.rooms[req.RoomID] = true
	c.mu.Unlock()

	// Add participant to room via chat service
	ctx := context.Background()
	_, err := c.hub.chatService.AddParticipant(ctx, req.RoomID, *userID, c.getCallsign(), chat.RoleMember)
	if err != nil {
		c.hub.logger.Warn().Err(err).Str("room_id", req.RoomID.String()).Msg("Failed to add participant to room")
		// Don't return error - user might already be in room
	}

	// Send confirmation
	c.sendMessage(TacticalWSMessage{
		Type:      MsgTypeChatRoomJoined,
		Payload:   map[string]string{"roomId": req.RoomID.String()},
		Timestamp: time.Now(),
		RoomID:    &req.RoomID,
	})

	// Broadcast user joined to room
	c.hub.BroadcastRoomUpdate(req.RoomID, nil, "user_joined")

	return nil
}

// handleLeaveRoom processes a leave room request
func (c *TacticalWSClient) handleLeaveRoom(payload json.RawMessage) error {
	var req JoinRoomPayload // Same payload structure
	if err := json.Unmarshal(payload, &req); err != nil {
		return fmt.Errorf("invalid leave room payload: %w", err)
	}

	c.mu.Lock()
	delete(c.rooms, req.RoomID)
	c.mu.Unlock()

	// Remove participant from room via chat service
	ctx := context.Background()
	err := c.hub.chatService.RemoveParticipant(ctx, req.RoomID, *c.userID)
	if err != nil {
		c.hub.logger.Warn().Err(err).Str("room_id", req.RoomID.String()).Msg("Failed to remove participant from room")
	}

	// Send confirmation
	c.sendMessage(TacticalWSMessage{
		Type:      MsgTypeChatRoomLeft,
		Payload:   map[string]string{"roomId": req.RoomID.String()},
		Timestamp: time.Now(),
		RoomID:    &req.RoomID,
	})

	// Broadcast user left to room
	c.hub.BroadcastRoomUpdate(req.RoomID, nil, "user_left")

	return nil
}

// handleReactToMessage processes a message reaction request
func (c *TacticalWSClient) handleReactToMessage(payload json.RawMessage) error {
	var req ReactToMessagePayload
	if err := json.Unmarshal(payload, &req); err != nil {
		return fmt.Errorf("invalid reaction payload: %w", err)
	}

	// Add reaction via chat service
	ctx := context.Background()
	err := c.hub.chatService.AddReaction(ctx, req.MessageID, *c.userID, req.ReactionType)
	if err != nil {
		return fmt.Errorf("failed to add reaction: %w", err)
	}

	// Create reaction object for broadcast
	reaction := &chat.MessageReaction{
		ID:           uuid.New(),
		MessageID:    req.MessageID,
		UserID:       *c.userID,
		ReactionType: req.ReactionType,
		CreatedAt:    time.Now(),
		Username:     &c.username,
		Callsign:     c.callsign,
	}

	// Broadcast reaction to all clients (room will be determined by message)
	c.hub.BroadcastMessageReaction(req.MessageID, reaction, "added")

	return nil
}

// handleAcknowledgeMessage processes a message acknowledgment request
func (c *TacticalWSClient) handleAcknowledgeMessage(payload json.RawMessage) error {
	var req AcknowledgeMessagePayload
	if err := json.Unmarshal(payload, &req); err != nil {
		return fmt.Errorf("invalid acknowledgment payload: %w", err)
	}

	// Acknowledge message via chat service
	ctx := context.Background()
	err := c.hub.chatService.AcknowledgeMessage(ctx, req.MessageID, *c.userID)
	if err != nil {
		return fmt.Errorf("failed to acknowledge message: %w", err)
	}

	// Send confirmation
	c.sendMessage(TacticalWSMessage{
		Type: MsgTypeMessageAck,
		Payload: map[string]interface{}{
			"messageId": req.MessageID.String(),
			"userId":    c.userID.String(),
		},
		Timestamp: time.Now(),
	})

	return nil
}

// handleTypingIndicator processes typing indicator messages
func (c *TacticalWSClient) handleTypingIndicator(payload json.RawMessage, isTyping bool) error {
	var roomPayload struct {
		RoomID uuid.UUID `json:"roomId"`
	}
	if err := json.Unmarshal(payload, &roomPayload); err != nil {
		return fmt.Errorf("invalid typing payload: %w", err)
	}

	// Check if user is in the room
	c.mu.Lock()
	isInRoom := c.rooms[roomPayload.RoomID]
	if isTyping {
		c.typingIn = &roomPayload.RoomID
	} else if c.typingIn != nil && *c.typingIn == roomPayload.RoomID {
		c.typingIn = nil
	}
	c.mu.Unlock()

	if !isInRoom {
		return fmt.Errorf("not a member of room %s", roomPayload.RoomID)
	}

	// Broadcast typing indicator to room
	c.hub.BroadcastTypingIndicator(roomPayload.RoomID, *c.userID, c.username, c.callsign, isTyping)

	return nil
}

// handleHeartbeat processes heartbeat messages
func (c *TacticalWSClient) handleHeartbeat() error {
	// Update last seen and send heartbeat response
	c.mu.Lock()
	c.lastSeen = time.Now()
	c.mu.Unlock()

	c.sendMessage(TacticalWSMessage{
		Type:      MsgTypeHeartbeat,
		Payload:   map[string]interface{}{"timestamp": time.Now().Unix()},
		Timestamp: time.Now(),
	})

	return nil
}

// sendError sends an error message to the client
func (c *TacticalWSClient) sendError(message string) {
	c.sendMessage(TacticalWSMessage{
		Type:      MsgTypeError,
		Payload:   map[string]string{"error": message},
		Timestamp: time.Now(),
	})
}

// sendMessage sends a message to the client
func (c *TacticalWSClient) sendMessage(msg TacticalWSMessage) {
	select {
	case c.send <- msg:
	default:
		c.hub.logger.Warn().Msg("Client send channel full, dropping message")
	}
}

// getCallsign returns the client's callsign or username
func (c *TacticalWSClient) getCallsign() string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.callsign != nil {
		return *c.callsign
	}
	return c.username
}
