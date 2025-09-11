#!/bin/bash
# GoTAK Load Testing Runner
# Executes comprehensive load testing scenarios with k6
set -e

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
LOAD_TEST_DIR="$PROJECT_ROOT/testing/load"
REPORTS_DIR="$PROJECT_ROOT/test-reports/load"

# Environment defaults
GOTAK_BASE_URL="${GOTAK_BASE_URL:-http://localhost:8080}"
GOTAK_WS_URL="${GOTAK_WS_URL:-ws://localhost:8080}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Logging functions
log() {
    echo -e "${BLUE}[LOAD-TEST]${NC} $1"
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

# Check prerequisites
check_prerequisites() {
    log "Checking load testing prerequisites..."
    
    # Check if k6 is installed
    if ! command -v k6 >/dev/null 2>&1; then
        error "k6 is not installed. Please install k6 for load testing:"
        error "  macOS: brew install k6"
        error "  Linux: sudo apt install k6"
        error "  Or download from: https://k6.io/docs/getting-started/installation/"
        exit 1
    fi
    
    local k6_version
    k6_version=$(k6 version | grep -o 'v[0-9]\+\.[0-9]\+\.[0-9]\+' | head -1)
    log "Found k6 version: $k6_version"
    
    # Check if GoTAK service is running
    log "Checking if GoTAK service is accessible..."
    if ! curl -s -f "$GOTAK_BASE_URL/health" >/dev/null 2>&1; then
        error "GoTAK service is not accessible at $GOTAK_BASE_URL"
        error "Please ensure the service is running before starting load tests"
        exit 1
    fi
    
    success "Prerequisites check passed"
}

# Setup test environment
setup_test_environment() {
    log "Setting up load testing environment..."
    
    # Create reports directory
    mkdir -p "$REPORTS_DIR"
    
    # Ensure test data exists
    log "Verifying test data availability..."
    # The load test script will use the seeded test users
    
    success "Test environment setup completed"
}

# Run specific load test scenario
run_load_test() {
    local scenario="${1:-all}"
    local duration="${2:-}"
    local vus="${3:-}"
    
    log "Running load test scenario: $scenario"
    
    # Prepare test file
    local test_file="$LOAD_TEST_DIR/k6-load-test.js"
    if [ ! -f "$test_file" ]; then
        error "Load test file not found: $test_file"
        exit 1
    fi
    
    # Generate timestamp for report files
    local timestamp=$(date '+%Y%m%d_%H%M%S')
    local report_base="$REPORTS_DIR/${scenario}_${timestamp}"
    
    # Prepare k6 options
    local k6_options=""
    
    # Set environment variables
    export GOTAK_BASE_URL="$GOTAK_BASE_URL"
    export GOTAK_WS_URL="$GOTAK_WS_URL"
    
    # Generate reports
    k6_options="$k6_options --out json=${report_base}.json"
    k6_options="$k6_options --out csv=${report_base}.csv"
    
    # Override duration and VUs if specified
    if [ -n "$duration" ]; then
        k6_options="$k6_options --duration $duration"
    fi
    if [ -n "$vus" ]; then
        k6_options="$k6_options --vus $vus"
    fi
    
    # Add scenario-specific options
    case "$scenario" in
        "baseline")
            k6_options="$k6_options --include-system-env-vars --quiet"
            ;;
        "stress")
            k6_options="$k6_options --include-system-env-vars"
            ;;
        "spike")
            k6_options="$k6_options --include-system-env-vars"
            ;;
        "websocket")
            k6_options="$k6_options --include-system-env-vars"
            ;;
        "database")
            k6_options="$k6_options --include-system-env-vars"
            ;;
        "quick")
            # Quick test with reduced load
            k6_options="$k6_options --vus 5 --duration 2m --quiet"
            ;;
        "all")
            # Run all scenarios (default k6 behavior)
            k6_options="$k6_options --include-system-env-vars"
            ;;
        *)
            error "Unknown scenario: $scenario"
            exit 1
            ;;
    esac
    
    log "Executing k6 load test..."
    log "Command: k6 run $k6_options $test_file"
    
    # Run the test
    local test_result=0
    if ! k6 run $k6_options "$test_file" 2>&1 | tee "${report_base}.log"; then
        test_result=1
        error "Load test failed"
    fi
    
    # Generate HTML report if possible
    generate_html_report "$report_base" "$scenario"
    
    # Show summary
    show_test_summary "$report_base" "$scenario" $test_result
    
    return $test_result
}

