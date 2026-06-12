-- Migration: Add Users and Authentication Tables
-- Description: Creates tables for user management, authentication sessions, and user preferences

-- Users table
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL DEFAULT 'operator',
    callsign VARCHAR(50),
    active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    last_login TIMESTAMP WITH TIME ZONE
);

-- 001_initial_schema also creates users (with is_active instead of active, no
-- callsign/last_login). When that shape already exists, the CREATE TABLE above
-- no-ops — reconcile the columns this migration depends on.
ALTER TABLE users ADD COLUMN IF NOT EXISTS active BOOLEAN DEFAULT true;
ALTER TABLE users ADD COLUMN IF NOT EXISTS callsign VARCHAR(50);
ALTER TABLE users ADD COLUMN IF NOT EXISTS last_login TIMESTAMP WITH TIME ZONE;

-- Sessions table for active sessions
CREATE TABLE IF NOT EXISTS sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash VARCHAR(255) NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    ip_address VARCHAR(45),
    user_agent TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- User preferences
CREATE TABLE IF NOT EXISTS user_preferences (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    theme VARCHAR(20) DEFAULT 'dark',
    language VARCHAR(10) DEFAULT 'en',
    notification_settings JSONB DEFAULT '{}',
    map_preferences JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Password reset tokens
CREATE TABLE IF NOT EXISTS password_reset_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash VARCHAR(255) NOT NULL UNIQUE,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    used BOOLEAN DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Login attempts for rate limiting
CREATE TABLE IF NOT EXISTS login_attempts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username VARCHAR(255),
    ip_address VARCHAR(45) NOT NULL,
    success BOOLEAN DEFAULT false,
    attempted_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for performance.
-- IF NOT EXISTS so this is idempotent: the users table + some of its indexes are
-- already created by 001_initial_schema, so unguarded CREATE INDEX would fail.
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
CREATE INDEX IF NOT EXISTS idx_users_active ON users(active);
CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_sessions_token_hash ON sessions(token_hash);
CREATE INDEX IF NOT EXISTS idx_sessions_expires_at ON sessions(expires_at);
CREATE INDEX IF NOT EXISTS idx_password_reset_tokens_token_hash ON password_reset_tokens(token_hash);
CREATE INDEX IF NOT EXISTS idx_password_reset_tokens_user_id ON password_reset_tokens(user_id);
CREATE INDEX IF NOT EXISTS idx_login_attempts_ip_address ON login_attempts(ip_address);
CREATE INDEX IF NOT EXISTS idx_login_attempts_username ON login_attempts(username);
CREATE INDEX IF NOT EXISTS idx_login_attempts_attempted_at ON login_attempts(attempted_at);

-- Create trigger to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Apply trigger to users table.
-- CREATE OR REPLACE so this is idempotent: 001_initial_schema already creates an
-- update_users_updated_at trigger. (Requires PostgreSQL 14+, image is PG 15.)
CREATE OR REPLACE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Apply trigger to user_preferences table
CREATE OR REPLACE TRIGGER update_user_preferences_updated_at BEFORE UPDATE ON user_preferences
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Insert default admin user (password: admin123)
-- Note: Change this password immediately in production!
INSERT INTO users (username, email, password_hash, role, callsign, active)
VALUES (
    'admin',
    'admin@gotak.local',
    '$2a$10$QVVSQmJPalYzWEd2LlcG4u0VqXuv7nH/fObMH68/8ISpLUD4zzwae', -- bcrypt hash of 'admin123'
    'admin',
    'ADMIN',
    true
) ON CONFLICT (username) DO NOTHING;

-- Add comments for documentation
COMMENT ON TABLE users IS 'System users for authentication and authorization';
COMMENT ON TABLE sessions IS 'Active user sessions for JWT token validation';
COMMENT ON TABLE user_preferences IS 'User-specific preferences and settings';
COMMENT ON TABLE password_reset_tokens IS 'Tokens for password reset functionality';
COMMENT ON TABLE login_attempts IS 'Track login attempts for rate limiting';

COMMENT ON COLUMN users.role IS 'User role: admin, operator, viewer';
COMMENT ON COLUMN users.callsign IS 'TAK callsign for the user';
COMMENT ON COLUMN sessions.token_hash IS 'SHA256 hash of the JWT token';
COMMENT ON COLUMN login_attempts.ip_address IS 'IP address of the login attempt';