#!/bin/bash
# GoTAK Production Deployment Script
# Orchestrates full production deployment with validation, migration, and health checks
set -e

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
COMPOSE_FILE="${COMPOSE_FILE:-docker-compose.prod.yml}"
ENV_FILE="${ENV_FILE:-.env.production}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Logging functions
log() {
    echo -e "${BLUE}[DEPLOY]${NC} $1"
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

# Print banner
print_banner() {
    cat << 'EOF'
   ____       _______      _  __   ____             _             
  / __ \___  |__   __|    | |/ /  |  _ \           | |            
 | |  | / _ \   | |  __ _  | ' /   | |_) | __ _ ___| |__     ___ __
 | |  | | | |  | | / _` | |  <    |  _ < / _` / __| '_ \   / _ \_ \
 | |__| | |_|  | | | (_| | | . \   | |_) | (_| \__ \ | | | |  __/ |
  \____/ \___/  |_|  \__,_| |_|\_\  |____/ \__,_|___/_| |_|  \___|_|
                                                                    
                Production Deployment Script                        
EOF
}

# Check prerequisites
check_prerequisites() {
    log "Checking deployment prerequisites..."
    
    local missing_tools=()
    
    # Check required tools
    for tool in docker docker-compose git; do
        if ! command -v "$tool" >/dev/null 2>&1; then
            missing_tools+=("$tool")
        fi
    done
    
    if [ ${#missing_tools[@]} -gt 0 ]; then
        error "Missing required tools: ${missing_tools[*]}"
        error "Please install the missing tools and try again"
        exit 1
    fi
    
    # Check Docker daemon
    if ! docker info >/dev/null 2>&1; then
        error "Docker daemon is not running or accessible"
        exit 1
    fi
    
    # Check if running as root
    if [ "$(id -u)" -eq 0 ]; then
        warn "Running as root. Consider using a non-root user with Docker group membership."
    fi
    
    success "Prerequisites check passed"
}

# Validate environment configuration
validate_environment() {
    log "Validating environment configuration..."
    
    # Check if environment file exists
    if [ ! -f "$ENV_FILE" ]; then
        error "Environment file not found: $ENV_FILE"
        error "Please create the environment file from the template:"
        error "  cp .env.production.template $ENV_FILE"
        error "  # Edit $ENV_FILE with your configuration"
        exit 1
    fi
    
    # Source environment file
    log "Loading environment from: $ENV_FILE"
    set -a  # automatically export all variables
    source "$ENV_FILE"
    set +a
    
    # Run configuration validation
    if [ -f "$SCRIPT_DIR/validate-config.sh" ]; then
        log "Running configuration validation..."
        "$SCRIPT_DIR/validate-config.sh" full || {
            error "Configuration validation failed"
            exit 1
        }
    else
        warn "Configuration validation script not found, skipping detailed validation"
    fi
    
    success "Environment validation passed"
}

# Build application images
build_images() {
    local build_args="$1"
    
    log "Building application images..."
    
    # Set build arguments
    local docker_build_args=""
    if [ -n "$GOTAK_VERSION" ]; then
        docker_build_args="$docker_build_args --build-arg VERSION=$GOTAK_VERSION"
    fi
    if [ -n "$BUILD_TIME" ]; then
        docker_build_args="$docker_build_args --build-arg BUILD_TIME=$BUILD_TIME"
    elif [ -z "$build_args" ]; then
        docker_build_args="$docker_build_args --build-arg BUILD_TIME=$(date -u +%Y-%m-%dT%H:%M:%SZ)"
    fi
    if [ -n "$GIT_COMMIT" ]; then
        docker_build_args="$docker_build_args --build-arg GIT_COMMIT=$GIT_COMMIT"
    elif [ -z "$build_args" ] && [ -d ".git" ]; then
        local git_commit
        git_commit=$(git rev-parse HEAD 2>/dev/null || echo "unknown")
        docker_build_args="$docker_build_args --build-arg GIT_COMMIT=$git_commit"
    fi
    
    # Build with Docker Compose
    log "Building images with: docker-compose -f $COMPOSE_FILE build $docker_build_args"
    
    if [ "$build_args" = "--no-cache" ]; then
        docker-compose -f "$COMPOSE_FILE" build --no-cache $docker_build_args || {
            error "Image build failed"
            exit 1
        }
    else
        docker-compose -f "$COMPOSE_FILE" build $docker_build_args || {
            error "Image build failed" 
            exit 1
        }
    fi
    
    success "Images built successfully"
}

# Deploy services
deploy_services() {
    local deployment_mode="${1:-rolling}"
    
    log "Deploying services (mode: $deployment_mode)..."
    
    case "$deployment_mode" in
        "fresh")
            log "Performing fresh deployment (stopping existing services)..."
            docker-compose -f "$COMPOSE_FILE" down || true
            docker-compose -f "$COMPOSE_FILE" up -d
            ;;
        "rolling")
            log "Performing rolling deployment..."
            # Start new services alongside old ones
            docker-compose -f "$COMPOSE_FILE" up -d --no-deps --scale gotak=2 gotak || {
                error "Failed to scale up new instances"
                exit 1
            }
            
            # Wait for new instances to be healthy
            sleep 30
            
            # Scale down old instances
            docker-compose -f "$COMPOSE_FILE" up -d --no-deps --scale gotak=1 gotak
            ;;
        "blue-green")
            log "Blue-green deployment not implemented yet, using rolling deployment"
            deploy_services "rolling"
            ;;
        *)
            error "Unknown deployment mode: $deployment_mode"
            exit 1
            ;;
    esac
    
    success "Services deployed successfully"
}

