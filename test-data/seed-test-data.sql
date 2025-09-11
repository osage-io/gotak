-- Test Data Seeding for GoTAK Integration Tests
-- This file seeds the database with test data for comprehensive integration testing

-- Test Users
INSERT INTO users (id, username, email, password_hash, role, group_id, created_at, updated_at) VALUES
('550e8400-e29b-41d4-a716-446655440001', 'test_admin', 'admin@test.gotak', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'admin', 'test_group', NOW(), NOW()),
('550e8400-e29b-41d4-a716-446655440002', 'test_user1', 'user1@test.gotak', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'user', 'test_group', NOW(), NOW()),
('550e8400-e29b-41d4-a716-446655440003', 'test_user2', 'user2@test.gotak', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'user', 'test_group', NOW(), NOW()),
('550e8400-e29b-41d4-a716-446655440004', 'test_operator', 'operator@test.gotak', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'operator', 'test_group', NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

-- Test Routes
INSERT INTO routes (id, name, description, created_by, group_id, geometry, distance, duration, route_type, vehicle, optimize, created_at, updated_at) VALUES
('550e8400-e29b-41d4-a716-446655440100', 'Test Route Alpha', 'Primary test route for integration testing', '550e8400-e29b-41d4-a716-446655440001', 'test_group', 
 '{"type":"LineString","coordinates":[[-76.6413,39.0458],[-76.6423,39.0468],[-76.6433,39.0478]]}', 
 2247.5, 300000000000, 'fastest', 'car', true, NOW(), NOW()),
('550e8400-e29b-41d4-a716-446655440101', 'Test Route Bravo', 'Secondary test route for parallel testing', '550e8400-e29b-41d4-a716-446655440002', 'test_group',
 '{"type":"LineString","coordinates":[[-76.6400,39.0450],[-76.6410,39.0460],[-76.6420,39.0470]]}', 
 1834.2, 250000000000, 'shortest', 'foot', false, NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

-- Test Waypoints
INSERT INTO waypoints (id, route_id, sequence, lat, lng, name, description, created_at) VALUES
('550e8400-e29b-41d4-a716-446655440200', '550e8400-e29b-41d4-a716-446655440100', 0, 39.0458, -76.6413, 'Start Point Alpha', 'Starting location for test route', NOW()),
('550e8400-e29b-41d4-a716-446655440201', '550e8400-e29b-41d4-a716-446655440100', 1, 39.0468, -76.6423, 'Waypoint Alpha 1', 'First intermediate waypoint', NOW()),
('550e8400-e29b-41d4-a716-446655440202', '550e8400-e29b-41d4-a716-446655440100', 2, 39.0478, -76.6433, 'End Point Alpha', 'Destination for test route', NOW()),
('550e8400-e29b-41d4-a716-446655440203', '550e8400-e29b-41d4-a716-446655440101', 0, 39.0450, -76.6400, 'Start Point Bravo', 'Starting location for secondary route', NOW()),
('550e8400-e29b-41d4-a716-446655440204', '550e8400-e29b-41d4-a716-446655440101', 1, 39.0460, -76.6410, 'Waypoint Bravo 1', 'Intermediate point', NOW()),
('550e8400-e29b-41d4-a716-446655440205', '550e8400-e29b-41d4-a716-446655440101', 2, 39.0470, -76.6420, 'End Point Bravo', 'End location', NOW())
ON CONFLICT (id) DO NOTHING;

-- Test Geofences
INSERT INTO geofences (id, name, description, type, geometry, enabled, alert_on_enter, alert_on_exit, created_by, group_id, created_at, updated_at) VALUES
('550e8400-e29b-41d4-a716-446655440300', 'Test Perimeter Alpha', 'Primary security perimeter for testing', 'circle',
 '{"center":{"lat":39.0458,"lng":-76.6413},"radius":1000}', true, true, true, 
 '550e8400-e29b-41d4-a716-446655440001', 'test_group', NOW(), NOW()),
('550e8400-e29b-41d4-a716-446655440301', 'Test Zone Bravo', 'Secondary monitoring zone', 'rectangle',
 '{"bounds":{"north":39.0480,"south":39.0440,"east":-76.6390,"west":-76.6430}}', true, true, false,
 '550e8400-e29b-41d4-a716-446655440001', 'test_group', NOW(), NOW()),
('550e8400-e29b-41d4-a716-446655440302', 'Test Polygon Charlie', 'Polygon geofence for complex shape testing', 'polygon',
 '{"coordinates":[[[39.0450,-76.6400],[39.0460,-76.6400],[39.0460,-76.6410],[39.0450,-76.6410],[39.0450,-76.6400]]]}', false, true, true,
 '550e8400-e29b-41d4-a716-446655440002', 'test_group', NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

-- Test Geofence Violations (historical data)
INSERT INTO geofence_violations (id, geofence_id, entity_id, violation_type, position, timestamp, acknowledged, acknowledged_by) VALUES
('550e8400-e29b-41d4-a716-446655440400', '550e8400-e29b-41d4-a716-446655440300', 'test-entity-001', 'enter', POINT(-76.6413, 39.0458), NOW() - INTERVAL '1 hour', false, NULL),
('550e8400-e29b-41d4-a716-446655440401', '550e8400-e29b-41d4-a716-446655440300', 'test-entity-001', 'exit', POINT(-76.6420, 39.0465), NOW() - INTERVAL '30 minutes', true, '550e8400-e29b-41d4-a716-446655440001'),
('550e8400-e29b-41d4-a716-446655440402', '550e8400-e29b-41d4-a716-446655440301', 'test-entity-002', 'enter', POINT(-76.6410, 39.0460), NOW() - INTERVAL '15 minutes', false, NULL)
ON CONFLICT (id) DO NOTHING;

-- Test Offline Areas
INSERT INTO offline_areas (id, name, bounds, min_zoom, max_zoom, layers, status, progress, size_mb, created_at, updated_at) VALUES
('550e8400-e29b-41d4-a716-446655440500', 'Test Area Alpha', 
 '{"north":39.0480,"south":39.0440,"east":-76.6390,"west":-76.6430}', 
 10, 14, '["satellite","streets"]', 'complete', 100.0, 45.2, NOW() - INTERVAL '2 hours', NOW() - INTERVAL '1 hour'),
('550e8400-e29b-41d4-a716-446655440501', 'Test Area Bravo', 
 '{"north":39.0500,"south":39.0460,"east":-76.6370,"west":-76.6410}', 
 8, 16, '["satellite","terrain","streets"]', 'downloading', 65.3, 127.8, NOW() - INTERVAL '30 minutes', NOW()),
('550e8400-e29b-41d4-a716-446655440502', 'Test Area Charlie', 
 '{"north":39.0520,"south":39.0480,"east":-76.6350,"west":-76.6390}', 
 12, 15, '["streets"]', 'pending', 0.0, 0.0, NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

-- Test Tactical Overlays
INSERT INTO tactical_overlays (id, name, type, geometry, style, metadata, created_by, group_id, created_at, updated_at) VALUES
('550e8400-e29b-41d4-a716-446655440600', 'Test Overlay Alpha', 'marker', 
 '{"type":"Point","coordinates":[-76.6413,39.0458]}',
 '{"color":"red","size":"large","icon":"warning"}',
 '{"priority":"high","classification":"unclassified","notes":"Test tactical marker"}',
 '550e8400-e29b-41d4-a716-446655440001', 'test_group', NOW(), NOW()),
('550e8400-e29b-41d4-a716-446655440601', 'Test Line Bravo', 'line',
 '{"type":"LineString","coordinates":[[-76.6400,39.0450],[-76.6420,39.0470]]}',
 '{"color":"blue","width":"3","dash":"solid"}',
 '{"type":"boundary","classification":"restricted","notes":"Test boundary line"}',
 '550e8400-e29b-41d4-a716-446655440002', 'test_group', NOW(), NOW()),
('550e8400-e29b-41d4-a716-446655440602', 'Test Area Charlie', 'area',
 '{"type":"Polygon","coordinates":[[[-76.6430,39.0440],[-76.6410,39.0440],[-76.6410,39.0460],[-76.6430,39.0460],[-76.6430,39.0440]]]}',
 '{"fillColor":"yellow","fillOpacity":"0.3","strokeColor":"orange","strokeWidth":"2"}',
 '{"zone":"restricted","access":"authorized-only","notes":"Test restricted area"}',
 '550e8400-e29b-41d4-a716-446655440003', 'test_group', NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

-- Test Missions (if missions table exists)
INSERT INTO missions (id, name, description, status, classification, priority, start_time, end_time, created_by, group_id, created_at, updated_at) VALUES
('550e8400-e29b-41d4-a716-446655440700', 'Test Mission Alpha', 'Primary integration test mission', 'active', 'UNCLASSIFIED', 'high', 
 NOW() - INTERVAL '1 day', NOW() + INTERVAL '1 day', '550e8400-e29b-41d4-a716-446655440001', 'test_group', NOW(), NOW()),
('550e8400-e29b-41d4-a716-446655440701', 'Test Mission Bravo', 'Secondary test mission for parallel execution', 'planning', 'RESTRICTED', 'medium',
 NOW() + INTERVAL '1 hour', NOW() + INTERVAL '2 days', '550e8400-e29b-41d4-a716-446655440002', 'test_group', NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

-- Commit the transaction
COMMIT;

-- Display seeding summary
SELECT 'Test data seeding completed successfully' AS status;
