-- GoTAK Database Schema Migration 001
-- Initial schema creation for production deployment
-- Version: 1.0.0
-- Created: 2024-01-01

-- Enable required extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "postgis";
CREATE EXTENSION IF NOT EXISTS "pg_trgm";  -- For text search
CREATE EXTENSION IF NOT EXISTS "btree_gist";  -- For advanced indexing

-- Migration tracking handled by migration script

-- ===========================================================================
-- USER MANAGEMENT TABLES
-- ===========================================================================

-- Users table
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    username VARCHAR(255) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    first_name VARCHAR(255),
    last_name VARCHAR(255),
    role VARCHAR(50) DEFAULT 'user',
    group_id VARCHAR(255),
    is_active BOOLEAN DEFAULT true,
    is_verified BOOLEAN DEFAULT false,
    last_login_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- User sessions
CREATE TABLE IF NOT EXISTS user_sessions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash VARCHAR(255) NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    last_used_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    ip_address INET,
    user_agent TEXT
);

-- ===========================================================================
-- MAPPING TABLES
-- ===========================================================================

-- Routes table
CREATE TABLE IF NOT EXISTS routes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    created_by UUID NOT NULL REFERENCES users(id),
    group_id VARCHAR(255),
    geometry GEOMETRY(LINESTRING, 4326),
    distance DOUBLE PRECISION,
    duration BIGINT, -- Duration in nanoseconds
    route_type VARCHAR(50) DEFAULT 'fastest',
    vehicle VARCHAR(50) DEFAULT 'car',
    optimize BOOLEAN DEFAULT false,
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Waypoints table
CREATE TABLE IF NOT EXISTS waypoints (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    route_id UUID NOT NULL REFERENCES routes(id) ON DELETE CASCADE,
    sequence INTEGER NOT NULL,
    lat DOUBLE PRECISION NOT NULL,
    lng DOUBLE PRECISION NOT NULL,
    name VARCHAR(255),
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(route_id, sequence)
);

-- Geofences table
CREATE TABLE IF NOT EXISTS geofences (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    type VARCHAR(50) NOT NULL, -- circle, rectangle, polygon
    geometry JSONB NOT NULL,
    enabled BOOLEAN DEFAULT true,
    alert_on_enter BOOLEAN DEFAULT true,
    alert_on_exit BOOLEAN DEFAULT false,
    created_by UUID NOT NULL REFERENCES users(id),
    group_id VARCHAR(255),
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Geofence violations/alerts
CREATE TABLE IF NOT EXISTS geofence_violations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    geofence_id UUID NOT NULL REFERENCES geofences(id),
    entity_id VARCHAR(255) NOT NULL, -- Entity that triggered the violation
    violation_type VARCHAR(50) NOT NULL, -- enter, exit
    position GEOMETRY(POINT, 4326) NOT NULL,
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    acknowledged BOOLEAN DEFAULT false,
    acknowledged_by UUID REFERENCES users(id),
    acknowledged_at TIMESTAMP WITH TIME ZONE
);

