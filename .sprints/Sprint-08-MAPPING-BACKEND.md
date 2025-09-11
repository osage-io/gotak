# Sprint 8: Mapping Backend Integration

**Duration:** 2 weeks  
**Theme:** Backend API Integration for Advanced Mapping Features  
**Sprint Goals:** Implement backend APIs and database support for the mapping components completed in Sprint 7  
**Priority:** High - Critical for making Sprint 7 mapping features fully functional

---

## 🎯 Objectives

1. **Mapping APIs**: REST endpoints for routes, geofences, measurements, and offline maps
2. **Database Schema**: Tables and relationships for mapping data persistence
3. **Real-time Features**: WebSocket integration for live mapping updates
4. **Backend Services**: Business logic and data processing for mapping features
5. **Integration Testing**: End-to-end testing of frontend-backend mapping functionality

---

## 📋 User Stories

### Epic: Mapping Backend Platform

**US-8.1: Route Management API**
```
As a tactical operator using the route management interface
I want my routes to be saved and shared with my team
So that we can coordinate movements and reuse planned routes
```

**Acceptance Criteria:**
- ✅ REST API endpoints for route CRUD operations
- ✅ Database schema for routes, waypoints, and route metadata
- ✅ Route optimization integration with routing services
- ✅ Route sharing and permissions system
- ✅ Import/export functionality for route data
- ✅ Route history and versioning

**US-8.2: Geofence Management API**
```
As a security operator managing geofences
I want real-time violation alerts and persistent geofence data
So that I can maintain continuous area monitoring and security
```

**Acceptance Criteria:**
- ✅ REST API for geofence CRUD with spatial queries
- ✅ Database schema with PostGIS spatial extensions
- ✅ Real-time geofence violation detection and alerts
- ✅ WebSocket notifications for boundary events
- ✅ Geofence analytics and reporting
- ✅ Bulk geofence operations and management

**US-8.3: Measurement Tools Backend**
```
As a field operator using measurement tools
I want my measurements saved and accessible for reporting
So that I can maintain measurement history and generate tactical reports
```

**Acceptance Criteria:**
- ✅ REST API for measurement data persistence
- ✅ Database schema for measurements with spatial data
- ✅ Measurement calculations and validation on backend
- ✅ Export functionality for measurements and reports
- ✅ Measurement sharing between team members
- ✅ Measurement templates and standards

**US-8.4: Offline Map Management Backend**
```
As a field operator preparing for remote operations
I want centralized management of offline map downloads
So that teams can coordinate offline map coverage and updates
```

**Acceptance Criteria:**
- ✅ REST API for offline map job management
- ✅ Database schema for download jobs and tile metadata
- ✅ Background processing for tile downloads
- ✅ Storage management and cleanup automation
- ✅ Download job scheduling and prioritization
- ✅ Team coordination for offline map coverage

**US-8.5: Real-time Mapping Collaboration**
```
As a team member using mapping tools
I want to see real-time updates from other team members
So that we can collaborate effectively on tactical planning
```

**Acceptance Criteria:**
- ✅ WebSocket integration for real-time mapping updates
- ✅ Live route sharing and collaborative editing
- ✅ Real-time geofence violation notifications
- ✅ Shared measurement updates and annotations
- ✅ Collaborative offline map planning
- ✅ Team presence indicators on maps

---

## 🛠️ Technical Implementation

### Backend API Structure

