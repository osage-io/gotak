-- Sprint 2: Authentication & Security Foundation
-- This migration extends the basic users table with authentication features
-- and adds supporting tables for roles, permissions, sessions, and security

-- Set search path
SET search_path TO gotak, public;

-- Extend users table with authentication fields
ALTER TABLE users ADD COLUMN IF NOT EXISTS first_name VARCHAR(255);
ALTER TABLE users ADD COLUMN IF NOT EXISTS last_name VARCHAR(255);
ALTER TABLE users ADD COLUMN IF NOT EXISTS password_hash VARCHAR(255);
ALTER TABLE users ADD COLUMN IF NOT EXISTS mfa_enabled BOOLEAN DEFAULT FALSE;
ALTER TABLE users ADD COLUMN IF NOT EXISTS mfa_secret VARCHAR(255);
ALTER TABLE users ADD COLUMN IF NOT EXISTS last_login TIMESTAMP WITH TIME ZONE;
ALTER TABLE users ADD COLUMN IF NOT EXISTS failed_attempts INTEGER DEFAULT 0;
ALTER TABLE users ADD COLUMN IF NOT EXISTS locked_until TIMESTAMP WITH TIME ZONE;

-- Add constraints and indexes for new fields
CREATE INDEX IF NOT EXISTS idx_users_last_login ON users(last_login);
CREATE INDEX IF NOT EXISTS idx_users_failed_attempts ON users(failed_attempts);
CREATE INDEX IF NOT EXISTS idx_users_locked_until ON users(locked_until);

-- Roles table for RBAC system
CREATE TABLE IF NOT EXISTS roles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) UNIQUE NOT NULL,
    description TEXT,
    is_system_role BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- User roles junction table
CREATE TABLE IF NOT EXISTS user_roles (
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    role_id UUID REFERENCES roles(id) ON DELETE CASCADE,
    granted_by UUID REFERENCES users(id),
    granted_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    expires_at TIMESTAMP WITH TIME ZONE,
    is_active BOOLEAN DEFAULT TRUE,
    PRIMARY KEY (user_id, role_id)
);

-- Permissions table for fine-grained access control
CREATE TABLE permissions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) UNIQUE NOT NULL,
    resource VARCHAR(255) NOT NULL,
    action VARCHAR(255) NOT NULL,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Role permissions junction table
CREATE TABLE role_permissions (
    role_id UUID REFERENCES roles(id) ON DELETE CASCADE,
    permission_id UUID REFERENCES permissions(id) ON DELETE CASCADE,
    granted_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    PRIMARY KEY (role_id, permission_id)
);

-- Refresh tokens table for JWT token management
CREATE TABLE refresh_tokens (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    token_hash VARCHAR(255) NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    revoked_at TIMESTAMP WITH TIME ZONE,
    last_used_at TIMESTAMP WITH TIME ZONE,
    device_info JSONB,
    ip_address INET
);

-- Password reset tokens table
CREATE TABLE password_reset_tokens (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    token_hash VARCHAR(255) NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    used_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    ip_address INET
);

-- User sessions table for session management
CREATE TABLE user_sessions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    session_token VARCHAR(255) UNIQUE NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_accessed_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    ip_address INET,
    user_agent TEXT,
    device_fingerprint VARCHAR(255),
    is_active BOOLEAN DEFAULT TRUE
);

-- Casbin RBAC policy rules table (required by Casbin adapter)
CREATE TABLE casbin_rules (
    id SERIAL PRIMARY KEY,
    ptype VARCHAR(255) NOT NULL,
    v0 VARCHAR(255),
    v1 VARCHAR(255),
    v2 VARCHAR(255),
    v3 VARCHAR(255),
    v4 VARCHAR(255),
    v5 VARCHAR(255)
);

