# TacticalMap Component

A comprehensive React mapping component for tactical operations, built on Leaflet with advanced features for route planning, geofencing, measurement tools, and real-time entity tracking.

## Features

### Core Mapping
- **Interactive Map**: Leaflet-based mapping with customizable layers
- **Multiple Map Layers**: OpenStreetMap, Satellite imagery, Topographic maps
- **Coordinate Display**: Real-time coordinate tracking with multiple formats (DD, DMS, MGRS)
- **Zoom Controls**: Built-in zoom and navigation controls
- **Scale Display**: Configurable scale indicator

### Entity Tracking
- **Real-time Positioning**: WebSocket-based live entity updates
- **Entity Classification**: Support for friendly, hostile, unknown entity types
- **Custom Markers**: Tactical symbols with callsigns and status indicators  
- **Entity Trails**: Historical movement tracking (planned)
- **Stale Entity Detection**: Automatic detection and marking of outdated positions

### Advanced Mapping Features

#### Route Planning
- **Interactive Route Creation**: Click-to-add waypoint system
- **Route Visualization**: Styled polylines with distance calculations
- **Route Management**: Create, edit, and delete routes
- **Backend Integration**: Save and load routes via REST API
- **Route Statistics**: Total distance, waypoint count, estimated travel time

#### Geofencing
- **Multiple Geofence Types**:
  - Circular geofences (center + radius)
  - Polygon geofences (custom shapes)
  - Rectangle geofences (bounding boxes)
- **Visual Overlays**: Semi-transparent colored shapes with dashed borders
- **Real-time Violation Detection**: Automatic alerts when entities enter/exit geofences
- **Geofence Management**: Create, modify, and remove geofences

#### Measurement Tools  
- **Distance Measurement**: Multi-point distance calculation with cumulative totals
- **Area Measurement**: Polygon area calculation with metric/imperial units
- **Bearing Measurement**: Direction and distance between two points
- **Interactive Tooltips**: Real-time measurement display
- **Persistent Measurements**: Measurements remain visible until manually cleared

## Component API

### Props

#### Basic Configuration
```typescript
interface TacticalMapProps {
  // Map setup
  initialCenter?: LatLng;           // Starting map center (default: US center)
  initialZoom?: number;             // Initial zoom level (default: 4)
  height?: string;                  // Container height (default: '100%')
  width?: string;                   // Container width (default: '100%')
  className?: string;               // Additional CSS classes
  
  // Display options
  showCoordinates?: boolean;        // Show coordinate display (default: true)
  coordinateFormat?: 'dd'|'dms'|'mgrs'; // Coordinate format (default: 'dd')
  showScale?: boolean;              // Show scale indicator (default: true)
  readOnly?: boolean;               // Disable interactive features (default: false)
}
```

#### Entity Tracking
```typescript  
interface EntityTrackingProps {
  showEntities?: boolean;           // Enable entity display (default: true)
  showTrails?: boolean;             // Show entity movement trails (default: false)
  showFriendlyOnly?: boolean;       // Filter to friendly entities only
  showHostileOnly?: boolean;        // Filter to hostile entities only
  autoCenter?: boolean;             // Auto-center on entity updates
  onEntityClick?: (entity: EntityPosition) => void;  // Entity click handler
}
```

#### Advanced Features
```typescript
interface AdvancedMappingProps {
  // Route planning
  enableRouting?: boolean;          // Enable route creation tools
  showRoutes?: boolean;             // Display existing routes
  onRouteCreated?: (route: Route) => void; // Route creation callback
  
  // Geofencing  
  enableGeofencing?: boolean;       // Enable geofence creation tools
  showGeofences?: boolean;          // Display existing geofences
  onGeofenceCreated?: (geofence: Geofence) => void; // Geofence creation callback
  
  // Measurement
  enableMeasurement?: boolean;      // Enable measurement tools
  onMeasurement?: (type: string, value: number, points: LatLng[]) => void;
  
  // Event handlers
  onMapClick?: (latLng: LatLng) => void;        // Map click handler
  onMapMove?: (center: LatLng, zoom: number) => void; // Map move handler
}
```

### Usage Example

```tsx
import { TacticalMap } from './components/maps/TacticalMap';

function MyTacticalApp() {
  const handleRouteCreated = (route) => {
    console.log('New route created:', route);
    // Save to backend, update UI, etc.
  };
  
  const handleGeofenceCreated = (geofence) => {
    console.log('New geofence created:', geofence);
    // Process geofence, set up monitoring, etc.
  };
  
  const handleMeasurement = (type, value, points) => {
    console.log(`${type} measured:`, value, 'at points:', points);
    // Log measurement, save to report, etc.
  };

  return (
    <TacticalMap
      initialCenter={{ lat: 40.7128, lng: -74.0060 }} // New York City
      initialZoom={10}
      height="100vh"
      
      // Enable all advanced features
      enableRouting={true}
      enableGeofencing={true}  
      enableMeasurement={true}
      
      // Show existing data
      showRoutes={true}
      showGeofences={true}
      showEntities={true}
      
      // Event handlers
      onRouteCreated={handleRouteCreated}
      onGeofenceCreated={handleGeofenceCreated}
      onMeasurement={handleMeasurement}
    />
  );
}
```

