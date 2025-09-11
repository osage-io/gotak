-- Rollback for GoTAK Database Schema Migration 001
-- This script completely removes the initial schema
-- Version: 1.0.0
-- Created: 2024-01-01

BEGIN;

-- Drop audit triggers first
DROP TRIGGER IF EXISTS audit_users_trigger ON users;
DROP TRIGGER IF EXISTS audit_routes_trigger ON routes;
DROP TRIGGER IF EXISTS audit_geofences_trigger ON geofences;

-- Drop update triggers
DROP TRIGGER IF EXISTS update_users_updated_at ON users;
DROP TRIGGER IF EXISTS update_routes_updated_at ON routes;
DROP TRIGGER IF EXISTS update_geofences_updated_at ON geofences;
DROP TRIGGER IF EXISTS update_offline_areas_updated_at ON offline_areas;
DROP TRIGGER IF EXISTS update_tactical_overlays_updated_at ON tactical_overlays;

-- Drop functions
DROP FUNCTION IF EXISTS audit_trigger_function();
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop tables in reverse order (respecting foreign key constraints)
DROP TABLE IF EXISTS audit_trail;
DROP TABLE IF EXISTS system_logs;
DROP TABLE IF EXISTS config_settings;
DROP TABLE IF EXISTS tactical_overlays;
DROP TABLE IF EXISTS offline_areas;
DROP TABLE IF EXISTS geofence_violations;
DROP TABLE IF EXISTS geofences;
DROP TABLE IF EXISTS waypoints;
DROP TABLE IF EXISTS routes;
DROP TABLE IF EXISTS user_sessions;
DROP TABLE IF EXISTS users;

-- Drop extensions (optional - may be used by other applications)
-- DROP EXTENSION IF EXISTS "btree_gist";
-- DROP EXTENSION IF EXISTS "pg_trgm";
-- DROP EXTENSION IF EXISTS "postgis";
-- DROP EXTENSION IF EXISTS "uuid-ossp";

-- Note: We don't drop schema_migrations table as it tracks migration history

COMMIT;