# Generate HTML report from k6 outputs
generate_html_report() {
    local report_base="$1"
    local scenario="$2"
    
    log "Generating HTML report..."
    
    # Create basic HTML report
    local html_file="${report_base}.html"
    
    cat > "$html_file" << EOF
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>GoTAK Load Test Report - $scenario</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; background-color: #f5f5f5; }
        .container { max-width: 1200px; margin: 0 auto; background: white; padding: 20px; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        h1 { color: #333; border-bottom: 2px solid #007acc; padding-bottom: 10px; }
        h2 { color: #555; margin-top: 30px; }
        .metric { background: #f8f9fa; padding: 15px; border-radius: 4px; margin: 10px 0; border-left: 4px solid #007acc; }
        .metric-name { font-weight: bold; color: #333; }
        .metric-value { font-size: 1.2em; color: #007acc; margin-top: 5px; }
        .status-pass { color: #28a745; font-weight: bold; }
        .status-fail { color: #dc3545; font-weight: bold; }
        .timestamp { color: #666; font-size: 0.9em; }
        pre { background: #f8f9fa; padding: 15px; border-radius: 4px; overflow-x: auto; }
        table { width: 100%; border-collapse: collapse; margin: 20px 0; }
        th, td { padding: 12px; text-align: left; border-bottom: 1px solid #ddd; }
        th { background-color: #f8f9fa; font-weight: bold; }
    </style>
</head>
<body>
    <div class="container">
        <h1>GoTAK Load Test Report</h1>
        <div class="timestamp">Generated: $(date)</div>
        <div class="timestamp">Scenario: $scenario</div>
        <div class="timestamp">Target: $GOTAK_BASE_URL</div>
        
        <h2>Test Configuration</h2>
        <div class="metric">
            <div class="metric-name">Base URL</div>
            <div class="metric-value">$GOTAK_BASE_URL</div>
        </div>
        <div class="metric">
            <div class="metric-name">WebSocket URL</div>
            <div class="metric-value">$GOTAK_WS_URL</div>
        </div>
        <div class="metric">
            <div class="metric-name">Test Scenario</div>
            <div class="metric-value">$scenario</div>
        </div>
        
        <h2>Raw Test Output</h2>
        <pre>$(cat "${report_base}.log" 2>/dev/null || echo "Log file not available")</pre>
        
        <h2>Files Generated</h2>
        <ul>
            <li><a href="$(basename "${report_base}.json")">JSON Report</a></li>
            <li><a href="$(basename "${report_base}.csv")">CSV Report</a></li>
            <li><a href="$(basename "${report_base}.log")">Test Log</a></li>
        </ul>
    </div>
</body>
</html>
EOF
    
    success "HTML report generated: $html_file"
}

# Show test summary
show_test_summary() {
    local report_base="$1"
    local scenario="$2"
    local test_result="$3"
    
    log "Load Test Summary"
    echo "=================="
    echo "Scenario: $scenario"
    echo "Timestamp: $(date)"
    
    if [ $test_result -eq 0 ]; then
        echo -e "Status: ${GREEN}PASSED${NC}"
    else
        echo -e "Status: ${RED}FAILED${NC}"
    fi
    
    echo "Reports generated:"
    echo "  HTML: ${report_base}.html"
    echo "  JSON: ${report_base}.json"
    echo "  CSV:  ${report_base}.csv"
    echo "  Log:  ${report_base}.log"
    echo ""
    
    # Extract key metrics from log if available
    if [ -f "${report_base}.log" ]; then
        echo "Key Metrics:"
        
        # Extract summary statistics
        if grep -q "✓\|✗" "${report_base}.log"; then
            echo "Checks:"
            grep "✓\|✗" "${report_base}.log" | tail -10
        fi
        
        # Extract performance metrics
        if grep -q "http_req_duration" "${report_base}.log"; then
            echo ""
            echo "Performance:"
            grep -A 5 "http_req_duration" "${report_base}.log" | head -10
        fi
    fi
}

# Run performance benchmark
run_performance_benchmark() {
    log "Running performance benchmark suite..."
    
    local benchmark_timestamp=$(date '+%Y%m%d_%H%M%S')
    local benchmark_report="$REPORTS_DIR/benchmark_${benchmark_timestamp}.txt"
    
    {
        echo "GoTAK Performance Benchmark Report"
        echo "Generated: $(date)"
        echo "=================================="
        echo ""
        
        # Quick baseline test
        echo "1. Baseline Performance Test (2 minutes)"
        echo "----------------------------------------"
        
    } > "$benchmark_report"
    
    # Run quick baseline
    if run_load_test "baseline" "2m" "10"; then
        echo "✓ Baseline test completed" >> "$benchmark_report"
    else
        echo "✗ Baseline test failed" >> "$benchmark_report"
    fi
    
    echo "" >> "$benchmark_report"
    
    # Quick stress test
    {
        echo "2. Stress Test (3 minutes)"
        echo "-------------------------"
        
    } >> "$benchmark_report"
    
    if run_load_test "stress" "3m" "25"; then
        echo "✓ Stress test completed" >> "$benchmark_report"
    else
        echo "✗ Stress test failed" >> "$benchmark_report"
    fi
    
    echo "" >> "$benchmark_report"
    
    # WebSocket test
    {
        echo "3. WebSocket Performance Test (2 minutes)"
        echo "----------------------------------------"
        
    } >> "$benchmark_report"
    
    if run_load_test "websocket" "2m" "20"; then
        echo "✓ WebSocket test completed" >> "$benchmark_report"
    else
        echo "✗ WebSocket test failed" >> "$benchmark_report"
    fi
    
    {
        echo ""
        echo "Benchmark completed at: $(date)"
        echo "Full reports available in: $REPORTS_DIR"
        
    } >> "$benchmark_report"
    
    success "Performance benchmark completed"
    echo "Benchmark report: $benchmark_report"
    
    # Show summary
    cat "$benchmark_report"
}

# Main function
main() {
    local command="${1:-benchmark}"
    shift || true
    
    case "$command" in
        "baseline")
            check_prerequisites
            setup_test_environment
            run_load_test "baseline" "$@"
            ;;
        "stress")
            check_prerequisites
            setup_test_environment
            run_load_test "stress" "$@"
            ;;
        "spike")
            check_prerequisites
            setup_test_environment
            run_load_test "spike" "$@"
            ;;
        "websocket")
            check_prerequisites
            setup_test_environment
            run_load_test "websocket" "$@"
            ;;
        "database")
            check_prerequisites
            setup_test_environment
            run_load_test "database" "$@"
            ;;
        "quick")
            check_prerequisites
            setup_test_environment
            run_load_test "quick" "$@"
            ;;
        "all")
            check_prerequisites
            setup_test_environment
            run_load_test "all" "$@"
            ;;
        "benchmark")
            check_prerequisites
            setup_test_environment
            run_performance_benchmark
            ;;
        "help"|*)
            cat << EOF
GoTAK Load Testing Tool

Usage: $0 <command> [duration] [vus]

Commands:
  baseline    Run baseline load test (normal operations)
  stress      Run stress test (high load)
  spike       Run spike test (sudden load increases)
  websocket   Run WebSocket performance test
  database    Run database-intensive test
  quick       Run quick test (2 minutes, low load)
  all         Run all test scenarios
  benchmark   Run performance benchmark suite
  help        Show this help message

Parameters:
  duration    Test duration (e.g., 5m, 30s)
  vus         Number of virtual users (e.g., 10, 50)

Environment Variables:
  GOTAK_BASE_URL   Base URL for HTTP tests (default: http://localhost:8080)
  GOTAK_WS_URL     WebSocket URL (default: ws://localhost:8080)

Examples:
  $0 benchmark          # Run full performance benchmark
  $0 baseline 5m 20     # Run baseline test for 5 minutes with 20 VUs
  $0 stress 10m 100     # Run stress test for 10 minutes with 100 VUs
  $0 quick              # Run quick validation test

Reports are saved to: $REPORTS_DIR/

Prerequisites:
  - k6 load testing tool (https://k6.io/)
  - GoTAK service running and accessible

EOF
            ;;
    esac
}

# Run main function
main "$@"
