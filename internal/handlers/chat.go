package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/dfedick/gotak/internal/chat"
	"github.com/dfedick/gotak/pkg/logger"
)

// ChatHandlers contains all chat-related HTTP handlers
type ChatHandlers struct {
	chatService *chat.Service
	logger      *logger.Logger
}

// NewChatHandlers creates a new ChatHandlers instance
func NewChatHandlers(chatService *chat.Service, logger *logger.Logger) *ChatHandlers {
	return &ChatHandlers{
		chatService: chatService,
		logger:      logger,
	}
}

// CreateRoom creates a new chat room
func (h *ChatHandlers) CreateRoom(w http.ResponseWriter, r *http.Request) {
	var req chat.CreateChatRoomRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Warn().Err(err).Msg("Failed to decode create room request")
		WriteJSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// TODO: Get user ID from authentication context
	userID := uuid.New() // For now, use a random UUID

	room, err := h.chatService.CreateRoom(r.Context(), req, userID)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to create chat room")
		WriteJSONError(w, "Failed to create room", http.StatusInternalServerError)
		return
	}

	WriteJSONResponse(w, room, http.StatusCreated)
}

// GetRoom retrieves a chat room by ID
func (h *ChatHandlers) GetRoom(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roomIDStr := vars["roomId"]
	
	roomID, err := uuid.Parse(roomIDStr)
	if err != nil {
		WriteJSONError(w, "Invalid room ID", http.StatusBadRequest)
		return
	}

	// TODO: Get user ID from authentication context
	userID := uuid.New()

	room, err := h.chatService.GetRoom(r.Context(), roomID, &userID)
	if err != nil {
		h.logger.Error().Err(err).Str("room_id", roomID.String()).Msg("Failed to get chat room")
		WriteJSONError(w, "Room not found", http.StatusNotFound)
		return
	}

	WriteJSONResponse(w, room, http.StatusOK)
}

// GetRooms lists chat rooms for the authenticated user
func (h *ChatHandlers) GetRooms(w http.ResponseWriter, r *http.Request) {
	// TODO: Get user ID from authentication context
	userID := uuid.New()

	// Parse optional room type filter
	var roomType *chat.ChatRoomType
	if typeStr := r.URL.Query().Get("type"); typeStr != "" {
		rt := chat.ChatRoomType(typeStr)
		roomType = &rt
	}

	rooms, err := h.chatService.GetRooms(r.Context(), userID, roomType)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to get chat rooms")
		WriteJSONError(w, "Failed to retrieve rooms", http.StatusInternalServerError)
		return
	}

	WriteJSONResponse(w, map[string]interface{}{
		"rooms": rooms,
		"count": len(rooms),
	}, http.StatusOK)
}

// SendMessage sends a message to a chat room
func (h *ChatHandlers) SendMessage(w http.ResponseWriter, r *http.Request) {
	var req chat.SendMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Warn().Err(err).Msg("Failed to decode send message request")
		WriteJSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// TODO: Get user ID and callsign from authentication context
	userID := uuid.New()
	callsign := "TestUser"

	message, err := h.chatService.SendMessage(r.Context(), req, &userID, callsign)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to send message")
		WriteJSONError(w, "Failed to send message", http.StatusInternalServerError)
		return
	}

	WriteJSONResponse(w, message, http.StatusCreated)
}

// GetMessages retrieves messages from a chat room
func (h *ChatHandlers) GetMessages(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roomIDStr := vars["roomId"]
	
	roomID, err := uuid.Parse(roomIDStr)
	if err != nil {
		WriteJSONError(w, "Invalid room ID", http.StatusBadRequest)
		return
	}

	// Parse query parameters
	req := chat.GetMessagesRequest{
		RoomID: roomID,
		Limit:  50, // Default limit
		Offset: 0,  // Default offset
	}

	// Parse optional parameters
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 && limit <= 100 {
			req.Limit = limit
		}
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil && offset >= 0 {
			req.Offset = offset
		}
	}

	if beforeIDStr := r.URL.Query().Get("beforeId"); beforeIDStr != "" {
		if beforeID, err := uuid.Parse(beforeIDStr); err == nil {
			req.BeforeID = &beforeID
		}
	}

	if afterIDStr := r.URL.Query().Get("afterId"); afterIDStr != "" {
		if afterID, err := uuid.Parse(afterIDStr); err == nil {
			req.AfterID = &afterID
		}
	}

	if messageTypeStr := r.URL.Query().Get("messageType"); messageTypeStr != "" {
		mt := chat.MessageType(messageTypeStr)
		req.MessageType = &mt
	}

	if priorityStr := r.URL.Query().Get("priority"); priorityStr != "" {
		p := chat.MessagePriority(priorityStr)
		req.Priority = &p
	}

	if classificationStr := r.URL.Query().Get("classification"); classificationStr != "" {
		c := chat.Classification(classificationStr)
		req.Classification = &c
	}

	if startTimeStr := r.URL.Query().Get("startTime"); startTimeStr != "" {
		if startTime, err := time.Parse(time.RFC3339, startTimeStr); err == nil {
			req.StartTime = &startTime
		}
	}

	if endTimeStr := r.URL.Query().Get("endTime"); endTimeStr != "" {
		if endTime, err := time.Parse(time.RFC3339, endTimeStr); err == nil {
			req.EndTime = &endTime
		}
	}

	// TODO: Get user ID from authentication context
	userID := uuid.New()

	messages, err := h.chatService.GetMessages(r.Context(), req, &userID)
	if err != nil {
		h.logger.Error().Err(err).Str("room_id", roomID.String()).Msg("Failed to get messages")
		WriteJSONError(w, "Failed to retrieve messages", http.StatusInternalServerError)
		return
	}

	WriteJSONResponse(w, map[string]interface{}{
		"messages": messages,
		"count":    len(messages),
		"roomId":   roomID,
		"limit":    req.Limit,
		"offset":   req.Offset,
	}, http.StatusOK)
}