**Route Management APIs**
```go
// internal/handlers/routes.go
func (h *Handler) SetupRouteRoutes(r chi.Router) {
    r.Route("/api/routes", func(r chi.Router) {
        r.Use(h.AuthMiddleware)
        
        // CRUD operations
        r.Post("/", h.CreateRoute)           // Create new route
        r.Get("/", h.ListRoutes)             // List routes with filtering
        r.Get("/{id}", h.GetRoute)           // Get specific route
        r.Put("/{id}", h.UpdateRoute)        // Update route
        r.Delete("/{id}", h.DeleteRoute)     // Delete route
        
        // Route operations
        r.Post("/{id}/optimize", h.OptimizeRoute)    // Optimize route
        r.Post("/{id}/share", h.ShareRoute)          // Share with team
        r.Get("/{id}/export", h.ExportRoute)         // Export route data
        
        // Waypoint management
        r.Post("/{id}/waypoints", h.AddWaypoint)     // Add waypoint
        r.Put("/{id}/waypoints/{wpId}", h.UpdateWaypoint)
        r.Delete("/{id}/waypoints/{wpId}", h.RemoveWaypoint)
    })
}

type RouteRequest struct {
    Name        string     `json:"name" validate:"required,max=100"`
    Description string     `json:"description" validate:"max=500"`
    Tags        []string   `json:"tags" validate:"dive,max=50"`
    Priority    Priority   `json:"priority" validate:"oneof=low medium high critical"`
    Waypoints   []Waypoint `json:"waypoints" validate:"required,min=2,dive"`
    RouteType   RouteType  `json:"route_type" validate:"required"`
    Vehicle     Vehicle    `json:"vehicle" validate:"required"`
    Optimize    bool       `json:"optimize"`
}

type RouteResponse struct {
    ID          uuid.UUID  `json:"id"`
    Name        string     `json:"name"`
    Description string     `json:"description"`
    CreatedBy   uuid.UUID  `json:"created_by"`
    GroupID     string     `json:"group_id"`
    
    // Route data
    Waypoints   []Waypoint `json:"waypoints"`
    Geometry    LineString `json:"geometry"`     // GeoJSON LineString
    Distance    float64    `json:"distance"`     // meters
    Duration    int        `json:"duration"`     // seconds
    
    // Metadata
    Tags        []string   `json:"tags"`
    Priority    Priority   `json:"priority"`
    RouteType   RouteType  `json:"route_type"`
    Vehicle     Vehicle    `json:"vehicle"`
    
    CreatedAt   time.Time  `json:"created_at"`
    UpdatedAt   time.Time  `json:"updated_at"`
}
```

**Geofence Management APIs**
```go
// internal/handlers/geofences.go
func (h *Handler) SetupGeofenceRoutes(r chi.Router) {
    r.Route("/api/geofences", func(r chi.Router) {
        r.Use(h.AuthMiddleware)
        
        // CRUD operations
        r.Post("/", h.CreateGeofence)
        r.Get("/", h.ListGeofences)
        r.Get("/{id}", h.GetGeofence)
        r.Put("/{id}", h.UpdateGeofence)
        r.Delete("/{id}", h.DeleteGeofence)
        
        // Geofence operations
        r.Post("/{id}/toggle", h.ToggleGeofence)     // Enable/disable
        r.Get("/{id}/violations", h.GetViolations)   // Get violation history
        r.Post("/bulk-toggle", h.BulkToggleGeofences)
        
        // Spatial queries
        r.Post("/intersect", h.FindIntersecting)     // Find overlapping geofences
        r.Post("/contains", h.FindContaining)        // Find geofences containing point
    })
}

type GeofenceRequest struct {
    Name        string            `json:"name" validate:"required,max=100"`
    Description string            `json:"description" validate:"max=500"`
    Type        GeofenceType      `json:"type" validate:"required,oneof=circle polygon rectangle"`
    Geometry    interface{}       `json:"geometry" validate:"required"`
    
    // Visual styling
    Color       string            `json:"color" validate:"required,hexcolor"`
    Opacity     float64           `json:"opacity" validate:"min=0,max=1"`
    BorderStyle GeofenceBorder    `json:"border_style"`
    
    // Monitoring settings
    Enabled     bool              `json:"enabled"`
    AlertOnEnter bool             `json:"alert_on_enter"`
    AlertOnExit  bool             `json:"alert_on_exit"`
    
    // Metadata
    Tags        []string          `json:"tags" validate:"dive,max=50"`
    Priority    Priority          `json:"priority"`
}
```

### Database Schema Design

