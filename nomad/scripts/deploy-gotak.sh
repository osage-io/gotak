#!/bin/bash

# GoTAK Nomad Deployment Script
# Usage: ./deploy-gotak.sh [environment] [options]

set -euo pipefail

# Default values
ENVIRONMENT="${1:-dev}"
DRY_RUN="${2:-false}"
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

# Function to check prerequisites
check_prerequisites() {
    log_info "Checking prerequisites..."
    
    # Check if nomad command is available
    if ! command -v nomad &> /dev/null; then
        log_error "nomad command not found. Please install Nomad CLI."
        exit 1
    fi
    
    # Check if docker command is available
    if ! command -v docker &> /dev/null; then
        log_error "docker command not found. Please install Docker."
        exit 1
    fi
    
    # Test Nomad connection
    if ! nomad node status &> /dev/null; then
        log_error "Cannot connect to Nomad at ${NOMAD_ADDR}"
        log_info "Make sure Nomad is running and NOMAD_ADDR is set correctly"
        exit 1
    fi
    
    log_success "Prerequisites check passed"
}

# Function to validate environment
validate_environment() {
    log_info "Validating environment: ${ENVIRONMENT}"
    
    local env_file="${NOMAD_DIR}/variables/env/${ENVIRONMENT}.hcl"
    if [[ ! -f "${env_file}" ]]; then
        log_error "Environment file not found: ${env_file}"
        log_info "Available environments:"
        find "${NOMAD_DIR}/variables/env" -name "*.hcl" -exec basename {} .hcl \; 2>/dev/null || echo "  No environment files found"
        exit 1
    fi
    
    log_success "Environment ${ENVIRONMENT} is valid"
}

# Function to build Docker images
build_images() {
    log_info "Building Docker images for Nomad deployment..."
    
    cd "${PROJECT_ROOT}"
    
    if [[ "${DRY_RUN}" == "true" ]]; then
        log_info "[DRY RUN] Would build images: gotak/server:latest, gotak/web:latest"
        return
    fi
    
    # Build images using Make
    if ! make docker-build-nomad; then
        log_error "Failed to build Docker images"
        exit 1
    fi
    
    log_success "Docker images built successfully"
}

# Function to create volumes for development
setup_volumes() {
    log_info "Setting up volumes for development..."
    
    if [[ "${DRY_RUN}" == "true" ]]; then
        log_info "[DRY RUN] Would create volumes using create-volumes.sh"
        return
    fi
    
    # Run volume creation script
    "${SCRIPT_DIR}/create-volumes.sh" "${ENVIRONMENT}"
    
    log_warn "Remember to configure host volumes in your Nomad client configuration"
    log_info "Or run Nomad in dev mode with the provided host volume arguments"
}

# Function to validate Nomad jobs
validate_jobs() {
    log_info "Validating Nomad job specifications..."
    
    local var_files=(
        "-var-file=${NOMAD_DIR}/variables/cluster.hcl"
        "-var-file=${NOMAD_DIR}/variables/env/${ENVIRONMENT}.hcl"
    )
    
    local jobs=(
        "postgres"
        "redis"
        "gotak-server"
    )
    
    for job in "${jobs[@]}"; do
        local job_file="${NOMAD_DIR}/jobs/${job}.nomad.hcl"
        log_info "Validating job: ${job}"
        
        if ! nomad job validate "${var_files[@]}" "${job_file}"; then
            log_error "Job validation failed: ${job}"
            exit 1
        fi
    done
    
    log_success "All job specifications are valid"
}

# Function to deploy a single job
deploy_job() {
    local job_name="$1"
    local job_file="${NOMAD_DIR}/jobs/${job_name}.nomad.hcl"
    
    log_info "Deploying job: ${job_name}"
    
    local var_files=(
        "-var-file=${NOMAD_DIR}/variables/cluster.hcl"
        "-var-file=${NOMAD_DIR}/variables/env/${ENVIRONMENT}.hcl"
    )
    
    if [[ "${DRY_RUN}" == "true" ]]; then
        log_info "[DRY RUN] Would deploy: nomad job run ${var_files[*]} ${job_file}"
        return
    fi
    
    if ! nomad job run "${var_files[@]}" "${job_file}"; then
        log_error "Failed to deploy job: ${job_name}"
        exit 1
    fi
    
    log_success "Job deployed: ${job_name}"
}

