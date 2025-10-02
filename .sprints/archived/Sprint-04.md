# Sprint 4: Interactive Maps & Positioning

**Duration:** 2 weeks  
**Sprint Goals:** Implement interactive maps with real-time position tracking and tactical overlays

## Objectives

1. **Interactive Map Foundation**: Integrate Leaflet.js for interactive mapping
2. **Real-time Position Display**: Show live entity positions on map
3. **Tactical Symbology**: Military-standard iconography and overlays
4. **Map Controls**: Pan, zoom, layer management, tactical controls
5. **Position History**: Track and display movement patterns

## User Stories

### Epic: Interactive Mapping Platform

**US-4.1: Base Map Integration**
```
As an operator
I want to see an interactive map interface
So that I can visualize the tactical situation geographically
```

**Acceptance Criteria:**
- Leaflet.js integrated with multiple base layers (OpenStreetMap, satellite, tactical)
- Map supports pan, zoom, and standard controls
- Responsive design works on desktop and mobile
- Map state persists between sessions
- Loading states and error handling implemented

**US-4.2: Real-time Entity Positioning**
```
As an operator
I want to see live positions of all entities on the map
So that I can maintain real-time situational awareness
```

**Acceptance Criteria:**
- Entities appear as military-standard icons on map
- Positions update in real-time via WebSocket
- Different entity types have distinct symbology
- Click on entity shows detailed information popup
- Entity trails show recent movement history

**US-4.3: Tactical Overlays and Controls**
```
As a mission commander
I want to add tactical overlays to the map
So that I can annotate the battlefield and share tactical information
```

**Acceptance Criteria:**
- Draw tools for points, lines, polygons
- Text annotations and labels
- Tactical symbols library (MIL-STD-2525)
- Layer management (show/hide different overlays)
- Overlay persistence and sharing between users

**US-4.4: Position History and Tracking**
```
As an analyst
I want to view historical position data
So that I can analyze movement patterns and reconstruct events
```

**Acceptance Criteria:**
- Timeline slider for historical position replay
- Trail visualization with time-based coloring
- Speed and direction indicators
- Export position data for analysis
- Performance optimized for large datasets

## Technical Implementation

### Frontend Components

**MapContainer.tsx**
```typescript
import { MapContainer, TileLayer, Marker, Popup } from 'react-leaflet';
import { useWebSocket } from '../hooks/useWebSocket';
import { usePositions } from '../hooks/usePositions';
import { TacticalOverlay } from './TacticalOverlay';
import { EntityMarker } from './EntityMarker';

export const TacticalMap: React.FC = () => {
  const { positions } = usePositions();
  const { connected } = useWebSocket();
  
  return (
    <MapContainer
      center={[39.0458, -76.6413]} // Fort Meade
      zoom={13}
      style={{ height: '100%', width: '100%' }}
    >
      <TileLayer
        url="https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png"
        attribution="&copy; OpenStreetMap contributors"
      />
      {positions.map(position => (
        <EntityMarker
          key={position.uid}
          position={position}
          onClick={handleEntityClick}
        />
      ))}
      <TacticalOverlay />
    </MapContainer>
  );
};
```

**EntityMarker.tsx**
```typescript
import { Marker, Popup } from 'react-leaflet';
import { divIcon } from 'leaflet';
import { Entity } from '../types/entity';
import { getMilSymbol } from '../utils/milSymbols';

interface EntityMarkerProps {
  position: Entity;
  onClick: (entity: Entity) => void;
}

export const EntityMarker: React.FC<EntityMarkerProps> = ({ position, onClick }) => {
  const icon = divIcon({
    html: getMilSymbol(position.type, position.affiliation),
    iconSize: [32, 32],
    className: 'tactical-marker'
  });

  return (
    <Marker
      position={[position.lat, position.lon]}
      icon={icon}
      eventHandlers={{
        click: () => onClick(position),
      }}
    >
      <Popup>
        <div className="entity-popup">
          <h3>{position.callsign}</h3>
          <p>Type: {position.type}</p>
          <p>Last Update: {position.time}</p>
          <p>Speed: {position.speed} m/s</p>
        </div>
      </Popup>
    </Marker>
  );
};
```

### Position Management Hook

**usePositions.ts**
```typescript
import { useState, useEffect } from 'react';
import { useWebSocket } from './useWebSocket';
import { Entity } from '../types/entity';

export const usePositions = () => {
  const [positions, setPositions] = useState<Map<string, Entity>>(new Map());
  const { lastMessage } = useWebSocket();

  useEffect(() => {
    if (lastMessage && lastMessage.type === 'position_update') {
      setPositions(prev => {
        const updated = new Map(prev);
        updated.set(lastMessage.data.uid, lastMessage.data);
        return updated;
      });
    }
  }, [lastMessage]);

  const getPositionsArray = () => Array.from(positions.values());
  
  const getPositionHistory = async (uid: string, timeRange: string) => {
    const response = await fetch(`/api/v1/entities/${uid}/history?range=${timeRange}`);
    return response.json();
  };

  return {
    positions: getPositionsArray(),
    getPositionHistory,
    entityCount: positions.size
  };
};
```

