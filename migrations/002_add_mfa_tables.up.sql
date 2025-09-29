-- Add MFA (Multi-Factor Authentication) tables for Sprint 10: Security & Compliance Framework

-- Enable necessary extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- MFA Factors: Store user's enrolled MFA methods
CREATE TABLE IF NOT EXISTS mfa_factors (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type VARCHAR(20) NOT NULL CHECK (type IN ('totp', 'sms', 'email', 'webauthn', 'backup')),
    name VARCHAR(255) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending' 
        CHECK (status IN ('pending', 'active', 'disabled', 'revoked', 'suspended')),
    
    -- Encrypted secret data (using pgcrypto)
    secret_encrypted BYTEA,
    
    -- Provider-specific metadata (JSON)
    metadata JSONB DEFAULT '{}',
    
    -- Encrypted backup codes
    backup_codes_encrypted BYTEA,
    
    -- Usage tracking
    last_used_at TIMESTAMP WITH TIME ZONE,
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Create partial unique index for active factors per type per user
CREATE UNIQUE INDEX IF NOT EXISTS idx_mfa_factors_unique_active 
    ON mfa_factors(user_id, type) 
    WHERE status = 'active';

-- MFA Enrollment Sessions: Track ongoing enrollment processes
CREATE TABLE IF NOT EXISTS mfa_enrollment_sessions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type VARCHAR(20) NOT NULL CHECK (type IN ('totp', 'sms', 'email', 'webauthn')),
    
    -- Encrypted secret data for enrollment
    secret_encrypted BYTEA NOT NULL,
    
    -- Provider-specific metadata
    metadata JSONB DEFAULT '{}',
    
    -- Session expiry
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Auth Challenges: Multi-factor authentication challenges
CREATE TABLE IF NOT EXISTS mfa_auth_challenges (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    
    -- Required MFA types for this challenge
    required_types TEXT[] NOT NULL,
    
    -- Challenge status
    status VARCHAR(20) NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'verified', 'failed', 'expired', 'canceled')),
    
    -- Completed factors
    completed_factors UUID[] DEFAULT '{}',
    
    -- Challenge expiry
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMP WITH TIME ZONE
);

-- MFA Challenges: Individual factor challenges within an auth challenge
CREATE TABLE IF NOT EXISTS mfa_challenges (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    auth_challenge_id UUID REFERENCES mfa_auth_challenges(id) ON DELETE CASCADE,
    factor_id UUID NOT NULL REFERENCES mfa_factors(id) ON DELETE CASCADE,
    type VARCHAR(20) NOT NULL CHECK (type IN ('totp', 'sms', 'email', 'webauthn', 'backup')),
    
    -- Challenge status
    status VARCHAR(20) NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'verified', 'failed', 'expired', 'canceled')),
    
    -- Challenge data (e.g., SMS code, encrypted)
    challenge_data_encrypted BYTEA,
    
    -- Attempt tracking
    attempts INTEGER NOT NULL DEFAULT 0,
    max_attempts INTEGER NOT NULL DEFAULT 3,
    
    -- Challenge expiry
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    verified_at TIMESTAMP WITH TIME ZONE
);

