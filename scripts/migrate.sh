#!/bin/bash
# GoTAK Database Migration Runner
# Handles schema migrations with apply, rollback, and status operations
set -e

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
MIGRATIONS_DIR="${MIGRATIONS_DIR:-$PROJECT_ROOT/migrations}"
ROLLBACKS_DIR="${ROLLBACKS_DIR:-$PROJECT_ROOT/migrations/rollbacks}"

# Database configuration from environment
DB_HOST="${POSTGRES_HOST:-localhost}"
DB_PORT="${POSTGRES_PORT:-5432}"
DB_NAME="${POSTGRES_DB:-gotak}"
DB_USER="${POSTGRES_USER:-gotak}"
DB_PASSWORD="${POSTGRES_PASSWORD}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Logging functions
log() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1" >&2
}

# Check if required environment variables are set
check_env() {
    if [ -z "$DB_PASSWORD" ]; then
        error "POSTGRES_PASSWORD environment variable is required"
        exit 1
    fi
    
    if [ ! -d "$MIGRATIONS_DIR" ]; then
        error "Migrations directory not found: $MIGRATIONS_DIR"
        exit 1
    fi
}

# Execute SQL command
execute_sql() {
    local sql="$1"
    local description="${2:-SQL command}"
    
    log "Executing: $description"
    
    PGPASSWORD="$DB_PASSWORD" psql \
        -h "$DB_HOST" \
        -p "$DB_PORT" \
        -U "$DB_USER" \
        -d "$DB_NAME" \
        -v ON_ERROR_STOP=1 \
        -c "$sql" || {
        error "Failed to execute: $description"
        return 1
    }
}

# Execute SQL file
execute_sql_file() {
    local file_path="$1"
    local description="${2:-$(basename "$file_path")}"
    
    if [ ! -f "$file_path" ]; then
        error "SQL file not found: $file_path"
        return 1
    fi
    
    log "Executing SQL file: $description"
    
    PGPASSWORD="$DB_PASSWORD" psql \
        -h "$DB_HOST" \
        -p "$DB_PORT" \
        -U "$DB_USER" \
        -d "$DB_NAME" \
        -v ON_ERROR_STOP=1 \
        -f "$file_path" || {
        error "Failed to execute SQL file: $file_path"
        return 1
    }
}

# Check database connection
check_connection() {
    log "Checking database connection..."
    
    if ! PGPASSWORD="$DB_PASSWORD" pg_isready \
        -h "$DB_HOST" \
        -p "$DB_PORT" \
        -U "$DB_USER" \
        -d "$DB_NAME" >/dev/null 2>&1; then
        error "Cannot connect to database at $DB_HOST:$DB_PORT"
        error "Please check your database connection settings"
        exit 1
    fi
    
    success "Database connection established"
}

# Initialize migration system
init_migration_system() {
    log "Initializing migration tracking system..."
    
    execute_sql "
        CREATE TABLE IF NOT EXISTS schema_migrations (
            version VARCHAR(255) PRIMARY KEY,
            applied_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
            description TEXT,
            checksum VARCHAR(64)
        );
        CREATE INDEX IF NOT EXISTS idx_schema_migrations_applied_at ON schema_migrations(applied_at);
    " "Create schema_migrations table"
    
    success "Migration tracking system initialized"
}

# Get applied migrations
get_applied_migrations() {
    PGPASSWORD="$DB_PASSWORD" psql \
        -h "$DB_HOST" \
        -p "$DB_PORT" \
        -U "$DB_USER" \
        -d "$DB_NAME" \
        -t -c "SELECT version FROM schema_migrations ORDER BY version;" 2>/dev/null | tr -d ' '
}

# Get available migrations
get_available_migrations() {
    find "$MIGRATIONS_DIR" -name "*.sql" -type f | sort | while read -r file; do
        basename "$file" .sql
    done
}

# Calculate file checksum
calculate_checksum() {
    local file_path="$1"
    if command -v sha256sum >/dev/null 2>&1; then
        sha256sum "$file_path" | cut -d' ' -f1
    elif command -v shasum >/dev/null 2>&1; then
        shasum -a 256 "$file_path" | cut -d' ' -f1
    else
        # Fallback to a simple hash
        md5 -q "$file_path" 2>/dev/null || echo "no-checksum"
    fi
}

