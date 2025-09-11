-- Migration: Remove mission planning system tables
-- Description: Drops all tables and functions created in the up migration

-- Drop triggers first
DROP TRIGGER IF EXISTS update_missions_updated_at ON missions;
DROP TRIGGER IF EXISTS update_tasks_updated_at ON tasks;
DROP TRIGGER IF EXISTS update_mission_resource_requests_updated_at ON mission_resource_requests;

-- Drop function
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop indexes (will be automatically dropped with tables, but explicit for clarity)
DROP INDEX IF EXISTS idx_missions_status;
DROP INDEX IF EXISTS idx_missions_commander;
DROP INDEX IF EXISTS idx_missions_created_by;
DROP INDEX IF EXISTS idx_missions_group;
DROP INDEX IF EXISTS idx_missions_dates;
DROP INDEX IF EXISTS idx_missions_classification;
DROP INDEX IF EXISTS idx_missions_priority;

DROP INDEX IF EXISTS idx_mission_objectives_mission;
DROP INDEX IF EXISTS idx_mission_objectives_priority;
DROP INDEX IF EXISTS idx_mission_objectives_completed;

DROP INDEX IF EXISTS idx_tasks_mission;
DROP INDEX IF EXISTS idx_tasks_assigned_to;
DROP INDEX IF EXISTS idx_tasks_status;
DROP INDEX IF EXISTS idx_tasks_priority;
DROP INDEX IF EXISTS idx_tasks_due_date;
DROP INDEX IF EXISTS idx_tasks_created_at;

DROP INDEX IF EXISTS idx_task_dependencies_task;
DROP INDEX IF EXISTS idx_task_dependencies_depends;

DROP INDEX IF EXISTS idx_mission_status_history_mission;
DROP INDEX IF EXISTS idx_mission_status_history_changed_by;
DROP INDEX IF EXISTS idx_mission_status_history_created_at;

DROP INDEX IF EXISTS idx_mission_milestones_mission;
DROP INDEX IF EXISTS idx_mission_milestones_date;
DROP INDEX IF EXISTS idx_mission_milestones_completed;

DROP INDEX IF EXISTS idx_mission_resources_mission;
DROP INDEX IF EXISTS idx_mission_resources_status;
DROP INDEX IF EXISTS idx_mission_resources_type;
DROP INDEX IF EXISTS idx_mission_resources_requested_by;
DROP INDEX IF EXISTS idx_mission_resources_required_date;

-- Drop tables in reverse dependency order
DROP TABLE IF EXISTS mission_resource_requests;
DROP TABLE IF EXISTS mission_milestones;
DROP TABLE IF EXISTS mission_status_history;
DROP TABLE IF EXISTS task_dependencies;
DROP TABLE IF EXISTS tasks;
DROP TABLE IF EXISTS mission_objectives;
DROP TABLE IF EXISTS missions;
