#!/bin/bash
# Integration test runner for GoTAK
set -e

echo "===== GoTAK Integration Test Runner ====="

# Configuration
COMPOSE_FILE="${COMPOSE_FILE:-docker-compose.test.yml}"
PROJECT_NAME="gotak-integration-test"
TIMEOUT="${TIMEOUT:-300}"
PARALLEL="${PARALLEL:-1}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

log() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

cleanup() {
    log "Cleaning up test environment..."
    docker-compose -f $COMPOSE_FILE -p $PROJECT_NAME down -v --remove-orphans || true
    docker system prune -f --volumes || true
}

wait_for_services() {
    log "Waiting for services to be ready..."
    
    local max_attempts=60
    local attempt=1
    
    while [ $attempt -le $max_attempts ]; do
        if docker-compose -f $COMPOSE_FILE -p $PROJECT_NAME exec -T gotak-test wget -q --spider http://localhost:8080/health 2>/dev/null; then
            success "Backend service is ready"
            return 0
        fi
        
        log "Attempt $attempt/$max_attempts - waiting for services..."
        sleep 5
        attempt=$((attempt + 1))
    done
    
    error "Services failed to become ready within timeout"
    return 1
}

setup_test_environment() {
    log "Setting up test environment..."
    
    # Clean up any existing containers
    cleanup
    
    # Build and start services
    log "Building test containers..."
    docker-compose -f $COMPOSE_FILE -p $PROJECT_NAME build --no-cache
    
    log "Starting test services..."
    docker-compose -f $COMPOSE_FILE -p $PROJECT_NAME up -d
    
    # Wait for services
    if ! wait_for_services; then
        error "Failed to start services"
        show_logs
        cleanup
        exit 1
    fi
    
    log "Services are ready, setting up test data..."
    
    # Run data setup
    docker-compose -f $COMPOSE_FILE -p $PROJECT_NAME exec -T gotak-test /app/test-data/setup.sh
    
    success "Test environment setup complete"
}

run_tests() {
    log "Running integration tests..."
    
    local test_pattern="${1:-./tests/e2e/...}"
    local test_flags=""
    
    # Add parallel flag if specified
    if [ "$PARALLEL" -gt 1 ]; then
        test_flags="$test_flags -p $PARALLEL"
    fi
    
    # Add timeout
    test_flags="$test_flags -timeout ${TIMEOUT}s"
    
    # Add verbose flag
    test_flags="$test_flags -v"
    
    # Set environment variables for tests
    export GOTAK_BASE_URL="http://localhost:8080"
    export GOTAK_WS_URL="ws://localhost:8080"
    export POSTGRES_HOST="localhost"
    export POSTGRES_PORT="5433"
    export POSTGRES_USER="gotak"
    export POSTGRES_PASSWORD="gotak_test_pass"
    export POSTGRES_DB="gotak_test"
    
    log "Running tests with pattern: $test_pattern"
    log "Test flags: $test_flags"
    
    # Run tests inside container
    if docker-compose -f $COMPOSE_FILE -p $PROJECT_NAME exec -T \
        -e GOTAK_BASE_URL="http://gotak-test:8080" \
        -e GOTAK_WS_URL="ws://gotak-test:8080" \
        gotak-test go test $test_flags $test_pattern; then
        success "All integration tests passed!"
        return 0
    else
        error "Integration tests failed!"
        return 1
    fi
}

run_performance_tests() {
    log "Running performance tests..."
    
    # Run specific performance test functions
    local perf_tests=(
        "TestPerformanceBaseline"
        "TestWebSocketPerformance"
        "TestConcurrentAccess"
    )
    
    for test in "${perf_tests[@]}"; do
        log "Running performance test: $test"
        
        if docker-compose -f $COMPOSE_FILE -p $PROJECT_NAME exec -T \
            -e GOTAK_BASE_URL="http://gotak-test:8080" \
            -e GOTAK_WS_URL="ws://gotak-test:8080" \
            gotak-test go test -v -timeout 120s -run "^$test$" ./tests/e2e/; then
            success "Performance test $test passed"
        else
            warn "Performance test $test failed"
        fi
    done
}

show_logs() {
    log "Showing service logs..."
    
    echo "=== Database Logs ==="
    docker-compose -f $COMPOSE_FILE -p $PROJECT_NAME logs --tail=50 postgres-test || true
    
    echo "=== GoTAK Server Logs ==="
    docker-compose -f $COMPOSE_FILE -p $PROJECT_NAME logs --tail=50 gotak-test || true
}

generate_report() {
    log "Generating test report..."
    
    local report_dir="./test-reports"
    mkdir -p $report_dir
    
    # Generate coverage report if available
    if docker-compose -f $COMPOSE_FILE -p $PROJECT_NAME exec -T gotak-test \
        go test -coverprofile=/tmp/coverage.out ./tests/e2e/ 2>/dev/null; then
        docker-compose -f $COMPOSE_FILE -p $PROJECT_NAME exec -T gotak-test \
            go tool cover -html=/tmp/coverage.out -o /app/test-reports/coverage.html
        log "Coverage report generated: test-reports/coverage.html"
    fi
    
    # Container and service status
    docker-compose -f $COMPOSE_FILE -p $PROJECT_NAME ps > $report_dir/container-status.txt
    
    # Service logs
    docker-compose -f $COMPOSE_FILE -p $PROJECT_NAME logs > $report_dir/service-logs.txt
    
    success "Test report generated in: $report_dir/"
}

main() {
    local command="${1:-all}"
    local test_pattern="$2"
    
    case $command in
        "setup")
            setup_test_environment
            ;;
        "test")
            run_tests "$test_pattern"
            ;;
        "perf")
            run_performance_tests
            ;;
        "logs")
            show_logs
            ;;
        "report")
            generate_report
            ;;
        "cleanup")
            cleanup
            ;;
        "all")
            trap cleanup EXIT
            setup_test_environment
            
            success "Running full integration test suite..."
            
            local test_result=0
            
            # Run main integration tests
            if ! run_tests "$test_pattern"; then
                test_result=1
            fi
            
            # Run performance tests
            log "Running performance benchmarks..."
            run_performance_tests
            
            # Generate report
            generate_report
            
            if [ $test_result -eq 0 ]; then
                success "Integration test suite completed successfully!"
            else
                error "Integration test suite failed!"
            fi
            
            exit $test_result
            ;;
        "help"|*)
            echo "Usage: $0 [command] [test_pattern]"
            echo ""
            echo "Commands:"
            echo "  all      - Run complete integration test suite (default)"
            echo "  setup    - Set up test environment only"
            echo "  test     - Run integration tests only"
            echo "  perf     - Run performance tests only"
            echo "  logs     - Show service logs"
            echo "  report   - Generate test report"
            echo "  cleanup  - Clean up test environment"
            echo "  help     - Show this help message"
            echo ""
            echo "Environment Variables:"
            echo "  COMPOSE_FILE  - Docker compose file to use (default: docker-compose.test.yml)"
            echo "  TIMEOUT       - Test timeout in seconds (default: 300)"
            echo "  PARALLEL      - Number of parallel test processes (default: 1)"
            echo ""
            echo "Examples:"
            echo "  $0 all                           # Run complete test suite"
            echo "  $0 test TestAuthentication       # Run specific test"
            echo "  $0 test ./tests/e2e/auth_test.go # Run specific test file"
            echo "  PARALLEL=4 $0 test              # Run tests in parallel"
            ;;
    esac
}

# Handle script interruption
trap cleanup INT TERM

# Run main function with all arguments
main "$@"
