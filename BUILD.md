# GoTAK Build & Push Guide

## Quick Start

### Build Locally (No Push)
```bash
./scripts/build-and-push.sh --no-push
```

### Build and Push to Docker Hub
```bash
# Login first
docker login

# Build and push
./scripts/build-and-push.sh -r docker.io/YOUR_USERNAME
```

### Build and Push to GitHub Container Registry
```bash
# Login first
echo $GITHUB_TOKEN | docker login ghcr.io -u YOUR_GITHUB_USERNAME --password-stdin

# Build and push
./scripts/build-and-push.sh -r ghcr.io/YOUR_GITHUB_USERNAME
```

## Common Usage Examples

### Production Release with Version Tag
```bash
# Tag with version number
./scripts/build-and-push.sh -r docker.io/YOUR_USERNAME -t v1.0.0

# This creates:
# - docker.io/YOUR_USERNAME/gotak-server:v1.0.0
# - docker.io/YOUR_USERNAME/gotak-server:latest
# - docker.io/YOUR_USERNAME/gotak-web:v1.0.0
# - docker.io/YOUR_USERNAME/gotak-web:latest
```

### Development Build (Local Only)
```bash
# Build without cache for clean build
./scripts/build-and-push.sh --no-push --no-cache
```

### Build Server Only
```bash
# Useful when only server code changes
./scripts/build-and-push.sh -r docker.io/YOUR_USERNAME --server-only
```

### Build Web UI Only
```bash
# Useful when only frontend changes
./scripts/build-and-push.sh -r docker.io/YOUR_USERNAME --web-only
```

## Registry Options

### Docker Hub
```bash
# Public repository
./scripts/build-and-push.sh -r docker.io/YOUR_USERNAME

# Private repository (requires docker login)
docker login
./scripts/build-and-push.sh -r docker.io/YOUR_ORGANIZATION
```

### GitHub Container Registry
```bash
# Requires GitHub personal access token with packages:write permission
echo $GITHUB_TOKEN | docker login ghcr.io -u YOUR_USERNAME --password-stdin
./scripts/build-and-push.sh -r ghcr.io/YOUR_USERNAME
```

### AWS ECR
```bash
# Get login token
aws ecr get-login-password --region us-east-1 | \
  docker login --username AWS --password-stdin 123456789.dkr.ecr.us-east-1.amazonaws.com

# Build and push
./scripts/build-and-push.sh -r 123456789.dkr.ecr.us-east-1.amazonaws.com
```

### Private Registry
```bash
# Login to your registry
docker login registry.company.com

# Build and push
./scripts/build-and-push.sh -r registry.company.com/gotak
```

## Script Options

| Option | Description | Example |
|--------|-------------|---------|
| `-r, --registry` | Docker registry URL | `docker.io/username` |
| `-t, --tag` | Image tag (default: latest) | `v1.0.0` |
| `-s, --server-only` | Build only server image | |
| `-w, --web-only` | Build only web UI image | |
| `--no-push` | Build locally without pushing | |
| `--no-cache` | Build without using cache | |
| `-h, --help` | Show help message | |

## Deployment After Push

Once images are pushed to your registry:

1. **Update production docker-compose.yml**:
```yaml
services:
  gotak:
    image: YOUR_REGISTRY/gotak-server:v1.0.0
    # ... rest of config

  web:
    image: YOUR_REGISTRY/gotak-web:v1.0.0
    # ... rest of config
```

2. **On production server**:
```bash
# Pull new images
docker-compose pull

# Restart with new images
docker-compose up -d

# Verify deployment
docker-compose ps
docker-compose logs --tail=50
```

## CI/CD Integration

### GitHub Actions Example
```yaml
name: Build and Push

on:
  push:
    tags:
      - 'v*'

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Login to GitHub Container Registry
        run: echo "${{ secrets.GITHUB_TOKEN }}" | docker login ghcr.io -u ${{ github.actor }} --password-stdin
      
      - name: Build and Push
        run: ./scripts/build-and-push.sh -r ghcr.io/${{ github.repository_owner }} -t ${{ github.ref_name }}
```

### GitLab CI Example
```yaml
build:
  stage: build
  script:
    - docker login -u $CI_REGISTRY_USER -p $CI_REGISTRY_PASSWORD $CI_REGISTRY
    - ./scripts/build-and-push.sh -r $CI_REGISTRY_IMAGE -t $CI_COMMIT_TAG
  only:
    - tags
```

## Troubleshooting

### Registry Authentication Issues
```bash
# Check if logged in
docker pull YOUR_REGISTRY/library/hello-world

# Re-authenticate
docker logout YOUR_REGISTRY
docker login YOUR_REGISTRY
```

### Build Failures
```bash
# Clean build without cache
./scripts/build-and-push.sh --no-push --no-cache

# Check Docker disk space
docker system df

# Clean up if needed
docker system prune -a
```

### Push Timeouts
```bash
# For large images, increase timeout
export DOCKER_CLIENT_TIMEOUT=300
export COMPOSE_HTTP_TIMEOUT=300

# Then run build script
./scripts/build-and-push.sh -r YOUR_REGISTRY
```

## Image Management

### List Local Images
```bash
docker images | grep gotak
```

### Remove Old Images
```bash
# Remove specific version
docker rmi gotak-server:old-version
docker rmi gotak-web:old-version

# Remove all gotak images
docker rmi $(docker images | grep gotak | awk '{print $3}')
```

### Check Image Size
```bash
docker images --format "table {{.Repository}}\t{{.Tag}}\t{{.Size}}" | grep gotak
```

## Security Notes

1. **Never commit registry credentials** to git
2. **Use CI/CD secrets** for automated builds
3. **Scan images for vulnerabilities**:
   ```bash
   docker scan gotak-server:latest
   ```
4. **Sign images** for production:
   ```bash
   docker trust sign YOUR_REGISTRY/gotak-server:v1.0.0
   ```

## Support

For issues with the build script, check:
1. Docker daemon is running: `docker info`
2. Sufficient disk space: `df -h`
3. Network connectivity to registry
4. Valid authentication tokens