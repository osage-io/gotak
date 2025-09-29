#!/bin/bash

# GoTAK Volume Creation Script for Development
# Creates host volumes for PostgreSQL and Redis data persistence

set -euo pipefail

# Configuration
VOLUME_ROOT="/tmp/gotak-nomad-volumes"
ENVIRONMENT="${1:-dev}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

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

# Create volume directories
create_directories() {
    log_info "Creating volume directories for environment: ${ENVIRONMENT}"
    
    local env_root="${VOLUME_ROOT}/${ENVIRONMENT}"
    
    # Create directory structure
    mkdir -p "${env_root}/postgres-data"
    mkdir -p "${env_root}/redis-data"
    
    # Set permissions
    chmod 755 "${env_root}"
    chmod 755 "${env_root}/postgres-data"
    chmod 755 "${env_root}/redis-data"
    
    log_success "Volume directories created in: ${env_root}"
}

# Register volumes with Nomad (requires restart of Nomad agents)
show_nomad_config() {
    log_info "Add the following to your Nomad client configuration:"
    echo
    cat << EOF
# Add to /etc/nomad.d/client.hcl or ~/.nomad.d/client.hcl

client {
  host_volume "gotak-postgres-data" {
    path      = "${VOLUME_ROOT}/${ENVIRONMENT}/postgres-data"
    read_only = false
  }
  
  host_volume "gotak-redis-data" {
    path      = "${VOLUME_ROOT}/${ENVIRONMENT}/redis-data" 
    read_only = false
  }
}
EOF
    echo
    log_warn "After adding this configuration, restart your Nomad agent for the volumes to be available."
    log_info "For development, you can also run Nomad in dev mode with host volumes:"
    echo
    echo "nomad agent -dev -bind=127.0.0.1 -log-level=INFO \\"
    echo "  -client-host-volume=gotak-postgres-data=${VOLUME_ROOT}/${ENVIRONMENT}/postgres-data \\"
    echo "  -client-host-volume=gotak-redis-data=${VOLUME_ROOT}/${ENVIRONMENT}/redis-data"
}

# Alternative: Create Nomad CSI volumes (if CSI plugin is available)
create_csi_volumes() {
    log_info "Creating CSI volumes for environment: ${ENVIRONMENT}"
    
    # Check if CSI plugin is available
    if ! nomad plugin status 2>/dev/null | grep -q "hostpath"; then
        log_warn "No CSI plugins detected. Skipping CSI volume creation."
        return
    fi
    
    # PostgreSQL volume
    cat << EOF > /tmp/postgres-volume.hcl
id           = "gotak-postgres-${ENVIRONMENT}"
name         = "gotak-postgres-${ENVIRONMENT}"
type         = "csi"
plugin_id    = "hostpath"
capacity_max = "2G"
capacity_min = "1G"

capability {
  access_mode     = "single-node-writer"
  attachment_mode = "file-system"
}

mount_options {
  fs_type = "ext4"
}
EOF

    # Redis volume  
    cat << EOF > /tmp/redis-volume.hcl
id           = "gotak-redis-${ENVIRONMENT}"
name         = "gotak-redis-${ENVIRONMENT}"
type         = "csi"
plugin_id    = "hostpath"
capacity_max = "500M"
capacity_min = "100M"

capability {
  access_mode     = "single-node-writer"
  attachment_mode = "file-system"
}

mount_options {
  fs_type = "ext4"
}
EOF

    # Create the volumes
    if nomad volume create /tmp/postgres-volume.hcl; then
        log_success "PostgreSQL CSI volume created"
    else
        log_error "Failed to create PostgreSQL CSI volume"
    fi
    
    if nomad volume create /tmp/redis-volume.hcl; then
        log_success "Redis CSI volume created"
    else
        log_error "Failed to create Redis CSI volume"
    fi
    
    # Clean up
    rm -f /tmp/postgres-volume.hcl /tmp/redis-volume.hcl
}

# Show volume status
show_status() {
    log_info "Volume status:"
    echo
    
    # Check directories
    if [[ -d "${VOLUME_ROOT}/${ENVIRONMENT}" ]]; then
        echo "Host directories:"
        ls -la "${VOLUME_ROOT}/${ENVIRONMENT}/"
        echo
    fi
    
    # Check Nomad volumes
    if command -v nomad &> /dev/null; then
        echo "Nomad volumes:"
        nomad volume status 2>/dev/null | grep -i gotak || echo "  No Nomad volumes found"
    fi
}

# Main function
main() {
    case "${1:-create}" in
        "help"|"-h"|"--help")
            echo "GoTAK Volume Creation Script"
            echo
            echo "Usage: $0 [environment] [command]"
            echo
            echo "Commands:"
            echo "  create    Create host volume directories (default)"
            echo "  csi       Create CSI volumes (requires CSI plugin)"
            echo "  config    Show Nomad configuration for host volumes"
            echo "  status    Show volume status"
            echo "  help      Show this help"
            echo
            echo "Examples:"
            echo "  $0 dev        # Create volumes for dev environment"
            echo "  $0 staging    # Create volumes for staging environment"
            echo "  $0 dev csi    # Create CSI volumes for dev environment"
            exit 0
            ;;
        "csi")
            create_csi_volumes
            ;;
        "config")
            show_nomad_config
            ;;
        "status")
            show_status
            ;;
        *)
            create_directories
            echo
            show_nomad_config
            ;;
    esac
}

# Check if environment is provided as second argument for commands
if [[ $# -gt 1 ]]; then
    ENVIRONMENT="$1"
    shift
fi

main "$@"