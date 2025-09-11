#!/bin/bash
# GoTAK Production Configuration Validation Script
# Validates environment variables and configuration files before startup
set -e

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

# Configuration
VALIDATION_ERRORS=0
VALIDATION_WARNINGS=0

# Required environment variables for production
REQUIRED_VARS=(
    "POSTGRES_HOST"
    "POSTGRES_PORT" 
    "POSTGRES_DB"
    "POSTGRES_USER"
    "POSTGRES_PASSWORD"
    "JWT_SECRET"
)

# Recommended environment variables
RECOMMENDED_VARS=(
    "REDIS_HOST"
    "REDIS_PORT"
    "REDIS_PASSWORD"
    "GOTAK_SERVER_NAME"
    "GOTAK_LOG_LEVEL"
    "TZ"
)

# Optional but important for production
OPTIONAL_PROD_VARS=(
    "BACKUP_S3_BUCKET"
    "SMTP_HOST"
    "SMTP_USER"
    "SMTP_PASSWORD"
    "GRAFANA_PASSWORD"
)

# Security-sensitive variables that should be properly secured
SENSITIVE_VARS=(
    "POSTGRES_PASSWORD"
    "JWT_SECRET"
    "REDIS_PASSWORD"
    "BACKUP_S3_SECRET_KEY"
    "SMTP_PASSWORD"
    "GRAFANA_PASSWORD"
)

# Increment error counter
add_error() {
    VALIDATION_ERRORS=$((VALIDATION_ERRORS + 1))
}

# Increment warning counter  
add_warning() {
    VALIDATION_WARNINGS=$((VALIDATION_WARNINGS + 1))
}

# Check if variable is set and not empty
check_required_var() {
    local var_name="$1"
    local var_value="${!var_name}"
    
    if [ -z "$var_value" ]; then
        error "Required environment variable '$var_name' is not set or empty"
        add_error
        return 1
    else
        log "✓ Required variable '$var_name' is set"
        return 0
    fi
}

# Check recommended variables with warnings
check_recommended_var() {
    local var_name="$1"
    local var_value="${!var_name}"
    
    if [ -z "$var_value" ]; then
        warn "Recommended environment variable '$var_name' is not set"
        add_warning
        return 1
    else
        log "✓ Recommended variable '$var_name' is set"
        return 0
    fi
}

# Check optional variables (just informational)
check_optional_var() {
    local var_name="$1"
    local var_value="${!var_name}"
    
    if [ -z "$var_value" ]; then
        log "○ Optional variable '$var_name' is not set"
        return 1
    else
        log "✓ Optional variable '$var_name' is set"
        return 0
    fi
}

# Validate specific configuration values
validate_postgres_config() {
    log "Validating PostgreSQL configuration..."
    
    # Check if port is numeric
    if ! [[ "$POSTGRES_PORT" =~ ^[0-9]+$ ]]; then
        error "POSTGRES_PORT must be a number, got: $POSTGRES_PORT"
        add_error
    fi
    
    # Check port range
    if [ "$POSTGRES_PORT" -lt 1 ] || [ "$POSTGRES_PORT" -gt 65535 ]; then
        error "POSTGRES_PORT must be between 1 and 65535, got: $POSTGRES_PORT"
        add_error
    fi
    
    # Check database name format
    if ! [[ "$POSTGRES_DB" =~ ^[a-zA-Z][a-zA-Z0-9_]*$ ]]; then
        error "POSTGRES_DB contains invalid characters: $POSTGRES_DB"
        add_error
    fi
    
    # Check username format
    if ! [[ "$POSTGRES_USER" =~ ^[a-zA-Z][a-zA-Z0-9_]*$ ]]; then
        error "POSTGRES_USER contains invalid characters: $POSTGRES_USER"
        add_error
    fi
}

# Validate JWT secret strength
validate_jwt_secret() {
    log "Validating JWT secret..."
    
    local jwt_length=${#JWT_SECRET}
    
    if [ "$jwt_length" -lt 32 ]; then
        error "JWT_SECRET is too short ($jwt_length chars). Minimum 32 characters required for security."
        add_error
    elif [ "$jwt_length" -lt 64 ]; then
        warn "JWT_SECRET is short ($jwt_length chars). Recommend 64+ characters for better security."
        add_warning
    else
        success "JWT_SECRET length is adequate ($jwt_length chars)"
    fi
    
    # Check for common weak secrets
    case "$JWT_SECRET" in
        "your-secret-key"|"secret"|"password"|"123456"|"changeme"|"default")
            error "JWT_SECRET appears to be a default or weak value. Use a strong, random secret."
            add_error
            ;;
        *"test"*|*"dev"*|*"example"*)
            warn "JWT_SECRET appears to contain test/dev/example text. Ensure this is a production secret."
            add_warning
            ;;
    esac
}