-- Security audit table for authentication events (extends audit.events)
CREATE TABLE audit.authentication_events (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES gotak.users(id),
    event_type VARCHAR(50) NOT NULL, -- login, logout, mfa_setup, password_change, etc.
    auth_method VARCHAR(50), -- local, vault_oidc, certificate
    success BOOLEAN NOT NULL,
    failure_reason VARCHAR(255),
    ip_address INET,
    user_agent TEXT,
    device_fingerprint VARCHAR(255),
    mfa_used BOOLEAN DEFAULT FALSE,
    session_id VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Performance indexes
CREATE INDEX idx_roles_name ON roles(name);
CREATE INDEX idx_user_roles_user_id ON user_roles(user_id);
CREATE INDEX idx_user_roles_role_id ON user_roles(role_id);
CREATE INDEX idx_user_roles_active ON user_roles(is_active);
CREATE INDEX idx_permissions_resource_action ON permissions(resource, action);
CREATE INDEX idx_role_permissions_role_id ON role_permissions(role_id);
CREATE INDEX idx_refresh_tokens_user_id ON refresh_tokens(user_id);
CREATE INDEX idx_refresh_tokens_expires_at ON refresh_tokens(expires_at);
CREATE INDEX idx_refresh_tokens_revoked ON refresh_tokens(revoked_at) WHERE revoked_at IS NULL;
CREATE INDEX idx_password_reset_tokens_user_id ON password_reset_tokens(user_id);
CREATE INDEX idx_password_reset_tokens_expires_at ON password_reset_tokens(expires_at);
CREATE INDEX idx_user_sessions_user_id ON user_sessions(user_id);
CREATE INDEX idx_user_sessions_session_token ON user_sessions(session_token);
CREATE INDEX idx_user_sessions_expires_at ON user_sessions(expires_at);
CREATE INDEX idx_user_sessions_active ON user_sessions(is_active);
CREATE INDEX idx_casbin_rules_ptype_v0_v1 ON casbin_rules(ptype, v0, v1);
CREATE INDEX idx_auth_events_user_id ON audit.authentication_events(user_id);
CREATE INDEX idx_auth_events_type ON audit.authentication_events(event_type);
CREATE INDEX idx_auth_events_created_at ON audit.authentication_events(created_at);

-- Insert default system roles for military hierarchy
INSERT INTO roles (name, description, is_system_role) VALUES
    ('system_admin', 'Full system administration access with all permissions', TRUE),
    ('mission_commander', 'Mission planning and personnel management authority', TRUE),
    ('operations_officer', 'Operational coordination and tactical oversight', TRUE),
    ('operator', 'Standard operational access for field personnel', TRUE),
    ('observer', 'Read-only access to operational picture and reports', TRUE),
    ('guest', 'Limited access for external stakeholders', TRUE);

-- Insert default permissions following military access patterns
INSERT INTO permissions (name, resource, action, description) VALUES
    -- System administration
    ('system.admin', 'system', 'admin', 'Full system administration'),
    ('system.config', 'system', 'configure', 'System configuration management'),
    
    -- User management
    ('users.create', 'users', 'create', 'Create new user accounts'),
    ('users.read', 'users', 'read', 'View user information'),
    ('users.update', 'users', 'update', 'Modify user accounts'),
    ('users.delete', 'users', 'delete', 'Delete user accounts'),
    ('users.admin', 'users', 'admin', 'Full user administration'),
    
    -- Role management
    ('roles.read', 'roles', 'read', 'View roles and permissions'),
    ('roles.assign', 'roles', 'assign', 'Assign roles to users'),
    ('roles.admin', 'roles', 'admin', 'Full role administration'),
    
    -- Mission and operations
    ('missions.create', 'missions', 'create', 'Create new missions'),
    ('missions.read', 'missions', 'read', 'View mission information'),
    ('missions.update', 'missions', 'update', 'Modify mission details'),
    ('missions.delete', 'missions', 'delete', 'Delete missions'),
    ('missions.execute', 'missions', 'execute', 'Execute and manage active missions'),
    
    -- CoT and situational awareness
    ('cot.read', 'cot', 'read', 'View cursor-on-target messages'),
    ('cot.send', 'cot', 'send', 'Send CoT messages'),
    ('cot.admin', 'cot', 'admin', 'Manage CoT routing and filters'),
    
    -- Reporting and analytics
    ('reports.read', 'reports', 'read', 'View reports and analytics'),
    ('reports.create', 'reports', 'create', 'Generate reports'),
    
    -- Audit and security
    ('audit.read', 'audit', 'read', 'View audit logs'),
    ('security.admin', 'security', 'admin', 'Security administration');

-- Assign default permissions to roles
WITH role_permission_assignments AS (
    SELECT 
        r.id as role_id,
        p.id as permission_id
    FROM roles r
    CROSS JOIN permissions p
    WHERE 
        -- System Admin gets all permissions
        (r.name = 'system_admin') OR
        -- Mission Commander permissions
        (r.name = 'mission_commander' AND p.name IN (
            'users.read', 'users.update', 'roles.read', 'roles.assign',
            'missions.create', 'missions.read', 'missions.update', 'missions.delete', 'missions.execute',
            'cot.read', 'cot.send', 'reports.read', 'reports.create', 'audit.read'
        )) OR
        -- Operations Officer permissions
        (r.name = 'operations_officer' AND p.name IN (
            'users.read', 'missions.read', 'missions.update', 'missions.execute',
            'cot.read', 'cot.send', 'reports.read', 'reports.create'
        )) OR
        -- Operator permissions
        (r.name = 'operator' AND p.name IN (
            'missions.read', 'cot.read', 'cot.send', 'reports.read'
        )) OR
        -- Observer permissions
        (r.name = 'observer' AND p.name IN (
            'missions.read', 'cot.read', 'reports.read'
        )) OR
        -- Guest permissions (minimal)
        (r.name = 'guest' AND p.name IN (
            'cot.read', 'reports.read'
        ))
)
INSERT INTO role_permissions (role_id, permission_id)
SELECT role_id, permission_id FROM role_permission_assignments;

-- Assign roles to existing development users
-- Admin user gets system_admin role
INSERT INTO user_roles (user_id, role_id, granted_by)
SELECT 
    u.id as user_id,
    r.id as role_id,
    u.id as granted_by  -- Self-granted for initial setup
FROM users u
CROSS JOIN roles r
WHERE u.username = 'admin' AND r.name = 'system_admin';

-- Commander user gets mission_commander role
INSERT INTO user_roles (user_id, role_id, granted_by)
SELECT 
    u.id as user_id,
    r.id as role_id,
    (SELECT id FROM users WHERE username = 'admin') as granted_by
FROM users u
CROSS JOIN roles r
WHERE u.username = 'commander' AND r.name = 'mission_commander';

-- Operator user gets operator role
INSERT INTO user_roles (user_id, role_id, granted_by)
SELECT 
    u.id as user_id,
    r.id as role_id,
    (SELECT id FROM users WHERE username = 'admin') as granted_by
FROM users u
CROSS JOIN roles r
WHERE u.username = 'operator' AND r.name = 'operator';

-- Create system configuration table if it doesn't exist (from migration 001)
CREATE TABLE IF NOT EXISTS system_config (
    key VARCHAR(255) PRIMARY KEY,
    value JSONB NOT NULL,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Insert or update system configuration to reflect Sprint 2 completion
INSERT INTO system_config (key, value, description) VALUES 
    ('database_version', '"2"', 'Database schema version'),
    ('auth_system_enabled', 'true', 'Authentication system is active'),
    ('rbac_system_enabled', 'true', 'Role-based access control is active'),
    ('mfa_available', 'true', 'Multi-factor authentication is available'),
    ('password_policy_enabled', 'true', 'Password complexity policies are enforced')
ON CONFLICT (key) DO UPDATE SET 
    value = EXCLUDED.value,
    updated_at = NOW();

-- Add table comments for documentation
COMMENT ON TABLE roles IS 'System roles for role-based access control';
COMMENT ON TABLE user_roles IS 'User role assignments with audit trail';
COMMENT ON TABLE permissions IS 'Fine-grained permissions for resources and actions';
COMMENT ON TABLE role_permissions IS 'Role to permission mappings';
COMMENT ON TABLE refresh_tokens IS 'JWT refresh tokens for secure authentication';
COMMENT ON TABLE password_reset_tokens IS 'Secure password reset tokens';
COMMENT ON TABLE user_sessions IS 'Active user sessions for session management';
COMMENT ON TABLE casbin_rules IS 'Casbin RBAC policy rules storage';
COMMENT ON TABLE audit.authentication_events IS 'Detailed authentication event logging';

-- Grant appropriate permissions to gotak user
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA gotak TO gotak;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA audit TO gotak;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA gotak TO gotak;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA audit TO gotak;
