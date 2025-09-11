#!/bin/bash
# GoTAK Database Backup and Restore Script
# Handles database backups with compression, retention, and S3 integration
set -e

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
BACKUP_DIR="${BACKUP_DIR:-$PROJECT_ROOT/backups}"
BACKUP_RETENTION_DAYS="${BACKUP_RETENTION_DAYS:-30}"

# Database configuration from environment
DB_HOST="${POSTGRES_HOST:-localhost}"
DB_PORT="${POSTGRES_PORT:-5432}"
DB_NAME="${POSTGRES_DB:-gotak}"
DB_USER="${POSTGRES_USER:-gotak}"
DB_PASSWORD="${POSTGRES_PASSWORD}"

# S3 Configuration (optional)
S3_BUCKET="${BACKUP_S3_BUCKET}"
S3_REGION="${BACKUP_S3_REGION:-us-west-2}"
S3_ACCESS_KEY="${BACKUP_S3_ACCESS_KEY}"
S3_SECRET_KEY="${BACKUP_S3_SECRET_KEY}"

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

# Check if S3 tools are available
check_s3_tools() {
    if [ -n "$S3_BUCKET" ]; then
        if ! command -v aws >/dev/null 2>&1; then
            error "AWS CLI is required for S3 integration but not found"
            error "Install AWS CLI or disable S3 integration"
            exit 1
        fi
        
        # Configure AWS credentials if provided
        if [ -n "$S3_ACCESS_KEY" ] && [ -n "$S3_SECRET_KEY" ]; then
            export AWS_ACCESS_KEY_ID="$S3_ACCESS_KEY"
            export AWS_SECRET_ACCESS_KEY="$S3_SECRET_KEY"
            export AWS_DEFAULT_REGION="$S3_REGION"
        fi
    fi
}

# Generate backup filename
generate_backup_filename() {
    local backup_type="${1:-full}"
    local timestamp=$(date '+%Y%m%d_%H%M%S')
    echo "${DB_NAME}_${backup_type}_${timestamp}.sql.gz"
}

# Create backup directory
ensure_backup_dir() {
    mkdir -p "$BACKUP_DIR"
    log "Using backup directory: $BACKUP_DIR"
}

# Perform database backup
create_backup() {
    local backup_type="${1:-full}"
    local backup_filename
    backup_filename=$(generate_backup_filename "$backup_type")
    local backup_path="$BACKUP_DIR/$backup_filename"
    
    log "Creating $backup_type backup: $backup_filename"
    
    case "$backup_type" in
        "full")
            # Full database backup with all data
            PGPASSWORD="$DB_PASSWORD" pg_dump \
                -h "$DB_HOST" \
                -p "$DB_PORT" \
                -U "$DB_USER" \
                -d "$DB_NAME" \
                --verbose \
                --format=custom \
                --compress=9 \
                --no-password \
                --file="$backup_path.tmp"
            ;;
        "schema")
            # Schema-only backup (structure without data)
            PGPASSWORD="$DB_PASSWORD" pg_dump \
                -h "$DB_HOST" \
                -p "$DB_PORT" \
                -U "$DB_USER" \
                -d "$DB_NAME" \
                --verbose \
                --schema-only \
                --format=custom \
                --compress=9 \
                --no-password \
                --file="$backup_path.tmp"
            ;;
        "data")
            # Data-only backup (data without structure)
            PGPASSWORD="$DB_PASSWORD" pg_dump \
                -h "$DB_HOST" \
                -p "$DB_PORT" \
                -U "$DB_USER" \
                -d "$DB_NAME" \
                --verbose \
                --data-only \
                --format=custom \
                --compress=9 \
                --no-password \
                --file="$backup_path.tmp"
            ;;
        *)
            error "Unknown backup type: $backup_type"
            return 1
            ;;
    esac
    
    # Move temporary file to final location
    mv "$backup_path.tmp" "${backup_path%.gz}"
    
    # Compress with gzip if not already compressed by pg_dump
    if [[ "$backup_path" == *.gz ]]; then
        gzip "${backup_path%.gz}"
    fi
    
    # Verify backup was created
    if [ ! -f "$backup_path" ]; then
        error "Backup file was not created: $backup_path"
        return 1
    fi
    
    local backup_size
    backup_size=$(du -h "$backup_path" | cut -f1)
    success "Backup created: $backup_filename ($backup_size)"
    
    # Upload to S3 if configured
    if [ -n "$S3_BUCKET" ]; then
        upload_to_s3 "$backup_path" "$backup_filename"
    fi
    
    echo "$backup_path"
}