-- MFA Events: Audit log for MFA operations
CREATE TABLE IF NOT EXISTS mfa_events (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    factor_id UUID REFERENCES mfa_factors(id) ON DELETE SET NULL,
    challenge_id UUID REFERENCES mfa_challenges(id) ON DELETE SET NULL,
    
    -- Event details
    event_type VARCHAR(50) NOT NULL, -- enrollment, challenge, verification, etc.
    result VARCHAR(20) NOT NULL,     -- success, failure, error
    
    -- Additional metadata
    metadata JSONB DEFAULT '{}',
    
    -- Request context
    ip_address INET,
    user_agent TEXT,
    
    -- Timestamp
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Certificate Identities: X.509 certificate-based authentication
CREATE TABLE IF NOT EXISTS certificate_identities (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    
    -- Certificate details
    subject_dn TEXT NOT NULL,
    issuer_dn TEXT NOT NULL,
    serial_number TEXT NOT NULL,
    fingerprint_sha256 CHAR(64) NOT NULL,
    
    -- Certificate data (PEM encoded)
    certificate_pem TEXT NOT NULL,
    
    -- Validity period
    not_before TIMESTAMP WITH TIME ZONE NOT NULL,
    not_after TIMESTAMP WITH TIME ZONE NOT NULL,
    
    -- Status
    status VARCHAR(20) NOT NULL DEFAULT 'active'
        CHECK (status IN ('active', 'revoked', 'expired', 'suspended')),
    
    -- Metadata
    metadata JSONB DEFAULT '{}',
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    last_used_at TIMESTAMP WITH TIME ZONE,
    
    -- Ensure unique active certificates
    UNIQUE(fingerprint_sha256)
);

-- Roles: Enhanced RBAC role definitions
CREATE TABLE IF NOT EXISTS roles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL UNIQUE,
    display_name VARCHAR(255) NOT NULL,
    description TEXT,
    
    -- Role hierarchy
    parent_role_id UUID REFERENCES roles(id) ON DELETE SET NULL,
    
    -- Role type
    type VARCHAR(20) NOT NULL DEFAULT 'custom'
        CHECK (type IN ('system', 'builtin', 'custom')),
    
    -- Permissions (array of permission strings)
    permissions TEXT[] DEFAULT '{}',
    
    -- Role status
    status VARCHAR(20) NOT NULL DEFAULT 'active'
        CHECK (status IN ('active', 'disabled', 'deprecated')),
    
    -- Metadata
    metadata JSONB DEFAULT '{}',
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Role Bindings: Associate users with roles
CREATE TABLE IF NOT EXISTS role_bindings (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    
    -- Binding scope (global, mission-specific, etc.)
    scope VARCHAR(50) NOT NULL DEFAULT 'global',
    scope_id UUID, -- Reference to mission, group, etc.
    
    -- Binding validity
    valid_from TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    valid_until TIMESTAMP WITH TIME ZONE,
    
    -- Status
    status VARCHAR(20) NOT NULL DEFAULT 'active'
        CHECK (status IN ('active', 'suspended', 'expired')),
    
    -- Metadata
    metadata JSONB DEFAULT '{}',
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    
    -- Ensure unique active role bindings (index created below)
    UNIQUE(user_id, role_id, scope, scope_id)
);

-- Attribute Policies: ABAC (Attribute-Based Access Control) policies
CREATE TABLE IF NOT EXISTS attribute_policies (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL UNIQUE,
    description TEXT,
    
    -- Policy rule (JSON Logic format)
    rule JSONB NOT NULL,
    
    -- Policy effect
    effect VARCHAR(10) NOT NULL CHECK (effect IN ('allow', 'deny')),
    
    -- Resources this policy applies to
    resources TEXT[] DEFAULT '{}',
    
    -- Policy priority (higher numbers take precedence)
    priority INTEGER NOT NULL DEFAULT 0,
    
    -- Policy status
    status VARCHAR(20) NOT NULL DEFAULT 'active'
        CHECK (status IN ('active', 'disabled', 'draft')),
    
    -- Policy metadata
    metadata JSONB DEFAULT '{}',
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    
    -- Policy version tracking
    version INTEGER NOT NULL DEFAULT 1
);

-- Encryption Keys: Key management for envelope encryption
CREATE TABLE IF NOT EXISTS encryption_keys (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    
    -- Key identifier and version
    key_id VARCHAR(255) NOT NULL,
    version INTEGER NOT NULL DEFAULT 1,
    
    -- Key type and purpose
    type VARCHAR(20) NOT NULL CHECK (type IN ('dek', 'kek', 'hmac', 'signing')),
    purpose VARCHAR(50) NOT NULL, -- user_data, mfa_secrets, certificates, etc.
    
    -- Key data (encrypted with KEK)
    key_data_encrypted BYTEA NOT NULL,
    
    -- Key algorithm and size
    algorithm VARCHAR(20) NOT NULL, -- AES-256, RSA-2048, etc.
    key_size INTEGER NOT NULL,
    
    -- Key status
    status VARCHAR(20) NOT NULL DEFAULT 'active'
        CHECK (status IN ('active', 'rotated', 'deprecated', 'destroyed')),
    
    -- Key lifecycle
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    activated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMP WITH TIME ZONE,
    destroyed_at TIMESTAMP WITH TIME ZONE,
    
    -- External key management system reference
    external_key_id VARCHAR(255),
    kms_provider VARCHAR(50), -- vault, aws-kms, gcp-kms, etc.
    
    -- Metadata
    metadata JSONB DEFAULT '{}',
    
    -- Ensure unique active key versions
    UNIQUE(key_id, version)
);

-- Security Events: Comprehensive security audit log
CREATE TABLE IF NOT EXISTS security_events (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    
    -- Event classification
    event_type VARCHAR(50) NOT NULL,
    event_category VARCHAR(30) NOT NULL, -- authentication, authorization, audit, etc.
    severity VARCHAR(10) NOT NULL CHECK (severity IN ('info', 'warning', 'error', 'critical')),
    
    -- Event source
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    resource_type VARCHAR(50),
    resource_id UUID,
    
    -- Event details
    message TEXT NOT NULL,
    details JSONB DEFAULT '{}',
    
    -- Request context
    ip_address INET,
    user_agent TEXT,
    session_id UUID,
    request_id UUID,
    
    -- Timestamp
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Create partial unique constraint for active encryption keys
CREATE UNIQUE INDEX idx_encryption_keys_unique_active ON encryption_keys(key_id) WHERE status = 'active';

-- Create partial unique constraint for active role bindings
CREATE UNIQUE INDEX idx_role_bindings_unique_active ON role_bindings(user_id, role_id, scope, scope_id) WHERE status = 'active';

-- Create indexes for performance
CREATE INDEX idx_mfa_factors_user_id ON mfa_factors(user_id);
CREATE INDEX idx_mfa_factors_type_status ON mfa_factors(type, status);
CREATE INDEX idx_security_events_type_created ON security_events(event_type, created_at);
CREATE INDEX idx_security_events_user_created ON security_events(user_id, created_at);
CREATE INDEX idx_security_events_severity_created ON security_events(severity, created_at);
CREATE INDEX idx_mfa_enrollment_sessions_user_id ON mfa_enrollment_sessions(user_id);
CREATE INDEX idx_mfa_enrollment_sessions_expires ON mfa_enrollment_sessions(expires_at);
CREATE INDEX idx_mfa_auth_challenges_user_id ON mfa_auth_challenges(user_id);
CREATE INDEX idx_mfa_auth_challenges_expires ON mfa_auth_challenges(expires_at);
CREATE INDEX idx_mfa_challenges_factor_id ON mfa_challenges(factor_id);
CREATE INDEX idx_mfa_challenges_auth_challenge_id ON mfa_challenges(auth_challenge_id);
CREATE INDEX idx_mfa_events_user_id ON mfa_events(user_id);
CREATE INDEX idx_mfa_events_created_at ON mfa_events(created_at);
CREATE INDEX idx_certificate_identities_user_id ON certificate_identities(user_id);
CREATE INDEX idx_certificate_identities_fingerprint ON certificate_identities(fingerprint_sha256);
CREATE INDEX idx_roles_name ON roles(name);
CREATE INDEX idx_roles_type_status ON roles(type, status);
CREATE INDEX idx_role_bindings_user_id ON role_bindings(user_id);
CREATE INDEX idx_role_bindings_role_id ON role_bindings(role_id);
CREATE INDEX idx_attribute_policies_status ON attribute_policies(status);
CREATE INDEX idx_encryption_keys_key_id ON encryption_keys(key_id);
CREATE INDEX idx_encryption_keys_status ON encryption_keys(status);

-- Create updated_at triggers for timestamp management
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Apply updated_at triggers
CREATE TRIGGER update_mfa_factors_updated_at
    BEFORE UPDATE ON mfa_factors
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_certificate_identities_updated_at
    BEFORE UPDATE ON certificate_identities
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_roles_updated_at
    BEFORE UPDATE ON roles
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_role_bindings_updated_at
    BEFORE UPDATE ON role_bindings
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_attribute_policies_updated_at
    BEFORE UPDATE ON attribute_policies
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Insert default system roles
INSERT INTO roles (id, name, display_name, description, type, permissions, status) VALUES
(uuid_generate_v4(), 'system:admin', 'System Administrator', 'Full system administration privileges', 'system', 
 ARRAY['*'], 'active'),

(uuid_generate_v4(), 'system:auditor', 'Security Auditor', 'Read-only access for security auditing', 'system',
 ARRAY['audit:read', 'security:read', 'users:read', 'roles:read'], 'active'),

(uuid_generate_v4(), 'tactical:commander', 'Tactical Commander', 'Mission command and control privileges', 'builtin',
 ARRAY['missions:*', 'entities:*', 'communications:*', 'maps:*'], 'active'),

(uuid_generate_v4(), 'tactical:operator', 'Tactical Operator', 'Standard tactical operations privileges', 'builtin',
 ARRAY['missions:read', 'missions:update', 'entities:*', 'communications:send', 'maps:read'], 'active'),

(uuid_generate_v4(), 'tactical:dispatcher', 'Dispatcher', 'Emergency dispatch and coordination privileges', 'builtin',
 ARRAY['missions:create', 'missions:read', 'missions:update', 'communications:*', 'alerts:*'], 'active'),

(uuid_generate_v4(), 'tactical:viewer', 'Tactical Viewer', 'Read-only tactical situation awareness', 'builtin',
 ARRAY['missions:read', 'entities:read', 'maps:read', 'communications:read'], 'active');

-- Insert default attribute policies for basic security controls
INSERT INTO attribute_policies (id, name, description, rule, effect, resources, priority, status) VALUES
(uuid_generate_v4(), 'time-based-access', 'Allow access only during business hours', 
 '{"and": [{">=": [{"var": "time.hour"}, 6]}, {"<=": [{"var": "time.hour"}, 18]}]}', 
 'allow', ARRAY['missions:*'], 10, 'active'),

(uuid_generate_v4(), 'ip-whitelist', 'Allow access only from approved networks',
 '{"in": [{"var": "request.ip_subnet"}, ["10.0.0.0/8", "192.168.0.0/16", "172.16.0.0/12"]]}',
 'allow', ARRAY['*'], 100, 'active'),

(uuid_generate_v4(), 'mfa-required-admin', 'Require MFA for administrative operations',
 '{"and": [{"in": ["admin", {"var": "user.roles"}]}, {"==": [{"var": "auth.mfa_verified"}, true]}]}',
 'allow', ARRAY['users:*', 'roles:*', 'security:*'], 50, 'active');

-- Add comments for documentation
COMMENT ON TABLE mfa_factors IS 'Stores enrolled multi-factor authentication methods for users';
COMMENT ON TABLE mfa_enrollment_sessions IS 'Tracks ongoing MFA enrollment processes with temporary secrets';
COMMENT ON TABLE mfa_auth_challenges IS 'Multi-factor authentication challenges requiring multiple factors';
COMMENT ON TABLE mfa_challenges IS 'Individual factor challenges within an authentication session';
COMMENT ON TABLE mfa_events IS 'Comprehensive audit log for all MFA operations';
COMMENT ON TABLE certificate_identities IS 'X.509 certificate-based authentication identities (CAC/PIV)';
COMMENT ON TABLE roles IS 'Role definitions for role-based access control (RBAC)';
COMMENT ON TABLE role_bindings IS 'Associates users with roles for specific scopes';
COMMENT ON TABLE attribute_policies IS 'Attribute-based access control (ABAC) policy definitions';
COMMENT ON TABLE encryption_keys IS 'Key management for envelope encryption and key lifecycle';
COMMENT ON TABLE security_events IS 'Comprehensive security audit and event logging';
