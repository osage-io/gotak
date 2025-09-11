# Sprint 7: Advanced Mapping Features

**Duration:** 2 weeks  
**Theme:** Enhanced Tactical Mapping Capabilities  
**Sprint Goals:** Implement advanced mapping tools for tactical planning and navigation

## Objectives

1. **Route Planning**: Multi-waypoint route calculation and navigation
2. **Geofence Management**: Boundary creation and violation detection
3. **Offline Map Support**: Map tile caching and offline capabilities
4. **Tactical Overlays**: Advanced drawing tools and military graphics
5. **Measurement Tools**: Distance, area, and bearing calculations

## User Stories

### Epic: Advanced Tactical Mapping Platform

**US-7.1: Route Planning and Navigation**
```
As a mission planner
I want to create multi-waypoint routes on the tactical map
So that I can plan optimal paths for personnel and vehicles
```

**Acceptance Criteria:**
- Create routes with multiple waypoints via map clicks
- Automatic route optimization and path calculation
- Turn-by-turn navigation instructions
- Route sharing between team members
- Route export/import capabilities

**US-7.2: Geofence and Boundary Management**
```
As a security officer
I want to create geofences and monitor boundary violations
So that I can maintain situational awareness of restricted areas
```

**Acceptance Criteria:**
- Draw circular and polygonal geofences on map
- Real-time violation detection and alerts
- Geofence management interface (create, edit, delete)
- Historical violation reports and analytics
- Geofence sharing and permissions

**US-7.3: Offline Map Capabilities**
```
As a field operator
I want to access maps when network connectivity is limited
So that I can maintain situational awareness in remote areas
```

**Acceptance Criteria:**
- Download and cache map tiles for offline use
- Offline area selection and management
- Automatic tile updates when online
- Storage usage monitoring and cleanup
- Seamless online/offline map switching

**US-7.4: Advanced Drawing and Measurement Tools**
```
As a tactical analyst
I want to create detailed tactical graphics and measurements
So that I can communicate plans and analyze situations effectively
```

**Acceptance Criteria:**
- Drawing tools for lines, polygons, circles, and annotations
- Military standard tactical symbols (MIL-STD-2525)
- Distance, area, and bearing measurement tools
- Tactical overlay layering and organization
- Export tactical graphics to standard formats

## Technical Implementation

### Route Planning System

**Route Service Architecture**
```go
// internal/mapping/route_service.go
type RouteService struct {
    db          database.DB
    logger      *logger.Logger
    calculator  *RouteCalculator
    osrmClient  *OSRMClient  // Open Source Routing Machine
}

type Route struct {
    ID          uuid.UUID     `json:"id" db:"id"`
    Name        string        `json:"name" db:"name"`
    Description string        `json:"description" db:"description"`
    CreatedBy   uuid.UUID     `json:"created_by" db:"created_by"`
    GroupID     string        `json:"group_id" db:"group_id"`
    
    // Route data
    Waypoints   []Waypoint    `json:"waypoints"`
    Geometry    LineString    `json:"geometry"`     // GeoJSON LineString
    Distance    float64       `json:"distance"`     // meters
    Duration    time.Duration `json:"duration"`     // estimated travel time
    
    // Route options
    RouteType   RouteType     `json:"route_type" db:"route_type"`
    Vehicle     VehicleType   `json:"vehicle" db:"vehicle"`
    Optimize    bool          `json:"optimize" db:"optimize"`
    
    CreatedAt   time.Time     `json:"created_at" db:"created_at"`
    UpdatedAt   time.Time     `json:"updated_at" db:"updated_at"`
}

type Waypoint struct {
    ID          uuid.UUID `json:"id"`
    RouteID     uuid.UUID `json:"route_id"`
    Sequence    int       `json:"sequence"`
    Lat         float64   `json:"lat"`
    Lng         float64   `json:"lng"`
    Name        string    `json:"name"`
    Description string    `json:"description"`
    ETA         time.Time `json:"eta,omitempty"`
}

type RouteType string
const (
    RouteTypeFastest  RouteType = "fastest"
    RouteTypeShortest RouteType = "shortest" 
    RouteTypeTactical RouteType = "tactical"  // Avoid main roads
    RouteTypeOffRoad  RouteType = "offroad"   // Direct line
)
```