# Upload backup to S3
upload_to_s3() {
    local local_path="$1"
    local filename="$2"
    local s3_path="s3://$S3_BUCKET/gotak-backups/$filename"
    
    log "Uploading backup to S3: $s3_path"
    
    if aws s3 cp "$local_path" "$s3_path" --region "$S3_REGION"; then
        success "Backup uploaded to S3: $s3_path"
        
        # Add metadata
        aws s3api put-object-tagging \
            --bucket "$S3_BUCKET" \
            --key "gotak-backups/$filename" \
            --tagging "TagSet=[{Key=Application,Value=GoTAK},{Key=BackupDate,Value=$(date '+%Y-%m-%d')},{Key=Environment,Value=production}]" \
            --region "$S3_REGION" || warn "Failed to add S3 object tags"
    else
        error "Failed to upload backup to S3"
        return 1
    fi
}

# Download backup from S3
download_from_s3() {
    local filename="$1"
    local local_path="$BACKUP_DIR/$filename"
    local s3_path="s3://$S3_BUCKET/gotak-backups/$filename"
    
    log "Downloading backup from S3: $s3_path"
    
    if aws s3 cp "$s3_path" "$local_path" --region "$S3_REGION"; then
        success "Backup downloaded from S3: $local_path"
        echo "$local_path"
    else
        error "Failed to download backup from S3"
        return 1
    fi
}

# Restore database from backup
restore_backup() {
    local backup_path="$1"
    local restore_options="${2:-}"
    
    if [ ! -f "$backup_path" ]; then
        error "Backup file not found: $backup_path"
        exit 1
    fi
    
    log "Restoring database from: $backup_path"
    
    # Check if backup is compressed
    local restore_file="$backup_path"
    if [[ "$backup_path" == *.gz ]]; then
        log "Decompressing backup file..."
        restore_file="${backup_path%.gz}"
        gunzip -c "$backup_path" > "$restore_file"
    fi
    
    # Determine restore command based on file format
    if file "$restore_file" | grep -q "PostgreSQL custom database dump"; then
        # Custom format backup - use pg_restore
        log "Restoring custom format backup..."
        
        PGPASSWORD="$DB_PASSWORD" pg_restore \
            -h "$DB_HOST" \
            -p "$DB_PORT" \
            -U "$DB_USER" \
            -d "$DB_NAME" \
            --verbose \
            --clean \
            --if-exists \
            --no-owner \
            --no-privileges \
            $restore_options \
            "$restore_file"
    else
        # Plain SQL format - use psql
        log "Restoring SQL format backup..."
        
        PGPASSWORD="$DB_PASSWORD" psql \
            -h "$DB_HOST" \
            -p "$DB_PORT" \
            -U "$DB_USER" \
            -d "$DB_NAME" \
            -v ON_ERROR_STOP=1 \
            -f "$restore_file"
    fi
    
    # Clean up temporary decompressed file
    if [[ "$backup_path" == *.gz ]] && [ -f "$restore_file" ]; then
        rm "$restore_file"
    fi
    
    success "Database restored successfully"
}