// AcknowledgeMessage acknowledges a message
func (h *ChatHandlers) AcknowledgeMessage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	messageIDStr := vars["messageId"]
	
	messageID, err := uuid.Parse(messageIDStr)
	if err != nil {
		WriteJSONError(w, "Invalid message ID", http.StatusBadRequest)
		return
	}

	// TODO: Get user ID from authentication context
	userID := uuid.New()

	if err := h.chatService.AcknowledgeMessage(r.Context(), messageID, userID); err != nil {
		h.logger.Error().Err(err).Str("message_id", messageID.String()).Msg("Failed to acknowledge message")
		WriteJSONError(w, "Failed to acknowledge message", http.StatusInternalServerError)
		return
	}

	WriteJSONResponse(w, map[string]string{
		"status":    "success",
		"messageId": messageID.String(),
		"userId":    userID.String(),
	}, http.StatusOK)
}

// AddReaction adds a reaction to a message
func (h *ChatHandlers) AddReaction(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	messageIDStr := vars["messageId"]
	
	messageID, err := uuid.Parse(messageIDStr)
	if err != nil {
		WriteJSONError(w, "Invalid message ID", http.StatusBadRequest)
		return
	}

	var reqBody struct {
		ReactionType chat.ReactionType `json:"reactionType"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		h.logger.Warn().Err(err).Msg("Failed to decode add reaction request")
		WriteJSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// TODO: Get user ID from authentication context
	userID := uuid.New()

	if err := h.chatService.AddReaction(r.Context(), messageID, userID, reqBody.ReactionType); err != nil {
		h.logger.Error().Err(err).Str("message_id", messageID.String()).Msg("Failed to add reaction")
		WriteJSONError(w, "Failed to add reaction", http.StatusInternalServerError)
		return
	}

	WriteJSONResponse(w, map[string]string{
		"status":       "success",
		"messageId":    messageID.String(),
		"userId":       userID.String(),
		"reactionType": string(reqBody.ReactionType),
	}, http.StatusCreated)
}

// GetRoomParticipants retrieves participants of a chat room
func (h *ChatHandlers) GetRoomParticipants(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roomIDStr := vars["roomId"]
	
	roomID, err := uuid.Parse(roomIDStr)
	if err != nil {
		WriteJSONError(w, "Invalid room ID", http.StatusBadRequest)
		return
	}

	participants, err := h.chatService.GetParticipants(r.Context(), roomID)
	if err != nil {
		h.logger.Error().Err(err).Str("room_id", roomID.String()).Msg("Failed to get room participants")
		WriteJSONError(w, "Failed to retrieve participants", http.StatusInternalServerError)
		return
	}

	WriteJSONResponse(w, map[string]interface{}{
		"participants": participants,
		"count":        len(participants),
		"roomId":       roomID,
	}, http.StatusOK)
}

// AddParticipant adds a user to a chat room
func (h *ChatHandlers) AddParticipant(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roomIDStr := vars["roomId"]
	
	roomID, err := uuid.Parse(roomIDStr)
	if err != nil {
		WriteJSONError(w, "Invalid room ID", http.StatusBadRequest)
		return
	}

	var reqBody struct {
		UserID   uuid.UUID           `json:"userId"`
		Callsign string              `json:"callsign,omitempty"`
		Role     chat.ParticipantRole `json:"role"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		h.logger.Warn().Err(err).Msg("Failed to decode add participant request")
		WriteJSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	participant, err := h.chatService.AddParticipant(r.Context(), roomID, reqBody.UserID, reqBody.Callsign, reqBody.Role)
	if err != nil {
		h.logger.Error().Err(err).Str("room_id", roomID.String()).Msg("Failed to add participant")
		WriteJSONError(w, "Failed to add participant", http.StatusInternalServerError)
		return
	}

	WriteJSONResponse(w, participant, http.StatusCreated)
}

// RemoveParticipant removes a user from a chat room
func (h *ChatHandlers) RemoveParticipant(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roomIDStr := vars["roomId"]
	userIDStr := vars["userId"]
	
	roomID, err := uuid.Parse(roomIDStr)
	if err != nil {
		WriteJSONError(w, "Invalid room ID", http.StatusBadRequest)
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		WriteJSONError(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	if err := h.chatService.RemoveParticipant(r.Context(), roomID, userID); err != nil {
		h.logger.Error().Err(err).Str("room_id", roomID.String()).Str("user_id", userID.String()).Msg("Failed to remove participant")
		WriteJSONError(w, "Failed to remove participant", http.StatusInternalServerError)
		return
	}

	WriteJSONResponse(w, map[string]string{
		"status": "success",
		"roomId": roomID.String(),
		"userId": userID.String(),
	}, http.StatusOK)
}

// GetStatistics returns chat system statistics
func (h *ChatHandlers) GetStatistics(w http.ResponseWriter, r *http.Request) {
	stats, err := h.chatService.GetStatistics(r.Context())
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to get chat statistics")
		WriteJSONError(w, "Failed to retrieve statistics", http.StatusInternalServerError)
		return
	}

	WriteJSONResponse(w, stats, http.StatusOK)
}

