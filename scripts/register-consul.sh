#!/usr/bin/env bash
set -euo pipefail

# ============================================================================
# Consul Service Registration Script for GoTAK
# ============================================================================
# Registers all GoTAK Docker services with Consul using localhost addresses
# and appropriate health checks (HTTP or TCP).
#
# Usage:
#   ./register-consul.sh           # Register all services
#   ./register-consul.sh --dry-run # Show what would be registered
# ============================================================================

CONSUL_ADDR=${CONSUL_HTTP_ADDR:-http://localhost:8500}
DRY_RUN=false

# Parse command line arguments
for arg in "$@"; do
  case $arg in
    --dry-run)
      DRY_RUN=true
      shift
      ;;
    -h|--help)
      echo "Usage: $0 [--dry-run]"
      echo "  --dry-run  Show what would be registered without making changes"
      exit 0
      ;;
  esac
done

# ============================================================================
# Check Consul availability
# ============================================================================
check_consul() {
  if ! curl -sf "$CONSUL_ADDR/v1/status/leader" &>/dev/null; then
    echo "Error: Consul is not reachable at $CONSUL_ADDR" >&2
    echo "Make sure Consul is running: consul agent -dev" >&2
    exit 1
  fi
  echo "✓ Consul is reachable at $CONSUL_ADDR"
}

# ============================================================================
# Deregister all existing services
# ============================================================================
deregister_all() {
  echo
  echo "=== Deregistering existing services ==="
  local service_ids
  service_ids=$(curl -sf "$CONSUL_ADDR/v1/agent/services" | jq -r '.[].ID' 2>/dev/null || echo "")
  
  if [[ -z "$service_ids" ]]; then
    echo "No services currently registered."
    return
  fi
  
  while IFS= read -r id; do
    if [[ -n "$id" ]]; then
      if [[ "$DRY_RUN" == "true" ]]; then
        echo "[DRY-RUN] Would deregister: $id"
      else
        if curl -sf -X PUT "$CONSUL_ADDR/v1/agent/service/deregister/$id" &>/dev/null; then
          echo "✓ Deregistered: $id"
        else
          echo "✗ Failed to deregister: $id" >&2
        fi
      fi
    fi
  done <<< "$service_ids"
}

# ============================================================================
# Register a service with Consul
# ============================================================================
register_service() {
  local json="$1"
  local service_name="$2"
  
  if [[ "$DRY_RUN" == "true" ]]; then
    echo "[DRY-RUN] Would register: $service_name"
    echo "$json" | jq . 2>/dev/null || echo "$json"
    return 0
  fi
  
  if echo "$json" | curl -sf -X PUT --data-binary @- "$CONSUL_ADDR/v1/agent/service/register" &>/dev/null; then
    echo "✓ Registered: $service_name"
  else
    echo "✗ Failed to register: $service_name" >&2
    return 1
  fi
}