-- Offline map areas
CREATE TABLE IF NOT EXISTS offline_areas (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    bounds JSONB NOT NULL, -- {north, south, east, west}
    min_zoom INTEGER DEFAULT 1,
    max_zoom INTEGER DEFAULT 18,
    layers TEXT[] DEFAULT ARRAY['streets'],
    status VARCHAR(50) DEFAULT 'pending', -- pending, downloading, complete, error
    progress DOUBLE PRECISION DEFAULT 0.0,
    size_mb DOUBLE PRECISION DEFAULT 0.0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Tactical overlays
CREATE TABLE IF NOT EXISTS tactical_overlays (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    type VARCHAR(50) NOT NULL, -- marker, line, area, text
    geometry JSONB NOT NULL,
    style JSONB,
    metadata JSONB,
    created_by UUID NOT NULL REFERENCES users(id),
    group_id VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- ===========================================================================
-- OPERATIONAL TABLES
-- ===========================================================================

-- System logs
CREATE TABLE IF NOT EXISTS system_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    level VARCHAR(20) NOT NULL,
    message TEXT NOT NULL,
    context JSONB,
    user_id UUID REFERENCES users(id),
    ip_address INET,
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Audit trail
CREATE TABLE IF NOT EXISTS audit_trail (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    table_name VARCHAR(255) NOT NULL,
    record_id UUID NOT NULL,
    action VARCHAR(50) NOT NULL, -- INSERT, UPDATE, DELETE
    old_data JSONB,
    new_data JSONB,
    user_id UUID REFERENCES users(id),
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Configuration settings
CREATE TABLE IF NOT EXISTS config_settings (
    key VARCHAR(255) PRIMARY KEY,
    value TEXT,
    value_type VARCHAR(50) DEFAULT 'string',
    description TEXT,
    is_public BOOLEAN DEFAULT false,
    updated_by UUID REFERENCES users(id),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- ===========================================================================
-- INDEXES FOR PERFORMANCE
-- ===========================================================================

-- User indexes
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_group_id ON users(group_id);
CREATE INDEX IF NOT EXISTS idx_users_role ON users(role);
CREATE INDEX IF NOT EXISTS idx_users_active ON users(is_active) WHERE is_active = true;

-- Session indexes
CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON user_sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_sessions_token ON user_sessions(token_hash);
CREATE INDEX IF NOT EXISTS idx_sessions_expires ON user_sessions(expires_at);

-- Route indexes
CREATE INDEX IF NOT EXISTS idx_routes_created_by ON routes(created_by);
CREATE INDEX IF NOT EXISTS idx_routes_group_id ON routes(group_id);
CREATE INDEX IF NOT EXISTS idx_routes_created_at ON routes(created_at);
CREATE INDEX IF NOT EXISTS idx_routes_geometry ON routes USING GIST(geometry);

-- Waypoint indexes
CREATE INDEX IF NOT EXISTS idx_waypoints_route_id ON waypoints(route_id);
CREATE INDEX IF NOT EXISTS idx_waypoints_sequence ON waypoints(route_id, sequence);

-- Geofence indexes
CREATE INDEX IF NOT EXISTS idx_geofences_created_by ON geofences(created_by);
CREATE INDEX IF NOT EXISTS idx_geofences_group_id ON geofences(group_id);
CREATE INDEX IF NOT EXISTS idx_geofences_enabled ON geofences(enabled) WHERE enabled = true;
CREATE INDEX IF NOT EXISTS idx_geofences_type ON geofences(type);

-- Geofence violation indexes
CREATE INDEX IF NOT EXISTS idx_violations_geofence_id ON geofence_violations(geofence_id);
CREATE INDEX IF NOT EXISTS idx_violations_entity_id ON geofence_violations(entity_id);
CREATE INDEX IF NOT EXISTS idx_violations_timestamp ON geofence_violations(timestamp);
CREATE INDEX IF NOT EXISTS idx_violations_position ON geofence_violations USING GIST(position);

-- Offline area indexes
CREATE INDEX IF NOT EXISTS idx_offline_areas_status ON offline_areas(status);
CREATE INDEX IF NOT EXISTS idx_offline_areas_created_at ON offline_areas(created_at);

-- Tactical overlay indexes
CREATE INDEX IF NOT EXISTS idx_overlays_created_by ON tactical_overlays(created_by);
CREATE INDEX IF NOT EXISTS idx_overlays_group_id ON tactical_overlays(group_id);
CREATE INDEX IF NOT EXISTS idx_overlays_type ON tactical_overlays(type);

-- System log indexes
CREATE INDEX IF NOT EXISTS idx_logs_level ON system_logs(level);
CREATE INDEX IF NOT EXISTS idx_logs_timestamp ON system_logs(timestamp);
CREATE INDEX IF NOT EXISTS idx_logs_user_id ON system_logs(user_id);

-- Audit trail indexes
CREATE INDEX IF NOT EXISTS idx_audit_table_name ON audit_trail(table_name);
CREATE INDEX IF NOT EXISTS idx_audit_record_id ON audit_trail(record_id);
CREATE INDEX IF NOT EXISTS idx_audit_timestamp ON audit_trail(timestamp);
CREATE INDEX IF NOT EXISTS idx_audit_user_id ON audit_trail(user_id);

-- ===========================================================================
-- FUNCTIONS AND TRIGGERS
-- ===========================================================================

-- Function to update the updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Apply updated_at triggers to tables (drop if exists first to make idempotent)
DROP TRIGGER IF EXISTS update_users_updated_at ON users;
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_routes_updated_at ON routes;
CREATE TRIGGER update_routes_updated_at BEFORE UPDATE ON routes
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_geofences_updated_at ON geofences;
CREATE TRIGGER update_geofences_updated_at BEFORE UPDATE ON geofences
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_offline_areas_updated_at ON offline_areas;
CREATE TRIGGER update_offline_areas_updated_at BEFORE UPDATE ON offline_areas
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_tactical_overlays_updated_at ON tactical_overlays;
CREATE TRIGGER update_tactical_overlays_updated_at BEFORE UPDATE ON tactical_overlays
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Function for audit trail
CREATE OR REPLACE FUNCTION audit_trigger_function()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'DELETE' THEN
        INSERT INTO audit_trail (table_name, record_id, action, old_data, user_id)
        VALUES (TG_TABLE_NAME, OLD.id, TG_OP, row_to_json(OLD), 
                COALESCE(current_setting('app.current_user_id', true)::UUID, NULL));
        RETURN OLD;
    ELSIF TG_OP = 'UPDATE' THEN
        INSERT INTO audit_trail (table_name, record_id, action, old_data, new_data, user_id)
        VALUES (TG_TABLE_NAME, NEW.id, TG_OP, row_to_json(OLD), row_to_json(NEW),
                COALESCE(current_setting('app.current_user_id', true)::UUID, NULL));
        RETURN NEW;
    ELSIF TG_OP = 'INSERT' THEN
        INSERT INTO audit_trail (table_name, record_id, action, new_data, user_id)
        VALUES (TG_TABLE_NAME, NEW.id, TG_OP, row_to_json(NEW),
                COALESCE(current_setting('app.current_user_id', true)::UUID, NULL));
        RETURN NEW;
    END IF;
    RETURN NULL;
END;
$$ language 'plpgsql';

-- Apply audit triggers to important tables (drop if exists first to make idempotent)
DROP TRIGGER IF EXISTS audit_users_trigger ON users;
CREATE TRIGGER audit_users_trigger
    AFTER INSERT OR UPDATE OR DELETE ON users
    FOR EACH ROW EXECUTE FUNCTION audit_trigger_function();

DROP TRIGGER IF EXISTS audit_routes_trigger ON routes;
CREATE TRIGGER audit_routes_trigger
    AFTER INSERT OR UPDATE OR DELETE ON routes
    FOR EACH ROW EXECUTE FUNCTION audit_trigger_function();

DROP TRIGGER IF EXISTS audit_geofences_trigger ON geofences;
CREATE TRIGGER audit_geofences_trigger
    AFTER INSERT OR UPDATE OR DELETE ON geofences
    FOR EACH ROW EXECUTE FUNCTION audit_trigger_function();

-- ===========================================================================
-- DEFAULT DATA
-- ===========================================================================

-- Insert default configuration settings
INSERT INTO config_settings (key, value, value_type, description, is_public) VALUES
('app.version', '1.0.0', 'string', 'Application version', true),
('app.name', 'GoTAK Server', 'string', 'Application name', true),
('app.max_upload_size', '104857600', 'integer', 'Maximum file upload size in bytes', false),
('app.session_timeout', '86400', 'integer', 'Session timeout in seconds', false),
('app.enable_registration', 'true', 'boolean', 'Enable user registration', true),
('app.enable_password_reset', 'true', 'boolean', 'Enable password reset', true),
('mapping.default_zoom', '10', 'integer', 'Default map zoom level', true),
('mapping.max_route_waypoints', '50', 'integer', 'Maximum waypoints per route', false),
('geofence.max_per_user', '100', 'integer', 'Maximum geofences per user', false),
('tak.max_message_size', '1048576', 'integer', 'Maximum TAK message size', false),
('tak.heartbeat_interval', '30', 'integer', 'TAK heartbeat interval in seconds', false)
ON CONFLICT (key) DO NOTHING;

-- Create default admin user (admin password set at deploy time; see ops notes (not stored in repo))
-- Password hash for 'admin123' using bcrypt
INSERT INTO users (id, username, email, password_hash, first_name, last_name, role, is_active, is_verified)
VALUES (
    '00000000-0000-0000-0000-000000000001',
    'admin',
    'admin@gotak.local',
    '$2a$10$7KvpZQsLHTeipj.MCCopyOtM7lp5qXwzYWqUmmA9Nbp3fjQKbikGy', -- bcrypt hash of the deploy-time admin password
    'System',
    'Administrator',
    'admin',
    true,
    true
) ON CONFLICT (id) DO NOTHING;

COMMIT;