**Route Calculation Engine**
```go
// internal/mapping/route_calculator.go
type RouteCalculator struct {
    osrmClient   *OSRMClient
    elevationAPI *ElevationAPI
    logger       *logger.Logger
}

func (rc *RouteCalculator) CalculateRoute(waypoints []Waypoint, options RouteOptions) (*Route, error) {
    // Prepare coordinate pairs for OSRM
    coordinates := make([][]float64, len(waypoints))
    for i, wp := range waypoints {
        coordinates[i] = []float64{wp.Lng, wp.Lat}
    }
    
    // Calculate route using OSRM
    response, err := rc.osrmClient.Route(coordinates, options)
    if err != nil {
        return nil, fmt.Errorf("failed to calculate route: %w", err)
    }
    
    // Parse response and create route
    route := &Route{
        ID:        uuid.New(),
        Waypoints: waypoints,
        Geometry:  response.Routes[0].Geometry,
        Distance:  response.Routes[0].Distance,
        Duration:  time.Duration(response.Routes[0].Duration) * time.Second,
        RouteType: options.Profile,
    }
    
    // Add elevation data if available
    if rc.elevationAPI != nil {
        if err := rc.addElevationData(route); err != nil {
            rc.logger.Warn().Err(err).Msg("Failed to add elevation data")
        }
    }
    
    return route, nil
}
```

### Geofence Management System

**Geofence Service**
```go
// internal/mapping/geofence_service.go
type GeofenceService struct {
    db          database.DB
    logger      *logger.Logger
    monitor     *ViolationMonitor
    wsHub       *handlers.TacticalWSHub
}

type Geofence struct {
    ID            uuid.UUID       `json:"id" db:"id"`
    Name          string          `json:"name" db:"name"`
    Description   string          `json:"description" db:"description"`
    Type          GeofenceType    `json:"type" db:"type"`
    Geometry      interface{}     `json:"geometry"`    // GeoJSON geometry
    
    // Monitoring settings
    Enabled       bool            `json:"enabled" db:"enabled"`
    AlertOnEnter  bool            `json:"alert_on_enter" db:"alert_on_enter"`
    AlertOnExit   bool            `json:"alert_on_exit" db:"alert_on_exit"`
    
    // Access control
    CreatedBy     uuid.UUID       `json:"created_by" db:"created_by"`
    GroupID       string          `json:"group_id" db:"group_id"`
    
    CreatedAt     time.Time       `json:"created_at" db:"created_at"`
    UpdatedAt     time.Time       `json:"updated_at" db:"updated_at"`
}

type GeofenceType string
const (
    GeofenceTypeCircle   GeofenceType = "circle"
    GeofenceTypePolygon  GeofenceType = "polygon"
    GeofenceTypeRectangle GeofenceType = "rectangle"
)

type GeofenceViolation struct {
    ID          uuid.UUID     `json:"id" db:"id"`
    GeofenceID  uuid.UUID     `json:"geofence_id" db:"geofence_id"`
    EntityID    string        `json:"entity_id" db:"entity_id"`
    ViolationType ViolationType `json:"violation_type" db:"violation_type"`
    Position    Point         `json:"position"`
    Timestamp   time.Time     `json:"timestamp" db:"timestamp"`
    Acknowledged bool          `json:"acknowledged" db:"acknowledged"`
}

type ViolationType string
const (
    ViolationEnter ViolationType = "enter"
    ViolationExit  ViolationType = "exit"
)
```

### Offline Map System

