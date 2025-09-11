#!/bin/bash
# Test data setup script for GoTAK integration tests
set -e

echo "Setting up test data for GoTAK integration tests..."

# Wait for database to be ready
echo "Waiting for database to be ready..."
until pg_isready -h "${POSTGRES_HOST:-localhost}" -p "${POSTGRES_PORT:-5432}" -U "${POSTGRES_USER:-gotak}"; do
    echo "Database not ready, waiting..."
    sleep 2
done

echo "Database is ready. Running migrations..."

# Run database migrations if they exist
if [ -f "/app/migrations/schema.sql" ]; then
    echo "Running schema migrations..."
    PGPASSWORD="${POSTGRES_PASSWORD}" psql \
        -h "${POSTGRES_HOST:-localhost}" \
        -p "${POSTGRES_PORT:-5432}" \
        -U "${POSTGRES_USER:-gotak}" \
        -d "${POSTGRES_DB:-gotak_test}" \
        -f /app/migrations/schema.sql || echo "Schema migration failed or already applied"
fi

# Seed test data
echo "Seeding test data..."
PGPASSWORD="${POSTGRES_PASSWORD}" psql \
    -h "${POSTGRES_HOST:-localhost}" \
    -p "${POSTGRES_PORT:-5432}" \
    -U "${POSTGRES_USER:-gotak}" \
    -d "${POSTGRES_DB:-gotak_test}" \
    -f /app/test-data/seed-test-data.sql

echo "Test data setup completed successfully!"

# Display test data summary
echo "Test data summary:"
PGPASSWORD="${POSTGRES_PASSWORD}" psql \
    -h "${POSTGRES_HOST:-localhost}" \
    -p "${POSTGRES_PORT:-5432}" \
    -U "${POSTGRES_USER:-gotak}" \
    -d "${POSTGRES_DB:-gotak_test}" \
    -c "
SELECT 
    'users' as table_name, count(*) as record_count FROM users WHERE group_id = 'test_group'
UNION ALL
SELECT 
    'routes', count(*) FROM routes WHERE group_id = 'test_group'
UNION ALL
SELECT 
    'geofences', count(*) FROM geofences WHERE group_id = 'test_group'
UNION ALL
SELECT 
    'offline_areas', count(*) FROM offline_areas
UNION ALL
SELECT 
    'tactical_overlays', count(*) FROM tactical_overlays WHERE group_id = 'test_group'
ORDER BY table_name;
"

echo "Test data setup is complete and ready for integration testing!"
