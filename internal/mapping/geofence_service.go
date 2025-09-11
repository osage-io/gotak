package mapping

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/dfedick/gotak/internal/position"
	"github.com/dfedick/gotak/pkg/database"
	"github.com/dfedick/gotak/pkg/logger"
)

// GeofenceService provides geofence management and violation monitoring
type GeofenceService struct {
	db             database.DB
	logger         *logger.Logger
	monitor        *ViolationMonitor
	positionSvc    *position.Service
	mu             sync.RWMutex
	geofences      map[uuid.UUID]*Geofence
	entityStates   map[string]*EntityGeofenceState
	alertCallbacks []AlertCallback
}

// AlertCallback defines a callback function for geofence alerts
type AlertCallback func(violation *GeofenceViolation, geofence *Geofence)

// EntityGeofenceState tracks which geofences an entity is currently inside
type EntityGeofenceState struct {
	EntityID        string
	InsideGeofences map[uuid.UUID]bool
	LastPosition    Point
	LastUpdate      time.Time
}

// NewGeofenceService creates a new geofence service
func NewGeofenceService(db database.DB, logger *logger.Logger, positionSvc *position.Service) *GeofenceService {
	gs := &GeofenceService{
		db:           db,
		logger:       logger,
		positionSvc:  positionSvc,
		geofences:    make(map[uuid.UUID]*Geofence),
		entityStates: make(map[string]*EntityGeofenceState),
	}

	// Initialize violation monitor
	gs.monitor = NewViolationMonitor(gs, logger)

	// Load geofences from database
	go gs.loadGeofencesFromDatabase()

	return gs
}

// Start begins monitoring for geofence violations
func (gs *GeofenceService) Start() error {
	gs.logger.Info().Msg("Starting geofence monitoring service")
	return gs.monitor.Start()
}

// Stop stops geofence monitoring
func (gs *GeofenceService) Stop() error {
	gs.logger.Info().Msg("Stopping geofence monitoring service")
	return gs.monitor.Stop()
}