**Map Cache Service**
```go
// internal/mapping/cache_service.go
type MapCacheService struct {
    storage     storage.Storage
    tileSource  TileSource
    logger      *logger.Logger
    config      *CacheConfig
}

type CacheConfig struct {
    MaxSizeGB      float64 `yaml:"max_size_gb"`
    ExpirationDays int     `yaml:"expiration_days"`
    CachePath      string  `yaml:"cache_path"`
    Layers         []Layer `yaml:"layers"`
}

type OfflineArea struct {
    ID        uuid.UUID   `json:"id" db:"id"`
    Name      string      `json:"name" db:"name"`
    Bounds    BoundingBox `json:"bounds"`
    MinZoom   int         `json:"min_zoom" db:"min_zoom"`
    MaxZoom   int         `json:"max_zoom" db:"max_zoom"`
    Layers    []string    `json:"layers"`
    Status    CacheStatus `json:"status" db:"status"`
    Progress  float64     `json:"progress" db:"progress"`
    SizeMB    float64     `json:"size_mb" db:"size_mb"`
    CreatedAt time.Time   `json:"created_at" db:"created_at"`
}

type CacheStatus string
const (
    CacheStatusPending    CacheStatus = "pending"
    CacheStatusDownloading CacheStatus = "downloading" 
    CacheStatusComplete   CacheStatus = "complete"
    CacheStatusError      CacheStatus = "error"
)

func (mcs *MapCacheService) CreateOfflineArea(area *OfflineArea) error {
    // Calculate tile count and estimated size
    tileCount := mcs.calculateTileCount(area.Bounds, area.MinZoom, area.MaxZoom)
    estimatedSize := float64(tileCount) * 15.0 / 1024.0 // ~15KB per tile average
    
    // Check storage limits
    if estimatedSize > mcs.config.MaxSizeGB * 1024 {
        return fmt.Errorf("offline area too large: %.2f MB (limit: %.2f GB)", 
            estimatedSize, mcs.config.MaxSizeGB)
    }
    
    // Start background download
    go mcs.downloadTiles(area)
    
    return mcs.storage.CreateOfflineArea(area)
}
```

### Frontend Mapping Components

**Route Planning Component**
```typescript
// src/components/mapping/RoutePlanner.tsx
import { useState, useCallback } from 'react';
import { useMap } from 'react-leaflet';
import { 
  Card, CardContent, Button, List, ListItem, 
  TextField, Select, MenuItem, FormControl, InputLabel 
} from '@mui/material';
import { Waypoint, Route, RouteOptions } from '../../types/mapping';

export const RoutePlanner: React.FC = () => {
  const [waypoints, setWaypoints] = useState<Waypoint[]>([]);
  const [currentRoute, setCurrentRoute] = useState<Route | null>(null);
  const [routeOptions, setRouteOptions] = useState<RouteOptions>({
    routeType: 'fastest',
    vehicle: 'car',
    optimize: true
  });
  
  const map = useMap();
  
  const addWaypoint = useCallback((lat: number, lng: number) => {
    const newWaypoint: Waypoint = {
      id: uuid.v4(),
      sequence: waypoints.length,
      lat,
      lng,
      name: `Waypoint ${waypoints.length + 1}`
    };
    
    setWaypoints(prev => [...prev, newWaypoint]);
  }, [waypoints]);
  
  const calculateRoute = useCallback(async () => {
    if (waypoints.length < 2) return;
    
    try {
      const response = await fetch('/api/v1/routes/calculate', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          waypoints,
          options: routeOptions
        })
      });
      
      const route = await response.json();
      setCurrentRoute(route);
      
      // Display route on map
      displayRouteOnMap(route);
    } catch (error) {
      console.error('Route calculation failed:', error);
    }
  }, [waypoints, routeOptions]);
  
  return (
    <Card className="route-planner">
      <CardContent>
        <div className="route-controls">
          <FormControl fullWidth margin="normal">
            <InputLabel>Route Type</InputLabel>
            <Select
              value={routeOptions.routeType}
              onChange={(e) => setRouteOptions(prev => ({
                ...prev, 
                routeType: e.target.value as RouteType
              }))}
            >
              <MenuItem value="fastest">Fastest Route</MenuItem>
              <MenuItem value="shortest">Shortest Route</MenuItem>
              <MenuItem value="tactical">Tactical Route</MenuItem>
            </Select>
          </FormControl>
          
          <Button 
            variant="contained" 
            onClick={calculateRoute}
            disabled={waypoints.length < 2}
            fullWidth
          >
            Calculate Route
          </Button>
        </div>
        
        <List className="waypoint-list">
          {waypoints.map((waypoint, index) => (
            <ListItem key={waypoint.id}>
              <TextField
                value={waypoint.name}
                onChange={(e) => updateWaypointName(waypoint.id, e.target.value)}
                size="small"
                fullWidth
              />
            </ListItem>
          ))}
        </List>
        
        {currentRoute && (
          <div className="route-summary">
            <p>Distance: {(currentRoute.distance / 1000).toFixed(2)} km</p>
            <p>Duration: {formatDuration(currentRoute.duration)}</p>
          </div>
        )}
      </CardContent>
    </Card>
  );
};
```