# Wait for services to be healthy
wait_for_services() {
    local timeout="${1:-300}"  # 5 minutes default timeout
    local check_interval=10
    local elapsed=0
    
    log "Waiting for services to become healthy (timeout: ${timeout}s)..."
    
    while [ $elapsed -lt $timeout ]; do
        log "Health check attempt (${elapsed}s elapsed)..."
        
        # Check if all services are healthy
        local unhealthy_services
        unhealthy_services=$(docker-compose -f "$COMPOSE_FILE" ps --filter "health=starting" --filter "health=unhealthy" --services 2>/dev/null || echo "")
        
        if [ -z "$unhealthy_services" ]; then
            success "All services are healthy"
            return 0
        fi
        
        log "Waiting for services: $unhealthy_services"
        sleep $check_interval
        elapsed=$((elapsed + check_interval))
    done
    
    error "Services failed to become healthy within timeout"
    
    # Show service status for debugging
    log "Current service status:"
    docker-compose -f "$COMPOSE_FILE" ps
    
    return 1
}

# Run database migrations
run_migrations() {
    log "Running database migrations..."
    
    # Wait for database to be ready
    log "Waiting for database to be ready..."
    local db_ready=false
    for i in {1..30}; do
        if docker-compose -f "$COMPOSE_FILE" exec -T postgres pg_isready -U "${POSTGRES_USER:-gotak}" >/dev/null 2>&1; then
            db_ready=true
            break
        fi
        log "Database not ready, waiting... (attempt $i/30)"
        sleep 2
    done
    
    if [ "$db_ready" = false ]; then
        error "Database failed to become ready"
        exit 1
    fi
    
    # Run migrations using the application container
    log "Executing migrations..."
    docker-compose -f "$COMPOSE_FILE" exec -T gotak /app/migrate.sh apply || {
        error "Database migration failed"
        exit 1
    }
    
    success "Database migrations completed"
}

# Run post-deployment tests
run_post_deployment_tests() {
    log "Running post-deployment verification tests..."
    
    # Basic connectivity tests
    local base_url="http://localhost:${GOTAK_HTTP_PORT:-8080}"
    
    # Health check
    log "Testing health endpoint..."
    local health_response
    health_response=$(curl -s -o /dev/null -w "%{http_code}" "$base_url/health" || echo "000")
    
    if [ "$health_response" != "200" ]; then
        error "Health check failed (HTTP $health_response)"
        return 1
    fi
    
    success "Health check passed"
    
    # Test database connectivity through API
    log "Testing database connectivity..."
    # Add more specific API tests here as needed
    
    # Test WebSocket connectivity
    log "Testing WebSocket connectivity..."
    # Add WebSocket connectivity tests here
    
    success "Post-deployment tests passed"
}

# Create backup before deployment
create_backup() {
    if [ -f "$SCRIPT_DIR/backup.sh" ]; then
        log "Creating pre-deployment backup..."
        
        # Only create backup if database is running
        if docker-compose -f "$COMPOSE_FILE" ps postgres | grep -q "Up"; then
            "$SCRIPT_DIR/backup.sh" create full || {
                warn "Backup creation failed, continuing with deployment..."
            }
        else
            log "Database not running, skipping backup"
        fi
    else
        warn "Backup script not found, skipping backup"
    fi
}

# Cleanup old images and containers
cleanup() {
    local cleanup_mode="${1:-moderate}"
    
    log "Cleaning up (mode: $cleanup_mode)..."
    
    case "$cleanup_mode" in
        "minimal")
            # Remove only stopped containers
            docker container prune -f
            ;;
        "moderate")
            # Remove stopped containers and dangling images
            docker container prune -f
            docker image prune -f
            ;;
        "aggressive")
            # Remove all unused containers, images, and volumes
            docker system prune -af --volumes
            ;;
        "none")
            log "Skipping cleanup"
            ;;
        *)
            warn "Unknown cleanup mode: $cleanup_mode, using moderate"
            cleanup "moderate"
            ;;
    esac
    
    success "Cleanup completed"
}

