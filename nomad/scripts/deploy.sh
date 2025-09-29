#!/bin/bash
set -e

# GoTAK Nomad Deployment Script
# Deploys the complete GoTAK stack to a Nomad cluster

# Configuration
NOMAD_ADDR=${NOMAD_ADDR:-"https://localhost:4646"}
NOMAD_SKIP_VERIFY=${NOMAD_SKIP_VERIFY:-"true"}
CONSUL_HTTP_ADDR=${CONSUL_HTTP_ADDR:-"http://localhost:8500"}
VAULT_ADDR=${VAULT_ADDR:-"https://vault.demoland.io:8200"}

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Helper functions
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
    
    # Check Nomad CLI
    if ! command -v nomad &> /dev/null; then
        log_error "Nomad CLI not found. Please install Nomad."
        exit 1
    fi
    
    # Check Consul CLI
    if ! command -v consul &> /dev/null; then
        log_warn "Consul CLI not found. Some features may not work."
    fi
    
    # Check Docker images exist
    if ! docker image inspect gotak-server:latest &> /dev/null; then
        log_error "Docker image 'gotak-server:latest' not found. Please build it first."
        log_info "Run: docker build -t gotak-server:latest ."
        exit 1
    fi
    
    if ! docker image inspect gotak-web:latest &> /dev/null; then
        log_error "Docker image 'gotak-web:latest' not found. Please build it first."
        log_info "Run: cd web && docker build -t gotak-web:latest ."
        exit 1
    fi
    
    log_info "Prerequisites check passed."
}

# Push Docker images to registry
push_images() {
    log_info "Pushing Docker images to registry..."
    
    REGISTRY=${DOCKER_REGISTRY:-"registry.demoland.io"}
    
    # Tag and push gotak-server
    docker tag gotak-server:latest ${REGISTRY}/gotak-server:latest
    docker push ${REGISTRY}/gotak-server:latest
    
    # Tag and push gotak-web
    docker tag gotak-web:latest ${REGISTRY}/gotak-web:latest
    docker push ${REGISTRY}/gotak-web:latest
    
    log_info "Docker images pushed successfully."
}

# Create Nomad namespace
create_namespace() {
    log_info "Creating Nomad namespace..."
    
    nomad namespace apply \
        -description "GoTAK Tactical Awareness Kit" \
        gotak || true
        
    log_info "Namespace created/updated."
}

# Create host volumes on Nomad clients
create_host_volumes() {
    log_info "Creating host volumes on Nomad clients..."
    
    # This needs to be configured in Nomad client configuration
    # Add to /etc/nomad.d/client.hcl on each client:
    cat << EOF
# Add this to your Nomad client configuration on each node:

client {
  host_volume "postgres-data" {
    path      = "/opt/nomad/volumes/postgres"
    read_only = false
  }
  
  host_volume "redis-data" {
    path      = "/opt/nomad/volumes/redis"
    read_only = false
  }
}
EOF
    
    log_warn "Host volumes need to be configured manually on each Nomad client."
    log_warn "See the output above for the configuration to add."
}

# Setup Vault PKI for TLS certificates
setup_vault_pki() {
    log_info "Setting up Vault PKI for TLS certificates..."
    
    if ! command -v vault &> /dev/null; then
        log_warn "Vault CLI not found. Skipping PKI setup."
        return
    fi
    
    # Enable PKI secrets engine
    vault secrets enable -path=pki_int pki || true
    
    # Configure PKI
    vault write pki_int/config/urls \
        issuing_certificates="${VAULT_ADDR}/v1/pki_int/ca" \
        crl_distribution_points="${VAULT_ADDR}/v1/pki_int/crl"
    
    # Create role for GoTAK server
    vault write pki_int/roles/gotak-server \
        allowed_domains="gotak.service.consul,gotak.demoland.io" \
        allow_subdomains=true \
        max_ttl="720h"
    
    # Create policy for Nomad
    vault policy write gotak-tls - <<EOF
path "pki_int/issue/gotak-server" {
  capabilities = ["create", "update"]
}
EOF
    
    log_info "Vault PKI setup completed."
}