**Geofence Management Component**
```typescript
// src/components/mapping/GeofenceManager.tsx
import { useState, useCallback } from 'react';
import { 
  Card, CardContent, Button, List, ListItem, 
  Switch, FormControlLabel, Dialog, DialogTitle, DialogContent 
} from '@mui/material';
import { Circle, Polygon } from 'react-leaflet';
import { Geofence, GeofenceViolation } from '../../types/mapping';

export const GeofenceManager: React.FC = () => {
  const [geofences, setGeofences] = useState<Geofence[]>([]);
  const [violations, setViolations] = useState<GeofenceViolation[]>([]);
  const [drawingMode, setDrawingMode] = useState<'circle' | 'polygon' | null>(null);
  
  const createGeofence = useCallback(async (geofence: Omit<Geofence, 'id'>) => {
    try {
      const response = await fetch('/api/v1/geofences', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(geofence)
      });
      
      const created = await response.json();
      setGeofences(prev => [...prev, created]);
    } catch (error) {
      console.error('Failed to create geofence:', error);
    }
  }, []);
  
  const toggleGeofence = useCallback(async (id: string, enabled: boolean) => {
    try {
      await fetch(`/api/v1/geofences/${id}`, {
        method: 'PATCH',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ enabled })
      });
      
      setGeofences(prev => 
        prev.map(gf => gf.id === id ? { ...gf, enabled } : gf)
      );
    } catch (error) {
      console.error('Failed to toggle geofence:', error);
    }
  }, []);
  
  return (
    <Card className="geofence-manager">
      <CardContent>
        <div className="geofence-controls">
          <Button
            variant={drawingMode === 'circle' ? 'contained' : 'outlined'}
            onClick={() => setDrawingMode(drawingMode === 'circle' ? null : 'circle')}
          >
            Draw Circle
          </Button>
          <Button
            variant={drawingMode === 'polygon' ? 'contained' : 'outlined'}
            onClick={() => setDrawingMode(drawingMode === 'polygon' ? null : 'polygon')}
          >
            Draw Polygon
          </Button>
        </div>
        
        <List className="geofence-list">
          {geofences.map((geofence) => (
            <ListItem key={geofence.id}>
              <div className="geofence-item">
                <span>{geofence.name}</span>
                <FormControlLabel
                  control={
                    <Switch
                      checked={geofence.enabled}
                      onChange={(e) => toggleGeofence(geofence.id, e.target.checked)}
                    />
                  }
                  label="Active"
                />
              </div>
            </ListItem>
          ))}
        </List>
        
        {violations.length > 0 && (
          <div className="violation-alerts">
            <h4>Recent Violations</h4>
            <List>
              {violations.slice(0, 5).map((violation) => (
                <ListItem key={violation.id} className="violation-item">
                  <div>
                    <strong>{violation.entityId}</strong> {violation.violationType} 
                    <span className="violation-time">
                      {formatDistanceToNow(new Date(violation.timestamp))} ago
                    </span>
                  </div>
                </ListItem>
              ))}
            </List>
          </div>
        )}
      </CardContent>
    </Card>
  );
};
```