# Validate Redis configuration
validate_redis_config() {
    if [ -n "$REDIS_HOST" ]; then
        log "Validating Redis configuration..."
        
        if [ -n "$REDIS_PORT" ] && ! [[ "$REDIS_PORT" =~ ^[0-9]+$ ]]; then
            error "REDIS_PORT must be a number, got: $REDIS_PORT"
            add_error
        fi
        
        if [ -z "$REDIS_PASSWORD" ]; then
            warn "REDIS_PASSWORD not set. Redis will be unprotected."
            add_warning
        fi
    fi
}

# Validate server configuration
validate_server_config() {
    log "Validating server configuration..."
    
    # Check log level
    if [ -n "$GOTAK_LOG_LEVEL" ]; then
        case "$GOTAK_LOG_LEVEL" in
            "debug"|"info"|"warn"|"error")
                log "✓ Log level '$GOTAK_LOG_LEVEL' is valid"
                ;;
            *)
                warn "Unknown log level '$GOTAK_LOG_LEVEL'. Valid levels: debug, info, warn, error"
                add_warning
                ;;
        esac
    fi
    
    # Check timezone
    if [ -n "$TZ" ]; then
        if [ ! -f "/usr/share/zoneinfo/$TZ" ]; then
            warn "Timezone '$TZ' may not be valid"
            add_warning
        fi
    fi
    
    # Check server name for production
    if [ -n "$GOTAK_SERVER_NAME" ]; then
        if [[ "$GOTAK_SERVER_NAME" == "localhost" ]]; then
            warn "GOTAK_SERVER_NAME is 'localhost'. Consider setting to actual domain name for production."
            add_warning
        fi
    fi
}

# Check file permissions and security
validate_file_security() {
    log "Validating file security..."
    
    # Check if running as root (security risk)
    if [ "$(id -u)" -eq 0 ]; then
        error "Running as root user. This is a security risk in production."
        add_error
    fi
    
    # Check configuration file permissions
    local config_file="${GOTAK_CONFIG_PATH:-/app/config/production.yaml}"
    if [ -f "$config_file" ]; then
        local permissions=$(stat -c "%a" "$config_file" 2>/dev/null || stat -f "%A" "$config_file" 2>/dev/null)
        if [[ "$permissions" =~ [2367] ]]; then
            warn "Configuration file '$config_file' has world-writable permissions ($permissions)"
            add_warning
        fi
    fi
}

# Validate network configuration
validate_network_config() {
    log "Validating network configuration..."
    
    # Check if all required ports are available
    local ports_to_check=("${GOTAK_HTTP_PORT:-8080}" "${GOTAK_TAK_PORT:-8087}" "${GOTAK_TLS_PORT:-8089}")
    
    for port in "${ports_to_check[@]}"; do
        if command -v netstat >/dev/null 2>&1; then
            if netstat -tuln 2>/dev/null | grep -q ":$port "; then
                warn "Port $port appears to be in use"
                add_warning
            fi
        fi
    done
}

# Check secret management
validate_secrets() {
    log "Validating secret management..."
    
    for var in "${SENSITIVE_VARS[@]}"; do
        local var_value="${!var}"
        
        if [ -n "$var_value" ]; then
            # Check if secret is hardcoded (potential security issue)
            if env | grep -q "^${var}="; then
                warn "Sensitive variable '$var' is set in environment. Consider using Docker secrets or external secret management."
                add_warning
            fi
            
            # Check for common insecure values
            case "$var_value" in
                "password"|"123456"|"admin"|"changeme"|"secret"|"default")
                    error "Variable '$var' has a weak/default value"
                    add_error
                    ;;
            esac
        fi
    done
}

# Test database connectivity
test_database_connectivity() {
    log "Testing database connectivity..."
    
    if command -v pg_isready >/dev/null 2>&1; then
        if timeout 10 pg_isready -h "${POSTGRES_HOST}" -p "${POSTGRES_PORT}" -U "${POSTGRES_USER}" >/dev/null 2>&1; then
            success "Database connectivity test passed"
        else
            error "Cannot connect to database at ${POSTGRES_HOST}:${POSTGRES_PORT}"
            add_error
        fi
    else
        warn "pg_isready not available, skipping database connectivity test"
        add_warning
    fi
}

# Test Redis connectivity
test_redis_connectivity() {
    if [ -n "$REDIS_HOST" ]; then
        log "Testing Redis connectivity..."
        
        if command -v redis-cli >/dev/null 2>&1; then
            local redis_port="${REDIS_PORT:-6379}"
            if [ -n "$REDIS_PASSWORD" ]; then
                if timeout 5 redis-cli -h "${REDIS_HOST}" -p "$redis_port" -a "${REDIS_PASSWORD}" ping >/dev/null 2>&1; then
                    success "Redis connectivity test passed"
                else
                    error "Cannot connect to Redis at ${REDIS_HOST}:${redis_port}"
                    add_error
                fi
            else
                if timeout 5 redis-cli -h "${REDIS_HOST}" -p "$redis_port" ping >/dev/null 2>&1; then
                    success "Redis connectivity test passed"
                else
                    warn "Cannot connect to Redis at ${REDIS_HOST}:${redis_port}"
                    add_warning
                fi
            fi
        else
            warn "redis-cli not available, skipping Redis connectivity test"
            add_warning
        fi
    fi
}