**Routes and Waypoints Schema**
```sql
-- migrations/008_create_routes_tables.up.sql

-- Routes table
CREATE TABLE routes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    created_by UUID NOT NULL REFERENCES users(id),
    group_id VARCHAR(50) NOT NULL,
    
    -- Route configuration
    route_type VARCHAR(20) NOT NULL CHECK (route_type IN ('fastest', 'shortest', 'tactical', 'offroad')),
    vehicle VARCHAR(20) NOT NULL CHECK (vehicle IN ('foot', 'vehicle', 'boat', 'aircraft')),
    optimize BOOLEAN DEFAULT false,
    
    -- Calculated route data
    geometry GEOMETRY(LINESTRING, 4326),  -- PostGIS spatial data
    distance REAL,                        -- meters
    duration INTEGER,                     -- seconds
    
    -- Metadata
    tags TEXT[] DEFAULT '{}',
    priority VARCHAR(10) DEFAULT 'medium' CHECK (priority IN ('low', 'medium', 'high', 'critical')),
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Waypoints table
CREATE TABLE waypoints (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    route_id UUID NOT NULL REFERENCES routes(id) ON DELETE CASCADE,
    sequence INTEGER NOT NULL,
    
    -- Position
    position GEOMETRY(POINT, 4326) NOT NULL,  -- PostGIS spatial data
    lat DOUBLE PRECISION NOT NULL,
    lng DOUBLE PRECISION NOT NULL,
    altitude REAL,
    
    -- Waypoint data
    name VARCHAR(100),
    description TEXT,
    waypoint_type VARCHAR(20) DEFAULT 'waypoint' CHECK (waypoint_type IN ('start', 'waypoint', 'checkpoint', 'destination')),
    
    -- Timing
    eta TIMESTAMPTZ,
    stop_duration INTEGER DEFAULT 0,  -- seconds
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    
    UNIQUE(route_id, sequence)
);

-- Indexes for performance
CREATE INDEX idx_routes_created_by ON routes(created_by);
CREATE INDEX idx_routes_group_id ON routes(group_id);
CREATE INDEX idx_routes_geometry ON routes USING GIST(geometry);
CREATE INDEX idx_routes_tags ON routes USING GIN(tags);
CREATE INDEX idx_waypoints_route_id ON waypoints(route_id);
CREATE INDEX idx_waypoints_position ON waypoints USING GIST(position);
```

**Geofences Schema**
```sql
-- migrations/008_create_geofences_tables.up.sql

-- Geofences table
CREATE TABLE geofences (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    created_by UUID NOT NULL REFERENCES users(id),
    group_id VARCHAR(50) NOT NULL,
    
    -- Geofence type and geometry
    geofence_type VARCHAR(20) NOT NULL CHECK (geofence_type IN ('circle', 'polygon', 'rectangle')),
    geometry GEOMETRY(POLYGON, 4326) NOT NULL,  -- PostGIS spatial data
    
    -- Visual styling
    color VARCHAR(7) NOT NULL DEFAULT '#ff0000',  -- Hex color
    opacity REAL DEFAULT 0.3 CHECK (opacity >= 0 AND opacity <= 1),
    border_style VARCHAR(20) DEFAULT 'solid' CHECK (border_style IN ('solid', 'dashed', 'dotted')),
    border_width INTEGER DEFAULT 2,
    
    -- Monitoring settings
    enabled BOOLEAN DEFAULT true,
    alert_on_enter BOOLEAN DEFAULT true,
    alert_on_exit BOOLEAN DEFAULT false,
    
    -- Metadata
    tags TEXT[] DEFAULT '{}',
    priority VARCHAR(10) DEFAULT 'medium' CHECK (priority IN ('low', 'medium', 'high', 'critical')),
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Geofence violations table
CREATE TABLE geofence_violations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    geofence_id UUID NOT NULL REFERENCES geofences(id) ON DELETE CASCADE,
    entity_id VARCHAR(100) NOT NULL,  -- Can be user, vehicle, or other entity
    
    -- Violation data
    violation_type VARCHAR(10) NOT NULL CHECK (violation_type IN ('enter', 'exit')),
    position GEOMETRY(POINT, 4326) NOT NULL,
    lat DOUBLE PRECISION NOT NULL,
    lng DOUBLE PRECISION NOT NULL,
    
    -- Timing
    detected_at TIMESTAMPTZ DEFAULT NOW(),
    acknowledged BOOLEAN DEFAULT false,
    acknowledged_by UUID REFERENCES users(id),
    acknowledged_at TIMESTAMPTZ,
    
    -- Context
    metadata JSONB DEFAULT '{}'
);

-- Indexes for spatial queries and performance
CREATE INDEX idx_geofences_created_by ON geofences(created_by);
CREATE INDEX idx_geofences_group_id ON geofences(group_id);
CREATE INDEX idx_geofences_geometry ON geofences USING GIST(geometry);
CREATE INDEX idx_geofences_enabled ON geofences(enabled) WHERE enabled = true;
CREATE INDEX idx_violations_geofence_id ON geofence_violations(geofence_id);
CREATE INDEX idx_violations_entity_id ON geofence_violations(entity_id);
CREATE INDEX idx_violations_detected_at ON geofence_violations(detected_at);
```