# Record migration as applied
record_migration() {
    local version="$1"
    local description="$2"
    local checksum="$3"
    
    execute_sql "
        INSERT INTO schema_migrations (version, description, checksum) 
        VALUES ('$version', '$description', '$checksum')
        ON CONFLICT (version) 
        DO UPDATE SET 
            applied_at = CURRENT_TIMESTAMP,
            description = EXCLUDED.description,
            checksum = EXCLUDED.checksum;
    " "Record migration $version"
}

# Remove migration record
remove_migration_record() {
    local version="$1"
    
    execute_sql "
        DELETE FROM schema_migrations WHERE version = '$version';
    " "Remove migration record $version"
}

# Apply a single migration
apply_migration() {
    local migration_file="$1"
    local version=$(basename "$migration_file" .sql)
    
    log "Applying migration: $version"
    
    # Check if migration is already applied
    local applied_migrations
    applied_migrations=$(get_applied_migrations)
    if echo "$applied_migrations" | grep -q "^$version$"; then
        warn "Migration $version is already applied, skipping"
        return 0
    fi
    
    # Calculate checksum
    local checksum
    checksum=$(calculate_checksum "$migration_file")
    
    # Extract description from migration file (first comment line)
    local description
    description=$(head -5 "$migration_file" | grep -E "^--" | head -1 | sed 's/^-- *//' || echo "Migration $version")
    
    # Execute migration
    execute_sql_file "$migration_file" "Migration $version"
    
    # Record migration
    record_migration "$version" "$description" "$checksum"
    
    success "Applied migration: $version"
}

# Apply all pending migrations
apply_all_migrations() {
    log "Applying all pending migrations..."
    
    local applied_count=0
    local available_migrations
    available_migrations=$(get_available_migrations)
    
    if [ -z "$available_migrations" ]; then
        warn "No migration files found in $MIGRATIONS_DIR"
        return 0
    fi
    
    echo "$available_migrations" | while read -r version; do
        local migration_file="$MIGRATIONS_DIR/${version}.sql"
        if [ -f "$migration_file" ]; then
            apply_migration "$migration_file"
            applied_count=$((applied_count + 1))
        fi
    done
    
    success "Migration process completed"
}

# Rollback a specific migration
rollback_migration() {
    local version="$1"
    
    log "Rolling back migration: $version"
    
    # Check if migration is applied
    local applied_migrations
    applied_migrations=$(get_applied_migrations)
    if ! echo "$applied_migrations" | grep -q "^$version$"; then
        warn "Migration $version is not applied, nothing to rollback"
        return 0
    fi
    
    # Look for rollback script
    local rollback_file="$ROLLBACKS_DIR/${version}_rollback.sql"
    if [ ! -f "$rollback_file" ]; then
        error "Rollback script not found: $rollback_file"
        error "Cannot rollback migration $version"
        return 1
    fi
    
    # Execute rollback
    execute_sql_file "$rollback_file" "Rollback $version"
    
    # Remove migration record
    remove_migration_record "$version"
    
    success "Rolled back migration: $version"
}

# Show migration status
show_status() {
    log "Migration Status"
    echo "=================="
    
    local applied_migrations
    applied_migrations=$(get_applied_migrations)
    
    local available_migrations
    available_migrations=$(get_available_migrations)
    
    echo "Applied migrations:"
    if [ -n "$applied_migrations" ]; then
        echo "$applied_migrations" | while read -r version; do
            if [ -n "$version" ]; then
                echo "  ✓ $version"
            fi
        done
    else
        echo "  (none)"
    fi
    
    echo ""
    echo "Pending migrations:"
    local has_pending=false
    echo "$available_migrations" | while read -r version; do
        if [ -n "$version" ] && ! echo "$applied_migrations" | grep -q "^$version$"; then
            echo "  ○ $version"
            has_pending=true
        fi
    done
    
    if [ "$has_pending" = "false" ]; then
        echo "  (none)"
    fi
    
    echo ""
    echo "Database: $DB_USER@$DB_HOST:$DB_PORT/$DB_NAME"
}

# Validate migration files
validate_migrations() {
    log "Validating migration files..."
    
    local errors=0
    local available_migrations
    available_migrations=$(get_available_migrations)
    
    echo "$available_migrations" | while read -r version; do
        local migration_file="$MIGRATIONS_DIR/${version}.sql"
        
        # Check file exists
        if [ ! -f "$migration_file" ]; then
            error "Migration file not found: $migration_file"
            errors=$((errors + 1))
            continue
        fi
        
        # Check file is readable
        if [ ! -r "$migration_file" ]; then
            error "Migration file not readable: $migration_file"
            errors=$((errors + 1))
            continue
        fi
        
        # Check SQL syntax (basic check)
        if ! grep -q ";" "$migration_file"; then
            warn "Migration file may be missing SQL statements: $migration_file"
        fi
        
        log "✓ $version"
    done
    
    if [ $errors -eq 0 ]; then
        success "All migration files validated successfully"
    else
        error "$errors validation errors found"
        exit 1
    fi
}