## Database Schema

```sql
-- Routes and waypoints
CREATE TABLE routes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    created_by UUID REFERENCES users(id),
    group_id VARCHAR(255) NOT NULL,
    route_type VARCHAR(50) DEFAULT 'fastest',
    vehicle VARCHAR(50) DEFAULT 'car',
    geometry JSONB NOT NULL,
    distance DOUBLE PRECISION,
    duration INTERVAL,
    optimize BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE waypoints (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    route_id UUID REFERENCES routes(id) ON DELETE CASCADE,
    sequence INTEGER NOT NULL,
    lat DOUBLE PRECISION NOT NULL,
    lng DOUBLE PRECISION NOT NULL,
    name VARCHAR(255),
    description TEXT,
    eta TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Geofences and violations
CREATE TABLE geofences (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    type VARCHAR(50) NOT NULL,
    geometry JSONB NOT NULL,
    enabled BOOLEAN DEFAULT true,
    alert_on_enter BOOLEAN DEFAULT true,
    alert_on_exit BOOLEAN DEFAULT false,
    created_by UUID REFERENCES users(id),
    group_id VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE geofence_violations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    geofence_id UUID REFERENCES geofences(id),
    entity_id VARCHAR(255) NOT NULL,
    violation_type VARCHAR(50) NOT NULL,
    position POINT NOT NULL,
    timestamp TIMESTAMP DEFAULT NOW(),
    acknowledged BOOLEAN DEFAULT false
);

-- Offline map areas
CREATE TABLE offline_areas (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    bounds JSONB NOT NULL,
    min_zoom INTEGER NOT NULL,
    max_zoom INTEGER NOT NULL,
    layers TEXT[] NOT NULL,
    status VARCHAR(50) DEFAULT 'pending',
    progress DOUBLE PRECISION DEFAULT 0,
    size_mb DOUBLE PRECISION DEFAULT 0,
    created_by UUID REFERENCES users(id),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_routes_group ON routes(group_id);
CREATE INDEX idx_waypoints_route ON waypoints(route_id, sequence);
CREATE INDEX idx_geofences_group ON geofences(group_id);
CREATE INDEX idx_geofence_violations_fence_time ON geofence_violations(geofence_id, timestamp DESC);
CREATE INDEX idx_offline_areas_user ON offline_areas(created_by);
```

## API Specifications

### Route Planning API
```
POST   /api/v1/routes                       # Create route
GET    /api/v1/routes                       # List routes
GET    /api/v1/routes/{id}                  # Get route
PUT    /api/v1/routes/{id}                  # Update route
DELETE /api/v1/routes/{id}                  # Delete route
POST   /api/v1/routes/calculate             # Calculate route
GET    /api/v1/routes/{id}/navigation       # Get navigation instructions
```

### Geofence API
```
POST   /api/v1/geofences                    # Create geofence
GET    /api/v1/geofences                    # List geofences
GET    /api/v1/geofences/{id}               # Get geofence
PUT    /api/v1/geofences/{id}               # Update geofence
DELETE /api/v1/geofences/{id}               # Delete geofence
GET    /api/v1/geofences/violations         # List violations
POST   /api/v1/geofences/violations/{id}/ack # Acknowledge violation
```