### Real-time WebSocket Integration

**WebSocket Handler for Mapping Updates**
```go
// internal/handlers/websocket_mapping.go
type MappingWSHub struct {
    clients    map[*MappingWSClient]bool
    register   chan *MappingWSClient
    unregister chan *MappingWSClient
    broadcast  chan *MappingUpdate
    
    // Geofence monitoring
    geofenceMonitor *GeofenceMonitor
    
    logger *logger.Logger
}

type MappingUpdate struct {
    Type      MappingUpdateType `json:"type"`
    Timestamp time.Time         `json:"timestamp"`
    UserID    uuid.UUID         `json:"user_id"`
    GroupID   string            `json:"group_id"`
    Data      interface{}       `json:"data"`
}

type MappingUpdateType string
const (
    UpdateTypeRouteCreated      MappingUpdateType = "route_created"
    UpdateTypeRouteUpdated      MappingUpdateType = "route_updated"
    UpdateTypeRouteDeleted      MappingUpdateType = "route_deleted"
    UpdateTypeGeofenceViolation MappingUpdateType = "geofence_violation"
    UpdateTypeGeofenceCreated   MappingUpdateType = "geofence_created"
    UpdateTypeGeofenceUpdated   MappingUpdateType = "geofence_updated"
    UpdateTypeMeasurementShared MappingUpdateType = "measurement_shared"
)

func (h *MappingWSHub) BroadcastRouteUpdate(route *Route, updateType MappingUpdateType) {
    update := &MappingUpdate{
        Type:      updateType,
        Timestamp: time.Now(),
        UserID:    route.CreatedBy,
        GroupID:   route.GroupID,
        Data:      route,
    }
    
    select {
    case h.broadcast <- update:
    default:
        h.logger.Warn().Msg("Failed to broadcast route update - channel full")
    }
}
```

### Backend Services Architecture

