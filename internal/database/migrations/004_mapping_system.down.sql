-- Down migration for Mapping System
-- Removes all mapping-related tables and indexes

-- Drop indexes first (order doesn't matter for indexes)
DROP INDEX IF EXISTS idx_tactical_overlays_expiration;
DROP INDEX IF EXISTS idx_tactical_overlays_tags;
DROP INDEX IF EXISTS idx_tactical_overlays_created_at;
DROP INDEX IF EXISTS idx_tactical_overlays_priority;
DROP INDEX IF EXISTS idx_tactical_overlays_classification;
DROP INDEX IF EXISTS idx_tactical_overlays_visible;
DROP INDEX IF EXISTS idx_tactical_overlays_type;
DROP INDEX IF EXISTS idx_tactical_overlays_group_id;

DROP INDEX IF EXISTS idx_offline_areas_bounds;
DROP INDEX IF EXISTS idx_offline_areas_created_at;
DROP INDEX IF EXISTS idx_offline_areas_status;

DROP INDEX IF EXISTS idx_violations_type;
DROP INDEX IF EXISTS idx_violations_acknowledged;
DROP INDEX IF EXISTS idx_violations_timestamp;
DROP INDEX IF EXISTS idx_violations_entity_id;
DROP INDEX IF EXISTS idx_violations_geofence_id;

DROP INDEX IF EXISTS idx_geofences_created_at;
DROP INDEX IF EXISTS idx_geofences_type;
DROP INDEX IF EXISTS idx_geofences_enabled;
DROP INDEX IF EXISTS idx_geofences_group_id;

DROP INDEX IF EXISTS idx_waypoints_location;
DROP INDEX IF EXISTS idx_waypoints_sequence;
DROP INDEX IF EXISTS idx_waypoints_route_id;

DROP INDEX IF EXISTS idx_routes_route_type;
DROP INDEX IF EXISTS idx_routes_created_at;
DROP INDEX IF EXISTS idx_routes_created_by;
DROP INDEX IF EXISTS idx_routes_group_id;

-- Drop tables in reverse dependency order
DROP TABLE IF EXISTS tactical_overlays;
DROP TABLE IF EXISTS offline_areas;
DROP TABLE IF EXISTS geofence_violations;
DROP TABLE IF EXISTS geofences;
DROP TABLE IF EXISTS waypoints;
DROP TABLE IF EXISTS routes;