### Backend Position API Extensions

**Position History Endpoint**
```go
// GET /api/v1/entities/{uid}/history
func (s *Server) handleGetPositionHistory(w http.ResponseWriter, r *http.Request) {
    uid := mux.Vars(r)["uid"]
    timeRange := r.URL.Query().Get("range") // "1h", "24h", "7d"
    
    positions, err := s.db.GetPositionHistory(r.Context(), uid, timeRange)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(positions)
}
```

**WebSocket Position Broadcasting**
```go
func (s *Server) broadcastPositionUpdate(position *Entity) {
    message := WSMessage{
        Type: "position_update",
        Data: position,
        Timestamp: time.Now(),
    }
    
    s.wsManager.BroadcastToGroup(position.GroupID, message)
}
```

### Database Schema Extensions

```sql
-- Enhanced entity positions table
CREATE TABLE entity_positions (
    id SERIAL PRIMARY KEY,
    uid VARCHAR(255) NOT NULL,
    callsign VARCHAR(255),
    type VARCHAR(100),
    affiliation VARCHAR(50),
    lat DOUBLE PRECISION NOT NULL,
    lon DOUBLE PRECISION NOT NULL,
    altitude DOUBLE PRECISION,
    speed DOUBLE PRECISION,
    course DOUBLE PRECISION,
    accuracy DOUBLE PRECISION,
    timestamp TIMESTAMP NOT NULL,
    group_id VARCHAR(255),
    classification VARCHAR(50) DEFAULT 'UNCLASSIFIED',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Index for efficient position queries
CREATE INDEX idx_entity_positions_uid_time ON entity_positions(uid, timestamp DESC);
CREATE INDEX idx_entity_positions_time ON entity_positions(timestamp DESC);
CREATE INDEX idx_entity_positions_group ON entity_positions(group_id);

-- Tactical overlays table
CREATE TABLE tactical_overlays (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    type VARCHAR(100) NOT NULL, -- 'point', 'line', 'polygon', 'text'
    geometry JSONB NOT NULL,     -- GeoJSON geometry
    properties JSONB,            -- Style and metadata
    group_id VARCHAR(255),
    created_by VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

## API Specifications

### REST Endpoints

**Position History API**
```
GET /api/v1/entities/{uid}/history
Query Parameters:
  - range: "1h" | "24h" | "7d" | "30d" (default: "24h")
  - limit: number (default: 1000)
  
Response:
{
  "uid": "entity-123",
  "positions": [
    {
      "lat": 39.0458,
      "lon": -76.6413,
      "altitude": 100,
      "timestamp": "2024-01-15T10:30:00Z",
      "speed": 5.2,
      "course": 45
    }
  ],
  "total_count": 1250
}
```

**Tactical Overlays API**
```
POST /api/v1/overlays
{
  "name": "Checkpoint Alpha",
  "type": "point",
  "geometry": {
    "type": "Point",
    "coordinates": [-76.6413, 39.0458]
  },
  "properties": {
    "icon": "checkpoint",
    "color": "#FF0000",
    "description": "Primary checkpoint"
  }
}
```

### WebSocket Messages

**Position Update**
```json
{
  "type": "position_update",
  "data": {
    "uid": "entity-123",
    "callsign": "Alpha-1",
    "type": "friendly-infantry",
    "lat": 39.0458,
    "lon": -76.6413,
    "timestamp": "2024-01-15T10:30:00Z"
  },
  "timestamp": "2024-01-15T10:30:00Z"
}
```

**Overlay Update**
```json
{
  "type": "overlay_update",
  "data": {
    "id": "overlay-456",
    "action": "create",
    "overlay": {
      "name": "Route Blue",
      "type": "line",
      "geometry": {...}
    }
  }
}
```

## Testing Strategy

### Unit Tests
```typescript
// Frontend position tracking tests
describe('usePositions', () => {
  test('updates positions from WebSocket messages', () => {
    const { result } = renderHook(() => usePositions());
    
    act(() => {
      mockWebSocketMessage({
        type: 'position_update',
        data: mockEntity
      });
    });
    
    expect(result.current.positions).toHaveLength(1);
  });
});
```

```go
// Backend position history tests
func TestGetPositionHistory(t *testing.T) {
    db := setupTestDB()
    
    // Insert test positions
    positions := generateTestPositions("entity-123", 24)
    for _, pos := range positions {
        db.InsertPosition(pos)
    }
    
    // Test 1 hour range
    result, err := db.GetPositionHistory("entity-123", "1h")
    assert.NoError(t, err)
    assert.True(t, len(result) > 0)
}
```

### Integration Tests
```go
func TestWebSocketPositionBroadcast(t *testing.T) {
    server := setupTestServer()
    client := connectWebSocketClient()
    
    // Send position update via REST API
    position := &Entity{
        UID: "test-entity",
        Lat: 39.0458,
        Lon: -76.6413,
    }
    
    resp, err := http.Post("/api/v1/entities/test-entity/position", 
        "application/json", positionJSON)
    assert.NoError(t, err)
    assert.Equal(t, http.StatusOK, resp.StatusCode)
    
    // Verify WebSocket broadcast
    message := <-client.Messages
    assert.Equal(t, "position_update", message.Type)
}
```

## Performance Considerations

### Frontend Optimization
```typescript
// Efficient position rendering with clustering
import MarkerClusterGroup from 'react-leaflet-markercluster';