**Route Service Implementation**
```go
// internal/services/route_service.go
type RouteService struct {
    storage      storage.Storage
    routingAPI   routing.Client     // OSRM or similar
    logger       *logger.Logger
    wsHub        *MappingWSHub
}

func (s *RouteService) CreateRoute(ctx context.Context, req *CreateRouteRequest) (*Route, error) {
    // Validate waypoints
    if len(req.Waypoints) < 2 {
        return nil, ErrInsufficientWaypoints
    }
    
    // Create route entity
    route := &Route{
        ID:          uuid.New(),
        Name:        req.Name,
        Description: req.Description,
        CreatedBy:   req.UserID,
        GroupID:     req.GroupID,
        RouteType:   req.RouteType,
        Vehicle:     req.Vehicle,
        Waypoints:   req.Waypoints,
        Tags:        req.Tags,
        Priority:    req.Priority,
        CreatedAt:   time.Now(),
        UpdatedAt:   time.Now(),
    }
    
    // Calculate route if optimization requested
    if req.Optimize {
        if err := s.calculateOptimalRoute(ctx, route); err != nil {
            s.logger.Warn().Err(err).Msg("Failed to optimize route, using direct path")
            s.calculateDirectRoute(route)
        }
    } else {
        s.calculateDirectRoute(route)
    }
    
    // Save to database
    if err := s.storage.CreateRoute(ctx, route); err != nil {
        return nil, fmt.Errorf("failed to create route: %w", err)
    }
    
    // Broadcast update to WebSocket clients
    s.wsHub.BroadcastRouteUpdate(route, UpdateTypeRouteCreated)
    
    s.logger.Info().
        Str("route_id", route.ID.String()).
        Str("name", route.Name).
        Int("waypoints", len(route.Waypoints)).
        Msg("Route created successfully")
    
    return route, nil
}

func (s *RouteService) calculateOptimalRoute(ctx context.Context, route *Route) error {
    // Extract coordinates for routing API
    coordinates := make([]routing.Coordinate, len(route.Waypoints))
    for i, wp := range route.Waypoints {
        coordinates[i] = routing.Coordinate{
            Lat: wp.Lat,
            Lng: wp.Lng,
        }
    }
    
    // Request route from OSRM or similar service
    routeResult, err := s.routingAPI.CalculateRoute(ctx, coordinates, routing.Options{
        Profile: string(route.Vehicle),
        Steps:   true,
        Overview: routing.OverviewFull,
    })
    if err != nil {
        return fmt.Errorf("routing API failed: %w", err)
    }
    
    // Convert to our geometry format
    route.Geometry = routeResult.Geometry
    route.Distance = routeResult.Distance
    route.Duration = time.Duration(routeResult.Duration) * time.Second
    
    return nil
}
```

### Performance and Optimization

**Database Query Optimization**
```go
// internal/storage/postgres/routes.go
func (s *PostgresStorage) ListRoutes(ctx context.Context, filter RouteFilter) ([]*Route, int, error) {
    query := `
        SELECT 
            r.id, r.name, r.description, r.created_by, r.group_id,
            r.route_type, r.vehicle, r.optimize,
            ST_AsGeoJSON(r.geometry) as geometry,
            r.distance, r.duration,
            r.tags, r.priority,
            r.created_at, r.updated_at,
            COUNT(*) OVER() as total_count
        FROM routes r
        WHERE 1=1
    `
    
    args := []interface{}{}
    argIndex := 0
    
    // Apply filters with proper indexing
    if filter.GroupID != "" {
        argIndex++
        query += fmt.Sprintf(" AND r.group_id = $%d", argIndex)
        args = append(args, filter.GroupID)
    }
    
    if filter.CreatedBy != nil {
        argIndex++
        query += fmt.Sprintf(" AND r.created_by = $%d", argIndex)
        args = append(args, *filter.CreatedBy)
    }
    
    if len(filter.Tags) > 0 {
        argIndex++
        query += fmt.Sprintf(" AND r.tags && $%d", argIndex)
        args = append(args, pq.Array(filter.Tags))
    }
    
    if filter.Priority != nil {
        argIndex++
        query += fmt.Sprintf(" AND r.priority = $%d", argIndex)
        args = append(args, *filter.Priority)
    }
    
    // Spatial filter for area-based queries
    if filter.BoundingBox != nil {
        argIndex++
        query += fmt.Sprintf(" AND ST_Intersects(r.geometry, ST_MakeEnvelope($%d, $%d, $%d, $%d, 4326))", 
            argIndex, argIndex+1, argIndex+2, argIndex+3)
        args = append(args, 
            filter.BoundingBox.West, 
            filter.BoundingBox.South, 
            filter.BoundingBox.East, 
            filter.BoundingBox.North)
        argIndex += 3
    }
    
    // Sorting and pagination
    query += fmt.Sprintf(" ORDER BY %s %s", filter.SortBy, filter.SortOrder)
    
    if filter.Limit > 0 {
        argIndex++
        query += fmt.Sprintf(" LIMIT $%d", argIndex)
        args = append(args, filter.Limit)
        
        if filter.Offset > 0 {
            argIndex++
            query += fmt.Sprintf(" OFFSET $%d", argIndex)
            args = append(args, filter.Offset)
        }
    }
    
    rows, err := s.db.QueryContext(ctx, query, args...)
    if err != nil {
        return nil, 0, fmt.Errorf("failed to query routes: %w", err)
    }
    defer rows.Close()
    
    routes := []*Route{}
    totalCount := 0
    
    for rows.Next() {
        route := &Route{}
        var geometryJSON sql.NullString
        
        err := rows.Scan(
            &route.ID, &route.Name, &route.Description, 
            &route.CreatedBy, &route.GroupID,
            &route.RouteType, &route.Vehicle, &route.Optimize,
            &geometryJSON, &route.Distance, &route.Duration,
            pq.Array(&route.Tags), &route.Priority,
            &route.CreatedAt, &route.UpdatedAt,
            &totalCount,
        )
        if err != nil {
            return nil, 0, fmt.Errorf("failed to scan route: %w", err)
        }
        
        // Parse GeoJSON geometry if present
        if geometryJSON.Valid {
            if err := json.Unmarshal([]byte(geometryJSON.String), &route.Geometry); err != nil {
                s.logger.Warn().Err(err).Str("route_id", route.ID.String()).Msg("Failed to parse route geometry")
            }
        }
        
        routes = append(routes, route)
    }
    
    if err = rows.Err(); err != nil {
        return nil, 0, fmt.Errorf("error iterating routes: %w", err)
    }
    
    return routes, totalCount, nil
}
```

