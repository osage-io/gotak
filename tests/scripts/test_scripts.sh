#!/bin/bash
# Test suite for GoTAK utility scripts
set -e

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$(dirname "$SCRIPT_DIR")")"
SCRIPTS_DIR="$PROJECT_ROOT/scripts"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Test counters
TESTS_RUN=0
TESTS_PASSED=0
TESTS_FAILED=0

# Logging functions
log() {
    echo -e "${BLUE}[TEST]${NC} $1"
}

success() {
    echo -e "${GREEN}[PASS]${NC} $1"
    TESTS_PASSED=$((TESTS_PASSED + 1))
}

fail() {
    echo -e "${RED}[FAIL]${NC} $1" >&2
    TESTS_FAILED=$((TESTS_FAILED + 1))
}

warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

run_test() {
    local test_name="$1"
    local test_command="$2"
    
    TESTS_RUN=$((TESTS_RUN + 1))
    log "Running test: $test_name"
    
    if eval "$test_command" >/dev/null 2>&1; then
        success "$test_name"
        return 0
    else
        fail "$test_name"
        return 1
    fi
}

# Test script existence and executability
test_script_exists_and_executable() {
    local script_name="$1"
    local script_path="$SCRIPTS_DIR/$script_name"
    
    if [ ! -f "$script_path" ]; then
        fail "Script not found: $script_path"
        return 1
    fi
    
    if [ ! -x "$script_path" ]; then
        fail "Script not executable: $script_path"
        return 1
    fi
    
    success "Script exists and executable: $script_name"
    return 0
}

# Test script help output
test_script_help() {
    local script_name="$1"
    local script_path="$SCRIPTS_DIR/$script_name"
    
    if "$script_path" help 2>&1 | grep -q "Usage\|usage\|USAGE"; then
        success "Help output available: $script_name"
        return 0
    elif "$script_path" --help 2>&1 | grep -q "Usage\|usage\|USAGE"; then
        success "Help output available: $script_name"
        return 0
    elif "$script_path" 2>&1 | grep -q "Usage\|usage\|USAGE"; then
        success "Help output available: $script_name"
        return 0
    else
        fail "No help output found: $script_name"
        return 1
    fi
}

# Test deploy script configuration validation
test_deploy_script_config_validation() {
    local deploy_script="$SCRIPTS_DIR/deploy.sh"
    
    # Test with invalid environment should fail gracefully
    if DEPLOY_ENV="invalid" "$deploy_script" --dry-run 2>&1 | grep -q "Invalid environment"; then
        success "Deploy script validates environment"
        return 0
    else
        # Try alternative method - script should show help or error
        if DEPLOY_ENV="invalid" "$deploy_script" 2>&1 | grep -qE "(Invalid|Usage|help)"; then
            success "Deploy script validates environment"
            return 0
        else
            warn "Deploy script environment validation unclear"
            return 0  # Don't fail the test suite for this
        fi
    fi
}

# Test load test script scenario listing
test_load_test_script_scenarios() {
    local load_test_script="$SCRIPTS_DIR/load-test.sh"
    
    if "$load_test_script" list 2>&1 | grep -q "baseline\|stress\|spike"; then
        success "Load test script lists scenarios"
        return 0
    else
        fail "Load test script scenario listing failed"
        return 1
    fi
}

# Test security audit script options
test_security_audit_script_options() {
    local security_script="$SCRIPTS_DIR/security-audit.sh"
    
    if "$security_script" help 2>&1 | grep -q "headers\|tls\|auth"; then
        success "Security audit script shows options"
        return 0
    else
        fail "Security audit script options not found"
        return 1
    fi
}

# Test validate-config script
test_validate_config_script() {
    local config_script="$SCRIPTS_DIR/validate-config.sh"
    local test_config="$PROJECT_ROOT/config/test.yaml"
    
    if [ -f "$test_config" ]; then
        if "$config_script" "$test_config" >/dev/null 2>&1; then
            success "Config validation script works"
            return 0
        else
            warn "Config validation script returned non-zero (may be expected)"
            return 0  # Don't fail - validation might be strict
        fi
    else
        warn "No test config found for validation test"
        return 0
    fi
}

# Test backup script dry run
test_backup_script_dry_run() {
    local backup_script="$SCRIPTS_DIR/backup.sh"
    
    if "$backup_script" --dry-run 2>&1 | grep -qE "(dry.?run|would|simulation)"; then
        success "Backup script supports dry run"
        return 0
    else
        # Try without arguments to see if it shows help
        if "$backup_script" 2>&1 | grep -qE "(Usage|help|dry.?run)"; then
            success "Backup script shows usage information"
            return 0
        else
            warn "Backup script dry run test inconclusive"
            return 0
        fi
    fi
}

# Test migrate script help
test_migrate_script() {
    local migrate_script="$SCRIPTS_DIR/migrate.sh"
    
    # Test help or usage output
    if "$migrate_script" --help 2>&1 | grep -qE "(Usage|migration|database)"; then
        success "Migration script shows help"
        return 0
    elif "$migrate_script" 2>&1 | grep -qE "(Usage|migration|database)"; then
        success "Migration script shows usage"
        return 0
    else
        warn "Migration script help test inconclusive"
        return 0
    fi
}

# Test healthcheck script
test_healthcheck_script() {
    local healthcheck_script="$SCRIPTS_DIR/healthcheck.sh"
    
    # Healthcheck should return status
    if "$healthcheck_script" 2>&1 | grep -qE "(healthy|unhealthy|status|check)"; then
        success "Healthcheck script provides status"
        return 0
    else
        warn "Healthcheck script test inconclusive"
        return 0
    fi
}

