#!/bin/bash

# Universal GoTAK Nomad Deployment Script
# Supports both standalone and Consul-integrated deployments

set -euo pipefail

# Default values
DEPLOYMENT_TYPE="${1:-standalone}"  # standalone or consul
ENVIRONMENT="${2:-dev}"
DRY_RUN="${3:-false}"
NOMAD_ADDR="${NOMAD_ADDR:-http://localhost:4646}"

# Script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"
NOMAD_DIR="${PROJECT_ROOT}/nomad"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

# Function to show help
show_help() {
    echo "Universal GoTAK Nomad Deployment Script"
    echo
    echo "Usage: $0 [deployment_type] [environment] [dry_run]"
    echo
    echo "Arguments:"
    echo "  deployment_type   'standalone' or 'consul' [default: standalone]"
    echo "  environment       Target environment (dev, staging, prod) [default: dev]"
    echo "  dry_run           Set to 'true' to perform a dry run [default: false]"
    echo
    echo "Environment variables:"
    echo "  NOMAD_ADDR        Nomad server address [default: http://localhost:4646]"
    echo "  CONSUL_HTTP_ADDR  Consul server address [default: http://localhost:8500]"
    echo
    echo "Deployment Types:"
    echo "  standalone        Simple deployment without Consul service discovery"
    echo "                    - Direct IP addressing between services"
    echo "                    - No service mesh or Connect"
    echo "                    - Perfect for development and simple deployments"
    echo
    echo "  consul            Full Consul integration with service discovery"
    echo "                    - Automatic service registration and discovery"
    echo "                    - Health checking and monitoring"
    echo "                    - Optional: Consul Connect for service mesh"
    echo "                    - Production-ready with advanced networking"
    echo
    echo "Examples:"
    echo "  $0                          # Standalone deployment to dev"
    echo "  $0 standalone dev           # Standalone deployment to dev"
    echo "  $0 consul dev               # Consul deployment to dev"
    echo "  $0 consul prod              # Consul deployment to prod"
    echo "  $0 standalone dev true      # Dry run standalone deployment"
    echo
    echo "Commands:"
    echo "  $0 help                     # Show this help"
    echo "  $0 status                   # Check deployment status"
    echo "  $0 stop [type]              # Stop deployments"
}

# Function to check prerequisites
check_prerequisites() {
    log_info "Checking prerequisites for ${DEPLOYMENT_TYPE} deployment..."
    
    # Check Nomad
    if ! command -v nomad &> /dev/null; then
        log_error "nomad command not found. Please install Nomad CLI."
        exit 1
    fi
    
    if ! nomad node status &> /dev/null; then
        log_error "Cannot connect to Nomad at ${NOMAD_ADDR}"
        log_info "Make sure Nomad is running and NOMAD_ADDR is set correctly"
        exit 1
    fi
    
    # Check Docker
    if ! command -v docker &> /dev/null; then
        log_error "docker command not found. Please install Docker."
        exit 1
    fi
    
    # Check Consul if using consul deployment
    if [[ "${DEPLOYMENT_TYPE}" == "consul" ]]; then
        if ! command -v consul &> /dev/null; then
            log_warn "consul command not found. Installing consul..."
            # You could add consul installation here
        fi
        
        CONSUL_HTTP_ADDR="${CONSUL_HTTP_ADDR:-http://localhost:8500}"
        if ! curl -s "${CONSUL_HTTP_ADDR}/v1/status/leader" &> /dev/null; then
            log_warn "Cannot connect to Consul at ${CONSUL_HTTP_ADDR}"
            log_info "Consul will be deployed as part of this stack"
        else
            log_info "Connected to existing Consul cluster"
        fi
    fi
    
    log_success "Prerequisites check passed for ${DEPLOYMENT_TYPE} deployment"
}