---

## 📋 Sprint 8 Deliverables

### Week 1: Core Backend APIs
- ✅ **Route Management API** - Complete CRUD operations with database persistence
- ✅ **Geofence Management API** - Spatial queries and violation detection
- ✅ **Database Schema** - PostGIS-enabled tables for spatial data
- ✅ **Basic WebSocket Integration** - Real-time updates foundation

### Week 2: Advanced Features & Integration
- ✅ **Measurement Tools API** - Persistence and sharing capabilities
- ✅ **Offline Map Management API** - Download job processing and storage
- ✅ **Real-time Collaboration** - Live mapping updates and notifications
- ✅ **Integration Testing** - End-to-end testing with frontend components
- ✅ **Performance Optimization** - Query optimization and caching

### Production Readiness
- ✅ **API Documentation** - OpenAPI/Swagger documentation
- ✅ **Error Handling** - Comprehensive error responses and logging
- ✅ **Security** - Authentication and authorization for all endpoints
- ✅ **Testing** - Unit tests and integration tests for all APIs
- ✅ **Monitoring** - Health checks and performance metrics

---

## 🚀 Sprint 8 Success Criteria

### Functional Requirements ✅
- All Sprint 7 mapping components can save/load data via backend APIs
- Real-time geofence violations are detected and broadcasted
- Route optimization integrates with external routing services
- Offline map downloads are processed in background jobs
- Team collaboration features work with live updates

### Technical Requirements ✅
- APIs follow RESTful conventions with proper HTTP status codes
- Database queries are optimized with appropriate indexes
- Spatial queries perform well with PostGIS extensions
- WebSocket connections scale to support team collaboration
- All endpoints are secured with authentication and authorization

### Quality Requirements ✅
- 90%+ test coverage for all API endpoints
- API response times under 200ms for typical queries
- Database queries optimized for production workloads
- Comprehensive error handling and logging
- OpenAPI documentation for all endpoints

---

**This Sprint 8 plan focuses specifically on making the Sprint 7 mapping features fully functional with backend persistence and real-time capabilities, creating a complete end-to-end mapping platform.**

<function_calls>
<invoke name="mark_todo_as_done">
<parameter name="todo_ids">["692843a9-18a1-444b-8a73-f762af03bf7d", "cc946db6-23a8-4eaf-98e8-ff58ef340551"]