# List available backups
list_backups() {
    local location="${1:-local}"
    
    case "$location" in
        "local")
            log "Local backups in $BACKUP_DIR:"
            if [ -d "$BACKUP_DIR" ] && [ "$(ls -A "$BACKUP_DIR" 2>/dev/null)" ]; then
                find "$BACKUP_DIR" -name "*.sql*" -type f -printf "%T@ %Tc %s %p\n" | sort -n | while read -r timestamp date size path; do
                    local filename=$(basename "$path")
                    local readable_size=$(numfmt --to=iec --suffix=B $size)
                    printf "  %-40s %s (%s)\n" "$filename" "$date" "$readable_size"
                done
            else
                echo "  (no local backups found)"
            fi
            ;;
        "s3")
            if [ -n "$S3_BUCKET" ]; then
                log "S3 backups in s3://$S3_BUCKET/gotak-backups/:"
                aws s3 ls "s3://$S3_BUCKET/gotak-backups/" --region "$S3_REGION" | while read -r date time size filename; do
                    if [ -n "$filename" ]; then
                        printf "  %-40s %s %s (%s)\n" "$filename" "$date" "$time" "$size"
                    fi
                done
            else
                warn "S3 backup not configured"
            fi
            ;;
        "all")
            list_backups "local"
            echo ""
            list_backups "s3"
            ;;
        *)
            error "Unknown location: $location"
            return 1
            ;;
    esac
}

# Clean up old backups
cleanup_backups() {
    local location="${1:-local}"
    local retention_days="${2:-$BACKUP_RETENTION_DAYS}"
    
    log "Cleaning up backups older than $retention_days days..."
    
    case "$location" in
        "local")
            if [ -d "$BACKUP_DIR" ]; then
                local deleted_count=0
                find "$BACKUP_DIR" -name "*.sql*" -type f -mtime "+$retention_days" | while read -r file; do
                    log "Deleting old backup: $(basename "$file")"
                    rm "$file"
                    deleted_count=$((deleted_count + 1))
                done
                success "Cleaned up local backups"
            fi
            ;;
        "s3")
            if [ -n "$S3_BUCKET" ]; then
                # S3 lifecycle policies are preferred for S3 cleanup
                # This is a manual implementation
                local cutoff_date=$(date -d "$retention_days days ago" '+%Y-%m-%d')
                aws s3api list-objects-v2 \
                    --bucket "$S3_BUCKET" \
                    --prefix "gotak-backups/" \
                    --query "Contents[?LastModified<='$cutoff_date'].Key" \
                    --output text \
                    --region "$S3_REGION" | while read -r key; do
                    if [ -n "$key" ] && [ "$key" != "None" ]; then
                        log "Deleting old S3 backup: $key"
                        aws s3 rm "s3://$S3_BUCKET/$key" --region "$S3_REGION"
                    fi
                done
                success "Cleaned up S3 backups"
            fi
            ;;
        "all")
            cleanup_backups "local" "$retention_days"
            cleanup_backups "s3" "$retention_days"
            ;;
        *)
            error "Unknown location: $location"
            return 1
            ;;
    esac
}

# Verify backup integrity
verify_backup() {
    local backup_path="$1"
    
    if [ ! -f "$backup_path" ]; then
        error "Backup file not found: $backup_path"
        return 1
    fi
    
    log "Verifying backup integrity: $(basename "$backup_path")"
    
    # Check if file is compressed and decompress for verification
    local verify_file="$backup_path"
    if [[ "$backup_path" == *.gz ]]; then
        if ! gunzip -t "$backup_path" 2>/dev/null; then
            error "Backup file is corrupted (gzip test failed)"
            return 1
        fi
        verify_file="${backup_path%.gz}.verify"
        gunzip -c "$backup_path" > "$verify_file"
    fi
    
    # Verify PostgreSQL dump format
    if file "$verify_file" | grep -q "PostgreSQL custom database dump"; then
        # Verify custom format
        if ! pg_restore --list "$verify_file" >/dev/null 2>&1; then
            error "Backup file is corrupted (pg_restore verification failed)"
            [ "$verify_file" != "$backup_path" ] && rm -f "$verify_file"
            return 1
        fi
    else
        # Verify SQL format (basic check)
        if ! head -10 "$verify_file" | grep -q -E "(CREATE|INSERT|COPY)"; then
            error "Backup file does not appear to contain valid SQL"
            [ "$verify_file" != "$backup_path" ] && rm -f "$verify_file"
            return 1
        fi
    fi
    
    # Clean up temporary file
    [ "$verify_file" != "$backup_path" ] && rm -f "$verify_file"
    
    success "Backup verification passed"
}

