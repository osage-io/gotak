package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/dfedick/gotak/internal/position"
	"github.com/dfedick/gotak/pkg/logger"
)

// PositionService interface for position operations
type PositionService interface {
	GetAllPositions() []*position.EntityPosition
	GetPosition(entityID string) (*position.EntityPosition, bool)
	GetTrail(entityID string) []position.PositionHistory
	RemoveEntity(entityID string)
	GetStatistics() map[string]interface{}
}

// PositionHandlers handles position-related HTTP requests
type PositionHandlers struct {
	positionService PositionService
	logger          *logger.Logger
}

// NewPositionHandlers creates new position handlers
func NewPositionHandlers(positionService PositionService, logger *logger.Logger) *PositionHandlers {
	return &PositionHandlers{
		positionService: positionService,
		logger:          logger,
	}
}

// GetAllPositions handles GET /api/v1/positions
func (h *PositionHandlers) GetAllPositions(w http.ResponseWriter, r *http.Request) {
	positions := h.positionService.GetAllPositions()
	
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(positions); err != nil {
		h.logger.Error().Err(err).Msg("Failed to encode positions response")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// GetPosition handles GET /api/v1/positions/{entityId}
func (h *PositionHandlers) GetPosition(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	entityID := vars["entityId"]
	
	if entityID == "" {
		http.Error(w, "Entity ID is required", http.StatusBadRequest)
		return
	}
	
	position, exists := h.positionService.GetPosition(entityID)
	if !exists {
		http.Error(w, "Entity not found", http.StatusNotFound)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(position); err != nil {
		h.logger.Error().Err(err).Msg("Failed to encode position response")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// GetPositionTrail handles GET /api/v1/positions/{entityId}/trail
func (h *PositionHandlers) GetPositionTrail(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	entityID := vars["entityId"]
	
	if entityID == "" {
		http.Error(w, "Entity ID is required", http.StatusBadRequest)
		return
	}
	
	// Check if entity exists
	_, exists := h.positionService.GetPosition(entityID)
	if !exists {
		http.Error(w, "Entity not found", http.StatusNotFound)
		return
	}
	
	trail := h.positionService.GetTrail(entityID)
	if trail == nil {
		trail = []position.PositionHistory{} // Return empty array instead of null
	}
	
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(trail); err != nil {
		h.logger.Error().Err(err).Msg("Failed to encode trail response")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// GetPositionStatistics handles GET /api/v1/positions/statistics
func (h *PositionHandlers) GetPositionStatistics(w http.ResponseWriter, r *http.Request) {
	stats := h.positionService.GetStatistics()
	
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(stats); err != nil {
		h.logger.Error().Err(err).Msg("Failed to encode statistics response")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// DeletePosition handles DELETE /api/v1/positions/{entityId}
func (h *PositionHandlers) DeletePosition(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	entityID := vars["entityId"]
	
	if entityID == "" {
		http.Error(w, "Entity ID is required", http.StatusBadRequest)
		return
	}
	
	// Check if entity exists before deletion
	_, exists := h.positionService.GetPosition(entityID)
	if !exists {
		http.Error(w, "Entity not found", http.StatusNotFound)
		return
	}
	
	h.positionService.RemoveEntity(entityID)
	
	w.WriteHeader(http.StatusNoContent)
}

// GetPositionsInBounds handles GET /api/v1/positions/bounds
// Query parameters: north, south, east, west (all required)
func (h *PositionHandlers) GetPositionsInBounds(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	
	// Parse bounding box parameters
	northStr := query.Get("north")
	southStr := query.Get("south")
	eastStr := query.Get("east")
	westStr := query.Get("west")
	
	if northStr == "" || southStr == "" || eastStr == "" || westStr == "" {
		http.Error(w, "All bounding box parameters (north, south, east, west) are required", http.StatusBadRequest)
		return
	}
	
	north, err := strconv.ParseFloat(northStr, 64)
	if err != nil {
		http.Error(w, "Invalid north parameter", http.StatusBadRequest)
		return
	}
	
	south, err := strconv.ParseFloat(southStr, 64)
	if err != nil {
		http.Error(w, "Invalid south parameter", http.StatusBadRequest)
		return
	}
	
	east, err := strconv.ParseFloat(eastStr, 64)
	if err != nil {
		http.Error(w, "Invalid east parameter", http.StatusBadRequest)
		return
	}
	
	west, err := strconv.ParseFloat(westStr, 64)
	if err != nil {
		http.Error(w, "Invalid west parameter", http.StatusBadRequest)
		return
	}
	
	// Validate bounding box
	if north <= south || east <= west {
		http.Error(w, "Invalid bounding box: north must be > south, east must be > west", http.StatusBadRequest)
		return
	}
	
	// Get all positions and filter by bounds
	allPositions := h.positionService.GetAllPositions()
	var positionsInBounds []*position.EntityPosition
	
	for _, pos := range allPositions {
		if pos.Lat >= south && pos.Lat <= north && pos.Lng >= west && pos.Lng <= east {
			positionsInBounds = append(positionsInBounds, pos)
		}
	}
	
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(positionsInBounds); err != nil {
		h.logger.Error().Err(err).Msg("Failed to encode bounded positions response")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// GetActivePositions handles GET /api/v1/positions/active
// Returns only non-stale positions
func (h *PositionHandlers) GetActivePositions(w http.ResponseWriter, r *http.Request) {
	allPositions := h.positionService.GetAllPositions()
	var activePositions []*position.EntityPosition
	
	now := time.Now()
	for _, pos := range allPositions {
		// Update stale status based on current time
		isStale := now.After(pos.StaleTime)
		if !isStale {
			activePositions = append(activePositions, pos)
		}
	}
	
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(activePositions); err != nil {
		h.logger.Error().Err(err).Msg("Failed to encode active positions response")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// GetFriendlyPositions handles GET /api/v1/positions/friendly
func (h *PositionHandlers) GetFriendlyPositions(w http.ResponseWriter, r *http.Request) {
	allPositions := h.positionService.GetAllPositions()
	var friendlyPositions []*position.EntityPosition
	
	for _, pos := range allPositions {
		if pos.IsFriendly {
			friendlyPositions = append(friendlyPositions, pos)
		}
	}
	
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(friendlyPositions); err != nil {
		h.logger.Error().Err(err).Msg("Failed to encode friendly positions response")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// GetHostilePositions handles GET /api/v1/positions/hostile
func (h *PositionHandlers) GetHostilePositions(w http.ResponseWriter, r *http.Request) {
	allPositions := h.positionService.GetAllPositions()
	var hostilePositions []*position.EntityPosition
	
	for _, pos := range allPositions {
		if pos.IsHostile {
			hostilePositions = append(hostilePositions, pos)
		}
	}
	
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(hostilePositions); err != nil {
		h.logger.Error().Err(err).Msg("Failed to encode hostile positions response")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}
