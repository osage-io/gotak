package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/dfedick/gotak/pkg/logger"
)
// Entity represents a tactical entity for the web API
type Entity struct {
	ID             string     `json:"id"`
	Callsign       string     `json:"callsign"`
	Type           string     `json:"type"`
	Affiliation    string     `json:"affiliation"`
	Lat            float64    `json:"lat"`
	Lng            float64    `json:"lng"`
	Altitude       *float64   `json:"altitude,omitempty"`
	Speed          *float64   `json:"speed,omitempty"`
	Course         *float64   `json:"course,omitempty"`
	LastUpdate     *time.Time `json:"lastUpdate,omitempty"`
	Classification string     `json:"classification"`
	Status         string     `json:"status"`
}

// PositionPoint represents a historical position point
type PositionPoint struct {
	Lat       float64   `json:"lat"`
	Lng       float64   `json:"lng"`
	Altitude  *float64  `json:"altitude,omitempty"`
	Speed     *float64  `json:"speed,omitempty"`
	Course    *float64  `json:"course,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

// EntityHistory represents historical positions for an entity
type EntityHistory struct {
	EntityID   string          `json:"entityId"`
	Callsign   string          `json:"callsign"`
	Positions  []PositionPoint `json:"positions"`
	TotalCount int             `json:"totalCount"`
}

// EntitiesResponse represents the response for the entities endpoint
type EntitiesResponse struct {
	Entities []Entity `json:"entities"`
	Count    int      `json:"count"`
}


// EntityService interface for accessing entity data
type EntityService interface {
	GetAllEntities() ([]Entity, error)
	GetEntityByID(id string) (*Entity, error)
	GetEntityHistory(id string, timeRange string, limit int) (*EntityHistory, error)
}

// MockEntityService provides mock data for development
type MockEntityService struct {
	entities map[string]Entity
	logger   *logger.Logger
}

// NewMockEntityService creates a new mock entity service
func NewMockEntityService(logger *logger.Logger) *MockEntityService {
	now := time.Now()
	altitude15 := 15.0
	altitude150 := 150.0
	speed25 := 2.5
	speed152 := 15.2
	speed250 := 25.0
	course45 := 45.0
	course180 := 180.0
	course270 := 270.0

	entities := map[string]Entity{
		"ALPHA-6": {
			ID:             "ALPHA-6",
			Callsign:       "ALPHA-6",
			Type:           "ground-infantry",
			Affiliation:    "friendly",
			Lat:            38.9072,
			Lng:            -77.0369,
			Altitude:       &altitude15,
			Speed:          &speed25,
			Course:         &course45,
			LastUpdate:     &now,
			Classification: "UNCLASSIFIED",
			Status:         "active",
		},
		"BRAVO-3": {
			ID:             "BRAVO-3",
			Callsign:       "BRAVO-3",
			Type:           "ground-vehicle-wheeled",
			Affiliation:    "friendly",
			Lat:            38.9122,
			Lng:            -77.0319,
			Altitude:       &altitude15,
			Speed:          &speed152,
			Course:         &course180,
			LastUpdate:     &now,
			Classification: "UNCLASSIFIED",
			Status:         "active",
		},
		"CHARLIE-1": {
			ID:             "CHARLIE-1",
			Callsign:       "CHARLIE-1",
			Type:           "air-rotorcraft",
			Affiliation:    "friendly",
			Lat:            38.9022,
			Lng:            -77.0419,
			Altitude:       &altitude150,
			Speed:          &speed250,
			Course:         &course270,
			LastUpdate:     &now,
			Classification: "UNCLASSIFIED",
			Status:         "active",
		},
	}

	return &MockEntityService{
		entities: entities,
		logger:   logger,
	}
}

// GetAllEntities returns all entities
func (m *MockEntityService) GetAllEntities() ([]Entity, error) {
	entities := make([]Entity, 0, len(m.entities))
	for _, entity := range m.entities {
		entities = append(entities, entity)
	}
	return entities, nil
}

// GetEntityByID returns a specific entity by ID
func (m *MockEntityService) GetEntityByID(id string) (*Entity, error) {
	entity, exists := m.entities[id]
	if !exists {
		return nil, fmt.Errorf("entity not found: %s", id)
	}
	return &entity, nil
}

// GetEntityHistory returns historical positions for an entity
func (m *MockEntityService) GetEntityHistory(id string, timeRange string, limit int) (*EntityHistory, error) {
	entity, exists := m.entities[id]
	if !exists {
		return nil, fmt.Errorf("entity not found: %s", id)
	}

	// Generate mock historical positions
	positions := make([]PositionPoint, 0, limit)
	now := time.Now()
	
	for i := 0; i < limit && i < 50; i++ {
		// Generate positions going back in time
		timestamp := now.Add(-time.Duration(i*5) * time.Minute)
		
		// Add some variation to the position
		latVar := (float64(i) * 0.0001) - 0.005
		lngVar := (float64(i) * 0.0001) - 0.005
		
		position := PositionPoint{
			Lat:       entity.Lat + latVar,
			Lng:       entity.Lng + lngVar,
			Altitude:  entity.Altitude,
			Speed:     entity.Speed,
			Course:    entity.Course,
			Timestamp: timestamp,
		}
		positions = append(positions, position)
	}

	return &EntityHistory{
		EntityID:   entity.ID,
		Callsign:   entity.Callsign,
		Positions:  positions,
		TotalCount: len(positions),
	}, nil
}

// HandleGetEntities handles GET /api/v1/entities
func HandleGetEntities(service EntityService, logger *logger.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Info().Str("remote_addr", r.RemoteAddr).Str("user_agent", r.UserAgent()).Msg("GET /api/v1/entities")

		entities, err := service.GetAllEntities()
		if err != nil {
			logger.Error().Err(err).Msg("Failed to get entities")
			writeErrorResponse(w, "Failed to retrieve entities", http.StatusInternalServerError)
			return
		}

		response := EntitiesResponse{
			Entities: entities,
			Count:    len(entities),
		}

		writeSuccessResponse(w, response)
	}
}

// HandleGetEntity handles GET /api/v1/entities/{id}
func HandleGetEntity(service EntityService, logger *logger.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		entityID := vars["id"]

		logger.Info().Str("entity_id", entityID).Str("remote_addr", r.RemoteAddr).Msg("GET /api/v1/entities/{id}")

		entity, err := service.GetEntityByID(entityID)
		if err != nil {
			logger.Warn().Str("entity_id", entityID).Err(err).Msg("Entity not found")
			writeErrorResponse(w, "Entity not found", http.StatusNotFound)
			return
		}

		writeSuccessResponse(w, entity)
	}
}

// HandleGetEntityHistory handles GET /api/v1/entities/{id}/history
func HandleGetEntityHistory(service EntityService, logger *logger.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		entityID := vars["id"]

		// Parse query parameters
		timeRange := r.URL.Query().Get("range")
		if timeRange == "" {
			timeRange = "24h"
		}

		limitStr := r.URL.Query().Get("limit")
		limit := 100 // default limit
		if limitStr != "" {
			if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
				limit = parsedLimit
			}
		}

		logger.Info().Str("entity_id", entityID).Str("time_range", timeRange).Int("limit", limit).Str("remote_addr", r.RemoteAddr).Msg("GET /api/v1/entities/{id}/history")

		history, err := service.GetEntityHistory(entityID, timeRange, limit)
		if err != nil {
			logger.Warn().Str("entity_id", entityID).Err(err).Msg("Failed to get entity history")
			writeErrorResponse(w, "Entity not found or history unavailable", http.StatusNotFound)
			return
		}

		writeSuccessResponse(w, history)
	}
}

// Helper functions for consistent API responses

func writeSuccessResponse(w http.ResponseWriter, data interface{}) {
	WriteJSONSuccess(w, data)
}

func writeErrorResponse(w http.ResponseWriter, message string, statusCode int) {
	WriteJSONError(w, message, statusCode)
}
