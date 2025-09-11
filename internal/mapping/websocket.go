package mapping

import (
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/dfedick/gotak/pkg/logger"
)

// MappingWSHub manages WebSocket connections for real-time mapping updates
type MappingWSHub struct {
	clients    map[*MappingWSClient]bool
	register   chan *MappingWSClient
	unregister chan *MappingWSClient
	broadcast  chan *MappingUpdate
	
	// Services for data access
	routeService    *RouteService
	geofenceService *GeofenceService
	
	// Geofence monitoring
	violationCallbacks []ViolationCallback
	
	mu     sync.RWMutex
	logger *logger.Logger
}

// MappingWSClient represents a WebSocket client for mapping updates
type MappingWSClient struct {
	send   chan *MappingUpdate
	userID uuid.UUID
	groupID string
	hub    *MappingWSHub
	
	// Subscriptions
	subscribedRoutes    map[uuid.UUID]bool
	subscribedGeofences map[uuid.UUID]bool
	subscribedAreas     []BoundingBox // Geographic areas of interest
	
	mu sync.RWMutex
}

// MappingUpdate represents a real-time mapping update
type MappingUpdate struct {
	Type      MappingUpdateType `json:"type"`
	Timestamp time.Time         `json:"timestamp"`
	UserID    uuid.UUID         `json:"user_id"`
	GroupID   string            `json:"group_id"`
	Data      interface{}       `json:"data"`
}

// MappingUpdateType represents the type of mapping update
type MappingUpdateType string

const (
	// Route updates
	UpdateTypeRouteCreated  MappingUpdateType = "route_created"
	UpdateTypeRouteUpdated  MappingUpdateType = "route_updated"
	UpdateTypeRouteDeleted  MappingUpdateType = "route_deleted"
	UpdateTypeRouteShared   MappingUpdateType = "route_shared"
	
	// Geofence updates
	UpdateTypeGeofenceCreated   MappingUpdateType = "geofence_created"
	UpdateTypeGeofenceUpdated   MappingUpdateType = "geofence_updated"
	UpdateTypeGeofenceDeleted   MappingUpdateType = "geofence_deleted"
	UpdateTypeGeofenceViolation MappingUpdateType = "geofence_violation"
	UpdateTypeGeofenceToggled   MappingUpdateType = "geofence_toggled"
	
	// Measurement updates
	UpdateTypeMeasurementCreated MappingUpdateType = "measurement_created"
	UpdateTypeMeasurementShared  MappingUpdateType = "measurement_shared"
	UpdateTypeMeasurementDeleted MappingUpdateType = "measurement_deleted"
	
	// Offline map updates
	UpdateTypeOfflineAreaCreated   MappingUpdateType = "offline_area_created"
	UpdateTypeOfflineAreaProgress  MappingUpdateType = "offline_area_progress"
	UpdateTypeOfflineAreaComplete  MappingUpdateType = "offline_area_complete"
	UpdateTypeOfflineAreaDeleted   MappingUpdateType = "offline_area_deleted"
	
	// Collaboration updates
	UpdateTypeUserPresence       MappingUpdateType = "user_presence"
	UpdateTypeCursorMovement     MappingUpdateType = "cursor_movement"
	UpdateTypeToolSelection      MappingUpdateType = "tool_selection"
)

// ViolationCallback defines a callback for geofence violations
type ViolationCallback func(violation *GeofenceViolation, geofence *Geofence)

// RouteUpdateData represents route update data
type RouteUpdateData struct {
	Route  *Route `json:"route"`
	Action string `json:"action"` // created, updated, deleted, shared
}

// GeofenceUpdateData represents geofence update data
type GeofenceUpdateData struct {
	Geofence *Geofence `json:"geofence"`
	Action   string    `json:"action"` // created, updated, deleted, toggled
}

