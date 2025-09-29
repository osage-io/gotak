#!/bin/sh
# GoTAK Production Entrypoint Script
set -e

echo "===== GoTAK Production Container Starting ====="

# Configuration
GOTAK_CONFIG_PATH=${GOTAK_CONFIG_PATH:-/app/config/production.yaml}
GOTAK_LOG_LEVEL=${GOTAK_LOG_LEVEL:-info}
POSTGRES_HOST=${POSTGRES_HOST:-postgres}
POSTGRES_PORT=${POSTGRES_PORT:-5432}
POSTGRES_DB=${POSTGRES_DB:-gotak}
POSTGRES_USER=${POSTGRES_USER:-gotak}

# Logging function
log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1"
}

error() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] ERROR: $1" >&2
}

# Validate required environment variables
validate_environment() {
    log "Validating environment configuration..."
    
    if [ -z "$POSTGRES_PASSWORD" ]; then
        error "POSTGRES_PASSWORD environment variable is required"
        exit 1
    fi
    
    if [ -z "$JWT_SECRET" ]; then
        error "JWT_SECRET environment variable is required"
        exit 1
    fi
    
    log "Environment validation passed"
}

# Wait for database to be ready
wait_for_database() {
    log "Waiting for database at $POSTGRES_HOST:$POSTGRES_PORT..."
    
    timeout=60
    while ! pg_isready -h "$POSTGRES_HOST" -p "$POSTGRES_PORT" -U "$POSTGRES_USER" -d "$POSTGRES_DB"; do
        timeout=$((timeout - 1))
        if [ $timeout -le 0 ]; then
            error "Database connection timeout"
            exit 1
        fi
        log "Database not ready, waiting... ($timeout seconds left)"
        sleep 2
    done
    
    log "Database is ready"
}

# Run database migrations
run_migrations() {
    log "Running database migrations..."
    
    # Use the migration script if available
    if [ -f "/app/migrate.sh" ]; then
        log "Using migration runner script..."
        /app/migrate.sh apply || {
            error "Migration script failed"
            exit 1
        }
    elif [ -d "/app/migrations" ] && [ "$(ls -A /app/migrations 2>/dev/null)" ]; then
        # Fallback to simple migration execution
        log "Using fallback migration execution..."
        for migration_file in /app/migrations/*.sql; do
            if [ -f "$migration_file" ]; then
                log "Applying migration: $(basename "$migration_file")"
                PGPASSWORD="$POSTGRES_PASSWORD" psql \
                    -h "$POSTGRES_HOST" \
                    -p "$POSTGRES_PORT" \
                    -U "$POSTGRES_USER" \
                    -d "$POSTGRES_DB" \
                    -f "$migration_file" \
                    -v ON_ERROR_STOP=1 || {
                    error "Migration failed: $(basename "$migration_file")"
                    exit 1
                }
            fi
        done
        log "Database migrations completed"
    else
        log "No migrations found, skipping..."
    fi
}

# Validate configuration file
validate_config() {
    log "Validating configuration and environment..."
    
    # Use comprehensive validation script if available
    if [ -f "/app/validate-config.sh" ]; then
        log "Running comprehensive configuration validation..."
        /app/validate-config.sh quick || {
            error "Configuration validation failed"
            # Run full validation to show detailed errors
            /app/validate-config.sh full
            exit 1
        }
    else
        # Fallback to basic validation
        log "Running basic configuration validation..."
        if [ ! -f "$GOTAK_CONFIG_PATH" ]; then
            error "Configuration file not found: $GOTAK_CONFIG_PATH"
            exit 1
        fi
        
        if [ ! -r "$GOTAK_CONFIG_PATH" ]; then
            error "Configuration file is not readable: $GOTAK_CONFIG_PATH"
            exit 1
        fi
        
        log "Basic configuration validation passed"
    fi
}

# Create necessary directories
setup_directories() {
    log "Setting up application directories..."
    
    # Ensure directories exist and have correct permissions
    for dir in "$GOTAK_DATA_DIR" "$GOTAK_LOG_DIR" "/app/certs"; do
        if [ ! -d "$dir" ]; then
            mkdir -p "$dir"
            log "Created directory: $dir"
        fi
    done
    
    log "Directory setup completed"
}

# Generate TLS certificates if needed
setup_certificates() {
    log "Checking TLS certificates..."
    
    CERT_DIR="/app/certs"
    CERT_FILE="$CERT_DIR/server.crt"
    KEY_FILE="$CERT_DIR/server.key"
    
    # Check if certificates exist
    if [ ! -f "$CERT_FILE" ] || [ ! -f "$KEY_FILE" ]; then
        if [ "$GOTAK_GENERATE_CERTS" = "true" ]; then
            log "Generating self-signed TLS certificates..."
            
            # Generate self-signed certificate
            openssl req -x509 -newkey rsa:4096 -keyout "$KEY_FILE" -out "$CERT_FILE" \
                -days 365 -nodes -subj "/CN=${GOTAK_SERVER_NAME:-localhost}" \
                2>/dev/null || {
                error "Failed to generate TLS certificates"
                exit 1
            }
            
            log "Self-signed certificates generated"
        else
            log "No TLS certificates found and GOTAK_GENERATE_CERTS is not enabled"
            log "TLS endpoints will not be available"
        fi
    else
        log "TLS certificates found"
    fi
}

# Health check before starting
pre_flight_check() {
    log "Running pre-flight checks..."
    
    # Check if required files exist
    if [ ! -f "/app/gotak-server" ]; then
        error "GoTAK server binary not found"
        exit 1
    fi
    
    # Check if server binary is executable
    if [ ! -x "/app/gotak-server" ]; then
        error "GoTAK server binary is not executable"
        exit 1
    fi
    
    log "Pre-flight checks passed"
}

# Signal handlers for graceful shutdown
shutdown_handler() {
    log "Received shutdown signal, stopping GoTAK server..."
    if [ -n "$GOTAK_PID" ]; then
        kill -TERM "$GOTAK_PID" 2>/dev/null || true
        wait "$GOTAK_PID" 2>/dev/null || true
    fi
    log "GoTAK server stopped"
    exit 0
}

# Set up signal handlers
trap shutdown_handler TERM INT

# Main execution
main() {
    log "Starting GoTAK production container..."
    log "Version: ${VERSION:-unknown}"
    log "Build time: ${BUILD_TIME:-unknown}"
    log "Git commit: ${GIT_COMMIT:-unknown}"
    
    # Run startup sequence
    validate_environment
    setup_directories
    
    # Only wait for database if not in standalone mode
    if [ "$GOTAK_STANDALONE" != "true" ]; then
        wait_for_database
        run_migrations
    fi
    
    validate_config
    setup_certificates
    pre_flight_check
    
    log "Starting GoTAK server with config: $GOTAK_CONFIG_PATH"
    log "Log level: $GOTAK_LOG_LEVEL"
    
    # Execute the main command with config argument
    if [ $# -gt 0 ]; then
        # Add config argument if GOTAK_CONFIG_PATH is set
        CONFIG_ARG=""
        if [ -n "$GOTAK_CONFIG_PATH" ]; then
            CONFIG_ARG="-config $GOTAK_CONFIG_PATH"
        fi
        
        log "Executing command: $* $CONFIG_ARG"
        exec "$@" $CONFIG_ARG &
        GOTAK_PID=$!
        wait $GOTAK_PID
    else
        log "No command provided, exiting..."
        exit 1
    fi
}

# Run main function
main "$@"