// CreateGeofence creates a new geofence
func (gs *GeofenceService) CreateGeofence(ctx context.Context, req *CreateGeofenceRequest, createdBy uuid.UUID, groupID string) (*Geofence, error) {
	// Validate geometry based on type
	if err := gs.validateGeofenceGeometry(req.Type, req.Geometry); err != nil {
		return nil, fmt.Errorf("invalid geometry: %w", err)
	}

	geofence := &Geofence{
		ID:           uuid.New(),
		Name:         req.Name,
		Description:  req.Description,
		Type:         req.Type,
		Geometry:     req.Geometry,
		Enabled:      req.Enabled,
		AlertOnEnter: req.AlertOnEnter,
		AlertOnExit:  req.AlertOnExit,
		CreatedBy:    createdBy,
		GroupID:      groupID,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Save to database
	if err := gs.saveGeofenceToDatabase(ctx, geofence); err != nil {
		return nil, fmt.Errorf("failed to save geofence: %w", err)
	}

	// Add to in-memory cache
	gs.mu.Lock()
	gs.geofences[geofence.ID] = geofence
	gs.mu.Unlock()

	gs.logger.Info().
		Str("geofence_id", geofence.ID.String()).
		Str("name", geofence.Name).
		Str("type", string(geofence.Type)).
		Bool("enabled", geofence.Enabled).
		Msg("Geofence created successfully")

	return geofence, nil
}

// GetGeofence retrieves a geofence by ID
func (gs *GeofenceService) GetGeofence(ctx context.Context, geofenceID uuid.UUID) (*Geofence, error) {
	gs.mu.RLock()
	geofence, exists := gs.geofences[geofenceID]
	gs.mu.RUnlock()

	if exists {
		return geofence, nil
	}

	// Load from database if not in cache
	return gs.getGeofenceFromDatabase(ctx, geofenceID)
}

// ListGeofences returns geofences for a group
func (gs *GeofenceService) ListGeofences(ctx context.Context, groupID string, limit, offset int) ([]*Geofence, error) {
	return gs.listGeofencesFromDatabase(ctx, groupID, limit, offset)
}

// UpdateGeofence updates an existing geofence
func (gs *GeofenceService) UpdateGeofence(ctx context.Context, geofenceID uuid.UUID, updates map[string]interface{}) (*Geofence, error) {
	if err := gs.updateGeofenceInDatabase(ctx, geofenceID, updates); err != nil {
		return nil, err
	}

	// Update in-memory cache
	if geofence, err := gs.getGeofenceFromDatabase(ctx, geofenceID); err == nil {
		gs.mu.Lock()
		gs.geofences[geofenceID] = geofence
		gs.mu.Unlock()
	}

	return gs.GetGeofence(ctx, geofenceID)
}

// DeleteGeofence removes a geofence
func (gs *GeofenceService) DeleteGeofence(ctx context.Context, geofenceID uuid.UUID) error {
	if err := gs.deleteGeofenceFromDatabase(ctx, geofenceID); err != nil {
		return err
	}

	// Remove from in-memory cache
	gs.mu.Lock()
	delete(gs.geofences, geofenceID)
	gs.mu.Unlock()

	gs.logger.Info().
		Str("geofence_id", geofenceID.String()).
		Msg("Geofence deleted successfully")

	return nil
}

// CheckEntityPosition checks if an entity position violates any geofences
func (gs *GeofenceService) CheckEntityPosition(entityID string, position Point) []*GeofenceViolation {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	var violations []*GeofenceViolation

	// Get or create entity state
	state, exists := gs.entityStates[entityID]
	if !exists {
		state = &EntityGeofenceState{
			EntityID:        entityID,
			InsideGeofences: make(map[uuid.UUID]bool),
		}
		gs.entityStates[entityID] = state
	}

	_ = state.LastPosition // Track previous position for potential movement detection
	state.LastPosition = position
	state.LastUpdate = time.Now()

	// Check each active geofence
	for _, geofence := range gs.geofences {
		if !geofence.Enabled {
			continue
		}

		wasInside := state.InsideGeofences[geofence.ID]
		isInside := gs.isPointInGeofence(position, geofence)
		
		// Update state
		state.InsideGeofences[geofence.ID] = isInside

		// Check for violations
		if !wasInside && isInside && geofence.AlertOnEnter {
			// Entity entered geofence
			violation := &GeofenceViolation{
				ID:            uuid.New(),
				GeofenceID:    geofence.ID,
				EntityID:      entityID,
				ViolationType: ViolationEnter,
				Position:      position,
				Timestamp:     time.Now(),
				Acknowledged:  false,
			}
			violations = append(violations, violation)
			
			// Trigger alert callbacks
			for _, callback := range gs.alertCallbacks {
				go callback(violation, geofence)
			}
		} else if wasInside && !isInside && geofence.AlertOnExit {
			// Entity exited geofence
			violation := &GeofenceViolation{
				ID:            uuid.New(),
				GeofenceID:    geofence.ID,
				EntityID:      entityID,
				ViolationType: ViolationExit,
				Position:      position,
				Timestamp:     time.Now(),
				Acknowledged:  false,
			}
			violations = append(violations, violation)
			
			// Trigger alert callbacks
			for _, callback := range gs.alertCallbacks {
				go callback(violation, geofence)
			}
		}
	}

	// Save violations to database
	if len(violations) > 0 {
		go gs.saveViolationsToDatabase(violations)
	}

	return violations
}

// AddAlertCallback adds a callback for geofence alerts
func (gs *GeofenceService) AddAlertCallback(callback AlertCallback) {
	gs.alertCallbacks = append(gs.alertCallbacks, callback)
}

// GetViolations retrieves geofence violations
func (gs *GeofenceService) GetViolations(ctx context.Context, geofenceID *uuid.UUID, entityID *string, limit, offset int) ([]*GeofenceViolation, error) {
	return gs.getViolationsFromDatabase(ctx, geofenceID, entityID, limit, offset)
}

// AcknowledgeViolation marks a violation as acknowledged
func (gs *GeofenceService) AcknowledgeViolation(ctx context.Context, violationID uuid.UUID, acknowledgedBy uuid.UUID) error {
	return gs.acknowledgeViolationInDatabase(ctx, violationID, acknowledgedBy)
}

// isPointInGeofence checks if a point is inside a geofence
func (gs *GeofenceService) isPointInGeofence(point Point, geofence *Geofence) bool {
	switch geofence.Type {
	case GeofenceTypeCircle:
		return gs.isPointInCircle(point, geofence.Geometry)
	case GeofenceTypePolygon:
		return gs.isPointInPolygon(point, geofence.Geometry)
	case GeofenceTypeRectangle:
		return gs.isPointInRectangle(point, geofence.Geometry)
	}
	return false
}

// isPointInCircle checks if a point is inside a circular geofence
func (gs *GeofenceService) isPointInCircle(point Point, geometry interface{}) bool {
	circleData, ok := geometry.(map[string]interface{})
	if !ok {
		return false
	}

	center, ok := circleData["center"].(map[string]interface{})
	if !ok {
		return false
	}

	centerLat, _ := center["lat"].(float64)
	centerLng, _ := center["lng"].(float64)
	radius, _ := circleData["radius"].(float64)

	centerPoint := Point{Lat: centerLat, Lng: centerLng}
	distance := gs.calculateDistance(point, centerPoint)
	
	return distance <= radius
}

// isPointInPolygon checks if a point is inside a polygonal geofence using ray casting
func (gs *GeofenceService) isPointInPolygon(point Point, geometry interface{}) bool {
	polygonData, ok := geometry.(map[string]interface{})
	if !ok {
		return false
	}

	coordinates, ok := polygonData["coordinates"].([]interface{})
	if !ok || len(coordinates) == 0 {
		return false
	}

	// Get outer ring (first coordinate array)
	outerRing, ok := coordinates[0].([]interface{})
	if !ok {
		return false
	}

	// Convert to points
	var polygon []Point
	for _, coord := range outerRing {
		if coordArray, ok := coord.([]interface{}); ok && len(coordArray) >= 2 {
			if lng, ok1 := coordArray[0].(float64); ok1 {
				if lat, ok2 := coordArray[1].(float64); ok2 {
					polygon = append(polygon, Point{Lat: lat, Lng: lng})
				}
			}
		}
	}

	if len(polygon) < 3 {
		return false
	}

	// Ray casting algorithm
	inside := false
	j := len(polygon) - 1

	for i := 0; i < len(polygon); i++ {
		xi, yi := polygon[i].Lng, polygon[i].Lat
		xj, yj := polygon[j].Lng, polygon[j].Lat

		if ((yi > point.Lat) != (yj > point.Lat)) &&
			(point.Lng < (xj-xi)*(point.Lat-yi)/(yj-yi)+xi) {
			inside = !inside
		}
		j = i
	}

	return inside
}

// isPointInRectangle checks if a point is inside a rectangular geofence
func (gs *GeofenceService) isPointInRectangle(point Point, geometry interface{}) bool {
	rectData, ok := geometry.(map[string]interface{})
	if !ok {
		return false
	}

	bounds, ok := rectData["bounds"].(map[string]interface{})
	if !ok {
		return false
	}

	north, _ := bounds["north"].(float64)
	south, _ := bounds["south"].(float64)
	east, _ := bounds["east"].(float64)
	west, _ := bounds["west"].(float64)

	return point.Lat >= south && point.Lat <= north &&
		point.Lng >= west && point.Lng <= east
}

// calculateDistance calculates distance between two points using Haversine formula
func (gs *GeofenceService) calculateDistance(p1, p2 Point) float64 {
	const earthRadius = 6371000 // Earth's radius in meters

	lat1Rad := p1.Lat * math.Pi / 180
	lat2Rad := p2.Lat * math.Pi / 180
	deltaLatRad := (p2.Lat - p1.Lat) * math.Pi / 180
	deltaLngRad := (p2.Lng - p1.Lng) * math.Pi / 180

	a := math.Sin(deltaLatRad/2)*math.Sin(deltaLatRad/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*
			math.Sin(deltaLngRad/2)*math.Sin(deltaLngRad/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadius * c
}

// validateGeofenceGeometry validates geofence geometry based on type
func (gs *GeofenceService) validateGeofenceGeometry(geoType GeofenceType, geometry interface{}) error {
	switch geoType {
	case GeofenceTypeCircle:
		return gs.validateCircleGeometry(geometry)
	case GeofenceTypePolygon:
		return gs.validatePolygonGeometry(geometry)
	case GeofenceTypeRectangle:
		return gs.validateRectangleGeometry(geometry)
	}
	return fmt.Errorf("unsupported geofence type: %s", geoType)
}

func (gs *GeofenceService) validateCircleGeometry(geometry interface{}) error {
	circleData, ok := geometry.(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid circle geometry format")
	}

	center, ok := circleData["center"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("circle geometry missing center")
	}

	if _, ok := center["lat"].(float64); !ok {
		return fmt.Errorf("circle center missing latitude")
	}

	if _, ok := center["lng"].(float64); !ok {
		return fmt.Errorf("circle center missing longitude")
	}

	if radius, ok := circleData["radius"].(float64); !ok || radius <= 0 {
		return fmt.Errorf("circle geometry missing valid radius")
	}

	return nil
}

func (gs *GeofenceService) validatePolygonGeometry(geometry interface{}) error {
	polygonData, ok := geometry.(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid polygon geometry format")
	}

	coordinates, ok := polygonData["coordinates"].([]interface{})
	if !ok || len(coordinates) == 0 {
		return fmt.Errorf("polygon geometry missing coordinates")
	}

	// Validate outer ring
	outerRing, ok := coordinates[0].([]interface{})
	if !ok || len(outerRing) < 4 {
		return fmt.Errorf("polygon outer ring must have at least 4 coordinates")
	}

	return nil
}

func (gs *GeofenceService) validateRectangleGeometry(geometry interface{}) error {
	rectData, ok := geometry.(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid rectangle geometry format")
	}

	bounds, ok := rectData["bounds"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("rectangle geometry missing bounds")
	}

	north, northOk := bounds["north"].(float64)
	south, southOk := bounds["south"].(float64)
	east, eastOk := bounds["east"].(float64)
	west, westOk := bounds["west"].(float64)

	if !northOk || !southOk || !eastOk || !westOk {
		return fmt.Errorf("rectangle bounds missing required coordinates")
	}

	if north <= south || east <= west {
		return fmt.Errorf("invalid rectangle bounds")
	}

	return nil
}

// ViolationMonitor monitors entity positions for geofence violations
type ViolationMonitor struct {
	geofenceService *GeofenceService
	logger          *logger.Logger
	ticker          *time.Ticker
	stopChan        chan struct{}
	running         bool
}

// NewViolationMonitor creates a new violation monitor
func NewViolationMonitor(gs *GeofenceService, logger *logger.Logger) *ViolationMonitor {
	return &ViolationMonitor{
		geofenceService: gs,
		logger:          logger,
		stopChan:        make(chan struct{}),
	}
}

// Start begins monitoring for violations
func (vm *ViolationMonitor) Start() error {
	if vm.running {
		return fmt.Errorf("monitor already running")
	}

	vm.ticker = time.NewTicker(5 * time.Second) // Check every 5 seconds
	vm.running = true

	go vm.monitorLoop()
	return nil
}

// Stop stops the violation monitor
func (vm *ViolationMonitor) Stop() error {
	if !vm.running {
		return nil
	}

	close(vm.stopChan)
	if vm.ticker != nil {
		vm.ticker.Stop()
	}
	vm.running = false
	return nil
}

// monitorLoop continuously monitors entity positions
func (vm *ViolationMonitor) monitorLoop() {
	for {
		select {
		case <-vm.ticker.C:
			vm.checkAllEntities()
		case <-vm.stopChan:
			return
		}
	}
}

// checkAllEntities checks all active entities against geofences
func (vm *ViolationMonitor) checkAllEntities() {
	if vm.geofenceService.positionSvc == nil {
		return
	}

	// Get all active entity positions
	positions := vm.geofenceService.positionSvc.GetAllPositions()
	
	for _, entityPos := range positions {
		if !entityPos.IsStale {
			point := Point{Lat: entityPos.Lat, Lng: entityPos.Lng}
			vm.geofenceService.CheckEntityPosition(entityPos.EntityID, point)
		}
	}
}

// Database operations
func (gs *GeofenceService) loadGeofencesFromDatabase() {
	ctx := context.Background()
	geofences, err := gs.listGeofencesFromDatabase(ctx, "", 1000, 0)
	if err != nil {
		gs.logger.Error().Err(err).Msg("Failed to load geofences from database")
		return
	}

	gs.mu.Lock()
	for _, geofence := range geofences {
		gs.geofences[geofence.ID] = geofence
	}
	gs.mu.Unlock()

	gs.logger.Info().Int("count", len(geofences)).Msg("Loaded geofences from database")
}

func (gs *GeofenceService) saveGeofenceToDatabase(ctx context.Context, geofence *Geofence) error {
	query := `
		INSERT INTO geofences (
			id, name, description, type, geometry, enabled, 
			alert_on_enter, alert_on_exit, created_by, group_id,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`
	
	geometryJSON, _ := json.Marshal(geofence.Geometry)
	
	_, err := gs.db.ExecContext(ctx, query,
		geofence.ID, geofence.Name, geofence.Description, geofence.Type, geometryJSON,
		geofence.Enabled, geofence.AlertOnEnter, geofence.AlertOnExit, geofence.CreatedBy,
		geofence.GroupID, geofence.CreatedAt, geofence.UpdatedAt,
	)
	
	return err
}

func (gs *GeofenceService) getGeofenceFromDatabase(ctx context.Context, geofenceID uuid.UUID) (*Geofence, error) {
	query := `
		SELECT id, name, description, type, geometry, enabled,
			   alert_on_enter, alert_on_exit, created_by, group_id,
			   created_at, updated_at
		FROM geofences WHERE id = $1
	`
	
	var geofence Geofence
	var geometryJSON []byte
	
	err := gs.db.QueryRowContext(ctx, query, geofenceID).Scan(
		&geofence.ID, &geofence.Name, &geofence.Description, &geofence.Type,
		&geometryJSON, &geofence.Enabled, &geofence.AlertOnEnter, &geofence.AlertOnExit,
		&geofence.CreatedBy, &geofence.GroupID, &geofence.CreatedAt, &geofence.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	
	json.Unmarshal(geometryJSON, &geofence.Geometry)
	return &geofence, nil
}

func (gs *GeofenceService) listGeofencesFromDatabase(ctx context.Context, groupID string, limit, offset int) ([]*Geofence, error) {
	var query string
	var args []interface{}
	
	if groupID != "" {
		query = `
			SELECT id, name, description, type, geometry, enabled,
				   alert_on_enter, alert_on_exit, created_by, group_id,
				   created_at, updated_at
			FROM geofences 
			WHERE group_id = $1 
			ORDER BY created_at DESC 
			LIMIT $2 OFFSET $3
		`
		args = []interface{}{groupID, limit, offset}
	} else {
		query = `
			SELECT id, name, description, type, geometry, enabled,
				   alert_on_enter, alert_on_exit, created_by, group_id,
				   created_at, updated_at
			FROM geofences 
			ORDER BY created_at DESC 
			LIMIT $1 OFFSET $2
		`
		args = []interface{}{limit, offset}
	}
	
	rows, err := gs.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var geofences []*Geofence
	for rows.Next() {
		var geofence Geofence
		var geometryJSON []byte
		
		err := rows.Scan(
			&geofence.ID, &geofence.Name, &geofence.Description, &geofence.Type,
			&geometryJSON, &geofence.Enabled, &geofence.AlertOnEnter, &geofence.AlertOnExit,
			&geofence.CreatedBy, &geofence.GroupID, &geofence.CreatedAt, &geofence.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		
		json.Unmarshal(geometryJSON, &geofence.Geometry)
		geofences = append(geofences, &geofence)
	}

	return geofences, nil
}

func (gs *GeofenceService) updateGeofenceInDatabase(ctx context.Context, geofenceID uuid.UUID, updates map[string]interface{}) error {
	// Similar to route service update implementation
	// Build dynamic update query...
	return nil
}

func (gs *GeofenceService) deleteGeofenceFromDatabase(ctx context.Context, geofenceID uuid.UUID) error {
	tx, err := gs.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Delete violations first
	_, err = tx.ExecContext(ctx, "DELETE FROM geofence_violations WHERE geofence_id = $1", geofenceID)
	if err != nil {
		return err
	}

	// Delete geofence
	_, err = tx.ExecContext(ctx, "DELETE FROM geofences WHERE id = $1", geofenceID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (gs *GeofenceService) saveViolationsToDatabase(violations []*GeofenceViolation) {
	ctx := context.Background()
	query := `
		INSERT INTO geofence_violations (
			id, geofence_id, entity_id, violation_type, position, timestamp, acknowledged
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	
	for _, violation := range violations {
		positionJSON, _ := json.Marshal(violation.Position)
		
		_, err := gs.db.ExecContext(ctx, query,
			violation.ID, violation.GeofenceID, violation.EntityID, violation.ViolationType,
			positionJSON, violation.Timestamp, violation.Acknowledged,
		)
		if err != nil {
			gs.logger.Error().Err(err).Msg("Failed to save geofence violation")
		}
	}
}

func (gs *GeofenceService) getViolationsFromDatabase(ctx context.Context, geofenceID *uuid.UUID, entityID *string, limit, offset int) ([]*GeofenceViolation, error) {
	// Implementation for retrieving violations with filtering...
	return nil, nil
}

func (gs *GeofenceService) acknowledgeViolationInDatabase(ctx context.Context, violationID uuid.UUID, acknowledgedBy uuid.UUID) error {
	query := `
		UPDATE geofence_violations 
		SET acknowledged = true, acknowledged_by = $1, acknowledged_at = $2
		WHERE id = $3
	`
	
	_, err := gs.db.ExecContext(ctx, query, acknowledgedBy, time.Now(), violationID)
	return err
}
