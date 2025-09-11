package position

import (
	"strconv"
	"sync"
	"time"

	"github.com/dfedick/gotak/pkg/cot"
)

// EntityPosition represents a tracked entity's position and metadata
type EntityPosition struct {
	EntityID    string     `json:"entityId"`
	UID         string     `json:"uid"`
	Type        string     `json:"type"`
	Callsign    string     `json:"callsign"`
	Group       string     `json:"group"`
	
	// Position data
	Lat         float64    `json:"lat"`
	Lng         float64    `json:"lng"`
	Altitude    *float64   `json:"altitude,omitempty"`
	Speed       *float64   `json:"speed,omitempty"`
	Course      *float64   `json:"course,omitempty"`
	
	// Metadata
	LastUpdate  time.Time  `json:"lastUpdate"`
	StaleTime   time.Time  `json:"staleTime"`
	IsStale     bool       `json:"isStale"`
	
	// Tactical information
	IsFriendly  bool       `json:"isFriendly"`
	IsHostile   bool       `json:"isHostile"`
	
	// Position accuracy
	CircularError *float64 `json:"circularError,omitempty"`
	LinearError   *float64 `json:"linearError,omitempty"`
}

// PositionHistory represents a position update for trail tracking
type PositionHistory struct {
	Lat       float64   `json:"lat"`
	Lng       float64   `json:"lng"`
	Altitude  *float64  `json:"altitude,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

// Service manages entity position tracking
type Service struct {
	entities map[string]*EntityPosition
	trails   map[string][]PositionHistory // Entity trails for movement history
	mu       sync.RWMutex
	
	// Configuration
	maxTrailPoints    int
	staleTimeout      time.Duration
	cleanupInterval   time.Duration
	
	// Cleanup ticker
	cleanupTicker *time.Ticker
	stopChan      chan struct{}
}

// NewService creates a new position service
func NewService() *Service {
	s := &Service{
		entities:        make(map[string]*EntityPosition),
		trails:          make(map[string][]PositionHistory),
		maxTrailPoints:  100,
		staleTimeout:    5 * time.Minute,
		cleanupInterval: 1 * time.Minute,
		stopChan:        make(chan struct{}),
	}
	
	// Start cleanup goroutine
	s.cleanupTicker = time.NewTicker(s.cleanupInterval)
	go s.cleanupLoop()
	
	return s
}

// UpdatePosition updates an entity's position from a CoT event
func (s *Service) UpdatePosition(event *cot.Event, callsign string) error {
	lat, lng, err := event.GetPosition()
	if err != nil {
		return err
	}
	
	s.mu.Lock()
	defer s.mu.Unlock()
	
	// Use callsign as entity ID, fallback to UID
	entityID := callsign
	if entityID == "" {
		entityID = event.UID
	}
	
	// Parse additional position data
	var altitude, speed, course, ce, le *float64
	
	if event.Point != nil {
		if event.Point.Hae != "" {
			if hae, err := parseFloat(event.Point.Hae); err == nil {
				altitude = &hae
			}
		}
		if event.Point.CE != "" {
			if ceVal, err := parseFloat(event.Point.CE); err == nil {
				ce = &ceVal
			}
		}
		if event.Point.LE != "" {
			if leVal, err := parseFloat(event.Point.LE); err == nil {
				le = &leVal
			}
		}
	}
	
	if event.Detail != nil && event.Detail.Track != nil {
		if event.Detail.Track.Speed != "" {
			if s, err := parseFloat(event.Detail.Track.Speed); err == nil {
				speed = &s
			}
		}
		if event.Detail.Track.Course != "" {
			if c, err := parseFloat(event.Detail.Track.Course); err == nil {
				course = &c
			}
		}
	}
	
	// Parse timestamp
	timestamp := time.Now()
	if event.Time != "" {
		if t, err := time.Parse(time.RFC3339Nano, event.Time); err == nil {
			timestamp = t
		}
	}
	
	// Parse stale time
	staleTime := timestamp.Add(s.staleTimeout)
	if event.Stale != "" {
		if t, err := time.Parse(time.RFC3339Nano, event.Stale); err == nil {
			staleTime = t
		}
	}
	
	// Determine if friendly/hostile based on CoT type
	isFriendly := cot.IsTypeAtom(event.Type)
	isHostile := cot.IsTypeBit(event.Type)
	
	// Get existing entity or create new one
	entity, exists := s.entities[entityID]
	if !exists {
		entity = &EntityPosition{
			EntityID: entityID,
		}
		s.entities[entityID] = entity
	}
	
	// Update position history trail
	if exists {
		// Add to trail if position has changed significantly (> 10 meters)
		if s.hasPositionChanged(entity.Lat, entity.Lng, lat, lng, 10.0) {
			s.addToTrail(entityID, entity.Lat, entity.Lng, entity.Altitude, entity.LastUpdate)
		}
	}
	
	// Update entity data
	entity.UID = event.UID
	entity.Type = event.Type
	entity.Callsign = callsign
	entity.Group = event.GetGroup()
	entity.Lat = lat
	entity.Lng = lng
	entity.Altitude = altitude
	entity.Speed = speed
	entity.Course = course
	entity.LastUpdate = timestamp
	entity.StaleTime = staleTime
	entity.IsStale = time.Now().After(staleTime)
	entity.IsFriendly = isFriendly
	entity.IsHostile = isHostile
	entity.CircularError = ce
	entity.LinearError = le
	
	return nil
}

// GetAllPositions returns all tracked entity positions
func (s *Service) GetAllPositions() []*EntityPosition {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	positions := make([]*EntityPosition, 0, len(s.entities))
	for _, entity := range s.entities {
		// Create a copy to avoid race conditions
		entityCopy := *entity
		positions = append(positions, &entityCopy)
	}
	
	return positions
}

// GetPosition returns a specific entity's position
func (s *Service) GetPosition(entityID string) (*EntityPosition, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	entity, exists := s.entities[entityID]
	if !exists {
		return nil, false
	}
	
	// Return a copy
	entityCopy := *entity
	return &entityCopy, true
}

// GetTrail returns the movement trail for an entity
func (s *Service) GetTrail(entityID string) []PositionHistory {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	trail, exists := s.trails[entityID]
	if !exists {
		return nil
	}
	
	// Return a copy
	trailCopy := make([]PositionHistory, len(trail))
	copy(trailCopy, trail)
	return trailCopy
}

// RemoveEntity removes an entity from tracking
func (s *Service) RemoveEntity(entityID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	delete(s.entities, entityID)
	delete(s.trails, entityID)
}

// CleanupStaleEntities removes entities that haven't updated recently
func (s *Service) CleanupStaleEntities() {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	now := time.Now()
	for entityID, entity := range s.entities {
		if now.After(entity.StaleTime) {
			entity.IsStale = true
			
			// Remove entities that have been stale for too long
			if now.Sub(entity.StaleTime) > s.staleTimeout {
				delete(s.entities, entityID)
				delete(s.trails, entityID)
			}
		}
	}
}

// GetActiveEntityCount returns the number of active (non-stale) entities
func (s *Service) GetActiveEntityCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	count := 0
	for _, entity := range s.entities {
		if !entity.IsStale {
			count++
		}
	}
	return count
}

// GetStatistics returns position service statistics
func (s *Service) GetStatistics() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	activeCount := 0
	staleCount := 0
	friendlyCount := 0
	hostileCount := 0
	
	for _, entity := range s.entities {
		if entity.IsStale {
			staleCount++
		} else {
			activeCount++
		}
		
		if entity.IsFriendly {
			friendlyCount++
		} else if entity.IsHostile {
			hostileCount++
		}
	}
	
	return map[string]interface{}{
		"total_entities":   len(s.entities),
		"active_entities":  activeCount,
		"stale_entities":   staleCount,
		"friendly_entities": friendlyCount,
		"hostile_entities": hostileCount,
		"trails_tracked":   len(s.trails),
	}
}

// Close stops the position service
func (s *Service) Close() {
	close(s.stopChan)
	if s.cleanupTicker != nil {
		s.cleanupTicker.Stop()
	}
}

// Private helper methods

func (s *Service) cleanupLoop() {
	for {
		select {
		case <-s.cleanupTicker.C:
			s.CleanupStaleEntities()
		case <-s.stopChan:
			return
		}
	}
}

func (s *Service) addToTrail(entityID string, lat, lng float64, altitude *float64, timestamp time.Time) {
	trail := s.trails[entityID]
	
	// Add new point
	trail = append(trail, PositionHistory{
		Lat:       lat,
		Lng:       lng,
		Altitude:  altitude,
		Timestamp: timestamp,
	})
	
	// Keep only the last N points
	if len(trail) > s.maxTrailPoints {
		trail = trail[len(trail)-s.maxTrailPoints:]
	}
	
	s.trails[entityID] = trail
}

func (s *Service) hasPositionChanged(oldLat, oldLng, newLat, newLng, thresholdMeters float64) bool {
	// Simple distance calculation (not exact but good enough for filtering)
	latDiff := newLat - oldLat
	lngDiff := newLng - oldLng
	
	// Rough conversion to meters (works for small distances)
	latMeters := latDiff * 111000 // 1 degree latitude ≈ 111km
	lngMeters := lngDiff * 111000 * cosApprox(oldLat)
	
	distanceMeters := sqrt(latMeters*latMeters + lngMeters*lngMeters)
	return distanceMeters > thresholdMeters
}

// Simple helper functions to avoid importing math package
func parseFloat(s string) (float64, error) {
	return strconv.ParseFloat(s, 64)
}

func cosApprox(degrees float64) float64 {
	// Simple cosine approximation for latitude adjustment
	// cos(lat) ≈ 1 - lat²/2 for small angles
	radians := degrees * 0.017453292519943295 // π/180
	return 1.0 - (radians*radians)/2.0
}

func sqrt(x float64) float64 {
	// Simple square root approximation using Newton's method
	if x < 0 {
		return 0
	}
	if x == 0 {
		return 0
	}
	
	guess := x / 2
	for i := 0; i < 10; i++ {
		guess = (guess + x/guess) / 2
	}
	return guess
}