export const PositionLayer: React.FC = () => {
  const positions = usePositions();
  
  return (
    <MarkerClusterGroup>
      {positions.map(position => (
        <EntityMarker key={position.uid} position={position} />
      ))}
    </MarkerClusterGroup>
  );
};
```

### Backend Optimization
```go
// Efficient position queries with spatial indexing
func (db *DB) GetNearbyPositions(lat, lon, radiusKm float64) ([]*Entity, error) {
    query := `
        SELECT uid, callsign, lat, lon, timestamp
        FROM entity_positions 
        WHERE ST_DWithin(
            ST_Point(lon, lat)::geography,
            ST_Point($1, $2)::geography,
            $3 * 1000
        )
        AND timestamp > NOW() - INTERVAL '1 hour'
        ORDER BY timestamp DESC
    `
    
    rows, err := db.Query(query, lon, lat, radiusKm)
    // ... process results
}
```

## Acceptance Criteria

### Core Map Functionality
- [ ] Interactive map loads with multiple base layers
- [ ] Map is responsive and works on mobile devices
- [ ] Pan, zoom, and standard controls function properly
- [ ] Map state persists between browser sessions

### Real-time Position Display
- [ ] Entity positions appear on map in real-time
- [ ] Military symbology displays correctly for different entity types
- [ ] Entity information popups show complete details
- [ ] Position updates occur within 1 second of receiving data

### Tactical Features
- [ ] Drawing tools create persistent overlays
- [ ] Overlay management (show/hide layers) works correctly
- [ ] Tactical symbols library provides standard military icons
- [ ] All overlays sync between connected users

### Performance Requirements
- [ ] Map renders smoothly with 100+ entities
- [ ] Position updates don't cause UI lag or stuttering
- [ ] Historical position queries complete within 2 seconds
- [ ] Memory usage remains stable during extended sessions

### Historical Analysis
- [ ] Position history displays movement trails
- [ ] Timeline controls allow replay of past movements
- [ ] Export functionality works for position data
- [ ] Historical queries handle large datasets efficiently

## Dependencies

### Required Packages
```json
{
  "dependencies": {
    "leaflet": "^1.9.4",
    "react-leaflet": "^4.2.1",
    "react-leaflet-markercluster": "^3.0.0",
    "leaflet.markercluster": "^1.5.3",
    "@types/leaflet": "^1.9.8"
  }
}
```

### Go Dependencies
```go
// go.mod additions
require (
    github.com/paulmach/orb v0.10.0
    github.com/twpayne/go-geom v1.5.3
)
```

### External Services (Optional)
- **Map Tiles**: OpenStreetMap, Mapbox, or military-specific tile servers
- **Geocoding**: Nominatim or commercial geocoding services
- **Spatial Database**: PostGIS extension for PostgreSQL (production)

## Definition of Done

### Code Quality
- [ ] All code reviewed and approved
- [ ] Unit tests written with 80%+ coverage
- [ ] Integration tests pass
- [ ] No security vulnerabilities in dependencies
- [ ] Performance benchmarks meet requirements

### Functionality
- [ ] All user stories completed and accepted
- [ ] Manual testing completed on desktop and mobile
- [ ] Error handling implemented for all failure scenarios
- [ ] Accessibility requirements met (WCAG 2.1)

### Documentation
- [ ] API endpoints documented in OpenAPI spec
- [ ] Component documentation updated
- [ ] User guide sections completed
- [ ] Deployment notes updated

---

**Sprint Review Date:** [TBD]  
**Sprint Retrospective Date:** [TBD]  
**Next Sprint Planning:** [TBD]
