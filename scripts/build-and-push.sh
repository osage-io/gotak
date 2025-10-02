#!/bin/bash

# GoTAK Build and Push Script
# This script builds and pushes only the custom GoTAK containers to your registry

#set -xe  # Exit on any error

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
COMPOSE_FILE="${PROJECT_ROOT}/docker-compose.yml"

# Default values
DEFAULT_REGISTRY=""
DEFAULT_TAG="latest"
BUILD_SERVER=true
BUILD_WEB=true
PUSH_IMAGES=true
NO_CACHE=false

# Function to print colored messages
print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to display usage
usage() {
    cat << EOF
Usage: $0 [OPTIONS]

Build and push custom GoTAK containers to a Docker registry.

OPTIONS:
    -r, --registry REGISTRY    Docker registry URL (e.g., docker.io/username, ghcr.io/username)
    -t, --tag TAG             Tag for the images (default: latest)
    -s, --server-only         Build only the server image
    -w, --web-only           Build only the web image
    --no-push                Build only, don't push
    --no-cache               Build without using cache
    -h, --help               Show this help message

EXAMPLES:
    # Build and push both images to Docker Hub
    $0 -r docker.io/myusername

    # Build and push to GitHub Container Registry with custom tag
    $0 -r ghcr.io/myusername -t v1.0.0

    # Build only locally without pushing
    $0 --no-push

    # Build server only and push to AWS ECR
    $0 -r 123456789.dkr.ecr.us-east-1.amazonaws.com/gotak -s

EOF
    exit 0
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -r|--registry)
            REGISTRY="$2"
            shift 2
            ;;
        -t|--tag)
            TAG="$2"
            shift 2
            ;;
        -s|--server-only)
            BUILD_WEB=false
            shift
            ;;
        -w|--web-only)
            BUILD_SERVER=false
            shift
            ;;
        --no-push)
            PUSH_IMAGES=false
            shift
            ;;
        --no-cache)
            NO_CACHE=true
            shift
            ;;
        -h|--help)
            usage
            ;;
        *)
            print_error "Unknown option: $1"
            usage
            ;;
    esac
done

# Set defaults if not provided
TAG="${TAG:-$DEFAULT_TAG}"

# Validate registry if pushing
if [[ "$PUSH_IMAGES" == "true" ]] && [[ -z "$REGISTRY" ]]; then
    print_error "Registry is required when pushing images. Use -r option or --no-push"
    exit 1
fi

# Change to project root
cd "$PROJECT_ROOT"

# Get version information
if [[ -f ".git/HEAD" ]]; then
    GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
    GIT_BRANCH=$(git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "unknown")
    GIT_DIRTY=$(git diff --quiet || echo "-dirty")
    VERSION="${TAG}${GIT_DIRTY}"
else
    GIT_COMMIT="unknown"
    GIT_BRANCH="unknown"
    VERSION="$TAG"
fi

BUILD_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

print_info "Build Configuration:"
echo "  Project Root: $PROJECT_ROOT"
echo "  Registry: ${REGISTRY:-local}"
echo "  Tag: $TAG"
echo "  Version: $VERSION"
echo "  Git Commit: $GIT_COMMIT"
echo "  Git Branch: $GIT_BRANCH"
echo "  Build Time: $BUILD_TIME"
echo "  Build Server: $BUILD_SERVER"
echo "  Build Web: $BUILD_WEB"
echo "  Push Images: $PUSH_IMAGES"
echo ""

# Build arguments
BUILD_ARGS="--build-arg VERSION=$VERSION --build-arg BUILD_TIME=$BUILD_TIME --build-arg GIT_COMMIT=$GIT_COMMIT"

if [[ "$NO_CACHE" == "true" ]]; then
    BUILD_ARGS="$BUILD_ARGS --no-cache"
fi

# Function to build and optionally push an image
build_and_push() {
    local SERVICE=$1
    local IMAGE_NAME=$2
    local DOCKERFILE=$3
    local CONTEXT=$4
    
    print_info "Building $SERVICE..."
    
    # Determine the primary build tag
    local PRIMARY_TAG
    if [[ -n "$REGISTRY" ]]; then
        # Build directly with registry prefix
        PRIMARY_TAG="$REGISTRY/$IMAGE_NAME:latest"
    else
        # Build locally without registry
        PRIMARY_TAG="$IMAGE_NAME:latest"
    fi
    
    # Build the image with the primary tag
    if docker build $BUILD_ARGS -t "$PRIMARY_TAG" -f "$DOCKERFILE" "$CONTEXT"; then
        print_success "$SERVICE built successfully: $PRIMARY_TAG"
        
        # Create version tag if we have a specific version (not 'latest')
        if [[ "$TAG" != "latest" ]]; then
            local VERSION_TAG
            if [[ -n "$REGISTRY" ]]; then
                VERSION_TAG="$REGISTRY/$IMAGE_NAME:$TAG"
            else
                VERSION_TAG="$IMAGE_NAME:$TAG"
            fi
            
            # Tag with version
            docker tag "$PRIMARY_TAG" "$VERSION_TAG"
            print_success "Tagged $VERSION_TAG"
        fi
        
        # Push to registry if configured
        if [[ "$PUSH_IMAGES" == "true" ]] && [[ -n "$REGISTRY" ]]; then
            print_info "Pushing $SERVICE to registry..."
            
            # Push latest tag
            if docker push "$PRIMARY_TAG"; then
                print_success "Pushed $PRIMARY_TAG"
            else
                print_error "Failed to push $PRIMARY_TAG"
                return 1
            fi
            
            # Push version tag if it exists and is different from latest
            if [[ "$TAG" != "latest" ]]; then
                if docker push "$REGISTRY/$IMAGE_NAME:$TAG"; then
                    print_success "Pushed $REGISTRY/$IMAGE_NAME:$TAG"
                else
                    print_error "Failed to push $REGISTRY/$IMAGE_NAME:$TAG"
                    return 1
                fi
            fi
        fi
    else
        print_error "Failed to build $SERVICE"
        return 1
    fi
}

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    print_error "Docker is not running or not installed"
    exit 1
