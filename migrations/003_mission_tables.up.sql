-- Migration: Add mission planning system tables
-- Description: Creates all tables needed for mission planning, task management, and timeline tracking

-- Missions table
CREATE TABLE missions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    status VARCHAR(50) DEFAULT 'planning' CHECK (status IN ('planning', 'approved', 'active', 'on_hold', 'completed', 'cancelled')),
    priority INTEGER DEFAULT 3 CHECK (priority BETWEEN 1 AND 5),
    classification VARCHAR(50) DEFAULT 'RESTRICTED' CHECK (classification IN ('UNCLASSIFIED', 'RESTRICTED', 'CONFIDENTIAL', 'SECRET', 'TOP_SECRET')),
    start_date TIMESTAMP,
    end_date TIMESTAMP,
    commander_id UUID REFERENCES users(id),
    created_by UUID REFERENCES users(id) NOT NULL,
    group_id VARCHAR(255) NOT NULL,
    latitude DOUBLE PRECISION,
    longitude DOUBLE PRECISION,
    location_name VARCHAR(255),
    location_description TEXT,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Mission objectives table
CREATE TABLE mission_objectives (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    mission_id UUID REFERENCES missions(id) ON DELETE CASCADE,
    description TEXT NOT NULL,
    priority INTEGER DEFAULT 3 CHECK (priority BETWEEN 1 AND 5),
    completed BOOLEAN DEFAULT false,
    completed_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Tasks table
CREATE TABLE tasks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    mission_id UUID REFERENCES missions(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    status VARCHAR(50) DEFAULT 'pending' CHECK (status IN ('pending', 'assigned', 'in_progress', 'completed', 'blocked', 'cancelled')),
    priority INTEGER DEFAULT 3 CHECK (priority BETWEEN 1 AND 5),
    assigned_to UUID REFERENCES users(id),
    estimated_hours INTEGER DEFAULT 0 CHECK (estimated_hours >= 0),
    actual_hours INTEGER DEFAULT 0 CHECK (actual_hours >= 0),
    due_date TIMESTAMP,
    completed_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Task dependencies table
CREATE TABLE task_dependencies (
    task_id UUID REFERENCES tasks(id) ON DELETE CASCADE,
    depends_on_task_id UUID REFERENCES tasks(id) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT NOW(),
    PRIMARY KEY (task_id, depends_on_task_id),
    CONSTRAINT no_self_dependency CHECK (task_id != depends_on_task_id)
);

-- Mission status history table
CREATE TABLE mission_status_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    mission_id UUID REFERENCES missions(id) ON DELETE CASCADE,
    old_status VARCHAR(50),
    new_status VARCHAR(50) NOT NULL,
    changed_by UUID REFERENCES users(id) NOT NULL,
    reason TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Mission milestones table
CREATE TABLE mission_milestones (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    mission_id UUID REFERENCES missions(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    milestone_date TIMESTAMP NOT NULL,
    completed BOOLEAN DEFAULT false,
    completed_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Resource requests for missions table
CREATE TABLE mission_resource_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    mission_id UUID REFERENCES missions(id) ON DELETE CASCADE,
    resource_type VARCHAR(100) NOT NULL CHECK (resource_type IN ('personnel', 'equipment', 'supply', 'transport')),
    resource_id VARCHAR(255),
    resource_name VARCHAR(255) NOT NULL,
    quantity INTEGER DEFAULT 1 CHECK (quantity > 0),
    required_date TIMESTAMP,
    status VARCHAR(50) DEFAULT 'requested' CHECK (status IN ('requested', 'approved', 'allocated', 'denied', 'cancelled')),
    requested_by UUID REFERENCES users(id) NOT NULL,
    approved_by UUID REFERENCES users(id),
    notes TEXT,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Performance indexes
CREATE INDEX idx_missions_status ON missions(status);
CREATE INDEX idx_missions_commander ON missions(commander_id);
CREATE INDEX idx_missions_created_by ON missions(created_by);
CREATE INDEX idx_missions_group ON missions(group_id);
CREATE INDEX idx_missions_dates ON missions(start_date, end_date);
CREATE INDEX idx_missions_classification ON missions(classification);
CREATE INDEX idx_missions_priority ON missions(priority);

CREATE INDEX idx_mission_objectives_mission ON mission_objectives(mission_id);
CREATE INDEX idx_mission_objectives_priority ON mission_objectives(priority);
CREATE INDEX idx_mission_objectives_completed ON mission_objectives(completed);

CREATE INDEX idx_tasks_mission ON tasks(mission_id);
CREATE INDEX idx_tasks_assigned_to ON tasks(assigned_to);
CREATE INDEX idx_tasks_status ON tasks(status);
CREATE INDEX idx_tasks_priority ON tasks(priority);
CREATE INDEX idx_tasks_due_date ON tasks(due_date);
CREATE INDEX idx_tasks_created_at ON tasks(created_at);

CREATE INDEX idx_task_dependencies_task ON task_dependencies(task_id);
CREATE INDEX idx_task_dependencies_depends ON task_dependencies(depends_on_task_id);

CREATE INDEX idx_mission_status_history_mission ON mission_status_history(mission_id);
CREATE INDEX idx_mission_status_history_changed_by ON mission_status_history(changed_by);
CREATE INDEX idx_mission_status_history_created_at ON mission_status_history(created_at);

CREATE INDEX idx_mission_milestones_mission ON mission_milestones(mission_id);
CREATE INDEX idx_mission_milestones_date ON mission_milestones(milestone_date);
CREATE INDEX idx_mission_milestones_completed ON mission_milestones(completed);

CREATE INDEX idx_mission_resources_mission ON mission_resource_requests(mission_id);
CREATE INDEX idx_mission_resources_status ON mission_resource_requests(status);
CREATE INDEX idx_mission_resources_type ON mission_resource_requests(resource_type);
CREATE INDEX idx_mission_resources_requested_by ON mission_resource_requests(requested_by);
CREATE INDEX idx_mission_resources_required_date ON mission_resource_requests(required_date);

-- Function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Triggers for updated_at columns
CREATE TRIGGER update_missions_updated_at 
    BEFORE UPDATE ON missions 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_tasks_updated_at 
    BEFORE UPDATE ON tasks 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_mission_resource_requests_updated_at 
    BEFORE UPDATE ON mission_resource_requests 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Insert default mission templates data (optional - for demonstration)
INSERT INTO missions (name, description, status, priority, classification, created_by, group_id) VALUES
('Training Exercise Template', 'Basic training exercise template for unit preparation', 'planning', 2, 'UNCLASSIFIED', 
 (SELECT id FROM users WHERE username = 'admin' LIMIT 1), 'default')
ON CONFLICT DO NOTHING;

-- Comments for documentation
COMMENT ON TABLE missions IS 'Core missions table storing mission planning and tracking data';
COMMENT ON TABLE mission_objectives IS 'Mission objectives and goals tracking';
COMMENT ON TABLE tasks IS 'Individual tasks within missions with assignment and progress tracking';
COMMENT ON TABLE task_dependencies IS 'Task dependency relationships for project management';
COMMENT ON TABLE mission_status_history IS 'Audit trail for mission status changes';
COMMENT ON TABLE mission_milestones IS 'Mission milestones and key dates';
COMMENT ON TABLE mission_resource_requests IS 'Resource allocation requests for missions';

COMMENT ON COLUMN missions.classification IS 'Security classification level (UNCLASSIFIED, RESTRICTED, CONFIDENTIAL, SECRET, TOP_SECRET)';
COMMENT ON COLUMN missions.priority IS 'Mission priority (1=Highest, 5=Lowest)';
COMMENT ON COLUMN missions.status IS 'Current mission status in workflow';
COMMENT ON COLUMN missions.metadata IS 'Additional mission metadata in JSON format';

COMMENT ON COLUMN tasks.status IS 'Current task status in workflow';
COMMENT ON COLUMN tasks.priority IS 'Task priority (1=Highest, 5=Lowest)';
COMMENT ON COLUMN tasks.estimated_hours IS 'Estimated time to complete task in hours';
COMMENT ON COLUMN tasks.actual_hours IS 'Actual time spent on task in hours';