## User Interface

### Layer Controls
- Located in top-right corner
- Switch between map layers (OSM, Satellite, Topographic)
- Smooth transitions between layer types

### Status Display  
- Top-left corner overlay
- Shows WebSocket connection status
- Entity count and loading states
- Error notifications

### Advanced Controls
- Bottom-left control panel with categorized tools
- **Routing Controls**: Route creation button with waypoint counter
- **Geofencing Controls**: Circle, Polygon, Rectangle geofence tools  
- **Measurement Controls**: Distance, Area, Bearing tools plus Clear button

### Interactive Dialogs
- **Drawing Status**: Center overlay during active drawing with instructions
- **Completion Dialogs**: Pop-up forms for naming and describing created routes/geofences
- **Input Validation**: Character limits and required field validation

### Coordinate Display
- Bottom-right corner information panel
- Current map center coordinates
- Mouse cursor position (when hovering)
- Current zoom level
- Statistics for routes and geofences

## Styling

The component uses a dark tactical theme with:
- Semi-transparent dark backgrounds with blur effects
- Color-coded elements (blue for routes, yellow for circular geofences, red for polygons)
- Monospace fonts for coordinate displays
- Responsive design for mobile devices
- High contrast support for accessibility
- Reduced motion support for users with vestibular disorders

## Backend Integration

### REST API Endpoints
The component integrates with the following backend services:

#### Routes
- `GET /api/routes` - List all routes
- `POST /api/routes` - Create new route  
- `PUT /api/routes/:id` - Update route
- `DELETE /api/routes/:id` - Delete route

#### Geofences
- `GET /api/geofences` - List all geofences
- `POST /api/geofences` - Create new geofence
- `PUT /api/geofences/:id` - Update geofence  
- `DELETE /api/geofences/:id` - Delete geofence

#### Entity Positions
- WebSocket connection to `/ws/positions` for real-time updates
- Automatic reconnection on connection loss
- Entity state management with staleness detection

### Data Models

#### Route
```typescript
interface Route {
  id: string;
  name?: string;
  description?: string;
  waypoints: Point[];
  distance?: number;
  created: Date;
  modified: Date;
}
```

#### Geofence  
```typescript
interface Geofence {
  id: string;
  name?: string;
  description?: string;
  type: 'circle' | 'polygon' | 'rectangle';
  geometry: GeofenceGeometry;
  active: boolean;
  created: Date;
  modified: Date;
}
```

#### Entity Position
```typescript  
interface EntityPosition {
  entityId: string;
  callsign?: string;
  lat: number;
  lng: number;
  altitude?: number;
  speed?: number;
  course?: number;
  type: 'friendly' | 'hostile' | 'unknown';
  lastUpdate: Date;
  isStale: boolean;
}
```

## Dependencies

### Required Packages
- `react` (^18.0.0)
- `leaflet` (^1.9.0) 
- `leaflet-draw` (^1.0.4)

### Peer Dependencies  
- `@types/leaflet` (for TypeScript)
- `@types/leaflet-draw` (for TypeScript)

### Included Assets
- Leaflet CSS and marker icons
- Custom tactical symbology
- Dark theme stylesheets

## Development

### File Structure
```
src/components/maps/
├── TacticalMap.tsx          # Main component
├── TacticalMap.css          # Component styles  
├── README.md               # This documentation
└── __tests__/
    └── TacticalMap.test.tsx # Unit tests

src/services/
├── mappingService.ts        # Backend API integration
└── mapUtils.ts             # Utility functions

src/utils/
├── mappingUtils.ts         # Geographic calculations  
└── coordinates.ts          # Coordinate formatting
```

### Testing
```bash
# Run unit tests
npm test src/components/maps/

# Run integration tests
npm test src/components/maps/ --testNamePattern="integration"

# Run with coverage
npm test -- --coverage src/components/maps/
```

### Building for Production
The component is optimized for production with:
- Code splitting for large mapping libraries
- Lazy loading of advanced features
- Efficient WebSocket connection management
- Memory leak prevention for long-running sessions

## Browser Compatibility

- **Modern Browsers**: Chrome 90+, Firefox 88+, Safari 14+, Edge 90+
- **Mobile**: iOS Safari 14+, Chrome Mobile 90+
- **Features**: WebGL support required for advanced rendering
- **Fallbacks**: Graceful degradation for unsupported features

## Performance Considerations

- **Entity Limits**: Optimized for up to 1,000 concurrent entities
- **Route Complexity**: Supports routes with up to 500 waypoints
- **Geofence Count**: Efficient rendering of up to 100 active geofences
- **Memory Management**: Automatic cleanup of stale data and unused resources
- **Network Optimization**: Efficient WebSocket message handling and reconnection logic

## Security Notes

- **Input Validation**: All user inputs are sanitized and validated
- **XSS Protection**: Safe HTML rendering for all dynamic content
- **CSRF Protection**: Secure API communication with proper headers
- **Rate Limiting**: Built-in throttling for rapid user interactions