# Validate configuration file
validate_config_file() {
    local config_file="${GOTAK_CONFIG_PATH:-/app/config/production.yaml}"
    
    log "Validating configuration file: $config_file"
    
    if [ ! -f "$config_file" ]; then
        error "Configuration file not found: $config_file"
        add_error
        return 1
    fi
    
    if [ ! -r "$config_file" ]; then
        error "Configuration file is not readable: $config_file"
        add_error
        return 1
    fi
    
    # Basic YAML syntax validation
    if command -v python3 >/dev/null 2>&1; then
        if ! python3 -c "import yaml; yaml.safe_load(open('$config_file'))" 2>/dev/null; then
            error "Configuration file has invalid YAML syntax"
            add_error
            return 1
        fi
        success "Configuration file YAML syntax is valid"
    else
        warn "Python3 not available, skipping YAML syntax validation"
        add_warning
    fi
}

# Generate configuration report
generate_report() {
    local report_file="/tmp/gotak-config-validation-$(date +%Y%m%d_%H%M%S).txt"
    
    {
        echo "GoTAK Configuration Validation Report"
        echo "Generated: $(date)"
        echo "======================================"
        echo ""
        echo "Environment Variables:"
        echo "---------------------"
        
        echo "Required variables:"
        for var in "${REQUIRED_VARS[@]}"; do
            if [ -n "${!var}" ]; then
                echo "  ✓ $var=***"
            else
                echo "  ✗ $var=(not set)"
            fi
        done
        
        echo ""
        echo "Recommended variables:"
        for var in "${RECOMMENDED_VARS[@]}"; do
            if [ -n "${!var}" ]; then
                echo "  ✓ $var=${!var}"
            else
                echo "  ○ $var=(not set)"
            fi
        done
        
        echo ""
        echo "Optional variables:"
        for var in "${OPTIONAL_PROD_VARS[@]}"; do
            if [ -n "${!var}" ]; then
                echo "  ✓ $var=(set)"
            else
                echo "  ○ $var=(not set)"
            fi
        done
        
        echo ""
        echo "Validation Summary:"
        echo "------------------"
        echo "Errors: $VALIDATION_ERRORS"
        echo "Warnings: $VALIDATION_WARNINGS"
        
        if [ $VALIDATION_ERRORS -eq 0 ] && [ $VALIDATION_WARNINGS -eq 0 ]; then
            echo "Status: PASSED"
        elif [ $VALIDATION_ERRORS -eq 0 ]; then
            echo "Status: PASSED (with warnings)"
        else
            echo "Status: FAILED"
        fi
        
    } > "$report_file"
    
    echo "$report_file"
}

# Main validation function
main() {
    local mode="${1:-full}"
    
    case "$mode" in
        "quick")
            log "Running quick configuration validation..."
            
            # Check only required variables
            for var in "${REQUIRED_VARS[@]}"; do
                check_required_var "$var"
            done
            ;;
            
        "full")
            log "Running full configuration validation..."
            
            # Check all variable categories
            log "Checking required variables..."
            for var in "${REQUIRED_VARS[@]}"; do
                check_required_var "$var"
            done
            
            log "Checking recommended variables..."
            for var in "${RECOMMENDED_VARS[@]}"; do
                check_recommended_var "$var"
            done
            
            log "Checking optional variables..."
            for var in "${OPTIONAL_PROD_VARS[@]}"; do
                check_optional_var "$var"
            done
            
            # Validate specific configurations
            validate_postgres_config
            validate_jwt_secret
            validate_redis_config
            validate_server_config
            validate_file_security
            validate_network_config
            validate_secrets
            validate_config_file
            ;;
            
        "connectivity")
            log "Testing service connectivity..."
            test_database_connectivity
            test_redis_connectivity
            ;;
            
        "report")
            log "Generating configuration report..."
            main full > /dev/null 2>&1
            local report_file
            report_file=$(generate_report)
            success "Configuration report generated: $report_file"
            cat "$report_file"
            ;;
            
        "help"|*)
            cat << EOF
GoTAK Configuration Validation Tool

Usage: $0 [mode]

Modes:
  quick        Quick validation (required variables only)
  full         Full validation (default)
  connectivity Test service connectivity
  report       Generate detailed validation report
  help         Show this help message

Examples:
  $0           # Run full validation
  $0 quick     # Run quick validation
  $0 report    # Generate and display report

EOF
            exit 0
            ;;
    esac
    
    # Summary
    echo ""
    if [ $VALIDATION_ERRORS -eq 0 ] && [ $VALIDATION_WARNINGS -eq 0 ]; then
        success "Configuration validation passed! (0 errors, 0 warnings)"
        exit 0
    elif [ $VALIDATION_ERRORS -eq 0 ]; then
        warn "Configuration validation passed with warnings ($VALIDATION_WARNINGS warnings)"
        exit 0
    else
        error "Configuration validation failed! ($VALIDATION_ERRORS errors, $VALIDATION_WARNINGS warnings)"
        exit 1
    fi
}

# Run main function
main "$@"
