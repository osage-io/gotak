-- Mission Planning Tables

-- Mission status enum
CREATE TYPE mission_status AS ENUM (
    'planning',
    'approved', 
    'active',
    'on_hold',
    'completed',
    'cancelled'
);

-- Task status enum
CREATE TYPE task_status AS ENUM (
    'pending',
    'assigned',
    'in_progress',
    'completed',
    'blocked',
    'cancelled'
);

-- Classification levels
CREATE TYPE classification AS ENUM (
    'unclassified',
    'confidential',
    'secret',
    'top_secret'
);

-- Missions table
CREATE TABLE missions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    status mission_status DEFAULT 'planning',
    priority INTEGER DEFAULT 0,
    classification classification DEFAULT 'unclassified',
    start_date TIMESTAMP WITH TIME ZONE,
    end_date TIMESTAMP WITH TIME ZONE,
    commander_id UUID REFERENCES users(id),
    created_by UUID REFERENCES users(id) NOT NULL,
    group_id VARCHAR(100),
    -- Location fields
    location_lat DOUBLE PRECISION,
    location_lon DOUBLE PRECISION,
    location_name VARCHAR(255),
    location_description TEXT,
    -- Metadata as JSONB
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for missions
CREATE INDEX idx_missions_status ON missions(status);
CREATE INDEX idx_missions_commander ON missions(commander_id);
CREATE INDEX idx_missions_group ON missions(group_id);
CREATE INDEX idx_missions_dates ON missions(start_date, end_date);

-- Mission objectives table
CREATE TABLE mission_objectives (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    mission_id UUID REFERENCES missions(id) ON DELETE CASCADE,
    description TEXT NOT NULL,
    completed BOOLEAN DEFAULT FALSE,
    priority INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_objectives_mission ON mission_objectives(mission_id);

-- Tasks table
CREATE TABLE tasks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    mission_id UUID REFERENCES missions(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    status task_status DEFAULT 'pending',
    priority INTEGER DEFAULT 0,
    assigned_to UUID REFERENCES users(id),
    estimated_hours INTEGER,
    actual_hours INTEGER,
    due_date TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for tasks
CREATE INDEX idx_tasks_mission ON tasks(mission_id);
CREATE INDEX idx_tasks_assigned ON tasks(assigned_to);
CREATE INDEX idx_tasks_status ON tasks(status);
CREATE INDEX idx_tasks_due_date ON tasks(due_date);

-- Task dependencies table
CREATE TABLE task_dependencies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    task_id UUID REFERENCES tasks(id) ON DELETE CASCADE,
    depends_on_task_id UUID REFERENCES tasks(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    UNIQUE(task_id, depends_on_task_id),
    CHECK (task_id != depends_on_task_id)
);

-- Mission participants table (for tracking who's involved)
CREATE TABLE mission_participants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    mission_id UUID REFERENCES missions(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    role VARCHAR(100),
    joined_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    UNIQUE(mission_id, user_id)
);

CREATE INDEX idx_participants_mission ON mission_participants(mission_id);
CREATE INDEX idx_participants_user ON mission_participants(user_id);

-- Mission resources table
CREATE TABLE mission_resources (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    mission_id UUID REFERENCES missions(id) ON DELETE CASCADE,
    resource_type VARCHAR(100) NOT NULL,
    resource_name VARCHAR(255) NOT NULL,
    quantity INTEGER DEFAULT 1,
    status VARCHAR(50) DEFAULT 'requested',
    notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_resources_mission ON mission_resources(mission_id);

-- Mission timeline events
CREATE TABLE mission_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    mission_id UUID REFERENCES missions(id) ON DELETE CASCADE,
    event_type VARCHAR(100) NOT NULL,
    description TEXT,
    user_id UUID REFERENCES users(id),
    event_data JSONB DEFAULT '{}',
    occurred_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_events_mission ON mission_events(mission_id);
CREATE INDEX idx_events_occurred ON mission_events(occurred_at);

-- Triggers for updated_at timestamps
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_missions_updated_at BEFORE UPDATE ON missions
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
    
CREATE TRIGGER update_objectives_updated_at BEFORE UPDATE ON mission_objectives
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
    
CREATE TRIGGER update_tasks_updated_at BEFORE UPDATE ON tasks
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
    
CREATE TRIGGER update_resources_updated_at BEFORE UPDATE ON mission_resources
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();