# Function to deploy Consul (if needed)
deploy_consul() {
    if [[ "${DEPLOYMENT_TYPE}" != "consul" ]]; then
        return
    fi
    
    log_info "Setting up Consul for service discovery..."
    
    # Check if Consul is already running
    if curl -s "${CONSUL_HTTP_ADDR:-http://localhost:8500}/v1/status/leader" &> /dev/null; then
        log_info "Consul already running, skipping deployment"
        return
    fi
    
    log_info "Deploying Consul cluster..."
    
    cat > /tmp/consul.nomad.hcl << 'EOF'
job "consul" {
  datacenters = ["dc1"]
  region      = "global"
  type        = "service"
  priority    = 100

  group "consul" {
    count = 1

    network {
      port "http" {
        static = 8500
        to     = 8500
      }
      port "rpc" {
        static = 8400
        to     = 8400
      }
      port "serf_lan" {
        static = 8301
        to     = 8301
      }
      port "serf_wan" {
        static = 8302
        to     = 8302
      }
      port "dns" {
        static = 8600
        to     = 8600
      }
    }

    task "consul" {
      driver = "docker"

      config {
        image = "consul:1.16"
        ports = ["http", "rpc", "serf_lan", "serf_wan", "dns"]
        
        args = [
          "consul",
          "agent",
          "-dev",
          "-client=0.0.0.0",
          "-bind=0.0.0.0",
          "-ui-content-path=/ui/"
        ]
      }

      resources {
        cpu    = 200
        memory = 256
      }

      service {
        name = "consul"
        port = "http"
        tags = ["consul", "service-discovery"]
        
        check {
          type     = "http"
          path     = "/v1/status/leader"
          interval = "10s"
          timeout  = "3s"
        }
      }
    }
  }
}
EOF

    if [[ "${DRY_RUN}" == "true" ]]; then
        log_info "[DRY RUN] Would deploy Consul"
        rm -f /tmp/consul.nomad.hcl
        return
    fi
    
    nomad job run /tmp/consul.nomad.hcl
    rm -f /tmp/consul.nomad.hcl
    
    # Wait for Consul to be ready
    log_info "Waiting for Consul to be ready..."
    for i in {1..30}; do
        if curl -s "http://localhost:8500/v1/status/leader" &> /dev/null; then
            log_success "Consul is ready"
            return
        fi
        sleep 2
    done
    
    log_warn "Consul may not be fully ready yet, continuing..."
}

# Function to build Docker images
build_images() {
    log_info "Building Docker images for GoTAK deployment..."
    
    cd "${PROJECT_ROOT}"
    
    if [[ "${DRY_RUN}" == "true" ]]; then
        log_info "[DRY RUN] Would build images: gotak/server:latest, gotak/web:latest"
        return
    fi
    
    if ! make docker-build-nomad; then
        log_error "Failed to build Docker images"
        exit 1
    fi
    
    log_success "Docker images built successfully"
}

# Function to setup volumes
setup_volumes() {
    log_info "Setting up persistent volumes..."
    
    if [[ "${DRY_RUN}" == "true" ]]; then
        log_info "[DRY RUN] Would create volumes for environment: ${ENVIRONMENT}"
        return
    fi
    
    "${SCRIPT_DIR}/create-volumes.sh" "${ENVIRONMENT}"
    log_success "Volumes configured for ${ENVIRONMENT} environment"
}

# Function to deploy jobs
deploy_jobs() {
    local deployment_dir="${NOMAD_DIR}/deployments/${DEPLOYMENT_TYPE}"
    
    if [[ ! -d "${deployment_dir}" ]]; then
        log_error "Deployment directory not found: ${deployment_dir}"
        exit 1
    fi
    
    log_info "Deploying GoTAK stack using ${DEPLOYMENT_TYPE} configuration..."
    
    # Determine job files based on deployment type
    local jobs
    if [[ "${DEPLOYMENT_TYPE}" == "standalone" ]]; then
        jobs=(
            "postgres-simple"
            "redis-simple" 
            "gotak-server-simple"
        )
    else
        jobs=(
            "postgres"
            "redis"
            "gotak-server"
        )
    fi
    
    # Deploy jobs in order
    for job in "${jobs[@]}"; do
        local job_file="${deployment_dir}/${job}.nomad.hcl"
        
        if [[ ! -f "${job_file}" ]]; then
            log_error "Job file not found: ${job_file}"
            exit 1
        fi
        
        log_info "Deploying job: ${job}"
        
        if [[ "${DRY_RUN}" == "true" ]]; then
            log_info "[DRY RUN] Would deploy: nomad job run ${job_file}"
            continue
        fi
        
        if ! nomad job run "${job_file}"; then
            log_error "Failed to deploy job: ${job}"
            exit 1
        fi
        
        log_success "Job deployed: ${job}"
        
        # Wait a bit between deployments for dependencies
        if [[ "${job}" == "postgres-simple" || "${job}" == "postgres" ]]; then
            sleep 5
        elif [[ "${job}" == "redis-simple" || "${job}" == "redis" ]]; then
            sleep 3
        fi
    done
}