// GeofenceViolationData represents geofence violation data
type GeofenceViolationData struct {
	Violation *GeofenceViolation `json:"violation"`
	Geofence  *Geofence          `json:"geofence"`
	Entity    EntityInfo         `json:"entity"`
}

// EntityInfo represents entity information for violations
type EntityInfo struct {
	ID       string    `json:"id"`
	Callsign string    `json:"callsign"`
	Type     string    `json:"type"`
	Position Point     `json:"position"`
	LastSeen time.Time `json:"last_seen"`
}

// OfflineAreaProgressData represents offline area download progress
type OfflineAreaProgressData struct {
	AreaID         uuid.UUID `json:"area_id"`
	Name           string    `json:"name"`
	Progress       float64   `json:"progress"`
	CompletedTiles int       `json:"completed_tiles"`
	TotalTiles     int       `json:"total_tiles"`
	EstimatedTime  *string   `json:"estimated_time,omitempty"`
	Status         string    `json:"status"`
}

// UserPresenceData represents user presence on the map
type UserPresenceData struct {
	UserID   uuid.UUID `json:"user_id"`
	Username string    `json:"username"`
	Callsign string    `json:"callsign"`
	Position *Point    `json:"position,omitempty"`
	Tool     string    `json:"tool,omitempty"` // route, geofence, measurement
	Online   bool      `json:"online"`
}

// NewMappingWSHub creates a new mapping WebSocket hub
func NewMappingWSHub(routeService *RouteService, geofenceService *GeofenceService, logger *logger.Logger) *MappingWSHub {
	hub := &MappingWSHub{
		clients:         make(map[*MappingWSClient]bool),
		register:        make(chan *MappingWSClient),
		unregister:      make(chan *MappingWSClient),
		broadcast:       make(chan *MappingUpdate, 256),
		routeService:    routeService,
		geofenceService: geofenceService,
		logger:          logger,
	}
	
	// Register geofence violation callback
	hub.violationCallbacks = append(hub.violationCallbacks, hub.handleGeofenceViolation)
	
	return hub
}

// Run starts the mapping WebSocket hub
func (h *MappingWSHub) Run() {
	h.logger.Info().Msg("Starting mapping WebSocket hub")
	
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
			
			h.logger.Info().
				Str("user_id", client.userID.String()).
				Str("group_id", client.groupID).
				Int("total_clients", len(h.clients)).
				Msg("Mapping WebSocket client connected")
			
			// Send initial presence update
			h.broadcastUserPresence(client, true)
			
		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			h.mu.Unlock()
			
			h.logger.Info().
				Str("user_id", client.userID.String()).
				Int("total_clients", len(h.clients)).
				Msg("Mapping WebSocket client disconnected")
			
			// Send offline presence update
			h.broadcastUserPresence(client, false)
			
		case update := <-h.broadcast:
			h.mu.RLock()
			for client := range h.clients {
				if h.shouldSendUpdateToClient(client, update) {
					select {
					case client.send <- update:
					default:
						// Client's send channel is full, disconnect
						delete(h.clients, client)
						close(client.send)
						h.logger.Warn().
							Str("user_id", client.userID.String()).
							Msg("Mapping WebSocket client send channel full, disconnecting")
					}
				}
			}
			h.mu.RUnlock()
		}
	}
}

// RegisterClient registers a new WebSocket client
func (h *MappingWSHub) RegisterClient(userID uuid.UUID, groupID string) *MappingWSClient {
	client := &MappingWSClient{
		send:                make(chan *MappingUpdate, 256),
		userID:              userID,
		groupID:             groupID,
		hub:                 h,
		subscribedRoutes:    make(map[uuid.UUID]bool),
		subscribedGeofences: make(map[uuid.UUID]bool),
		subscribedAreas:     make([]BoundingBox, 0),
	}
	
	h.register <- client
	return client
}

// UnregisterClient unregisters a WebSocket client
func (h *MappingWSHub) UnregisterClient(client *MappingWSClient) {
	h.unregister <- client
}

