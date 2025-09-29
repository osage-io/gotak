#!/bin/sh
# GoTAK Production Health Check Script
set -e

# Configuration
HEALTH_ENDPOINT="${HEALTH_ENDPOINT:-http://localhost:8080/health}"
TAK_PORT="${TAK_PORT:-8087}"
TIMEOUT="${TIMEOUT:-10}"

# Health check function
check_http_health() {
    # Try to reach the health endpoint
    if curl -f -s --max-time "$TIMEOUT" "$HEALTH_ENDPOINT" >/dev/null 2>&1; then
        return 0
    else
        return 1
    fi
}

# Check if TAK port is listening
check_tak_port() {
    if nc -z localhost "$TAK_PORT" 2>/dev/null; then
        return 0
    else
        return 1
    fi
}

# Check if process is running
check_process() {
    if pgrep -f gotak-server >/dev/null 2>&1; then
        return 0
    else
        return 1
    fi
}

# Main health check
main() {
    local failed=0
    
    # Check if process is running
    if ! check_process; then
        echo "UNHEALTHY: GoTAK server process not running"
        failed=1
    fi
    
    # Check HTTP health endpoint
    if ! check_http_health; then
        echo "UNHEALTHY: HTTP health endpoint not responding"
        failed=1
    fi
    
    # Check TAK port
    if ! check_tak_port; then
        echo "UNHEALTHY: TAK port $TAK_PORT not listening"
        failed=1
    fi
    
    if [ $failed -eq 0 ]; then
        echo "HEALTHY: All checks passed"
        exit 0
    else
        exit 1
    fi
}

# Run health check
main "$@"
