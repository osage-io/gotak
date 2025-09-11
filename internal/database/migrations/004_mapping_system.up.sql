-- Mapping System Migration
-- Adds comprehensive mapping functionality including routes, geofences, and offline map caching

-- Routes table for storing calculated routes with waypoints
CREATE TABLE routes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    created_by UUID REFERENCES auth_users(id) ON DELETE SET NULL,
    group_id VARCHAR(100) NOT NULL, -- TAK group/team identifier
    
    -- Route geometry and metrics
    geometry JSONB NOT NULL, -- GeoJSON LineString geometry
    distance DECIMAL(12,2) NOT NULL, -- Distance in meters
    duration BIGINT NOT NULL, -- Duration in nanoseconds
    
    -- Route configuration
    route_type VARCHAR(20) NOT NULL DEFAULT 'fastest', -- fastest, shortest, tactical, offroad
    vehicle VARCHAR(20) NOT NULL DEFAULT 'car', -- car, truck, bicycle, foot, motorcycle
    optimize BOOLEAN DEFAULT FALSE, -- Whether route was optimized
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    -- Constraints
    CONSTRAINT valid_route_type CHECK (route_type IN ('fastest', 'shortest', 'tactical', 'offroad')),
    CONSTRAINT valid_vehicle_type CHECK (vehicle IN ('car', 'truck', 'bicycle', 'foot', 'motorcycle'))
);

-- Waypoints table for storing individual points along routes
CREATE TABLE waypoints (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    route_id UUID NOT NULL REFERENCES routes(id) ON DELETE CASCADE,
    sequence INTEGER NOT NULL, -- Order of waypoint in route
    
    -- Geographic position
    lat DECIMAL(10, 8) NOT NULL,
    lng DECIMAL(11, 8) NOT NULL,
    
    -- Waypoint metadata
    name VARCHAR(255),
    description TEXT,
    eta TIMESTAMP WITH TIME ZONE, -- Estimated time of arrival
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    -- Constraints
    UNIQUE(route_id, sequence),
    CONSTRAINT valid_latitude CHECK (lat BETWEEN -90 AND 90),
    CONSTRAINT valid_longitude CHECK (lng BETWEEN -180 AND 180)
);

-- Geofences table for boundary monitoring
CREATE TABLE geofences (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    type VARCHAR(20) NOT NULL, -- circle, polygon, rectangle
    geometry JSONB NOT NULL, -- GeoJSON-like geometry definition
    
    -- Monitoring configuration
    enabled BOOLEAN DEFAULT TRUE,
    alert_on_enter BOOLEAN DEFAULT FALSE,
    alert_on_exit BOOLEAN DEFAULT FALSE,
    
    -- Access control
    created_by UUID REFERENCES auth_users(id) ON DELETE SET NULL,
    group_id VARCHAR(100) NOT NULL, -- TAK group/team identifier
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    -- Constraints
    CONSTRAINT valid_geofence_type CHECK (type IN ('circle', 'polygon', 'rectangle'))
);

-- Geofence violations table for tracking boundary breaches
CREATE TABLE geofence_violations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    geofence_id UUID NOT NULL REFERENCES geofences(id) ON DELETE CASCADE,
    entity_id VARCHAR(255) NOT NULL, -- Entity/callsign that violated boundary
    violation_type VARCHAR(10) NOT NULL, -- enter, exit
    
    -- Violation details
    position JSONB NOT NULL, -- Geographic position where violation occurred
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    
    -- Acknowledgment tracking
    acknowledged BOOLEAN DEFAULT FALSE,
    acknowledged_by UUID REFERENCES auth_users(id) ON DELETE SET NULL,
    acknowledged_at TIMESTAMP WITH TIME ZONE,
    
    -- Constraints
    CONSTRAINT valid_violation_type CHECK (violation_type IN ('enter', 'exit'))
);

-- Offline areas table for map tile caching
CREATE TABLE offline_areas (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    
    -- Geographic bounds
    bounds_north DECIMAL(10, 8) NOT NULL,
    bounds_south DECIMAL(10, 8) NOT NULL,
    bounds_east DECIMAL(11, 8) NOT NULL,
    bounds_west DECIMAL(11, 8) NOT NULL,
    
    -- Zoom levels
    min_zoom INTEGER NOT NULL DEFAULT 1,
    max_zoom INTEGER NOT NULL DEFAULT 18,
    
    -- Layer configuration
    layers JSONB NOT NULL, -- Array of layer IDs to cache
    
    -- Cache status and metrics
    status VARCHAR(20) NOT NULL DEFAULT 'pending', -- pending, downloading, complete, error
    progress DECIMAL(5,2) DEFAULT 0, -- Progress percentage (0-100)
    size_mb DECIMAL(10,2) DEFAULT 0, -- Cache size in megabytes
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    -- Constraints
    CONSTRAINT valid_cache_status CHECK (status IN ('pending', 'downloading', 'complete', 'error')),
    CONSTRAINT valid_progress CHECK (progress BETWEEN 0 AND 100),
    CONSTRAINT valid_zoom_levels CHECK (min_zoom <= max_zoom AND min_zoom >= 1 AND max_zoom <= 20),
    CONSTRAINT valid_bounds CHECK (bounds_south <= bounds_north AND bounds_west <= bounds_east)
);

