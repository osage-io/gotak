-- Chat System Migration
-- Adds comprehensive chat functionality with rooms, messages, and participants

-- Chat rooms (channels, groups, or private conversations)
CREATE TABLE chat_rooms (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    type VARCHAR(50) NOT NULL DEFAULT 'group', -- 'group', 'private', 'tactical', 'emergency'
    classification VARCHAR(20) DEFAULT 'UNCLASSIFIED', -- Security classification
    created_by UUID REFERENCES auth_users(id) ON DELETE SET NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    is_active BOOLEAN DEFAULT TRUE,
    
    -- Tactical-specific fields
    mission_id UUID, -- Optional reference to missions
    geographic_bounds JSONB, -- Optional geographic constraints
    
    -- Room settings
    settings JSONB DEFAULT '{}'::jsonb, -- Room configuration (max participants, etc.)
    
    -- Indexing
    CONSTRAINT valid_room_type CHECK (type IN ('group', 'private', 'tactical', 'emergency')),
    CONSTRAINT valid_classification CHECK (classification IN ('UNCLASSIFIED', 'RESTRICTED', 'CONFIDENTIAL', 'SECRET', 'TOP_SECRET'))
);

-- Chat room participants/members
CREATE TABLE chat_room_participants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    room_id UUID NOT NULL REFERENCES chat_rooms(id) ON DELETE CASCADE,
    user_id UUID REFERENCES auth_users(id) ON DELETE CASCADE,
    callsign VARCHAR(100), -- TAK callsign for tactical users
    
    -- Participant role and permissions
    role VARCHAR(50) DEFAULT 'member', -- 'admin', 'moderator', 'member', 'observer'
    permissions JSONB DEFAULT '{}'::jsonb,
    
    -- Timestamps
    joined_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_seen TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    is_active BOOLEAN DEFAULT TRUE,
    
    -- Constraints
    UNIQUE(room_id, user_id),
    CONSTRAINT valid_participant_role CHECK (role IN ('admin', 'moderator', 'member', 'observer'))
);

-- Chat messages
CREATE TABLE chat_messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    room_id UUID NOT NULL REFERENCES chat_rooms(id) ON DELETE CASCADE,
    sender_id UUID REFERENCES auth_users(id) ON DELETE SET NULL,
    sender_callsign VARCHAR(100), -- TAK callsign for non-authenticated senders
    
    -- Message content
    message_text TEXT NOT NULL,
    message_type VARCHAR(50) DEFAULT 'text', -- 'text', 'system', 'position', 'emergency', 'tactical_report'
    priority VARCHAR(20) DEFAULT 'normal', -- 'low', 'normal', 'high', 'urgent', 'emergency'
    
    -- Message metadata
    cot_event_uid VARCHAR(255), -- Reference to original CoT event UID
    cot_event_type VARCHAR(100), -- Original CoT event type (b-t-f, etc.)
    cot_raw_xml TEXT, -- Original CoT XML for replay/analysis
    
    -- Geographic information (for position-based messages)
    location_lat DECIMAL(10, 8),
    location_lng DECIMAL(11, 8),
    location_alt DECIMAL(10, 3),
    
    -- Tactical information
    classification VARCHAR(20) DEFAULT 'UNCLASSIFIED',
    keywords TEXT[], -- Message keywords for searching
    
    -- Threading and replies
    reply_to_id UUID REFERENCES chat_messages(id) ON DELETE SET NULL,
    thread_id UUID, -- For message threading
    
    -- Status tracking
    is_deleted BOOLEAN DEFAULT FALSE,
    deleted_at TIMESTAMP WITH TIME ZONE,
    deleted_by UUID REFERENCES auth_users(id) ON DELETE SET NULL,
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    -- Message acknowledgment (for critical messages)
    requires_ack BOOLEAN DEFAULT FALSE,
    
    -- Indexing constraints
    CONSTRAINT valid_message_type CHECK (message_type IN ('text', 'system', 'position', 'emergency', 'tactical_report')),
    CONSTRAINT valid_priority CHECK (priority IN ('low', 'normal', 'high', 'urgent', 'emergency')),
    CONSTRAINT valid_classification CHECK (classification IN ('UNCLASSIFIED', 'RESTRICTED', 'CONFIDENTIAL', 'SECRET', 'TOP_SECRET'))
);