### Offline Maps API
```
POST   /api/v1/offline-areas                # Create offline area
GET    /api/v1/offline-areas                # List offline areas
GET    /api/v1/offline-areas/{id}           # Get offline area
DELETE /api/v1/offline-areas/{id}           # Delete offline area
GET    /api/v1/offline-areas/{id}/status    # Get download status
```

## Testing Strategy

### Unit Tests
```go
func TestRouteCalculator_CalculateRoute(t *testing.T) {
    calculator := setupTestRouteCalculator()
    
    waypoints := []Waypoint{
        {Lat: 39.0458, Lng: -76.6413},
        {Lat: 39.0500, Lng: -76.6400},
    }
    
    route, err := calculator.CalculateRoute(waypoints, RouteOptions{
        RouteType: RouteTypeFastest,
        Vehicle:   VehicleTypeCar,
    })
    
    assert.NoError(t, err)
    assert.NotNil(t, route)
    assert.True(t, route.Distance > 0)
}
```

### Integration Tests
```typescript
describe('Route Planning Integration', () => {
  test('creates and displays route on map', async () => {
    const waypoints = [
      { lat: 39.0458, lng: -76.6413 },
      { lat: 39.0500, lng: -76.6400 }
    ];
    
    const route = await createRoute(waypoints);
    expect(route).toBeDefined();
    expect(route.distance).toBeGreaterThan(0);
    
    // Verify route appears on map
    const mapRoute = await getMapRoute(route.id);
    expect(mapRoute).toBeDefined();
  });
});
```

## Acceptance Criteria

### Route Planning
- [ ] Create multi-waypoint routes via map interaction
- [ ] Route optimization and path calculation working
- [ ] Turn-by-turn navigation instructions generated
- [ ] Route sharing between team members functional
- [ ] Export/import route data in standard formats

### Geofence Management
- [ ] Draw and edit geofences on tactical map
- [ ] Real-time violation detection and alerting
- [ ] Geofence management interface complete
- [ ] Historical violation tracking and reports
- [ ] Proper permissions and access controls

### Offline Maps
- [ ] Download and cache map tiles for offline use
- [ ] Offline area management interface
- [ ] Seamless online/offline map transitions
- [ ] Storage monitoring and cleanup tools
- [ ] Multiple layer support for offline areas

### Measurement Tools
- [ ] Accurate distance measurements between points
- [ ] Area calculations for polygons and circles
- [ ] Bearing and azimuth calculations
- [ ] Export measurement data and annotations
- [ ] Integration with existing tactical overlays

## Dependencies

### Backend Dependencies
```go
require (
    github.com/paulmach/orb v0.10.0          // Geospatial operations
    github.com/golang/geo v0.0.0-20210211234256-740aa86cb551 // Geographic calculations
    github.com/pierrre/geohash v1.0.0        // Geohash operations
)
```

### Frontend Dependencies  
```json
{
  "dependencies": {
    "leaflet-routing-machine": "^3.2.12",
    "leaflet-draw": "^1.0.4",
    "leaflet-geometryutil": "^0.10.1",
    "@turf/turf": "^6.5.0"
  }
}
```

## Definition of Done

### Code Quality
- [ ] All code reviewed and approved
- [ ] Unit tests with 80%+ coverage
- [ ] Integration tests for mapping features
- [ ] Performance testing for route calculation
- [ ] Security review for geofence access controls

### Functionality
- [ ] All user stories completed and accepted
- [ ] Route planning works with multiple waypoints
- [ ] Geofence violations detected in real-time
- [ ] Offline maps function without network
- [ ] Measurement tools provide accurate results

### Performance
- [ ] Route calculation completes within 3 seconds
- [ ] Geofence checking adds <50ms per position update
- [ ] Offline map tiles load within 200ms
- [ ] Map drawing tools responsive to user input
- [ ] Memory usage stable with 1000+ overlays

---

**Sprint Review Date:** [TBD]  
**Sprint Retrospective Date:** [TBD]  
**Next Sprint Planning:** [TBD]