// BroadcastRouteUpdate broadcasts a route update to all relevant clients
func (h *MappingWSHub) BroadcastRouteUpdate(route *Route, updateType MappingUpdateType) {
	update := &MappingUpdate{
		Type:      updateType,
		Timestamp: time.Now(),
		UserID:    route.CreatedBy,
		GroupID:   route.GroupID,
		Data: RouteUpdateData{
			Route:  route,
			Action: string(updateType),
		},
	}
	
	select {
	case h.broadcast <- update:
		h.logger.Debug().
			Str("route_id", route.ID.String()).
			Str("update_type", string(updateType)).
			Msg("Broadcasting route update")
	default:
		h.logger.Warn().Msg("Failed to broadcast route update - channel full")
	}
}

// BroadcastGeofenceUpdate broadcasts a geofence update to all relevant clients
func (h *MappingWSHub) BroadcastGeofenceUpdate(geofence *Geofence, updateType MappingUpdateType) {
	update := &MappingUpdate{
		Type:      updateType,
		Timestamp: time.Now(),
		UserID:    geofence.CreatedBy,
		GroupID:   geofence.GroupID,
		Data: GeofenceUpdateData{
			Geofence: geofence,
			Action:   string(updateType),
		},
	}
	
	select {
	case h.broadcast <- update:
		h.logger.Debug().
			Str("geofence_id", geofence.ID.String()).
			Str("update_type", string(updateType)).
			Msg("Broadcasting geofence update")
	default:
		h.logger.Warn().Msg("Failed to broadcast geofence update - channel full")
	}
}

// BroadcastGeofenceViolation broadcasts a geofence violation to all relevant clients
func (h *MappingWSHub) BroadcastGeofenceViolation(violation *GeofenceViolation, geofence *Geofence, entityInfo EntityInfo) {
	update := &MappingUpdate{
		Type:      UpdateTypeGeofenceViolation,
		Timestamp: violation.Timestamp,
		UserID:    uuid.Nil, // System-generated
		GroupID:   geofence.GroupID,
		Data: GeofenceViolationData{
			Violation: violation,
			Geofence:  geofence,
			Entity:    entityInfo,
		},
	}
	
	select {
	case h.broadcast <- update:
		h.logger.Info().
			Str("geofence_id", geofence.ID.String()).
			Str("entity_id", violation.EntityID).
			Str("violation_type", string(violation.ViolationType)).
			Msg("Broadcasting geofence violation")
	default:
		h.logger.Warn().Msg("Failed to broadcast geofence violation - channel full")
	}
}

// BroadcastOfflineAreaProgress broadcasts offline area download progress
func (h *MappingWSHub) BroadcastOfflineAreaProgress(areaID uuid.UUID, name string, progress float64, completed, total int, status string) {
	update := &MappingUpdate{
		Type:      UpdateTypeOfflineAreaProgress,
		Timestamp: time.Now(),
		UserID:    uuid.Nil, // System-generated
		GroupID:   "", // Global update
		Data: OfflineAreaProgressData{
			AreaID:         areaID,
			Name:           name,
			Progress:       progress,
			CompletedTiles: completed,
			TotalTiles:     total,
			Status:         status,
		},
	}
	
	select {
	case h.broadcast <- update:
		h.logger.Debug().
			Str("area_id", areaID.String()).
			Float64("progress", progress).
			Msg("Broadcasting offline area progress")
	default:
		h.logger.Warn().Msg("Failed to broadcast offline area progress - channel full")
	}
}