# Main command handler
main() {
    local command="$1"
    shift
    
    case "$command" in
        "create")
            local backup_type="${1:-full}"
            check_env
            check_connection
            check_s3_tools
            ensure_backup_dir
            create_backup "$backup_type"
            ;;
        "restore")
            local backup_path="$1"
            local restore_options="$2"
            
            if [ -z "$backup_path" ]; then
                error "Backup file path is required"
                echo "Usage: $0 restore <backup_file> [restore_options]"
                exit 1
            fi
            
            # If backup path is just a filename, check if it's available locally or in S3
            if [ ! -f "$backup_path" ] && [ "$(basename "$backup_path")" = "$backup_path" ]; then
                local local_backup="$BACKUP_DIR/$backup_path"
                if [ -f "$local_backup" ]; then
                    backup_path="$local_backup"
                elif [ -n "$S3_BUCKET" ]; then
                    log "Backup not found locally, attempting to download from S3..."
                    check_s3_tools
                    backup_path=$(download_from_s3 "$backup_path")
                fi
            fi
            
            check_env
            check_connection
            verify_backup "$backup_path"
            restore_backup "$backup_path" "$restore_options"
            ;;
        "list")
            local location="${1:-all}"
            check_s3_tools
            list_backups "$location"
            ;;
        "cleanup")
            local location="${1:-all}"
            local retention_days="$2"
            check_s3_tools
            cleanup_backups "$location" "$retention_days"
            ;;
        "verify")
            local backup_path="$1"
            if [ -z "$backup_path" ]; then
                error "Backup file path is required"
                echo "Usage: $0 verify <backup_file>"
                exit 1
            fi
            verify_backup "$backup_path"
            ;;
        "help"|*)
            cat << EOF
GoTAK Database Backup and Restore Tool

Usage: $0 <command> [options]

Commands:
  create [type]        Create database backup (types: full, schema, data)
  restore <file>       Restore database from backup file
  list [location]      List available backups (locations: local, s3, all)
  cleanup [location]   Clean up old backups (locations: local, s3, all)
  verify <file>        Verify backup file integrity
  help                 Show this help message

Environment Variables:
  POSTGRES_HOST          Database host (default: localhost)
  POSTGRES_PORT          Database port (default: 5432)
  POSTGRES_DB            Database name (default: gotak)
  POSTGRES_USER          Database user (default: gotak)
  POSTGRES_PASSWORD      Database password (required)
  BACKUP_DIR             Local backup directory (default: ./backups)
  BACKUP_RETENTION_DAYS  Backup retention period (default: 30)
  
  S3 Configuration (optional):
  BACKUP_S3_BUCKET       S3 bucket for backup storage
  BACKUP_S3_REGION       S3 region (default: us-west-2)
  BACKUP_S3_ACCESS_KEY   S3 access key
  BACKUP_S3_SECRET_KEY   S3 secret key

Examples:
  $0 create full                    # Create full backup
  $0 create schema                  # Create schema-only backup
  $0 restore backup_20240101.sql.gz # Restore from backup
  $0 list local                     # List local backups
  $0 cleanup all 7                  # Clean up backups older than 7 days
  $0 verify backup_20240101.sql.gz  # Verify backup integrity

EOF
            ;;
    esac
}

# Run main function
main "$@"