# Function to check deployment status
check_status() {
    log_info "Checking deployment status..."
    
    local jobs=("gotak-postgres" "gotak-redis" "gotak-server")
    
    echo
    echo "=== Job Status ==="
    for job in "${jobs[@]}"; do
        echo "Job: ${job}"
        nomad job status "${job}" 2>/dev/null | head -10 || echo "  Job not found or failed"
        echo
    done
    
    echo "=== Service Status ==="
    if command -v consul &> /dev/null; then
        echo "Consul services:"
        consul catalog services 2>/dev/null | grep -E "(postgres|redis|gotak-)" || echo "  No services found (Consul not available or no services registered)"
    else
        echo "  Consul CLI not available - cannot check service status"
    fi
}

# Function to show deployment info
show_info() {
    log_info "GoTAK deployment information:"
    echo
    echo "Environment: ${ENVIRONMENT}"
    echo "Nomad Address: ${NOMAD_ADDR}"
    echo
    echo "Services will be available at:"
    echo "  GoTAK Server API: http://localhost:8080"
    echo "  GoTAK CoT TCP: localhost:8087"
    echo "  GoTAK CoT TLS: localhost:8089"
    echo "  PostgreSQL: localhost:5432"
    echo "  Redis: localhost:6379"
    echo
    echo "Access the web UI at: http://localhost:8080"
    echo
    echo "To check logs:"
    echo "  nomad alloc logs -f -job gotak-server"
    echo "  nomad alloc logs -f -job gotak-postgres"
    echo "  nomad alloc logs -f -job gotak-redis"
}

# Function to show help
show_help() {
    echo "GoTAK Nomad Deployment Script"
    echo
    echo "Usage: $0 [environment] [dry-run]"
    echo
    echo "Arguments:"
    echo "  environment    Target environment (dev, staging, prod) [default: dev]"
    echo "  dry-run        Set to 'true' to perform a dry run [default: false]"
    echo
    echo "Environment variables:"
    echo "  NOMAD_ADDR     Nomad server address [default: http://localhost:4646]"
    echo
    echo "Examples:"
    echo "  $0                    # Deploy to dev environment"
    echo "  $0 dev                # Deploy to dev environment"
    echo "  $0 staging            # Deploy to staging environment"
    echo "  $0 dev true           # Dry run for dev environment"
    echo
    echo "Available commands:"
    echo "  $0 help               # Show this help"
    echo "  $0 status             # Check deployment status"
    echo "  $0 stop               # Stop all GoTAK jobs"
    echo "  $0 restart            # Restart all GoTAK jobs"
}

# Function to stop all jobs
stop_jobs() {
    log_info "Stopping GoTAK jobs..."
    
    local jobs=("gotak-server" "gotak-redis" "gotak-postgres")
    
    for job in "${jobs[@]}"; do
        if nomad job status "${job}" &> /dev/null; then
            log_info "Stopping job: ${job}"
            nomad job stop "${job}"
        else
            log_warn "Job not found: ${job}"
        fi
    done
    
    log_success "Stop command completed"
}

# Function to restart all jobs
restart_jobs() {
    log_info "Restarting GoTAK jobs..."
    
    stop_jobs
    sleep 5
    
    # Deploy jobs in dependency order
    deploy_job "postgres"
    sleep 5
    
    deploy_job "redis"
    sleep 3
    
    deploy_job "gotak-server"
    
    log_success "Restart completed"
}

# Main deployment function
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
            stop_jobs
            exit 0
            ;;
        "restart")
            restart_jobs
            exit 0
            ;;
    esac
    
    echo "=== GoTAK Nomad Deployment ==="
    echo
    
    check_prerequisites
    validate_environment
    
    if [[ "${DRY_RUN}" == "true" ]]; then
        log_warn "DRY RUN MODE - No actual changes will be made"
        echo
    fi
    
    build_images
    setup_volumes
    validate_jobs
    
    # Deploy jobs in dependency order
    deploy_job "postgres"
    sleep 5  # Give PostgreSQL time to start
    
    deploy_job "redis" 
    sleep 3  # Give Redis time to start
    
    deploy_job "gotak-server"
    
    log_success "GoTAK deployment completed successfully!"
    
    if [[ "${DRY_RUN}" != "true" ]]; then
        echo
        sleep 10  # Give services time to start
        check_status
        show_info
    fi
}

# Run main function
main "$@"