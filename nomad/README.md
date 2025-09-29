# GoTAK Nomad Deployment

This directory contains HashiCorp Nomad job specifications and deployment tools for running GoTAK (Team Awareness Kit) on a Nomad cluster.

## Architecture Overview

GoTAK consists of the following components:

- **GoTAK Server**: Main TAK server handling CoT (Cursor on Target) protocol, WebSocket connections, and REST API
- **PostgreSQL with PostGIS**: Geospatial database for storing TAK data  
- **Redis**: Caching and session storage
- **Web UI**: React-based tactical web interface (served by GoTAK Server)

## Directory Structure

```
nomad/
├── README.md                    # This file
├── jobs/                        # Nomad job specifications
│   ├── postgres.nomad.hcl      # PostgreSQL with PostGIS
│   ├── redis.nomad.hcl         # Redis cache
│   └── gotak-server.nomad.hcl  # GoTAK server with integrated web UI
├── variables/                   # Environment configuration
│   ├── cluster.hcl             # Cluster-wide settings
│   └── env/                    # Environment-specific settings
│       └── dev.hcl            # Development configuration
├── config/                      # Application configuration templates
│   └── server-nomad.yaml.tpl   # GoTAK server config template
└── scripts/                     # Deployment and management scripts
    ├── deploy-gotak.sh         # Main deployment script
    └── create-volumes.sh       # Volume setup script
```

## Quick Start

### Prerequisites

1. **Nomad Cluster**: Running Nomad cluster (single node or multi-node)
2. **Docker**: Docker daemon running on Nomad clients
3. **Consul** (Optional): For service discovery and health checking

### 1. Build Docker Images

```bash
# Build GoTAK images for Nomad deployment
make docker-build-nomad
```

### 2. Setup Volumes (Development)

```bash
# Create host volumes for development
./nomad/scripts/create-volumes.sh dev
```

**Note**: For development, you'll need to configure Nomad client with host volumes or run in dev mode.

### 3. Configure Nomad Client (Development)

Add to your Nomad client configuration:

```hcl
client {
  host_volume "gotak-postgres-data" {
    path      = "/tmp/gotak-nomad-volumes/dev/postgres-data"
    read_only = false
  }
  
  host_volume "gotak-redis-data" {
    path      = "/tmp/gotak-nomad-volumes/dev/redis-data"
    read_only = false
  }
}
```

Or run Nomad in dev mode with volumes:
```bash
nomad agent -dev -bind=127.0.0.1 -log-level=INFO \
  -client-host-volume=gotak-postgres-data=/tmp/gotak-nomad-volumes/dev/postgres-data \
  -client-host-volume=gotak-redis-data=/tmp/gotak-nomad-volumes/dev/redis-data
```

### 4. Deploy GoTAK Stack

```bash
# Deploy to development environment
./nomad/scripts/deploy-gotak.sh dev

# Or perform a dry run first
./nomad/scripts/deploy-gotak.sh dev true
```

### 5. Access GoTAK

Once deployed, GoTAK will be available at:

- **Web UI**: http://localhost:8080
- **API**: http://localhost:8080/api
- **CoT TCP**: localhost:8087
- **CoT TLS**: localhost:8089

## Configuration

### Environment Variables

Configuration is managed through HCL variable files:

- `variables/cluster.hcl`: Cluster-wide settings (datacenter, registry, etc.)
- `variables/env/dev.hcl`: Environment-specific settings (resources, replicas, features)

### Key Configuration Options

**Resource Allocation (dev.hcl)**:
```hcl
gotak_server_cpu = 200      # CPU MHz
gotak_server_memory = 256   # Memory MB
gotak_server_replicas = 1   # Number of instances
```

**Feature Toggles**:
```hcl
enable_debug = true         # Debug logging
enable_tls = false         # TLS for CoT connections
enable_metrics = true      # Prometheus metrics
enable_auth = false        # Authentication
```

## Network Ports

| Service | Port | Protocol | Description |
|---------|------|----------|-------------|
| GoTAK API/Web | 8080 | TCP | REST API and Web UI |
| GoTAK CoT | 8087 | TCP/UDP | TAK CoT protocol |
| GoTAK CoT TLS | 8089 | TCP | Secure TAK CoT protocol |
| PostgreSQL | 5432 | TCP | Database connections |
| Redis | 6379 | TCP | Cache connections |

## Deployment Commands

### Deploy Stack
```bash
# Deploy all services
./nomad/scripts/deploy-gotak.sh dev

# Deploy with dry run
./nomad/scripts/deploy-gotak.sh dev true
```

### Check Status
```bash
# Check deployment status
./nomad/scripts/deploy-gotak.sh status

# Check individual job
nomad job status gotak-server
```

### Manage Jobs
```bash
# Stop all services
./nomad/scripts/deploy-gotak.sh stop

# Restart all services
./nomad/scripts/deploy-gotak.sh restart

# Scale a service
nomad job scale gotak-server 2
```

### Logs and Debugging
```bash
# View logs
nomad alloc logs -f -job gotak-server
nomad alloc logs -f -job gotak-postgres
nomad alloc logs -f -job gotak-redis

# Access allocation shell
nomad alloc exec -job gotak-server sh

# Check service health (if Consul is available)
consul catalog services
consul health service gotak-api
```

## Service Discovery

Services are registered with Consul (if available) with the following names:

- `gotak-api`: REST API and Web UI (port 8080)
- `gotak-cot`: CoT TCP protocol (port 8087)  
- `gotak-cot-udp`: CoT UDP protocol (port 8087)
- `gotak-cot-tls`: CoT TLS protocol (port 8089)
- `postgres`: PostgreSQL database (port 5432)
- `redis`: Redis cache (port 6379)

Services can connect using Consul DNS:
- `gotak-api.service.consul:8080`
- `postgres.service.consul:5432`
- `redis.service.consul:6379`

## Getting Started

The fastest way to get GoTAK running on Nomad:

```bash
# 1. Build images
make docker-build-nomad

# 2. Setup volumes and deploy
./nomad/scripts/deploy-gotak.sh dev

# 3. Check status
./nomad/scripts/deploy-gotak.sh status

# 4. Access at http://localhost:8080
```

## Troubleshooting

### Common Issues

**Job Validation Fails**:
```bash
# Check job syntax
nomad job validate -var-file=variables/cluster.hcl -var-file=variables/env/dev.hcl jobs/gotak-server.nomad.hcl
```

**Volume Mount Errors**:
```bash
# Check host volumes are configured
nomad node status -verbose <node-id>

# Verify volume paths exist
ls -la /tmp/gotak-nomad-volumes/dev/
```

**Service Discovery Issues**:
```bash
# Check Consul services
consul catalog services
consul health service postgres
```

## Next Steps

1. **Test the Deployment**: Start with development environment
2. **Configure Production**: Create production variable files  
3. **Set up Monitoring**: Deploy observability stack
4. **Security Hardening**: Implement Vault, mTLS, and RBAC

For more information, see the main GoTAK documentation and the deployment scripts in the `scripts/` directory.