# ============================================================================
# Service Definitions
# ============================================================================
define_services() {
  # PostgreSQL (main dev database)
  POSTGRES_JSON=$(cat <<'EOF'
{
  "Name": "postgres",
  "ID": "postgres-localhost-5432",
  "Address": "127.0.0.1",
  "Port": 5432,
  "Tags": ["db", "postgres", "gotak", "primary"],
  "Check": {
    "TCP": "127.0.0.1:5432",
    "Interval": "15s",
    "Timeout": "3s"
  }
}
EOF
)

  # PostgreSQL (test database)
  POSTGRES_TEST_JSON=$(cat <<'EOF'
{
  "Name": "postgres-test",
  "ID": "postgres-test-localhost-5433",
  "Address": "127.0.0.1",
  "Port": 5433,
  "Tags": ["db", "postgres", "gotak", "test"],
  "Check": {
    "TCP": "127.0.0.1:5433",
    "Interval": "15s",
    "Timeout": "3s"
  }
}
EOF
)

  # Redis
  REDIS_JSON=$(cat <<'EOF'
{
  "Name": "redis",
  "ID": "redis-localhost-6379",
  "Address": "127.0.0.1",
  "Port": 6379,
  "Tags": ["cache", "redis", "gotak"],
  "Check": {
    "TCP": "127.0.0.1:6379",
    "Interval": "15s",
    "Timeout": "3s"
  }
}
EOF
)

  # NATS
  NATS_JSON=$(cat <<'EOF'
{
  "Name": "nats",
  "ID": "nats-localhost-4222",
  "Address": "127.0.0.1",
  "Port": 4222,
  "Tags": ["messaging", "nats", "gotak"],
  "Check": {
    "HTTP": "http://127.0.0.1:8222/healthz",
    "Interval": "15s",
    "Timeout": "5s"
  }
}
EOF
)

  # GoTAK Server - API endpoint
  GOTAK_API_JSON=$(cat <<'EOF'
{
  "Name": "gotak-api",
  "ID": "gotak-api-localhost-8090",
  "Address": "127.0.0.1",
  "Port": 8090,
  "Tags": ["api", "tak", "gotak", "server", "http"],
  "Check": {
    "HTTP": "http://127.0.0.1:8090/health",
    "Interval": "15s",
    "Timeout": "5s"
  }
}
EOF
)

  # GoTAK Server - TAK protocol endpoint
  GOTAK_TAK_JSON=$(cat <<'EOF'
{
  "Name": "gotak-tak",
  "ID": "gotak-tak-localhost-8087",
  "Address": "127.0.0.1",
  "Port": 8087,
  "Tags": ["tak", "protocol", "gotak", "tcp"],
  "Check": {
    "TCP": "127.0.0.1:8087",
    "Interval": "15s",
    "Timeout": "3s"
  }
}
EOF
)

  # GoTAK Web UI
  GOTAK_WEB_JSON=$(cat <<'EOF'
{
  "Name": "gotak-web",
  "ID": "gotak-web-localhost-8080",
  "Address": "127.0.0.1",
  "Port": 8080,
  "Tags": ["web", "ui", "gotak", "frontend", "http"],
  "Check": {
    "HTTP": "http://127.0.0.1:8080/",
    "Interval": "30s",
    "Timeout": "5s"
  }
}
EOF
)

  # Jaeger UI
  JAEGER_JSON=$(cat <<'EOF'
{
  "Name": "jaeger-ui",
  "ID": "jaeger-ui-localhost-16686",
  "Address": "127.0.0.1",
  "Port": 16686,
  "Tags": ["tracing", "jaeger", "monitoring", "gotak", "ui"],
  "Check": {
    "HTTP": "http://127.0.0.1:16686/",
    "Interval": "30s",
    "Timeout": "5s"
  }
}
EOF
)

  # Adminer
  ADMINER_JSON=$(cat <<'EOF'
{
  "Name": "adminer",
  "ID": "adminer-localhost-8081",
  "Address": "127.0.0.1",
  "Port": 8081,
  "Tags": ["db", "admin", "ui", "gotak", "http"],
  "Check": {
    "HTTP": "http://127.0.0.1:8081/",
    "Interval": "30s",
    "Timeout": "5s"
  }
}
EOF
)
}

# ============================================================================
# Main execution
# ============================================================================
main() {
  echo "=== GoTAK Consul Service Registration ==="
  echo "Consul Address: $CONSUL_ADDR"
  [[ "$DRY_RUN" == "true" ]] && echo "Mode: DRY RUN (no changes will be made)"
  
  check_consul
  deregister_all
  
  echo
  echo "=== Registering services ==="
  define_services
  
  register_service "$POSTGRES_JSON" "postgres (localhost:5432)"
  register_service "$POSTGRES_TEST_JSON" "postgres-test (localhost:5433)"
  register_service "$REDIS_JSON" "redis (localhost:6379)"
  register_service "$NATS_JSON" "nats (localhost:4222)"
  register_service "$GOTAK_API_JSON" "gotak-api (localhost:8090)"
  register_service "$GOTAK_TAK_JSON" "gotak-tak (localhost:8087)"
  register_service "$GOTAK_WEB_JSON" "gotak-web (localhost:8080)"
  register_service "$JAEGER_JSON" "jaeger-ui (localhost:16686)"
  register_service "$ADMINER_JSON" "adminer (localhost:8081)"
  
  echo
  echo "=== Registration complete ==="
  echo "Visit Consul UI at: $CONSUL_ADDR/ui"
  
  if [[ "$DRY_RUN" != "true" ]]; then
    echo
    echo "Checking service health status in 5 seconds..."
    sleep 5
    
    critical_count=$(curl -sf "$CONSUL_ADDR/v1/health/state/critical" | jq length 2>/dev/null || echo "unknown")
    passing_count=$(curl -sf "$CONSUL_ADDR/v1/health/state/passing" | jq length 2>/dev/null || echo "unknown")
    
    echo "Health status: $passing_count passing, $critical_count critical"
    
    if [[ "$critical_count" != "0" && "$critical_count" != "unknown" ]]; then
      echo
      echo "⚠ Warning: Some services are not healthy yet. They may still be starting up."
      echo "Run this command to check status:"
      echo "  curl -s localhost:8500/v1/health/state/critical | jq -r '.[] | \"\\(.ServiceName): \\(.Output)\"'"
    fi
  fi
}

main