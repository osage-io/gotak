-- Sprint 2 Down Migration: Rollback Authentication & Security Foundation
-- This migration removes all authentication and RBAC related tables and columns

-- Set search path
SET search_path TO gotak, public;

-- Drop authentication event audit table
DROP TABLE IF EXISTS audit.authentication_events CASCADE;

-- Drop Casbin rules table
DROP TABLE IF EXISTS casbin_rules CASCADE;

-- Drop user sessions table
DROP TABLE IF EXISTS user_sessions CASCADE;

-- Drop password reset tokens table
DROP TABLE IF EXISTS password_reset_tokens CASCADE;

-- Drop refresh tokens table
DROP TABLE IF EXISTS refresh_tokens CASCADE;

-- Drop role permissions junction table
DROP TABLE IF EXISTS role_permissions CASCADE;

-- Drop permissions table
DROP TABLE IF EXISTS permissions CASCADE;

-- Drop user roles junction table
DROP TABLE IF EXISTS user_roles CASCADE;

-- Drop roles table
DROP TABLE IF EXISTS roles CASCADE;

-- Remove authentication fields from users table
ALTER TABLE users DROP COLUMN IF EXISTS locked_until;
ALTER TABLE users DROP COLUMN IF EXISTS failed_attempts;
ALTER TABLE users DROP COLUMN IF EXISTS last_login;
ALTER TABLE users DROP COLUMN IF EXISTS mfa_secret;
ALTER TABLE users DROP COLUMN IF EXISTS mfa_enabled;
ALTER TABLE users DROP COLUMN IF EXISTS password_hash;

-- Drop indexes that were added for authentication fields
DROP INDEX IF EXISTS idx_users_locked_until;
DROP INDEX IF EXISTS idx_users_failed_attempts;
DROP INDEX IF EXISTS idx_users_last_login;

-- Revert system configuration changes
DELETE FROM system_config WHERE key IN (
    'auth_system_enabled',
    'rbac_system_enabled', 
    'mfa_available',
    'password_policy_enabled'
);

UPDATE system_config 
SET 
    value = '"1"',
    updated_at = NOW()
WHERE key = 'database_version';