# Setup Consul services
setup_consul_services() {
    log_info "Setting up Consul service configurations..."
    
    if ! command -v consul &> /dev/null; then
        log_warn "Consul CLI not found. Skipping Consul setup."
        return
    fi
    
    # Register service defaults
    consul config write - <<EOF
Kind: ServiceDefaults
Name: gotak-server
Protocol: http
EOF

    consul config write - <<EOF
Kind: ServiceDefaults
Name: gotak-web
Protocol: http
EOF
    
    # Create intentions for service communication
    consul intention create -allow gotak-web gotak-server
    consul intention create -allow gotak-server gotak-postgres
    consul intention create -allow gotak-server gotak-redis
    consul intention create -allow gotak-server gotak-nats
    
    log_info "Consul services configured."
}

# Deploy the Nomad job
deploy_job() {
    log_info "Deploying GoTAK stack to Nomad..."
    
    # Update image references in job file if using registry
    if [ ! -z "${DOCKER_REGISTRY}" ]; then
        sed -i.bak "s|gotak-server:latest|${DOCKER_REGISTRY}/gotak-server:latest|g" ../jobs/gotak-stack.nomad.hcl
        sed -i.bak "s|gotak-web:latest|${DOCKER_REGISTRY}/gotak-web:latest|g" ../jobs/gotak-stack.nomad.hcl
    fi
    
    # Plan the deployment
    log_info "Planning deployment..."
    nomad job plan ../jobs/gotak-stack.nomad.hcl
    
    # Run the job
    log_info "Running job..."
    nomad job run -check-index 0 ../jobs/gotak-stack.nomad.hcl
    
    log_info "Deployment initiated."
}

# Monitor deployment status
monitor_deployment() {
    log_info "Monitoring deployment status..."
    
    # Wait for allocation to be running
    sleep 5
    
    # Get job status
    nomad job status gotak-stack
    
    # Get allocation status
    ALLOC_ID=$(nomad job status gotak-stack | grep running | head -1 | awk '{print $1}')
    if [ ! -z "$ALLOC_ID" ]; then
        nomad alloc status $ALLOC_ID
    fi
    
    log_info "Deployment monitoring complete."
}

# Setup ingress (Traefik/Consul API Gateway)
setup_ingress() {
    log_info "Setting up ingress for external access..."
    
    # Check if Traefik is running
    if nomad job status traefik &> /dev/null; then
        log_info "Traefik detected. Routes should be automatically configured via tags."
    else
        log_warn "Traefik not detected. You may need to setup ingress manually."
        
        # Provide manual ingress configuration
        cat << EOF
# If using Consul API Gateway, apply this configuration:

apiVersion: consul.hashicorp.com/v1alpha1
kind: Gateway
metadata:
  name: gotak-gateway
spec:
  gatewayClassName: consul-api-gateway
  listeners:
  - protocol: HTTPS
    port: 443
    name: https
    tls:
      certificateRefs:
      - name: gotak-cert
    allowedRoutes:
      namespaces:
        from: All
---
apiVersion: consul.hashicorp.com/v1alpha1
kind: HTTPRoute
metadata:
  name: gotak-web
spec:
  parentRefs:
  - name: gotak-gateway
  hostnames:
  - gotak.demoland.io
  rules:
  - backendRefs:
    - name: gotak-web
      port: 80
EOF
    fi
}

# Main deployment flow
main() {
    log_info "Starting GoTAK deployment to Nomad cluster..."
    log_info "Nomad Address: ${NOMAD_ADDR}"
    log_info "Consul Address: ${CONSUL_HTTP_ADDR}"
    log_info "Vault Address: ${VAULT_ADDR}"
    
    # Run deployment steps
    check_prerequisites
    
    read -p "Push Docker images to registry? (y/n) " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        push_images
    fi
    
    create_namespace
    create_host_volumes
    
    read -p "Setup Vault PKI? (y/n) " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        setup_vault_pki
    fi
    
    read -p "Setup Consul services? (y/n) " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        setup_consul_services
    fi
    
    deploy_job
    monitor_deployment
    setup_ingress
    
    log_info "Deployment complete!"
    log_info "Access GoTAK at: https://gotak.demoland.io"
    log_info "Jaeger UI at: https://jaeger.gotak.demoland.io"
    
    # Show service endpoints
    echo
    log_info "Service Endpoints:"
    consul catalog services | grep gotak
    
    echo
    log_info "To check job status: nomad job status gotak-stack"
    log_info "To view logs: nomad alloc logs <alloc-id> <task-name>"
}

# Run main function
main "$@"
