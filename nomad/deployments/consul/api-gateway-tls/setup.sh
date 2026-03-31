#!/bin/bash
# Demoland API Gateway TLS Setup Script
# This script:
# 1. Creates the inline-certificate config entry with your wildcard cert
# 2. Applies all Consul config entries
# 3. Runs the Nomad job

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CERT_DIR="${HOME}/sw/demoland-wildcard-certs"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check prerequisites
check_prerequisites() {
    log_info "Checking prerequisites..."
    
    if ! command -v consul &> /dev/null; then
        log_error "consul command not found. Please install Consul."
        exit 1
    fi
    
    if ! command -v nomad &> /dev/null; then
        log_error "nomad command not found. Please install Nomad."
        exit 1
    fi
    
    # Check certificate files exist
    if [[ ! -f "${CERT_DIR}/wildcard_demoland_io_fullchain.crt" ]]; then
        log_error "Certificate file not found: ${CERT_DIR}/wildcard_demoland_io_fullchain.crt"
        exit 1
    fi
    
    if [[ ! -f "${CERT_DIR}/wildcard_demoland_io.key" ]]; then
        log_error "Private key file not found: ${CERT_DIR}/wildcard_demoland_io.key"
        exit 1
    fi
    
    log_info "Prerequisites check passed."
}

# Generate the inline-certificate.hcl with actual certificate content
generate_inline_certificate() {
    log_info "Generating inline-certificate.hcl with certificate content..."
    
    local CERT_CONTENT
    local KEY_CONTENT
    
    # Read certificate and key content
    CERT_CONTENT=$(cat "${CERT_DIR}/wildcard_demoland_io_fullchain.crt")
    KEY_CONTENT=$(cat "${CERT_DIR}/wildcard_demoland_io.key")
    
    # Create the inline-certificate.hcl file
    cat > "${SCRIPT_DIR}/inline-certificate.hcl" <<EOF
# Consul Inline Certificate Configuration Entry
# Auto-generated from wildcard certificate for *.demoland.io
# Generated: $(date)

Kind = "inline-certificate"
Name = "demoland-wildcard-cert"

Certificate = <<CERT
${CERT_CONTENT}
CERT

PrivateKey = <<KEY
${KEY_CONTENT}
KEY
EOF

    log_info "Generated inline-certificate.hcl"
}

# Apply Consul config entries in the correct order
apply_consul_configs() {
    log_info "Applying Consul config entries..."
    
    # 1. First, apply the inline certificate (required by api-gateway)
    log_info "Applying inline-certificate..."
    consul config write "${SCRIPT_DIR}/inline-certificate.hcl"
    
    # 2. Apply the API gateway config
    log_info "Applying api-gateway config..."
    consul config write "${SCRIPT_DIR}/api-gateway.hcl"
    
    # 3. Apply service intentions
    log_info "Applying service intentions..."
    for intention_file in "${SCRIPT_DIR}"/intentions-*.hcl; do
        if [[ -f "$intention_file" ]]; then
            log_info "  Applying $(basename "$intention_file")..."
            consul config write "$intention_file"
        fi
    done
    
    # 4. Apply routes
    log_info "Applying HTTP routes..."
    for route_file in "${SCRIPT_DIR}"/http-route-*.hcl; do
        if [[ -f "$route_file" ]]; then
            log_info "  Applying $(basename "$route_file")..."
            consul config write "$route_file"
        fi
    done
    
    log_info "Applying TCP routes..."
    for route_file in "${SCRIPT_DIR}"/tcp-route-*.hcl; do
        if [[ -f "$route_file" ]]; then
            log_info "  Applying $(basename "$route_file")..."
            consul config write "$route_file"
        fi
    done
    
    log_info "All Consul config entries applied."
}

# Run the Nomad job
run_nomad_job() {
    log_info "Running Nomad job..."
    
    # Check if job already exists
    if nomad job status demoland-api-gateway &> /dev/null; then
        log_warn "Job 'demoland-api-gateway' already exists. Stopping it first..."
        nomad job stop -purge demoland-api-gateway || true
        sleep 2
    fi
    
    nomad job run "${SCRIPT_DIR}/gateway.nomad.hcl"
    
    log_info "Nomad job submitted."
}

# Verify deployment
verify_deployment() {
    log_info "Verifying deployment..."
    
    # Wait a few seconds for the gateway to start
    sleep 5
    
    # Check Nomad job status
    if nomad job status demoland-api-gateway | grep -q "running"; then
        log_info "Nomad job is running."
    else
        log_warn "Nomad job may not be running yet. Check with: nomad job status demoland-api-gateway"
    fi
    
    # Check Consul service
    if consul catalog services | grep -q "demoland-gateway"; then
        log_info "Gateway service registered in Consul."
    else
        log_warn "Gateway service not yet registered. It may take a moment."
    fi
    
    echo ""
    log_info "Deployment complete!"
    echo ""
    echo "Your services should be accessible at:"
    echo "  - https://gotak.demoland.io"
    echo "  - https://lurch.demoland.io"
    echo "  - https://opencode.demoland.io"
    echo "  - CoT TLS: port 8089"
    echo ""
    echo "Useful commands:"
    echo "  nomad job status demoland-api-gateway"
    echo "  consul catalog services"
    echo "  consul config read -kind api-gateway -name demoland-gateway"
    echo "  curl -k https://localhost:443 -H 'Host: gotak.demoland.io'"
}

# Cleanup function (for teardown)
cleanup() {
    log_info "Cleaning up..."
    
    # Stop Nomad job
    nomad job stop -purge demoland-api-gateway 2>/dev/null || true
    
    # Delete Consul config entries (in reverse order)
    consul config delete -kind http-route -name gotak-demoland-route 2>/dev/null || true
    consul config delete -kind http-route -name lurch-demoland-route 2>/dev/null || true
    consul config delete -kind http-route -name opencode-demoland-route 2>/dev/null || true
    consul config delete -kind tcp-route -name gotak-cot-tls-route 2>/dev/null || true
    consul config delete -kind service-intentions -name gotak-api 2>/dev/null || true
    consul config delete -kind service-intentions -name lurch 2>/dev/null || true
    consul config delete -kind service-intentions -name opencode 2>/dev/null || true
    consul config delete -kind service-intentions -name gotak-cot 2>/dev/null || true
    consul config delete -kind api-gateway -name demoland-gateway 2>/dev/null || true
    consul config delete -kind inline-certificate -name demoland-wildcard-cert 2>/dev/null || true
    
    # Remove generated file
    rm -f "${SCRIPT_DIR}/inline-certificate.hcl"
    
    log_info "Cleanup complete."
}

# Main
main() {
    case "${1:-deploy}" in
        deploy)
            check_prerequisites
            generate_inline_certificate
            apply_consul_configs
            run_nomad_job
            verify_deployment
            ;;
        cleanup|teardown)
            cleanup
            ;;
        config-only)
            check_prerequisites
            generate_inline_certificate
            apply_consul_configs
            log_info "Config entries applied. Run 'nomad job run gateway.nomad.hcl' to start the gateway."
            ;;
        *)
            echo "Usage: $0 [deploy|cleanup|config-only]"
            echo ""
            echo "Commands:"
            echo "  deploy      - Full deployment (default)"
            echo "  cleanup     - Remove all resources"
            echo "  config-only - Apply Consul configs only (no Nomad job)"
            exit 1
            ;;
    esac
}

main "$@"