// broadcastUserPresence broadcasts user presence updates
func (h *MappingWSHub) broadcastUserPresence(client *MappingWSClient, online bool) {
	update := &MappingUpdate{
		Type:      UpdateTypeUserPresence,
		Timestamp: time.Now(),
		UserID:    client.userID,
		GroupID:   client.groupID,
		Data: UserPresenceData{
			UserID:   client.userID,
			Username: "", // Would be filled from user service
			Callsign: "", // Would be filled from user service
			Online:   online,
		},
	}
	
	select {
	case h.broadcast <- update:
	default:
		h.logger.Warn().Msg("Failed to broadcast user presence - channel full")
	}
}

// shouldSendUpdateToClient determines if an update should be sent to a specific client
func (h *MappingWSHub) shouldSendUpdateToClient(client *MappingWSClient, update *MappingUpdate) bool {
	client.mu.RLock()
	defer client.mu.RUnlock()
	
	// Don't send updates back to the originator
	if update.UserID == client.userID && update.Type != UpdateTypeGeofenceViolation {
		return false
	}
	
	// Group-based filtering
	if update.GroupID != "" && update.GroupID != client.groupID {
		return false
	}
	
	// Type-specific filtering
	switch update.Type {
	case UpdateTypeRouteCreated, UpdateTypeRouteUpdated, UpdateTypeRouteDeleted, UpdateTypeRouteShared:
		// Send route updates to all clients in the same group
		return true
		
	case UpdateTypeGeofenceCreated, UpdateTypeGeofenceUpdated, UpdateTypeGeofenceDeleted, UpdateTypeGeofenceToggled:
		// Send geofence updates to all clients in the same group
		return true
		
	case UpdateTypeGeofenceViolation:
		// Send geofence violations to all clients in the same group
		return true
		
	case UpdateTypeOfflineAreaCreated, UpdateTypeOfflineAreaProgress, UpdateTypeOfflineAreaComplete, UpdateTypeOfflineAreaDeleted:
		// Send offline area updates to all clients (global)
		return true
		
	case UpdateTypeUserPresence, UpdateTypeCursorMovement, UpdateTypeToolSelection:
		// Send collaboration updates to all clients in the same group
		return true
		
	default:
		return false
	}
}

// handleGeofenceViolation is a callback function for geofence violations
func (h *MappingWSHub) handleGeofenceViolation(violation *GeofenceViolation, geofence *Geofence) {
	// Create entity info (in a real implementation, this would fetch from entity service)
	entityInfo := EntityInfo{
		ID:       violation.EntityID,
		Callsign: violation.EntityID, // Fallback
		Type:     "unknown",
		Position: violation.Position,
		LastSeen: violation.Timestamp,
	}
	
	h.BroadcastGeofenceViolation(violation, geofence, entityInfo)
}

// SubscribeToRoute allows a client to subscribe to updates for a specific route
func (c *MappingWSClient) SubscribeToRoute(routeID uuid.UUID) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.subscribedRoutes[routeID] = true
}

// UnsubscribeFromRoute allows a client to unsubscribe from updates for a specific route
func (c *MappingWSClient) UnsubscribeFromRoute(routeID uuid.UUID) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.subscribedRoutes, routeID)
}

// SubscribeToGeofence allows a client to subscribe to updates for a specific geofence
func (c *MappingWSClient) SubscribeToGeofence(geofenceID uuid.UUID) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.subscribedGeofences[geofenceID] = true
}

// UnsubscribeFromGeofence allows a client to unsubscribe from updates for a specific geofence
func (c *MappingWSClient) UnsubscribeFromGeofence(geofenceID uuid.UUID) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.subscribedGeofences, geofenceID)
}

// SubscribeToArea allows a client to subscribe to updates within a geographic area
func (c *MappingWSClient) SubscribeToArea(bounds BoundingBox) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.subscribedAreas = append(c.subscribedAreas, bounds)
}

// Send sends an update to the client
func (c *MappingWSClient) Send(update *MappingUpdate) bool {
	select {
	case c.send <- update:
		return true
	default:
		return false
	}
}

// Close closes the client's send channel
func (c *MappingWSClient) Close() {
	c.hub.UnregisterClient(c)
}