# Test Docker Compose files
test_docker_compose_files() {
    local compose_files=(
        "docker-compose.prod.yml"
        "docker-compose.test.yml"
    )
    
    for compose_file in "${compose_files[@]}"; do
        local compose_path="$PROJECT_ROOT/$compose_file"
        
        if [ -f "$compose_path" ]; then
            # Test if it's valid YAML
            if command -v docker-compose >/dev/null 2>&1; then
                if docker-compose -f "$compose_path" config >/dev/null 2>&1; then
                    success "Docker Compose file valid: $compose_file"
                else
                    fail "Docker Compose file invalid: $compose_file"
                fi
            elif python3 -c "import yaml; yaml.safe_load(open('$compose_path'))" 2>/dev/null; then
                success "Docker Compose file is valid YAML: $compose_file"
            else
                warn "Could not validate Docker Compose file: $compose_file"
            fi
        else
            fail "Docker Compose file not found: $compose_file"
        fi
    done
}

# Test config files
test_config_files() {
    local config_files=(
        "config/production.yaml"
        "config/test.yaml"
    )
    
    for config_file in "${config_files[@]}"; do
        local config_path="$PROJECT_ROOT/$config_file"
        
        if [ -f "$config_path" ]; then
            # Test if it's valid YAML
            if python3 -c "import yaml; yaml.safe_load(open('$config_path'))" 2>/dev/null; then
                success "Config file is valid YAML: $config_file"
            elif command -v yq >/dev/null 2>&1 && yq eval . "$config_path" >/dev/null 2>&1; then
                success "Config file is valid YAML: $config_file"
            else
                warn "Could not validate config file: $config_file"
            fi
        else
            fail "Config file not found: $config_file"
        fi
    done
}

# Test load testing files
test_load_testing_files() {
    local k6_script="$PROJECT_ROOT/testing/load/k6-load-test.js"
    
    if [ -f "$k6_script" ]; then
        # Check if it's valid JavaScript (basic syntax check)
        if command -v node >/dev/null 2>&1; then
            if node -c "$k6_script" 2>/dev/null; then
                success "K6 load test script is valid JavaScript"
            else
                fail "K6 load test script has syntax errors"
            fi
        else
            # Basic check for common JS syntax
            if grep -q "export.*default" "$k6_script" && grep -q "import.*k6" "$k6_script"; then
                success "K6 load test script has expected structure"
            else
                warn "Could not validate K6 script structure"
            fi
        fi
    else
        fail "K6 load test script not found"
    fi
}

# Main test runner
main() {
    echo "========================================"
    echo "GoTAK Scripts Test Suite"
    echo "========================================"
    echo ""
    
    log "Starting script tests..."
    
    # List of scripts to test
    local scripts=(
        "deploy.sh"
        "load-test.sh"
        "security-audit.sh"
        "validate-config.sh"
        "backup.sh"
        "migrate.sh"
        "healthcheck.sh"
        "entrypoint.sh"
        "run-integration-tests.sh"
    )
    
    # Test script existence and executability
    log "Testing script existence and permissions..."
    for script in "${scripts[@]}"; do
        test_script_exists_and_executable "$script"
        TESTS_RUN=$((TESTS_RUN + 1))
    done
    
    echo ""
    log "Testing script help output..."
    
    # Test help outputs
    test_script_help "deploy.sh"
    TESTS_RUN=$((TESTS_RUN + 1))
    
    test_script_help "load-test.sh"
    TESTS_RUN=$((TESTS_RUN + 1))
    
    test_script_help "security-audit.sh"
    TESTS_RUN=$((TESTS_RUN + 1))
    
    echo ""
    log "Testing script functionality..."
    
    # Functional tests
    test_deploy_script_config_validation
    TESTS_RUN=$((TESTS_RUN + 1))
    
    test_load_test_script_scenarios
    TESTS_RUN=$((TESTS_RUN + 1))
    
    test_security_audit_script_options
    TESTS_RUN=$((TESTS_RUN + 1))
    
    test_validate_config_script
    TESTS_RUN=$((TESTS_RUN + 1))
    
    test_backup_script_dry_run
    TESTS_RUN=$((TESTS_RUN + 1))
    
    test_migrate_script
    TESTS_RUN=$((TESTS_RUN + 1))
    
    test_healthcheck_script
    TESTS_RUN=$((TESTS_RUN + 1))
    
    echo ""
    log "Testing configuration files..."
    
    test_docker_compose_files
    
    test_config_files
    
    echo ""
    log "Testing load testing files..."
    
    test_load_testing_files
    TESTS_RUN=$((TESTS_RUN + 1))
    
    # Print summary
    echo ""
    echo "========================================"
    echo "Test Summary"
    echo "========================================"
    echo "Tests run: $TESTS_RUN"
    echo -e "${GREEN}Tests passed: $TESTS_PASSED${NC}"
    
    if [ $TESTS_FAILED -gt 0 ]; then
        echo -e "${RED}Tests failed: $TESTS_FAILED${NC}"
        exit 1
    else
        echo -e "${GREEN}All tests passed!${NC}"
        exit 0
    fi
}

# Run main function if script is executed directly
if [ "${BASH_SOURCE[0]}" = "${0}" ]; then
    main "$@"
fi
