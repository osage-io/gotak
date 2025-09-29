-- Test migration to fix syntax issues
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Test the attribute_policies table creation
CREATE TABLE IF NOT EXISTS attribute_policies_test (
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

SELECT 'Test migration completed successfully' as status;