fi

# Login to registry if needed
check_registry_authentication() {
    print_info "Checking registry authentication..."

    # Determine the image to pull for authentication check
    local AUTH_TEST_IMAGE
    case "$REGISTRY" in
        docker.io/*)
            # For docker.io registries, check authentication against a public image on Docker Hub
            AUTH_TEST_IMAGE="hello-world"
            ;;
        *)
            # For other registries, attempt to pull a non-existent image from the registry base.
            # This relies on the registry returning an authentication error before a "not found" error.
            AUTH_TEST_IMAGE="${REGISTRY%/}/_auth_check_nonexistent_image:latest"
            ;;
    esac

    if ! docker pull "$AUTH_TEST_IMAGE" > /dev/null 2>&1; then
        print_warning "Not authenticated to $REGISTRY"
        
        # Special handling for different registries
        case "$REGISTRY" in
            ghcr.io/*)
                print_info "Please authenticate with: echo \$GITHUB_TOKEN | docker login ghcr.io -u USERNAME --password-stdin"
                ;;
            *.dkr.ecr.*.amazonaws.com/*)
                print_info "Please authenticate with: aws ecr get-login-password | docker login --username AWS --password-stdin $REGISTRY"
                ;;
            docker.io/*)
                print_info "Please authenticate with: docker login"
                ;;
            *)
                print_info "Please authenticate with: docker login $REGISTRY"
                ;;
        esac
        
        print_error "Please login to registry and run this script again"
        exit 1
    fi
    print_success "Registry authentication verified"
}

# Login to registry if needed
if [[ "$PUSH_IMAGES" == "true" ]] && [[ -n "$REGISTRY" ]]; then
    check_registry_authentication
fi

# Build GoTAK Server
if [[ "$BUILD_SERVER" == "true" ]]; then
    print_info "="
    print_info "Building GoTAK Server..."
    print_info "="
    
    build_and_push "gotak-server" "gotak-server" "Dockerfile" "."
    
    if [[ $? -eq 0 ]]; then
        print_success "GoTAK Server build complete"
    else
        print_error "GoTAK Server build failed"
        exit 1
    fi
fi

# Build GoTAK Web UI
if [[ "$BUILD_WEB" == "true" ]]; then
    print_info "="
    print_info "Building GoTAK Web UI..."
    print_info "="
    
    # Check if web directory exists
    if [[ -d "$PROJECT_ROOT/web" ]]; then
        build_and_push "gotak-web" "gotak-web" "web/Dockerfile" "web"
        
        if [[ $? -eq 0 ]]; then
            print_success "GoTAK Web UI build complete"
        else
            print_error "GoTAK Web UI build failed"
            exit 1
        fi
    else
        print_warning "Web directory not found, skipping web build"
    fi
fi

# Summary
echo ""
print_info "="
print_info "Build Summary"
print_info "="

if [[ "$BUILD_SERVER" == "true" ]]; then
    echo "  Server Image: gotak-server:$TAG"
    if [[ -n "$REGISTRY" ]]; then
        echo "    Registry: $REGISTRY/gotak-server:$TAG"
    fi
fi

if [[ "$BUILD_WEB" == "true" ]] && [[ -d "$PROJECT_ROOT/web" ]]; then
    echo "  Web Image: gotak-web:$TAG"
    if [[ -n "$REGISTRY" ]]; then
        echo "    Registry: $REGISTRY/gotak-web:$TAG"
    fi
fi

if [[ "$PUSH_IMAGES" == "true" ]] && [[ -n "$REGISTRY" ]]; then
    echo ""
    print_success "Images have been pushed to $REGISTRY"
    echo ""
    print_info "To deploy on production server:"
    echo "  1. Update docker-compose.yml with registry prefix:"
    echo "     image: $REGISTRY/gotak-server:$TAG"
    echo "     image: $REGISTRY/gotak-web:$TAG"
    echo "  2. Run: docker-compose pull"
    echo "  3. Run: docker-compose up -d"
else
    echo ""
    print_success "Images built locally"
    echo ""
    print_info "To run locally:"
    echo "  docker-compose up -d"
fi

print_success "Build and push script completed successfully!"