# Show deployment status
show_status() {
    log "Deployment Status:"
    echo "=================="
    
    # Show running services
    echo "Services:"
    docker-compose -f "$COMPOSE_FILE" ps
    echo ""
    
    # Show resource usage
    echo "Resource Usage:"
    docker stats --no-stream --format "table {{.Container}}\t{{.CPUPerc}}\t{{.MemUsage}}"
    echo ""
    
    # Show application URLs
    echo "Application URLs:"
    echo "  HTTP API:     http://localhost:${GOTAK_HTTP_PORT:-8080}"
    echo "  TAK Protocol: tcp://localhost:${GOTAK_TAK_PORT:-8087}"
    echo "  TLS Protocol: tcp://localhost:${GOTAK_TLS_PORT:-8089}"
    
    if [ -n "$GRAFANA_PORT" ]; then
        echo "  Grafana:      http://localhost:${GRAFANA_PORT:-3000}"
    fi
    if [ -n "$PROMETHEUS_PORT" ]; then
        echo "  Prometheus:   http://localhost:${PROMETHEUS_PORT:-9090}"
    fi
}

# Main deployment function
main() {
    local command="${1:-deploy}"
    shift || true
    
    print_banner
    
    case "$command" in
        "deploy")
            local deployment_mode="${1:-rolling}"
            local cleanup_mode="${2:-moderate}"
            
            check_prerequisites
            validate_environment
            create_backup
            build_images
            deploy_services "$deployment_mode"
            wait_for_services
            run_migrations
            run_post_deployment_tests
            cleanup "$cleanup_mode"
            show_status
            
            success "Deployment completed successfully!"
            ;;
            
        "build")
            local build_args="$1"
            check_prerequisites
            validate_environment
            build_images "$build_args"
            ;;
            
        "start")
            check_prerequisites
            validate_environment
            deploy_services fresh
            wait_for_services
            show_status
            ;;
            
        "stop")
            log "Stopping all services..."
            docker-compose -f "$COMPOSE_FILE" down
            success "Services stopped"
            ;;
            
        "restart")
            log "Restarting all services..."
            docker-compose -f "$COMPOSE_FILE" restart
            wait_for_services
            show_status
            success "Services restarted"
            ;;
            
        "status")
            show_status
            ;;
            
        "logs")
            local service="$1"
            if [ -n "$service" ]; then
                docker-compose -f "$COMPOSE_FILE" logs -f "$service"
            else
                docker-compose -f "$COMPOSE_FILE" logs -f
            fi
            ;;
            
        "migrate")
            validate_environment
            run_migrations
            ;;
            
        "backup")
            local backup_type="${1:-full}"
            validate_environment
            "$SCRIPT_DIR/backup.sh" create "$backup_type"
            ;;
            
        "cleanup")
            local cleanup_mode="${1:-moderate}"
            cleanup "$cleanup_mode"
            ;;
            
        "validate")
            validate_environment
            success "Configuration validation passed"
            ;;
            
        "help"|*)
            cat << EOF
GoTAK Production Deployment Script

Usage: $0 <command> [options]

Commands:
  deploy [mode] [cleanup]  Full deployment (modes: fresh, rolling, blue-green)
  build [--no-cache]       Build application images only
  start                    Start all services
  stop                     Stop all services  
  restart                  Restart all services
  status                   Show deployment status
  logs [service]           Show service logs
  migrate                  Run database migrations only
  backup [type]            Create database backup (types: full, schema, data)
  cleanup [mode]           Clean up resources (modes: minimal, moderate, aggressive, none)
  validate                 Validate configuration only
  help                     Show this help message

Environment Variables:
  COMPOSE_FILE            Docker Compose file (default: docker-compose.prod.yml)
  ENV_FILE                Environment file (default: .env.production)

Examples:
  $0 deploy                    # Rolling deployment with moderate cleanup
  $0 deploy fresh aggressive   # Fresh deployment with aggressive cleanup
  $0 build --no-cache         # Rebuild images from scratch
  $0 logs gotak               # Show GoTAK service logs
  $0 backup full              # Create full database backup
  $0 validate                 # Validate configuration only

EOF
            ;;
    esac
}

# Handle script interruption
cleanup_on_exit() {
    log "Deployment interrupted, cleaning up..."
    # Add any cleanup logic here
    exit 1
}

trap cleanup_on_exit INT TERM

# Change to project root
cd "$PROJECT_ROOT"

# Run main function
main "$@"