-- Tactical overlays table for map annotations and graphics
CREATE TABLE tactical_overlays (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    type VARCHAR(20) NOT NULL, -- symbol, route, area, boundary, threat_circle, range_ring, line, polygon, circle, marker
    
    -- Geometry and style
    geometry JSONB NOT NULL, -- GeoJSON-like geometry
    style JSONB NOT NULL, -- Styling information (color, weight, opacity, etc.)
    
    -- Tactical metadata
    classification VARCHAR(20) DEFAULT 'UNCLASSIFIED',
    priority VARCHAR(10) NOT NULL DEFAULT 'MEDIUM', -- LOW, MEDIUM, HIGH, CRITICAL
    effective_time TIMESTAMP WITH TIME ZONE,
    expiration_time TIMESTAMP WITH TIME ZONE,
    source VARCHAR(255),
    tags TEXT[],
    attributes JSONB DEFAULT '{}'::jsonb,
    
    -- Visibility and editing
    visible BOOLEAN DEFAULT TRUE,
    editable BOOLEAN DEFAULT TRUE,
    
    -- Access control
    created_by UUID REFERENCES auth_users(id) ON DELETE SET NULL,
    group_id VARCHAR(100) NOT NULL, -- TAK group/team identifier
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    -- Constraints
    CONSTRAINT valid_overlay_type CHECK (type IN ('symbol', 'route', 'area', 'boundary', 'threat_circle', 'range_ring', 'line', 'polygon', 'circle', 'marker')),
    CONSTRAINT valid_classification CHECK (classification IN ('UNCLASSIFIED', 'RESTRICTED', 'CONFIDENTIAL', 'SECRET', 'TOP_SECRET')),
    CONSTRAINT valid_priority CHECK (priority IN ('LOW', 'MEDIUM', 'HIGH', 'CRITICAL'))
);

-- Create indexes for performance optimization

-- Routes indexes
CREATE INDEX idx_routes_group_id ON routes(group_id);
CREATE INDEX idx_routes_created_by ON routes(created_by);
CREATE INDEX idx_routes_created_at ON routes(created_at DESC);
CREATE INDEX idx_routes_route_type ON routes(route_type);

-- Waypoints indexes  
CREATE INDEX idx_waypoints_route_id ON waypoints(route_id);
CREATE INDEX idx_waypoints_sequence ON waypoints(route_id, sequence);
CREATE INDEX idx_waypoints_location ON waypoints(lat, lng);

-- Geofences indexes
CREATE INDEX idx_geofences_group_id ON geofences(group_id);
CREATE INDEX idx_geofences_enabled ON geofences(enabled) WHERE enabled = TRUE;
CREATE INDEX idx_geofences_type ON geofences(type);
CREATE INDEX idx_geofences_created_at ON geofences(created_at DESC);

-- Geofence violations indexes
CREATE INDEX idx_violations_geofence_id ON geofence_violations(geofence_id);
CREATE INDEX idx_violations_entity_id ON geofence_violations(entity_id);
CREATE INDEX idx_violations_timestamp ON geofence_violations(timestamp DESC);
CREATE INDEX idx_violations_acknowledged ON geofence_violations(acknowledged) WHERE acknowledged = FALSE;
CREATE INDEX idx_violations_type ON geofence_violations(violation_type);

-- Offline areas indexes
CREATE INDEX idx_offline_areas_status ON offline_areas(status);
CREATE INDEX idx_offline_areas_created_at ON offline_areas(created_at DESC);
CREATE INDEX idx_offline_areas_bounds ON offline_areas(bounds_north, bounds_south, bounds_east, bounds_west);

-- Tactical overlays indexes
CREATE INDEX idx_tactical_overlays_group_id ON tactical_overlays(group_id);
CREATE INDEX idx_tactical_overlays_type ON tactical_overlays(type);
CREATE INDEX idx_tactical_overlays_visible ON tactical_overlays(visible) WHERE visible = TRUE;
CREATE INDEX idx_tactical_overlays_classification ON tactical_overlays(classification);
CREATE INDEX idx_tactical_overlays_priority ON tactical_overlays(priority);
CREATE INDEX idx_tactical_overlays_created_at ON tactical_overlays(created_at DESC);
CREATE INDEX idx_tactical_overlays_tags ON tactical_overlays USING GIN(tags);
CREATE INDEX idx_tactical_overlays_expiration ON tactical_overlays(expiration_time) WHERE expiration_time IS NOT NULL;