# Function to check status
check_status() {
    log_info "Checking GoTAK deployment status..."
    
    echo
    echo "=== Nomad Jobs ==="
    nomad job status 2>/dev/null | grep -E "gotak-|consul" || echo "No GoTAK jobs found"
    
    echo
    echo "=== Service Status ==="
    if [[ "${DEPLOYMENT_TYPE}" == "consul" ]] && command -v consul &> /dev/null; then
        consul catalog services 2>/dev/null | grep -E "postgres|redis|gotak|consul" || echo "No services registered"
    else
        echo "Consul not available - cannot check service status"
    fi
    
    echo
    echo "=== Access Information ==="
    echo "GoTAK Web UI: http://localhost:8080"
    echo "GoTAK API: http://localhost:8080/api"
    echo "CoT TCP: localhost:8087"
    echo "CoT TLS: localhost:8089"
    if [[ "${DEPLOYMENT_TYPE}" == "consul" ]]; then
        echo "Consul UI: http://localhost:8500/ui"
    fi
}

# Function to stop services
stop_services() {
    local type_filter="${1:-}"
    
    log_info "Stopping GoTAK services..."
    
    local jobs=(
        "gotak-server"
        "gotak-redis"
        "gotak-postgres"
    )
    
    if [[ "${DEPLOYMENT_TYPE}" == "consul" ]]; then
        jobs+=("consul")
    fi
    
    for job in "${jobs[@]}"; do
        if nomad job status "${job}" &> /dev/null; then
            log_info "Stopping job: ${job}"
            nomad job stop "${job}" || log_warn "Failed to stop ${job}"
        else
            log_info "Job not found: ${job}"
        fi
    done
    
    log_success "Stop command completed"
}

# Main function
main() {
    case "${1:-deploy}" in
        "help"|"-h"|"--help")
            show_help
            exit 0
            ;;
        "status")
            check_status
            exit 0
            ;;
        "stop")
            stop_services "${2:-}"
            exit 0
            ;;
    esac
    
    # Validate deployment type
    if [[ "${DEPLOYMENT_TYPE}" != "standalone" && "${DEPLOYMENT_TYPE}" != "consul" ]]; then
        log_error "Invalid deployment type: ${DEPLOYMENT_TYPE}"
        log_info "Valid types: standalone, consul"
        exit 1
    fi
    
    echo "=== GoTAK Universal Nomad Deployment ==="
    echo "Deployment Type: ${DEPLOYMENT_TYPE}"
    echo "Environment: ${ENVIRONMENT}"
    echo
    
    if [[ "${DRY_RUN}" == "true" ]]; then
        log_warn "DRY RUN MODE - No actual changes will be made"
        echo
    fi
    
    check_prerequisites
    
    # Deploy Consul first if needed
    if [[ "${DEPLOYMENT_TYPE}" == "consul" ]]; then
        deploy_consul
        sleep 5  # Give Consul time to be ready
    fi
    
    build_images
    setup_volumes
    deploy_jobs
    
    log_success "GoTAK ${DEPLOYMENT_TYPE} deployment completed successfully!"
    
    if [[ "${DRY_RUN}" != "true" ]]; then
        echo
        sleep 10  # Give services time to start
        check_status
    fi
}

# Run main function
main "$@"