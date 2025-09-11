package mapping

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/dfedick/gotak/pkg/database"
	"github.com/dfedick/gotak/pkg/logger"
)

// RouteService provides route planning and navigation functionality
type RouteService struct {
	db         database.DB
	logger     *logger.Logger
	calculator *RouteCalculator
	osrmURL    string
	httpClient *http.Client
}

// NewRouteService creates a new route service
func NewRouteService(db database.DB, logger *logger.Logger, osrmURL string) *RouteService {
	calculator := NewRouteCalculator(osrmURL, logger)
	
	return &RouteService{
		db:         db,
		logger:     logger,
		calculator: calculator,
		osrmURL:    osrmURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// CreateRoute creates a new route with waypoints
func (rs *RouteService) CreateRoute(ctx context.Context, req *CreateRouteRequest, createdBy uuid.UUID, groupID string) (*Route, error) {
	// Validate waypoints
	if len(req.Waypoints) < 2 {
		return nil, fmt.Errorf("route requires at least 2 waypoints")
	}

	// Calculate route using OSRM
	routeData, err := rs.calculator.CalculateRoute(req.Waypoints, req.Options)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate route: %w", err)
	}

	// Create route record
	route := &Route{
		ID:          uuid.New(),
		Name:        req.Name,
		Description: req.Description,
		CreatedBy:   createdBy,
		GroupID:     groupID,
		Geometry:    routeData.Geometry,
		Distance:    routeData.Distance,
		Duration:    routeData.Duration,
		RouteType:   req.Options.RouteType,
		Vehicle:     req.Options.Vehicle,
		Optimize:    req.Options.Optimize,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Create waypoints
	waypoints := make([]Waypoint, len(req.Waypoints))
	for i, wp := range req.Waypoints {
		waypoints[i] = Waypoint{
			ID:        uuid.New(),
			RouteID:   route.ID,
			Sequence:  i,
			Lat:       wp.Lat,
			Lng:       wp.Lng,
			Name:      fmt.Sprintf("Waypoint %d", i+1),
			CreatedAt: time.Now(),
		}
	}
	route.Waypoints = waypoints

	// Save to database
	if err := rs.saveRouteToDatabase(ctx, route); err != nil {
		return nil, fmt.Errorf("failed to save route: %w", err)
	}

	rs.logger.Info().
		Str("route_id", route.ID.String()).
		Str("name", route.Name).
		Int("waypoints", len(waypoints)).
		Float64("distance_km", route.Distance/1000).
		Msg("Route created successfully")

	return route, nil
}

// GetRoute retrieves a route by ID
func (rs *RouteService) GetRoute(ctx context.Context, routeID uuid.UUID) (*Route, error) {
	route, err := rs.getRouteFromDatabase(ctx, routeID)
	if err != nil {
		return nil, err
	}
	return route, nil
}

// ListRoutes returns routes for a group with optional filtering
func (rs *RouteService) ListRoutes(ctx context.Context, groupID string, limit, offset int) ([]*Route, error) {
	routes, err := rs.listRoutesFromDatabase(ctx, groupID, limit, offset)
	if err != nil {
		return nil, err
	}
	return routes, nil
}

// UpdateRoute updates an existing route
func (rs *RouteService) UpdateRoute(ctx context.Context, routeID uuid.UUID, updates map[string]interface{}) (*Route, error) {
	if err := rs.updateRouteInDatabase(ctx, routeID, updates); err != nil {
		return nil, err
	}
	return rs.GetRoute(ctx, routeID)
}

// DeleteRoute removes a route
func (rs *RouteService) DeleteRoute(ctx context.Context, routeID uuid.UUID) error {
	return rs.deleteRouteFromDatabase(ctx, routeID)
}

// RecalculateRoute recalculates an existing route with new options
func (rs *RouteService) RecalculateRoute(ctx context.Context, routeID uuid.UUID, options RouteOptions) (*Route, error) {
	route, err := rs.GetRoute(ctx, routeID)
	if err != nil {
		return nil, err
	}

	// Convert waypoints to points
	points := make([]Point, len(route.Waypoints))
	for i, wp := range route.Waypoints {
		points[i] = Point{Lat: wp.Lat, Lng: wp.Lng}
	}

	// Recalculate route
	routeData, err := rs.calculator.CalculateRoute(points, options)
	if err != nil {
		return nil, fmt.Errorf("failed to recalculate route: %w", err)
	}

	// Update route data
	updates := map[string]interface{}{
		"geometry":    routeData.Geometry,
		"distance":    routeData.Distance,
		"duration":    routeData.Duration.Nanoseconds(),
		"route_type":  options.RouteType,
		"vehicle":     options.Vehicle,
		"optimize":    options.Optimize,
		"updated_at":  time.Now(),
	}

	return rs.UpdateRoute(ctx, routeID, updates)
}

// RouteCalculator handles route calculation using external routing services
type RouteCalculator struct {
	osrmURL    string
	logger     *logger.Logger
	httpClient *http.Client
}

// NewRouteCalculator creates a new route calculator
func NewRouteCalculator(osrmURL string, logger *logger.Logger) *RouteCalculator {
	return &RouteCalculator{
		osrmURL: osrmURL,
		logger:  logger,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// CalculateRoute calculates a route using OSRM
func (rc *RouteCalculator) CalculateRoute(waypoints []Point, options RouteOptions) (*RouteData, error) {
	if rc.osrmURL == "" {
		// Fallback to direct line calculation if OSRM not available
		return rc.calculateDirectRoute(waypoints), nil
	}

	// Build coordinates string for OSRM
	coords := make([]string, len(waypoints))
	for i, wp := range waypoints {
		coords[i] = fmt.Sprintf("%.6f,%.6f", wp.Lng, wp.Lat)
	}
	coordsStr := strings.Join(coords, ";")

	// Map our vehicle types to OSRM profiles
	profile := rc.mapVehicleTypeToProfile(options.Vehicle)
	
	// Build OSRM URL
	osrmURL := fmt.Sprintf("%s/route/v1/%s/%s", 
		strings.TrimSuffix(rc.osrmURL, "/"), 
		profile, 
		coordsStr)

	// Add query parameters
	params := url.Values{}
	params.Set("overview", "full")
	params.Set("geometries", "geojson")
	params.Set("steps", "true")
	
	if options.Optimize {
		params.Set("optimize", "true")
	}

	osrmURL += "?" + params.Encode()

	// Make request to OSRM
	resp, err := rc.httpClient.Get(osrmURL)
	if err != nil {
		rc.logger.Error().Err(err).Msg("Failed to connect to OSRM service")
		// Fallback to direct route
		return rc.calculateDirectRoute(waypoints), nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		rc.logger.Error().Int("status", resp.StatusCode).Msg("OSRM service returned error")
		// Fallback to direct route
		return rc.calculateDirectRoute(waypoints), nil
	}

	// Parse OSRM response
	var osrmResp OSRMResponse
	if err := json.NewDecoder(resp.Body).Decode(&osrmResp); err != nil {
		rc.logger.Error().Err(err).Msg("Failed to parse OSRM response")
		return rc.calculateDirectRoute(waypoints), nil
	}

	if len(osrmResp.Routes) == 0 {
		return rc.calculateDirectRoute(waypoints), nil
	}

	route := osrmResp.Routes[0]
	
	return &RouteData{
		Geometry: LineString{
			Type:        "LineString",
			Coordinates: route.Geometry.Coordinates,
		},
		Distance: route.Distance,
		Duration: time.Duration(route.Duration) * time.Second,
	}, nil
}

// calculateDirectRoute creates a direct line route as fallback
func (rc *RouteCalculator) calculateDirectRoute(waypoints []Point) *RouteData {
	coordinates := make([][]float64, len(waypoints))
	totalDistance := 0.0

	for i, wp := range waypoints {
		coordinates[i] = []float64{wp.Lng, wp.Lat}
		
		if i > 0 {
			// Calculate distance from previous waypoint
			dist := rc.calculateDistance(waypoints[i-1], wp)
			totalDistance += dist
		}
	}

	// Estimate duration based on average speed (50 km/h)
	avgSpeedKmh := 50.0
	durationHours := totalDistance / 1000.0 / avgSpeedKmh
	duration := time.Duration(durationHours * float64(time.Hour))

	return &RouteData{
		Geometry: LineString{
			Type:        "LineString",
			Coordinates: coordinates,
		},
		Distance: totalDistance,
		Duration: duration,
	}
}

// calculateDistance calculates distance between two points using Haversine formula
func (rc *RouteCalculator) calculateDistance(p1, p2 Point) float64 {
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

// mapVehicleTypeToProfile maps our vehicle types to OSRM profiles
func (rc *RouteCalculator) mapVehicleTypeToProfile(vehicle VehicleType) string {
	switch vehicle {
	case VehicleTypeCar:
		return "driving"
	case VehicleTypeTruck:
		return "driving"
	case VehicleTypeBicycle:
		return "cycling"
	case VehicleTypeFoot:
		return "walking"
	case VehicleTypeMotorcycle:
		return "driving"
	default:
		return "driving"
	}
}

// RouteData contains calculated route information
type RouteData struct {
	Geometry LineString
	Distance float64
	Duration time.Duration
}

// OSRM API response structures
type OSRMResponse struct {
	Code   string      `json:"code"`
	Routes []OSRMRoute `json:"routes"`
}

type OSRMRoute struct {
	Geometry OSRMGeometry `json:"geometry"`
	Distance float64      `json:"distance"`
	Duration float64      `json:"duration"`
	Legs     []OSRMLeg    `json:"legs"`
}

type OSRMGeometry struct {
	Type        string      `json:"type"`
	Coordinates [][]float64 `json:"coordinates"`
}

type OSRMLeg struct {
	Distance float64    `json:"distance"`
	Duration float64    `json:"duration"`
	Steps    []OSRMStep `json:"steps"`
}

type OSRMStep struct {
	Distance     float64                `json:"distance"`
	Duration     float64                `json:"duration"`
	Geometry     OSRMGeometry           `json:"geometry"`
	Name         string                 `json:"name"`
	Mode         string                 `json:"mode"`
	Maneuver     OSRMManeuver           `json:"maneuver"`
	Intersections []OSRMIntersection    `json:"intersections"`
}

type OSRMManeuver struct {
	Type      string    `json:"type"`
	Modifier  string    `json:"modifier,omitempty"`
	Location  []float64 `json:"location"`
	Bearing   int       `json:"bearing_before"`
}

type OSRMIntersection struct {
	Location []float64 `json:"location"`
	Bearings []int     `json:"bearings"`
	Entry    []bool    `json:"entry"`
	In       int       `json:"in,omitempty"`
	Out      int       `json:"out,omitempty"`
}

// Database operations (implementation will depend on your database layer)

func (rs *RouteService) saveRouteToDatabase(ctx context.Context, route *Route) error {
	// Start transaction
	tx, err := rs.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Insert route
	query := `
		INSERT INTO routes (
			id, name, description, created_by, group_id, 
			geometry, distance, duration, route_type, vehicle, optimize,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`
	
	geometryJSON, _ := json.Marshal(route.Geometry)
	
	_, err = tx.ExecContext(ctx, query,
		route.ID, route.Name, route.Description, route.CreatedBy, route.GroupID,
		geometryJSON, route.Distance, route.Duration.Nanoseconds(), route.RouteType, route.Vehicle, route.Optimize,
		route.CreatedAt, route.UpdatedAt,
	)
	if err != nil {
		return err
	}

	// Insert waypoints
	waypointQuery := `
		INSERT INTO waypoints (
			id, route_id, sequence, lat, lng, name, description, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	
	for _, wp := range route.Waypoints {
		_, err = tx.ExecContext(ctx, waypointQuery,
			wp.ID, wp.RouteID, wp.Sequence, wp.Lat, wp.Lng, wp.Name, wp.Description, wp.CreatedAt,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (rs *RouteService) getRouteFromDatabase(ctx context.Context, routeID uuid.UUID) (*Route, error) {
	// Get route
	query := `
		SELECT id, name, description, created_by, group_id, 
			   geometry, distance, duration, route_type, vehicle, optimize,
			   created_at, updated_at
		FROM routes WHERE id = $1
	`
	
	var route Route
	var geometryJSON []byte
	var durationNs int64
	
	err := rs.db.QueryRowContext(ctx, query, routeID).Scan(
		&route.ID, &route.Name, &route.Description, &route.CreatedBy, &route.GroupID,
		&geometryJSON, &route.Distance, &durationNs, &route.RouteType, &route.Vehicle, &route.Optimize,
		&route.CreatedAt, &route.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	
	route.Duration = time.Duration(durationNs)
	json.Unmarshal(geometryJSON, &route.Geometry)

	// Get waypoints
	waypointQuery := `
		SELECT id, route_id, sequence, lat, lng, name, description, eta, created_at
		FROM waypoints WHERE route_id = $1 ORDER BY sequence
	`
	
	rows, err := rs.db.QueryContext(ctx, waypointQuery, routeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var waypoints []Waypoint
	for rows.Next() {
		var wp Waypoint
		err := rows.Scan(&wp.ID, &wp.RouteID, &wp.Sequence, &wp.Lat, &wp.Lng, 
			&wp.Name, &wp.Description, &wp.ETA, &wp.CreatedAt)
		if err != nil {
			return nil, err
		}
		waypoints = append(waypoints, wp)
	}
	
	route.Waypoints = waypoints
	return &route, nil
}

func (rs *RouteService) listRoutesFromDatabase(ctx context.Context, groupID string, limit, offset int) ([]*Route, error) {
	query := `
		SELECT id, name, description, created_by, group_id, 
			   geometry, distance, duration, route_type, vehicle, optimize,
			   created_at, updated_at
		FROM routes 
		WHERE group_id = $1 
		ORDER BY created_at DESC 
		LIMIT $2 OFFSET $3
	`
	
	rows, err := rs.db.QueryContext(ctx, query, groupID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var routes []*Route
	for rows.Next() {
		var route Route
		var geometryJSON []byte
		var durationNs int64
		
		err := rows.Scan(
			&route.ID, &route.Name, &route.Description, &route.CreatedBy, &route.GroupID,
			&geometryJSON, &route.Distance, &durationNs, &route.RouteType, &route.Vehicle, &route.Optimize,
			&route.CreatedAt, &route.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		
		route.Duration = time.Duration(durationNs)
		json.Unmarshal(geometryJSON, &route.Geometry)
		
		// Load waypoints for each route
		route.Waypoints, _ = rs.getWaypointsForRoute(ctx, route.ID)
		
		routes = append(routes, &route)
	}

	return routes, nil
}

func (rs *RouteService) getWaypointsForRoute(ctx context.Context, routeID uuid.UUID) ([]Waypoint, error) {
	query := `
		SELECT id, route_id, sequence, lat, lng, name, description, eta, created_at
		FROM waypoints WHERE route_id = $1 ORDER BY sequence
	`
	
	rows, err := rs.db.QueryContext(ctx, query, routeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var waypoints []Waypoint
	for rows.Next() {
		var wp Waypoint
		err := rows.Scan(&wp.ID, &wp.RouteID, &wp.Sequence, &wp.Lat, &wp.Lng, 
			&wp.Name, &wp.Description, &wp.ETA, &wp.CreatedAt)
		if err != nil {
			return nil, err
		}
		waypoints = append(waypoints, wp)
	}
	
	return waypoints, nil
}

func (rs *RouteService) updateRouteInDatabase(ctx context.Context, routeID uuid.UUID, updates map[string]interface{}) error {
	if len(updates) == 0 {
		return nil
	}

	// Build dynamic update query
	setParts := make([]string, 0, len(updates))
	args := make([]interface{}, 0, len(updates)+1)
	argIndex := 1

	for key, value := range updates {
		setParts = append(setParts, fmt.Sprintf("%s = $%d", key, argIndex))
		args = append(args, value)
		argIndex++
	}

	query := fmt.Sprintf("UPDATE routes SET %s WHERE id = $%d", 
		strings.Join(setParts, ", "), argIndex)
	args = append(args, routeID)

	_, err := rs.db.ExecContext(ctx, query, args...)
	return err
}

func (rs *RouteService) deleteRouteFromDatabase(ctx context.Context, routeID uuid.UUID) error {
	tx, err := rs.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Delete waypoints first
	_, err = tx.ExecContext(ctx, "DELETE FROM waypoints WHERE route_id = $1", routeID)
	if err != nil {
		return err
	}

	// Delete route
	_, err = tx.ExecContext(ctx, "DELETE FROM routes WHERE id = $1", routeID)
	if err != nil {
		return err
	}

	return tx.Commit()
}