# Create a new migration file template
create_migration() {
    local description="$1"
    
    if [ -z "$description" ]; then
        error "Migration description is required"
        echo "Usage: $0 create <description>"
        exit 1
    fi
    
    # Generate version number (timestamp-based)
    local version
    version=$(date '+%Y%m%d%H%M%S')
    
    # Clean description for filename
    local clean_description
    clean_description=$(echo "$description" | tr '[:upper:]' '[:lower:]' | sed 's/[^a-z0-9]/_/g' | sed 's/_\+/_/g' | sed 's/^_\|_$//g')
    
    local migration_file="$MIGRATIONS_DIR/${version}_${clean_description}.sql"
    local rollback_file="$ROLLBACKS_DIR/${version}_${clean_description}_rollback.sql"
    
    # Create directories if they don't exist
    mkdir -p "$MIGRATIONS_DIR"
    mkdir -p "$ROLLBACKS_DIR"
    
    # Create migration file
    cat > "$migration_file" << EOF
-- Migration: $description
-- Version: $version
-- Created: $(date '+%Y-%m-%d %H:%M:%S')

BEGIN;

-- TODO: Add your migration SQL here
-- Example:
-- CREATE TABLE example_table (
--     id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
--     name VARCHAR(255) NOT NULL,
--     created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
-- );

-- Record migration
INSERT INTO schema_migrations (version, description) 
VALUES ('${version}_${clean_description}', '$description')
ON CONFLICT (version) DO NOTHING;

COMMIT;
EOF

    # Create rollback file
    cat > "$rollback_file" << EOF
-- Rollback for: $description
-- Version: $version
-- Created: $(date '+%Y-%m-%d %H:%M:%S')

BEGIN;

-- TODO: Add your rollback SQL here
-- This should undo the changes made in the migration
-- Example:
-- DROP TABLE IF EXISTS example_table;

COMMIT;
EOF

    success "Created migration files:"
    echo "  Migration: $migration_file"
    echo "  Rollback:  $rollback_file"
    echo ""
    echo "Please edit these files to add your SQL statements."
}

# Main command handler
main() {
    local command="$1"
    shift
    
    case "$command" in
        "apply")
            check_env
            check_connection
            init_migration_system
            if [ $# -gt 0 ]; then
                # Apply specific migration
                local version="$1"
                local migration_file="$MIGRATIONS_DIR/${version}.sql"
                apply_migration "$migration_file"
            else
                # Apply all pending migrations
                apply_all_migrations
            fi
            ;;
        "rollback")
            check_env
            check_connection
            if [ $# -eq 0 ]; then
                error "Migration version is required for rollback"
                echo "Usage: $0 rollback <version>"
                exit 1
            fi
            rollback_migration "$1"
            ;;
        "status")
            check_env
            check_connection
            init_migration_system
            show_status
            ;;
        "validate")
            check_env
            validate_migrations
            ;;
        "create")
            create_migration "$*"
            ;;
        "init")
            check_env
            check_connection
            init_migration_system
            ;;
        "help"|*)
            cat << EOF
GoTAK Database Migration Tool

Usage: $0 <command> [options]

Commands:
  apply [version]     Apply all pending migrations or specific migration
  rollback <version>  Rollback a specific migration
  status             Show migration status
  validate           Validate migration files
  create <desc>      Create new migration template files
  init               Initialize migration tracking system
  help               Show this help message

Environment Variables:
  POSTGRES_HOST      Database host (default: localhost)
  POSTGRES_PORT      Database port (default: 5432)
  POSTGRES_DB        Database name (default: gotak)
  POSTGRES_USER      Database user (default: gotak)
  POSTGRES_PASSWORD  Database password (required)
  MIGRATIONS_DIR     Migrations directory (default: ./migrations)

Examples:
  $0 apply                    # Apply all pending migrations
  $0 apply 001_initial        # Apply specific migration
  $0 rollback 001_initial     # Rollback specific migration
  $0 status                   # Show migration status
  $0 create "add users table" # Create new migration files
  $0 validate                 # Validate all migration files

EOF
            ;;
    esac
}

# Run main function
main "$@"
