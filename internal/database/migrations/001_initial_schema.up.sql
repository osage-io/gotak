-- Initial schema for GoTAK
-- This migration sets up the basic foundation tables

-- Enable required extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Create application schema
CREATE SCHEMA IF NOT EXISTS gotak;
CREATE SCHEMA IF NOT EXISTS audit;

-- Set search path
SET search_path TO gotak, public;

-- Basic users table (will be expanded in Sprint 2)
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    username VARCHAR(255) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    first_name VARCHAR(255),
    last_name VARCHAR(255),
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Audit events table for comprehensive logging
CREATE TABLE audit.events (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    event_time TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    user_id UUID REFERENCES gotak.users(id),
    action VARCHAR(100) NOT NULL,
    resource VARCHAR(255),
    resource_id VARCHAR(255),
    details JSONB,
    ip_address INET,
    user_agent TEXT,
    session_id VARCHAR(255),
    trace_id VARCHAR(255)
);

-- System configuration table
CREATE TABLE system_config (
    key VARCHAR(255) PRIMARY KEY,
    value JSONB NOT NULL,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes for performance
CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_active ON users(is_active);
CREATE INDEX idx_users_created_at ON users(created_at);

CREATE INDEX idx_audit_events_time ON audit.events(event_time);
CREATE INDEX idx_audit_events_user ON audit.events(user_id);
CREATE INDEX idx_audit_events_action ON audit.events(action);
CREATE INDEX idx_audit_events_resource ON audit.events(resource);
CREATE INDEX idx_audit_events_trace ON audit.events(trace_id);

CREATE INDEX idx_system_config_key ON system_config(key);

-- Insert initial system configuration
INSERT INTO system_config (key, value, description) VALUES 
    ('database_version', '"1"', 'Database schema version'),
    ('created_at', to_jsonb(NOW()::text), 'Database creation timestamp'),
    ('environment', '"development"', 'Environment identifier');

-- Insert initial development users
INSERT INTO users (username, email, first_name, last_name) VALUES 
    ('admin', 'admin@gotak.dev', 'Admin', 'User'),
    ('operator', 'operator@gotak.dev', 'Operator', 'User'),
    ('commander', 'commander@gotak.dev', 'Commander', 'User');

-- Grant permissions
GRANT USAGE ON SCHEMA gotak TO gotak;
GRANT USAGE ON SCHEMA audit TO gotak;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA gotak TO gotak;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA audit TO gotak;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA gotak TO gotak;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA audit TO gotak;

-- Set default privileges for future objects
ALTER DEFAULT PRIVILEGES IN SCHEMA gotak GRANT ALL ON TABLES TO gotak;
ALTER DEFAULT PRIVILEGES IN SCHEMA audit GRANT ALL ON TABLES TO gotak;
ALTER DEFAULT PRIVILEGES IN SCHEMA gotak GRANT ALL ON SEQUENCES TO gotak;
ALTER DEFAULT PRIVILEGES IN SCHEMA audit GRANT ALL ON SEQUENCES TO gotak;

-- Comments for documentation
COMMENT ON SCHEMA gotak IS 'Main GoTAK application schema';
COMMENT ON SCHEMA audit IS 'Audit logging schema for security and compliance';
COMMENT ON TABLE users IS 'Basic user accounts (will be extended in Sprint 2)';
COMMENT ON TABLE audit.events IS 'Comprehensive audit trail for all system events';
COMMENT ON TABLE system_config IS 'System configuration and metadata storage';
