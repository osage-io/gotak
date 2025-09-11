-- GoTAK Development Database Initialization
-- This script sets up the initial database structure for development

-- Enable required extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Create basic schemas
CREATE SCHEMA IF NOT EXISTS gotak;
CREATE SCHEMA IF NOT EXISTS audit;

-- Set search path
SET search_path TO gotak, public;

-- Basic migrations table for tracking schema changes
CREATE TABLE IF NOT EXISTS schema_migrations (
    version BIGINT PRIMARY KEY,
    dirty BOOLEAN NOT NULL DEFAULT FALSE,
    applied_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Development user for testing (will be replaced by proper auth in Sprint 2)
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    username VARCHAR(255) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Insert development test users
INSERT INTO users (username, email) VALUES 
    ('admin', 'admin@gotak.dev'),
    ('operator', 'operator@gotak.dev'),
    ('commander', 'commander@gotak.dev')
ON CONFLICT (username) DO NOTHING;

-- Audit logging table
CREATE TABLE IF NOT EXISTS audit.events (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    event_time TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    user_id UUID REFERENCES users(id),
    action VARCHAR(100) NOT NULL,
    resource VARCHAR(255),
    resource_id VARCHAR(255),
    details JSONB,
    ip_address INET,
    user_agent TEXT
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_audit_events_time ON audit.events(event_time);
CREATE INDEX IF NOT EXISTS idx_audit_events_user ON audit.events(user_id);
CREATE INDEX IF NOT EXISTS idx_audit_events_action ON audit.events(action);

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

COMMENT ON DATABASE gotak_dev IS 'GoTAK Development Database';
COMMENT ON SCHEMA gotak IS 'Main application schema';
COMMENT ON SCHEMA audit IS 'Audit logging schema';