-- Message acknowledgments (for tracking who has seen critical messages)
CREATE TABLE message_acknowledgments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    message_id UUID NOT NULL REFERENCES chat_messages(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES auth_users(id) ON DELETE CASCADE,
    acknowledged_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    UNIQUE(message_id, user_id)
);

-- Message reactions/status indicators
CREATE TABLE message_reactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    message_id UUID NOT NULL REFERENCES chat_messages(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES auth_users(id) ON DELETE CASCADE,
    reaction_type VARCHAR(50) NOT NULL, -- 'roger', 'wilco', 'negative', 'like', 'important'
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    UNIQUE(message_id, user_id, reaction_type),
    CONSTRAINT valid_reaction CHECK (reaction_type IN ('roger', 'wilco', 'negative', 'like', 'important', 'question'))
);

-- Create indexes for performance
CREATE INDEX idx_chat_rooms_type ON chat_rooms(type);
CREATE INDEX idx_chat_rooms_active ON chat_rooms(is_active) WHERE is_active = TRUE;
CREATE INDEX idx_chat_rooms_classification ON chat_rooms(classification);

CREATE INDEX idx_chat_room_participants_room_id ON chat_room_participants(room_id);
CREATE INDEX idx_chat_room_participants_user_id ON chat_room_participants(user_id);
CREATE INDEX idx_chat_room_participants_active ON chat_room_participants(room_id, is_active) WHERE is_active = TRUE;

CREATE INDEX idx_chat_messages_room_id ON chat_messages(room_id);
CREATE INDEX idx_chat_messages_sender_id ON chat_messages(sender_id);
CREATE INDEX idx_chat_messages_created_at ON chat_messages(created_at DESC);
CREATE INDEX idx_chat_messages_type_priority ON chat_messages(message_type, priority);
CREATE INDEX idx_chat_messages_classification ON chat_messages(classification);
CREATE INDEX idx_chat_messages_location ON chat_messages(location_lat, location_lng) WHERE location_lat IS NOT NULL;
CREATE INDEX idx_chat_messages_cot_uid ON chat_messages(cot_event_uid) WHERE cot_event_uid IS NOT NULL;
CREATE INDEX idx_chat_messages_thread ON chat_messages(thread_id) WHERE thread_id IS NOT NULL;
CREATE INDEX idx_chat_messages_active ON chat_messages(room_id, created_at DESC) WHERE is_deleted = FALSE;

CREATE INDEX idx_message_acknowledgments_message_id ON message_acknowledgments(message_id);
CREATE INDEX idx_message_acknowledgments_user_id ON message_acknowledgments(user_id);

CREATE INDEX idx_message_reactions_message_id ON message_reactions(message_id);
CREATE INDEX idx_message_reactions_type ON message_reactions(reaction_type);

-- Create default "All Users" chat room
INSERT INTO chat_rooms (id, name, description, type, classification) 
VALUES (
    '00000000-0000-0000-0000-000000000001', 
    'All Users', 
    'General communication channel for all tactical users', 
    'group', 
    'UNCLASSIFIED'
);

-- Create emergency chat room
INSERT INTO chat_rooms (id, name, description, type, classification) 
VALUES (
    '00000000-0000-0000-0000-000000000002', 
    'Emergency', 
    'Emergency communications and alerts', 
    'emergency', 
    'UNCLASSIFIED'
);

-- Create tactical coordination room
INSERT INTO chat_rooms (id, name, description, type, classification) 
VALUES (
    '00000000-0000-0000-0000-000000000003', 
    'Tactical Coordination', 
    'Mission planning and tactical coordination', 
    'tactical', 
    'RESTRICTED'
);
